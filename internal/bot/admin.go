package bot

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ragbot/internal/util"
)

// StartAdminBot запускает Telegram-бота для администрирования базы знаний.
// Параметры:
//   - db         : указатель на SQL-соединение
//   - token      : токен административного бота (из ENV)
//   - allowedIDs : слайс разрешённых chat_id администраторов
func StartAdminBot(db *sql.DB, token string, allowedIDs []int64) {
	defer util.Recover("StartAdminBot")
	bot, err := tgbotapi.NewBotAPI(token)
	for err != nil {
		log.Printf("Admin bot init error: %v\n", err)
		time.Sleep(1 * time.Second)
		log.Println("Trying to connect to Telegram bot API again...")
		bot, err = tgbotapi.NewBotAPI(token)
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

		// /help
		if strings.HasPrefix(text, "/help") {
			bot.Send(tgbotapi.NewMessage(chatID,
				"Команды администратора:\n"+
					"/help — эта справка\n"+
					"/myid — получить свой chat_id\n"+
					"/delete <id> — удалить фрагмент по ID\n"+
					"/update <id> <текст> — обновить фрагмент по ID\n\n"+
					"Все остальное будет интерпретировано как запись в базу знаний"))
			continue
		}

		// Все остальные команды должны выполняться только админами
		if !allowed[chatID] {
			continue
		}

		switch {
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
			_, _ = db.ExecContext(context.Background(),
				"UPDATE chunks SET content=$1, embedding=NULL, processed_at=NULL WHERE id=$2",
				content, id,
			)
			bot.Send(tgbotapi.NewMessage(chatID, fmt.Sprintf("Обновлён фрагмент %d", id)))

		default:
			content := strings.Trim(text, " ")
			_, _ = db.ExecContext(context.Background(),
				"INSERT INTO chunks(content) VALUES($1)",
				content,
			)
			bot.Send(tgbotapi.NewMessage(chatID, "Добавлено"))
		}
	}
}
