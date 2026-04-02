package services

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/pocky-ops-bot/internal/clients/ai"
	"github.com/pocky-ops-bot/internal/tools"
)

// mockAICompleter implements AICompleter for testing.
// Supports multiple responses for tool call loop testing.
type mockAICompleter struct {
	response  *ai.ChatResponse // single response (legacy)
	responses []*ai.ChatResponse // multiple responses for sequential calls
	err       error
	requests  []ai.ChatRequest
	callIdx   int
}

func (m *mockAICompleter) Complete(ctx context.Context, req ai.ChatRequest) (*ai.ChatResponse, error) {
	m.requests = append(m.requests, req)
	if m.err != nil {
		return nil, m.err
	}
	if len(m.responses) > 0 {
		idx := m.callIdx
		if idx >= len(m.responses) {
			idx = len(m.responses) - 1
		}
		m.callIdx++
		return m.responses[idx], nil
	}
	return m.response, nil
}

// mockToolExecutor implements ToolExecutor for testing.
type mockToolExecutor struct {
	definitions []ai.ToolDefinition
	results     map[string]tools.ToolResult // keyed by tool name
	calls       []ai.ToolCall
}

func (m *mockToolExecutor) Definitions() []ai.ToolDefinition {
	return m.definitions
}

func (m *mockToolExecutor) Execute(ctx context.Context, call ai.ToolCall) tools.ToolResult {
	m.calls = append(m.calls, call)
	if result, ok := m.results[call.Name]; ok {
		result.CallID = call.ID
		return result
	}
	return tools.ToolResult{CallID: call.ID, Content: "unknown tool", IsError: true}
}

func TestChatService_GenerateResponse(t *testing.T) {
	mock := &mockAICompleter{
		response: &ai.ChatResponse{
			Content:      "Hello! How can I help?",
			Model:        "gemini-2.0-flash",
			InputTokens:  10,
			OutputTokens: 8,
		},
	}

	service := NewChatService(mock, "You are Pocky", nil)

	reply, err := service.GenerateResponse(context.Background(), nil, "Hello")
	if err != nil {
		t.Fatalf("GenerateResponse() error = %v", err)
	}

	if reply != "Hello! How can I help?" {
		t.Errorf("reply = %q, want %q", reply, "Hello! How can I help?")
	}

	// Verify AI request
	if len(mock.requests) != 1 {
		t.Fatalf("expected 1 AI request, got %d", len(mock.requests))
	}

	req := mock.requests[0]
	if req.System != "You are Pocky" {
		t.Errorf("System = %q, want %q", req.System, "You are Pocky")
	}
	if len(req.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(req.Messages))
	}
	if req.Messages[0].Role != ai.RoleUser || req.Messages[0].Content != "Hello" {
		t.Errorf("message = %+v, want {user, Hello}", req.Messages[0])
	}
}

func TestChatService_GenerateResponse_WithHistory(t *testing.T) {
	mock := &mockAICompleter{
		response: &ai.ChatResponse{
			Content: "Your name is Alice.",
		},
	}

	service := NewChatService(mock, "", nil)

	history := []ai.ChatMessage{
		{Role: ai.RoleUser, Content: "My name is Alice"},
		{Role: ai.RoleAssistant, Content: "Nice to meet you, Alice!"},
	}

	_, err := service.GenerateResponse(context.Background(), history, "What is my name?")
	if err != nil {
		t.Fatalf("GenerateResponse() error = %v", err)
	}

	// Verify AI received full history + new message
	req := mock.requests[0]
	if len(req.Messages) != 3 {
		t.Fatalf("expected 3 messages (2 history + 1 new), got %d", len(req.Messages))
	}

	if req.Messages[0].Content != "My name is Alice" {
		t.Errorf("messages[0].Content = %q, want %q", req.Messages[0].Content, "My name is Alice")
	}
	if req.Messages[2].Content != "What is my name?" {
		t.Errorf("messages[2].Content = %q, want %q", req.Messages[2].Content, "What is my name?")
	}
}

func TestChatService_GenerateResponse_DoesNotMutateHistory(t *testing.T) {
	mock := &mockAICompleter{
		response: &ai.ChatResponse{Content: "reply"},
	}

	service := NewChatService(mock, "", nil)

	history := []ai.ChatMessage{
		{Role: ai.RoleUser, Content: "first"},
	}

	_, err := service.GenerateResponse(context.Background(), history, "second")
	if err != nil {
		t.Fatalf("GenerateResponse() error = %v", err)
	}

	// History should NOT be mutated by the service
	if len(history) != 1 {
		t.Errorf("history was mutated: len = %d, want 1", len(history))
	}
}

