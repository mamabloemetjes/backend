package lib

import (
	"fmt"
	"mamabloemetjes_server/structs"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ParseToken parses and validates a JWT token string and returns the claims
func ParseToken(tokenStr string, isAccessToken bool, secret string) (*structs.AuthClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenMalformed
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Safely extract and validate claims
		subStr, ok := claims["sub"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid sub claim")
		}

		sub, err := uuid.Parse(subStr)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID in sub claim: %w", err)
		}

		email, ok := claims["email"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid email claim")
		}

		role, ok := claims["role"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid role claim")
		}

		iat, ok := claims["iat"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid iat claim")
		}

		exp, ok := claims["exp"].(float64)
		if !ok {
			return nil, fmt.Errorf("invalid exp claim")
		}

		jtiStr, ok := claims["jti"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid jti claim")
		}

		jti, err := uuid.Parse(jtiStr)
		if err != nil {
			return nil, fmt.Errorf("invalid UUID in jti claim: %w", err)
		}

		return &structs.AuthClaims{
			Sub:   sub,
			Email: email,
			Role:  role,
			Iat:   time.Unix(int64(iat), 0),
			Exp:   time.Unix(int64(exp), 0),
			Jti:   jti,
		}, nil
	}
	return nil, jwt.ErrInvalidKey
}

func ExtractClaims(r *http.Request, secret string) (*structs.AuthClaims, error) {
	accessToken, err := GetCookieValue(AccessCookieName, r)
	if err != nil {
		return nil, err
	}

	claims, err := ParseToken(
		accessToken,
		true,
		secret,
	)
	if err != nil {
		return nil, err
	}

	return claims, nil
}
