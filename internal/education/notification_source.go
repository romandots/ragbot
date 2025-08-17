package education

import (
	"context"

	"ragbot/internal/bot"
	"ragbot/internal/repository"
)

// NotificationSource wraps the notification Telegram bot as a notification source.
type NotificationSource struct {
	Token      string
	AllowedIDs []int64
}

func (n *NotificationSource) Start(ctx context.Context, repo *repository.Repository) {
	go bot.StartNotificationBot(n.Token, n.AllowedIDs)
}