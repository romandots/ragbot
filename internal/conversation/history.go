package conversation

import (
	"context"
	"log"

	"ragbot/internal/repository"
)

type HistoryItem = repository.HistoryItem

func AppendHistory(repo *repository.Repository, chatID int64, role, text string) {
	if err := repo.AppendHistory(context.Background(), chatID, role, text); err != nil {
		log.Printf("append history error: %v", err)
	}
}

func GetHistory(repo *repository.Repository, chatID int64) []HistoryItem {
	items, err := repo.GetHistory(context.Background(), chatID, 20)
	if err != nil {
		log.Printf("get history query error: %v", err)
		return nil
	}
	return items
}
