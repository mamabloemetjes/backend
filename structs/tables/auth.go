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

type Address struct {
	tableName  struct{}   `bun:"table:addresses,alias:a"`
	Id         uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	UserId     *uuid.UUID `bun:"user_id,type:uuid" json:"user_id,omitempty"` // Nullable for guest orders
	Street     string     `bun:"street,notnull" json:"street"`
	HouseNo    string     `bun:"house_no,notnull" json:"house_no"`
	PostalCode string     `bun:"postal_code,notnull" json:"postal_code"`
	City       string     `bun:"city,notnull" json:"city"`
	Country    string     `bun:"country,notnull" json:"country"` // "NL"
	CreatedAt  time.Time  `bun:"created_at,notnull,default:now()" json:"created_at"`
	UpdatedAt  time.Time  `bun:"updated_at,notnull,default:now()" json:"updated_at"`
}
