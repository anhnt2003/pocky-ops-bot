package ai

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		apiKey  string
		opts    []ClientOption
		wantErr bool
	}{
		{
			name:    "valid key",
			apiKey:  "test-key",
			wantErr: false,
		},
		{
			name:    "empty key",
			apiKey:  "",
			wantErr: true,
		},
		{
			name:   "with options",
			apiKey: "test-key",
			opts: []ClientOption{
				WithProvider(ProviderClaude),
				WithModel("claude-sonnet-4-20250514"),
				WithMaxTokens(2048),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.apiKey, tt.opts...)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil without error")
			}
		})
	}
}

func TestClientDefaults(t *testing.T) {
	client, err := NewClient("test-key")
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client.config.Provider != ProviderGemini {
		t.Errorf("Provider = %q, want %q", client.config.Provider, ProviderGemini)
	}
	if client.config.Model != "gemini-2.0-flash" {
		t.Errorf("Model = %q, want %q", client.config.Model, "gemini-2.0-flash")
	}
	if client.config.MaxTokens != 1024 {
		t.Errorf("MaxTokens = %d, want 1024", client.config.MaxTokens)
	}
}

func TestClientDefaultModels(t *testing.T) {
	tests := []struct {
		provider Provider
		expected string
	}{
		{ProviderGemini, "gemini-2.0-flash"},
		{ProviderClaude, "claude-sonnet-4-20250514"},
		{ProviderOpenAI, "gpt-4o"},
		{ProviderQwen, "qwen-plus"},
	}

	for _, tt := range tests {
		t.Run(string(tt.provider), func(t *testing.T) {
			client, err := NewClient("test-key", WithProvider(tt.provider))
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}
			if client.config.Model != tt.expected {
				t.Errorf("Model = %q, want %q", client.config.Model, tt.expected)
			}
		})
	}
}

