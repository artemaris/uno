package storage

import (
	"sync"
	"uno/cmd/shortener/models"
)

type Storage interface {
	Save(shortID, originalURL, userID string)
	Get(shortID string) (string, bool)
	FindByOriginal(originalURL string) (string, bool)
	SaveBatch(pairs map[string]string, userID string) error
	GetUserURLs(userID string) []models.UserURL
}

type InMemoryStorage struct {
	data  map[string]string
	users map[string][]models.UserURL
	mu    sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data:  make(map[string]string),
		users: make(map[string][]models.UserURL),
	}
}

func (s *InMemoryStorage) Save(shortID, originalURL, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[shortID] = originalURL
	s.users[userID] = append(s.users[userID], models.UserURL{
		ShortURL:    shortID,
		OriginalURL: originalURL,
	})
}

func (s *InMemoryStorage) Get(shortID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.data[shortID]
	return url, ok
}

func (s *InMemoryStorage) FindByOriginal(originalURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for id, url := range s.data {
		if url == originalURL {
			return id, true
		}
	}
	return "", false
}

func (s *InMemoryStorage) SaveBatch(pairs map[string]string, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for shortID, originalURL := range pairs {
		s.data[shortID] = originalURL
		s.users[userID] = append(s.users[userID], models.UserURL{
			ShortURL:    shortID,
			OriginalURL: originalURL,
		})
	}
	return nil
}

func (s *InMemoryStorage) GetUserURLs(userID string) []models.UserURL {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.users[userID]
}
