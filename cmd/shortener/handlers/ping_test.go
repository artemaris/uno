package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestPingHandler(t *testing.T) {
	req1 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	res1 := httptest.NewRecorder()
	h1 := PingHandler(nil)
	h1.ServeHTTP(res1, req1)
	if res1.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", res1.Code)
	}

	conn, err := pgx.Connect(context.Background(), "postgres://invalid")
	if err == nil {
		defer conn.Close(context.Background())
	}
	req2 := httptest.NewRequest(http.MethodGet, "/ping", nil)
	res2 := httptest.NewRecorder()
	h2 := PingHandler(conn)
	h2.ServeHTTP(res2, req2)
	expected := http.StatusInternalServerError
	if conn == nil {
		expected = http.StatusOK
	}
	if res2.Code != expected {
		t.Fatalf("expected %d, got %d", expected, res2.Code)
	}
}
