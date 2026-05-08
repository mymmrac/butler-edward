package memory

import (
	"encoding/json"
	"fmt"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/provider"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/storage"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool"
)

// ForgetTool represents a tool that allows to forget things.
type ForgetTool struct {
	storage storage.Storage
}

// NewForgetTool creates a new ForgetTool.
func NewForgetTool(storage storage.Storage) *ForgetTool {
	return &ForgetTool{storage: storage}
}

// Definition returns tool definition.
func (t *ForgetTool) Definition() provider.ToolDefinition {
	return provider.ToolDefinition{
		Type: provider.ToolTypeFunction,
		Function: &provider.ToolFunction{
			Name:        "forget",
			Description: "Allows forgetting things from the memory by keyword.",
			//nolint:goconst
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"keyword": map[string]any{
						"type":        "string",
						"description": "Short descriptive keyword to forget.",
						"pattern":     keywordPatter,
					},
				},
				"required": []string{"keyword"},
			},
		},
	}
}

// Call calls the tool.
func (t *ForgetTool) Call(ctx *tool.Context, args json.RawMessage) (*tool.Result, error) {
	var in struct {
		Keyword string `json:"keyword"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return tool.ErrorResult("Invalid arguments", fmt.Errorf("invalid args: %w", err))
	}

	err := t.storage.Set(ctx, ctx.UserID, KeyPrefix+in.Keyword, nil)
	if err != nil {
		return tool.ErrorResult("Failed to forget", fmt.Errorf("storage set: %w", err))
	}

	return tool.SuccessResult("Forgotten "+in.Keyword, fmt.Sprintf("Forgotten %q", in.Keyword))
}
