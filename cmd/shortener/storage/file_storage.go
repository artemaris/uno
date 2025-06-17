package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

type fileStorage struct {
	filePath        string
	mu              sync.RWMutex
	originalToShort map[string]string
	shortToOriginal map[string]string
	file            *os.File
}

type record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewFileStorage(path string) (Storage, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open file storage: %w", err)
	}

	fs := &fileStorage{
		filePath:        path,
		originalToShort: make(map[string]string),
		shortToOriginal: make(map[string]string),
		file:            file,
	}

	if err := fs.load(); err != nil {
		return nil, err
	}

	return fs, nil
}

func (fs *fileStorage) load() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.file.Seek(0, 0)
	scanner := bufio.NewScanner(fs.file)
	for scanner.Scan() {
		var r record
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			continue
		}
		fs.originalToShort[r.OriginalURL] = r.ShortURL
		fs.shortToOriginal[r.ShortURL] = r.OriginalURL
	}
	return scanner.Err()
}

func (fs *fileStorage) Save(shortID, originalURL string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// не сохраняем повторно
	if _, exists := fs.originalToShort[originalURL]; exists {
		return
	}

	fs.originalToShort[originalURL] = shortID
	fs.shortToOriginal[shortID] = originalURL

	rec := record{
		UUID:        uuid.NewString(),
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}
	jsonData, err := json.Marshal(rec)
	if err != nil {
		return
	}
	fs.file.Write(append(jsonData, '\n'))
}

func (fs *fileStorage) Get(shortID string) (string, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	url, ok := fs.shortToOriginal[shortID]
	return url, ok
}

func (fs *fileStorage) FindByOriginal(originalURL string) (string, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	id, ok := fs.originalToShort[originalURL]
	return id, ok
}

func (fs *fileStorage) SaveBatch(pairs map[string]string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for shortID, originalURL := range pairs {
		if _, exists := fs.originalToShort[originalURL]; exists {
			continue
		}

		fs.originalToShort[originalURL] = shortID
		fs.shortToOriginal[shortID] = originalURL

		rec := record{
			UUID:        uuid.NewString(),
			ShortURL:    shortID,
			OriginalURL: originalURL,
		}
		jsonData, err := json.Marshal(rec)
		if err != nil {
			continue
		}
		fs.file.Write(append(jsonData, '\n'))
	}
	return nil
}
