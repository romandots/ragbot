package bot

import (
	"database/sql"
	"fmt"
	"log"
	"ragbot/internal/conversation"

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
		chatID := update.Message.Chat.ID
		userText := update.Message.Text

		// 1) Вызываем обработку вопроса (RAG + учёт истории)
		answer, err := handler.ProcessQuestionWithHistory(db, aiClient, chatID, userText)
		if err != nil {
			answer = fmt.Sprintf("Ошибка: %v", err)
		}

		// 2) Сохраняем вопрос пользователя в историю
		conversation.AppendHistory(chatID, "user", userText)

		// 3) Сохраняем ответ бота в историю
		conversation.AppendHistory(chatID, "assistant", answer)

		// 4) Отправляем ответ
		bot.Send(tgbotapi.NewMessage(chatID, answer))
	}
}
