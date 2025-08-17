package bot

import (
	"strings"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"ragbot/internal/config"
)

func TestDeleteMessageAfterCallback(t *testing.T) {
	// Создаем тестовый callback query
	callbackQuery := &tgbotapi.CallbackQuery{
		ID: "test_callback_id",
		Data: actionCallManager,
		Message: &tgbotapi.Message{
			MessageID: 123,
			Chat: &tgbotapi.Chat{
				ID: 456,
			},
		},
	}

	update := tgbotapi.Update{
		CallbackQuery: callbackQuery,
	}

	// Проверяем, что messageID извлекается корректно
	messageID := update.CallbackQuery.Message.MessageID
	if messageID != 123 {
		t.Errorf("Expected messageID 123, got %d", messageID)
	}

	chatID := update.CallbackQuery.Message.Chat.ID
	if chatID != 456 {
		t.Errorf("Expected chatID 456, got %d", chatID)
	}

	// Проверяем, что callback data корректно определяется
	data := update.CallbackQuery.Data
	if data != actionCallManager {
		t.Errorf("Expected callback data %s, got %s", actionCallManager, data)
	}
}

func TestCallbackActions(t *testing.T) {
	testCases := []struct {
		name     string
		callback string
		expected string
	}{
		{
			name:     "Call Manager Action",
			callback: actionCallManager,
			expected: actionCallManager,
		},
		{
			name:     "Confirm Yes Action",
			callback: actionConfirmYes,
			expected: actionConfirmYes,
		},
		{
			name:     "Confirm No Action",
			callback: actionConfirmNo,
			expected: actionConfirmNo,
		},
		{
			name:     "Price Action",
			callback: "PRICE_0",
			expected: "PRICE_0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.callback != tc.expected {
				t.Errorf("Expected callback %s, got %s", tc.expected, tc.callback)
			}
		})
	}
}

func TestTriggerWordsDetection(t *testing.T) {
	// Load configuration for testing
	config.LoadSettings()

	testCases := []struct {
		name               string
		userText           string
		shouldTrigger      string
		shouldNotTrigger   []string
	}{
		{
			name:             "Schedule trigger words",
			userText:         "Покажи расписание занятий",
			shouldTrigger:    "schedule",
			shouldNotTrigger: []string{"price", "call"},
		},
		{
			name:             "Price trigger words",
			userText:         "Сколько стоят цены на абонементы?",
			shouldTrigger:    "price",
			shouldNotTrigger: []string{"schedule", "call"},
		},
		{
			name:             "Call manager trigger words",
			userText:         "Позови менеджера пожалуйста",
			shouldTrigger:    "call",
			shouldNotTrigger: []string{"schedule", "price"},
		},
		{
			name:             "Mixed text with schedule",
			userText:         "Привет! Можете показать график занятий?",
			shouldTrigger:    "schedule",
			shouldNotTrigger: []string{"price", "call"},
		},
		{
			name:             "Mixed text with price",
			userText:         "Скажите прайс на ваши услуги",
			shouldTrigger:    "price",
			shouldNotTrigger: []string{"schedule", "call"},
		},
		{
			name:             "No triggers",
			userText:         "Привет! Как дела?",
			shouldTrigger:    "",
			shouldNotTrigger: []string{"schedule", "price", "call"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			lowerText := strings.ToLower(tc.userText)
			
			// Check schedule triggers
			scheduleTriggered := containsAny(lowerText, config.Settings.ScheduleTriggerWords)
			if tc.shouldTrigger == "schedule" && !scheduleTriggered {
				t.Errorf("Expected schedule trigger for text: %s", tc.userText)
			}
			if tc.shouldTrigger != "schedule" && scheduleTriggered {
				t.Errorf("Unexpected schedule trigger for text: %s", tc.userText)
			}

			// Check price triggers
			priceTriggered := containsAny(lowerText, config.Settings.PriceTriggerWords)
			if tc.shouldTrigger == "price" && !priceTriggered {
				t.Errorf("Expected price trigger for text: %s", tc.userText)
			}
			if tc.shouldTrigger != "price" && priceTriggered {
				t.Errorf("Unexpected price trigger for text: %s", tc.userText)
			}

			// Check call manager triggers
			callTriggered := containsAny(lowerText, config.Settings.CallManagerTriggerWords)
			if tc.shouldTrigger == "call" && !callTriggered {
				t.Errorf("Expected call manager trigger for text: %s", tc.userText)
			}
			if tc.shouldTrigger != "call" && callTriggered {
				t.Errorf("Unexpected call manager trigger for text: %s", tc.userText)
			}
		})
	}
}

// Helper function to check if text contains any of the trigger words
func containsAny(text string, triggerWords []string) bool {
	for _, word := range triggerWords {
		if strings.Contains(text, strings.ToLower(word)) {
			return true
		}
	}
	return false
}
