// Package ai provides provider-agnostic type definitions for AI completion APIs.
package ai

import "encoding/json"

// Provider represents the AI service provider.
type Provider string

const (
	ProviderGemini Provider = "gemini"
	ProviderClaude Provider = "claude"
	ProviderOpenAI Provider = "openai"
	ProviderQwen   Provider = "qwen"
)

// Role represents the role of a message participant.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// ToolDefinition describes a tool the LLM can call.
type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"` // JSON Schema object
}

// ToolCall represents a tool invocation requested by the LLM.
type ToolCall struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

// ChatMessage represents a single message in a conversation.
type ChatMessage struct {
	Role    Role   `json:"role"`
	Content string `json:"content"`

	// ToolCalls is populated when Role=assistant and the LLM wants to call tools.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// ToolCallID identifies which tool call this message is a result for (Role=tool).
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// ChatRequest represents a request to the AI completion API.
type ChatRequest struct {
	Model       string           `json:"model,omitempty"`
	Messages    []ChatMessage    `json:"messages"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
	System      string           `json:"system,omitempty"`
	Tools       []ToolDefinition `json:"tools,omitempty"`
}

// ChatResponse represents the AI completion response.
type ChatResponse struct {
	Content      string
	Model        string
	InputTokens  int
	OutputTokens int

	// ToolCalls is non-nil when the LLM wants to call tools instead of (or in addition to) text.
	ToolCalls []ToolCall
}
