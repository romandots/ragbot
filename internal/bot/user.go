package bot

import (
	"fmt"
	"log"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	ai "ragbot/internal/ai"
	"ragbot/internal/config"
	"ragbot/internal/conversation"
	"ragbot/internal/handler"
	"ragbot/internal/repository"
	"ragbot/internal/amo"
	"ragbot/internal/util"
)

type contactState struct {
	Stage int
	Name  string
}

var (
	stateMu      sync.Mutex
	contactSteps = make(map[int64]*contactState)
	userBot      *tgbotapi.BotAPI
	repo         *repository.Repository
	aiClient     *ai.AIClient
)

// StartUserBot launches Telegram bot for users.
func StartUserBot(r *repository.Repository, AIClient *ai.AIClient, token string) {
	defer util.Recover("StartUserBot")

	aiClient = AIClient
	repo = r
	userBot = connect(token)
	log.Println("User bot connected to Telegram API")

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := userBot.GetUpdatesChan(u)

	log.Println("User bot started")
	for update := range updates {
		if update.CallbackQuery != nil {
			chatID := update.CallbackQuery.Message.Chat.ID
			conversation.EnsureSession(repo, chatID)
			data := update.CallbackQuery.Data
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
			if _, err := userBot.Request(callback); err != nil {
				log.Printf("Callback answer error: %v", err)
			}

			// Handle actions
			switch data {
			case "CALL_MANAGER":
				callManagerAction(chatID)
			default:
				log.Printf("Unknown CallbackQuery data: %s", data)
			}
			continue
		}

		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		conversation.EnsureSession(repo, chatID)
		userText := update.Message.Text

		stateMu.Lock()
		st, ok := contactSteps[chatID]
		stateMu.Unlock()
		if ok {
			switch st.Stage {
			case 1:
				conversation.AppendHistory(repo, chatID, "user", userText)
				conversation.UpdateName(repo, chatID, userText)
				stateMu.Lock()
				st.Stage = 2
				st.Name = userText
				stateMu.Unlock()
				replyToUser(chatID, "Напишите ваш телефон для связи")
				continue
			case 2:
				conversation.AppendHistory(repo, chatID, "user", userText)
				conversation.UpdatePhone(repo, chatID, userText)
				stateMu.Lock()
				delete(contactSteps, chatID)
				stateMu.Unlock()
				replyToUser(chatID, "Наш менеджер свяжется с вами в ближайшее время")
				info, err := conversation.GetChatInfoByChatID(db, chatID)
				if err == nil {
					link := fmt.Sprintf("%s/chat/%s", config.Config.BaseURL, info.ID)
					adminMsg := fmt.Sprintf("%s (%s): %s\n\n%s", info.Name.String, info.Phone.String, info.Summary.String, link)
					SendToAllAdmins(adminMsg)
					amo.SendLead(info.Name.String, info.Phone.String, info.Summary.String+"\n\n"+link)
				}
				continue
			}
		}

		lowerRequest := strings.ToLower(userText)
		if util.ContainsStringFromSlice(lowerRequest, config.Settings.CallManagerTriggerWords) {
			conversation.AppendHistory(repo, chatID, "user", userText)
			userBot.Send(callMeBackButton(chatID))
			continue
		}

		answer, err := handler.ProcessQuestionWithHistory(repo, aiClient, chatID, userText)
		if err != nil {
			SendToAllAdmins(fmt.Sprintf("Возникла ошибка: %s", err))
			answer = "Возникла ошибка. Пожалуйста, попробуйте повторить ваш запрос позднее."
		}

		conversation.AppendHistory(repo, chatID, "user", userText)
		conversation.AppendHistory(repo, chatID, "assistant", answer)

		lowerAnswer := strings.ToLower(answer)
		if util.ContainsStringFromSlice(lowerAnswer, config.Settings.CallManagerTriggerWordsInAnswer) {
			userBot.Send(callMeBackButton(chatID))
			continue
		}

		replyToUser(chatID, answer)
	}
}

func callManagerAction(chatID int64) {
	conversation.AppendHistory(repo, chatID, "user", "** хочет, чтобы ему перезвонили **")

	summary, err := summarize(repo, aiClient, chatID)
	if err == nil {
		conversation.UpdateSummary(repo, chatID, summary)
	} else {
		log.Printf("summary error: %v", err)
	}

	stateMu.Lock()
	contactSteps[chatID] = &contactState{Stage: 1}
	stateMu.Unlock()

	replyToUser(chatID, "Как к вам можно обращаться?")

	info, err := conversation.GetChatInfoByChatID(repo, chatID)
	if err != nil {
		log.Printf("Разговор не найден")
		return
	}
	link := fmt.Sprintf("%s/chat/%s", config.Config.BaseURL, info.ID)
	adminMsg := fmt.Sprintf("%s (%s): %s\n\n%s", info.Name.String, info.Phone.String, info.Summary.String, link)

	SendToAllAdmins(adminMsg)
}

func summarize(repo *repository.Repository, aiClient *ai.AIClient, chatID int64) (string, error) {
	hist := conversation.GetHistory(repo, chatID)
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

func replyToUser(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	userBot.Send(msg)
	conversation.AppendHistory(repo, chatID, "assistant", message)
}

func callMeBackButton(chatID int64) tgbotapi.MessageConfig {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Хочу, чтобы мне перезвонили", "CALL_MANAGER"),
		),
	)
	msg := tgbotapi.NewMessage(chatID, "Чтобы продолжить общение с нашим менеджером, нажмите кнопку:")
	msg.ReplyMarkup = keyboard
	return msg
}
