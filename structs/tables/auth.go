package tables

import (
	"time"

	"github.com/google/uuid"
)

type AuthResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type User struct {
	tableName     struct{}  `bun:"table:users,alias:u"`
	Id            uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Username      string    `json:"username" bun:"username,unique,notnull"`
	Email         string    `json:"email" bun:"email,unique,notnull"`
	PasswordHash  string    `json:"-" bun:"password_hash,notnull"`
	Role          string    `json:"role" bun:"role,notnull,default:'user'"`
	LastLogin     time.Time `json:"last_login" bun:"last_login,default:now()"`
	EmailVerified bool      `json:"email_verified" bun:"email_verified,notnull,default:false"`
	CreatedAt     time.Time `json:"created_at" bun:"created_at,notnull,default:now()"`
}
