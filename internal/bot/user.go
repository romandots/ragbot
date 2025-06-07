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
	"ragbot/internal/util"
)

type contactState struct {
	Stage int
	Name  string
}

const chatUrlFormat = "%s/chat/%s"

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
				requestUserName(chatID, userText, st)
				continue
			case 2:
				requestUserPhoneNumber(chatID, userText)
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

func replyToUser(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	userBot.Send(msg)
	conversation.AppendHistory(repo, chatID, "assistant", message)
}
