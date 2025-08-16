package handler

import (
	"html/template"
	"net/http"
	"strconv"

	"ragbot/internal/repository"
	"ragbot/internal/util"
)

var chatsTemplate = template.Must(template.New("chats").Parse(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <script src="https://cdn.tailwindcss.com"></script>
    <script>tailwind.config={darkMode:'media'}</script>
    <title>Чаты</title>
</head>
<body class="min-h-screen bg-gray-100 dark:bg-gray-900 text-gray-900 dark:text-gray-100">
<div class="max-w-6xl mx-auto p-4 space-y-4">
    <h1 class="text-2xl font-bold">Чаты</h1>
    <div class="overflow-x-auto">
    <table class="min-w-full bg-white dark:bg-gray-800 rounded shadow">
        <thead class="bg-gray-200 dark:bg-gray-700">
        <tr>
            <th class="px-4 py-2 text-left">Дата</th>
            <th class="px-4 py-2 text-left">Пользователь</th>
            <th class="px-4 py-2 text-left">Чат</th>
            <th class="px-4 py-2 text-left">Лид</th>
            <th class="px-4 py-2 text-left">Последнее сообщение</th>
        </tr>
        </thead>
        <tbody>
        {{range .Chats}}
        <tr class="border-t border-gray-200 dark:border-gray-700">
            <td class="px-4 py-2 whitespace-nowrap"><a class="text-blue-600 dark:text-blue-400" href="/chat/{{.ID}}">{{.LastAt.Format "2006-01-02 15:04"}}</a></td>
            <td class="px-4 py-2 whitespace-nowrap">
			{{if .Name.Valid}}
				{{.Name.String}}
				{{if .Username.Valid}}
					(@{{.Username.String}})
				{{end}}
			{{else}}
				{{if .Username.Valid}}
					@{{.Username.String}}
				{{end}}
			{{end}}
			</td>
            <td class="px-4 py-2 whitespace-nowrap">{{.Title.String}}</td>
            <td class="px-4 py-2 whitespace-nowrap">{{if .HasDeal}}✔{{else}}—{{end}}</td>
            <td class="px-4 py-2">{{.LastMsg}}</td>
        </tr>
        {{end}}
        </tbody>
    </table>
    </div>
    <div class="flex justify-between">
        {{if .HasPrev}}<a class="text-blue-600 dark:text-blue-400" href="?page={{.PrevPage}}">Предыдущая</a>{{else}}<span></span>{{end}}
        {{if .HasNext}}<a class="text-blue-600 dark:text-blue-400" href="?page={{.NextPage}}">Следующая</a>{{end}}
    </div>
</div>
</body>
</html>`))

func ChatsHandler(repo *repository.Repository) http.HandlerFunc {
	const limit = 20
	return func(w http.ResponseWriter, r *http.Request) {
		defer util.Recover("ChatsHandler")

		if !authorize(w, r) {
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
