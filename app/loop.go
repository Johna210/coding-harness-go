package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type Loop struct {
	client        openai.Client
	model         string
	messages      []openai.ChatCompletionMessageParamUnion
	tools         map[string]Tool
	skillLocation string
}

// Initialize the conversation
func Initialize(apiKey, baseUrl, model string) *Loop {
	loop := Loop{
		client:        openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl)),
		model:         model,
		tools:         make(map[string]Tool),
		skillLocation: ".claude/skills",
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

func (l *Loop) executeTool(toolCallID, toolName string, arguments string) error {
	tool, exists := l.tools[toolName]
	if !exists {
		return fmt.Errorf("tool not found: %s", toolName)
	}

	output, err := tool.Execution(arguments)
	if err != nil {
		return err
	}

	l.messages = append(l.messages, openai.ToolMessage(output, toolCallID))
	return nil
}

func (l *Loop) resolveSkills() error {
	contentList, err := ProcessSubdirectories(l.skillLocation, "SKILL.md")
	if err != nil {
		return err
	}
	l.messages = append(l.messages, openai.AssistantMessage(strings.Join(contentList, "\n")))
	return nil
}

func (l *Loop) invokeSkill(skillName string) string {
	path, err := FindDirectoryRecursive(l.skillLocation, skillName)
	if err != nil {
		return ""
	}

	filePath := path + "/SKILL.md"

	content, err := ReadAfterFirstBlankLine(filePath)
	if err != nil {
		return ""
	}

	return content
}

func (l *Loop) Run(prompt string) error {
	err := l.resolveSkills()
	if err != nil {
		return err
	}

	content := prompt

	if prompt[0] == '/' {
		content = l.invokeSkill(prompt[1:])
	}
	l.messages = []openai.ChatCompletionMessageParamUnion{openai.UserMessage(content)}

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
					err := l.executeTool(toolCall.ID, toolCall.Function.Name,
						toolCall.Function.Arguments)

					if err != nil {
						return err
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
