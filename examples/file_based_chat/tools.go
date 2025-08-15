package file_based_chat

import (
	"context"
	"fmt"
	"os"

	"github.com/KennethanCeyer/adk-go/tools"
)

type ReadFileTool struct{}

func NewReadFileTool() tools.Tool {
	return &ReadFileTool{}
}

func (t *ReadFileTool) Name() string {
	return "read_file"
}

func (t *ReadFileTool) Description() string {
	return "Reads the content of a specified file."
}

func (t *ReadFileTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"filename": map[string]any{
				"type":        "string",
				"description": "The name of the file to read.",
			},
		},
		"required": []string{"filename"},
	}
}

func (t *ReadFileTool) Execute(ctx context.Context, args any) (any, error) {
	argsMap, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid arguments type, expected map[string]any")
	}
	filename, ok := argsMap["filename"].(string)
	if !ok {
		return nil, fmt.Errorf("filename is a required argument and must be a string")
	}

	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w", filename, err)
	}

	return map[string]any{
		"content": string(content),
	}, nil
}

type WriteFileTool struct{}

func NewWriteFileTool() tools.Tool {
	return &WriteFileTool{}
}

func (t *WriteFileTool) Name() string {
	return "write_file"
}

func (t *WriteFileTool) Description() string {
	return "Writes content to a specified file, overwriting it if it exists."
}

func (t *WriteFileTool) Parameters() any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"filename": map[string]any{
				"type":        "string",
				"description": "The name of the file to write to.",
			},
			"content": map[string]any{
				"type":        "string",
				"description": "The content to write to the file.",
			},
		},
		"required": []string{"filename", "content"},
	}
}

func (t *WriteFileTool) Execute(ctx context.Context, args any) (any, error) {
	argsMap, ok := args.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid arguments type, expected map[string]any")
	}
	filename, ok := argsMap["filename"].(string)
	if !ok {
		return nil, fmt.Errorf("filename is a required argument and must be a string")
	}
	content, ok := argsMap["content"].(string)
	if !ok {
		return nil, fmt.Errorf("content is a required argument and must be a string")
	}

	err := os.WriteFile(filename, []byte(content), 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to write to file '%s': %w", filename, err)
	}

	return map[string]any{
		"status":  "success",
		"message": fmt.Sprintf("Successfully wrote %d bytes to %s", len(content), filename),
	}, nil
}
