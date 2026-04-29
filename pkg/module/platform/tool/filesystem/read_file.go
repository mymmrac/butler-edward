package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mymmrac/butler-edward/pkg/module/platform/provider"
)

// ReadFileTool represents a tool that reads a file from the filesystem.
type ReadFileTool struct {
	root *os.Root
}

// NewReadFileTool creates a new ReadFileTool.
func NewReadFileTool(root *os.Root) *ReadFileTool {
	return &ReadFileTool{root: root}
}

// Definition returns tool definition.
func (t *ReadFileTool) Definition() provider.ToolDefinition {
	return provider.ToolDefinition{
		Type: provider.ToolTypeFunction,
		Function: &provider.ToolFunction{
			Name:        "read_file",
			Description: "Reads the contents of a file given its path.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{"type": "string", "description": "The path to the file"},
				},
				"required": []string{"path"},
			},
		},
	}
}

// Call reads the content of the file at the specified path.
func (t *ReadFileTool) Call(_ context.Context, args json.RawMessage) (string, error) {
	var parsed struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &parsed); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	content, err := t.root.ReadFile(parsed.Path)
	if err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	return string(content), nil
}
