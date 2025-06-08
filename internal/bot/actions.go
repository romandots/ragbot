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

func finalizeContactRequest(chatID int64) {
	stateMu.Lock()
	delete(contactSteps, chatID)
	stateMu.Unlock()
	replyToUser(chatID, msgManagerWillCall)
	info, err := conversation.GetChatInfoByChatID(repo, chatID)
	if err == nil {
		link := fmt.Sprintf(chatUrlFormat, config.Config.BaseURL, info.ID)
		adminMsg := fmt.Sprintf(msgAdminSummaryFormat, info.Name.String, info.Phone.String, info.Summary.String, link)
		SendToAllAdmins(adminMsg)
		amo.SendLead(info.Name.String, info.Phone.String, info.Summary.String+"\n\n"+link)
	}
}

func requestUserPhoneNumber(chatID int64, userText string) {
	conversation.AppendHistory(repo, chatID, "user", userText)
	conversation.UpdatePhone(repo, chatID, userText)
	finalizeContactRequest(chatID)
}

func requestUserName(chatID int64, userText string, st *contactState) {
	conversation.AppendHistory(repo, chatID, "user", userText)
	conversation.UpdateName(repo, chatID, userText)
	stateMu.Lock()
	st.Stage = 2
	st.Name = userText
	stateMu.Unlock()
	replyToUser(chatID, msgAskPhone)
}

func callManagerAction(chatID int64) {
	conversation.AppendHistory(repo, chatID, "user", historyCallRequested)

	summary, err := summarize(repo, aiClient, chatID)
	if err == nil {
		conversation.UpdateSummary(repo, chatID, summary)
	} else {
		log.Printf("summary error: %v", err)
	}

	info, err := conversation.GetChatInfoByChatID(repo, chatID)
	if err == nil && info.Name.Valid && info.Phone.Valid && info.Name.String != "" && info.Phone.String != "" {
		stateMu.Lock()
		contactSteps[chatID] = &contactState{Stage: 3}
		stateMu.Unlock()
		msg := confirmContactButton(chatID, info.Name.String, info.Phone.String)
		userBot.Send(msg)
		return
	}

	stateMu.Lock()
	contactSteps[chatID] = &contactState{Stage: 1}
	stateMu.Unlock()

	replyToUser(chatID, msgAskName)
}

func summarize(repo *repository.Repository, aiClient *ai.AIClient, chatID int64) (string, error) {
	hist := conversation.GetHistory(repo, chatID)
	var sb strings.Builder
	for _, h := range hist {
		if h.Role == "user" {
			sb.WriteString(promptUserPrefix + h.Content + "\n")
		} else {
			sb.WriteString(promptAssistantPrefix + h.Content + "\n")
		}
	}
	prompt := fmt.Sprintf(promptSummarizeFormat, sb.String())
	summary, err := aiClient.GenerateResponse(prompt)
	return summary, err
}