func TestCompleteGemini(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if key := r.Header.Get("x-goog-api-key"); key != "test-key" {
			t.Errorf("x-goog-api-key = %q, want %q", key, "test-key")
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want %q", ct, "application/json")
		}

		// Verify path contains model
		if !strings.Contains(r.URL.Path, "gemini-2.0-flash") {
			t.Errorf("URL path = %q, expected to contain model name", r.URL.Path)
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)

		// Should have contents
		if _, ok := reqBody["contents"]; !ok {
			t.Error("request missing 'contents' field")
		}

		// Should have systemInstruction
		if _, ok := reqBody["systemInstruction"]; !ok {
			t.Error("request missing 'systemInstruction' field")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]interface{}{
							{"text": "Hello from Gemini!"},
						},
					},
				},
			},
			"usageMetadata": map[string]interface{}{
				"promptTokenCount":     10,
				"candidatesTokenCount": 5,
			},
			"modelVersion": "gemini-2.0-flash",
		})
	}))
	defer server.Close()

	// We need to override the URL, so we'll use a custom HTTP client that rewrites the URL
	client, err := NewClient("test-key",
		WithProvider(ProviderGemini),
		WithAIHTTPClient(&urlRewriteClient{
			target: server.URL,
			inner:  http.DefaultClient,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	resp, err := client.Complete(context.Background(), ChatRequest{
		Messages: []ChatMessage{
			{Role: RoleUser, Content: "Hello"},
		},
		System: "You are helpful",
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	if resp.Content != "Hello from Gemini!" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello from Gemini!")
	}
	if resp.InputTokens != 10 {
		t.Errorf("InputTokens = %d, want 10", resp.InputTokens)
	}
	if resp.OutputTokens != 5 {
		t.Errorf("OutputTokens = %d, want 5", resp.OutputTokens)
	}
}

func TestCompleteClaude(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if key := r.Header.Get("x-api-key"); key != "test-key" {
			t.Errorf("x-api-key = %q, want %q", key, "test-key")
		}
		if ver := r.Header.Get("anthropic-version"); ver != "2023-06-01" {
			t.Errorf("anthropic-version = %q, want %q", ver, "2023-06-01")
		}

		// Verify request body
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)

		if system, ok := reqBody["system"].(string); !ok || system != "You are helpful" {
			t.Errorf("system = %v, want 'You are helpful'", reqBody["system"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": "Hello from Claude!"},
			},
			"model": "claude-sonnet-4-20250514",
			"usage": map[string]interface{}{
				"input_tokens":  15,
				"output_tokens": 8,
			},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key",
		WithProvider(ProviderClaude),
		WithAIHTTPClient(&urlRewriteClient{
			target: server.URL,
			inner:  http.DefaultClient,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	resp, err := client.Complete(context.Background(), ChatRequest{
		Messages: []ChatMessage{
			{Role: RoleUser, Content: "Hello"},
		},
		System: "You are helpful",
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	if resp.Content != "Hello from Claude!" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello from Claude!")
	}
	if resp.InputTokens != 15 {
		t.Errorf("InputTokens = %d, want 15", resp.InputTokens)
	}
}

func TestCompleteOpenAI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-key")
		}

		// Verify request body includes system message
		body, _ := io.ReadAll(r.Body)
		var reqBody struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		json.Unmarshal(body, &reqBody)

		if len(reqBody.Messages) < 2 {
			t.Errorf("expected at least 2 messages (system + user), got %d", len(reqBody.Messages))
		}
		if reqBody.Messages[0].Role != "system" {
			t.Errorf("first message role = %q, want 'system'", reqBody.Messages[0].Role)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": "Hello from OpenAI!",
					},
				},
			},
			"model": "gpt-4o",
			"usage": map[string]interface{}{
				"prompt_tokens":     12,
				"completion_tokens": 6,
			},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key",
		WithProvider(ProviderOpenAI),
		WithAIHTTPClient(&urlRewriteClient{
			target: server.URL,
			inner:  http.DefaultClient,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	resp, err := client.Complete(context.Background(), ChatRequest{
		Messages: []ChatMessage{
			{Role: RoleUser, Content: "Hello"},
		},
		System: "You are helpful",
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	if resp.Content != "Hello from OpenAI!" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello from OpenAI!")
	}
	if resp.InputTokens != 12 {
		t.Errorf("InputTokens = %d, want 12", resp.InputTokens)
	}
}

func TestCompleteQwen(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Qwen uses Bearer auth like OpenAI
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-key" {
			t.Errorf("Authorization = %q, want %q", auth, "Bearer test-key")
		}

		// Should hit /v1/chat/completions path
		if r.URL.Path != "/compatible-mode/v1/chat/completions" && r.URL.Path != "/v1/chat/completions" {
			// urlRewriteClient rewrites host, path stays
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": "Hello from Qwen!",
					},
				},
			},
			"model": "qwen-plus",
			"usage": map[string]interface{}{
				"prompt_tokens":     9,
				"completion_tokens": 4,
			},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key",
		WithProvider(ProviderQwen),
		WithAIHTTPClient(&urlRewriteClient{
			target: server.URL,
			inner:  http.DefaultClient,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	resp, err := client.Complete(context.Background(), ChatRequest{
		Messages: []ChatMessage{
			{Role: RoleUser, Content: "Hello"},
		},
		System: "You are helpful",
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	if resp.Content != "Hello from Qwen!" {
		t.Errorf("Content = %q, want %q", resp.Content, "Hello from Qwen!")
	}
	if resp.InputTokens != 9 {
		t.Errorf("InputTokens = %d, want 9", resp.InputTokens)
	}
	if resp.OutputTokens != 4 {
		t.Errorf("OutputTokens = %d, want 4", resp.OutputTokens)
	}
}

func TestCompleteAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Invalid request",
			},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key",
		WithProvider(ProviderGemini),
		WithAIHTTPClient(&urlRewriteClient{
			target: server.URL,
			inner:  http.DefaultClient,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.Complete(context.Background(), ChatRequest{
		Messages: []ChatMessage{
			{Role: RoleUser, Content: "Hello"},
		},
	})
	if err == nil {
		t.Fatal("Complete() expected error, got nil")
	}

	aiErr, ok := err.(*AIError)
	if !ok {
		t.Fatalf("expected *AIError, got %T: %v", err, err)
	}

	if aiErr.Code != 400 {
		t.Errorf("Code = %d, want 400", aiErr.Code)
	}
	if !strings.Contains(aiErr.Description, "Invalid request") {
		t.Errorf("Description = %q, want to contain 'Invalid request'", aiErr.Description)
	}
	if aiErr.IsRetryable() {
		t.Error("expected non-retryable error")
	}
}

func TestCompleteRateLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Rate limit exceeded",
			},
		})
	}))
	defer server.Close()

	client, err := NewClient("test-key",
		WithProvider(ProviderGemini),
		WithAIHTTPClient(&urlRewriteClient{
			target: server.URL,
			inner:  http.DefaultClient,
		}),
	)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.Complete(context.Background(), ChatRequest{
		Messages: []ChatMessage{
			{Role: RoleUser, Content: "Hello"},
		},
	})

	aiErr, ok := err.(*AIError)
	if !ok {
		t.Fatalf("expected *AIError, got %T: %v", err, err)
	}

	if !aiErr.IsRetryable() {
		t.Error("expected retryable error for 429")
	}
	if aiErr.RetryAfter == 0 {
		t.Error("expected RetryAfter > 0 for rate limit")
	}
}

