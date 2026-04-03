// Package services provides business logic for the bot.
package services

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/pocky-ops-bot/internal/clients/llm"
	"github.com/pocky-ops-bot/internal/tools"
)

// AICompleter is the interface for AI completion.
// Defined at the consumer side for testability.
type AICompleter interface {
	Complete(ctx context.Context, req llm.ChatRequest) (*llm.ChatResponse, error)
}

// ToolExecutor executes tool calls and provides tool definitions.
// Defined at the consumer side for testability.
type ToolExecutor interface {
	Definitions() []llm.ToolDefinition
	Execute(ctx context.Context, call llm.ToolCall) tools.ToolResult
}

// ChatService handles AI conversation generation.
// It is stateless — history is managed by the caller.
type ChatService struct {
	ai            AICompleter
	tools         ToolExecutor
	systemPrompt  string
	maxToolRounds int
	vietnamese    bool
	logger        *slog.Logger
}

// ChatServiceOption is a functional option for configuring ChatService.
type ChatServiceOption func(*ChatService)

// WithTools sets the tool executor for function calling support.
func WithTools(executor ToolExecutor) ChatServiceOption {
	return func(s *ChatService) {
		s.tools = executor
	}
}

// WithVietnamese forces the AI to always respond in Vietnamese.
func WithVietnamese() ChatServiceOption {
	return func(s *ChatService) {
		s.vietnamese = true
	}
}

// WithMaxToolRounds sets the maximum number of tool call rounds (safety limit).
func WithMaxToolRounds(n int) ChatServiceOption {
	return func(s *ChatService) {
		s.maxToolRounds = n
	}
}

// NewChatService creates a new ChatService.
func NewChatService(completer AICompleter, systemPrompt string, logger *slog.Logger, opts ...ChatServiceOption) *ChatService {
	if logger == nil {
		logger = slog.Default()
	}
	s := &ChatService{
		ai:            completer,
		systemPrompt:  systemPrompt,
		maxToolRounds: 5,
		logger:        logger,
	}
	for _, opt := range opts {
		opt(s)
	}

	// Append Vietnamese language instruction if enabled
	if s.vietnamese {
		s.systemPrompt += "\n\nAlways respond in Vietnamese (tiếng Việt)."
	}

	// Enhance system prompt with tool descriptions so the LLM knows to use them
	if s.tools != nil {
		defs := s.tools.Definitions()
		if len(defs) > 0 {
			var toolList strings.Builder; toolList.WriteString("\n\nYou have access to the following tools:\n")
			for _, def := range defs {
				fmt.Fprintf(&toolList, "- %s: %s\n", def.Name, def.Description)
			}
			toolList.WriteString("\nWhen the user asks about their portfolio, balance, prices, P&L, " +
				"futures positions, margin, leverage, open orders, trade history, or funding fees, " +
				"you MUST use these tools to fetch real-time data. " +
				"Do not make up or estimate values — always call the tools first.")
			s.systemPrompt += toolList.String()
		}
	}

	return s
}

// GenerateResponse calls the AI with the given history and user text.
// History is owned by the caller — this method does not store anything.
// If tools are configured, handles the tool call loop automatically.
func (s *ChatService) GenerateResponse(ctx context.Context, history []llm.ChatMessage, userText string) (string, error) {
	// Build messages: history + current user message
	messages := make([]llm.ChatMessage, len(history), len(history)+1)
	copy(messages, history)
	messages = append(messages, llm.ChatMessage{
		Role:    llm.RoleUser,
		Content: userText,
	})

	req := llm.ChatRequest{
		Messages: messages,
		System:   s.systemPrompt,
	}

	// Attach tool definitions if available
	if s.tools != nil {
		req.Tools = s.tools.Definitions()
	}

	// Tool call loop
	for round := 0; round <= s.maxToolRounds; round++ {
		resp, err := s.ai.Complete(ctx, req)
		if err != nil {
			return "", fmt.Errorf("ai completion failed: %w", err)
		}

		s.logger.Info("ai response",
			slog.Int("round", round),
			slog.Int("tool_calls", len(resp.ToolCalls)),
			slog.Int("input_tokens", resp.InputTokens),
			slog.Int("output_tokens", resp.OutputTokens),
		)

		// No tool calls → final response
		if len(resp.ToolCalls) == 0 {
			return resp.Content, nil
		}

		if s.tools == nil {
			// LLM returned tool calls but no executor configured
			return resp.Content, nil
		}

		// Append assistant message with tool calls
		req.Messages = append(req.Messages, llm.ChatMessage{
			Role:      llm.RoleAssistant,
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
		})

		// Execute each tool call and append results
		for _, call := range resp.ToolCalls {
			s.logger.Info("executing tool",
				slog.String("tool", call.Name),
				slog.String("call_id", call.ID),
			)

			result := s.tools.Execute(ctx, call)

			s.logger.Info("tool result",
				slog.String("tool", call.Name),
				slog.Bool("is_error", result.IsError),
				slog.Int("content_len", len(result.Content)),
			)

			req.Messages = append(req.Messages, llm.ChatMessage{
				Role:       llm.RoleTool,
				Content:    result.Content,
				ToolCallID: result.CallID,
			})
		}
	}

	return "", fmt.Errorf("tool call loop exceeded maximum rounds (%d)", s.maxToolRounds)
}
