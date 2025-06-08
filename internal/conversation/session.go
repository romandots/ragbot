package conversation

import (
	"context"
	"database/sql"
	"log"

	"ragbot/internal/repository"
)

type ChatInfo = repository.ChatInfo

func EnsureSession(repo *repository.Repository, chatID int64) (string, error) {
	uuid, err := repo.EnsureSession(context.Background(), chatID)
	if err != nil {
		log.Printf("ensure session error: %v", err)
		return "", err
	}
	return uuid, nil
}

func GetChatInfoByChatID(repo *repository.Repository, chatID int64) (ChatInfo, error) {
	return repo.GetChatInfoByChatID(context.Background(), chatID)
}

func GetChatInfoByUUID(repo *repository.Repository, uuid string) (ChatInfo, error) {
	return repo.GetChatInfoByUUID(context.Background(), uuid)
}

func UpdateSummary(repo *repository.Repository, chatID int64, summary, title, interest string) {
	if err := repo.UpdateSummary(context.Background(), chatID, summary, title, interest); err != nil {
		log.Printf("update summary error: %v", err)
	}
}

func UpdateName(repo *repository.Repository, chatID int64, name string) {
	if err := repo.UpdateName(context.Background(), chatID, name); err != nil {
		log.Printf("update name error: %v", err)
	}
}

func UpdatePhone(repo *repository.Repository, chatID int64, phone string) {
	if err := repo.UpdatePhone(context.Background(), chatID, phone); err != nil {
		log.Printf("update phone error: %v", err)
	}
}

func UpdateAmoContactID(repo *repository.Repository, chatID int64, contactID sql.NullInt64) {
	if err := repo.UpdateAmoContactID(context.Background(), chatID, contactID); err != nil {
		log.Printf("update amo contact id error: %v", err)
	}
}

func ClearAmoContactID(repo *repository.Repository, chatID int64) {
	UpdateAmoContactID(repo, chatID, sql.NullInt64{})
}

func GetFullHistory(repo *repository.Repository, chatID int64) []HistoryItem {
	items, err := repo.GetFullHistory(context.Background(), chatID)
	if err != nil {
		log.Printf("get full history query error: %v", err)
		return nil
	}
	return items
}
