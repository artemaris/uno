package noosexit

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	// Запускаем тест анализатора с тестовыми данными
	analysistest.Run(t, analysistest.TestData(), Analyzer, "a", "b")
}
