package tables

import (
	"time"

	"github.com/google/uuid"
)

type EmailVerification struct {
	tableName struct{}  `bun:"table:email_verifications,alias:ev"`
	Id        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" validate:"omitempty,uuid4"`
	UserId    uuid.UUID `bun:"user_id,notnull,type:uuid" validate:"required,uuid4"`
	Token     string    `bun:"token,notnull,unique" validate:"required,min=32"`
	ExpiresAt time.Time `bun:"expires_at,notnull" validate:"required"`
	Used      bool      `bun:"used,notnull,default:false"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	User      *User     `bun:"rel:belongs-to,join:user_id=id,on_delete:cascade" validate:"omitempty"`
}
