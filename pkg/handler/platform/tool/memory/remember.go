package memory

import (
	"encoding/json"
	"fmt"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/provider"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/storage"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool"
)

// RememberTool represents a tool that allows to remember things.
type RememberTool struct {
	storage storage.Storage
}

// NewRememberTool creates a new RememberTool.
func NewRememberTool(storage storage.Storage) *RememberTool {
	return &RememberTool{storage: storage}
}

// Definition returns tool definition.
func (t *RememberTool) Definition() provider.ToolDefinition {
	return provider.ToolDefinition{
		Type: provider.ToolTypeFunction,
		Function: &provider.ToolFunction{
			Name:        "remember",
			Description: "Allows adding new things to the memory by keyword.",
			//nolint:goconst
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"keyword": map[string]any{
						"type": "string",
						"description": "Short descriptive keyword to remember the thing by. " +
							"Keyword should be lowercase and contain only letters, numbers, and dashes.",
						"pattern": keywordPatter,
					},
					"thing": map[string]any{
						"type":        "string",
						"description": "Thing to remember.",
					},
				},
				"required": []string{"keyword", "thing"},
			},
		},
	}
}

// Call calls the tool.
func (t *RememberTool) Call(ctx *tool.Context, args json.RawMessage) (*tool.Result, error) {
	var in struct {
		Keyword string `json:"keyword"`
		Thing   string `json:"thing"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return tool.ErrorResult("Invalid arguments", fmt.Errorf("invalid args: %w", err))
	}

	err := t.storage.Set(ctx, ctx.UserID, KeyPrefix+in.Keyword, []byte(in.Thing))
	if err != nil {
		return tool.ErrorResult("Failed to remember", fmt.Errorf("storage set: %w", err))
	}

	return tool.SuccessResult("Remembered "+in.Keyword, fmt.Sprintf("Remembered %q: %q", in.Keyword, in.Thing))
}
