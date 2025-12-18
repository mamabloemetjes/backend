package lib

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateRandomToken generates a cryptographically secure random token
func GenerateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
