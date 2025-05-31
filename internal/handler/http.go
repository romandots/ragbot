package handler

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"ragbot/internal/ai"
)

// QueryRequest/Response модели JSON для эндпоинта /query
type QueryRequest struct {
	Question string `json:"question"`
}
type QueryResponse struct {
	Answer string `json:"answer"`
}

// StartHTTP запускает HTTP-сервер с endpoint-ами /health и /query
func StartHTTP(db *sql.DB, aiClient *ai.AIClient) {
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		var req QueryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		// Заменили вызов processQuestion на ProcessQuestion
		answer, err := ProcessQuestion(db, aiClient, req.Question)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(QueryResponse{Answer: answer})
	})

	log.Println("HTTP server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
