package storage

import "sync"

type Storage interface {
	Save(shortID, originalURL string)
	Get(shortID string) (string, bool)
	FindByOriginal(originalURL string) (string, bool)
	SaveBatch(pairs map[string]string) error
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

func (s *InMemoryStorage) SaveBatch(pairs map[string]string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for shortID, originalURL := range pairs {
		s.data[shortID] = originalURL
	}
	return nil
}
