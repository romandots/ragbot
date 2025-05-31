package main

import (
	_ "database/sql"
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"

	"ragbot/internal/ai"
	"ragbot/internal/bot"
	"ragbot/internal/config"
	"ragbot/internal/db"
	"ragbot/internal/handler"
)

func main() {
	cfg := config.LoadConfig()

	// Подключаем БД
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer database.Close()

	// Создаём AIClient (по стратегии)
	aiClient := ai.NewAIClient() // берёт USE_LOCAL_MODEL, OPENAI_API_KEY из env

	// Запускаем HTTP-сервер
	go handler.StartHTTP(database, aiClient)

	// Запускаем Telegram-ботов
	go bot.StartUserBot(database, aiClient, cfg.UserTelegramToken)
	go bot.StartAdminBot(database, aiClient, cfg.AdminTelegramToken, cfg.AdminChatIDs)

	select {} // блокируем main
}
