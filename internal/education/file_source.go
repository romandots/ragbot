package education

import (
	"bufio"
	"context"
	"log"
	"os"
	"strings"
	"time"

	"ragbot/internal/repository"
	"ragbot/internal/util"
)

const fileSource = "file"

// FileSource loads chunks from a text file on a schedule.
type FileSource struct {
	Path     string
	Interval time.Duration
}

func (f *FileSource) Start(ctx context.Context, repo *repository.Repository) {
	go f.run(ctx, repo)
}

func (f *FileSource) run(ctx context.Context, repo *repository.Repository) {
	defer util.Recover("FileSource.run")
	f.process(repo)
	ticker := time.NewTicker(f.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			f.process(repo)
		}
	}
}

func (f *FileSource) process(repo *repository.Repository) {
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
		id, err := repo.AddChunk(context.Background(), line, fileSource)
		if err != nil {
			log.Printf("file source insert error: %v", err)
			continue
		}
		if id != 0 {
			log.Printf("Chunk added #%d: %s", id, line)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("file source scan error: %v", err)
	}
}
