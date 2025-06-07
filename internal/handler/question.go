package handler

import (
	"context"
	"fmt"

	"ragbot/internal/ai"
	"ragbot/internal/config"
	"ragbot/internal/conversation"
	"ragbot/internal/repository"
	"ragbot/internal/util"
)

// ProcessQuestionWithHistory builds prompt using conversation history and knowledge fragments.
func ProcessQuestionWithHistory(
	repo *repository.Repository,
	aiClient *ai.AIClient,
	chatID int64,
	question string,
) (string, error) {
	defer util.Recover("ProcessQuestionWithHistory")
	var histText string
	if chatID != 0 {
		history := conversation.GetHistory(repo, chatID)
		histText = "История беседы:\n"
		for _, item := range history {
			if item.Role == "user" {
				histText += "Пользователь: " + item.Content + "\n"
			} else if item.Role == "assistant" {
				histText += "Помощник: " + item.Content + "\n"
			}
		}
	}

	queryVec, err := aiClient.GenerateEmbedding(question)
	if err != nil {
		return "", err
	}

	fragments, err := repo.SearchChunks(context.Background(), queryVec, 5)
	if err != nil {
		return "", fmt.Errorf("DB query error: %v", err)
	}

	var fragText string
	fragText = "Используй фрагменты базы знаний:\n---\n"
	for _, c := range fragments {
		fragText += c + "\n"
	}
	fragText += "---\n"

	prompt := config.LoadSettings().Preamble + "\n" + histText + "\n" + fragText + "Вопрос: " + question + "\nОтвет:\n"

	fmt.Println("Prompt: " + prompt)
	return aiClient.GenerateResponse(prompt)
}
