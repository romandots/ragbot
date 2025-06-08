package bot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func callMeBackButton(chatID int64) tgbotapi.MessageConfig {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(msgCallManagerButton, actionCallManager),
		),
	)
	msg := tgbotapi.NewMessage(chatID, msgCallManagerPrompt)
	msg.ReplyMarkup = keyboard
	return msg
}

func confirmContactButton(chatID int64, name, phone string) tgbotapi.MessageConfig {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(msgConfirmYes, actionConfirmYes),
			tgbotapi.NewInlineKeyboardButtonData(msgConfirmNo, actionConfirmNo),
		),
	)
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf(msgConfirmContactFormat, name, phone))
	msg.ReplyMarkup = keyboard
	return msg
}
