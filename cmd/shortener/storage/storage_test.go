package storage

import (
	"testing"
)

func TestNewInMemoryStorage(t *testing.T) {
	store := NewInMemoryStorage()
	if store == nil {
		t.Error("NewInMemoryStorage should return non-nil storage")
	}
}

func TestInMemoryStorage_Save(t *testing.T) {
	store := NewInMemoryStorage()

	// Test saving a URL
	store.Save("test-id", "https://example.com", "user1")
	// Save doesn't return error, so just verify it doesn't panic
}

func TestInMemoryStorage_Get(t *testing.T) {
	store := NewInMemoryStorage()

	// Save a URL first
	originalURL := "https://example.com"
	store.Save("test-id", originalURL, "user1")

	// Test getting the URL
	url, deleted, exists := store.Get("test-id")
	if !exists {
		t.Error("Get should return true for existing ID")
	}
	if deleted {
		t.Error("URL should not be deleted")
	}
	if url != originalURL {
		t.Errorf("Expected URL %s, got %s", originalURL, url)
	}

	// Test getting non-existent URL
	_, _, exists = store.Get("non-existent")
	if exists {
		t.Error("Get should return false for non-existent ID")
	}
}

func TestInMemoryStorage_FindByOriginal(t *testing.T) {
	store := NewInMemoryStorage()

	// Save a URL first
	originalURL := "https://example.com"
	store.Save("test-id", originalURL, "user1")

	// Test finding by original URL
	id, exists := store.FindByOriginal(originalURL)
	if !exists {
		t.Error("FindByOriginal should return true for existing URL")
	}
	if id != "test-id" {
		t.Errorf("Expected ID %s, got %s", "test-id", id)
	}

	// Test finding non-existent URL
	_, exists = store.FindByOriginal("https://non-existent.com")
	if exists {
		t.Error("FindByOriginal should return false for non-existent URL")
	}
}

func TestInMemoryStorage_SaveBatch(t *testing.T) {
	store := NewInMemoryStorage()

	// Test batch save
	pairs := map[string]string{
		"id1": "https://example1.com",
		"id2": "https://example2.com",
		"id3": "https://example3.com",
	}

	err := store.SaveBatch(pairs, "user1")
	if err != nil {
		t.Errorf("SaveBatch should not return error: %v", err)
	}

	// Verify all URLs were saved
	for id, originalURL := range pairs {
		url, deleted, exists := store.Get(id)
		if !exists {
			t.Errorf("URL with ID %s should exist", id)
		}
		if deleted {
			t.Errorf("URL with ID %s should not be deleted", id)
		}
		if url != originalURL {
			t.Errorf("Expected URL %s for ID %s, got %s", originalURL, id, url)
		}
	}
}

func TestInMemoryStorage_GetUserURLs(t *testing.T) {
	store := NewInMemoryStorage()

	// Save URLs for different users
	store.Save("id1", "https://example1.com", "user1")
	store.Save("id2", "https://example2.com", "user1")
	store.Save("id3", "https://example3.com", "user2")

	// Test getting URLs for user1
	urls, err := store.GetUserURLs("user1")
	if err != nil {
		t.Errorf("GetUserURLs should not return error: %v", err)
	}
	if len(urls) != 2 {
		t.Errorf("Expected 2 URLs for user1, got %d", len(urls))
	}

	// Test getting URLs for user2
	urls, err = store.GetUserURLs("user2")
	if err != nil {
		t.Errorf("GetUserURLs should not return error: %v", err)
	}
	if len(urls) != 1 {
		t.Errorf("Expected 1 URL for user2, got %d", len(urls))
	}

	// Test getting URLs for non-existent user
	urls, err = store.GetUserURLs("non-existent")
	if err == nil {
		t.Error("GetUserURLs should return error for non-existent user")
	}
	if len(urls) != 0 {
		t.Errorf("Expected 0 URLs for non-existent user, got %d", len(urls))
	}
}

func TestInMemoryStorage_DeleteURLs(t *testing.T) {
	store := NewInMemoryStorage()

	// Save URLs first
	store.Save("id1", "https://example1.com", "user1")
	store.Save("id2", "https://example2.com", "user1")
	store.Save("id3", "https://example3.com", "user1")

	// Test deleting URLs
	urlsToDelete := []string{"id1", "id2"}
	err := store.DeleteURLs("user1", urlsToDelete)
	if err != nil {
		t.Errorf("DeleteURLs should not return error: %v", err)
	}

	// Verify URLs were deleted
	_, deleted, exists := store.Get("id1")
	if !exists {
		t.Error("URL with ID id1 should still exist")
	}
	if !deleted {
		t.Error("URL with ID id1 should be marked as deleted")
	}

	_, deleted, exists = store.Get("id2")
	if !exists {
		t.Error("URL with ID id2 should still exist")
	}
	if !deleted {
		t.Error("URL with ID id2 should be marked as deleted")
	}

	// Verify non-deleted URL still exists and is not deleted
	_, deleted, exists = store.Get("id3")
	if !exists {
		t.Error("URL with ID id3 should still exist")
	}
	if deleted {
		t.Error("URL with ID id3 should not be deleted")
	}
}
