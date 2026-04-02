package tools

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/pocky-ops-bot/internal/clients/ai"
)

// Registry holds all available tools indexed by name.
type Registry struct {
	tools  map[string]Tool
	logger *slog.Logger
}

// NewRegistry creates a new Registry with the given logger.
// If logger is nil, slog.Default() is used.
func NewRegistry(logger *slog.Logger) *Registry {
	if logger == nil {
		logger = slog.Default()
	}
	return &Registry{
		tools:  make(map[string]Tool),
		logger: logger,
	}
}

// Register adds a tool to the registry indexed by its definition name.
func (r *Registry) Register(tool Tool) {
	name := tool.Definition().Name
	r.tools[name] = tool
	r.logger.Info("tool registered", slog.String("name", name))
}

// Definitions returns a slice of all tool definitions (for sending to the LLM).
func (r *Registry) Definitions() []ai.ToolDefinition {
	defs := make([]ai.ToolDefinition, 0, len(r.tools))
	for _, tool := range r.tools {
		defs = append(defs, tool.Definition())
	}
	return defs
}

// Get looks up a tool by name.
func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

// Execute runs a tool call and returns the result.
// If the tool is not found, it returns a ToolResult with IsError set to true.
func (r *Registry) Execute(ctx context.Context, call ai.ToolCall) ToolResult {
	tool, ok := r.tools[call.Name]
	if !ok {
		return ToolResult{
			CallID:  call.ID,
			IsError: true,
			Content: fmt.Sprintf("unknown tool: %s", call.Name),
		}
	}

	r.logger.Info("executing tool",
		slog.String("name", call.Name),
		slog.String("call_id", call.ID),
	)

	start := time.Now()
	result, err := tool.Execute(ctx, call.Arguments)
	duration := time.Since(start)

	if err != nil {
		r.logger.Error("tool execution failed",
			slog.String("name", call.Name),
			slog.String("error", err.Error()),
			slog.Duration("duration", duration),
		)
		return ToolResult{
			CallID:  call.ID,
			IsError: true,
			Content: err.Error(),
		}
	}

	r.logger.Info("tool execution completed",
		slog.String("name", call.Name),
		slog.Duration("duration", duration),
		slog.Int("content_length", len(result)),
	)

	return ToolResult{
		CallID:  call.ID,
		IsError: false,
		Content: result,
	}
}
