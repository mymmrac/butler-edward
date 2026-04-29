package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mymmrac/butler-edward/pkg/module/platform/provider"
)

// ListFileTool represents a tool that lists files in the filesystem.
type ListFileTool struct {
	root *os.Root
}

// NewListFileTool creates a new ListFileTool.
func NewListFileTool(root *os.Root) *ListFileTool {
	return &ListFileTool{root: root}
}

// Definition returns tool definition.
func (t *ListFileTool) Definition() provider.ToolDefinition {
	return provider.ToolDefinition{
		Type: provider.ToolTypeFunction,
		Function: &provider.ToolFunction{
			Name:        "list_files",
			Description: "Lists files or directories for a given path.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{"type": "string", "description": "The path to list files"},
				},
				"required": []string{"path"},
			},
		},
	}
}

// Call reads the content of the file at the specified path.
func (t *ListFileTool) Call(_ context.Context, args json.RawMessage) (string, error) {
	var parsed struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &parsed); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	dir, err := t.root.Open(parsed.Path)
	if err != nil {
		return "", fmt.Errorf("open dir: %w", err)
	}

	entries, err := dir.ReadDir(-1)
	if err != nil {
		return "", fmt.Errorf("read dir: %w", err)
	}

	sb := &strings.Builder{}
	for _, entry := range entries {
		sb.WriteByte(fileTypeChar(entry))
		sb.WriteByte(' ')
		sb.WriteString(entry.Name())
		sb.WriteByte('\n')
	}
	return sb.String(), nil
}

func fileTypeChar(e os.DirEntry) byte {
	mode := e.Type()

	switch {
	case mode.IsDir():
		return 'd'
	case mode&os.ModeSymlink != 0:
		return 'l'
	case mode&os.ModeNamedPipe != 0:
		return 'p'
	case mode&os.ModeSocket != 0:
		return 's'
	case mode&os.ModeDevice != 0:
		if mode&os.ModeCharDevice != 0 {
			return 'c'
		}
		return 'b'
	default:
		return '-'
	}
}
