package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"uno/cmd/shortener/models"

	"github.com/google/uuid"
)

type fileStorage struct {
	filePath        string
	mu              sync.RWMutex
	originalToShort map[string]string
	shortToOriginal map[string]string
	file            *os.File
	userURLs        map[string][]models.UserURL
}

type record struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	UserID      string `json:"user_id"`
	DeletedFlag bool   `json:"deleted_flag"`
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
		userURLs:        make(map[string][]models.UserURL),
	}

	if err := fs.load(); err != nil {
		return nil, err
	}

	return fs, nil
}

func (fs *fileStorage) load() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.userURLs == nil {
		fs.userURLs = make(map[string][]models.UserURL)
	}

	_, err := fs.file.Seek(0, 0)
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(fs.file)
	for scanner.Scan() {
		var r record
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			continue
		}
		if r.DeletedFlag {
			continue
		}
		fs.originalToShort[r.OriginalURL] = r.ShortURL
		fs.shortToOriginal[r.ShortURL] = r.OriginalURL
		fs.userURLs[r.UserID] = append(fs.userURLs[r.UserID], models.UserURL{
			ShortURL:    r.ShortURL,
			OriginalURL: r.OriginalURL,
		})
	}
	return scanner.Err()
}

func (fs *fileStorage) Save(shortID, originalURL, userID string) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, exists := fs.originalToShort[originalURL]; exists {
		return
	}

	fs.originalToShort[originalURL] = shortID
	fs.shortToOriginal[shortID] = originalURL

	rec := record{
		UUID:        uuid.NewString(),
		ShortURL:    shortID,
		OriginalURL: originalURL,
		UserID:      userID,
	}
	fs.userURLs[userID] = append(fs.userURLs[userID], models.UserURL{
		ShortURL:    shortID,
		OriginalURL: originalURL,
	})
	jsonData, err := json.Marshal(rec)
	if err != nil {
		return
	}
	_, _ = fs.file.Write(append(jsonData, '\n'))
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

func (fs *fileStorage) SaveBatch(pairs map[string]string, userID string) error {
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
			UserID:      userID,
		}
		fs.userURLs[userID] = append(fs.userURLs[userID], models.UserURL{
			ShortURL:    shortID,
			OriginalURL: originalURL,
		})
		jsonData, err := json.Marshal(rec)
		if err != nil {
			continue
		}
		_, _ = fs.file.Write(append(jsonData, '\n'))
	}
	return nil
}

func (fs *fileStorage) DeleteURLs(userID string, ids []string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	idsSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idsSet[id] = struct{}{}
	}

	for _, uid := range []string{userID} {
		urls, ok := fs.userURLs[uid]
		if !ok {
			continue
		}
		var updatedURLs []models.UserURL
		for _, u := range urls {
			if _, toDelete := idsSet[u.ShortURL]; toDelete {
				// Mark as deleted and write record
				rec := record{
					UUID:        uuid.NewString(),
					ShortURL:    u.ShortURL,
					OriginalURL: u.OriginalURL,
					UserID:      userID,
					DeletedFlag: true,
				}
				jsonData, err := json.Marshal(rec)
				if err == nil {
					_, _ = fs.file.Write(append(jsonData, '\n'))
				}
				delete(fs.shortToOriginal, u.ShortURL)
				delete(fs.originalToShort, u.OriginalURL)
			} else {
				updatedURLs = append(updatedURLs, u)
			}
		}
		fs.userURLs[uid] = updatedURLs
	}

	return nil
}

func (fs *fileStorage) GetUserURLs(userID string) ([]models.UserURL, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	urls, ok := fs.userURLs[userID]
	if !ok {
		return nil, nil
	}
	return urls, nil
}
