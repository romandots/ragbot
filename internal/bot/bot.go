package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"time"
)

func connect(token string) *tgbotapi.BotAPI {
	var bot *tgbotapi.BotAPI
	var err error
	try := 0
	for bot, err = tgbotapi.NewBotAPI(token); err != nil; try++ {
		log.Printf("Telegram bot init error: %v\nWaiting for %d seconds and trying again.", err, try)
		time.Sleep(time.Duration(try) * time.Second)
		log.Println("Trying to connect to Telegram bot API again...")
	}
	return bot
}

// deleteWebhook удаляет активный webhook перед использованием getUpdates
func deleteWebhook(bot *tgbotapi.BotAPI) {
	config := tgbotapi.DeleteWebhookConfig{
		DropPendingUpdates: true,
	}
	_, err := bot.Request(config)
	if err != nil {
		log.Printf("Warning: Failed to delete webhook (may not be set): %v", err)
	} else {
		log.Println("Webhook deleted successfully")
	}
}
