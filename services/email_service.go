package services

import (
	"context"
	"fmt"
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs"
	"mamabloemetjes_server/structs/tables"
	"sync"
	"time"

	"github.com/MonkyMars/gecho"
	"github.com/resend/resend-go/v3"
)

var (
	client     *resend.Client
	clientOnce = sync.Once{}
)

type EmailService struct {
	logger      *gecho.Logger
	cfg         *structs.Config
	client      *resend.Client
	db          *database.DB
	authService *AuthService
}

func NewEmailService(logger *gecho.Logger, cfg *structs.Config, db *database.DB) *EmailService {
	return &EmailService{
		logger:      logger,
		cfg:         cfg,
		db:          db,
		client:      getEmailClient(cfg.Email.ApiKey),
		authService: NewAuthService(cfg, logger, db),
	}
}

func getEmailClient(apiKey string) *resend.Client {
	clientOnce.Do(func() {
		client = resend.NewClient(apiKey)
	})
	return client
}

func (es *EmailService) SendEmail(to string, subject string, body string) error {

	params := &resend.SendEmailRequest{
		From:    es.cfg.Email.From,
		To:      []string{to},
		Html:    body,
		Subject: subject,
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		es.logger.Error("Failed to send email", gecho.Field("error", err), gecho.Field("to", to))
		return err
	}

	return nil
}

func (es *EmailService) SendVerificationEmail(user *tables.User) (*tables.EmailVerification, error) {
	token, err := lib.GenerateRandomToken()
	if err != nil {
		es.logger.Error("Failed to generate verification token", gecho.Field("error", err))
		return nil, err
	}

	// Calculate expiration time
	expiration := time.Now().Add(es.cfg.Email.VerificationTokenExpiry)

	// Create struct
	emailVerif := &tables.EmailVerification{
		UserId:    user.Id,
		Token:     token,
		ExpiresAt: expiration,
		CreatedAt: time.Now(),
	}

	// Store hashed token in database with association to user
	result, err := database.Query[tables.EmailVerification](es.db).Insert(context.Background(), emailVerif)
	if err != nil {
		es.logger.Error("Failed to store email verification token", gecho.Field("error", err))
		return nil, err
	}

	// Send email with link
	verificationLink := fmt.Sprintf("%s/auth/verify-email?token=%s&user_id=%s", es.cfg.Server.ServerURL, token, user.Id.String())

	emailBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
		</head>
		<body>
			<p>Please verify your email by clicking the following link:</p>
			<p><a href="%s">Verify Email</a></p>
			<p>This link will expire in %.0f minutes.</p>
			<p>If you did not create an account, please ignore this email.</p>

			<p>Link not working? Copy and paste the following URL into your browser:</p>
			<p>%s</p>

			<p>Best regards,<br/>The MamaBloemetjes Team</p>
		</body>
		</html>
	`, verificationLink, time.Until(expiration).Minutes(), verificationLink)

	err = es.SendEmail(user.Email, "Verify your email", emailBody)
	if err != nil {
		es.logger.Error("Failed to send verification email", gecho.Field("error", err), gecho.Field("to", user.Email))
		return nil, err
	}

	return result, err
}
