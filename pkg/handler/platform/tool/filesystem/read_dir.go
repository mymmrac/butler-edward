package filesystem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/provider"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool"
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
			//nolint:goconst
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
func (t *ReadDirTool) Call(_ context.Context, args json.RawMessage) (*tool.Result, error) {
	var in struct {
		Path string `json:"path"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return tool.ErrorResult("Invalid arguments", fmt.Errorf("invalid args: %w", err))
	}

	dir, err := t.root.Open(in.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return tool.ErrorResult("Directory doesn't exist", fmt.Errorf("open dir: %w", err))
		}
		return tool.ErrorResult("Failed to open directory", fmt.Errorf("open dir: %w", err))
	}

	entries, err := dir.ReadDir(-1)
	if err != nil {
		return tool.ErrorResult("Failed to read directory", fmt.Errorf("read dir: %w", err))
	}

	if len(entries) == 0 {
		return tool.SuccessResult("No files found in directory", "No files found in: "+in.Path)
	}

	sb := &strings.Builder{}
	_, _ = fmt.Fprintf(sb, "Files in %s:\n", in.Path)
	for _, entry := range entries {
		_ = sb.WriteByte(fileTypeChar(entry))
		_ = sb.WriteByte(' ')
		_, _ = fmt.Fprint(sb, entry.Name())
		_ = sb.WriteByte('\n')
	}
	return tool.SuccessResult(fmt.Sprintf("Found %d files in directory", len(entries)), sb.String())
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
