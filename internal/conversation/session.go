package conversation

import (
	"context"
	"database/sql"
	"log"
)

type ChatInfo struct {
	ChatID  int64
	Summary sql.NullString
}

// EnsureSession makes sure a record exists for the given chatID and returns its uuid.
func EnsureSession(db *sql.DB, chatID int64) (string, error) {
	if db == nil {
		return "", nil
	}
	var uuid string
	err := db.QueryRowContext(context.Background(), `SELECT uuid FROM conversations WHERE chat_id=$1`, chatID).Scan(&uuid)
	if err == sql.ErrNoRows {
		err = db.QueryRowContext(context.Background(), `INSERT INTO conversations(chat_id) VALUES($1) RETURNING uuid`, chatID).Scan(&uuid)
	}
	if err != nil {
		log.Printf("ensure session error: %v", err)
		return "", err
	}
	return uuid, nil
}

// GetChatInfoByUUID returns chat id and summary by uuid.
func GetChatInfoByUUID(db *sql.DB, uuid string) (ChatInfo, error) {
	var info ChatInfo
	if db == nil {
		return info, sql.ErrConnDone
	}
	err := db.QueryRowContext(context.Background(), `SELECT chat_id, summary FROM conversations WHERE uuid=$1`, uuid).Scan(&info.ChatID, &info.Summary)
	if err != nil {
		return info, err
	}
	return info, nil
}

// UpdateSummary saves summary for chat.
func UpdateSummary(db *sql.DB, chatID int64, summary string) {
	if db == nil {
		return
	}
	_, err := db.ExecContext(context.Background(), `UPDATE conversations SET summary=$1, updated_at=NOW() WHERE chat_id=$2`, summary, chatID)
	if err != nil {
		log.Printf("update summary error: %v", err)
	}
}

// GetFullHistory returns entire chat history in chronological order.
func GetFullHistory(db *sql.DB, chatID int64) []HistoryItem {
	if db == nil {
		return nil
	}
	rows, err := db.QueryContext(context.Background(), `SELECT role, content FROM conversation_history WHERE chat_id=$1 ORDER BY id ASC`, chatID)
	if err != nil {
		log.Printf("get full history query error: %v", err)
		return nil
	}
	defer rows.Close()

	var items []HistoryItem
	for rows.Next() {
		var it HistoryItem
		if err := rows.Scan(&it.Role, &it.Content); err != nil {
			log.Printf("get full history scan error: %v", err)
			return items
		}
		items = append(items, it)
	}
	return items
}
