package handler

import (
	"database/sql"
	"html/template"
	"net/http"
	"os"

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
            <tr><td class="px-4 py-2">Уникальные чаты</td><td class="px-4 py-2">{{.UniqueChats}}</td></tr>
            <tr class="border-t border-gray-200 dark:border-gray-700"><td class="px-4 py-2">Лиды</td><td class="px-4 py-2">{{.Deals}}</td></tr>
            <tr class="border-t border-gray-200 dark:border-gray-700"><td class="px-4 py-2">Конверсия</td><td class="px-4 py-2">{{printf "%.2f" .Conversion}}%</td></tr>
            <tr class="border-t border-gray-200 dark:border-gray-700"><td class="px-4 py-2">/rasp</td><td class="px-4 py-2">{{.RaspCount}}</td></tr>
            <tr class="border-t border-gray-200 dark:border-gray-700"><td class="px-4 py-2">/address</td><td class="px-4 py-2">{{.AddrCount}}</td></tr>
            <tr class="border-t border-gray-200 dark:border-gray-700"><td class="px-4 py-2">/prices</td><td class="px-4 py-2">{{.PriceCount}}</td></tr>
        </tbody>
    </table>
    <table class="min-w-full bg-white dark:bg-gray-800 rounded shadow divide-y divide-gray-200 dark:divide-gray-700">
        <thead class="bg-gray-200 dark:bg-gray-700">
            <tr>
                <th class="px-4 py-2 text-left">Chat ID</th>
                <th class="px-4 py-2 text-left">Username</th>
                <th class="px-4 py-2 text-left">Сообщений до лида</th>
            </tr>
        </thead>
        <tbody>
            {{range .MsgCounts}}
            <tr class="border-t border-gray-200 dark:border-gray-700">
                <td class="px-4 py-2">{{.ChatID}}</td>
                <td class="px-4 py-2">{{.Username.String}}</td>
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
		ChatID   int64
		Username sql.NullString
		Count    int
	}
	return func(w http.ResponseWriter, r *http.Request) {
		defer util.Recover("StatsHandler")
		user, pass, ok := r.BasicAuth()
		if !ok || user != os.Getenv("STATS_USER") || pass != os.Getenv("STATS_PASS") {
			w.Header().Set("WWW-Authenticate", "Basic realm=restricted")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := r.Context()
		uniqueChats, _ := repo.CountUniqueChats(ctx)
		deals, _ := repo.CountDeals(ctx)
		conv := 0.0
		if uniqueChats > 0 {
			conv = float64(deals) / float64(uniqueChats) * 100
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
			UniqueChats int
			Deals       int
			Conversion  float64
			RaspCount   int
			AddrCount   int
			PriceCount  int
			MsgCounts   []msgCount
		}{
			UniqueChats: uniqueChats,
			Deals:       deals,
			Conversion:  conv,
			RaspCount:   raspCount,
			AddrCount:   addrCount,
			PriceCount:  priceCount,
			MsgCounts:   msgCounts,
		}
		statsTemplate.Execute(w, data)
	}
}
