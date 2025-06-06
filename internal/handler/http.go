package handler

import (
	"database/sql"
	"log"
	"net/http"

	"ragbot/internal/ai"
	"ragbot/internal/util"
)

// QueryRequest/Response модели JSON для эндпоинта /query
type QueryRequest struct {
	Question string `json:"question"`
}
type QueryResponse struct {
	Answer string `json:"answer"`
}

// StartHTTP запускает HTTP-сервер с endpoint-ами /health, /query и /chat/{uuid}
func StartHTTP(db *sql.DB, aiClient *ai.AIClient) {
	defer util.Recover("StartHTTP")
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/chat/", ChatHandler(db))

	log.Println("HTTP server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
