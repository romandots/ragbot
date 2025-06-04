package conversation

import (
	"context"
	"database/sql"
	"log"
)

// HistoryItem — одна запись истории (пользователь или бот).
type HistoryItem struct {
	Role    string // "user" или "assistant"
	Content string
}

// AppendHistory добавляет новый элемент в историю конкретного чата в БД.
func AppendHistory(db *sql.DB, chatID int64, role, text string) {
	if db == nil {
		return
	}
	_, err := db.ExecContext(context.Background(),
		`INSERT INTO conversation_history(chat_id, role, content) VALUES ($1, $2, $3)`,
		chatID, role, text,
	)
	if err != nil {
		log.Printf("append history error: %v", err)
	}
}

// GetHistory возвращает последние 20 сообщений чата из БД.
func GetHistory(db *sql.DB, chatID int64) []HistoryItem {
	if db == nil {
		return nil
	}
	rows, err := db.QueryContext(context.Background(),
		`SELECT role, content FROM conversation_history
         WHERE chat_id=$1 ORDER BY id DESC LIMIT 20`, chatID)
	if err != nil {
		log.Printf("get history query error: %v", err)
		return nil
	}
	defer rows.Close()

	var items []HistoryItem
	for rows.Next() {
		var it HistoryItem
		if err := rows.Scan(&it.Role, &it.Content); err != nil {
			log.Printf("get history scan error: %v", err)
			return items
		}
		items = append(items, it)
	}
	// Результат идёт в обратном порядке, разворачиваем
	for i, j := 0, len(items)-1; i < j; i, j = i+1, j-1 {
		items[i], items[j] = items[j], items[i]
	}
	return items
}
