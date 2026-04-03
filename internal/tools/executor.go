package tools

import (
	"context"
	"encoding/json"

	"github.com/pocky-ops-bot/internal/clients/llm"
)

// Tool is the interface each concrete tool must implement.
type Tool interface {
	// Definition returns the tool's metadata for the LLM.
	Definition() llm.ToolDefinition

	// Execute runs the tool with the given JSON arguments.
	// Returns a JSON string on success, or an error.
	Execute(ctx context.Context, arguments json.RawMessage) (string, error)
}
