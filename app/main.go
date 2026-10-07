package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/codecrafters-io/claude-code-starter-go/app/tools"
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

	messages := []openai.ChatCompletionMessageParamUnion{
		{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(prompt),
				},
			},
		},
	}

	for {
		resp, err := client.Chat.Completions.New(context.Background(),
			openai.ChatCompletionNewParams{
				Model:    model,
				Messages: messages,
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
					openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
						Name:        "Write",
						Description: openai.String("Write content to a file"),
						Parameters: shared.FunctionParameters{
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
					}),
					openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
						Name:        "Bash",
						Description: openai.String("Execute a shell command"),
						Parameters: shared.FunctionParameters{
							"type":     "object",
							"required": []string{"command"},
							"properties": map[string]any{
								"command": map[string]any{
									"type":        "string",
									"description": "The command to execute",
								},
							},
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

		tool_calls := resp.Choices[0].Message.ToolCalls
		msg := resp.Choices[0].Message

		if len(tool_calls) == 0 {
			fmt.Print(msg.Content)
			return
		}

		messages = append(messages, msg.ToParam())

		for _, tool := range tool_calls {
			var result string
			switch tool.Function.Name {
			case "Read":
				var params tools.ReadParameters

				err := json.Unmarshal([]byte(tool.Function.Arguments), &params)

				if err != nil {
					panic("Tool parameters parse error")
				}

				content, err := os.ReadFile(params.FilePath)

				if err != nil {
					result = fmt.Sprintf("error: %v", err)
				} else {
					result = string(content)
				}
			case "Write":
				var params tools.WriteParameters

				err := json.Unmarshal([]byte(tool.Function.Arguments), &params)

				if err != nil {
					panic("Tool parameters parse error")
				}

				data := []byte(params.Content)

				os.WriteFile(params.FilePath, data, 0644)

				if err != nil {
					panic("Write file error")
				}

			case "Bash":
				var params tools.BashParameters

				err := json.Unmarshal([]byte(tool.Function.Arguments), &params)

				if err != nil {
					panic("Tool parameters parse error")
				}

				command := exec.Command("sh", "-c", params.Command)

				stdout, err := command.Output()

				if err != nil {
					result = fmt.Sprintf("%s", err)
				}

				result = string(stdout)

			default:
				result = "error: unknown tool error " + tool.Function.Name
			}

			messages = append(messages, openai.ToolMessage(result, tool.ID))
		}
	}
}
