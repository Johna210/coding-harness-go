package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
)

type Tool interface {
	Name() string
	Definition() openai.ChatCompletionToolUnionParam
	Execution(rawArgs string) (string, error)
}

type ReadTool struct{}

func (rt ReadTool) Name() string {
	return "Read"
}

func (rt ReadTool) Definition() openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        "Read",
		Description: openai.String("Read and return the contents of a file"),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"file_path": map[string]any{
					"type":        "string",
					"description": "The path to the file to read",
				},
			},
			"required": []string{"file_path"},
		},
	})
}

func (rt ReadTool) Execution(rawArgs string) (string, error) {
	var arguments struct {
		FilePath string `json:"file_path"`
	}
	err := json.Unmarshal([]byte(rawArgs), &arguments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}

	output, err := ReadFile(arguments.FilePath)

	if err != nil {
		return "", err
	}

	return output, nil
}

type WriteTool struct{}

func (wt WriteTool) Name() string {
	return "Write"
}

func (wt WriteTool) Definition() openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        "Write",
		Description: openai.String("Write contents to a file"),
		Parameters: openai.FunctionParameters{
			"type":     "object",
			"required": []string{"file_path", "content"},
			"properties": map[string]any{
				"file_path": map[string]any{
					"type":        "string",
					"description": "The path of the file to write to",
				},
				"content": map[string]any{
					"type":        "string",
					"description": "The content to write to the file",
				},
			},
		},
	})
}

func (wt WriteTool) Execution(rawArgs string) (string, error) {
	var arguments struct {
		FilePath string `json:"file_path"`
		Content  string `json:"content"`
	}

	err := json.Unmarshal([]byte(rawArgs), &arguments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}

	err = WriteFile(arguments.FilePath, arguments.Content)
	if err != nil {
		return "", err
	}
	return "File written successfully", nil
}

type BashTool struct{}

func (bt BashTool) Name() string {
	return "Bash"
}

func (bt BashTool) Definition() openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
		Name:        "Bash",
		Description: openai.String("Execute a shell command"),
		Parameters: openai.FunctionParameters{
			"type":     "object",
			"required": []string{"command"},
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "The command to execute",
				},
			},
		},
	})
}

func (bt BashTool) Execution(rawArgs string) (string, error) {
	var arguments struct {
		Command string `json:"command"`
	}

	err := json.Unmarshal([]byte(rawArgs), &arguments)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}

	output, err := RunShellCommand(arguments.Command)
	if err != nil {
		return "", err
	}
	return output, nil
}
