package config

import (
	"flag"
)

type Config struct {
	Address string
	BaseURL string
}

func NewConfig() *Config {
	addr := flag.String("a", "localhost:8080", "http service address")
	baseURL := flag.String("b", "http://localhost:8080", "http base url")

	flag.Parse()

	return &Config{
		Address: *addr,
		BaseURL: *baseURL,
	}
}
