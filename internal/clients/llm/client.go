// Package llm provides a provider-agnostic LLM completion client.
package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

// HTTPClient defines the interface for HTTP operations.
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// ClientConfig holds configuration options for the AI client.
type ClientConfig struct {
	// Provider is the AI service provider (gemini, claude, openai, qwen).
	Provider Provider

	// APIKey is the API key for authentication.
	APIKey string

	// Model is the model name to use.
	Model string

	// BaseURL overrides the default API base URL for the provider.
	// If empty, the provider's default URL is used.
	BaseURL string

	// MaxTokens is the maximum number of tokens in the response.
	MaxTokens int

	// Timeout is the HTTP request timeout.
	Timeout time.Duration

	// HTTPClient is the HTTP client to use for requests.
	HTTPClient HTTPClient

	// Logger is the structured logger.
	Logger *slog.Logger
}

// validate checks the configuration and applies defaults.
func (c *ClientConfig) validate() error {
	if c.APIKey == "" {
		return fmt.Errorf("ai: api key is required")
	}

	if c.Provider == "" {
		c.Provider = ProviderGemini
	}

	if c.Model == "" {
		switch c.Provider {
		case ProviderGemini:
			c.Model = "gemini-2.0-flash"
		case ProviderClaude:
			c.Model = "claude-sonnet-4-20250514"
		case ProviderOpenAI:
			c.Model = "gpt-4o"
		case ProviderQwen:
			c.Model = "qwen-plus"
		}
	}

	if c.MaxTokens <= 0 {
		c.MaxTokens = 1024
	}

	if c.Timeout <= 0 {
		c.Timeout = 60 * time.Second
	}

	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{
			Timeout: c.Timeout,
		}
	}

	if c.Logger == nil {
		c.Logger = slog.Default()
	}

	return nil
}

// Client is a provider-agnostic AI completion client.
type Client struct {
	config ClientConfig
}

// ClientOption is a functional option for configuring the AI client.
type ClientOption func(*ClientConfig)

// WithProvider sets the AI provider.
func WithProvider(p Provider) ClientOption {
	return func(c *ClientConfig) {
		c.Provider = p
	}
}

// WithModel sets the model name.
func WithModel(model string) ClientOption {
	return func(c *ClientConfig) {
		c.Model = model
	}
}

// WithBaseURL overrides the default API base URL for the provider.
func WithBaseURL(url string) ClientOption {
	return func(c *ClientConfig) {
		c.BaseURL = url
	}
}

// WithMaxTokens sets the maximum response tokens.
func WithMaxTokens(n int) ClientOption {
	return func(c *ClientConfig) {
		c.MaxTokens = n
	}
}

// WithLLMTimeout sets the HTTP request timeout.
func WithLLMTimeout(d time.Duration) ClientOption {
	return func(c *ClientConfig) {
		c.Timeout = d
	}
}

// WithLLMHTTPClient sets the HTTP client.
func WithLLMHTTPClient(client HTTPClient) ClientOption {
	return func(c *ClientConfig) {
		c.HTTPClient = client
	}
}

// WithLLMLogger sets the logger.
func WithLLMLogger(logger *slog.Logger) ClientOption {
	return func(c *ClientConfig) {
		c.Logger = logger
	}
}

// NewClient creates a new AI client with functional options.
func NewClient(apiKey string, opts ...ClientOption) (*Client, error) {
	config := ClientConfig{
		APIKey: apiKey,
	}

	for _, opt := range opts {
		opt(&config)
	}

	if err := config.validate(); err != nil {
		return nil, err
	}

	return &Client{config: config}, nil
}

// Complete sends a chat completion request to the configured AI provider.
func (c *Client) Complete(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	if req.Model == "" {
		req.Model = c.config.Model
	}
	if req.MaxTokens <= 0 {
		req.MaxTokens = c.config.MaxTokens
	}

	c.config.Logger.Debug("ai completion request",
		slog.String("provider", string(c.config.Provider)),
		slog.String("model", req.Model),
		slog.Int("messages", len(req.Messages)),
	)

	switch c.config.Provider {
	case ProviderGemini:
		return c.completeGemini(ctx, req)
	case ProviderClaude:
		return c.completeClaude(ctx, req)
	case ProviderOpenAI, ProviderQwen:
		return c.completeOpenAICompatible(ctx, req)
	default:
		return nil, fmt.Errorf("ai: unsupported provider: %s", c.config.Provider)
	}
}

