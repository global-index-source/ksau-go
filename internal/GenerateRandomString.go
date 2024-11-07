package internal

import (
	"time"

	"golang.org/x/exp/rand"
)

// Randomize the seed
func init() {
	rand.Seed(uint64(time.Now().UnixNano()))
}

var alphanumericRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")

// Iterate over the letterRunes and generate a random string
func GenerateRandomString(length int) string {
	randomRunesBuffer := make([]rune, length)

	for i := range randomRunesBuffer {
		randomRunesBuffer[i] = alphanumericRunes[rand.Intn(len(alphanumericRunes))]
	}

	return string(randomRunesBuffer) // Convert the buffer to a string
}
