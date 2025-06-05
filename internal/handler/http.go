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

// StartHTTP запускает HTTP-сервер с endpoint-ами /health, /query и /chat/{uuid}
func StartHTTP(db *sql.DB, aiClient *ai.AIClient) {
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/query", func(w http.ResponseWriter, r *http.Request) {
		var req QueryRequest
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		answer, err := ProcessQuestionWithHistory(db, aiClient, 0, req.Question)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(QueryResponse{Answer: answer})
	})

	http.HandleFunc("/chat/", ChatHandler(db))

	log.Println("HTTP server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
