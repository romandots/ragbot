package handler

import (
	"fmt"
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

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		if config.Config.UserTelegramBotName != "" {
			url := fmt.Sprintf(telegramWebUrlFormat, config.Config.UserTelegramBotName)
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		}

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		if err := repo.Ping(ctx); err != nil {
			log.Printf("Health check failed: %v", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("Database connection error: " + err.Error()))
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	http.HandleFunc("/chat/", ChatHandler(repo))

	log.Println("HTTP server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
