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
	tableName    struct{}  `pg:"users,alias:u"`
	Id           uuid.UUID `json:"id" pg:"id,pk,type:uuid,default:gen_random_uuid()"`
	Username     string    `json:"username" pg:"username,unique,notnull"`
	Email        string    `json:"email" pg:"email,unique,notnull"`
	PasswordHash string    `json:"-" pg:"password_hash,notnull"`
	Role         string    `json:"role" pg:"role,notnull,default:'user'"`
	CreatedAt    time.Time `json:"created_at" pg:"created_at,notnull,default:now()"`
}
