# URL Shortener Service

Сервис для сокращения URL с поддержкой пользователей, пакетной обработки и различных типов хранилищ.

## Возможности

- Сокращение URL через текстовый API и JSON API
- Пакетное сокращение URL
- Аутентификация пользователей через cookies
- Получение списка URL пользователя
- Асинхронное удаление URL
- Поддержка различных типов хранилищ (память, файл, PostgreSQL)
- Сжатие ответов в gzip
- Логирование запросов
- Проверка доступности базы данных

## Архитектура

Проект построен с использованием чистой архитектуры:

- **Handlers** - HTTP обработчики для различных эндпоинтов
- **Storage** - интерфейс и реализации хранилищ
- **Models** - структуры данных и JSON сериализация
- **Middleware** - промежуточное ПО для аутентификации, сжатия и логирования
- **Config** - конфигурация сервиса
- **Utils** - вспомогательные функции

## API Endpoints

### POST /
Сокращение URL в текстовом формате.

**Request Body:** `https://example.com`

**Response:** `http://localhost:8080/AbCdEfGh`

**Status:** 201 Created

### POST /api/shorten
Сокращение URL через JSON API.

**Request Body:**
```json
{
  "url": "https://example.com"
}
```

**Response:**
```json
{
  "result": "http://localhost:8080/AbCdEfGh"
}
```

**Status:** 201 Created

### POST /api/shorten/batch
Пакетное сокращение URL.

**Request Body:**
```json
[
  {
    "correlation_id": "1",
    "original_url": "https://example1.com"
  },
  {
    "correlation_id": "2", 
    "original_url": "https://example2.com"
  }
]
```

**Response:**
```json
[
  {
    "correlation_id": "1",
    "short_url": "http://localhost:8080/AbCdEfGh"
  },
  {
    "correlation_id": "2",
    "short_url": "http://localhost:8080/IjKlMnOp"
  }
]
```

**Status:** 201 Created

### GET /{shortID}
Перенаправление по сокращенному URL.

**Status:** 307 Temporary Redirect (Location header содержит оригинальный URL)

### GET /api/user/urls
Получение всех URL пользователя.

**Response:**
```json
[
  {
    "short_url": "http://localhost:8080/AbCdEfGh",
    "original_url": "https://example1.com",
    "deleted": false
  }
]
```

**Status:** 200 OK или 204 No Content

### DELETE /api/user/urls
Асинхронное удаление URL пользователя.

**Request Body:** `["AbCdEfGh", "IjKlMnOp"]`

**Status:** 202 Accepted

### GET /ping
Проверка доступности базы данных.

**Status:** 200 OK или 503 Service Unavailable

## Конфигурация

Сервис поддерживает конфигурацию через переменные окружения и флаги командной строки:

| Переменная окружения | Флаг | Описание | По умолчанию |
|---------------------|------|----------|--------------|
| `SERVER_ADDRESS` | `-a` | Адрес HTTP сервера | `localhost:8080` |
| `BASE_URL` | `-b` | Базовый URL для генерации сокращенных ссылок | `http://localhost:8080` |
| `FILE_STORAGE_PATH` | `-f` | Путь к файлу хранилища | `/tmp/short-url-db.json` |
| `DATABASE_DSN` | `-d` | Строка подключения к PostgreSQL | - |

## Запуск

### С компиляцией
```bash
go build -o shortener ./cmd/shortener
./shortener
```

### С переменными окружения
```bash
export SERVER_ADDRESS=:8080
export BASE_URL=https://myshortener.com
export DATABASE_DSN="postgres://user:pass@localhost/shortener"
go run ./cmd/shortener
```

### С флагами командной строки
```bash
go run ./cmd/shortener -a :8080 -b https://myshortener.com -d "postgres://user:pass@localhost/shortener"
```

## Тестирование

Запуск всех тестов:
```bash
go test ./...
```

Запуск тестов с покрытием:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

Запуск примеров:
```bash
go test -run Example ./cmd/shortener/handlers
```

## Форматирование кода

Проект использует стандартные инструменты Go для форматирования кода:

### Автоматическое форматирование
```bash
./format.sh
```

### Ручное форматирование
```bash
# Форматирование с помощью gofmt
gofmt -w -s .

# Форматирование импортов с помощью goimports
go install golang.org/x/tools/cmd/goimports@latest
export PATH=$PATH:$(go env GOPATH)/bin
goimports -w .
```

### Проверка форматирования
```bash
# Проверка gofmt
gofmt -l -s .

# Проверка goimports
goimports -l .
```

## Покрытие тестами

Текущее покрытие тестами: **45.5%**

- handlers: 72.2%
- middleware: 91.7%
- models: 34.3%
- storage: 49.0%
- utils: 100.0%

## Документация

Проект документирован в формате godoc. Для просмотра документации:

```bash
go doc ./cmd/shortener/handlers
go doc ./cmd/shortener/storage
go doc ./cmd/shortener/models
```

## Примеры использования

Смотрите файл `cmd/shortener/handlers/example_test.go` для примеров работы с каждым эндпоинтом.

## Лицензия

MIT
