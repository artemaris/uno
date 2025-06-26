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

	urls, err := store.GetUserURLs(userID)
	if err != nil {
		t.Fatalf("failed to get user URLs: %v", err)
	}
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

func TestFileStorage_DeleteURLs(t *testing.T) {
	testFile := "temp/file_storage_test_delete.json"
	defer os.Remove(testFile)

	store, err := NewFileStorage(testFile)
	if err != nil {
		t.Fatalf("failed to create file storage: %v", err)
	}

	userID := "user1"
	shortID1 := "id1"
	shortID2 := "id2"
	originalURL1 := "https://example.com/1"
	originalURL2 := "https://example.com/2"

	// Save two URLs
	store.Save(shortID1, originalURL1, userID)
	store.Save(shortID2, originalURL2, userID)

	// Delete one
	err = store.DeleteURLs(userID, []string{shortID1})
	if err != nil {
		t.Fatalf("failed to delete URL: %v", err)
	}

	// Check Get returns nothing for deleted
	if _, ok := store.Get(shortID1); ok {
		t.Errorf("expected deleted shortID1 to be absent")
	}

	// Check Get returns remaining one
	if val, ok := store.Get(shortID2); !ok || val != originalURL2 {
		t.Errorf("expected shortID2 to be present")
	}

	// Check GetUserURLs returns only undeleted URL
	urls, err := store.GetUserURLs(userID)
	if err != nil {
		t.Fatalf("failed to get user URLs: %v", err)
	}
	if len(urls) != 1 || urls[0].ShortURL != shortID2 {
		t.Errorf("unexpected result from GetUserURLs: %+v", urls)
	}
}
