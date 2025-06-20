package config

import (
	"log"
	"os"
	"ragbot/internal/util"
	"strconv"
	"strings"
)

type AppConfig struct {
	BaseURL             string
	DatabaseURL         string
	UseLocalModel       bool
	OpenAIAPIKey        string
	UserTelegramToken   string
	UserTelegramBotName string
	AdminTelegramToken  string
	AdminChatIDs        []int64
	EducationFilePath   string
	UseExternalSource   bool
	YandexYMLURL        string
	AmoDomain           string
	AmoAccessToken      string
	AdminUsername       string
	AdminPassword       string
	TelegramChannel     string
}

type AppSettings struct {
	Preamble                        string
	CallManagerTriggerWords         []string
	CallManagerTriggerWordsInAnswer []string
}

var Config *AppConfig
var Settings *AppSettings

func LoadConfig() *AppConfig {
	if Config != nil {
		return Config
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "localhost:8080"
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

	userBotName := os.Getenv("USER_TELEGRAM_BOT_NAME")

	adminToken := os.Getenv("ADMIN_TELEGRAM_TOKEN")
	if adminToken == "" {
		log.Fatalln("ADMIN_TELEGRAM_TOKEN not set")
	}

	eduFile := os.Getenv("EDUCATION_FILE_PATH")
	ymlURL := os.Getenv("YANDEX_YML_URL")
	amoDomain := os.Getenv("AMO_DOMAIN")
	amoToken := os.Getenv("AMO_ACCESS_TOKEN")
	telegramChannel := os.Getenv("TELEGRAM_CHANNEL")
	useExternal := false
	if os.Getenv("USE_EXTERNAL_SOURCE") == "true" {
		useExternal = true
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

	Config = &AppConfig{
		BaseURL:             baseURL,
		DatabaseURL:         url,
		UseLocalModel:       useLocal,
		OpenAIAPIKey:        apiKey,
		UserTelegramToken:   userToken,
		UserTelegramBotName: userBotName,
		AdminTelegramToken:  adminToken,
		AdminChatIDs:        adminIDs,
		EducationFilePath:   eduFile,
		UseExternalSource:   useExternal,
		YandexYMLURL:        ymlURL,
		AmoDomain:           amoDomain,
		AmoAccessToken:      amoToken,
		TelegramChannel:     telegramChannel,
		AdminUsername:       util.GetEnvString("ADMIN_USERNAME", "admin"),
		AdminPassword:       util.GetEnvString("ADMIN_PASSWORD", "secret"),
	}

	return Config
}

func LoadSettings() *AppSettings {
	if Settings != nil {
		return Settings
	}

	callManagerTriggerWords := os.Getenv("CALL_MANAGER_TRIGGER_WORDS")
	if callManagerTriggerWords == "" {
		callManagerTriggerWords = "позвать,позови,менеджер,оператор"
	}

	callManagerTriggerWordsInAnswer := os.Getenv("CALL_MANAGER_TRIGGER_WORDS_IN_ANSWER")
	if callManagerTriggerWordsInAnswer == "" {
		callManagerTriggerWordsInAnswer = "заказать звонок,позвать менеджера,вам перезвонил,оператор"
	}

	Settings = &AppSettings{
		Preamble:                        os.Getenv("PREAMBLE"),
		CallManagerTriggerWords:         strings.Split(callManagerTriggerWords, ","),
		CallManagerTriggerWordsInAnswer: strings.Split(callManagerTriggerWordsInAnswer, ","),
	}

	return Settings
}
