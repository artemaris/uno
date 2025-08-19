package utils

import (
	"crypto/rand"
	"encoding/base64"
)

const idLength = 8

func GenerateShortID() string {
	bytes := make([]byte, idLength)
	rand.Read(bytes)
	return base64.URLEncoding.EncodeToString(bytes)[:idLength]
}
