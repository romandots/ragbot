package handler

import (
	"fmt"
	"log"
	"net/http"
	"os"

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

		ip := r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr
		}
		ua := r.UserAgent()
		ref := r.Referer()
		q := r.URL.Query()
		repo.AddVisit(r.Context(), ip, ua, ref,
			q.Get("utm_source"), q.Get("utm_medium"), q.Get("utm_campaign"), q.Get("utm_content"), q.Get("utm_term"))

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
	http.HandleFunc("/chats", ChatsHandler(repo))

	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != os.Getenv("STATS_USER") || pass != os.Getenv("STATS_PASS") {
			w.Header().Set("WWW-Authenticate", "Basic realm=restricted")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := r.Context()
		uniqueChats, _ := repo.CountUniqueChats(ctx)
		deals, _ := repo.CountDeals(ctx)
		conv := 0.0
		if uniqueChats > 0 {
			conv = float64(deals) / float64(uniqueChats) * 100
		}
		raspCount, _ := repo.CountCommandUsage(ctx, "/rasp")
		addrCount, _ := repo.CountCommandUsage(ctx, "/address")
		priceCount, _ := repo.CountCommandUsage(ctx, "/prices")

		fmt.Fprintf(w, "Unique chats: %d\n", uniqueChats)
		fmt.Fprintf(w, "Deals: %d\n", deals)
		fmt.Fprintf(w, "Conversion: %.2f%%\n", conv)
		fmt.Fprintf(w, "/rasp: %d\n", raspCount)
		fmt.Fprintf(w, "/address: %d\n", addrCount)
		fmt.Fprintf(w, "/prices: %d\n", priceCount)

		msgCounts, _ := repo.MessageCountsBeforeDeal(ctx)
		fmt.Fprintf(w, "Messages before deal:\n")
		for _, m := range msgCounts {
			fmt.Fprintf(w, "chat %d (%s): %d messages\n", m.ChatID, m.Username.String, m.Count)
		}
	})

	log.Println("HTTP server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
