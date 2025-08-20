package models

import (
	"encoding/json"
	"testing"
)

func TestBatchRequest_MarshalJSON(t *testing.T) {
	req := BatchRequest{
		CorrelationID: "corr-123",
		OriginalURL:   "https://example.com",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Errorf("Failed to marshal BatchRequest: %v", err)
	}

	var unmarshaled BatchRequest
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal BatchRequest: %v", err)
	}

	if unmarshaled.CorrelationID != req.CorrelationID {
		t.Errorf("Expected CorrelationID %s, got %s", req.CorrelationID, unmarshaled.CorrelationID)
	}

	if unmarshaled.OriginalURL != req.OriginalURL {
		t.Errorf("Expected OriginalURL %s, got %s", req.OriginalURL, unmarshaled.OriginalURL)
	}
}

func TestBatchResponse_MarshalJSON(t *testing.T) {
	resp := BatchResponse{
		CorrelationID: "corr-123",
		ShortURL:      "http://localhost:8080/abc123",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Errorf("Failed to marshal BatchResponse: %v", err)
	}

	var unmarshaled BatchResponse
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal BatchResponse: %v", err)
	}

	if unmarshaled.CorrelationID != resp.CorrelationID {
		t.Errorf("Expected CorrelationID %s, got %s", resp.CorrelationID, unmarshaled.CorrelationID)
	}

	if unmarshaled.ShortURL != resp.ShortURL {
		t.Errorf("Expected ShortURL %s, got %s", resp.ShortURL, unmarshaled.ShortURL)
	}
}

func TestUserURL_MarshalJSON(t *testing.T) {
	userURL := UserURL{
		ShortURL:    "abc123",
		OriginalURL: "https://example.com",
		Deleted:     false,
	}

	data, err := json.Marshal(userURL)
	if err != nil {
		t.Errorf("Failed to marshal UserURL: %v", err)
	}

	var unmarshaled UserURL
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal UserURL: %v", err)
	}

	if unmarshaled.ShortURL != userURL.ShortURL {
		t.Errorf("Expected ShortURL %s, got %s", userURL.ShortURL, unmarshaled.ShortURL)
	}

	if unmarshaled.OriginalURL != userURL.OriginalURL {
		t.Errorf("Expected OriginalURL %s, got %s", userURL.OriginalURL, unmarshaled.OriginalURL)
	}

	if unmarshaled.Deleted != userURL.Deleted {
		t.Errorf("Expected Deleted %v, got %v", userURL.Deleted, unmarshaled.Deleted)
	}
}

func TestUserURL_Deleted(t *testing.T) {
	userURL := UserURL{
		ShortURL:    "abc123",
		OriginalURL: "https://example.com",
		Deleted:     true,
	}

	if !userURL.Deleted {
		t.Error("UserURL should be marked as deleted")
	}

	// Also verify other fields to use them
	if userURL.ShortURL != "abc123" {
		t.Errorf("Expected ShortURL abc123, got %s", userURL.ShortURL)
	}

	if userURL.OriginalURL != "https://example.com" {
		t.Errorf("Expected OriginalURL https://example.com, got %s", userURL.OriginalURL)
	}
}

func TestBatchRequest_UnmarshalJSON(t *testing.T) {
	jsonData := `{"correlation_id":"corr-123","original_url":"https://example.com"}`

	var req BatchRequest
	err := json.Unmarshal([]byte(jsonData), &req)
	if err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	if req.CorrelationID != "corr-123" {
		t.Errorf("Expected CorrelationID corr-123, got %s", req.CorrelationID)
	}

	if req.OriginalURL != "https://example.com" {
		t.Errorf("Expected OriginalURL https://example.com, got %s", req.OriginalURL)
	}
}

func TestBatchResponse_UnmarshalJSON(t *testing.T) {
	jsonData := `{"correlation_id":"corr-123","short_url":"http://localhost:8080/abc123"}`

	var resp BatchResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	if err != nil {
		t.Errorf("Failed to unmarshal JSON: %v", err)
	}

	if resp.CorrelationID != "corr-123" {
		t.Errorf("Expected CorrelationID corr-123, got %s", resp.CorrelationID)
	}

	if resp.ShortURL != "http://localhost:8080/abc123" {
		t.Errorf("Expected ShortURL http://localhost:8080/abc123, got %s", resp.ShortURL)
	}
}
