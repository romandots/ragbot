package embedding

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/pgvector/pgvector-go"
	ai "ragbot/internal/ai"
	"ragbot/internal/util"
)

// StartWorker runs a goroutine that periodically embeds unprocessed chunks.
func StartWorker(db *sql.DB, aiClient *ai.AIClient) {
	go func() {
		defer util.Recover("embedding worker")
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			process(db, aiClient)
			<-ticker.C
		}
	}()
}

func process(db *sql.DB, aiClient *ai.AIClient) {
	defer util.Recover("embedding process")
	rows, err := db.QueryContext(context.Background(), "SELECT id, content FROM chunks WHERE processed_at IS NULL LIMIT 5")
	if err != nil {
		log.Printf("embedding worker query error: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var content string
		if err := rows.Scan(&id, &content); err != nil {
			log.Printf("embedding worker scan error: %v", err)
			continue
		}
		emb, err := aiClient.GenerateEmbedding(content)
		if err != nil {
			log.Printf("embedding generation error: %v", err)
			continue
		}
		_, err = db.ExecContext(context.Background(), "UPDATE chunks SET embedding=$1, processed_at=NOW() WHERE id=$2", pgvector.NewVector(emb), id)
		if err != nil {
			log.Printf("embedding update error: %v", err)
		}
		// small delay to avoid rate limits
		time.Sleep(200 * time.Millisecond)
	}
}
