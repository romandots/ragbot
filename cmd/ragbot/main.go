package main

import (
	"context"
	_ "database/sql"
	"log"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"ragbot/internal/ai"
	"ragbot/internal/bot"
	"ragbot/internal/config"
	"ragbot/internal/db"
	"ragbot/internal/education"
	"ragbot/internal/embedding"
	"ragbot/internal/handler"
	"ragbot/internal/util"
)

func main() {
	defer util.Recover("main")
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

	// Запускаем Telegram-бот и источники данных
	go bot.StartUserBot(database, aiClient, cfg.UserTelegramToken)

	ctx := context.Background()
	sources := []education.Source{
		&education.AdminSource{Token: cfg.AdminTelegramToken, AllowedIDs: cfg.AdminChatIDs},
	}
	if cfg.EducationFilePath != "" {
		sources = append(sources, &education.FileSource{Path: cfg.EducationFilePath, Interval: time.Hour})
	}
	if cfg.UseExternalSource {
		sources = append(sources, &education.ExternalDBSource{})
	}
	for _, s := range sources {
		s.Start(ctx, database)
	}

	embedding.StartWorker(database, aiClient)

	select {} // блокируем main
}
