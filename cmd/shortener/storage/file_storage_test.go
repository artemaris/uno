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
		t.Errorf("failed to create file storage: %v", err)
	}

	shortID := "abc123"
	originalURL := "https://example.com"
	userID := "test-user"

	store.Save(shortID, originalURL, userID)

	got, deleted, exists := store.Get(shortID)
	if !exists {
		t.Error("expected key to be found")
	}
	if deleted {
		t.Error("expected URL not to be deleted")
	}
	if got != originalURL {
		t.Errorf("expected %s, got %s", originalURL, got)
	}

	urls, err := store.GetUserURLs(userID)
	if err != nil {
		t.Errorf("failed to get user URLs: %v", err)
	}
	if len(urls) != 1 {
		t.Errorf("expected 1 URL for user, got %d", len(urls))
	}
	if urls[0].ShortURL != shortID || urls[0].OriginalURL != originalURL {
		t.Errorf("unexpected user URL: %+v", urls[0])
	}

	data, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("failed to read file: %v", err)
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
		t.Errorf("failed to create file storage: %v", err)
	}

	userID := "user1"
	shortID1 := "id1"
	shortID2 := "id2"
	originalURL1 := "https://example.com/1"
	originalURL2 := "https://example.com/2"

	store.Save(shortID1, originalURL1, userID)
	store.Save(shortID2, originalURL2, userID)

	err = store.DeleteURLs(userID, []string{shortID1})
	if err != nil {
		t.Errorf("failed to delete URL: %v", err)
	}

	_, deleted, exists := store.Get(shortID1)
	if !exists {
		t.Error("expected shortID1 to exist")
	}
	if !deleted {
		t.Error("expected shortID1 to be marked as deleted")
	}

	val, deleted, exists := store.Get(shortID2)
	if !exists || deleted || val != originalURL2 {
		t.Errorf("expected shortID2 to be present and not deleted")
	}

	urls, err := store.GetUserURLs(userID)
	if err != nil {
		t.Errorf("failed to get user URLs: %v", err)
	}
	if len(urls) != 1 || urls[0].ShortURL != shortID2 {
		t.Errorf("unexpected result from GetUserURLs: %+v", urls)
	}
}
