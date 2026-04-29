package tool

import (
	"context"
	"encoding/json"

	"github.com/mymmrac/butler-edward/pkg/module/platform/provider"
)

// Tool representation.
type Tool interface {
	// Definition returns tool definition.
	Definition() provider.ToolDefinition

	// Call calls the tool.
	Call(ctx context.Context, args json.RawMessage) (string, error)
}
