package db

import (
	"database/sql"
	"embed"
	"errors"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

// Connect открывает соединение с Postgres
func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}

	goose.SetBaseFS(migrationsFS)
	goose.SetDialect("postgres")
	if err := goose.Up(db, "migrations"); err != nil && !errors.Is(err, goose.ErrNoNextVersion) {
		db.Close()
		return nil, err
	}

	// Здесь можно настроить пул, ping и т.д.
	return db, nil
}
