package ai

import (
	"ragbot/internal/config"
)

type ModelStrategy interface {
	GenerateEmbedding(text string) ([]float32, error)
	GenerateResponse(prompt string) (string, error)
	TranscribeAudio(audioData []byte, format string) (string, error)
}

type AIClient struct {
	strategy ModelStrategy
}

func NewAIClient() *AIClient {
	if config.Config.UseLocalModel {
		return &AIClient{strategy: NewLocalStrategy()}
	}
	return &AIClient{strategy: NewGPTStrategy(config.Config.OpenAIAPIKey)}
}

func (a *AIClient) GenerateEmbedding(text string) ([]float32, error) {
	return a.strategy.GenerateEmbedding(text)
}

func (a *AIClient) GenerateResponse(prompt string) (string, error) {
	return a.strategy.GenerateResponse(prompt)
}

func (a *AIClient) TranscribeAudio(audioData []byte, format string) (string, error) {
	return a.strategy.TranscribeAudio(audioData, format)
}
