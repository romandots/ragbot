package bot

import (
	"database/sql"
	"fmt"
	"log"
	"ragbot/internal/config"
	"ragbot/internal/conversation"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	ai "ragbot/internal/ai"
	"ragbot/internal/handler" // поправлен импорт
	"ragbot/internal/util"
)

type contactState struct {
	Stage int // 1 - expect name, 2 - expect phone
	Name  string
}

var (
	stateMu      sync.Mutex
	contactSteps = make(map[int64]*contactState)
	userBot      *tgbotapi.BotAPI
	db           *sql.DB
	aiClient     *ai.AIClient
)

// StartUserBot запускает Telegram-бота для пользователей.
func StartUserBot(dbConn *sql.DB, AIClient *ai.AIClient, token string) {
	defer util.Recover("StartUserBot")

	aiClient = AIClient
	db = dbConn
	userBot = connect(token)
	log.Println("User bot connected to Telegram API")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := userBot.GetUpdatesChan(u)

	log.Println("User bot started")
	for update := range updates {
		if update.CallbackQuery != nil {
			chatID := update.CallbackQuery.Message.Chat.ID
			conversation.EnsureSession(db, chatID)
			data := update.CallbackQuery.Data

			// 2.1) Сразу подтверждаем колбек, чтобы Telegram убрал «часики»
			//     Второй аргумент (text) может быть пустым, т.е. просто «OK»
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
			if _, err := userBot.Request(callback); err != nil {
				log.Printf("Callback answer error: %v", err)
			}

			// 2.2) В зависимости от data реагируем
			switch data {
			case "CALL_MANAGER":
				callManagerAction(chatID)

			// При необходимости можно добавить другие data-коды
			default:
				// Если неожиданный код, просто ничего не делаем или логируем
				log.Printf("Unknown CallbackQuery data: %s", data)
			}
			continue
		}

		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		conversation.EnsureSession(db, chatID)
		userText := update.Message.Text

		stateMu.Lock()
		st, ok := contactSteps[chatID]
		stateMu.Unlock()
		if ok {
			switch st.Stage {
			case 1:
				conversation.AppendHistory(db, chatID, "user", userText)
				conversation.UpdateName(db, chatID, userText)
				stateMu.Lock()
				st.Stage = 2
				st.Name = userText
				stateMu.Unlock()
				reply(chatID, "Укажите телефон для связи")
				continue
			case 2:
				conversation.AppendHistory(db, chatID, "user", userText)
				conversation.UpdatePhone(db, chatID, userText)
				stateMu.Lock()
				delete(contactSteps, chatID)
				stateMu.Unlock()
				reply(chatID, "Наш менеджер свяжется с вами в ближайшее время")
				continue
			}
		}

		// Если пользователь хочет «записаться» — предлагаем кнопку «Вызвать менеджера»
		lowerRequest := strings.ToLower(userText)
		if util.ContainsStringFromSlice(lowerRequest, config.Settings.CallManagerTriggerWords) {
			conversation.AppendHistory(db, chatID, "user", userText)
			userBot.Send(callMeBackButton(chatID))
			continue
		}

		// Вызываем обработку вопроса (RAG + учёт истории)
		answer, err := handler.ProcessQuestionWithHistory(db, aiClient, chatID, userText)
		if err != nil {
			SendToAllAdmins(fmt.Sprintf("Возникла ошибка: %s", err))
			answer = "Возникла ошибка. Пожалуйста, попробуйте повторить ваш запрос позднее."
		}

		conversation.AppendHistory(db, chatID, "user", userText)
		conversation.AppendHistory(db, chatID, "assistant", answer)

		lowerAnswer := strings.ToLower(answer)
		if strings.Contains(lowerAnswer, "позвать менеджера") {
			userBot.Send(callMeBackButton(chatID))
			continue
		}

		reply(chatID, answer)
	}
}

func callManagerAction(chatID int64) {
	conversation.AppendHistory(db, chatID, "user", "нажал кнопку Вызвать менеджера")

	summary, err := summarize(db, aiClient, chatID)
	if err == nil {
		conversation.UpdateSummary(db, chatID, summary)
	} else {
		log.Printf("summary error: %v", err)
	}

	stateMu.Lock()
	contactSteps[chatID] = &contactState{Stage: 1}
	stateMu.Unlock()

	reply(chatID, "Как к вам можно обращаться?")

	// Уведомляем администраторов
	info, err := conversation.GetChatInfoByChatID(db, chatID)
	if err != nil {
		log.Printf("Разговор не найден")
		return
	}
	link := fmt.Sprintf("%s/chat/%s", config.Config.BaseURL, info.ID)
	adminMsg := fmt.Sprintf("%s (%s): %s\n\n%s", info.Name.String, info.Phone.String, info.Summary.String, link)

	SendToAllAdmins(adminMsg)
}

func summarize(db *sql.DB, aiClient *ai.AIClient, chatID int64) (string, error) {
	// Генерируем краткое резюме последних сообщений
	hist := conversation.GetHistory(db, chatID)
	var sb strings.Builder
	for _, h := range hist {
		if h.Role == "user" {
			sb.WriteString("Пользователь: " + h.Content + "\n")
		} else {
			sb.WriteString("Помощник: " + h.Content + "\n")
		}
	}
	prompt := "Суммаризируй диалог пользователя в двух предложениях:\n" + sb.String() + "\nРезюме:"
	summary, err := aiClient.GenerateResponse(prompt)
	return summary, err
}

func reply(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	userBot.Send(msg)
	conversation.AppendHistory(db, chatID, "assistant", message)
}

func callMeBackButton(chatID int64) tgbotapi.MessageConfig {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Позвать менеджера в чат", "CALL_MANAGER"),
		),
	)
	msg := tgbotapi.NewMessage(chatID, "Если вы хотите записаться на занятие, нажмите кнопку:")
	msg.ReplyMarkup = keyboard
	return msg
}
