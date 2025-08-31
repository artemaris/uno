package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Hello, World!")

	// Это должно вызвать ошибку анализатора
	os.Exit(1) // want "прямой вызов os.Exit в функции main запрещен"
}

func notMain() {
	// Это НЕ должно вызвать ошибку, так как это не функция main
	os.Exit(1)
}

// В не-main пакетах os.Exit разрешен
