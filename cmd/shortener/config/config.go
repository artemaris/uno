package config

import (
	"encoding/json"
	"flag"
	"os"
)

const (
	defaultAddress     = "localhost:8080"
	defaultBaseURL     = "http://localhost:8080"
	defaultStoragePath = "/tmp/short-url-db.json"
	defaultCertFile    = "cert.pem"
	defaultKeyFile     = "key.pem"
)

// Config содержит конфигурационные параметры сервиса сокращения URL
type Config struct {
	Address         string // Адрес HTTP сервера (например, "localhost:8080")
	BaseURL         string // Базовый URL для генерации сокращенных ссылок
	FileStoragePath string // Путь к файлу для хранения данных (если используется файловое хранилище)
	DatabaseDSN     string // Строка подключения к PostgreSQL (если используется база данных)
	EnablePprof     bool   // Включение pprof сервера (только для разработки)
	EnableHTTPS     bool   // Включение HTTPS сервера
	CertFile        string // Путь к файлу сертификата
	KeyFile         string // Путь к файлу приватного ключа
}

// JSONConfig представляет структуру JSON конфигурации
type JSONConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

// NewConfig создает новый экземпляр конфигурации, читая параметры из:
// 1. Флагов командной строки (наивысший приоритет)
// 2. Переменных окружения
// 3. JSON файла конфигурации
// 4. Значений по умолчанию (наименьший приоритет)
//
// Поддерживаемые переменные окружения:
// - SERVER_ADDRESS: адрес сервера
// - BASE_URL: базовый URL
// - FILE_STORAGE_PATH: путь к файлу хранилища
// - DATABASE_DSN: строка подключения к PostgreSQL
// - ENABLE_PPROF: включение pprof сервера (true/false, только для разработки)
// - ENABLE_HTTPS: включение HTTPS сервера (true/false)
// - CERT_FILE: путь к файлу сертификата (по умолчанию "cert.pem")
// - KEY_FILE: путь к файлу приватного ключа (по умолчанию "key.pem")
// - CONFIG: путь к JSON файлу конфигурации
//
// Поддерживаемые флаги командной строки:
// - -a: адрес сервера
// - -b: базовый URL
// - -f: путь к файлу хранилища
// - -d: строка подключения к PostgreSQL
// - -pprof: включение pprof сервера (только для разработки)
// - -s: включение HTTPS сервера
// - -cert: путь к файлу сертификата
// - -key: путь к файлу приватного ключа
// - -c/-config: путь к JSON файлу конфигурации
func NewConfig() *Config {
	addressFlag := flag.String("a", defaultAddress, "http service address")
	baseURLFlag := flag.String("b", defaultBaseURL, "http base url")
	filePathFlag := flag.String("f", defaultStoragePath, "storage path")
	dsnFlag := flag.String("d", "", "PostgreSQL DSN")
	pprofFlag := flag.Bool("pprof", false, "enable pprof server (development only)")
	httpsFlag := flag.Bool("s", false, "enable HTTPS server")
	certFlag := flag.String("cert", defaultCertFile, "path to certificate file")
	keyFlag := flag.String("key", defaultKeyFile, "path to private key file")
	configFlag := flag.String("c", "", "path to JSON config file")
	flag.Parse()

	// Загружаем JSON конфигурацию (если указана)
	jsonConfig := loadJSONConfig(*configFlag)

	addr := os.Getenv("SERVER_ADDRESS")
	if addr == "" {
		addr = *addressFlag
	}
	if addr == defaultAddress && jsonConfig.ServerAddress != "" {
		addr = jsonConfig.ServerAddress
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = *baseURLFlag
	}
	if baseURL == defaultBaseURL && jsonConfig.BaseURL != "" {
		baseURL = jsonConfig.BaseURL
	}

	fileStorage := os.Getenv("FILE_STORAGE_PATH")
	if fileStorage == "" {
		fileStorage = *filePathFlag
	}
	if fileStorage == defaultStoragePath && jsonConfig.FileStoragePath != "" {
		fileStorage = jsonConfig.FileStoragePath
	}

	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = *dsnFlag
	}
	if dsn == "" && jsonConfig.DatabaseDSN != "" {
		dsn = jsonConfig.DatabaseDSN
	}

	enablePprof := *pprofFlag
	if pprofEnv := os.Getenv("ENABLE_PPROF"); pprofEnv != "" {
		enablePprof = pprofEnv == "true" || pprofEnv == "1"
	}

	enableHTTPS := *httpsFlag
	if httpsEnv := os.Getenv("ENABLE_HTTPS"); httpsEnv != "" {
		enableHTTPS = httpsEnv == "true" || httpsEnv == "1"
	}
	if !enableHTTPS && jsonConfig.EnableHTTPS {
		enableHTTPS = jsonConfig.EnableHTTPS
	}

	certFile := *certFlag
	if certEnv := os.Getenv("CERT_FILE"); certEnv != "" {
		certFile = certEnv
	}

	keyFile := *keyFlag
	if keyEnv := os.Getenv("KEY_FILE"); keyEnv != "" {
		keyFile = keyEnv
	}

	// Если включен HTTPS, обновляем BaseURL для использования https://
	if enableHTTPS && baseURL == *baseURLFlag {
		baseURL = "https://" + addr
	}

	return &Config{
		Address:         addr,
		BaseURL:         baseURL,
		FileStoragePath: fileStorage,
		DatabaseDSN:     dsn,
		EnablePprof:     enablePprof,
		EnableHTTPS:     enableHTTPS,
		CertFile:        certFile,
		KeyFile:         keyFile,
	}
}

// loadJSONConfig загружает конфигурацию из JSON файла
func loadJSONConfig(configPath string) JSONConfig {
	// Если путь к конфигурации не указан через флаг, проверяем переменную окружения
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	// Если путь к конфигурации не указан, возвращаем пустую конфигурацию
	if configPath == "" {
		return JSONConfig{}
	}

	// Читаем файл конфигурации
	data, err := os.ReadFile(configPath)
	if err != nil {
		// Если файл не найден или не может быть прочитан, возвращаем пустую конфигурацию
		return JSONConfig{}
	}

	// Парсим JSON
	var jsonConfig JSONConfig
	if err := json.Unmarshal(data, &jsonConfig); err != nil {
		// Если JSON невалидный, возвращаем пустую конфигурацию
		return JSONConfig{}
	}

	return jsonConfig
}