// completeGemini sends a request to Google Gemini API.
// POST https://generativelanguage.googleapis.com/v1beta/models/{model}:generateContent
func (c *Client) completeGemini(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Build Gemini request body
	type part struct {
		Text             string          `json:"text,omitempty"`
		FunctionCall     *functionCall   `json:"functionCall,omitempty"`
		FunctionResponse *functionResp   `json:"functionResponse,omitempty"`
	}
	type content struct {
		Role  string `json:"role"`
		Parts []part `json:"parts"`
	}
	type generationConfig struct {
		MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
		Temperature     float64 `json:"temperature,omitempty"`
	}
	type systemInstruction struct {
		Parts []part `json:"parts"`
	}
	type funcDecl struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Parameters  json.RawMessage `json:"parameters,omitempty"`
	}
	type geminiTool struct {
		FunctionDeclarations []funcDecl `json:"functionDeclarations"`
	}

	type geminiRequest struct {
		Contents          []content          `json:"contents"`
		GenerationConfig  *generationConfig  `json:"generationConfig,omitempty"`
		SystemInstruction *systemInstruction `json:"systemInstruction,omitempty"`
		Tools             []geminiTool       `json:"tools,omitempty"`
	}

	gemReq := geminiRequest{}

	// Set system instruction if provided
	if req.System != "" {
		gemReq.SystemInstruction = &systemInstruction{
			Parts: []part{{Text: req.System}},
		}
	}

	// Add tool definitions if provided
	if len(req.Tools) > 0 {
		var decls []funcDecl
		for _, t := range req.Tools {
			decls = append(decls, funcDecl{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			})
		}
		gemReq.Tools = []geminiTool{{FunctionDeclarations: decls}}
	}

	// Convert messages (skip system role, Gemini uses systemInstruction)
	for _, msg := range req.Messages {
		if msg.Role == RoleSystem {
			continue
		}

		switch {
		case msg.Role == RoleAssistant && len(msg.ToolCalls) > 0:
			// Assistant message with tool calls → model role with functionCall parts
			var parts []part
			if msg.Content != "" {
				parts = append(parts, part{Text: msg.Content})
			}
			for _, tc := range msg.ToolCalls {
				parts = append(parts, part{
					FunctionCall: &functionCall{
						Name: tc.Name,
						Args: tc.Arguments,
					},
				})
			}
			gemReq.Contents = append(gemReq.Contents, content{
				Role:  "model",
				Parts: parts,
			})

		case msg.Role == RoleTool:
			// Tool result → function role with functionResponse part
			gemReq.Contents = append(gemReq.Contents, content{
				Role: "function",
				Parts: []part{{
					FunctionResponse: &functionResp{
						Name:     msg.ToolCallID,
						Response: json.RawMessage(msg.Content),
					},
				}},
			})

		default:
			role := "user"
			if msg.Role == RoleAssistant {
				role = "model"
			}
			gemReq.Contents = append(gemReq.Contents, content{
				Role:  role,
				Parts: []part{{Text: msg.Content}},
			})
		}
	}

	gemReq.GenerationConfig = &generationConfig{
		MaxOutputTokens: req.MaxTokens,
	}
	if req.Temperature > 0 {
		gemReq.GenerationConfig.Temperature = req.Temperature
	}

	apiURL := fmt.Sprintf("%s/v1beta/models/%s:generateContent", c.baseURL(), req.Model)

	httpReq, err := c.buildRequest(ctx, http.MethodPost, apiURL, gemReq)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("x-goog-api-key", c.config.APIKey)

	return c.doRequest(httpReq, c.config.Provider, c.parseGeminiResponse)
}

