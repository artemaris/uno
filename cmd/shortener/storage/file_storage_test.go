package storage

import (
	"os"
	"strings"
	"testing"
)

func TestFileStorage_SaveAndGet(t *testing.T) {
	testFile := "/tmp/file_storage_test.json"
	defer os.Remove(testFile)

	store, err := NewFileStorage(testFile)
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}

	fs, ok := store.(*fileStorage)
	if !ok {
		t.Fatal("store is not *fileStorage")
	}

	shortID := "abc123"
	originalURL := "https://example.com"

	store.Save(shortID, originalURL)
	fs.Flush()

	got, ok := store.Get(shortID)
	if !ok {
		t.Fatal("expected key to be found")
	}
	if got != originalURL {
		t.Fatalf("expected %s, got %s", originalURL, got)
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, shortID) || !strings.Contains(content, originalURL) {
		t.Error("file does not contain expected content")
	}
}
