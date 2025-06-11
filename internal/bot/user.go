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
	}

	_, err := userBot.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		log.Printf("Failed registering commands for user bot: %v", err)
	}
}

func handleUserMessage(update tgbotapi.Update) {
	chatID := update.Message.Chat.ID
	conversation.EnsureSession(repo, chatID)
	userText := update.Message.Text
	var answer string

	// Обработка команды /start - инициализируем общение как если бы пользователь написал "Привет"
	if userText == "/start" {
		userText = "Привет"
	}

	defer func() {
		conversation.AppendHistory(repo, chatID, "user", userText)
		if answer != "" {
			conversation.AppendHistory(repo, chatID, "assistant", answer)
		}
	}()

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
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, b := range branches {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(fmt.Sprintf(msgScheduleLinkFormat, b.Title), b.ScheduleLink),
		))
	}
	msg := tgbotapi.NewMessage(chatID, msgScheduleTitle)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	userBot.Send(msg)
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
	var rows [][]tgbotapi.InlineKeyboardButton
	priceMu.Lock()
	priceMap = make(map[string]string)
	for i, p := range prices {
		key := fmt.Sprintf("PRICE_%d", i)
		properties := ""
		if p.Hours != "" {
			properties = properties + fmt.Sprintf(msgPassHoursFormat, p.Hours)
		}
		if p.GuestVisits != "" {
			properties = properties + fmt.Sprintf(msgGuestVisitsFormat, p.GuestVisits)
		}
		if p.FreezeAllowed != "" {
			properties = properties + msgPassFreezeAllowed
		}
		if p.Lifetime != "" {
			properties = properties + fmt.Sprintf(msgPassLifetimeFormat, p.Lifetime)
		}
		properties = tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, properties)
		if p.Price != "" {
			properties = properties + fmt.Sprintf(msgPriceFormat, p.Price)
		}
		name := tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, p.Name)
		description := tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, p.Description)
		priceMap[key] = fmt.Sprintf(msgPriceDescriptionFormat, name, description, properties)
		label := fmt.Sprintf(msgPriceButtonFormat, p.Name, p.Price)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, key),
		))
	}
	priceMu.Unlock()
	msg := tgbotapi.NewMessage(chatID, msgPricesTitle)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	userBot.Send(msg)
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
		if strings.HasPrefix(data, "PRICE_") {
			priceMu.RLock()
			desc := priceMap[data]
			priceMu.RUnlock()
			if desc != "" {
				replyToUserMarkdownV2(chatID, desc)
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