// Gemini function calling types (package-level for reuse in request/response).
type functionCall struct {
	Name string          `json:"name"`
	Args json.RawMessage `json:"args"`
}
type functionResp struct {
	Name     string          `json:"name"`
	Response json.RawMessage `json:"response"`
}

// completeClaude sends a request to Anthropic Claude API.
// POST https://api.anthropic.com/v1/messages
func (c *Client) completeClaude(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	type claudeContentBlock struct {
		Type      string          `json:"type"`
		Text      string          `json:"text,omitempty"`
		ID        string          `json:"id,omitempty"`
		Name      string          `json:"name,omitempty"`
		Input     json.RawMessage `json:"input,omitempty"`
		ToolUseID string          `json:"tool_use_id,omitempty"`
		Content   string          `json:"content,omitempty"`
	}

	type claudeMessage struct {
		Role    string      `json:"role"`
		Content interface{} `json:"content"` // string or []claudeContentBlock
	}

	type claudeTool struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		InputSchema json.RawMessage `json:"input_schema"`
	}

	type claudeRequest struct {
		Model     string          `json:"model"`
		Messages  []claudeMessage `json:"messages"`
		MaxTokens int             `json:"max_tokens"`
		System    string          `json:"system,omitempty"`
		Tools     []claudeTool    `json:"tools,omitempty"`
	}

	claudeReq := claudeRequest{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
		System:    req.System,
	}

	// Add tool definitions if provided
	if len(req.Tools) > 0 {
		for _, t := range req.Tools {
			claudeReq.Tools = append(claudeReq.Tools, claudeTool{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: t.Parameters,
			})
		}
	}

	for _, msg := range req.Messages {
		if msg.Role == RoleSystem {
			continue
		}

		switch {
		case msg.Role == RoleAssistant && len(msg.ToolCalls) > 0:
			// Assistant with tool calls → content blocks
			var blocks []claudeContentBlock
			if msg.Content != "" {
				blocks = append(blocks, claudeContentBlock{
					Type: "text",
					Text: msg.Content,
				})
			}
			for _, tc := range msg.ToolCalls {
				blocks = append(blocks, claudeContentBlock{
					Type:  "tool_use",
					ID:    tc.ID,
					Name:  tc.Name,
					Input: tc.Arguments,
				})
			}
			claudeReq.Messages = append(claudeReq.Messages, claudeMessage{
				Role:    "assistant",
				Content: blocks,
			})

		case msg.Role == RoleTool:
			// Tool result → user message with tool_result content block
			claudeReq.Messages = append(claudeReq.Messages, claudeMessage{
				Role: "user",
				Content: []claudeContentBlock{{
					Type:      "tool_result",
					ToolUseID: msg.ToolCallID,
					Content:   msg.Content,
				}},
			})

		default:
			claudeReq.Messages = append(claudeReq.Messages, claudeMessage{
				Role:    string(msg.Role),
				Content: msg.Content,
			})
		}
	}

	apiURL := c.baseURL() + "/v1/messages"

	httpReq, err := c.buildRequest(ctx, http.MethodPost, apiURL, claudeReq)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("x-api-key", c.config.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	return c.doRequest(httpReq, c.config.Provider, c.parseClaudeResponse)
}

// defaultBaseURL returns the default API base URL for a provider.
func defaultBaseURL(provider Provider) string {
	switch provider {
	case ProviderGemini:
		return "https://generativelanguage.googleapis.com"
	case ProviderClaude:
		return "https://api.anthropic.com"
	case ProviderOpenAI:
		return "https://api.openai.com"
	case ProviderQwen:
		return "https://dashscope.aliyuncs.com/compatible-mode"
	default:
		return ""
	}
}

// baseURL returns the configured or default base URL for the current provider.
func (c *Client) baseURL() string {
	if c.config.BaseURL != "" {
		return c.config.BaseURL
	}
	return defaultBaseURL(c.config.Provider)
}

