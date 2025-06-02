package config

import (
	"flag"
	"os"
)

const (
	defaultAddress     = "localhost:8080"
	defaultBaseURL     = "http://localhost:8080"
	defaultStoragePath = "data/storage.jsonl"
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
	dsn := *dsnFlag
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

	return &Config{
		Address:         addr,
		BaseURL:         baseURL,
		FileStoragePath: fileStorage,
		DatabaseDSN:     dsn,
	}
}
