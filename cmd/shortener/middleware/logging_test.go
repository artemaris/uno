package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestLoggingMiddleware(t *testing.T) {
	// Create a test logger
	logger := zaptest.NewLogger(t)

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	// Create middleware
	middleware := LoggingMiddleware(logger)

	// Test request
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Real-IP", "192.168.1.1")

	rec := httptest.NewRecorder()

	// Record start time
	start := time.Now()

	// Call middleware
	middleware(handler).ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	if rec.Body.String() != "Hello, World!" {
		t.Errorf("Expected body 'Hello, World!', got '%s'", rec.Body.String())
	}

	// Verify that middleware didn't panic and completed
	elapsed := time.Since(start)
	if elapsed > 5*time.Second {
		t.Error("Middleware took too long to complete")
	}
}

func TestLoggingMiddleware_WithError(t *testing.T) {
	// Create a test logger
	logger := zaptest.NewLogger(t)

	// Create a test handler that returns an error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	})

	// Create middleware
	middleware := LoggingMiddleware(logger)

	// Test request
	req := httptest.NewRequest("POST", "/error", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	rec := httptest.NewRecorder()

	// Call middleware
	middleware(handler).ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}

	if rec.Body.String() != "Internal Server Error" {
		t.Errorf("Expected body 'Internal Server Error', got '%s'", rec.Body.String())
	}
}

func TestLoggingMiddleware_WithCustomHeaders(t *testing.T) {
	// Create a test logger
	logger := zaptest.NewLogger(t)

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom-Header", "custom-value")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	})

	// Create middleware
	middleware := LoggingMiddleware(logger)

	// Test request with custom headers
	req := httptest.NewRequest("GET", "/custom", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	req.Header.Set("X-Request-ID", "req-123")
	req.Header.Set("Authorization", "Bearer token123")

	rec := httptest.NewRecorder()

	// Call middleware
	middleware(handler).ServeHTTP(rec, req)

	// Verify response
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	if rec.Header().Get("X-Custom-Header") != "custom-value" {
		t.Errorf("Expected custom header 'custom-value', got '%s'", rec.Header().Get("X-Custom-Header"))
	}

	if rec.Body.String() != "Success" {
		t.Errorf("Expected body 'Success', got '%s'", rec.Body.String())
	}
}

func TestLoggingMiddleware_WithNilLogger(t *testing.T) {
	// Test with nil logger (should panic)
	var logger *zap.Logger

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Create middleware
	middleware := LoggingMiddleware(logger)

	// Test request
	req := httptest.NewRequest("GET", "/nil-logger", nil)
	req.RemoteAddr = "127.0.0.1:12345"

	rec := httptest.NewRecorder()

	// Call middleware (should panic with nil logger)
	defer func() {
		if r := recover(); r != nil {
			// Expected panic with nil logger
			t.Logf("Expected panic with nil logger: %v", r)
		} else {
			t.Error("Expected panic with nil logger, but no panic occurred")
		}
	}()

	middleware(handler).ServeHTTP(rec, req)

	// If we get here, the test should fail
	t.Error("Expected panic with nil logger")
}
