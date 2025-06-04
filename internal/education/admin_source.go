package education

import (
	"context"
	"database/sql"
	"ragbot/internal/bot"
)

// AdminSource wraps the admin Telegram bot as a knowledge source.
type AdminSource struct {
	Token      string
	AllowedIDs []int64
}

func (a *AdminSource) Start(ctx context.Context, db *sql.DB) {
	// bot.StartAdminBot blocks, so run it in a goroutine
	go bot.StartAdminBot(db, a.Token, a.AllowedIDs)
}
