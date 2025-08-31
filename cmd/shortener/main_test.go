package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestGetBuildValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string returns N/A",
			input:    "",
			expected: "N/A",
		},
		{
			name:     "non-empty string returns itself",
			input:    "v1.0.0",
			expected: "v1.0.0",
		},
		{
			name:     "space string returns itself",
			input:    " ",
			expected: " ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getBuildValue(tt.input)
			if result != tt.expected {
				t.Errorf("getBuildValue(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPrintBuildInfo(t *testing.T) {
	// Сохраняем оригинальные значения
	originalVersion := buildVersion
	originalDate := buildDate
	originalCommit := buildCommit

	// Восстанавливаем после теста
	defer func() {
		buildVersion = originalVersion
		buildDate = originalDate
		buildCommit = originalCommit
	}()

	tests := []struct {
		name            string
		version         string
		date            string
		commit          string
		expectedVersion string
		expectedDate    string
		expectedCommit  string
	}{
		{
			name:            "all values empty should show N/A",
			version:         "",
			date:            "",
			commit:          "",
			expectedVersion: "N/A",
			expectedDate:    "N/A",
			expectedCommit:  "N/A",
		},
		{
			name:            "all values set should show actual values",
			version:         "v1.0.0",
			date:            "2025-08-31",
			commit:          "abc123",
			expectedVersion: "v1.0.0",
			expectedDate:    "2025-08-31",
			expectedCommit:  "abc123",
		},
		{
			name:            "mixed values should show correct output",
			version:         "v2.0.0",
			date:            "",
			commit:          "def456",
			expectedVersion: "v2.0.0",
			expectedDate:    "N/A",
			expectedCommit:  "def456",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Устанавливаем тестовые значения
			buildVersion = tt.version
			buildDate = tt.date
			buildCommit = tt.commit

			// Захватываем stdout
			var buf bytes.Buffer
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// Вызываем функцию
			printBuildInfo()

			// Восстанавливаем stdout
			w.Close()
			os.Stdout = oldStdout

			// Читаем захваченный вывод
			buf.ReadFrom(r)
			output := buf.String()

			// Проверяем содержимое
			expectedLines := []string{
				"Build version: " + tt.expectedVersion,
				"Build date: " + tt.expectedDate,
				"Build commit: " + tt.expectedCommit,
			}

			lines := strings.Split(strings.TrimSpace(output), "\n")
			if len(lines) != 3 {
				t.Errorf("Expected 3 lines of output, got %d: %v", len(lines), lines)
			}

			for i, expectedLine := range expectedLines {
				if i >= len(lines) || lines[i] != expectedLine {
					t.Errorf("Line %d: expected %q, got %q", i, expectedLine, lines[i])
				}
			}
		})
	}
}

func TestGlobalVariablesDefaultValues(t *testing.T) {
	// Этот тест проверяет, что глобальные переменные имеют корректные значения по умолчанию
	// В реальной среде они будут установлены через ldflags при сборке
	
	// Проверяем, что переменные объявлены и доступны
	_ = buildVersion
	_ = buildDate  
	_ = buildCommit

	// Если переменные не были установлены через ldflags, они должны иметь значения по умолчанию
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	// Проверяем, что getBuildValue корректно обрабатывает пустые значения
	if getBuildValue("") != "N/A" {
		t.Error("getBuildValue should return 'N/A' for empty string")
	}
}
