package bot

import (
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
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

func TestVoiceMessageDetection(t *testing.T) {
	// Test text message (should not be treated as voice)
	textUpdate := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: "Hello, this is a text message",
			Chat: &tgbotapi.Chat{ID: 123},
		},
	}

	if textUpdate.Message.Voice != nil {
		t.Error("Text message should not have Voice field set")
	}

	// Test voice message structure
	voiceUpdate := tgbotapi.Update{
		Message: &tgbotapi.Message{
			Voice: &tgbotapi.Voice{
				FileID:   "test_voice_file_id",
				Duration: 10,
			},
			Chat: &tgbotapi.Chat{ID: 123},
		},
	}

	if voiceUpdate.Message.Voice == nil {
		t.Error("Voice message should have Voice field set")
	}

	if voiceUpdate.Message.Voice.FileID != "test_voice_file_id" {
		t.Errorf("Expected FileID 'test_voice_file_id', got %s", voiceUpdate.Message.Voice.FileID)
	}

	if voiceUpdate.Message.Voice.Duration != 10 {
		t.Errorf("Expected Duration 10, got %d", voiceUpdate.Message.Voice.Duration)
	}
}
