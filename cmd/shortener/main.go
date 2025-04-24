package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

var (
	urlToStore = make(map[string]string)
	mu         sync.Mutex
)

// функция main вызывается автоматически при запуске приложения
func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

// функция run будет полезна при инициализации зависимостей сервера перед запуском
func run() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			shortenUrlHandler(w, r)
		} else {
			redirectHandler(w, r)
		}

	})
	fmt.Println("Listening on port 8080")
	return http.ListenAndServe(":8080", nil)
}

func shortenUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// разрешаем только POST-запросы
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Content-Type") != `text/plain` {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	longUrl := string(body)

	hash := md5.Sum([]byte(longUrl))
	shortUrl := hex.EncodeToString(hash[:])[:6]
	mu.Lock()
	urlToStore[shortUrl] = longUrl
	mu.Unlock()

	shortUrl = fmt.Sprintf("http://localhost:8080/%s", shortUrl)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "201 Created\n%s", shortUrl)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	shortUrl := strings.TrimPrefix(r.URL.Path, "/")
	if shortUrl == "" {
		http.NotFound(w, r)
		return
	}

	mu.Lock()
	originalUrl, ok := urlToStore[shortUrl]
	mu.Unlock()

	if !ok {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusTemporaryRedirect)
	fmt.Fprintf(w, "307 Temporary Redirect\n%s", originalUrl)
}
