package embedding

import (
	"context"
	"log"
	"time"

	ai "ragbot/internal/ai"
	"ragbot/internal/repository"
	"ragbot/internal/util"
)

// StartWorker runs a goroutine that periodically embeds unprocessed chunks.
func StartWorker(repo *repository.Repository, aiClient *ai.AIClient) {
	go func() {
		defer util.Recover("embedding worker")
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			process(repo, aiClient)
			<-ticker.C
		}
	}()
}

func process(repo *repository.Repository, aiClient *ai.AIClient) {
	defer util.Recover("embedding process")
	chunks, err := repo.GetUnprocessedChunks(context.Background(), 5)
	if err != nil {
		log.Printf("embedding worker query error: %v", err)
		return
	}
	for _, ch := range chunks {
		emb, err := aiClient.GenerateEmbedding(ch.Content)
		if err != nil {
			log.Printf("embedding generation error: %v", err)
			continue
		}
		if err := repo.UpdateChunkEmbedding(context.Background(), ch.ID, emb); err != nil {
			log.Printf("embedding update error: %v", err)
		}
		time.Sleep(200 * time.Millisecond)
	}
}
