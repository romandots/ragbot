package bot

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ragbot/internal/tansultant"
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

func statsButton(chatID int64, url string) tgbotapi.MessageConfig {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(msgStatsButton, url),
		),
	)
	msg := tgbotapi.NewMessage(chatID, msgStatsPrompt)
	msg.ReplyMarkup = keyboard
	return msg
}

func chatsButton(chatID int64, url string) tgbotapi.MessageConfig {
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(msgChatsButton, url),
		),
	)
	msg := tgbotapi.NewMessage(chatID, msgChatsPrompt)
	msg.ReplyMarkup = keyboard
	return msg
}

func channelButton(chatID int64, channel string) tgbotapi.MessageConfig {
	url := fmt.Sprintf("https://t.me/%s", channel)
	keyboard := tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("@"+channel, url),
		),
	)
	msg := tgbotapi.NewMessage(chatID, msgChannelPrompt)
	msg.ReplyMarkup = keyboard
	return msg
}

func scheduleButtons(chatID int64, branches []tansultant.Branch) tgbotapi.MessageConfig {
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, b := range branches {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(fmt.Sprintf(msgScheduleLinkFormat, b.Title), b.ScheduleLink),
		))
	}
	msg := tgbotapi.NewMessage(chatID, msgScheduleTitle)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	return msg
}

func priceButtons(chatID int64, prices []tansultant.Price) (tgbotapi.MessageConfig, map[string]string) {
	var rows [][]tgbotapi.InlineKeyboardButton
	m := make(map[string]string)
	for i, p := range prices {
		key := fmt.Sprintf("PRICE_%d", i)
		properties := ""
		if p.Hours != "" {
			properties += fmt.Sprintf(msgPassHoursFormat, p.Hours)
		}
		if p.GuestVisits != "" {
			properties += fmt.Sprintf(msgGuestVisitsFormat, p.GuestVisits)
		}
		if p.FreezeAllowed != "" {
			properties += msgPassFreezeAllowed
		}
		if p.Lifetime != "" {
			properties += fmt.Sprintf(msgPassLifetimeFormat, p.Lifetime)
		}
		properties = tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, properties)
		if p.Price != "" {
			properties += fmt.Sprintf(msgPriceFormat, p.Price)
		}
		name := tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, p.Name)
		description := tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, p.Description)
		m[key] = fmt.Sprintf(msgPriceDescriptionFormat, name, description, properties)
		label := fmt.Sprintf(msgPriceButtonFormat, p.Name, p.Price)
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(label, key),
		))
	}
	msg := tgbotapi.NewMessage(chatID, msgPricesTitle)
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(rows...)
	return msg, m
}
