package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Loop struct {
	client   openai.Client
	model    string
	messages []openai.ChatCompletionMessageParamUnion
	tools    map[string]Tool
}

// Initialize the conversation
func Intialize(prompt, apiKey, baseUrl, model string) *Loop {
	loop := Loop{
		client:   openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl)),
		model:    model,
		messages: []openai.ChatCompletionMessageParamUnion{openai.UserMessage(prompt)},
		tools:    make(map[string]Tool),
	}

	return &loop
}

func (l *Loop) RegisterTool(tool Tool) {
	l.tools[tool.Name()] = tool
}

func (l *Loop) ToolDefinitions() []openai.ChatCompletionToolUnionParam {
	definitions := []openai.ChatCompletionToolUnionParam{}

	for _, tool := range l.tools {
		definitions = append(definitions, tool.Definition())
	}
	return definitions
}

func (l *Loop) Chat(prompt string) {
	resp, err := l.client.Chat.Completions.New(context.Background(),
		openai.ChatCompletionNewParams{
			Model:    l.model,
			Messages: l.messages,
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	assistantMessage := resp.Choices[0].Message.ToParam()
	l.messages = append(l.messages, assistantMessage)
}

func (l *Loop) executeTool(toolCallID, toolName, arguments string) error {
	tool := l.tools[toolName]

	output, err := tool.Execution(arguments)
	if err != nil {
		return err
	}

	l.messages = append(l.messages, openai.ToolMessage(output, toolCallID))
	return nil
}

func (l *Loop) Run() error {
	for {
		resp, err := l.client.Chat.Completions.New(context.Background(),
			openai.ChatCompletionNewParams{
				Model:    l.model,
				Messages: l.messages,
				Tools:    l.ToolDefinitions(),
			},
		)

		if err != nil {
			return err
		}

		if len(resp.Choices) == 0 {
			panic("No choices in response")
		}

		for _, choice := range resp.Choices {
			if choice.FinishReason == "tool_calls" {
				l.messages = append(l.messages, choice.Message.ToParam())
				for _, toolCall := range choice.Message.ToolCalls {
					if toolCall.Function.Name == "Read" {
						var arguments FileArgs
						err := json.Unmarshal([]byte(toolCall.Function.Arguments), &arguments)
						if err != nil {
							fmt.Fprintf(os.Stderr, "error: %v\n", err)
							os.Exit(1)
						}
						err = l.executeTool(toolCall.ID, "Read", arguments.FilePath)
						if err != nil {
							fmt.Fprintf(os.Stderr, "error: %v\n", err)
							os.Exit(1)
						}
						continue
					}
				}
			}
			if choice.FinishReason == "stop" {
				fmt.Print(choice.Message.Content)
				os.Exit(0)
			}
		}
	}
}
