package handler

import (
	"html/template"
	"net/http"
	"os"
	"strconv"

	"ragbot/internal/repository"
	"ragbot/internal/util"
)

var chatsTemplate = template.Must(template.New("chats").Parse(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <script src="https://cdn.tailwindcss.com"></script>
    <title>Чаты</title>
</head>
<body class="bg-gray-100">
<div class="max-w-4xl mx-auto p-4">
    <table class="min-w-full bg-white rounded shadow">
        <thead class="bg-gray-200">
        <tr>
            <th class="px-2 py-1 text-left">Дата</th>
            <th class="px-2 py-1 text-left">Чат</th>
            <th class="px-2 py-1 text-left">Пользователь</th>
            <th class="px-2 py-1 text-left">Лид</th>
            <th class="px-2 py-1 text-left">Последнее сообщение</th>
        </tr>
        </thead>
        <tbody>
        {{range .Chats}}
        <tr class="border-t">
            <td class="px-2 py-1">{{.LastAt.Format "2006-01-02 15:04"}}</td>
            <td class="px-2 py-1"><a class="text-blue-600" href="/chat/{{.ID}}">{{.Title.String}}</a></td>
            <td class="px-2 py-1">{{.Name.String}} {{if .Username.Valid}}(@{{.Username.String}}){{end}}</td>
            <td class="px-2 py-1">{{if .HasDeal}}✔{{else}}—{{end}}</td>
            <td class="px-2 py-1">{{.LastMsg}}</td>
        </tr>
        {{end}}
        </tbody>
    </table>
    <div class="mt-4 flex justify-between">
        {{if .HasPrev}}<a class="text-blue-600" href="?page={{.PrevPage}}">Предыдущая</a>{{else}}<span></span>{{end}}
        {{if .HasNext}}<a class="text-blue-600" href="?page={{.NextPage}}">Следующая</a>{{end}}
    </div>
</div>
</body>
</html>`))

func ChatsHandler(repo *repository.Repository) http.HandlerFunc {
	const limit = 20
	return func(w http.ResponseWriter, r *http.Request) {
		defer util.Recover("ChatsHandler")

		user, pass, ok := r.BasicAuth()
		if !ok || user != os.Getenv("STATS_USER") || pass != os.Getenv("STATS_PASS") {
			w.Header().Set("WWW-Authenticate", "Basic realm=restricted")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 1 {
			page = 1
		}
		chats, err := repo.ListChats(r.Context(), limit+1, (page-1)*limit)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		hasNext := len(chats) > limit
		if hasNext {
			chats = chats[:limit]
		}
		data := struct {
			Chats    []repository.ChatSummary
			PrevPage int
			NextPage int
			HasPrev  bool
			HasNext  bool
		}{
			Chats:    chats,
			PrevPage: page - 1,
			NextPage: page + 1,
			HasPrev:  page > 1,
			HasNext:  hasNext,
		}
		chatsTemplate.Execute(w, data)
	}
}
