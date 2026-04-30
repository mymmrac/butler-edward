package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/mymmrac/butler-edward/pkg/module/platform/provider"
)

// ReadDirTool represents a tool that lists files in the filesystem.
type ReadDirTool struct {
	root *os.Root
}

// NewReadDirTool creates a new ReadDirTool.
func NewReadDirTool(root *os.Root) *ReadDirTool {
	return &ReadDirTool{root: root}
}

// Definition returns tool definition.
func (t *ReadDirTool) Definition() provider.ToolDefinition {
	return provider.ToolDefinition{
		Type: provider.ToolTypeFunction,
		Function: &provider.ToolFunction{
			Name: "read_dir",
			Description: "Lists files and directories for a given path. " +
				"Returns a list of files in the format `{type} {name}`, " +
				"where the type is `-` for files and `d` for directories.",
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "The path to list files.",
					},
				},
				"required": []string{"path"},
			},
		},
	}
}

// Call reads the content of the file at the specified path.
func (t *ReadDirTool) Call(_ context.Context, args json.RawMessage) (string, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	dir, err := t.root.Open(in.Path)
	if err != nil {
		return "", fmt.Errorf("open dir: %w", err)
	}

	entries, err := dir.ReadDir(-1)
	if err != nil {
		return "", fmt.Errorf("read dir: %w", err)
	}

	if len(entries) == 0 {
		return "No files found in: " + in.Path, nil
	}

	sb := &strings.Builder{}
	_, _ = fmt.Fprintf(sb, "Files in %s:\n", in.Path)
	for _, entry := range entries {
		_ = sb.WriteByte(fileTypeChar(entry))
		_ = sb.WriteByte(' ')
		_, _ = fmt.Fprint(sb, entry.Name())
		_ = sb.WriteByte('\n')
	}
	return sb.String(), nil
}

func fileTypeChar(e os.DirEntry) byte {
	mode := e.Type()
	switch {
	case mode.IsDir():
		return 'd'
	default:
		return '-'
	}
}
