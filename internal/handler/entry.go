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
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprintf(w, "<!DOCTYPE html><html><head><meta charset=\"utf-8\"><title>Redirect</title><script>window.location.replace('%s');</script></head><body><noscript><a href=\"%s\">Перейти в чат с ассистентом</a></noscript></body></html>", url, url)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
