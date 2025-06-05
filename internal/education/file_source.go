package education

import (
	"bufio"
	"context"
	"database/sql"
	"log"
	"os"
	"strings"
	"time"
)

// FileSource loads chunks from a text file on a schedule.
type FileSource struct {
	Path     string
	Interval time.Duration
}

func (f *FileSource) Start(ctx context.Context, db *sql.DB) {
	go f.run(ctx, db)
}

func (f *FileSource) run(ctx context.Context, db *sql.DB) {
	f.process(db)
	ticker := time.NewTicker(f.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			f.process(db)
		}
	}
}

func (f *FileSource) process(db *sql.DB) {
	file, err := os.Open(f.Path)
	if err != nil {
		log.Printf("file source open error: %v", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		_, err := db.ExecContext(context.Background(),
			"INSERT INTO chunks(content) VALUES($1) ON CONFLICT DO NOTHING",
			line,
		)
		if err != nil {
			log.Printf("file source insert error: %v", err)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("file source scan error: %v", err)
	}
}
