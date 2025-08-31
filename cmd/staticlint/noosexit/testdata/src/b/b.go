package notmain

import "os"

func main() {
	// Это НЕ должно вызвать ошибку, так как пакет не main
	os.Exit(1)
}

func someFunction() {
	os.Exit(0)
}
