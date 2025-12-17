package lib

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateCSRFToken generates a cryptographically secure random token
func GenerateCSRFToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate CSRF token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}
