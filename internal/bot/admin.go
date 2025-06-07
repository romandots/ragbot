package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ragbot/internal/repository"
	"ragbot/internal/util"
)

const source = "admin"

var adminChats []int64
var adminBot *tgbotapi.BotAPI

// StartAdminBot launches Telegram bot for knowledge base administration.
func StartAdminBot(repo *repository.Repository, token string, allowedIDs []int64) {
	defer util.Recover("StartAdminBot")

	adminBot = connect(token)
	log.Println("Admin bot connected to Telegram API")

	adminChats = allowedIDs
	allowed := make(map[int64]bool)
	for _, id := range allowedIDs {
		allowed[id] = true
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := adminBot.GetUpdatesChan(u)

	log.Println("Admin bot started")
	for update := range updates {
		if update.Message == nil {
			continue
		}
		chatID := update.Message.Chat.ID
		text := strings.TrimSpace(update.Message.Text)

		if text == "/myid" {
			replyToAdmin(chatID, fmt.Sprintf("Ваш CHAT ID: %d", chatID))
			continue
		}

		if strings.HasPrefix(text, "/help") {
			replyToAdmin(chatID,
				"Команды администратора:\n"+
					"/help — эта справка\n"+
					"/myid — получить свой chat_id\n"+
					"/delete <id> — удалить фрагмент по ID\n"+
					"/update <id> <текст> — обновить фрагмент по ID\n\n"+
					"Все остальное будет интерпретировано как запись в базу знаний")
			continue
		}

		if !allowed[chatID] {
			continue
		}

		switch {
		case strings.HasPrefix(text, "/delete "):
			idStr := strings.TrimPrefix(text, "/delete ")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				replyToAdmin(chatID, "Неверный ID")
				continue
			}
			if err := repo.DeleteChunk(context.Background(), id); err != nil {
				replyToAdmin(chatID, "Ошибка удаления")
				continue
			}
			replyToAdmin(chatID, fmt.Sprintf("Удалён фрагмент %d", id))

		case strings.HasPrefix(text, "/update "):
			parts := strings.SplitN(strings.TrimPrefix(text, "/update "), " ", 2)
			if len(parts) < 2 {
				replyToAdmin(chatID, "Использование: /update <id> <новый текст>")
				continue
			}
			id, err := strconv.Atoi(parts[0])
			if err != nil {
				replyToAdmin(chatID, "Неверный ID")
				continue
			}
			content := parts[1]
			if err := repo.UpdateChunk(context.Background(), id, content); err != nil {
				replyToAdmin(chatID, "Ошибка обновления")
				continue
			}
			replyToAdmin(chatID, fmt.Sprintf("Обновлён фрагмент %d", id))

		default:
			content := strings.Trim(text, " ")
			added, err := repo.AddChunk(context.Background(), content, source)
			if err != nil {
				replyToAdmin(chatID, "Ошибка добавления")
				continue
			}
			if added {
				replyToAdmin(chatID, "Добавлено")
			} else {
				replyToAdmin(chatID, "Уже существует")
			}
		}
	}
}

func SendToAllAdmins(message string) {
	for _, adminChatID := range adminChats {
		replyToAdmin(adminChatID, message)
	}
}

func replyToAdmin(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	adminBot.Send(msg)
}
