package config

import (
	"testing"
)

func TestLoadSettings(t *testing.T) {
	settings := LoadSettings()
	
	// Test that default schedule trigger words are set
	if len(settings.ScheduleTriggerWords) == 0 {
		t.Error("ScheduleTriggerWords should not be empty")
	}
	
	// Test that default price trigger words are set  
	if len(settings.PriceTriggerWords) == 0 {
		t.Error("PriceTriggerWords should not be empty")
	}
	
	// Test that known trigger words are present
	scheduleFound := false
	for _, word := range settings.ScheduleTriggerWords {
		if word == "расписание" {
			scheduleFound = true
			break
		}
	}
	if !scheduleFound {
		t.Error("Expected 'расписание' to be in ScheduleTriggerWords")
	}
	
	priceFound := false
	for _, word := range settings.PriceTriggerWords {
		if word == "цены" {
			priceFound = true
			break
		}
	}
	if !priceFound {
		t.Error("Expected 'цены' to be in PriceTriggerWords")
	}
}