package bot

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pgvector/pgvector-go"
	ai "ragbot/internal/ai"
)

// StartAdminBot запускает Telegram-бота для администрирования базы знаний.
// Параметры:
//   - db            : указатель на SQL-соединение
//   - aiClient      : экземпляр AIClient для генерации эмбеддингов
//   - token         : токен административного бота (из ENV)
//   - allowedIDs    : слайс разрешённых chat_id администраторов
func StartAdminBot(db *sql.DB, aiClient *ai.AIClient, token string, allowedIDs []int64) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Admin bot init error: %v", err)
	}

	// Собираем множество разрешённых ChatID
	allowed := make(map[int64]bool)
	for _, id := range allowedIDs {
		allowed[id] = true
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Println("Admin bot started")
	for update := range updates {
		if update.Message == nil {
			continue
		}
		chatID := update.Message.Chat.ID
		text := strings.TrimSpace(update.Message.Text)

		// Команда /myid доступна всем пользователям
		if text == "/myid" {
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Your chat_id is: %d", chatID)))
			continue
		}

		// Все остальные команды должны выполняться только админами
		if !allowed[chatID] {
			continue
		}

		switch {
		// /add <текст> — добавить новый фрагмент в chunks
		case strings.HasPrefix(text, "/add "):
			content := strings.TrimPrefix(text, "/add ")
			// Для добавления фрагмента нужен эмбеддинг через aiClient
			emb, err := aiClient.GenerateEmbedding(content)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Embedding error: %v", err)))
				continue
			}
			_, _ = db.ExecContext(context.Background(),
				"INSERT INTO chunks(content, embedding) VALUES($1, $2)",
				content, pgvector.NewVector(emb),
			)
			bot.Send(tgbotapi.NewMessage(chatID, "Добавлено"))

		// /delete <id> — удалить фрагмент по id
		case strings.HasPrefix(text, "/delete "):
			idStr := strings.TrimPrefix(text, "/delete ")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Неверный ID"))
				continue
			}
			_, _ = db.ExecContext(context.Background(),
				"DELETE FROM chunks WHERE id = $1", id,
			)
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Удалён фрагмент %d", id)))

		// /update <id> <новый текст> — обновить существующий фрагмент
		case strings.HasPrefix(text, "/update "):
			parts := strings.SplitN(strings.TrimPrefix(text, "/update "), " ", 2)
			if len(parts) < 2 {
				bot.Send(tgbotapi.NewMessage(chatID, "Использование: /update <id> <новый текст>"))
				continue
			}
			id, err := strconv.Atoi(parts[0])
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, "Неверный ID"))
				continue
			}
			content := parts[1]
			emb, err := aiClient.GenerateEmbedding(content)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Embedding error: %v", err)))
				continue
			}
			_, _ = db.ExecContext(context.Background(),
				"UPDATE chunks SET content=$1, embedding=$2 WHERE id=$3",
				content, pgvector.NewVector(emb), id,
			)
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Обновлён фрагмент %d", id)))

		default:
			bot.Send(tgbotapi.NewMessage(chatID,
				"Команды администратора:\n"+
					"/myid — получить свой chat_id\n"+
					"/add <текст> — добавить фрагмент\n"+
					"/delete <id> — удалить фрагмент по ID\n"+
					"/update <id> <текст> — обновить фрагмент по ID"))
		}
	}
}
