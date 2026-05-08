package filesystem

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/provider"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool"
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
			Description: "Writes the contents of a file given its path. Supports actions: overwrite and append.",
			//nolint:goconst
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "The path to the file.",
					},
					"content": map[string]any{
						"type":        "string",
						"description": "The content to write to the file.",
					},
					"create": map[string]any{
						"type":        "boolean",
						"description": "Whether to create the file if it doesn't exist.",
						"default":     true,
					},
					"action": map[string]any{
						"type":        "string",
						"description": "The action to perform on the file.",
						"enum":        []string{"overwrite", "append"},
						"default":     "overwrite",
					},
				},
				"required": []string{"path", "content"},
			},
		},
	}
}

// Call writes the content to the file at the specified path.
func (t *WriteFileTool) Call(_ *tool.Context, args json.RawMessage) (*tool.Result, error) {
	var in struct {
		Path    string `json:"path"`
		Content string `json:"content"`
		Create  bool   `json:"create"`
		Action  string `json:"action"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return tool.ErrorResult("Invalid arguments", fmt.Errorf("invalid args: %w", err))
	}

	flags := os.O_WRONLY
	if in.Create {
		if err := t.root.MkdirAll(filepath.Dir(in.Path), 0o755); err != nil {
			return tool.ErrorResult("Failed to create directory", fmt.Errorf("create dir: %w", err))
		}
		flags |= os.O_CREATE
	}

	switch in.Action {
	case "overwrite":
		flags |= os.O_TRUNC
	case "append":
		flags |= os.O_APPEND
	default:
		return tool.ErrorResult("Invalid action", fmt.Errorf("invalid action: %q", in.Action))
	}

	file, err := t.root.OpenFile(in.Path, flags, 0o644)
	if err != nil {
		return tool.ErrorResult("Failed to open file", fmt.Errorf("open file: %w", err))
	}
	defer func() { _ = file.Close() }()

	_, err = file.WriteString(in.Content)
	if err != nil {
		return tool.ErrorResult("Failed to write file", fmt.Errorf("write file: %w", err))
	}

	return tool.SuccessResult("File written", "File written: "+in.Path)
}
