package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

type fileStorage struct {
	filePath string
	mu       sync.RWMutex
	data     map[string]string
	file     *os.File
	queue    chan record
	done     chan struct{}
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
		filePath: path,
		data:     make(map[string]string),
		file:     file,
	}

	fs.queue = make(chan record, 1000)
	fs.done = make(chan struct{})
	go fs.runWriter()

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
		fs.data[r.ShortURL] = r.OriginalURL
	}
	return scanner.Err()
}

func (fs *fileStorage) Save(shortID, originalURL string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.data[shortID] = originalURL

	rec := record{
		UUID:        uuid.NewString(),
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}
	_, err := json.Marshal(rec)
	if err != nil {
		return
	}
	fs.queue <- record{}
}

func (fs *fileStorage) Get(shortID string) (string, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	url, ok := fs.data[shortID]
	return url, ok
}

func (fs *fileStorage) FindByOriginal(originalURL string) (string, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	for shortID, url := range fs.data {
		if url == originalURL {
			return shortID, true
		}
	}
	return "", false
}

func (fs *fileStorage) SaveBatch(pairs map[string]string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	for shortID, originalURL := range pairs {
		fs.data[shortID] = originalURL

		rec := record{
			UUID:        uuid.NewString(),
			ShortURL:    shortID,
			OriginalURL: originalURL,
		}
		_, err := json.Marshal(rec)
		if err != nil {
			continue
		}
		fs.queue <- record{}
	}
	return nil
}

func (fs *fileStorage) runWriter() {
	buffer := make([]record, 0, 100)

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case rec := <-fs.queue:
			buffer = append(buffer, rec)
			if len(buffer) >= 100 {
				fs.flush(buffer)
				buffer = buffer[:0]
			}
		case <-ticker.C:
			if len(buffer) > 0 {
				fs.flush(buffer)
				buffer = buffer[:0]
			}
		case <-fs.done:
			fs.flush(buffer)
			return
		}
	}
}

func (fs *fileStorage) flush(records []record) {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	for _, rec := range records {
		data, err := json.Marshal(rec)
		if err != nil {
			continue
		}
		fs.file.Write(append(data, '\n'))
	}
}
