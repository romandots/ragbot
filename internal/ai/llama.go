package ai

import "fmt"

// LocalStrategy заглушка локальной LLaMA стратегии
// TODO: интегрировать llama.cpp или другую локальную модель
type LocalStrategy struct{}

func NewLocalStrategy() *LocalStrategy { return &LocalStrategy{} }

func (l *LocalStrategy) GenerateEmbedding(text string) ([]float32, error) {
	return nil, fmt.Errorf("local embedding not implemented")
}

func (l *LocalStrategy) GenerateResponse(prompt string) (string, error) {
	return "", fmt.Errorf("local generation not implemented")
}

func (l *LocalStrategy) TranscribeAudio(audioData []byte, format string) (string, error) {
	return "", fmt.Errorf("local transcription not implemented")
}
