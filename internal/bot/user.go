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

const (
	actionCallManager = "CALL_MANAGER"
	actionConfirmYes  = "CONFIRM_YES"
	actionConfirmNo   = "CONFIRM_NO"
)

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
			handleCallbackQuery(update)
			continue
		}

		if update.Message == nil {
			continue
		}

		handleUserMessage(update)
	}
}

func handleUserMessage(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	conversation.EnsureSession(repo, chatID)
	userText := update.Message.Text
	var answer string

	defer func() {
		conversation.AppendHistory(repo, chatID, "user", userText)
		if answer != "" {
			conversation.AppendHistory(repo, chatID, "assistant", answer)
		}
	}()

	stateMu.Lock()
	st, ok := contactSteps[chatID]
	stateMu.Unlock()
	if ok {
		switch st.Stage {
		case 1:
			requestUserName(chatID, userText, st)
			return
		case 2:
			requestUserPhoneNumber(chatID, userText)
			return
		case 3:
			lower := strings.ToLower(userText)
			if strings.Contains(lower, "да") {
				conversation.AppendHistory(repo, chatID, "user", historyConfirmYes)
				finalizeContactRequest(chatID)
				return
			}
			if strings.Contains(lower, "нет") {
				conversation.AppendHistory(repo, chatID, "user", historyConfirmNo)
				conversation.ClearAmoContactID(repo, chatID)
				stateMu.Lock()
				contactSteps[chatID] = &contactState{Stage: 1}
				stateMu.Unlock()
				replyToUser(chatID, msgAskName)
				return
			}
		}
	}

	lowerRequest := strings.ToLower(userText)
	if util.ContainsStringFromSlice(lowerRequest, config.Settings.CallManagerTriggerWords) {
		userBot.Send(callMeBackButton(chatID))
		return
	}

	answer, err := handler.ProcessQuestionWithHistory(repo, aiClient, chatID, userText)
	if err != nil {
		SendToAllAdmins(fmt.Sprintf(msgAdminErrorFormat, err))
		answer = msgUserError
	} else {
		lowerAnswer := strings.ToLower(answer)
		if util.ContainsStringFromSlice(lowerAnswer, config.Settings.CallManagerTriggerWordsInAnswer) {
			userBot.Send(callMeBackButton(chatID))
			return
		}
	}

	replyToUser(chatID, answer)
}

func handleCallbackQuery(update tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	conversation.EnsureSession(repo, chatID)
	data := update.CallbackQuery.Data
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	if _, err := userBot.Request(callback); err != nil {
		log.Printf("Callback answer error: %v", err)
	}

	// Handle actions
	switch data {
	case actionCallManager:
		callManagerAction(chatID)
	case actionConfirmYes:
		conversation.AppendHistory(repo, chatID, "user", historyConfirmYes)
		finalizeContactRequest(chatID)
	case actionConfirmNo:
		conversation.AppendHistory(repo, chatID, "user", historyConfirmNo)
		conversation.ClearAmoContactID(repo, chatID)
		stateMu.Lock()
		contactSteps[chatID] = &contactState{Stage: 1}
		stateMu.Unlock()
		replyToUser(chatID, msgAskName)
	default:
		log.Printf("Unknown CallbackQuery data: %s", data)
	}
}

func replyToUser(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	userBot.Send(msg)
}
