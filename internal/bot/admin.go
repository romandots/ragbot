package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ragbot/internal/config"
	"ragbot/internal/repository"
	"ragbot/internal/util"
)

const source = "admin"

var adminChats []int64
var adminBot *tgbotapi.BotAPI

// StartAdminBot launches Telegram bot for knowledge base administration.
func StartAdminBot(repo *repository.Repository, token string, allowedIDs []int64) {
	defer util.Recover("StartAdminBot")

	adminBot = connect(token)
	log.Println("Admin bot connected to Telegram API")

	registerAdminCommands()
	handleAdminUpdates(repo, allowedIDs)
}

func handleAdminUpdates(repo *repository.Repository, allowedIDs []int64) {
	adminChats = allowedIDs
	allowed := make(map[int64]bool)
	for _, id := range allowedIDs {
		allowed[id] = true
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := adminBot.GetUpdatesChan(u)

	log.Println("Admin bot started")
	for update := range updates {
		if update.Message == nil {
			continue
		}
		handleAdminMessage(repo, update, allowed)
	}
}

func handleAdminMessage(repo *repository.Repository, update tgbotapi.Update, allowed map[int64]bool) bool {
	chatID := update.Message.Chat.ID

	if !allowed[chatID] {
		return true
	}

	if handleAdminCommand(repo, update, chatID) {
		return true
	}

	text := strings.TrimSpace(update.Message.Text)
	content := strings.Trim(text, " ")
	id, err := repo.AddChunk(context.Background(), content, source)
	if err != nil {
		replyToAdmin(chatID, fmt.Sprintf(msgAdminAddError, content))
		return true
	}
	if id != 0 {
		SendToAllAdmins(fmt.Sprintf(msgAdminAdded, id, content))
	} else {
		replyToAdmin(chatID, msgAdminExists)
	}
	return false
}

func handleAdminCommand(repo *repository.Repository, update tgbotapi.Update, chatID int64) bool {
	if update.Message.IsCommand() {
		cmd := update.Message.Command()
		args := update.Message.CommandArguments()

		switch cmd {
		case "start", "myid":
			replyToAdmin(chatID, fmt.Sprintf(msgAdminMyIDFormat, chatID))
			return true
		case "help":
			replyToAdmin(chatID, msgAdminHelp)
			return true
		case "delete":
			idStr := strings.Fields(args)
			if len(idStr) == 0 {
				replyToAdmin(chatID, msgAdminInvalidID)
				return true
			}
			id, err := strconv.Atoi(idStr[0])
			if err != nil {
				replyToAdmin(chatID, msgAdminInvalidID)
				return true
			}
			content, err := repo.DeleteChunk(context.Background(), id)
			if err != nil {
				replyToAdmin(chatID, fmt.Sprintf(msgAdminDeleteError, id))
				return true
			}
			replyToAdmin(chatID, fmt.Sprintf(msgAdminDeletedFormat, id, content))
			return true
		case "update":
			parts := strings.SplitN(args, " ", 2)
			if len(parts) < 2 {
				replyToAdmin(chatID, msgAdminUpdateUsage)
				return true
			}
			id, err := strconv.Atoi(parts[0])
			if err != nil {
				replyToAdmin(chatID, msgAdminInvalidID)
				return true
			}
			content := parts[1]
			if err := repo.UpdateChunk(context.Background(), id, content); err != nil {
				replyToAdmin(chatID, fmt.Sprintf(msgAdminUpdateError, id, content))
				return true
			}
			replyToAdmin(chatID, fmt.Sprintf(msgAdminUpdatedFormat, id, content))
			return true
		case "list":
			chunks, err := repo.ListChunksWithoutExtID(context.Background())
			if err != nil {
				replyToAdmin(chatID, fmt.Sprintf("Ошибка получения списка: %v", err))
				return true
			}
			var sb strings.Builder
			for _, c := range chunks {
				escapedText := tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, c.Content)
				sb.WriteString(fmt.Sprintf("*%d* %s\n\n", c.ID, escapedText))
			}
			replyToAdminMarkdownV2(chatID, sb.String())
			return true
		case "stats":
			url := fmt.Sprintf("%s/stats", config.Config.BaseURL)
			adminBot.Send(statsButton(chatID, url))
			return true
		case "chats":
			url := fmt.Sprintf("%s/chats", config.Config.BaseURL)
			adminBot.Send(chatsButton(chatID, url))
			return true
		}
	}
	return false
}

func registerAdminCommands() {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Получить ваш chat ID"},
		{Command: "help", Description: "Показать справку по командам"},
		{Command: "update", Description: "Обновить фрагмент: /update <id> <текст>"},
		{Command: "delete", Description: "Удалить фрагмент: /delete <id>"},
		{Command: "list", Description: "Показать все фрагменты"},
		{Command: "stats", Description: "Открыть статистику"},
		{Command: "chats", Description: "Открыть список чатов"},
	}

	_, err := adminBot.Request(tgbotapi.NewSetMyCommands(commands...))
	if err != nil {
		log.Printf("Failed registering commands: %v", err)
	}
}

func SendToAllAdmins(message string) {
	for _, adminChatID := range adminChats {
		replyToAdmin(adminChatID, message)
	}
}

func replyToAdmin(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	_, err := adminBot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %s", err.Error())
	}
}

func replyToAdminMarkdownV2(chatID int64, message string) {
	msg := tgbotapi.NewMessage(chatID, message)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	_, err := adminBot.Send(msg)
	if err != nil {
		log.Printf("Error sending message: %s", err.Error())
	}
}