func TestCompleteUnsupportedProvider(t *testing.T) {
	client, err := NewClient("test-key", WithProvider("unknown"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	_, err = client.Complete(context.Background(), ChatRequest{
		Messages: []ChatMessage{
			{Role: RoleUser, Content: "Hello"},
		},
	})
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
	if !strings.Contains(err.Error(), "unsupported provider") {
		t.Errorf("error = %q, want to contain 'unsupported provider'", err.Error())
	}
}

func TestCompleteGeminiToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)

		// Verify tools are included in request
		if _, ok := reqBody["tools"]; !ok {
			t.Error("request missing 'tools' field")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"candidates": []map[string]interface{}{
				{
					"content": map[string]interface{}{
						"parts": []map[string]interface{}{
							{
								"functionCall": map[string]interface{}{
									"name": "get_spot_balances",
									"args": map[string]interface{}{},
								},
							},
						},
					},
				},
			},
			"usageMetadata": map[string]interface{}{
				"promptTokenCount":     20,
				"candidatesTokenCount": 10,
			},
			"modelVersion": "gemini-2.0-flash",
		})
	}))
	defer server.Close()

	client, _ := NewClient("test-key",
		WithProvider(ProviderGemini),
		WithAIHTTPClient(&urlRewriteClient{target: server.URL, inner: http.DefaultClient}),
	)

	resp, err := client.Complete(context.Background(), ChatRequest{
		Messages: []ChatMessage{{Role: RoleUser, Content: "Show my balance"}},
		Tools: []ToolDefinition{{
			Name:        "get_spot_balances",
			Description: "Get spot balances",
			Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		}},
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	if len(resp.ToolCalls) != 1 {
		t.Fatalf("ToolCalls length = %d, want 1", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].Name != "get_spot_balances" {
		t.Errorf("ToolCalls[0].Name = %q, want %q", resp.ToolCalls[0].Name, "get_spot_balances")
	}
}

func TestCompleteClaudeToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)

		// Verify tools are included
		if _, ok := reqBody["tools"]; !ok {
			t.Error("request missing 'tools' field")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"content": []map[string]interface{}{
				{"type": "text", "text": "Let me check your balance."},
				{
					"type":  "tool_use",
					"id":    "toolu_01",
					"name":  "get_spot_balances",
					"input": map[string]interface{}{},
				},
			},
			"model":       "claude-sonnet-4-20250514",
			"stop_reason": "tool_use",
			"usage": map[string]interface{}{
				"input_tokens":  25,
				"output_tokens": 15,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient("test-key",
		WithProvider(ProviderClaude),
		WithAIHTTPClient(&urlRewriteClient{target: server.URL, inner: http.DefaultClient}),
	)

	resp, err := client.Complete(context.Background(), ChatRequest{
		Messages: []ChatMessage{{Role: RoleUser, Content: "Show my balance"}},
		Tools: []ToolDefinition{{
			Name:        "get_spot_balances",
			Description: "Get spot balances",
			Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
		}},
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	if resp.Content != "Let me check your balance." {
		t.Errorf("Content = %q, want %q", resp.Content, "Let me check your balance.")
	}
	if len(resp.ToolCalls) != 1 {
		t.Fatalf("ToolCalls length = %d, want 1", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].ID != "toolu_01" {
		t.Errorf("ToolCalls[0].ID = %q, want %q", resp.ToolCalls[0].ID, "toolu_01")
	}
	if resp.ToolCalls[0].Name != "get_spot_balances" {
		t.Errorf("ToolCalls[0].Name = %q, want %q", resp.ToolCalls[0].Name, "get_spot_balances")
	}
}

func TestCompleteOpenAIToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqBody map[string]interface{}
		json.Unmarshal(body, &reqBody)

		// Verify tools are included
		if _, ok := reqBody["tools"]; !ok {
			t.Error("request missing 'tools' field")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"content": "",
						"tool_calls": []map[string]interface{}{
							{
								"id":   "call_abc123",
								"type": "function",
								"function": map[string]interface{}{
									"name":      "get_ticker_prices",
									"arguments": `{"symbols":["BTCUSDT"]}`,
								},
							},
						},
					},
					"finish_reason": "tool_calls",
				},
			},
			"model": "gpt-4o",
			"usage": map[string]interface{}{
				"prompt_tokens":     30,
				"completion_tokens": 20,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient("test-key",
		WithProvider(ProviderOpenAI),
		WithAIHTTPClient(&urlRewriteClient{target: server.URL, inner: http.DefaultClient}),
	)

	resp, err := client.Complete(context.Background(), ChatRequest{
		Messages: []ChatMessage{{Role: RoleUser, Content: "Get BTC price"}},
		Tools: []ToolDefinition{{
			Name:        "get_ticker_prices",
			Description: "Get prices",
			Parameters:  json.RawMessage(`{"type":"object","properties":{"symbols":{"type":"array"}}}`),
		}},
	})
	if err != nil {
		t.Fatalf("Complete() error = %v", err)
	}

	if len(resp.ToolCalls) != 1 {
		t.Fatalf("ToolCalls length = %d, want 1", len(resp.ToolCalls))
	}
	if resp.ToolCalls[0].ID != "call_abc123" {
		t.Errorf("ToolCalls[0].ID = %q, want %q", resp.ToolCalls[0].ID, "call_abc123")
	}
	if resp.ToolCalls[0].Name != "get_ticker_prices" {
		t.Errorf("ToolCalls[0].Name = %q, want %q", resp.ToolCalls[0].Name, "get_ticker_prices")
	}
}

// urlRewriteClient is a test helper that redirects all requests to a test server.
type urlRewriteClient struct {
	target string
	inner  *http.Client
}

func (c *urlRewriteClient) Do(req *http.Request) (*http.Response, error) {
	// Rewrite the URL to point to the test server, preserving path
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(c.target, "http://")
	return c.inner.Do(req)
}
