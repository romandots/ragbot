package handler

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pgvector/pgvector-go"
	"ragbot/internal/ai"
)

// ProcessQuestion выполняет RAG-логику и возвращает ответ
type QuestionHandler func(db *sql.DB, aiClient *ai.AIClient, question string) (string, error)

func ProcessQuestion(db *sql.DB, aiClient *ai.AIClient, question string) (string, error) {
	// 1) Эмбеддинг вопроса
	queryVec, err := aiClient.GenerateEmbedding(question)
	if err != nil {
		return "", err
	}

	// 2) Поиск похожих чанков в БД
	rows, err := db.QueryContext(context.Background(),
		`SELECT content FROM chunks ORDER BY embedding <-> $1 LIMIT 5`,
		pgvector.NewVector(queryVec),
	)
	if err != nil {
		return "", fmt.Errorf("DB query error: %v", err)
	}
	defer rows.Close()

	// Собираем найденные фрагменты
	var fragments []string
	for rows.Next() {
		var c string
		err := rows.Scan(&c)
		if err != nil {
			return "", fmt.Errorf("Row scan error: %v", err)
		}
		fragments = append(fragments, c)
	}

	// 3) Формируем prompt
	prompt := "Ты — помощник CRM. Используй фрагменты базы знаний:\n---\n"
	for _, c := range fragments {
		prompt += c + "\n"
	}
	prompt += "---\nВопрос: " + question + "\nОтвет:\n"

	// 4) Генерация ответа через AIClient
	return aiClient.GenerateResponse(prompt)
}
