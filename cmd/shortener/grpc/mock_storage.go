package grpc

import (
	"uno/cmd/shortener/models"
	"uno/cmd/shortener/storage"

	"github.com/stretchr/testify/mock"
)

// MockStorage - мок для тестирования
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Save(shortID, originalURL, userID string) {
	m.Called(shortID, originalURL, userID)
}

func (m *MockStorage) Get(id string) (string, bool, bool) {
	args := m.Called(id)
	return args.String(0), args.Bool(1), args.Bool(2)
}

func (m *MockStorage) FindByOriginal(originalURL string) (string, bool) {
	args := m.Called(originalURL)
	return args.String(0), args.Bool(1)
}

func (m *MockStorage) SaveBatch(pairs map[string]string, userID string) error {
	args := m.Called(pairs, userID)
	return args.Error(0)
}

func (m *MockStorage) GetUserURLs(userID string) ([]models.UserURL, error) {
	args := m.Called(userID)
	return args.Get(0).([]models.UserURL), args.Error(1)
}

func (m *MockStorage) DeleteURLs(userID string, ids []string) error {
	args := m.Called(userID, ids)
	return args.Error(0)
}

func (m *MockStorage) GetStats() (storage.Stats, error) {
	args := m.Called()
	return args.Get(0).(storage.Stats), args.Error(1)
}

// Убеждаемся, что MockStorage реализует интерфейс Storage
var _ storage.Storage = (*MockStorage)(nil)
