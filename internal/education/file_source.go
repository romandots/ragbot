package education

import (
	"bufio"
	"context"
	"database/sql"
	"log"
	"os"
	"strings"
	"time"

	"ragbot/internal/util"
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
	defer util.Recover("FileSource.run")
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
	defer util.Recover("FileSource.process")
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
			"INSERT INTO chunks(content, source) VALUES($1, $2) ON CONFLICT (content) DO NOTHING",
			line, f.Path)
		if err != nil {
			log.Printf("file source insert error: %v", err)
		}
		log.Printf("Chunk added: %s\n", line)
	}
	if err := scanner.Err(); err != nil {
		log.Printf("file source scan error: %v", err)
	}
}