func TestChatService_GenerateResponse_AIError(t *testing.T) {
	mock := &mockAICompleter{
		err: fmt.Errorf("api error"),
	}

	service := NewChatService(mock, "", nil)

	_, err := service.GenerateResponse(context.Background(), nil, "Hello")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestChatService_GenerateResponse_WithToolsAttachesDefinitions(t *testing.T) {
	mock := &mockAICompleter{
		response: &ai.ChatResponse{Content: "No tools needed"},
	}

	tools := &mockToolExecutor{
		definitions: []ai.ToolDefinition{
			{Name: "test_tool", Description: "A test tool", Parameters: json.RawMessage(`{}`)},
		},
	}

	service := NewChatService(mock, "", nil, WithTools(tools))

	_, err := service.GenerateResponse(context.Background(), nil, "Hello")
	if err != nil {
		t.Fatalf("GenerateResponse() error = %v", err)
	}

	// Verify tool definitions were attached to the request
	if len(mock.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(mock.requests))
	}
	if len(mock.requests[0].Tools) != 1 {
		t.Errorf("expected 1 tool definition, got %d", len(mock.requests[0].Tools))
	}
	if mock.requests[0].Tools[0].Name != "test_tool" {
		t.Errorf("tool name = %q, want %q", mock.requests[0].Tools[0].Name, "test_tool")
	}
}

func TestChatService_GenerateResponse_ToolCallLoop(t *testing.T) {
	// Simulate: LLM calls tool, gets result, then returns final text
	aiMock := &mockAICompleter{
		responses: []*ai.ChatResponse{
			{
				Content: "",
				ToolCalls: []ai.ToolCall{
					{ID: "call_1", Name: "get_balance", Arguments: json.RawMessage(`{}`)},
				},
				InputTokens:  20,
				OutputTokens: 10,
			},
			{
				Content:      "Your balance is 100 USDT.",
				InputTokens:  50,
				OutputTokens: 15,
			},
		},
	}

	tools := &mockToolExecutor{
		definitions: []ai.ToolDefinition{
			{Name: "get_balance", Description: "Get balance"},
		},
		results: map[string]tools.ToolResult{
			"get_balance": {Content: `[{"asset":"USDT","free":"100.00"}]`},
		},
	}

	service := NewChatService(aiMock, "system prompt", nil, WithTools(tools))

	reply, err := service.GenerateResponse(context.Background(), nil, "Show balance")
	if err != nil {
		t.Fatalf("GenerateResponse() error = %v", err)
	}

	if reply != "Your balance is 100 USDT." {
		t.Errorf("reply = %q, want %q", reply, "Your balance is 100 USDT.")
	}

	// Verify tool was executed
	if len(tools.calls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(tools.calls))
	}
	if tools.calls[0].Name != "get_balance" {
		t.Errorf("tool call name = %q, want %q", tools.calls[0].Name, "get_balance")
	}

	// Verify 2 AI requests were made (initial + after tool result)
	if len(aiMock.requests) != 2 {
		t.Fatalf("expected 2 AI requests, got %d", len(aiMock.requests))
	}

	// Second request should include tool result messages
	secondReq := aiMock.requests[1]
	// Messages: user + assistant(tool_call) + tool(result)
	if len(secondReq.Messages) != 3 {
		t.Fatalf("second request messages = %d, want 3", len(secondReq.Messages))
	}
	if secondReq.Messages[1].Role != ai.RoleAssistant {
		t.Errorf("messages[1].Role = %q, want %q", secondReq.Messages[1].Role, ai.RoleAssistant)
	}
	if secondReq.Messages[2].Role != ai.RoleTool {
		t.Errorf("messages[2].Role = %q, want %q", secondReq.Messages[2].Role, ai.RoleTool)
	}
}

