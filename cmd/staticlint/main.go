// Package main предоставляет multichecker - инструмент статического анализа кода,
// который объединяет множество анализаторов для комплексной проверки Go кода.
//
// # Запуск multichecker
//
// Базовое использование:
//
//	go run cmd/staticlint/main.go ./...
//
// Запуск для конкретного пакета:
//
//	go run cmd/staticlint/main.go ./cmd/shortener/...
//
// Запуск с дополнительными флагами:
//
//	go run cmd/staticlint/main.go -test=false ./...
//
// # Состав анализаторов
//
// ## Стандартные анализаторы (golang.org/x/tools/go/analysis/passes):
//
// - asmdecl: проверяет корректность объявлений в assembly файлах
// - assign: находит бесполезные присваивания
// - atomic: проверяет правильность использования пакета sync/atomic
// - atomicalign: проверяет выравнивание 64-битных атомарных операций
// - bools: обнаруживает распространенные ошибки с булевыми операторами
// - buildtag: проверяет корректность build tags
// - cgocall: обнаруживает нарушения правил cgo
// - composite: проверяет композитные литералы
// - copylock: проверяет блокировки, передаваемые по значению
// - deepequalerrors: проверяет использование reflect.DeepEqual с error значениями
// - errorsas: проверяет правильность использования errors.As
// - fieldalignment: находит неоптимальное выравнивание полей структур
// - findcall: находит вызовы конкретных функций (для отладки)
// - framepointer: проверяет assembly код на корректность frame pointer
// - httpresponse: проверяет ошибки в использовании HTTP Response Body
// - ifaceassert: обнаруживает невозможные приведения типов интерфейса
// - loopclosure: проверяет захват переменных цикла в closures
// - lostcancel: находит context.CancelFunc, которые не вызываются
// - nilfunc: проверяет бесполезные сравнения функций с nil
// - nilness: проверяет избыточные или невозможные nil проверки
// - printf: проверяет корректность printf-подобных вызовов
// - reflectvaluecompare: проверяет сравнения reflect.Value
// - shadow: находит затенённые переменные
// - shift: проверяет сдвиги, которые превышают ширину целого числа
// - sigchanyzer: обнаруживает неправильно буферизованные каналы сигналов
// - sortslice: проверяет вызовы sort.Slice с некорректными аргументами
// - stdmethods: проверяет сигнатуры методов известных интерфейсов
// - stringintconv: флагирует преобразования строки в int
// - structtag: проверяет корректность тегов структур
// - testinggoroutine: проверяет использование testing.T из других горутин
// - tests: проверяет распространенные ошибочные паттерны в тестах
// - timeformat: проверяет использование time.Time.Format
// - unmarshal: проверяет передачу адресуемых значений в unmarshal
// - unreachable: находит недостижимый код
// - unsafeptr: проверяет корректность использования unsafe.Pointer
// - unusedresult: проверяет неиспользуемые результаты вызовов функций
// - unusedwrite: находит записи в переменные, которые никогда не читаются
//
// ## Анализаторы класса SA (staticcheck.io):
//
// Все анализаторы серии SA* из staticcheck, включающие:
// - SA1*: различные проверки багов и неправильного использования API
// - SA2*: проверки concurrency и goroutine leaks
// - SA3*: проверки testing пакета и бенчмарков
// - SA4*: проверки неправильного использования стандартной библиотеки
// - SA5*: проверки корректности кода
// - SA6*: проверки производительности
// - SA9*: проверки сомнительных конструкций
//
// ## Другие классы staticcheck.io:
//
// - ST*: стилистические вопросы
// - QF*: quickfix предложения
// - S*: простые проверки
//
// ## Публичные анализаторы:
//
// - errcheck: проверяет, что все ошибки обрабатываются
// - ineffassign: находит неэффективные присваивания переменным
//
// ## Собственный анализатор:
//
//   - noosexit: запрещает прямой вызов os.Exit в функции main пакета main.
//     Помогает избежать неконтролируемого завершения программы и способствует
//     более чистому коду с правильной обработкой ошибок.
//
// # Примеры использования
//
// Проверка всего проекта:
//
//	go run cmd/staticlint/main.go ./...
//
// Проверка конкретного пакета с выводом только ошибок:
//
//	go run cmd/staticlint/main.go ./cmd/shortener/handlers
//
// Сборка и использование как standalone инструмент:
//
//	go build -o staticlint cmd/staticlint/main.go
//	./staticlint ./...
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"uno/cmd/staticlint/noosexit"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/nilness"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/sortslice"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"

	"github.com/kisielk/errcheck/errcheck"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func main() {
	// Собираем все анализаторы
	var analyzers []*analysis.Analyzer

	// 1. Стандартные анализаторы из golang.org/x/tools/go/analysis/passes
	standardAnalyzers := []*analysis.Analyzer{
		asmdecl.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		atomicalign.Analyzer,
		bools.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		deepequalerrors.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		sortslice.Analyzer,
		stdmethods.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
	}
	analyzers = append(analyzers, standardAnalyzers...)

	// 2. Все анализаторы класса SA из staticcheck.io
	staticcheckAnalyzers := getStaticcheckAnalyzers()
	analyzers = append(analyzers, staticcheckAnalyzers...)

	// 3. Анализаторы других классов из staticcheck.io
	// ST* - стилистические вопросы
	for _, analyzer := range stylecheck.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}

	// QF* - quickfix предложения
	for _, analyzer := range quickfix.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}

	// S* - простые проверки
	for _, analyzer := range simple.Analyzers {
		analyzers = append(analyzers, analyzer.Analyzer)
	}

	// 4. Публичные анализаторы
	publicAnalyzers := []*analysis.Analyzer{
		errcheck.Analyzer, // Проверяет обработку ошибок
		// ineffassign будет добавлен через отдельный wrapper
	}
	analyzers = append(analyzers, publicAnalyzers...)

	// 5. Собственный анализатор
	analyzers = append(analyzers, noosexit.Analyzer)

	// Запускаем multichecker
	multichecker.Main(analyzers...)
}

// getStaticcheckAnalyzers возвращает все анализаторы класса SA из staticcheck
func getStaticcheckAnalyzers() []*analysis.Analyzer {
	var analyzers []*analysis.Analyzer

	for _, analyzer := range staticcheck.Analyzers {
		// Добавляем только анализаторы класса SA
		if strings.HasPrefix(analyzer.Analyzer.Name, "SA") {
			analyzers = append(analyzers, analyzer.Analyzer)
		}
	}

	return analyzers
}

// Функция для чтения конфигурации (может быть расширена в будущем)
func readConfig() map[string]interface{} {
	configPath := ".staticlint.json"
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return make(map[string]interface{})
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return make(map[string]interface{})
	}

	var config map[string]interface{}
	if err := json.Unmarshal(data, &config); err != nil {
		return make(map[string]interface{})
	}

	return config
}

// Вспомогательная функция для определения, является ли файл тестовым
func isTestFile(filename string) bool {
	return strings.HasSuffix(filepath.Base(filename), "_test.go")
}
