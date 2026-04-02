package tools

// ToolResult is the outcome of executing a tool call.
type ToolResult struct {
	CallID  string `json:"call_id"`
	Content string `json:"content"` // JSON for success, error message for failure
	IsError bool   `json:"is_error"`
}
