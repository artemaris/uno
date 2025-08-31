package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGzipMiddleware(t *testing.T) {
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hello, World!"))
	})

	// Test without gzip encoding
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	middleware := GzipMiddleware(handler)
	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	if rec.Body.String() != "Hello, World!" {
		t.Errorf("Expected body 'Hello, World!', got '%s'", rec.Body.String())
	}

	// Test with gzip encoding
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec = httptest.NewRecorder()

	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Errorf("Expected Content-Encoding gzip, got %s", rec.Header().Get("Content-Encoding"))
	}

	// Verify gzipped content
	reader, err := gzip.NewReader(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read gzipped content: %v", err)
	}

	if string(body) != "Hello, World!" {
		t.Errorf("Expected decompressed body 'Hello, World!', got '%s'", string(body))
	}
}

func TestGzipMiddleware_AlwaysGzipWhenAccepted(t *testing.T) {
	// Create a test handler with small content
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("Hi"))
	})

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()

	middleware := GzipMiddleware(handler)
	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	// The middleware always gzips when Accept-Encoding includes gzip
	if rec.Header().Get("Content-Encoding") != "gzip" {
		t.Error("Content should be gzipped when Accept-Encoding includes gzip")
	}

	// Verify the gzipped content
	reader, err := gzip.NewReader(bytes.NewReader(rec.Body.Bytes()))
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v", err)
	}
	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("Failed to read gzipped content: %v", err)
	}

	if string(body) != "Hi" {
		t.Errorf("Expected decompressed body 'Hi', got '%s'", string(body))
	}
}

func TestGzipMiddleware_ContentTypeFiltering(t *testing.T) {
	// Test that all content types are gzipped when Accept-Encoding includes gzip
	testCases := []struct {
		contentType string
		shouldGzip  bool
	}{
		{"text/plain", true},
		{"application/json", true},
		{"image/png", true},       // The middleware gzips everything when accepted
		{"application/pdf", true}, // The middleware gzips everything when accepted
	}

	for _, tc := range testCases {
		t.Run(tc.contentType, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", tc.contentType)
				w.Write([]byte("Test content"))
			})

			req := httptest.NewRequest("GET", "/", nil)
			req.Header.Set("Accept-Encoding", "gzip")
			rec := httptest.NewRecorder()

			middleware := GzipMiddleware(handler)
			middleware.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Errorf("Expected status 200, got %d", rec.Code)
			}

			// The middleware always gzips when Accept-Encoding includes gzip
			isGzipped := rec.Header().Get("Content-Encoding") == "gzip"
			if isGzipped != tc.shouldGzip {
				t.Errorf("Content-Type %s: expected gzipped=%v, got gzipped=%v",
					tc.contentType, tc.shouldGzip, isGzipped)
			}
		})
	}
}

func TestGzipMiddleware_WithGzippedRequest(t *testing.T) {
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "failed to read body", http.StatusInternalServerError)
			return
		}
		w.Write(body)
	})

	// Create gzipped request body
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	gw.Write([]byte("Hello, Gzipped Request!"))
	gw.Close()

	req := httptest.NewRequest("POST", "/", bytes.NewReader(buf.Bytes()))
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()

	middleware := GzipMiddleware(handler)
	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	if rec.Body.String() != "Hello, Gzipped Request!" {
		t.Errorf("Expected body 'Hello, Gzipped Request!', got '%s'", rec.Body.String())
	}
}
