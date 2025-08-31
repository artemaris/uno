# Staticlint - Multichecker для статического анализа Go кода

Staticlint - это комплексный инструмент статического анализа, объединяющий множество анализаторов для тщательной проверки Go кода.

## Быстрый старт

### Сборка

```bash
go build -o staticlint cmd/staticlint/main.go
```

### Использование

```bash
# Анализ всего проекта
./staticlint ./...

# Анализ конкретного пакета
./staticlint ./cmd/shortener/handlers

# Анализ с включением только собственного анализатора
./staticlint -noosexit ./cmd/shortener/main.go
```

## Состав анализаторов

### Стандартные анализаторы (golang.org/x/tools/go/analysis/passes)

#### Безопасность и корректность
- **asmdecl** - проверяет корректность объявлений в assembly файлах
- **atomic** - проверяет правильность использования пакета sync/atomic
- **atomicalign** - проверяет выравнивание 64-битных атомарных операций
- **cgocall** - обнаруживает нарушения правил cgo
- **copylock** - проверяет блокировки, передаваемые по значению
- **lostcancel** - находит context.CancelFunc, которые не вызываются
- **unsafeptr** - проверяет корректность использования unsafe.Pointer

#### Логические ошибки
- **assign** - находит бесполезные присваивания
- **bools** - обнаруживает распространенные ошибки с булевыми операторами
- **composite** - проверяет композитные литералы
- **deepequalerrors** - проверяет использование reflect.DeepEqual с error значениями
- **errorsas** - проверяет правильность использования errors.As
- **ifaceassert** - обнаруживает невозможные приведения типов интерфейса
- **loopclosure** - проверяет захват переменных цикла в closures
- **nilfunc** - проверяет бесполезные сравнения функций с nil
- **nilness** - проверяет избыточные или невозможные nil проверки

#### Форматирование и стиль
- **printf** - проверяет корректность printf-подобных вызовов
- **reflectvaluecompare** - проверяет сравнения reflect.Value
- **shadow** - находит затенённые переменные
- **stringintconv** - флагирует преобразования строки в int
- **structtag** - проверяет корректность тегов структур

#### Оптимизация и производительность
- **fieldalignment** - находит неоптимальное выравнивание полей структур
- **shift** - проверяет сдвиги, которые превышают ширину целого числа
- **unusedresult** - проверяет неиспользуемые результаты вызовов функций
- **unusedwrite** - находит записи в переменные, которые никогда не читаются

#### Тестирование
- **testinggoroutine** - проверяет использование testing.T из других горутин
- **tests** - проверяет распространенные ошибочные паттерны в тестах

#### Специализированные
- **buildtag** - проверяет корректность build tags
- **findcall** - находит вызовы конкретных функций (для отладки)
- **framepointer** - проверяет assembly код на корректность frame pointer
- **httpresponse** - проверяет ошибки в использовании HTTP Response Body
- **sigchanyzer** - обнаруживает неправильно буферизованные каналы сигналов
- **sortslice** - проверяет вызовы sort.Slice с некорректными аргументами
- **stdmethods** - проверяет сигнатуры методов известных интерфейсов
- **timeformat** - проверяет использование time.Time.Format
- **unmarshal** - проверяет передачу адресуемых значений в unmarshal
- **unreachable** - находит недостижимый код

### Анализаторы Staticcheck (honnef.co/go/tools)

#### Класс SA* (Ошибки и неправильное использование)
Все анализаторы серии SA*, включая:
- **SA1\*** - различные проверки багов и неправильного использования API
- **SA2\*** - проверки concurrency и goroutine leaks
- **SA3\*** - проверки testing пакета и бенчмарков
- **SA4\*** - проверки неправильного использования стандартной библиотеки
- **SA5\*** - проверки корректности кода
- **SA6\*** - проверки производительности
- **SA9\*** - проверки сомнительных конструкций

