package system

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/provider"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool"
)

// TimeTool represents a tool that returns the current time.
type TimeTool struct{}

// NewTimeTool creates a new TimeTool.
func NewTimeTool() *TimeTool {
	return &TimeTool{}
}

// Definition returns the definition of the tool.
func (t *TimeTool) Definition() provider.ToolDefinition {
	return provider.ToolDefinition{
		Type: provider.ToolTypeFunction,
		Function: &provider.ToolFunction{
			Name:        "get_time",
			Description: "Allows getting the current time with an optional timezone and format.",
			//nolint:goconst
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"format": map[string]any{
						"type":        "string",
						"description": "Golang time layout used to format the time (e.g. '2006-01-02T15:04:05Z07:00').",
						"default":     time.DateTime,
					},
					"timezone": map[string]any{
						"type":        "string",
						"description": "Timezone to use for the time (e.g. Eurepo/Kyiv).",
						"default":     time.Local.String(), //nolint:gosmopolitan
					},
				},
			},
		},
	}
}

// Call calls the tool.
func (t *TimeTool) Call(_ context.Context, args json.RawMessage) (*tool.Result, error) {
	var in struct {
		Format   string `json:"format"`
		Timezone string `json:"timezone"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return tool.ErrorResult("Invalid arguments", fmt.Errorf("invalid args: %w", err))
	}

	if in.Format == "" {
		in.Format = time.DateTime
	}
	if in.Timezone == "" {
		in.Timezone = time.Local.String() //nolint:gosmopolitan
	}

	loc, err := time.LoadLocation(in.Timezone)
	if err != nil {
		return tool.ErrorResult("Invalid timezone", fmt.Errorf("invalid timezone: %w", err))
	}

	return tool.SuccessResult("Got current time", "Time: "+time.Now().In(loc).Format(in.Format))
}
