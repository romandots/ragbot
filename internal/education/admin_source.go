package education

import (
	"context"

	"ragbot/internal/bot"
	"ragbot/internal/repository"
)

// AdminSource wraps the admin Telegram bot as a knowledge source.
type AdminSource struct {
	Token      string
	AllowedIDs []int64
}

func (a *AdminSource) Start(ctx context.Context, repo *repository.Repository) {
	go bot.StartAdminBot(repo, a.Token, a.AllowedIDs)
}
