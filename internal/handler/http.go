package handler

import (
	"log"
	"net/http"

	"ragbot/internal/ai"
	"ragbot/internal/repository"
	"ragbot/internal/util"
)

type QueryRequest struct {
	Question string `json:"question"`
}

type QueryResponse struct {
	Answer string `json:"answer"`
}

func StartHTTP(repo *repository.Repository, aiClient *ai.AIClient) {
	defer util.Recover("StartHTTP")
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/chat/", ChatHandler(repo))

	log.Println("HTTP server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
