package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/middleware"
	"uno/cmd/shortener/storage"
)

func TestUserURLsHandler(t *testing.T) {
	// Create test storage
	store := storage.NewInMemoryStorage()

	// Add some test URLs
	store.Save("id1", "https://example1.com", "user1")
	store.Save("id2", "https://example2.com", "user1")
	store.Save("id3", "https://example3.com", "user2")

	// Create test config
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	// Create handler
	handler := UserURLsHandler(cfg, store)

	// Test request with user1
	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextUserIDKey, "user1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Check content type
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", rec.Header().Get("Content-Type"))
	}

	// Verify response body contains user1's URLs
	body := rec.Body.String()
	if len(body) == 0 {
		t.Error("Response body should not be empty")
	}

	// Check that response contains expected URLs
	if !contains(body, "id1") || !contains(body, "id2") {
		t.Error("Response should contain user1's URLs")
	}

	// Check that response doesn't contain user2's URLs
	if contains(body, "id3") {
		t.Error("Response should not contain user2's URLs")
	}
}

func TestUserURLsHandler_NoUserID(t *testing.T) {
	// Create test storage
	store := storage.NewInMemoryStorage()

	// Create test config
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	// Create handler
	handler := UserURLsHandler(cfg, store)

	// Test request without user ID
	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	rec := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestUserURLsHandler_EmptyUserURLs(t *testing.T) {
	// Create test storage
	store := storage.NewInMemoryStorage()

	// Create test config
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	// Create handler
	handler := UserURLsHandler(cfg, store)

	// Test request with user that has no URLs
	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextUserIDKey, "user-with-no-urls")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(rec, req)

	// Verify response - GetUserURLs returns error for non-existent users
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}

	// Check error message - http.Error adds a newline
	body := rec.Body.String()
	expectedMessage := "failed to get user URLs\n"
	if body != expectedMessage {
		t.Errorf("Expected error message '%s', got '%s'", expectedMessage, body)
	}
}

func TestUserURLsHandler_WithDeletedURLs(t *testing.T) {
	// Create test storage
	store := storage.NewInMemoryStorage()

	// Add some test URLs
	store.Save("id1", "https://example1.com", "user1")
	store.Save("id2", "https://example2.com", "user1")

	// Delete one URL
	store.DeleteURLs("user1", []string{"id1"})

	// Create test config
	cfg := &config.Config{BaseURL: "http://localhost:8080"}

	// Create handler
	handler := UserURLsHandler(cfg, store)

	// Test request
	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	ctx := context.WithValue(req.Context(), middleware.ContextUserIDKey, "user1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// Check that response contains both URLs (deleted ones are filtered out by the handler)
	body := rec.Body.String()
	if !contains(body, "id2") {
		t.Error("Response should contain non-deleted URL id2")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			contains(s[1:len(s)-1], substr))))
}
