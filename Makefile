# Makefile для сборки URL shortener

# Получаем информацию о сборке
BUILD_VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%d_%H:%M:%S")
BUILD_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Флаги для линковщика
LDFLAGS = -ldflags "-X main.buildVersion=$(BUILD_VERSION) -X main.buildDate=$(BUILD_DATE) -X main.buildCommit=$(BUILD_COMMIT)"

# Основной бинарник
BINARY_NAME = shortener
BINARY_PATH = cmd/shortener/main.go

# Staticlint
STATICLINT_NAME = staticlint
STATICLINT_PATH = cmd/staticlint/main.go

.PHONY: all build build-with-version test clean lint fmt help run

# Цель по умолчанию
all: build

# Простая сборка без информации о версии
build:
	go build -o $(BINARY_NAME) $(BINARY_PATH)

# Сборка с информацией о версии
build-with-version:
	go build $(LDFLAGS) -o $(BINARY_NAME) $(BINARY_PATH)

# Сборка staticlint
build-staticlint:
	go build -o $(STATICLINT_NAME) $(STATICLINT_PATH)

# Запуск тестов
test:
	go test -v ./...

# Запуск тестов с покрытием
test-coverage:
	go test -coverprofile=coverage.out ./cmd/shortener/...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out | tail -1

# Линтинг (требует сборки staticlint)
lint: build-staticlint
	./$(STATICLINT_NAME) ./cmd/shortener/...

# Форматирование кода
fmt:
	gofmt -w .
	go run golang.org/x/tools/cmd/goimports@latest -w .

# Запуск приложения
run: build-with-version
	./$(BINARY_NAME)

# Очистка
clean:
	rm -f $(BINARY_NAME) $(STATICLINT_NAME) coverage.out coverage.html

# Показать информацию о сборке
version-info:
	@echo "Build Version: $(BUILD_VERSION)"
	@echo "Build Date: $(BUILD_DATE)" 
	@echo "Build Commit: $(BUILD_COMMIT)"

# Полная сборка (форматирование, тесты, линтинг, сборка)
ci: fmt test lint build-with-version

# Справка
help:
	@echo "Доступные команды:"
	@echo "  build             - Простая сборка без версии"
	@echo "  build-with-version - Сборка с информацией о версии"
	@echo "  build-staticlint  - Сборка staticlint"
	@echo "  test              - Запуск тестов"
	@echo "  test-coverage     - Запуск тестов с покрытием"
	@echo "  lint              - Статический анализ кода"
	@echo "  fmt               - Форматирование кода"
	@echo "  run               - Сборка и запуск приложения"
	@echo "  clean             - Очистка файлов сборки"
	@echo "  version-info      - Показать информацию о сборке"
	@echo "  ci                - Полный цикл: fmt + test + lint + build"
	@echo "  help              - Показать эту справку"
