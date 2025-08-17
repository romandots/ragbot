package bot

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	ai "ragbot/internal/ai"
	"ragbot/internal/config"
	"ragbot/internal/conversation"
	"ragbot/internal/handler"
	"ragbot/internal/repository"
	"ragbot/internal/tansultant"
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
	tansClient   *tansultant.Client
	priceMu      sync.RWMutex
	priceMap     = make(map[string]string)
)

// StartUserBot launches Telegram bot for users.
func StartUserBot(r *repository.Repository, ac *ai.AIClient, tc *tansultant.Client, token string) {
	defer util.Recover("StartUserBot")

	aiClient = ac
	tansClient = tc
	repo = r
	userBot = connect(token)
	log.Println("User bot connected to Telegram API")

	registerUserCommands()
	handleUserUpdates()
}

func handleUserUpdates() {
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

func registerUserCommands() {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Начать общение с ассистентом"},
		{Command: "address", Description: "Показать адреса студий"},
		{Command: "prices", Description: "Показать цены на обучение"},
		{Command: "rasp", Description: "Показать расписание занятий"},
		{Command: "call", Description: "Заказать обратный звонок от менеджера"},
		{Command: "channel", Description: "Перейти в телеграм-канал ШТБП"},
	}

	_, err := userBot.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		log.Printf("Failed registering commands for user bot: %v", err)
	}
}

func handleUserMessage(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	username := ""
	if update.Message.From != nil {
		username = update.Message.From.UserName
	}
	conversation.EnsureSession(repo, chatID, username)
	
	// Handle voice messages
	if update.Message.Voice != nil {
		userText, err := handleVoiceMessage(update.Message.Voice)
		if err != nil {
			log.Printf("Voice transcription error: %v", err)
			replyToUser(chatID, "Извините, не удалось распознать голосовое сообщение. Попробуйте отправить текстом.")
			return
		}
		// Continue processing as text message
		processTextMessage(chatID, username, userText)
		return
	}
	
	userText := update.Message.Text
	
	// Обработка команды /start - инициализируем общение как если бы пользователь написал "Привет"
	if userText == "/start" {
		userText = "Привет"
	}
	
	processTextMessage(chatID, username, userText)
}

func processTextMessage(chatID int64, username, userText string) {
	var answer string

	defer func() {
		conversation.AppendHistory(repo, chatID, "user", userText)
		if answer != "" {
			conversation.AppendHistory(repo, chatID, "assistant", answer)
		}
	}()

	// Create a minimal update for handleUserCommand
	update := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: userText,
			Chat: &tgbotapi.Chat{ID: chatID},
			From: &tgbotapi.User{UserName: username},
		},
	}
	
	if handleUserCommand(update, chatID) {
		return
	}

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

func handleVoiceMessage(voice *tgbotapi.Voice) (string, error) {
	// Get file info from Telegram
	fileConfig := tgbotapi.FileConfig{FileID: voice.FileID}
	file, err := userBot.GetFile(fileConfig)
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %v", err)
	}

	// Download the voice file
	fileURL := file.Link(userBot.Token)
	resp, err := http.Get(fileURL)
	if err != nil {
		return "", fmt.Errorf("failed to download voice file: %v", err)
	}
	defer resp.Body.Close()

	// Read file data
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read voice file: %v", err)
	}

	// Determine file format (Telegram voice messages are usually OGG)
	format := "ogg"

	// Transcribe using AI client
	transcription, err := aiClient.TranscribeAudio(audioData, format)
	if err != nil {
		return "", fmt.Errorf("failed to transcribe audio: %v", err)
	}

	return transcription, nil
}

func handleUserCommand(update tgbotapi.Update, chatID int64) bool {
	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "address":
			sendAddresses(chatID)
			return true
		case "prices":
			sendPrices(chatID)
			return true
		case "rasp":
			sendSchedule(chatID)
			return true
		case "call":
			userBot.Send(callMeBackButton(chatID))
			return true
		case "channel":
			userBot.Send(channelButton(chatID, config.Config.TelegramChannel))
			return true
		default:
		}
	}
	return false
}

