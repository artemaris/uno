package storage

import (
	"os"
	"strings"
	"testing"
)

func TestFileStorage_SaveAndGet(t *testing.T) {
	testFile := "temp/file_storage_test.json"
	defer os.Remove(testFile)

	store, err := NewFileStorage(testFile)
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}

	shortID := "abc123"
	originalURL := "https://example.com"
	userID := "test-user"

	store.Save(shortID, originalURL, userID)

	got, ok := store.Get(shortID)
	if !ok {
		t.Fatal("expected key to be found")
	}
	if got != originalURL {
		t.Fatalf("expected %s, got %s", originalURL, got)
	}

	urls := store.GetUserURLs(userID)
	if len(urls) != 1 {
		t.Fatalf("expected 1 URL for user, got %d", len(urls))
	}
	if urls[0].ShortURL != shortID || urls[0].OriginalURL != originalURL {
		t.Errorf("unexpected user URL: %+v", urls[0])
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
