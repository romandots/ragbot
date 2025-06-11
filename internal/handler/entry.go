package handler

import (
	"fmt"
	"net/http"
	"ragbot/internal/config"
	"ragbot/internal/repository"
)

func HandleEntry(repo *repository.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
	}
}
