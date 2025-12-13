package utils

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	// ? Characters used for generating short-code
	shortCodeChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	charsLength    = int64(len(shortCodeChars))
)

func IsValidShortCode(code string) bool {
	if len(code) == 0 {
		return false
	}

	for _, char := range code {
		if !strings.ContainsRune(shortCodeChars, char) {
			return false
		}
	}

	return true
}

func GenerateShortCode(length int) (string, error) {
	if length <= 0 {
		length = 7
	}

	result := make([]byte, length)
	for i := range length {
		num, err := rand.Int(rand.Reader, big.NewInt(charsLength))
		if err != nil {
			return "", err
		}
		result[i] = shortCodeChars[num.Int64()]
	}

	return string(result), nil
}
