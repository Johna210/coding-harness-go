package main

import "github.com/openai/openai-go/v3"

type Tool interface {
	Name() string
	Definition() openai.ChatCompletionToolUnionParam
	Execution(arguments string) (string, error)
}

type ReadTool struct{}

func (rt ReadTool) Name() string {
	return "Read"
}

func (rt *ReadTool) Definition() openai.ChatCompletionToolUnionParam {
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

func (rt *ReadTool) Execution(arguments string) (string, error) {
	output, err := ReadFile(arguments)

	if err != nil {
		return "", err
	}

	return output, nil
}
