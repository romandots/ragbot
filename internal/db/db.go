package db

import (
	"database/sql"
)

// Connect открывает соединение с Postgres
func Connect(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	// Здесь можно настроить пул, ping и т.д.
	return db, nil
}
