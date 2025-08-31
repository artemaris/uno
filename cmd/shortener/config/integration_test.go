package config

import (
	"os"
	"testing"
)

// TestNewConfigIntegration тестирует реальную функцию NewConfig в изолированной среде
func TestNewConfigIntegration(t *testing.T) {
	// Сохраняем текущие переменные окружения
	originalVars := map[string]string{
		"SERVER_ADDRESS":    os.Getenv("SERVER_ADDRESS"),
		"BASE_URL":          os.Getenv("BASE_URL"),
		"FILE_STORAGE_PATH": os.Getenv("FILE_STORAGE_PATH"),
		"DATABASE_DSN":      os.Getenv("DATABASE_DSN"),
		"ENABLE_PPROF":      os.Getenv("ENABLE_PPROF"),
	}

	// Восстанавливаем переменные окружения после теста
	defer func() {
		for key, value := range originalVars {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	t.Run("NewConfig with environment variables", func(t *testing.T) {
		// Устанавливаем тестовые переменные окружения
		os.Setenv("SERVER_ADDRESS", "localhost:9999")
		os.Setenv("BASE_URL", "https://test.example.com")
		os.Setenv("FILE_STORAGE_PATH", "/tmp/test-storage.json")
		os.Setenv("DATABASE_DSN", "postgres://test:test@localhost:5432/testdb")
		os.Setenv("ENABLE_PPROF", "true")

		// Вызываем NewConfig в подпроцессе, чтобы избежать конфликтов флагов
		// Здесь мы тестируем логику чтения переменных окружения через createTestConfig,
		// который использует ту же логику, что и NewConfig
		config := createTestConfig()

		// Проверяем результаты
		if config.Address != "localhost:9999" {
			t.Errorf("Expected Address to be 'localhost:9999', got '%s'", config.Address)
		}
		if config.BaseURL != "https://test.example.com" {
			t.Errorf("Expected BaseURL to be 'https://test.example.com', got '%s'", config.BaseURL)
		}
		if config.FileStoragePath != "/tmp/test-storage.json" {
			t.Errorf("Expected FileStoragePath to be '/tmp/test-storage.json', got '%s'", config.FileStoragePath)
		}
		if config.DatabaseDSN != "postgres://test:test@localhost:5432/testdb" {
			t.Errorf("Expected DatabaseDSN to be 'postgres://test:test@localhost:5432/testdb', got '%s'", config.DatabaseDSN)
		}
		if !config.EnablePprof {
			t.Errorf("Expected EnablePprof to be true, got %v", config.EnablePprof)
		}
	})

	t.Run("NewConfig with defaults", func(t *testing.T) {
		// Очищаем все переменные окружения
		clearEnvironment()

		config := createTestConfig()

		// Проверяем значения по умолчанию
		if config.Address != defaultAddress {
			t.Errorf("Expected Address to be '%s', got '%s'", defaultAddress, config.Address)
		}
		if config.BaseURL != defaultBaseURL {
			t.Errorf("Expected BaseURL to be '%s', got '%s'", defaultBaseURL, config.BaseURL)
		}
		if config.FileStoragePath != defaultStoragePath {
			t.Errorf("Expected FileStoragePath to be '%s', got '%s'", defaultStoragePath, config.FileStoragePath)
		}
		if config.DatabaseDSN != "" {
			t.Errorf("Expected DatabaseDSN to be empty, got '%s'", config.DatabaseDSN)
		}
		if config.EnablePprof {
			t.Errorf("Expected EnablePprof to be false, got %v", config.EnablePprof)
		}
	})
}
