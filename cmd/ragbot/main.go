package main

import (
	"context"
	"log"
	"ragbot/internal/tansultant"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"ragbot/internal/ai"
	"ragbot/internal/bot"
	"ragbot/internal/config"
	"ragbot/internal/db"
	"ragbot/internal/education"
	"ragbot/internal/embedding"
	"ragbot/internal/handler"
	"ragbot/internal/repository"
	"ragbot/internal/util"
)

func main() {
	defer util.Recover("main")
	cfg := config.LoadConfig()
	config.LoadSettings()

	database, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("DB connection error: %v", err)
	}
	defer database.Close()
	repo := repository.New(database)

	aiClient := ai.NewAIClient()
	tansClient := tansultant.NewClient()

	go handler.StartHTTP(repo, aiClient)

	go bot.StartUserBot(repo, aiClient, tansClient, cfg.UserTelegramToken)

	startEducationSourcesHandlers(cfg, repo)

	embedding.StartWorker(repo, aiClient)

	select {}
}

func startEducationSourcesHandlers(cfg *config.AppConfig, repo *repository.Repository) {
	ctx := context.Background()
	sources := []education.Source{
		&education.AdminSource{Token: cfg.AdminTelegramToken, AllowedIDs: cfg.AdminChatIDs},
		&education.NotificationSource{Token: cfg.NotificationTelegramToken, AllowedIDs: cfg.NotificationChatIDs},
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
		s.Start(ctx, repo)
	}
}
