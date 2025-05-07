package config

import (
	"flag"
	"os"
)

const (
	defaultAddress = "localhost:8080"
	defaultBaseURL = "http://localhost:8080"
)

type Config struct {
	Address string
	BaseURL string
}

func NewConfig() *Config {
	addressFlag := flag.String("a", defaultAddress, "http service address")
	baseURLFlag := flag.String("b", defaultBaseURL, "http base url")
	flag.Parse()

	addr := os.Getenv("SERVER_ADDRESS")
	if addr == "" {
		addr = *addressFlag
	}

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = *baseURLFlag
	}

	return &Config{
		Address: addr,
		BaseURL: baseURL,
	}
}
