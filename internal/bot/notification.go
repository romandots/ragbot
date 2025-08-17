package bot

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ragbot/internal/util"
)

var notificationChats []int64
var notificationBot *tgbotapi.BotAPI

// StartNotificationBot launches Telegram bot for notifications only.
func StartNotificationBot(token string, allowedIDs []int64) {
	defer util.Recover("StartNotificationBot")

	notificationBot = connect(token)
	log.Println("Notification bot connected to Telegram API")

	registerNotificationCommands()
	handleNotificationUpdates(allowedIDs)
}

func handleNotificationUpdates(allowedIDs []int64) {
	notificationChats = allowedIDs
	allowed := make(map[int64]bool)
	for _, id := range allowedIDs {
		allowed[id] = true
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := notificationBot.GetUpdatesChan(u)

	log.Println("Notification bot started")
	for update := range updates {
		if update.Message == nil {
			continue
		}
		handleNotificationMessage(update, allowed)
	}
}

func handleNotificationMessage(update tgbotapi.Update, allowed map[int64]bool) {
	chatID := update.Message.Chat.ID

	if !allowed[chatID] {
		return
	}

	// Only handle /chatId command for getting chat ID
	if update.Message.IsCommand() {
		cmd := update.Message.Command()
		switch cmd {
		case "chatId", "start":
			replyToNotification(chatID, fmt.Sprintf("Ваш CHAT ID: %d", chatID))
		default:
			replyToNotification(chatID, "Этот бот только для получения уведомлений. Доступная команда: /chatId")
		}
	}
	// Ignore all non-command messages - this bot is read-only for notifications
}

func registerNotificationCommands() {
	commands := []tgbotapi.BotCommand{
		{Command: "chatId", Description: "Получить ваш chat ID"},
	}

	_, err := notificationBot.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		log.Printf("Failed registering notification commands: %v", err)
	}
}

// SendToAllNotifications sends notification message to all configured notification chats.
func SendToAllNotifications(message string) {
	if notificationBot == nil {
		log.Printf("Notification bot not initialized, skipping notification: %s", message)
		return
	}
	for _, notificationChatID := range notificationChats {
		replyToNotification(notificationChatID, message)
	}
}

func replyToNotification(chatID int64, message string) {
	if notificationBot == nil {
		log.Printf("Notification bot not initialized, cannot send to chat %d: %s", chatID, message)
		return
	}
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := notificationBot.Send(msg)
	if err != nil {
		log.Printf("Error sending notification message: %s", err.Error())
	}
}