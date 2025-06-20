package handler

import (
	"database/sql"
	"html/template"
	"net/http"
	"ragbot/internal/repository"
	"ragbot/internal/util"
)

var statsTemplate = template.Must(template.New("stats").Parse(`<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <script src="https://cdn.tailwindcss.com"></script>
    <script>tailwind.config={darkMode:'media'}</script>
    <title>Статистика</title>
</head>
<body class="min-h-screen bg-gray-100 dark:bg-gray-900 text-gray-900 dark:text-gray-100">
<div class="max-w-3xl mx-auto p-4 space-y-6">
    <h1 class="text-2xl font-bold">Статистика</h1>
    <table class="min-w-full bg-white dark:bg-gray-800 rounded shadow divide-y divide-gray-200 dark:divide-gray-700">
        <thead class="bg-gray-200 dark:bg-gray-700">
            <tr>
                <th class="px-4 py-2 text-left">Показатель</th>
                <th class="px-4 py-2 text-left">Значение</th>
            </tr>
        </thead>
        <tbody>
            <tr><td class="px-4 py-2">Переходы в бота</td><td class="px-4 py-2">{{.Visits}}</td></tr>
            <tr class="border-t border-gray-200 dark:border-gray-700"><td class="px-4 py-2">Диалоги с ботом</td><td class="px-4 py-2">{{.UniqueChats}}</td></tr>
            <tr class="border-t border-gray-200 dark:border-gray-700"><td class="px-4 py-2">Лиды</td><td class="px-4 py-2">{{.Deals}}</td></tr>
            <tr class="border-t-2 border-gray-200 dark:border-gray-700"><td class="px-4 py-2">Конверсия заходов в беседы</td><td class="px-4 py-2">{{printf "%.2f" .VisitToChatConversion}}%</td></tr>
            <tr class="border-t border-gray-200 dark:border-gray-700"><td class="px-4 py-2">Конверсия бесед в лиды</td><td class="px-4 py-2">{{printf "%.2f" .ChatToLeadConversion}}%</td></tr>
            <tr class="border-t border-gray-200 dark:border-gray-700"><td class="px-4 py-2">Конверсия переходов в лиды</td><td class="px-4 py-2">{{printf "%.2f" .VisitToLeadConversion}}%</td></tr>
            <tr class="border-t-2 border-gray-200 dark:border-gray-700"><td class="px-4 py-2">Нажатий кнопки «Расписание»</td><td class="px-4 py-2">{{.RaspCount}}</td></tr>
            <tr class="border-t border-gray-200 dark:border-gray-700"><td class="px-4 py-2">Нажатий кнопки «Адреса»</td><td class="px-4 py-2">{{.AddrCount}}</td></tr>
            <tr class="border-t border-gray-200 dark:border-gray-700"><td class="px-4 py-2">Нажатий кнопки «Цены»</td><td class="px-4 py-2">{{.PriceCount}}</td></tr>
        </tbody>
    </table>
    <table class="min-w-full bg-white dark:bg-gray-800 rounded shadow divide-y divide-gray-200 dark:divide-gray-700">
        <thead class="bg-gray-200 dark:bg-gray-700">
            <tr>
                <th class="px-4 py-2 text-left">Chat ID</th>
                <th class="px-4 py-2 text-left">Пользователь</th>
                <th class="px-4 py-2 text-left">Сообщений до лида</th>
            </tr>
        </thead>
        <tbody>
            {{range .MsgCounts}}
            <tr class="border-t border-gray-200 dark:border-gray-700">
                <td class="px-4 py-2"><a class="text-blue-600 dark:text-blue-400" href="/chat/{{.ID}}">{{.ChatID}}</a></td>
				<td class="px-4 py-2">
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
                <td class="px-4 py-2">{{.Count}}</td>
            </tr>
            {{end}}
        </tbody>
    </table>
</div>
</body>
</html>`))

func StatsHandler(repo *repository.Repository) http.HandlerFunc {
	type msgCount struct {
		ID       string
		ChatID   int64
		Username sql.NullString
		Name     sql.NullString
		Count    int
	}
	return func(w http.ResponseWriter, r *http.Request) {
		defer util.Recover("StatsHandler")
		if !authorize(w, r) {
			return
		}
		ctx := r.Context()
		visits, _ := repo.CountUniqueVisits(ctx)
		uniqueChats, _ := repo.CountUniqueChats(ctx)
		deals, _ := repo.CountDeals(ctx)
		chatToLeadConv := 0.0
		visitToChatConv := 0.0
		visitToLeadConv := 0.0
		if uniqueChats > 0 {
			chatToLeadConv = float64(deals) / float64(uniqueChats) * 100
		}
		if visits > 0 {
			visitToChatConv = float64(uniqueChats) / float64(visits) * 100
			visitToLeadConv = float64(deals) / float64(visits) * 100
		}
		raspCount, _ := repo.CountCommandUsage(ctx, "/rasp")
		addrCount, _ := repo.CountCommandUsage(ctx, "/address")
		priceCount, _ := repo.CountCommandUsage(ctx, "/prices")
		msgCountsRaw, _ := repo.MessageCountsBeforeDeal(ctx)
		var msgCounts []msgCount
		for _, m := range msgCountsRaw {
			msgCounts = append(msgCounts, msgCount(m))
		}
		data := struct {
			Visits                int
			UniqueChats           int
			VisitToChatConversion float64
			Deals                 int
			ChatToLeadConversion  float64
			VisitToLeadConversion float64
			RaspCount             int
			AddrCount             int
			PriceCount            int
			MsgCounts             []msgCount
		}{
			Visits:                visits,
			UniqueChats:           uniqueChats,
			VisitToChatConversion: visitToChatConv,
			Deals:                 deals,
			ChatToLeadConversion:  chatToLeadConv,
			VisitToLeadConversion: visitToLeadConv,
			RaspCount:             raspCount,
			AddrCount:             addrCount,
			PriceCount:            priceCount,
			MsgCounts:             msgCounts,
		}
		statsTemplate.Execute(w, data)
	}
}
