# gRPC Support

Проект теперь поддерживает gRPC API в дополнение к HTTP API. Все функции доступны через оба протокола.

## Запуск с gRPC

### Через флаги командной строки:
```bash
./shortener -grpc -grpc-addr localhost:9090 -a localhost:8080
```

### Через переменные окружения:
```bash
ENABLE_GRPC=true GRPC_ADDRESS=localhost:9090 ./shortener
```

### Через JSON конфигурацию:
```json
{
  "server_address": "localhost:8080",
  "base_url": "http://localhost:8080",
  "enable_grpc": true,
  "grpc_address": "localhost:9090"
}
```

## Доступные gRPC методы

- `ShortenURL` - сокращение URL
- `GetURL` - получение оригинального URL по сокращенному
- `BatchShortenURL` - пакетное сокращение URL
- `GetUserURLs` - получение URL пользователя
- `DeleteUserURLs` - удаление URL пользователя
- `Ping` - проверка доступности сервиса
- `GetStats` - получение статистики

## Тестирование gRPC

Для тестирования gRPC API можно использовать встроенный клиент:

```bash
go build -o grpc-client ./cmd/grpc-client
./grpc-client
```

## Конфигурация

gRPC сервер поддерживает все те же конфигурации, что и HTTP сервер:
- Адрес сервера
- Базовый URL
- Хранилище (файловое, PostgreSQL)
- HTTPS (для HTTP сервера)
- Доверенная подсеть для статистики

## Особенности

- gRPC и HTTP серверы работают независимо
- Можно запускать только HTTP, только gRPC или оба одновременно
- Graceful shutdown поддерживается для обоих серверов
- UserID автоматически генерируется для gRPC запросов
