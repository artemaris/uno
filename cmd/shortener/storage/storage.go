package storage

import "sync"

type Storage interface {
	Save(shortID, originalURL string)
	Get(shortID string) (string, bool)
}

type InMemoryStorage struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data: make(map[string]string),
	}
}

func (s *InMemoryStorage) Save(shortID, originalURL string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[shortID] = originalURL
}

func (s *InMemoryStorage) Get(shortID string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	url, ok := s.data[shortID]
	return url, ok
}
