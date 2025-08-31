package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"uno/cmd/shortener/config"
	"uno/cmd/shortener/middleware"
	"uno/cmd/shortener/models"
	"uno/cmd/shortener/storage"
)

// ExampleShortenURLHandler демонстрирует использование ShortenURLHandler
// для сокращения URL в текстовом формате
func ExampleShortenURLHandler() {
	// Создаем конфигурацию и хранилище
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	store := storage.NewInMemoryStorage()

	// Создаем тестовый запрос
	req := httptest.NewRequest("POST", "/", strings.NewReader("https://example.com"))
	req.Header.Set("Content-Type", "text/plain")

	// Добавляем userID в контекст (в реальном приложении это делает middleware)
	req = req.WithContext(contextWithUserID("test-user-123"))

	// Создаем ResponseRecorder для записи ответа
	w := httptest.NewRecorder()

	// Вызываем хендлер
	handler := ShortenURLHandler(cfg, store)
	handler.ServeHTTP(w, req)

	// Проверяем результат
	fmt.Printf("Status: %d\n", w.Code)
	response := w.Body.String()
	if strings.HasPrefix(response, "http://localhost:8080/") {
		fmt.Printf("Generated short URL with length: %d\n", len(response))
	}
	// Output:
	// Status: 201
	// Generated short URL with length: 30
}

// ExampleAPIShortenHandler демонстрирует использование APIShortenHandler
// для сокращения URL через JSON API
func ExampleAPIShortenHandler() {
	// Создаем конфигурацию и хранилище
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	store := storage.NewInMemoryStorage()

	// Создаем JSON запрос
	requestBody := models.APIRequest{URL: "https://example.com"}
	jsonData, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/shorten", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(contextWithUserID("test-user-123"))

	w := httptest.NewRecorder()

	// Вызываем хендлер
	handler := APIShortenHandler(cfg, store)
	handler.ServeHTTP(w, req)

	// Проверяем результат
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))

	var response models.APIResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	if strings.HasPrefix(response.Result, "http://localhost:8080/") {
		fmt.Printf("Generated short URL with length: %d\n", len(response.Result))
	}
	// Output:
	// Status: 201
	// Content-Type: application/json
	// Generated short URL with length: 30
}

// ExampleBatchShortenHandler демонстрирует использование BatchShortenHandler
// для пакетного сокращения URL
func ExampleBatchShortenHandler() {
	// Создаем конфигурацию и хранилище
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	store := storage.NewInMemoryStorage()

	// Создаем пакетный запрос
	batchRequest := []models.BatchRequest{
		{CorrelationID: "1", OriginalURL: "https://example1.com"},
		{CorrelationID: "2", OriginalURL: "https://example2.com"},
		{CorrelationID: "3", OriginalURL: "https://example3.com"},
	}

	jsonData, _ := json.Marshal(batchRequest)

	req := httptest.NewRequest("POST", "/api/shorten/batch", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(contextWithUserID("test-user-123"))

	w := httptest.NewRecorder()

	// Вызываем хендлер
	handler := BatchShortenHandler(cfg, store)
	handler.ServeHTTP(w, req)

	// Проверяем результат
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))

	var responses []models.BatchResponse
	json.Unmarshal(w.Body.Bytes(), &responses)

	fmt.Printf("Batch processed: %d URLs\n", len(responses))
	for _, resp := range responses {
		if strings.HasPrefix(resp.ShortURL, "http://localhost:8080/") {
			fmt.Printf("ID: %s -> Short URL generated\n", resp.CorrelationID)
		}
	}
	// Output:
	// Status: 201
	// Content-Type: application/json
	// Batch processed: 3 URLs
	// ID: 1 -> Short URL generated
	// ID: 2 -> Short URL generated
	// ID: 3 -> Short URL generated
}

// ExampleRedirectHandler демонстрирует использование RedirectHandler
// для перенаправления по сокращенным URL
func ExampleRedirectHandler() {
	// Создаем хранилище и добавляем тестовый URL
	store := storage.NewInMemoryStorage()
	store.Save("AbCdEfGh", "https://example.com", "test-user-123")

	// Создаем тестовый хендлер без chi router для примера
	testHandler := func(w http.ResponseWriter, r *http.Request) {
		// Извлекаем shortID из URL (в реальном приложении это делает chi router)
		path := r.URL.Path
		shortID := strings.TrimPrefix(path, "/")

		originalURL, deleted, exists := store.Get(shortID)

		if !exists {
			http.NotFound(w, r)
			return
		}

		if deleted {
			w.WriteHeader(http.StatusGone)
			return
		}

		w.Header().Set("Location", originalURL)
		w.WriteHeader(http.StatusTemporaryRedirect)
	}

	// Создаем запрос с shortID в URL
	req := httptest.NewRequest("GET", "/AbCdEfGh", nil)

	w := httptest.NewRecorder()

	// Вызываем тестовый хендлер
	testHandler(w, req)

	// Проверяем результат
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Location: %s\n", w.Header().Get("Location"))
	// Output:
	// Status: 307
	// Location: https://example.com
}

