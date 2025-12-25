package lib

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

// GenerateRandomToken generates a cryptographically secure random token
func GenerateRandomToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// GenerateSKU generates a SKU from a product name and optional suffix length
func GenerateSKU(productName string, suffixLength int) (string, error) {
	// Take first 3 letters of the product name (uppercase, alphanumeric only)
	namePart := strings.ToUpper(productName)
	namePart = strings.Map(func(r rune) rune {
		if r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' {
			return r
		}
		return -1
	}, namePart)
	if len(namePart) > 3 {
		namePart = namePart[:3]
	}

	// Generate random suffix
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, suffixLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	for i := range b {
		b[i] = letters[int(b[i])%len(letters)]
	}

	return fmt.Sprintf("%s-%s", namePart, string(b)), nil
}
