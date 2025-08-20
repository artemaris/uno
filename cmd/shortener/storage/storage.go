package storage

import (
	"fmt"
	"sync"
	"uno/cmd/shortener/models"
)

// Storage определяет интерфейс для хранения и управления сокращенными URL
type Storage interface {
	// Save сохраняет связь между сокращенным ID и оригинальным URL для конкретного пользователя
	Save(shortID, originalURL, userID string)
	
	// Get возвращает оригинальный URL по сокращенному ID, флаг удаления и существования
	Get(shortID string) (originalURL string, deleted bool, exists bool)
	
	// FindByOriginal ищет существующий сокращенный ID для оригинального URL
	FindByOriginal(originalURL string) (string, bool)
	
	// SaveBatch сохраняет пакет URL для конкретного пользователя
	SaveBatch(pairs map[string]string, userID string) error
	
	// GetUserURLs возвращает все URL для конкретного пользователя
	GetUserURLs(userID string) ([]models.UserURL, error)
	
	// DeleteURLs помечает указанные URL как удаленные для конкретного пользователя
	DeleteURLs(userID string, ids []string) error
}

// InMemoryStorage реализует интерфейс Storage с хранением данных в памяти
type InMemoryStorage struct {
	data    map[string]string        // Сокращенный ID -> оригинальный URL
	users   map[string][]models.UserURL // Пользователь -> список его URL
	deleted map[string]bool          // Сокращенный ID -> флаг удаления
	mu      sync.RWMutex             // Мьютекс для безопасного доступа к данным
}

// NewInMemoryStorage создает новый экземпляр InMemoryStorage
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		data:    make(map[string]string),
		users:   make(map[string][]models.UserURL),
		deleted: make(map[string]bool),
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
	s.deleted[shortID] = false
}

func (s *InMemoryStorage) Get(shortID string) (string, bool, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	url, exists := s.data[shortID]
	if !exists {
		return "", false, false
	}

	deleted := s.deleted[shortID]
	return url, deleted, true
}

func (s *InMemoryStorage) FindByOriginal(originalURL string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for id, url := range s.data {
		if s.deleted[id] {
			continue
		}
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
		s.deleted[shortID] = false
	}
	return nil
}

func (s *InMemoryStorage) GetUserURLs(userID string) ([]models.UserURL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	urls, ok := s.users[userID]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	return urls, nil
}

func (s *InMemoryStorage) DeleteURLs(userID string, ids []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	idSet := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		idSet[id] = struct{}{}
		s.deleted[id] = true
	}
	for i, url := range s.users[userID] {
		if _, ok := idSet[url.ShortURL]; ok {
			s.users[userID][i].Deleted = true
		}
	}
	return nil
}
