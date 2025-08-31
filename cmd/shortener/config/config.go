package config

import (
	"flag"
	"os"
)

const (
	defaultAddress     = "localhost:8080"
	defaultBaseURL     = "http://localhost:8080"
	defaultStoragePath = "/tmp/short-url-db.json"
)

// Config содержит конфигурационные параметры сервиса сокращения URL
type Config struct {
	Address         string // Адрес HTTP сервера (например, "localhost:8080")
	BaseURL         string // Базовый URL для генерации сокращенных ссылок
	FileStoragePath string // Путь к файлу для хранения данных (если используется файловое хранилище)
	DatabaseDSN     string // Строка подключения к PostgreSQL (если используется база данных)
	EnablePprof     bool   // Включение pprof сервера (только для разработки)
}

// NewConfig создает новый экземпляр конфигурации, читая параметры из:
// 1. Переменных окружения (приоритет)
// 2. Флагов командной строки
// 3. Значений по умолчанию
//
// Поддерживаемые переменные окружения:
// - SERVER_ADDRESS: адрес сервера
// - BASE_URL: базовый URL
// - FILE_STORAGE_PATH: путь к файлу хранилища
// - DATABASE_DSN: строка подключения к PostgreSQL
// - ENABLE_PPROF: включение pprof сервера (true/false, только для разработки)
//
// Поддерживаемые флаги командной строки:
// - -a: адрес сервера
// - -b: базовый URL
// - -f: путь к файлу хранилища
// - -d: строка подключения к PostgreSQL
// - -pprof: включение pprof сервера (только для разработки)
func NewConfig() *Config {
	addressFlag := flag.String("a", defaultAddress, "http service address")
	baseURLFlag := flag.String("b", defaultBaseURL, "http base url")
	filePathFlag := flag.String("f", defaultStoragePath, "storage path")
	dsnFlag := flag.String("d", "", "PostgreSQL DSN")
	pprofFlag := flag.Bool("pprof", false, "enable pprof server (development only)")
	flag.Parse()

	addr := os.Getenv("SERVER_ADDRESS")
	if addr == "" {
		addr = *addressFlag
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = *baseURLFlag
	}

	fileStorage := os.Getenv("FILE_STORAGE_PATH")
	if fileStorage == "" {
		fileStorage = *filePathFlag
	}

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = *dsnFlag
	}

	enablePprof := *pprofFlag
	if pprofEnv := os.Getenv("ENABLE_PPROF"); pprofEnv != "" {
		enablePprof = pprofEnv == "true" || pprofEnv == "1"
	}

	return &Config{
		Address:         addr,
		BaseURL:         baseURL,
		FileStoragePath: fileStorage,
		DatabaseDSN:     dsn,
		EnablePprof:     enablePprof,
	}
}
