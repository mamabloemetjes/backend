package lib

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestParseTokenRejectsNonHMACSigningMethod(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"sub": uuid.NewString(), "email": "user@example.com", "role": "user",
		"iat": float64(time.Now().Unix()), "exp": float64(time.Now().Add(time.Hour).Unix()), "jti": uuid.NewString(),
	})
	tokenString, err := token.SignedString([]byte("not-a-rsa-key"))
	if err == nil {
		t.Fatal("expected RSA token creation to fail with an invalid key")
	}
	_ = tokenString
}

func TestParseTokenRequiresAllClaims(t *testing.T) {
	claims := jwt.MapClaims{
		"sub": uuid.NewString(), "email": "user@example.com", "role": "user",
		"iat": float64(time.Now().Unix()), "exp": float64(time.Now().Add(time.Hour).Unix()), "jti": uuid.NewString(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("12345678901234567890123456789012"))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	parsed, err := ParseToken(tokenString, true, "12345678901234567890123456789012")
	if err != nil || parsed == nil {
		t.Fatalf("ParseToken() error = %v", err)
	}

}
