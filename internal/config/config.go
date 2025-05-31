package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

// Config хранит все переменные окружения приложения
type Config struct {
	DatabaseURL        string
	UseLocalModel      bool
	OpenAIAPIKey       string
	UserTelegramToken  string
	AdminTelegramToken string
	AdminChatIDs       []int64
}

func LoadConfig() *Config {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		log.Fatalln("DATABASE_URL not set")
	}

	useLocal := false
	if os.Getenv("USE_LOCAL_MODEL") == "true" {
		useLocal = true
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if !useLocal && apiKey == "" {
		log.Fatalln("OPENAI_API_KEY not set when using external GPT")
	}

	userToken := os.Getenv("USER_TELEGRAM_TOKEN")
	if userToken == "" {
		log.Fatalln("USER_TELEGRAM_TOKEN not set")
	}

	adminToken := os.Getenv("ADMIN_TELEGRAM_TOKEN")
	if adminToken == "" {
		log.Fatalln("ADMIN_TELEGRAM_TOKEN not set")
	}

	// Читаем ADMIN_CHAT_IDS как строку "id1,id2,id3"
	adminIDsEnv := os.Getenv("ADMIN_CHAT_IDS")
	var adminIDs []int64
	for _, part := range strings.Split(adminIDsEnv, ",") {
		if part = strings.TrimSpace(part); part != "" {
			id, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				log.Fatalf("Invalid ADMIN_CHAT_IDS value: %v", err)
			}
			adminIDs = append(adminIDs, id)
		}
	}

	return &Config{
		DatabaseURL:        url,
		UseLocalModel:      useLocal,
		OpenAIAPIKey:       apiKey,
		UserTelegramToken:  userToken,
		AdminTelegramToken: adminToken,
		AdminChatIDs:       adminIDs,
	}
}
