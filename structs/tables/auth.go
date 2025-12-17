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
	tableName    struct{}  `bun:"table:users,alias:u"`
	Id           uuid.UUID `json:"id" bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Username     string    `json:"username" bun:"username,unique,notnull"`
	Email        string    `json:"email" bun:"email,unique,notnull"`
	PasswordHash string    `json:"-" bun:"password_hash,notnull"`
	Role         string    `json:"role" bun:"role,notnull,default:'user'"`
	CreatedAt    time.Time `json:"created_at" bun:"created_at,notnull,default:now()"`
}
