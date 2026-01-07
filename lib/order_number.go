package lib

import (
	"fmt"
	"math/rand"
	"time"
)

// GenerateOrderNumber generates a unique order number in the format: MB-XXXX
// where XXXX is a random 4-character alphanumeric string
func GenerateOrderNumber() string {
	// Use a local rand.Source + rand.Rand for thread safety
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)

	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 4

	// Generate random part
	randomPart := make([]byte, length)
	for i := range randomPart {
		randomPart[i] = chars[r.Intn(len(chars))]
	}

	return fmt.Sprintf("MB-%s", string(randomPart))
}
