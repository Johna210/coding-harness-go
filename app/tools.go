package main

import "github.com/openai/openai-go/v3"

type Tool interface {
	Name() string
	Definition() openai.ChatCompletionToolUnionParam
	Execution(arguments FileArgs) (string, error)
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

func (rt ReadTool) Execution(arguments FileArgs) (string, error) {
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

func (wt WriteTool) Execution(arguments FileArgs) (string, error) {
	err := WriteFile(arguments.FilePath, *arguments.Content)
	if err != nil {
		return "", err
	}
	return "File written successfully", nil
}