func TestChatService_GenerateResponse_MultipleToolCalls(t *testing.T) {
	// LLM calls two tools in one round, then returns final text
	aiMock := &mockAICompleter{
		responses: []*ai.ChatResponse{
			{
				ToolCalls: []ai.ToolCall{
					{ID: "call_1", Name: "get_balance", Arguments: json.RawMessage(`{}`)},
					{ID: "call_2", Name: "get_prices", Arguments: json.RawMessage(`{"symbols":["BTCUSDT"]}`)},
				},
			},
			{
				Content: "BTC balance: 0.5 ($50,000)",
			},
		},
	}

	tools := &mockToolExecutor{
		definitions: []ai.ToolDefinition{
			{Name: "get_balance", Description: "Get balance"},
			{Name: "get_prices", Description: "Get prices"},
		},
		results: map[string]tools.ToolResult{
			"get_balance": {Content: `[{"asset":"BTC","free":"0.5"}]`},
			"get_prices":  {Content: `[{"symbol":"BTCUSDT","price":"100000"}]`},
		},
	}

	service := NewChatService(aiMock, "", nil, WithTools(tools))

	reply, err := service.GenerateResponse(context.Background(), nil, "Portfolio")
	if err != nil {
		t.Fatalf("GenerateResponse() error = %v", err)
	}

	if reply != "BTC balance: 0.5 ($50,000)" {
		t.Errorf("reply = %q, want %q", reply, "BTC balance: 0.5 ($50,000)")
	}

	// Both tools should have been called
	if len(tools.calls) != 2 {
		t.Fatalf("expected 2 tool calls, got %d", len(tools.calls))
	}
}

func TestChatService_GenerateResponse_ToolCallLoopExceeded(t *testing.T) {
	// LLM keeps calling tools forever
	aiMock := &mockAICompleter{
		response: &ai.ChatResponse{
			ToolCalls: []ai.ToolCall{
				{ID: "call_loop", Name: "infinite", Arguments: json.RawMessage(`{}`)},
			},
		},
	}

	tools := &mockToolExecutor{
		definitions: []ai.ToolDefinition{{Name: "infinite"}},
		results: map[string]tools.ToolResult{
			"infinite": {Content: "looping"},
		},
	}

	service := NewChatService(aiMock, "", nil, WithTools(tools), WithMaxToolRounds(2))

	_, err := service.GenerateResponse(context.Background(), nil, "loop")
	if err == nil {
		t.Fatal("expected error for exceeded tool rounds")
	}
}

func TestChatService_GenerateResponse_ToolError(t *testing.T) {
	// Tool returns an error, LLM should still get the error and respond
	aiMock := &mockAICompleter{
		responses: []*ai.ChatResponse{
			{
				ToolCalls: []ai.ToolCall{
					{ID: "call_1", Name: "failing_tool", Arguments: json.RawMessage(`{}`)},
				},
			},
			{
				Content: "Sorry, the tool failed.",
			},
		},
	}

	tools := &mockToolExecutor{
		definitions: []ai.ToolDefinition{{Name: "failing_tool"}},
		results: map[string]tools.ToolResult{
			"failing_tool": {Content: "connection timeout", IsError: true},
		},
	}

	service := NewChatService(aiMock, "", nil, WithTools(tools))

	reply, err := service.GenerateResponse(context.Background(), nil, "do thing")
	if err != nil {
		t.Fatalf("GenerateResponse() error = %v", err)
	}

	if reply != "Sorry, the tool failed." {
		t.Errorf("reply = %q, want %q", reply, "Sorry, the tool failed.")
	}

	// Verify the error was passed to the LLM
	secondReq := aiMock.requests[1]
	toolMsg := secondReq.Messages[len(secondReq.Messages)-1]
	if toolMsg.Content != "connection timeout" {
		t.Errorf("tool result content = %q, want %q", toolMsg.Content, "connection timeout")
	}
}

func TestChatService_WithMaxToolRounds(t *testing.T) {
	service := NewChatService(&mockAICompleter{response: &ai.ChatResponse{Content: "ok"}}, "", nil, WithMaxToolRounds(10))
	if service.maxToolRounds != 10 {
		t.Errorf("maxToolRounds = %d, want 10", service.maxToolRounds)
	}
}

func TestChatService_DefaultMaxToolRounds(t *testing.T) {
	service := NewChatService(&mockAICompleter{response: &ai.ChatResponse{Content: "ok"}}, "", nil)
	if service.maxToolRounds != 5 {
		t.Errorf("maxToolRounds = %d, want 5", service.maxToolRounds)
	}
}
