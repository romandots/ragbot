package handler

import (
	"log"
	"net/http"
	"ragbot/internal/config"

	"ragbot/internal/ai"
	"ragbot/internal/repository"
	"ragbot/internal/util"
)

const telegramWebUrlFormat = "https://t.me/%s"

type QueryRequest struct {
	Question string `json:"question"`
}

type QueryResponse struct {
	Answer string `json:"answer"`
}

func StartHTTP(repo *repository.Repository, aiClient *ai.AIClient) {
	defer util.Recover("StartHTTP")

	http.HandleFunc("/", HandleEntry(repo))
	http.HandleFunc("/health", HandleHealth(repo))
	http.HandleFunc("/chat/", ChatHandler(repo))
	http.HandleFunc("/chats", ChatsHandler(repo))
	http.HandleFunc("/stats", StatsHandler(repo))

	log.Println("HTTP server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func authorize(w http.ResponseWriter, r *http.Request) bool {
	user, pass, ok := r.BasicAuth()
	if !ok || user != config.Config.AdminUsername || pass != config.Config.AdminPassword {
		w.Header().Set("WWW-Authenticate", "Basic realm=restricted")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}
