package handler

import (
	"context"
	"database/sql"
	"fmt"
	"ragbot/internal/config"
	"ragbot/internal/conversation"

	"github.com/pgvector/pgvector-go"
	"ragbot/internal/ai"
)

// ProcessQuestionWithHistory собирает историю, фрагменты из chunks и формирует единый prompt.
func ProcessQuestionWithHistory(
	db *sql.DB,
	aiClient *ai.AIClient,
	chatID int64,
	question string,
) (string, error) {
	var histText string
	if chatID != 0 {
		// 1) Получаем всю историю сообщений для этого chatID
		history := conversation.GetHistory(db, chatID)

		// 2) Формируем блок истории в виде текста
		//    Например:
		//    "История беседы:\nПользователь: ...\nПомощник: ...\nПользователь: ...\n"
		histText = "История беседы:\n"
		for _, item := range history {
			if item.Role == "user" {
				histText += "Пользователь: " + item.Content + "\n"
			} else if item.Role == "assistant" {
				histText += "Помощник: " + item.Content + "\n"
			}
		}
	}

	// 3) Делаем эмбеддинг текущего вопроса через AIClient
	queryVec, err := aiClient.GenerateEmbedding(question)
	if err != nil {
		return "", err
	}

	// 4) Ищем фрагменты из chunks (top 5)
	rows, err := db.QueryContext(context.Background(),
		`SELECT content FROM chunks ORDER BY embedding <-> $1 LIMIT 5`,
		pgvector.NewVector(queryVec),
	)
	if err != nil {
		return "", fmt.Errorf("DB query error: %v", err)
	}
	defer rows.Close()

	var fragments []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return "", fmt.Errorf("Row scan error: %v", err)
		}
		fragments = append(fragments, c)
	}

	// 5) Формируем блок фрагментов
	var fragText string
	fragText = "Используй фрагменты базы знаний:\n---\n"
	for _, c := range fragments {
		fragText += c + "\n"
	}
	fragText += "---\n"

	// 6) Собираем окончательный prompt:
	//    <Preamble> +
	//    <histText> +
	//    <fragText> +
	//    "Вопрос: <question>\nОтвет:\n"
	prompt := config.LoadSettings().Preamble + "\n" + histText + "\n" + fragText + "Вопрос: " + question + "\nОтвет:\n"

	// 7) Генерируем ответ по полному prompt
	fmt.Println("Prompt: " + prompt)
	return aiClient.GenerateResponse(prompt)
}
