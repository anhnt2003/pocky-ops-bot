package tools

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"testing"

	"github.com/pocky-ops-bot/internal/clients/llm"
)

// mockTool implements Tool with configurable behavior.
type mockTool struct {
	def      llm.ToolDefinition
	result   string
	err      error
	called   bool
	lastArgs json.RawMessage
}

func (m *mockTool) Definition() llm.ToolDefinition {
	return m.def
}

func (m *mockTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	m.called = true
	m.lastArgs = arguments
	return m.result, m.err
}

func newMockTool(name, description, result string, err error) *mockTool {
	return &mockTool{
		def: llm.ToolDefinition{
			Name:        name,
			Description: description,
			Parameters:  json.RawMessage(`{"type":"object"}`),
		},
		result: result,
		err:    err,
	}
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry(nil)
	tool := newMockTool("test_tool", "A test tool", "ok", nil)

	r.Register(tool)

	got, ok := r.Get("test_tool")
	if !ok {
		t.Fatal("expected tool to be found after Register")
	}
	if got.Definition().Name != "test_tool" {
		t.Errorf("got name %q, want %q", got.Definition().Name, "test_tool")
	}
}

func TestRegistry_Register_Multiple(t *testing.T) {
	r := NewRegistry(nil)
	tools := []*mockTool{
		newMockTool("tool_a", "Tool A", "a", nil),
		newMockTool("tool_b", "Tool B", "b", nil),
		newMockTool("tool_c", "Tool C", "c", nil),
	}

	for _, tool := range tools {
		r.Register(tool)
	}

	for _, tool := range tools {
		name := tool.Definition().Name
		got, ok := r.Get(name)
		if !ok {
			t.Errorf("expected tool %q to be found", name)
			continue
		}
		if got.Definition().Name != name {
			t.Errorf("got name %q, want %q", got.Definition().Name, name)
		}
	}

	// Verify unknown tool is not found
	_, ok := r.Get("nonexistent")
	if ok {
		t.Error("expected nonexistent tool to not be found")
	}
}

func TestRegistry_Definitions(t *testing.T) {
	r := NewRegistry(nil)
	r.Register(newMockTool("tool_a", "Tool A", "", nil))
	r.Register(newMockTool("tool_b", "Tool B", "", nil))

	defs := r.Definitions()
	if len(defs) != 2 {
		t.Fatalf("expected 2 definitions, got %d", len(defs))
	}

	// Collect names (order is not guaranteed)
	names := make(map[string]bool)
	for _, d := range defs {
		names[d.Name] = true
	}

	if !names["tool_a"] {
		t.Error("expected tool_a in definitions")
	}
	if !names["tool_b"] {
		t.Error("expected tool_b in definitions")
	}
}

func TestRegistry_Execute_Success(t *testing.T) {
	r := NewRegistry(nil)
	tool := newMockTool("echo", "Echo tool", `{"echo":"hello"}`, nil)
	r.Register(tool)

	call := llm.ToolCall{
		ID:        "call-1",
		Name:      "echo",
		Arguments: json.RawMessage(`{"text":"hello"}`),
	}

	result := r.Execute(context.Background(), call)

	if result.CallID != "call-1" {
		t.Errorf("CallID = %q, want %q", result.CallID, "call-1")
	}
	if result.IsError {
		t.Errorf("IsError = true, want false")
	}
	if result.Content != `{"echo":"hello"}` {
		t.Errorf("Content = %q, want %q", result.Content, `{"echo":"hello"}`)
	}
	if !tool.called {
		t.Error("expected tool Execute to be called")
	}
}

func TestRegistry_Execute_Error(t *testing.T) {
	r := NewRegistry(nil)
	tool := newMockTool("fail", "Failing tool", "", errors.New("something broke"))
	r.Register(tool)

	call := llm.ToolCall{
		ID:        "call-2",
		Name:      "fail",
		Arguments: json.RawMessage(`{}`),
	}

	result := r.Execute(context.Background(), call)

	if result.CallID != "call-2" {
		t.Errorf("CallID = %q, want %q", result.CallID, "call-2")
	}
	if !result.IsError {
		t.Error("IsError = false, want true")
	}
	if result.Content != "something broke" {
		t.Errorf("Content = %q, want %q", result.Content, "something broke")
	}
}

func TestRegistry_Execute_UnknownTool(t *testing.T) {
	r := NewRegistry(nil)

	call := llm.ToolCall{
		ID:        "call-3",
		Name:      "nonexistent",
		Arguments: json.RawMessage(`{}`),
	}

	result := r.Execute(context.Background(), call)

	if result.CallID != "call-3" {
		t.Errorf("CallID = %q, want %q", result.CallID, "call-3")
	}
	if !result.IsError {
		t.Error("IsError = false, want true")
	}
	if result.Content != "unknown tool: nonexistent" {
		t.Errorf("Content = %q, want %q", result.Content, "unknown tool: nonexistent")
	}
}

func TestRegistry_Execute_Logging(t *testing.T) {
	logger := slog.Default()
	r := NewRegistry(logger)
	r.Register(newMockTool("logged", "Logged tool", "ok", nil))

	call := llm.ToolCall{
		ID:        "call-4",
		Name:      "logged",
		Arguments: json.RawMessage(`{}`),
	}

	result := r.Execute(context.Background(), call)
	if result.IsError {
		t.Errorf("unexpected error: %s", result.Content)
	}
	if result.Content != "ok" {
		t.Errorf("Content = %q, want %q", result.Content, "ok")
	}
}
