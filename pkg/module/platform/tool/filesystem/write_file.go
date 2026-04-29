package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/mymmrac/butler-edward/pkg/module/platform/provider"
)

// WriteFileTool represents a tool that writes a file to the filesystem.
type WriteFileTool struct {
	root *os.Root
}

// NewWriteFileTool creates a new WriteFileTool.
func NewWriteFileTool(root *os.Root) *WriteFileTool {
	return &WriteFileTool{root: root}
}

// Definition returns the definition of the tool.
func (t *WriteFileTool) Definition() provider.ToolDefinition {
	return provider.ToolDefinition{
		Type: provider.ToolTypeFunction,
		Function: &provider.ToolFunction{
			Name:        "write_file",
			Description: "Writes the contents of a file given its path.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":    map[string]any{"type": "string", "description": "The path to the file"},
					"content": map[string]any{"type": "string", "description": "The content to write to the file"},
				},
				"required": []string{"path", "content"},
			},
		},
	}
}

// Call writes the content to the file at the specified path.
func (t *WriteFileTool) Call(_ context.Context, args json.RawMessage) (string, error) {
	var parsed struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(args, &parsed); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	err := t.root.WriteFile(parsed.Path, []byte(parsed.Content), 0o644)
	if err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return "Written successfully", nil
}