// completeOpenAICompatible sends a request to OpenAI-compatible APIs (OpenAI, Qwen).
func (c *Client) completeOpenAICompatible(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	type oaiToolCallFunc struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	}
	type oaiToolCall struct {
		ID       string          `json:"id"`
		Type     string          `json:"type"`
		Function oaiToolCallFunc `json:"function"`
	}
	type openAIMessage struct {
		Role       string        `json:"role"`
		Content    string        `json:"content,omitempty"`
		ToolCalls  []oaiToolCall `json:"tool_calls,omitempty"`
		ToolCallID string        `json:"tool_call_id,omitempty"`
	}
	type oaiFunction struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		Parameters  json.RawMessage `json:"parameters,omitempty"`
	}
	type oaiTool struct {
		Type     string      `json:"type"`
		Function oaiFunction `json:"function"`
	}

	type openAIRequest struct {
		Model     string          `json:"model"`
		Messages  []openAIMessage `json:"messages"`
		MaxTokens int             `json:"max_tokens,omitempty"`
		Tools     []oaiTool       `json:"tools,omitempty"`
	}

	oaiReq := openAIRequest{
		Model:     req.Model,
		MaxTokens: req.MaxTokens,
	}

	// Add tool definitions if provided
	if len(req.Tools) > 0 {
		for _, t := range req.Tools {
			oaiReq.Tools = append(oaiReq.Tools, oaiTool{
				Type: "function",
				Function: oaiFunction{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  t.Parameters,
				},
			})
		}
	}

	// Include system prompt as a system message for OpenAI
	if req.System != "" {
		oaiReq.Messages = append(oaiReq.Messages, openAIMessage{
			Role:    "system",
			Content: req.System,
		})
	}

	for _, msg := range req.Messages {
		if msg.Role == RoleSystem {
			continue
		}

		switch {
		case msg.Role == RoleAssistant && len(msg.ToolCalls) > 0:
			// Assistant with tool calls
			var calls []oaiToolCall
			for _, tc := range msg.ToolCalls {
				calls = append(calls, oaiToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: oaiToolCallFunc{
						Name:      tc.Name,
						Arguments: string(tc.Arguments),
					},
				})
			}
			oaiReq.Messages = append(oaiReq.Messages, openAIMessage{
				Role:      "assistant",
				Content:   msg.Content,
				ToolCalls: calls,
			})

		case msg.Role == RoleTool:
			// Tool result
			oaiReq.Messages = append(oaiReq.Messages, openAIMessage{
				Role:       "tool",
				Content:    msg.Content,
				ToolCallID: msg.ToolCallID,
			})

		default:
			oaiReq.Messages = append(oaiReq.Messages, openAIMessage{
				Role:    string(msg.Role),
				Content: msg.Content,
			})
		}
	}

	apiURL := c.baseURL()

	httpReq, err := c.buildRequest(ctx, http.MethodPost, apiURL, oaiReq)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+c.config.APIKey)

	return c.doRequest(httpReq, c.config.Provider, c.parseOpenAIResponse)
}

// buildRequest creates an HTTP request with JSON body.
func (c *Client) buildRequest(ctx context.Context, method, url string, body interface{}) (*http.Request, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("ai: failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("ai: failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// responseParser is a function that parses a provider-specific response body.
type responseParser func(body []byte) (*ChatResponse, error)

// doRequest executes an HTTP request and parses the response.
func (c *Client) doRequest(req *http.Request, provider Provider, parser responseParser) (*ChatResponse, error) {
	resp, err := c.config.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ai: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ai: failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, c.parseErrorResponse(body, resp.StatusCode, provider)
	}

	return parser(body)
}

// parseErrorResponse extracts an LLMError from a failed response.
func (c *Client) parseErrorResponse(body []byte, statusCode int, provider Provider) error {
	aiErr := &LLMError{
		Code:     statusCode,
		Provider: provider,
	}

	// Try to extract error description from JSON
	var errResp struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &errResp) == nil && errResp.Error.Message != "" {
		aiErr.Description = errResp.Error.Message
	} else {
		aiErr.Description = string(body)
	}

	// Check for retry-after header info in the response body
	if statusCode == 429 {
		var rateLimitResp struct {
			Error struct {
				RetryAfter float64 `json:"retry_after"`
			} `json:"error"`
		}
		if json.Unmarshal(body, &rateLimitResp) == nil && rateLimitResp.Error.RetryAfter > 0 {
			aiErr.RetryAfter = int(rateLimitResp.Error.RetryAfter)
		}
		if aiErr.RetryAfter == 0 {
			aiErr.RetryAfter = 5 // default 5s for rate limits
		}
	}

	return aiErr
}

