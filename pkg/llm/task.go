package llm

import (
	"context"
	"fmt"
	"os"

	"github.com/sashabaranov/go-openai"
)

const (
	PRIMECODEGEN_LLM_ENDPOINT = "PRIMECODEGEN_LLM_ENDPOINT"
	PRIMECODEGEN_LLM_APIKEY   = "PRIMECODEGEN_LLM_APIKEY"
	PRIMECODEGEN_LLM_MODEL    = "PRIMECODEGEN_LLM_MODEL"
)

// LLMChatCompletion performs a simple chat completion using the LLM configured via environment variables.
func LLMChatCompletion(systemMessage string, userMessage string) (string, error) {
	// env
	endpoint := os.Getenv(PRIMECODEGEN_LLM_ENDPOINT)
	apiKey := os.Getenv(PRIMECODEGEN_LLM_APIKEY)
	model := os.Getenv(PRIMECODEGEN_LLM_MODEL)
	if endpoint == "" || apiKey == "" || model == "" {
		return "", fmt.Errorf("endpoint, apiKey or model not configured for LLM, please set %s, %s and %s", PRIMECODEGEN_LLM_ENDPOINT, PRIMECODEGEN_LLM_APIKEY, PRIMECODEGEN_LLM_MODEL)
	}

	// client
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = endpoint
	client := openai.NewClientWithConfig(config)

	// request
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemMessage,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: userMessage,
				},
			},
		},
	)
	if err != nil {
		return "", fmt.Errorf("llm chat completion error: %w", err)
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response from LLM")
	}
	return resp.Choices[0].Message.Content, nil
}