func sendAddresses(chatID int64) {
	if tansClient == nil {
		log.Printf("Failed retrieving addresses: Tansultant client is not instantiated")
		replyToUser(chatID, msgServiceUnavailable)
		return
	}
	branches, err := tansClient.Branches()
	if err != nil || len(branches) == 0 {
		log.Printf("Failed retrieving addresses: %s", err.Error())
		replyToUser(chatID, msgInfoUnavailable)
		return
	}
	var sb strings.Builder
	for _, b := range branches {
		sb.WriteString(fmt.Sprintf(msgAddressFormat, b.Title, b.Address))
	}
	replyToUser(chatID, sb.String())
}

func sendSchedule(chatID int64) {
	if tansClient == nil {
		replyToUser(chatID, msgServiceUnavailable)
		return
	}
	branches, err := tansClient.Branches()
	if err != nil || len(branches) == 0 {
		log.Printf("Failed retrieving schedule: %s", err.Error())
		replyToUser(chatID, msgInfoUnavailable)
		return
	}
	userBot.Send(scheduleButtons(chatID, branches))
}

func sendPrices(chatID int64) {
	if tansClient == nil {
		replyToUser(chatID, msgServiceUnavailable)
		return
	}
	prices, err := tansClient.Prices()
	if err != nil || len(prices) == 0 {
		log.Printf("Failed retrieving prices: %s", err.Error())
		replyToUser(chatID, msgInfoUnavailable)
		return
	}
	priceMu.Lock()
	msg, m := priceButtons(chatID, prices)
	priceMap = m
	priceMu.Unlock()
	userBot.Send(msg)
}

func handleCallbackQuery(update tgbotapi.Update) {
	chatID := update.CallbackQuery.Message.Chat.ID
	messageID := update.CallbackQuery.Message.MessageID
	username := ""
	if update.CallbackQuery.From != nil {
		username = update.CallbackQuery.From.UserName
	}
	conversation.EnsureSession(repo, chatID, username)
	data := update.CallbackQuery.Data
	callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
	if _, err := userBot.Request(callback); err != nil {
		log.Printf("Callback answer error: %v", err)
	}

	// Handle actions
	switch data {
	case actionCallManager:
		callManagerAction(chatID)
		// Удаляем сообщение с кнопкой после нажатия
		deleteMessage(chatID, messageID)
	case actionConfirmYes:
		conversation.AppendHistory(repo, chatID, "user", historyConfirmYes)
		finalizeContactRequest(chatID)
		// Удаляем сообщение с кнопкой после нажатия
		deleteMessage(chatID, messageID)
	case actionConfirmNo:
		conversation.AppendHistory(repo, chatID, "user", historyConfirmNo)
		conversation.ClearAmoContactID(repo, chatID)
		stateMu.Lock()
		contactSteps[chatID] = &contactState{Stage: 1}
		stateMu.Unlock()
		replyToUser(chatID, msgAskName)
		// Удаляем сообщение с кнопкой после нажатия
		deleteMessage(chatID, messageID)
	default:
		if strings.HasPrefix(data, "PRICE_") {
			priceMu.RLock()
			desc := priceMap[data]
			priceMu.RUnlock()
			if desc != "" {
				replyToUserMarkdownV2(chatID, desc)
				// Удаляем сообщение с кнопкой после нажатия
				deleteMessage(chatID, messageID)
			} else {
				log.Printf("Unknown price callback: %s", data)
			}
		} else {
			log.Printf("Unknown CallbackQuery data: %s", data)
		}
	}
}

func replyToUser(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := userBot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %s", err.Error())
	}
}

func replyToUserMarkdownV2(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	_, err := userBot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %s", err.Error())
	}
}

// deleteMessage удаляет сообщение по chatID и messageID
func deleteMessage(chatID int64, messageID int) {
	deleteMsg := tgbotapi.NewDeleteMessage(chatID, messageID)
	_, err := userBot.Request(deleteMsg)
	if err != nil {
		log.Printf("Error deleting message: %s", err.Error())
	}
}
