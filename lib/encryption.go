package lib

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
)

// Encrypt encrypts plaintext using AES-GCM
func Encrypt(plaintext string, key string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return "", errors.New("encryption key must be 32 bytes for AES-256")
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// Create a nonce
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	// Encrypt and prepend nonce to ciphertext
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)

	// Return base64 encoded string
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-GCM
func Decrypt(ciphertext string, key string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	keyBytes := []byte(key)
	if len(keyBytes) != 32 {
		return "", errors.New("encryption key must be 32 bytes for AES-256")
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(keyBytes)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// MustEncrypt encrypts or panics (for testing/development)
func MustEncrypt(plaintext string, key string) string {
	encrypted, err := Encrypt(plaintext, key)
	if err != nil {
		panic(err)
	}
	return encrypted
}

// MustDecrypt decrypts or panics (for testing/development)
func MustDecrypt(ciphertext string, key string) string {
	decrypted, err := Decrypt(ciphertext, key)
	if err != nil {
		panic(err)
	}
	return decrypted
}
