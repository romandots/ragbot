package ai

import (
	"context"
	"fmt"
	go_openai "github.com/sashabaranov/go-openai"
)

// GPTStrategy использует OpenAI API
type GPTStrategy struct {
	client *go_openai.Client
}

// NewGPTStrategy создаёт GPTStrategy с заданным API-ключом
func NewGPTStrategy(apiKey string) *GPTStrategy {
	return &GPTStrategy{client: go_openai.NewClient(apiKey)}
}

func (g *GPTStrategy) GenerateEmbedding(text string) ([]float32, error) {
	resp, err := g.client.CreateEmbeddings(context.Background(), go_openai.EmbeddingRequest{
		Model: go_openai.AdaEmbeddingV2,
		Input: []string{text},
	})
	if err != nil {
		return nil, fmt.Errorf("OpenAI embedding error: %v", err)
	}
	return resp.Data[0].Embedding, nil
}

func (g *GPTStrategy) GenerateResponse(prompt string) (string, error) {
	chatReq := go_openai.ChatCompletionRequest{
		Model:       go_openai.GPT4oMini, //go_openai.GPT3Dot5Turbo,
		Messages:    []go_openai.ChatCompletionMessage{{Role: "system", Content: prompt}},
		MaxTokens:   512,
		Temperature: 0.2,
	}
	resp, err := g.client.CreateChatCompletion(context.Background(), chatReq)
	if err != nil {
		return "", fmt.Errorf("OpenAI chat error: %v", err)
	}
	return resp.Choices[0].Message.Content, nil
}
