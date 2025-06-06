package bot

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"ragbot/internal/conversation"

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
)

// StartUserBot запускает Telegram-бота для пользователей.
// Параметры:
//   - db       : указатель на SQL-соединение
//   - aiClient : экземпляр AIClient для генерации ответов
//   - token    : токен пользовательского бота (из ENV)
func StartUserBot(db *sql.DB, aiClient *ai.AIClient, token string) {
	defer util.Recover("StartUserBot")
	bot, err := tgbotapi.NewBotAPI(token)
	for err != nil {
		log.Printf("User bot init error: %v\n", err)
		time.Sleep(1 * time.Second)
		log.Println("Trying to connect to Telegram bot API again...")
		bot, err = tgbotapi.NewBotAPI(token)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	log.Println("User bot started")
	for update := range updates {
		if update.CallbackQuery != nil {
			chatID := update.CallbackQuery.Message.Chat.ID
			conversation.EnsureSession(db, chatID)
			data := update.CallbackQuery.Data

			// 2.1) Сразу подтверждаем колбек, чтобы Telegram убрал «часики»
			//     Второй аргумент (text) может быть пустым, т.е. просто «OK»
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
			if _, err := bot.Request(callback); err != nil {
				log.Printf("Callback answer error: %v", err)
			}

			// 2.2) В зависимости от data реагируем
			switch data {
			case "CALL_MANAGER":
				// Сохраняем в историю факт нажатия
				conversation.AppendHistory(db, chatID, "user", "нажал кнопку Вызвать менеджера")

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
				if err == nil {
					conversation.UpdateSummary(db, chatID, summary)
				} else {
					log.Printf("summary error: %v", err)
				}

				stateMu.Lock()
				contactSteps[chatID] = &contactState{Stage: 1}
				stateMu.Unlock()

				msg := tgbotapi.NewMessage(chatID, "Как вас зовут?")
				bot.Send(msg)
				conversation.AppendHistory(db, chatID, "assistant", "Как вас зовут?")

				// Если нужно — уведомляем менеджера (в админ-чат или личным сообщением)
				// Например, пусть менеджер сидит в чате с ID = ADMIN_CHAT_ID
				// adminMsg := tgbotapi.NewMessage(ADMIN_CHAT_ID, fmt.Sprintf("Пользователь %d хочет записаться на занятие.", chatID))
				// bot.Send(adminMsg)

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
				msg := tgbotapi.NewMessage(chatID, "Укажите номер телефона")
				bot.Send(msg)
				conversation.AppendHistory(db, chatID, "assistant", "Укажите номер телефона")
				continue
			case 2:
				conversation.AppendHistory(db, chatID, "user", userText)
				conversation.UpdatePhone(db, chatID, userText)
				stateMu.Lock()
				delete(contactSteps, chatID)
				stateMu.Unlock()
				msg := tgbotapi.NewMessage(chatID, "Менеджер уже уведомлён и свяжется с вами в ближайшее время.")
				bot.Send(msg)
				conversation.AppendHistory(db, chatID, "assistant", "Менеджер уже уведомлён и свяжется с вами в ближайшее время.")
				continue
			}
		}

		lower := strings.ToLower(userText)

		// 0) Если пользователь хочет «записаться» — предлагаем кнопку «Вызвать менеджера»
		if strings.Contains(lower, "записаться") || strings.Contains(lower, "позови") || strings.Contains(lower, "человека") {
			conversation.AppendHistory(db, chatID, "user", userText)
			// Формируем inline-клавиатуру с одной кнопкой
			// При нажатии будет послан callback с данными "CALL_MANAGER"
			msg := getCallManagerButton(chatID)
			bot.Send(msg)
			continue
		}

		// 1) Вызываем обработку вопроса (RAG + учёт истории)
		answer, err := handler.ProcessQuestionWithHistory(db, aiClient, chatID, userText)
		if err != nil {
			answer = fmt.Sprintf("Ошибка: %v", err)
		}
		lowerAnswer := strings.ToLower(answer)

		// 2) Сохраняем вопрос пользователя в историю
		conversation.AppendHistory(db, chatID, "user", userText)

		// 3) Сохраняем ответ бота в историю
		conversation.AppendHistory(db, chatID, "assistant", answer)

		if strings.Contains(lowerAnswer, "позвать менеджера") {
			// Формируем inline-клавиатуру с одной кнопкой
			// При нажатии будет послан callback с данными "CALL_MANAGER"
			msg := getCallManagerButton(chatID)
			bot.Send(msg)
			continue
		}

		// 4) Отправляем ответ
		bot.Send(tgbotapi.NewMessage(chatID, answer))
	}
}

func getCallManagerButton(chatID int64) tgbotapi.MessageConfig {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Позвать менеджера в чат", "CALL_MANAGER"),
		),
	)
	msg := tgbotapi.NewMessage(chatID, "Если вы хотите записаться на занятие, нажмите кнопку:")
	msg.ReplyMarkup = keyboard
	return msg
}
