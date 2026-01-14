package ai

import (
	"testing"
)

func TestTranscriptionInterface(t *testing.T) {
	// Test that GPTStrategy implements the interface correctly
	strategy := NewGPTStrategy("test-api-key")
	
	// This should compile without error, proving the interface is implemented
	var _ ModelStrategy = strategy
	
	// Test that LocalStrategy implements the interface correctly
	localStrategy := NewLocalStrategy()
	var _ ModelStrategy = localStrategy
}

func TestGPTStrategyTranscriptionCall(t *testing.T) {
	// This test checks that the transcription method can be called
	// without actually making API calls (which would require a real API key)
	strategy := NewGPTStrategy("test-api-key")
	
	// Test with empty data - should return an error
	_, err := strategy.TranscribeAudio([]byte{}, "ogg")
	if err == nil {
		t.Error("Expected error for empty audio data with invalid API key")
	}
	
	// The error should be an OpenAI API error, not a compilation error
	if err != nil && len(err.Error()) == 0 {
		t.Error("Error message should not be empty")
	}
}

func TestLocalStrategyTranscriptionStub(t *testing.T) {
	// Test that LocalStrategy returns expected error
	strategy := NewLocalStrategy()
	
	_, err := strategy.TranscribeAudio([]byte{}, "ogg")
	if err == nil {
		t.Error("LocalStrategy should return error for transcription")
	}
	
	expectedMsg := "local transcription not implemented"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}