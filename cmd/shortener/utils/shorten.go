package utils

import (
	"crypto/rand"
	"math/big"
)

const idLength = 8
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateShortID генерирует случайный сокращенный идентификатор длиной 8 символов
// Использует криптографически стойкий генератор случайных чисел для создания
// уникальных идентификаторов из набора букв и цифр
// Возвращает ошибку, если генератор случайных чисел недоступен
func GenerateShortID() (string, error) {
	id := make([]byte, idLength)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range id {
		randomIndex, err := rand.Int(rand.Reader, charsetLen)
		if err != nil {
			return "", err
		}
		id[i] = charset[randomIndex.Int64()]
	}

	return string(id), nil
}
