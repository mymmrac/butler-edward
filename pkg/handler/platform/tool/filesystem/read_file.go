package filesystem

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/provider"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool"
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
func (t *ReadFileTool) Call(_ context.Context, args json.RawMessage) (*tool.Result, error) {
	var in struct {
		Path    string `json:"path"`
		Offset  int64  `json:"offset"`
		MaxSize int64  `json:"maxSize"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return tool.ErrorResult("Invalid arguments", fmt.Errorf("invalid args: %w", err))
	}

	if in.Offset < 0 {
		return tool.ErrorResult("Invalid arguments", fmt.Errorf("invalid args: offset cannot be negative"))
	}
	if in.MaxSize < 0 {
		return tool.ErrorResult("Invalid arguments", fmt.Errorf("invalid args: max size cannot be negative"))
	}

	if in.MaxSize > maxFileSize {
		in.MaxSize = maxFileSize
	}

	file, err := t.root.Open(in.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return tool.ErrorResult("File doesn't exist", fmt.Errorf("open file: %w", err))
		}
		return tool.ErrorResult("Failed to open file", fmt.Errorf("open file: %w", err))
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return tool.ErrorResult("Failed to stat file", fmt.Errorf("stat file: %w", err))
	}

	if fileInfo.IsDir() {
		return tool.ErrorResult("File is a directory", fmt.Errorf("file is a directory"))
	}

	size := fileInfo.Size()
	if size == 0 {
		return tool.SuccessResult("File is empty", "File is empty: "+in.Path)
	}
	if in.Offset >= size {
		return tool.ErrorResult("Offset is past the end of the file", fmt.Errorf("offset is past the end of the file"))
	}

	if in.Offset != 0 {
		_, err = file.Seek(in.Offset, 0)
		if err != nil {
			return tool.ErrorResult("Failed to seek file", fmt.Errorf("seek file: %w", err))
		}
	}

	sizeToRead := min(size-in.Offset, in.MaxSize)

	content := make([]byte, sizeToRead)
	if _, err = io.ReadFull(file, content); err != nil {
		return tool.ErrorResult("Failed to read file", fmt.Errorf("read file: %w", err))
	}

	header := fmt.Sprintf("File: %s\nSize: %d bytes\nOffset: %d bytes\nRead %d bytes\n\n",
		filepath.Base(in.Path), size, in.Offset, sizeToRead,
	)

	return tool.SuccessResult(fmt.Sprintf("Read %d bytes from file", sizeToRead), header+string(content))
}
