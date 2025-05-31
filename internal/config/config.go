package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

// Config хранит все переменные окружения приложения
type config struct {
	DatabaseURL        string
	UseLocalModel      bool
	OpenAIAPIKey       string
	UserTelegramToken  string
	AdminTelegramToken string
	AdminChatIDs       []int64
}

type settings struct {
	Preamble string
}

var Config *config
var Settings *settings

func LoadConfig() *config {
	if Config != nil {
		return Config
	}

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

	Config := &config{
		DatabaseURL:        url,
		UseLocalModel:      useLocal,
		OpenAIAPIKey:       apiKey,
		UserTelegramToken:  userToken,
		AdminTelegramToken: adminToken,
		AdminChatIDs:       adminIDs,
	}

	return Config
}

func LoadSettings() *settings {
	if Settings != nil {
		return Settings
	}

	Settings := &settings{
		Preamble: `
Ты — помощник, консультант, менеджер по продажам и администратор школы танцев «Без правил». 
Если тебя спросят, кто ты, ты должен представляться как "танцультант школы танцев «Без правил»".
Хоть ты и не человек, но ты обладаешь знаниями, достаточными для того, чтобы ответить на часто задаваемые
вопросы о школе танцев «Без правил». Если у тебя нет точного отвеа на поставленный вопрос, честно в этом признайся
и предложи пригласить в чат реального администратора школы танцев «Без правил». Твоя основная задача — 
помочь клиенту выбрать танцевальное направление, ответить на вопросы о режиме работы, расписании, адресах
студий и др. Когда у клиента больше не остается вопросов, консультации должны заканчиваться предложением посетить пробное 
занятие или приобрести абонемент на выбранное направление.
`,
	}

	return Settings
}