#### Класс ST* (Стилистические проверки)
- Проверки соответствия стилевым рекомендациям Go
- Правила именования
- Форматирование комментариев

#### Класс QF* (Quickfix предложения)
- Автоматические исправления кода
- Предложения по улучшению

#### Класс S* (Простые проверки)
- Упрощения кода
- Удаление избыточности

### Публичные анализаторы

#### errcheck
Проверяет, что все возвращаемые ошибки обрабатываются.

**Примеры проблем:**
```go
// Плохо
file, err := os.Open("file.txt")
defer file.Close() // err не проверен!

// Хорошо
file, err := os.Open("file.txt")
if err != nil {
    return err
}
defer file.Close()
```

### Собственный анализатор

#### noosexit
Запрещает прямой вызов `os.Exit` в функции `main` пакета `main`.

**Обоснование:**
- Способствует более чистому коду с правильной обработкой ошибок
- Предотвращает неконтролируемое завершение программы
- Облегчает тестирование приложения

**Примеры:**
```go
// Плохо - будет обнаружено анализатором
package main

import "os"

func main() {
    os.Exit(1) // Ошибка: прямой вызов os.Exit в функции main запрещен
}

// Хорошо
package main

import (
    "fmt"
    "os"
)

func main() {
    if err := run(); err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1) // Допустимо в функции main при обработке ошибок
    }
}

func run() error {
    // Основная логика приложения
    return nil
}
```

## Примеры использования

### Базовый анализ
```bash
./staticlint ./...
```

### Анализ с фокусом на определенные классы проблем
```bash
# Только ошибки безопасности (SA*)
./staticlint -SA1001 -SA1002 -SA1003 ./...

# Только стилистические проблемы
./staticlint -ST1000 -ST1001 ./...

# Только собственный анализатор
./staticlint -noosexit ./...
```

### Анализ с исправлениями
```bash
# Применить автоматические исправления
./staticlint -fix ./...
```

### Анализ конкретных файлов
```bash
./staticlint ./cmd/shortener/main.go
./staticlint ./cmd/shortener/handlers/*.go
```

## Конфигурация

Staticlint поддерживает конфигурационный файл `.staticlint.json` в корне проекта:

```json
{
  "exclude": [
    "vendor/",
    "testdata/"
  ],
  "enable_all": false,
  "custom_rules": {
    "noosexit": {
      "enabled": true
    }
  }
}
```

## Интеграция в CI/CD

### GitHub Actions
```yaml
name: Static Analysis
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21
    - name: Build staticlint
      run: go build -o staticlint cmd/staticlint/main.go
    - name: Run static analysis
      run: ./staticlint ./...
```

### Makefile интеграция
```makefile
.PHONY: lint
lint:
	go build -o staticlint cmd/staticlint/main.go
	./staticlint ./...

.PHONY: lint-fix
lint-fix:
	go build -o staticlint cmd/staticlint/main.go
	./staticlint -fix ./...
```

## Расширение

Для добавления собственных анализаторов:

1. Создайте новый анализатор в директории `cmd/staticlint/analyzers/`
2. Зарегистрируйте его в `main.go`
3. Добавьте тесты и документацию

Пример структуры нового анализатора:
```go
package myanalyzer

import (
    "golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
    Name: "myanalyzer",
    Doc:  "описание анализатора",
    Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
    // Логика анализатора
    return nil, nil
}
```

## Отладка

Для отладки анализаторов используйте флаги:

```bash
# Подробный вывод
./staticlint -debug=fpstv ./...

# JSON вывод для автоматической обработки
./staticlint -json ./...

# Профилирование
./staticlint -cpuprofile=cpu.prof ./...
```

## Лицензия

Этот проект использует анализаторы с различными лицензиями:
- Стандартные анализаторы Go: BSD-3-Clause
- Staticcheck: MIT
- errcheck: MIT
- Собственные анализаторы: MIT
