package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/shared"
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
	mode := os.Getenv("MODE")
	model := "anthropic/claude-haiku-4.5"

	if baseUrl == "" {
		baseUrl = "https://openrouter.ai/api/v1"
	}

	if mode == "local" {
		baseUrl = "http://localhost:11434/v1"
		model = "qwen3:8b"
	}

	if apiKey == "" && mode != "local" {
		panic("Env variable OPENROUTER_API_KEY not found")
	}

	client := openai.NewClient(option.WithAPIKey(apiKey), option.WithBaseURL(baseUrl))
	resp, err := client.Chat.Completions.New(context.Background(),
		openai.ChatCompletionNewParams{
			Model: model,
			Messages: []openai.ChatCompletionMessageParamUnion{
				{
					OfUser: &openai.ChatCompletionUserMessageParam{
						Content: openai.ChatCompletionUserMessageParamContentUnion{
							OfString: openai.String(prompt),
						},
					},
				},
			},
			Tools: []openai.ChatCompletionToolUnionParam{
				openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
					Name:        "Read",
					Description: openai.String("Read and return the content of the file"),
					Parameters: shared.FunctionParameters{
						"type": openai.String("object"),
						"properties": map[string]any{
							"file_path": map[string]any{
								"type":        "string",
								"description": "The path to the file to read",
							},
						},
						"required": []string{"file_path"},
					},
				}),
			},
		},
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	if len(resp.Choices) == 0 {
		panic("No choices in response")
	}

	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	fmt.Print(resp.Choices[0].Message.Content)
}
