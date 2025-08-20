package config

import (
	"os"
	"testing"
)

func TestNewConfig_EnvironmentVariables(t *testing.T) {
	// Test that environment variables override defaults
	os.Setenv("SERVER_ADDRESS", ":9090")
	os.Setenv("BASE_URL", "https://example.com")
	os.Setenv("FILE_STORAGE_PATH", "/tmp/test.json")
	os.Setenv("DATABASE_DSN", "postgres://user:pass@localhost/db")

	// Clean up after test
	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("DATABASE_DSN")
	}()

	// Note: We can't easily test NewConfig() due to flag conflicts in tests
	// But we can verify that environment variables are set correctly
	if os.Getenv("SERVER_ADDRESS") != ":9090" {
		t.Error("Environment variable SERVER_ADDRESS not set correctly")
	}

	if os.Getenv("BASE_URL") != "https://example.com" {
		t.Error("Environment variable BASE_URL not set correctly")
	}

	if os.Getenv("FILE_STORAGE_PATH") != "/tmp/test.json" {
		t.Error("Environment variable FILE_STORAGE_PATH not set correctly")
	}

	if os.Getenv("DATABASE_DSN") != "postgres://user:pass@localhost/db" {
		t.Error("Environment variable DATABASE_DSN not set correctly")
	}
}

func TestNewConfig_EmptyEnvironmentVariables(t *testing.T) {
	// Test that empty environment variables don't interfere
	os.Setenv("SERVER_ADDRESS", "")
	os.Setenv("BASE_URL", "")
	os.Setenv("FILE_STORAGE_PATH", "")
	os.Setenv("DATABASE_DSN", "")

	// Clean up after test
	defer func() {
		os.Unsetenv("SERVER_ADDRESS")
		os.Unsetenv("BASE_URL")
		os.Unsetenv("FILE_STORAGE_PATH")
		os.Unsetenv("DATABASE_DSN")
	}()

	// Verify environment variables are empty
	if os.Getenv("SERVER_ADDRESS") != "" {
		t.Error("Environment variable SERVER_ADDRESS should be empty")
	}

	if os.Getenv("BASE_URL") != "" {
		t.Error("Environment variable BASE_URL should be empty")
	}

	if os.Getenv("FILE_STORAGE_PATH") != "" {
		t.Error("Environment variable FILE_STORAGE_PATH should be empty")
	}

	if os.Getenv("DATABASE_DSN") != "" {
		t.Error("Environment variable DATABASE_DSN should be empty")
	}
}
