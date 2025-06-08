package handler

import (
	"html/template"
	"net/http"
	"strings"

	"ragbot/internal/conversation"
	"ragbot/internal/repository"
	"ragbot/internal/util"
)

var chatTemplate = template.Must(template.New("chat").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script src="https://cdn.tailwindcss.com"></script>
<title>{{.Title}}</title>
</head>
<body class="bg-gray-100">
<div class="max-w-2xl mx-auto p-4">
    {{if .Name}}
    <div class="bg-yellow-100 p-3 rounded shadow mb-4">
        <h2 class="font-bold mb-2">Контакт</h2>
        <p>{{.Name}}</p>
                {{if .Phone}}
                <p>{{.Phone}}</p>
                {{end}}
    </div>
    {{end}}
    {{if .Summary}}
    <div class="bg-yellow-100 p-3 rounded shadow mb-4">
        <h2 class="font-bold mb-2">Суть обращения</h2>
        <p>{{.Summary}}</p>
    </div>
    {{end}}
    <div class="bg-white p-4 rounded shadow">
        {{range .History}}
        <div class="mb-2">
            {{if eq .Role "user"}}
            <span class="font-semibold">{{$.Name}}:</span>
            {{else}}
            <span class="text-blue-600">Ассистент:</span>
            {{end}}
            <span>{{.Content}}</span>
        </div>
        {{end}}
    </div>
</div>
</body>
</html>`))

func ChatHandler(repo *repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer util.Recover("ChatHandler")
		uuid := strings.TrimPrefix(strings.TrimSuffix(r.URL.Path, "/"), "/chat/")
		info, err := conversation.GetChatInfoByUUID(repo, uuid)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		history := conversation.GetFullHistory(repo, info.ChatID)
		data := struct {
			Title   string
			Summary string
			Name    string
			Phone   string
			History []conversation.HistoryItem
		}{
			Title:   info.Title.String,
			Summary: info.Summary.String,
			Name:    info.Name.String,
			Phone:   info.Phone.String,
			History: history,
		}
		chatTemplate.Execute(w, data)
	}
}
