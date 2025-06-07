package bot

import (
	"fmt"
	"log"
	"ragbot/internal/ai"
	"ragbot/internal/amo"
	"ragbot/internal/config"
	"ragbot/internal/conversation"
	"ragbot/internal/repository"
	"strings"
)

func requestUserPhoneNumber(chatID int64, userText string) {
	conversation.AppendHistory(repo, chatID, "user", userText)
	conversation.UpdatePhone(repo, chatID, userText)
	stateMu.Lock()
	delete(contactSteps, chatID)
	stateMu.Unlock()
	replyToUser(chatID, "Наш менеджер свяжется с вами в ближайшее время")
	info, err := conversation.GetChatInfoByChatID(repo, chatID)
	if err == nil {
		link := fmt.Sprintf(chatUrlFormat, config.Config.BaseURL, info.ID)
		adminMsg := fmt.Sprintf("%s (%s): %s\n\n%s", info.Name.String, info.Phone.String, info.Summary.String, link)
		SendToAllAdmins(adminMsg)
		amo.SendLead(info.Name.String, info.Phone.String, info.Summary.String+"\n\n"+link)
	}
}

func requestUserName(chatID int64, userText string, st *contactState) {
	conversation.AppendHistory(repo, chatID, "user", userText)
	conversation.UpdateName(repo, chatID, userText)
	stateMu.Lock()
	st.Stage = 2
	st.Name = userText
	stateMu.Unlock()
	replyToUser(chatID, "Напишите ваш телефон для связи")
}

func callManagerAction(chatID int64) {
	conversation.AppendHistory(repo, chatID, "user", "** хочет, чтобы ему перезвонили **")

	summary, err := summarize(repo, aiClient, chatID)
	if err == nil {
		conversation.UpdateSummary(repo, chatID, summary)
	} else {
		log.Printf("summary error: %v", err)
	}

	stateMu.Lock()
	contactSteps[chatID] = &contactState{Stage: 1}
	stateMu.Unlock()

	replyToUser(chatID, "Как к вам можно обращаться?")
}

func summarize(repo *repository.Repository, aiClient *ai.AIClient, chatID int64) (string, error) {
	hist := conversation.GetHistory(repo, chatID)
	var sb strings.Builder
	for _, h := range hist {
		if h.Role == "user" {
			sb.WriteString("Пользователь: " + h.Content + "\n")
		} else {
			sb.WriteString("Помощник: " + h.Content + "\n")
		}
	}
	prompt := "Суммаризируй диалог пользователя в двух предложениях:\n" + sb.String() + "\nРезюме:"
	summary, err := aiClient.GenerateResponse(prompt)
	return summary, err
}
