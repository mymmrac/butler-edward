package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/provider"
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
			Description: "Reads the contents of a file given its path. Supports pagination with offset and max size.",
			//nolint:goconst
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path": map[string]any{
						"type":        "string",
						"description": "The path to the file.",
					},
					"offset": map[string]any{
						"type":        "integer",
						"description": "Byte offset to start reading from.",
						"default":     0,
					},
					"maxSize": map[string]any{
						"type":        "integer",
						"description": "Maximum number of bytes to read.",
						"default":     maxFileSize,
					},
				},
				"required": []string{"path"},
			},
		},
	}
}

const maxFileSize = 64 * 1024 // 64KB

// Call reads the content of the file at the specified path.
func (t *ReadFileTool) Call(_ context.Context, args json.RawMessage) (string, error) {
	var in struct {
		Path    string `json:"path"`
		Offset  int64  `json:"offset"`
		MaxSize int64  `json:"maxSize"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	if in.Offset < 0 {
		return "", fmt.Errorf("invalid argument: offset cannot be negative")
	}
	if in.MaxSize < 0 {
		return "", fmt.Errorf("invalid argument: max size cannot be negative")
	}

	if in.MaxSize > maxFileSize {
		in.MaxSize = maxFileSize
	}

	file, err := t.root.Open(in.Path)
	if err != nil {
		return "", fmt.Errorf("open file: %w", err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("stat file: %w", err)
	}

	if fileInfo.IsDir() {
		return "", fmt.Errorf("file is a directory")
	}

	size := fileInfo.Size()
	if size == 0 {
		return "", fmt.Errorf("file is empty")
	}
	if in.Offset >= size {
		return "", fmt.Errorf("offset is past the end of the file")
	}

	if in.Offset != 0 {
		_, err = file.Seek(in.Offset, 0)
		if err != nil {
			return "", fmt.Errorf("seek file: %w", err)
		}
	}

	sizeToRead := min(size-in.Offset, in.MaxSize)

	content := make([]byte, sizeToRead)
	if _, err = io.ReadFull(file, content); err != nil {
		return "", fmt.Errorf("read file: %w", err)
	}

	header := fmt.Sprintf("File: %s\nSize: %d bytes\nOffset: %d bytes\nRead %d bytes\n\n",
		filepath.Base(in.Path), size, in.Offset, sizeToRead,
	)

	return header + string(content), nil
}
