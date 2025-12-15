package structs

import (
	"time"

	"github.com/google/uuid"
)

type ArgonParams struct {
	Memory  uint32
	Time    uint32
	Threads uint8
	KeyLen  uint32
	SaltLen uint32
}

type AuthClaims struct {
	Sub   uuid.UUID `json:"sub"`
	Email string    `json:"email"`
	Role  string    `json:"role"`
	Iat   time.Time `json:"iat"`
	Exp   time.Time `json:"exp"`
	Jti   uuid.UUID `json:"jti"`
}

type AuthRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}
