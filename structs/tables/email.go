package tables

import (
	"time"

	"github.com/google/uuid"
)

type EmailVerification struct {
	tableName struct{}  `bun:"table:email_verifications,alias:ev"`
	Id        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserId    uuid.UUID `bun:"user_id,notnull,type:uuid"`
	Token     string    `bun:"token,notnull,unique"`
	ExpiresAt time.Time `bun:"expires_at,notnull"`
	Used      bool      `bun:"used,notnull,default:false"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	User      *User     `bun:"rel:belongs-to,join:user_id=id,on_delete:cascade"`
}
