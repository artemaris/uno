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

type Config struct {
	Address         string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
}

func NewConfig() *Config {
	addressFlag := flag.String("a", defaultAddress, "http service address")
	baseURLFlag := flag.String("b", defaultBaseURL, "http base url")
	filePathFlag := flag.String("f", defaultStoragePath, "storage path")
	dsnFlag := flag.String("d", "", "PostgreSQL DSN")
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

	return &Config{
		Address:         addr,
		BaseURL:         baseURL,
		FileStoragePath: fileStorage,
		DatabaseDSN:     dsn,
	}
}
