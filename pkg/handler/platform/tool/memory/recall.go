package memory

import (
	"encoding/json"
	"fmt"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/provider"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/storage"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool"
)

// RecallTool represents a tool that allows to recall things.
type RecallTool struct {
	storage storage.Storage
}

// NewRecallTool creates a new RecallTool.
func NewRecallTool(storage storage.Storage) *RecallTool {
	return &RecallTool{storage: storage}
}

// Definition returns tool definition.
func (t *RecallTool) Definition() provider.ToolDefinition {
	return provider.ToolDefinition{
		Type: provider.ToolTypeFunction,
		Function: &provider.ToolFunction{
			Name: "recall",
			Description: "Allows recalling things from the memory by keyword. " +
				"Only already known things can be recalled. ",
			//nolint:goconst
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"keyword": map[string]any{
						"type":        "string",
						"description": "Short descriptive keyword to recall.",
						"pattern":     keywordPatter,
					},
				},
				"required": []string{"keyword"},
			},
		},
	}
}

// Call calls the tool.
func (t *RecallTool) Call(ctx *tool.Context, args json.RawMessage) (*tool.Result, error) {
	var in struct {
		Keyword string `json:"keyword"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return tool.ErrorResult("Invalid arguments", fmt.Errorf("invalid args: %w", err))
	}

	value, err := t.storage.Get(ctx, ctx.UserID, KeyPrefix+in.Keyword)
	if err != nil {
		return tool.ErrorResult("Failed to recall", fmt.Errorf("storage get: %w", err))
	}

	if value == nil {
		return tool.SuccessResult("No such thing", "No such thing: "+in.Keyword)
	}
	return tool.SuccessResult("Recalled "+in.Keyword, fmt.Sprintf("Recalled %q: %q", in.Keyword, value))
}
