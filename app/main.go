package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	var prompt string
	flag.StringVar(&prompt, "p", "", "Prompt to send to LLM")
	flag.Parse()

	if prompt == "" {
		panic("Prompt must not be empty")
	}

	apiKey := os.Getenv("OPENROUTER_API_KEY")
	baseUrl := os.Getenv("OPENROUTER_BASE_URL")
	if baseUrl == "" {
		baseUrl = "https://openrouter.ai/api/v1"
	}

	if apiKey == "" {
		panic("Env variable OPENROUTER_API_KEY not found")
	}

	loop := Initialize(apiKey, baseUrl, "anthropic/claude-haiku-4.5")

	loop.RegisterTool(&ReadTool{})
	loop.RegisterTool(&WriteTool{})
	loop.RegisterTool(&BashTool{})

	loop.Run(prompt)

	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")
}
