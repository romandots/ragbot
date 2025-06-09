package bot

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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
		chatID := update.Message.Chat.ID

		if !allowed[chatID] {
			continue
		}

		if update.Message.IsCommand() {
			cmd := update.Message.Command()
			args := update.Message.CommandArguments()

			switch cmd {
			case "start", "myid":
				replyToAdmin(chatID, fmt.Sprintf(msgAdminMyIDFormat, chatID))
				continue
			case "help":
				replyToAdmin(chatID, msgAdminHelp)
				continue
			case "delete":
				idStr := strings.Fields(args)
				if len(idStr) == 0 {
					replyToAdmin(chatID, msgAdminInvalidID)
					continue
				}
				id, err := strconv.Atoi(idStr[0])
				if err != nil {
					replyToAdmin(chatID, msgAdminInvalidID)
					continue
				}
				content, err := repo.DeleteChunk(context.Background(), id)
				if err != nil {
					replyToAdmin(chatID, fmt.Sprintf(msgAdminDeleteError, id))
					continue
				}
				replyToAdmin(chatID, fmt.Sprintf(msgAdminDeletedFormat, id, content))
				continue
			case "update":
				parts := strings.SplitN(args, " ", 2)
				if len(parts) < 2 {
					replyToAdmin(chatID, msgAdminUpdateUsage)
					continue
				}
				id, err := strconv.Atoi(parts[0])
				if err != nil {
					replyToAdmin(chatID, msgAdminInvalidID)
					continue
				}
				content := parts[1]
				if err := repo.UpdateChunk(context.Background(), id, content); err != nil {
					replyToAdmin(chatID, fmt.Sprintf(msgAdminUpdateError, id, content))
					continue
				}
				replyToAdmin(chatID, fmt.Sprintf(msgAdminUpdatedFormat, id, content))
				continue
			case "list":
				chunks, err := repo.ListChunksWithoutExtID(context.Background())
				if err != nil {
					replyToAdmin(chatID, fmt.Sprintf("Ошибка получения списка: %v", err))
					continue
				}
				var sb strings.Builder
				for _, c := range chunks {
					escapedText := tgbotapi.EscapeText(tgbotapi.ModeMarkdownV2, c.Content)
					sb.WriteString(fmt.Sprintf("*%d* %s\n\n", c.ID, escapedText))
				}
				replyToAdminMarkdownV2(chatID, sb.String())
				continue
			}
		}

		text := strings.TrimSpace(update.Message.Text)
		content := strings.Trim(text, " ")
		id, err := repo.AddChunk(context.Background(), content, source)
		if err != nil {
			replyToAdmin(chatID, fmt.Sprintf(msgAdminAddError, content))
			continue
		}
		if id != 0 {
			SendToAllAdmins(fmt.Sprintf(msgAdminAdded, id, content))
		} else {
			replyToAdmin(chatID, msgAdminExists)
		}
	}
}

func registerAdminCommands() {
	commands := []tgbotapi.BotCommand{
		{Command: "start", Description: "Получить ваш chat ID"},
		{Command: "help", Description: "Показать справку по командам"},
		{Command: "update", Description: "Обновить фрагмент: /update <id> <текст>"},
		{Command: "delete", Description: "Удалить фрагмент: /delete <id>"},
		{Command: "list", Description: "Показать все фрагменты"},
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
