package provider

import (
	"context"
	"encoding/json"
)

// Provider representation.
type Provider interface {
	// Name returns provider name.
	Name() string

	// Models returns list of available models.
	Models(ctx context.Context) ([]Model, error)
	// Chat sends messages to the LLM provider.
	Chat(ctx context.Context, model string, messages []Message, tools []ToolDefinition) (*Response, error)
}

// Model representation.
type Model struct {
	// Name is the name of the model.
	Name string
}

// Message representation.
type Message struct {
	// Role is the role of the message sender.
	Role MessageRole
	// Name is the name of the message sender.
	Name string
	// Content is the content of the message.
	Content string
	// ToolCalls is the list of tool calls in the response.
	ToolCalls []ToolCall
	// ToolCallID is the ID of the tool call if the message is a tool call response.
	ToolCallID string
}

// MessageRole represents the role of a message sender.
type MessageRole string

const (
	// MessageRoleUser is the role of a user message.
	MessageRoleUser MessageRole = "user"
	// MessageRoleSystem is the role of a system message.
	MessageRoleSystem MessageRole = "system"
	// MessageRoleAssistant is the role of an assistant message.
	MessageRoleAssistant MessageRole = "assistant"
	// MessageRoleTool is the role of a tool message.
	MessageRoleTool MessageRole = "tool"
)

// ToolDefinition representation.
type ToolDefinition struct {
	// Type is the type of the tool.
	Type ToolType
	// Function is the function of the tool if Type is "function".
	Function *ToolFunction
}

// ToolType represents the type of tool.
type ToolType string

// ToolTypeFunction is the type of function tool.
const ToolTypeFunction ToolType = "function"

// ToolFunction representation.
type ToolFunction struct {
	// Name is the name of the function.
	Name string
	// Description is the description of the function.
	Description string
	// Parameters is the parameters of the function as JSON Schema.
	Parameters map[string]any
}

// Response representation.
type Response struct {
	// Content is the content of the response.
	Content string
	// ToolCalls is the list of tool calls in the response.
	ToolCalls []ToolCall
}

// ToolCall representation.
type ToolCall struct {
	// ID is the ID of the tool call.
	ID string
	// Type is the type of the tool call.
	Type ToolType
	// Function is the function of the tool call if Type is "function".
	Function *ToolFunctionCall
}

// ToolFunctionCall representation.
type ToolFunctionCall struct {
	// Name is the name of the function.
	Name string
	// Arguments is the arguments of the function as JSON.
	Arguments json.RawMessage
}
