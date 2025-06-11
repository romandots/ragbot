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
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <script src="https://cdn.tailwindcss.com"></script>
    <script>tailwind.config={darkMode:'media'}</script>
    <title>{{.Title}}</title>
</head>
<body class="min-h-screen bg-gray-100 dark:bg-gray-900 text-gray-900 dark:text-gray-100">
<div class="max-w-2xl mx-auto p-4 space-y-4">
    <h1 class="text-2xl font-bold">{{.Title}}</h1>
    {{if .Name}}
    <div class="p-4 rounded shadow">
        <h2 class="font-semibold mb-1">Контакт</h2>
        <p>{{.Name}}</p>
        {{if .Phone}}
        <p>{{.Phone}}</p>
        {{end}}
        {{if .Username}}
        <p><a href="https://t.me/{{.Username}}">@{{.Username}}</a></p>
        {{end}}
    </div>
    {{end}}
    {{if .Summary}}
    <div class="bg-yellow-50 dark:bg-yellow-900/20 p-4 rounded shadow">
        <h2 class="font-semibold mb-1">Суть обращения</h2>
        <p>{{.Summary}}</p>
    </div>
    {{end}}
    <div class="bg-white dark:bg-gray-800 p-4 rounded shadow space-y-2">
        {{range .History}}
			{{if eq .Role "user"}}
			<div class="bg-blue-100 dark:bg-blue-900/40 rounded p-3 mb-2">
				<span class="font-semibold">@{{$.Username}}:</span>
				<span>{{.Content}}</span>
			</div>
			{{else}}
			<div class="bg-gray-100 dark:bg-gray-900/40 rounded p-3 mb-2">
				<span class="text-blue-600 dark:text-blue-400">Ассистент:</span>
				<span>{{.Content}}</span>
			</div>
			{{end}}
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
			Title    string
			Summary  string
			Name     string
			Username string
			Phone    string
			History  []conversation.HistoryItem
		}{
			Title:    info.Title.String,
			Summary:  info.Summary.String,
			Name:     info.Name.String,
			Username: info.Username.String,
			Phone:    info.Phone.String,
			History:  history,
		}
		chatTemplate.Execute(w, data)
	}
}
