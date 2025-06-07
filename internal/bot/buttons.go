package bot

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

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
