package bot

import (
	"database/sql"
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	ai "ragbot/internal/ai"
	"ragbot/internal/handler" // поправлен импорт
)

// StartUserBot запускает Telegram-бота для пользователей.
// Параметры:
//   - db       : указатель на SQL-соединение
//   - aiClient : экземпляр AIClient для генерации ответов
//   - token    : токен пользовательского бота (из ENV)
func StartUserBot(db *sql.DB, aiClient *ai.AIClient, token string) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("User bot init error: %v", err)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Println("User bot started")
	for update := range updates {
		if update.Message == nil {
			continue
		}
		question := update.Message.Text
		// Вызовем handler.ProcessQuestion (раньше было handler.ProcessQuestion, но handler не был импортирован)
		answer, err := handler.ProcessQuestion(db, aiClient, question)
		msgText := answer
		if err != nil {
			msgText = fmt.Sprintf("Ошибка: %v", err)
		}
		bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msgText))
	}
}
