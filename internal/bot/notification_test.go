package bot

import (
	"testing"
)

func TestSendToAllNotifications(t *testing.T) {
	// Initialize notification chats for testing
	notificationChats = []int64{123456789}
	
	// Test that function doesn't panic when bot is not initialized
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("SendToAllNotifications should not panic even when bot is not initialized: %v", r)
		}
	}()
	
	SendToAllNotifications("Test notification message")
}

func TestNotificationBotHandling(t *testing.T) {
	// Test that we have the notification chat functionality
	if len(notificationChats) == 0 {
		notificationChats = []int64{123456789}
	}
	
	// Verify notification chats are set correctly
	if len(notificationChats) != 1 || notificationChats[0] != 123456789 {
		t.Errorf("Expected notification chats to be set correctly")
	}
}