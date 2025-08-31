package storage

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"uno/cmd/shortener/models"

	"github.com/google/uuid"
)

// FileStorage реализует интерфейс Storage с хранением данных в файле
// Поддерживает персистентность данных и асинхронные операции
type FileStorage struct {
	filePath        string                      // Путь к файлу хранилища
	mu              sync.RWMutex                // Мьютекс для безопасного доступа к данным
	originalToShort map[string]string           // Оригинальный URL -> сокращенный ID
	shortToOriginal map[string]string           // Сокращенный ID -> оригинальный URL
	file            *os.File                    // Файл для записи данных
	userURLs        map[string][]models.UserURL // Пользователь -> список его URL
	deleted         map[string]bool             // Сокращенный ID -> флаг удаления
}

// record представляет запись в файле хранилища
type record struct {
	UUID        string `json:"uuid"`         // Уникальный идентификатор записи
	ShortURL    string `json:"short_url"`    // Сокращенный URL
	OriginalURL string `json:"original_url"` // Оригинальный URL
	UserID      string `json:"user_id"`      // Идентификатор пользователя
	DeletedFlag bool   `json:"deleted_flag"` // Флаг удаления
}

// NewFileStorage создает новый экземпляр FileStorage
// Создает директорию для файла, если она не существует
// Загружает существующие данные из файла при инициализации
func NewFileStorage(path string) (Storage, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("failed to open file storage: %w", err)
	}

	fs := &FileStorage{
		filePath:        path,
		originalToShort: make(map[string]string),
		shortToOriginal: make(map[string]string),
		file:            file,
		userURLs:        make(map[string][]models.UserURL),
		deleted:         make(map[string]bool),
	}

	if err := fs.load(); err != nil {
		return nil, err
	}

	return fs, nil
}

// load загружает данные из файла в память
// Читает файл построчно и восстанавливает состояние хранилища
func (fs *FileStorage) load() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if fs.userURLs == nil {
		fs.userURLs = make(map[string][]models.UserURL)
	}
	if fs.deleted == nil {
		fs.deleted = make(map[string]bool)
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
		u := models.UserURL{
			ShortURL:    r.ShortURL,
			OriginalURL: r.OriginalURL,
			Deleted:     r.DeletedFlag,
		}
		fs.userURLs[r.UserID] = append(fs.userURLs[r.UserID], u)
		if r.DeletedFlag {
			fs.deleted[r.ShortURL] = true
			continue
		}
		fs.originalToShort[r.OriginalURL] = r.ShortURL
		fs.shortToOriginal[r.ShortURL] = r.OriginalURL
		fs.deleted[r.ShortURL] = false
	}
	return scanner.Err()
}

// Save сохраняет связь между сокращенным ID и оригинальным URL для конкретного пользователя
// Записывает данные в файл для персистентности
// Если URL уже существует, операция игнорируется
func (fs *FileStorage) Save(shortID, originalURL, userID string) {
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
	fs.deleted[shortID] = false
	jsonData, err := json.Marshal(rec)
	if err != nil {
		return
	}
	if _, err := fs.file.Write(append(jsonData, '\n')); err != nil {
		log.Printf("FileStorage: failed to write record to file: %v", err)
	}
}

// Get возвращает оригинальный URL по сокращенному ID, флаг удаления и существования
func (fs *FileStorage) Get(shortID string) (string, bool, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	originalURL, exists := fs.shortToOriginal[shortID]
	if !exists {
		return "", false, false
	}

	deleted := fs.deleted[shortID]
	return originalURL, deleted, true
}

// FindByOriginal ищет существующий сокращенный ID для оригинального URL
func (fs *FileStorage) FindByOriginal(originalURL string) (string, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	id, ok := fs.originalToShort[originalURL]
	return id, ok
}

// SaveBatch сохраняет пакет URL для конкретного пользователя
// Обрабатывает каждый URL аналогично методу Save
func (fs *FileStorage) SaveBatch(pairs map[string]string, userID string) error {
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
		fs.deleted[shortID] = false
		jsonData, err := json.Marshal(rec)
		if err != nil {
			continue
		}
		_, _ = fs.file.Write(append(jsonData, '\n'))
	}
	return nil
}

// DeleteURLs помечает указанные URL как удаленные для конкретного пользователя
// Записывает информацию об удалении в файл для персистентности
func (fs *FileStorage) DeleteURLs(userID string, ids []string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	toDelete := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		toDelete[id] = struct{}{}
	}

	for i, u := range fs.userURLs[userID] {
		if _, del := toDelete[u.ShortURL]; del && !fs.deleted[u.ShortURL] {
			fs.userURLs[userID][i].Deleted = true
			fs.deleted[u.ShortURL] = true
			rec := record{
				UUID:        uuid.NewString(),
				ShortURL:    u.ShortURL,
				OriginalURL: u.OriginalURL,
				UserID:      userID,
				DeletedFlag: true,
			}
			if b, err := json.Marshal(rec); err == nil {
				fs.file.Write(append(b, '\n'))
			}
		}
	}
	return nil
}

// GetUserURLs возвращает все не удаленные URL для конкретного пользователя
// Фильтрует удаленные URL из результата
func (fs *FileStorage) GetUserURLs(userID string) ([]models.UserURL, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	urls, ok := fs.userURLs[userID]
	if !ok {
		return nil, nil
	}

	var filtered []models.UserURL
	for _, u := range urls {
		if !u.Deleted {
			filtered = append(filtered, u)
		}
	}
	return filtered, nil
}
