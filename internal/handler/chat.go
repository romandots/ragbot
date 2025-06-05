package handler

import (
	"database/sql"
	"html/template"
	"net/http"
	"strings"

	"ragbot/internal/conversation"
)

var chatTemplate = template.Must(template.New("chat").Parse(`<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script src="https://cdn.tailwindcss.com"></script>
<title>Chat</title>
</head>
<body class="bg-gray-100">
<div class="max-w-2xl mx-auto p-4">
    {{if .Summary}}
    <div class="bg-yellow-100 p-3 rounded shadow mb-4">
        <h2 class="font-bold mb-2">Summary</h2>
        <p>{{.Summary}}</p>
    </div>
    {{end}}
    <div class="bg-white p-4 rounded shadow">
        {{range .History}}
        <div class="mb-2">
            {{if eq .Role "user"}}
            <span class="font-semibold">User:</span>
            {{else}}
            <span class="text-blue-600">Assistant:</span>
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
		uuid := strings.TrimPrefix(r.URL.Path, "/chat/")
		info, err := conversation.GetChatInfoByUUID(db, uuid)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		history := conversation.GetFullHistory(db, info.ChatID)
		data := struct {
			Summary string
			History []conversation.HistoryItem
		}{
			Summary: info.Summary.String,
			History: history,
		}
		chatTemplate.Execute(w, data)
	}
}
