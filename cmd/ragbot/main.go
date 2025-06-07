package main

import (
	"context"
	"database/sql"
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
	config.LoadSettings()

	// Connect to DB
	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer database.Close()

	// Init AI client
	aiClient := ai.NewAIClient()

	// Start web server
	go handler.StartHTTP(database, aiClient)

	// Start user bot
	go bot.StartUserBot(database, aiClient, cfg.UserTelegramToken)

	// Start education sources
	startEducationSourcesHandlers(cfg, database)

	// Start embedding worker
	embedding.StartWorker(database, aiClient)

	// Block the main goroutine
	select {}
}

func startEducationSourcesHandlers(cfg *config.AppConfig, database *sql.DB) {
	ctx := context.Background()
	sources := []education.Source{
		&education.AdminSource{Token: cfg.AdminTelegramToken, AllowedIDs: cfg.AdminChatIDs},
	}
	if cfg.EducationFilePath != "" {
		sources = append(sources, &education.FileSource{Path: cfg.EducationFilePath, Interval: time.Hour})
	}
	if cfg.YandexYMLURL != "" {
		sources = append(sources, &education.YandexYMLSource{URL: cfg.YandexYMLURL, Interval: time.Hour})
	}
	if cfg.UseExternalSource {
		sources = append(sources, &education.ExternalDBSource{})
	}
	for _, s := range sources {
		s.Start(ctx, database)
	}
}
