#!/bin/bash

# Скрипт для автоматического форматирования Go кода
# Использует gofmt и goimports для приведения кода к стандартам Go

echo "Форматирование Go кода..."

# Проверяем, установлен ли goimports
if ! command -v goimports &> /dev/null; then
    echo "Устанавливаем goimports..."
    go install golang.org/x/tools/cmd/goimports@latest
    export PATH=$PATH:$(go env GOPATH)/bin
fi

# Форматируем код с помощью gofmt
echo "Запуск gofmt..."
gofmt -w -s .

# Форматируем импорты с помощью goimports
echo "Запуск goimports..."
goimports -w .

echo "Форматирование завершено!"

# Проверяем результат
echo "Проверка форматирования..."
if [ -n "$(gofmt -l -s .)" ]; then
    echo "Ошибка: найдены неотформатированные файлы:"
    gofmt -l -s .
    exit 1
fi

if [ -n "$(goimports -l .)" ]; then
    echo "Ошибка: найдены файлы с неправильными импортами:"
    goimports -l .
    exit 1
fi

echo "Все файлы отформатированы правильно!"
