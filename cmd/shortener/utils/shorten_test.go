package utils

import (
	"strings"
	"testing"
)

func TestGenerateShortID(t *testing.T) {
	// Test that ID is generated
	id, err := GenerateShortID()
	if err != nil {
		t.Fatalf("GenerateShortID returned error: %v", err)
	}
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
	id1, err := GenerateShortID()
	if err != nil {
		t.Fatalf("GenerateShortID returned error: %v", err)
	}
	id2, err := GenerateShortID()
	if err != nil {
		t.Fatalf("GenerateShortID returned error: %v", err)
	}
	if id1 == id2 {
		t.Error("Generated IDs should be different")
	}
}

func TestGenerateShortID_Uniqueness(t *testing.T) {
	// Generate multiple IDs and check for uniqueness
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id, err := GenerateShortID()
		if err != nil {
			t.Fatalf("GenerateShortID returned error: %v", err)
		}
		if ids[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		ids[id] = true
	}
}

func TestGenerateShortID_ErrorHandling(t *testing.T) {
	// Проверяем, что функция корректно возвращает ошибку и пустую строку при ошибке
	// В нормальных условиях crypto/rand не должен возвращать ошибку,
	// но мы проверяем корректность сигнатуры функции
	id, err := GenerateShortID()
	if err != nil {
		// Если есть ошибка, ID должен быть пустым
		if id != "" {
			t.Errorf("Expected empty ID when error occurs, got: %s", id)
		}
	} else {
		// Если ошибки нет, ID должен быть валидным
		if id == "" {
			t.Error("Expected non-empty ID when no error occurs")
		}
		if len(id) != idLength {
			t.Errorf("Expected ID length %d, got %d", idLength, len(id))
		}
	}
}