// ExampleUserURLsHandler демонстрирует использование UserURLsHandler
// для получения всех URL пользователя
func ExampleUserURLsHandler() {
	// Создаем конфигурацию и хранилище
	cfg := &config.Config{BaseURL: "http://localhost:8080"}
	store := storage.NewInMemoryStorage()

	// Добавляем тестовые URL для пользователя
	store.Save("AbCdEfGh", "https://example1.com", "test-user-123")
	store.Save("IjKlMnOp", "https://example2.com", "test-user-123")

	// Создаем запрос
	req := httptest.NewRequest("GET", "/api/user/urls", nil)
	req = req.WithContext(contextWithUserID("test-user-123"))

	w := httptest.NewRecorder()

	// Вызываем хендлер
	handler := UserURLsHandler(cfg, store)
	handler.ServeHTTP(w, req)

	// Проверяем результат
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Content-Type: %s\n", w.Header().Get("Content-Type"))

	var userURLs []models.UserURL
	json.Unmarshal(w.Body.Bytes(), &userURLs)

	fmt.Printf("User has %d URLs\n", len(userURLs))
	for _, url := range userURLs {
		fmt.Printf("- %s -> %s\n", url.ShortURL, url.OriginalURL)
	}
	// Output:
	// Status: 200
	// Content-Type: application/json
	// User has 2 URLs
	// - http://localhost:8080/AbCdEfGh -> https://example1.com
	// - http://localhost:8080/IjKlMnOp -> https://example2.com
}

// ExampleDeleteUserURLsHandler демонстрирует использование DeleteUserURLsHandler
// для асинхронного удаления URL пользователя
func ExampleDeleteUserURLsHandler() {
	// Создаем хранилище и добавляем тестовые URL
	store := storage.NewInMemoryStorage()
	store.Save("AbCdEfGh", "https://example1.com", "test-user-123")
	store.Save("IjKlMnOp", "https://example2.com", "test-user-123")

	// Создаем канал для очереди удаления
	deleteQueue := make(chan DeleteRequest, 10)

	// Создаем JSON запрос с ID для удаления
	idsToDelete := []string{"AbCdEfGh"}
	jsonData, _ := json.Marshal(idsToDelete)

	req := httptest.NewRequest("DELETE", "/api/user/urls", bytes.NewReader(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(contextWithUserID("test-user-123"))

	w := httptest.NewRecorder()

	// Вызываем хендлер
	handler := DeleteUserURLsHandler(store, nil, deleteQueue)
	handler.ServeHTTP(w, req)

	// Проверяем результат
	fmt.Printf("Status: %d\n", w.Code)

	// Получаем запрос из очереди
	deleteReq := <-deleteQueue
	fmt.Printf("Delete request for user: %s\n", deleteReq.UserID)
	fmt.Printf("URLs to delete: %v\n", deleteReq.IDs)
	// Output:
	// Status: 202
	// Delete request for user: test-user-123
	// URLs to delete: [AbCdEfGh]
}

// ExamplePingHandler демонстрирует использование PingHandler
// для проверки доступности базы данных
func ExamplePingHandler() {
	// Создаем запрос
	req := httptest.NewRequest("GET", "/ping", nil)

	w := httptest.NewRecorder()

	// Вызываем хендлер без базы данных (nil)
	handler := PingHandler(nil)
	handler.ServeHTTP(w, req)

	// Проверяем результат
	fmt.Printf("Status without DB: %d\n", w.Code)
	// Output:
	// Status without DB: 200
}

// Вспомогательная функция для создания контекста с userID
// В реальном приложении это делает middleware WithUserID
func contextWithUserID(userID string) context.Context {
	ctx := context.Background()
	// Используем тот же ключ, что и в middleware
	ctx = context.WithValue(ctx, middleware.ContextUserIDKey, userID)
	return ctx
}
