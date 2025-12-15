package lib

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("invalid hash format")
	ErrIncompatibleVersion = errors.New("incompatible version of argon2")
)

// Argon2HashParts contains the decoded parts of an Argon2 hash
type Argon2HashParts struct {
	Memory  uint32
	Time    uint32
	Threads uint8
	KeyLen  uint32
	Salt    []byte
	Hash    []byte
}

// DecodeArgon2Hash decodes an Argon2id hash string into its component parts
// Expected format: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
func DecodeArgon2Hash(encodedHash string) (*Argon2HashParts, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, ErrInvalidHash
	}

	// Check algorithm
	if parts[1] != "argon2id" {
		return nil, ErrInvalidHash
	}

	// Check version
	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, err
	}
	if version != argon2.Version {
		return nil, ErrIncompatibleVersion
	}

	// Parse parameters
	var memory, time uint32
	var threads uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &time, &threads)
	if err != nil {
		return nil, err
	}

	// Decode salt
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, err
	}

	// Decode hash
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, err
	}

	return &Argon2HashParts{
		Memory:  memory,
		Time:    time,
		Threads: threads,
		KeyLen:  uint32(len(hash)),
		Salt:    salt,
		Hash:    hash,
	}, nil
}

// SecureCompare performs a constant-time comparison of two byte slices
// This prevents timing attacks when comparing password hashes
func SecureCompare(a, b []byte) bool {
	return subtle.ConstantTimeCompare(a, b) == 1
}