// parseGeminiResponse parses a Gemini API response.
func (c *Client) parseGeminiResponse(body []byte) (*ChatResponse, error) {
	var resp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text         string        `json:"text,omitempty"`
					FunctionCall *functionCall `json:"functionCall,omitempty"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		UsageMetadata struct {
			PromptTokenCount     int `json:"promptTokenCount"`
			CandidatesTokenCount int `json:"candidatesTokenCount"`
		} `json:"usageMetadata"`
		ModelVersion string `json:"modelVersion"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("ai: failed to parse gemini response: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("ai: gemini returned no content")
	}

	result := &ChatResponse{
		Model:        resp.ModelVersion,
		InputTokens:  resp.UsageMetadata.PromptTokenCount,
		OutputTokens: resp.UsageMetadata.CandidatesTokenCount,
	}

	// Extract text and function calls from parts
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.FunctionCall != nil {
			result.ToolCalls = append(result.ToolCalls, ToolCall{
				ID:        part.FunctionCall.Name, // Gemini uses name as ID
				Name:      part.FunctionCall.Name,
				Arguments: part.FunctionCall.Args,
			})
		} else if part.Text != "" {
			result.Content += part.Text
		}
	}

	return result, nil
}

// parseClaudeResponse parses a Claude API response.
func (c *Client) parseClaudeResponse(body []byte) (*ChatResponse, error) {
	var resp struct {
		Content []struct {
			Type  string          `json:"type"`
			Text  string          `json:"text,omitempty"`
			ID    string          `json:"id,omitempty"`
			Name  string          `json:"name,omitempty"`
			Input json.RawMessage `json:"input,omitempty"`
		} `json:"content"`
		Model    string `json:"model"`
		StopReason string `json:"stop_reason"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("ai: failed to parse claude response: %w", err)
	}

	if len(resp.Content) == 0 {
		return nil, fmt.Errorf("ai: claude returned no content")
	}

	result := &ChatResponse{
		Model:        resp.Model,
		InputTokens:  resp.Usage.InputTokens,
		OutputTokens: resp.Usage.OutputTokens,
	}

	// Extract text and tool_use blocks
	for _, block := range resp.Content {
		switch block.Type {
		case "text":
			result.Content += block.Text
		case "tool_use":
			result.ToolCalls = append(result.ToolCalls, ToolCall{
				ID:        block.ID,
				Name:      block.Name,
				Arguments: block.Input,
			})
		}
	}

	// If there are no tool calls and no text, that's an error
	if result.Content == "" && len(result.ToolCalls) == 0 {
		return nil, fmt.Errorf("ai: claude returned no text or tool_use content")
	}

	return result, nil
}

// parseOpenAIResponse parses an OpenAI API response.
func (c *Client) parseOpenAIResponse(body []byte) (*ChatResponse, error) {
	var resp struct {
		Choices []struct {
			Message struct {
				Content   string `json:"content"`
				ToolCalls []struct {
					ID       string `json:"id"`
					Type     string `json:"type"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls,omitempty"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Model string `json:"model"`
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
		} `json:"usage"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("ai: failed to parse openai response: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("ai: openai returned no choices")
	}

	result := &ChatResponse{
		Content:      resp.Choices[0].Message.Content,
		Model:        resp.Model,
		InputTokens:  resp.Usage.PromptTokens,
		OutputTokens: resp.Usage.CompletionTokens,
	}

	// Extract tool calls if present
	for _, tc := range resp.Choices[0].Message.ToolCalls {
		result.ToolCalls = append(result.ToolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: json.RawMessage(tc.Function.Arguments),
		})
	}

	return result, nil
}
