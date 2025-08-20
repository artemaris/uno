package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"uno/cmd/shortener/middleware"
	"uno/cmd/shortener/storage"

	"go.uber.org/zap/zaptest"
)

func TestDeleteUserURLsHandler(t *testing.T) {
	// Create test storage
	store := storage.NewInMemoryStorage()

	// Add some test URLs
	store.Save("id1", "https://example1.com", "user1")
	store.Save("id2", "https://example2.com", "user1")
	store.Save("id3", "https://example3.com", "user1")

	// Create test logger
	logger := zaptest.NewLogger(t)

	// Create delete queue
	deleteQueue := make(chan DeleteRequest, 10)

	// Create handler
	handler := DeleteUserURLsHandler(store, logger, deleteQueue)

	// Test request body - the handler expects just an array of strings
	urlsToDelete := []string{"id1", "id2"}

	body, _ := json.Marshal(urlsToDelete)

	// Create request
	req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	// Add user ID to context using the correct key
	ctx := context.WithValue(req.Context(), middleware.ContextUserIDKey, "user1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusAccepted {
		t.Errorf("Expected status 202, got %d", rec.Code)
	}

	// Check if delete request was sent to queue
	select {
	case deleteReq := <-deleteQueue:
		if deleteReq.UserID != "user1" {
			t.Errorf("Expected user ID user1, got %s", deleteReq.UserID)
		}
		if len(deleteReq.IDs) != 2 {
			t.Errorf("Expected 2 IDs, got %d", len(deleteReq.IDs))
		}
		// Check if IDs are in the request
		idsMap := make(map[string]bool)
		for _, id := range deleteReq.IDs {
			idsMap[id] = true
		}
		if !idsMap["id1"] || !idsMap["id2"] {
			t.Error("Expected IDs id1 and id2 in delete request")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Delete request was not sent to queue")
	}
}

func TestDeleteUserURLsHandler_InvalidJSON(t *testing.T) {
	// Create test storage
	store := storage.NewInMemoryStorage()

	// Create test logger
	logger := zaptest.NewLogger(t)

	// Create delete queue
	deleteQueue := make(chan DeleteRequest, 10)

	// Create handler
	handler := DeleteUserURLsHandler(store, logger, deleteQueue)

	// Test with invalid JSON
	req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	// Add user ID to context using the correct key
	ctx := context.WithValue(req.Context(), middleware.ContextUserIDKey, "user1")
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}

func TestDeleteUserURLsHandler_NoUserID(t *testing.T) {
	// Create test storage
	store := storage.NewInMemoryStorage()

	// Create test logger
	logger := zaptest.NewLogger(t)

	// Create delete queue
	deleteQueue := make(chan DeleteRequest, 10)

	// Create handler
	handler := DeleteUserURLsHandler(store, logger, deleteQueue)

	// Test without user ID in context
	req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewBufferString("[]"))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	// Call handler
	handler.ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", rec.Code)
	}
}

func TestRunDeletionWorker(t *testing.T) {
	// Create test storage
	store := storage.NewInMemoryStorage()

	// Add some test URLs
	store.Save("id1", "https://example1.com", "user1")
	store.Save("id2", "https://example2.com", "user1")

	// Create test logger
	logger := zaptest.NewLogger(t)

	// Create delete queue
	deleteQueue := make(chan DeleteRequest, 10)

	// Create context with cancel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start deletion worker
	go RunDeletionWorker(ctx, store, logger, deleteQueue)

	// Send delete request
	deleteRequest := DeleteRequest{
		UserID: "user1",
		IDs:    []string{"id1", "id2"},
	}

	deleteQueue <- deleteRequest

	// Wait a bit for processing
	time.Sleep(100 * time.Millisecond)

	// Verify URLs were marked as deleted
	_, deleted1, exists1 := store.Get("id1")
	if !exists1 {
		t.Error("URL id1 should still exist")
	}
	if !deleted1 {
		t.Error("URL id1 should be marked as deleted")
	}

	_, deleted2, exists2 := store.Get("id2")
	if !exists2 {
		t.Error("URL id2 should still exist")
	}
	if !deleted2 {
		t.Error("URL id2 should be marked as deleted")
	}
}
