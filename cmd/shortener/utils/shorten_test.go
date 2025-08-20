package utils

import (
	"strings"
	"testing"
)

func TestGenerateShortID(t *testing.T) {
	// Test that ID is generated
	id := GenerateShortID()
	if id == "" {
		t.Error("Generated ID should not be empty")
	}

	// Test ID length
	if len(id) != idLength {
		t.Errorf("Expected ID length %d, got %d", idLength, len(id))
	}

	// Test that ID contains only valid characters
	for _, char := range id {
		if !strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789", char) {
			t.Errorf("Invalid character in ID: %c", char)
		}
	}

	// Test that multiple IDs are different (very unlikely to be same)
	id1 := GenerateShortID()
	id2 := GenerateShortID()
	if id1 == id2 {
		t.Error("Generated IDs should be different")
	}
}

func TestGenerateShortID_Uniqueness(t *testing.T) {
	// Generate multiple IDs and check for uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := GenerateShortID()
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}
