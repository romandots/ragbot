package handler

import (
	"database/sql"
	"html/template"
	"net/http"
	"strings"

	"ragbot/internal/conversation"
	"ragbot/internal/util"
)

var chatTemplate = template.Must(template.New("chat").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script src="https://cdn.tailwindcss.com"></script>
<title>Чат с {{.Name}}</title>
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
            <span class="font-semibold">{{.Name}}:</span>
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

func ChatHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer util.Recover("ChatHandler")
		uuid := strings.TrimPrefix(r.URL.Path, "/chat/")
		info, err := conversation.GetChatInfoByUUID(db, uuid)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		history := conversation.GetFullHistory(db, info.ChatID)
		data := struct {
			Summary string
			Name    string
			Phone   string
			History []conversation.HistoryItem
		}{
			Summary: info.Summary.String,
			Name:    info.Name.String,
			Phone:   info.Phone.String,
			History: history,
		}
		chatTemplate.Execute(w, data)
	}
}
