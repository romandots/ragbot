package bot

import (
	"fmt"
	"log"
	"ragbot/internal/amo"
	"ragbot/internal/config"
	"ragbot/internal/conversation"
	"strings"
)

func finalizeContactRequest(chatID int64) {
	stateMu.Lock()
	delete(contactSteps, chatID)
	stateMu.Unlock()
	info, err := conversation.GetChatInfoByChatID(repo, chatID)
	if err != nil {
		errMsg := fmt.Sprintf("Error sending lead to AMO: %v", err)
		SendToAllAdmins(errMsg)
		log.Println(errMsg)
		replyToUser(chatID, "Извините, возниклка какая-то ошибка. Попробуйте повторить ваш запрос позднее.")
		return
	}

	link := fmt.Sprintf(chatUrlFormat, config.Config.BaseURL, info.ID)
	adminMsg := fmt.Sprintf(msgAdminSummaryFormat, info.Name.String, info.Phone.String, info.Summary.String, link)
	SendToAllAdmins(adminMsg)

	err = amo.SendLeadToAMO(repo, &info, link)
	if err == nil {
		replyToUser(chatID, msgManagerWillCall)
		return
	}

	errMsg := fmt.Sprintf("Error sending lead to AMO: %v", err)
	SendToAllAdmins(errMsg)
	log.Println(errMsg)
	replyToUser(chatID, "Извините, возниклка какая-то ошибка. Попробуйте повторить ваш запрос позднее.")
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

	summary, title, interest, err := summarize(chatID)
	if err == nil {
		conversation.UpdateSummary(repo, chatID, summary, title, interest)
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

func extractGist(chatID int64, prompt string) (string, error) {
	hist := conversation.GetHistory(repo, chatID)
	var sb strings.Builder
	for _, h := range hist {
		if h.Role == "user" {
			sb.WriteString(promptUserPrefix + h.Content + "\n")
		} else {
			sb.WriteString(promptAssistantPrefix + h.Content + "\n")
		}
	}
	summary, err := aiClient.GenerateResponse(fmt.Sprintf(prompt, sb.String()))
	return summary, err
}

func summarize(chatID int64) (summary, title, interest string, err error) {
	hist := conversation.GetHistory(repo, chatID)
	var sb strings.Builder
	for _, h := range hist {
		if h.Role == "user" {
			sb.WriteString(promptUserPrefix + h.Content + "\n")
		} else {
			sb.WriteString(promptAssistantPrefix + h.Content + "\n")
		}
	}
	summary, err = aiClient.GenerateResponse(fmt.Sprintf(promptSummarizeGist, sb.String()))
	if err != nil {
		return
	}

	title, err = aiClient.GenerateResponse(fmt.Sprintf(promptSummarizeTitle, summary))
	if err != nil {
		return
	}

	interest, err = aiClient.GenerateResponse(fmt.Sprintf(promptSummarizeInterest, summary))

	return
}
