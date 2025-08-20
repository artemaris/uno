package utils

import (
	"crypto/rand"
	"math/big"
)

const idLength = 8
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateShortID() string {
	id := make([]byte, idLength)
	charsetLen := big.NewInt(int64(len(charset)))

	for i := range id {
		randomIndex, _ := rand.Int(rand.Reader, charsetLen)
		id[i] = charset[randomIndex.Int64()]
	}

	return string(id)
}
