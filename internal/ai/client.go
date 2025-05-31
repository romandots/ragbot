package ai

import "os"

// AIClient обёртка над ModelStrategy
type AIClient struct {
	strategy ModelStrategy
}

// NewAIClient создаёт AIClient на основе переменных окружения
func NewAIClient() *AIClient {
	// useLocal берётся из ENV
	if os.Getenv("USE_LOCAL_MODEL") == "true" {
		return &AIClient{strategy: NewLocalStrategy()}
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	return &AIClient{strategy: NewGPTStrategy(apiKey)}
}

func (a *AIClient) GenerateEmbedding(text string) ([]float32, error) {
	return a.strategy.GenerateEmbedding(text)
}

func (a *AIClient) GenerateResponse(prompt string) (string, error) {
	return a.strategy.GenerateResponse(prompt)
}
