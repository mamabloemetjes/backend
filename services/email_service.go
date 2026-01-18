package services

import (
	"context"
	"fmt"
	"mamabloemetjes_server/database"
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs"
	"mamabloemetjes_server/structs/tables"
	"strings"
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

func (es *EmailService) SendEmail(to []string, subject string, body string) error {
	params := &resend.SendEmailRequest{
		From:    es.cfg.Email.From,
		To:      to,
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
	resendLink := fmt.Sprintf("%s/email/resend?id=%s", es.cfg.Server.FrontendURL, user.Id.String())

	emailBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
				.content { padding: 20px; background-color: #f9f9f9; }
				.button { display: inline-block; padding: 15px 30px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
				.footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
				.divider { margin: 30px 0; border-top: 2px solid #ddd; }
			</style>
		</head>
		<body>
			<div class="container">
				<!-- Dutch Version -->
				<div class="header">
					<h1>Verifieer je e-mailadres</h1>
				</div>
				<div class="content">
					<p>Verifieer je e-mailadres door op de volgende link te klikken:</p>
					<p style="text-align: center;">
						<a href="%s" class="button">Verifieer E-mail</a>
					</p>
					<p>Deze link verloopt over %.0f minuten.</p>
					<p>Als je geen account hebt aangemaakt, kun je deze e-mail negeren.</p>

					<p>Link werkt niet? Kopieer en plak de volgende URL in je browser:</p>
					<p style="word-break: break-all;">%s</p>

					<p style="margin-top: 20px; padding: 15px; background-color: #f0f0f0; border-left: 4px solid #4CAF50;">
						<strong>Link verlopen?</strong><br>
						Als deze verificatielink is verlopen, kun je een nieuwe aanvragen: <a href="%s" style="color: #4CAF50; text-decoration: underline;">Klik hier om een nieuwe verificatie e-mail te ontvangen</a>
					</p>
				</div>

				<div class="divider"></div>

				<!-- English Version -->
				<div class="header">
					<h1>Verify your email address</h1>
				</div>
				<div class="content">
					<p>Please verify your email by clicking the following link:</p>
					<p style="text-align: center;">
						<a href="%s" class="button">Verify Email</a>
					</p>
					<p>This link will expire in %.0f minutes.</p>
					<p>If you did not create an account, please ignore this email.</p>

					<p>Link not working? Copy and paste the following URL into your browser:</p>
					<p style="word-break: break-all;">%s</p>

					<p style="margin-top: 20px; padding: 15px; background-color: #f0f0f0; border-left: 4px solid #4CAF50;">
						<strong>Link expired?</strong><br>
						If this verification link has expired, you can request a new one: <a href="%s" style="color: #4CAF50; text-decoration: underline;">Click here to receive a new verification email</a>
					</p>
				</div>

				<div class="footer">
					<p>MamaBloemetjes | Fresh Flowers Delivered with Love</p>
				</div>
			</div>
		</body>
		</html>
	`, verificationLink, time.Until(expiration).Minutes(), verificationLink, resendLink, verificationLink, time.Until(expiration).Minutes(), verificationLink, resendLink)

	err = es.SendEmail([]string{user.Email}, "Verify your email", emailBody)
	if err != nil {
		es.logger.Error("Failed to send verification email", gecho.Field("error", err), gecho.Field("to", user.Email))
		return nil, err
	}

	return result, err
}

// SendOrderConfirmationEmail sends a bilingual order confirmation email
func (es *EmailService) SendOrderConfirmationEmail(email, name, orderNumber string, orderLines []*tables.OrderLine, address *tables.Address) error {
	// Calculate total
	var total uint64
	for _, line := range orderLines {
		total += line.LineTotal
	}

	// Format total as currency
	totalFormatted := fmt.Sprintf("€%.2f", float64(total)/100)

	// Build order items list
	itemsListNL := ""
	itemsListEN := ""
	var itemsBuilderNL, itemsBuilderEN strings.Builder
	for _, line := range orderLines {
		lineTotal := fmt.Sprintf("€%.2f", float64(line.LineTotal)/100)
		fmt.Fprintf(&itemsBuilderNL, "<li>%dx %s - %s</li>", line.Quantity, line.ProductName, lineTotal)
		fmt.Fprintf(&itemsBuilderEN, "<li>%dx %s - %s</li>", line.Quantity, line.ProductName, lineTotal)
	}
	itemsListNL = itemsBuilderNL.String()
	itemsListEN = itemsBuilderEN.String()

	// Format address
	addressFormatted := fmt.Sprintf("%s %s<br>%s %s<br>%s",
		address.Street, address.HouseNo, address.PostalCode, address.City, address.Country)

	emailBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
				.content { padding: 20px; background-color: #f9f9f9; }
				.order-details { background-color: white; padding: 15px; margin: 15px 0; border-radius: 5px; }
				.footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
				ul { list-style-type: none; padding: 0; }
				li { padding: 5px 0; border-bottom: 1px solid #eee; }
				.divider { margin: 30px 0; border-top: 2px solid #ddd; }
			</style>
		</head>
		<body>
			<div class="container">
				<!-- Dutch Version -->
				<div class="header">
					<h1>Bedankt voor je bestelling!</h1>
				</div>
				<div class="content">
					<p>Beste %s,</p>
					<p>Je bestelling is ontvangen. Hieronder vind je de details van je bestelling.</p>

					<div class="order-details">
						<h3>Bestelnummer: <strong>%s</strong></h3>
						<h4>Bestellijst:</h4>
						<ul>%s</ul>
						<p><strong>Totaal: %s</strong></p>

						<h4>Bezorgadres:</h4>
						<p>%s</p>
					</div>

					<p><strong>Betaling via Tikkie:</strong></p>
					<p>Je ontvangt binnenkort een e-mail met een Tikkie betaallink. Zodra de betaling is ontvangen, ga ik aan de slag met je bestelling!</p>

					<p>Vragen? Neem contact met mij op via %s</p>
				</div>

				<div class="divider"></div>

				<!-- English Version -->
				<div class="header">SendEmail
					<h1>Thank you for your order!</h1>
				</div>user
				<div class="content">
					<p>Dear %s,</p>
					<p>The order has beenr recieved. Below you will find the details of your order.</p>

					<div class="order-details">
						<h3>Order Number: <strong>%s</strong></h3>
						<h4>Order Items:</h4>
						<ul>%s</ul>
						<p><strong>Total: %s</strong></p>

						<h4>Delivery Address:</h4>
						<p>%s</p>
					</div>

					<p><strong>Payment via Tikkie:</strong></p>
					<p>You will soon receive an email with a Tikkie payment link. Once the payment is received, I will start preparing your order!</p>

					<p>Questions? Contact me at %s</p>
				</div>

				<div class="footer">
					<p>MamaBloemetjes | Fresh Flowers Delivered with Love</p>
				</div>
			</div>
		</body>
		</html>
	`, name, orderNumber, itemsListNL, totalFormatted, addressFormatted, es.cfg.Email.SupportEmail,
		name, orderNumber, itemsListEN, totalFormatted, addressFormatted, es.cfg.Email.SupportEmail)

	subject := fmt.Sprintf("Bevestiging van je bestelling %s / Order confirmation %s", orderNumber, orderNumber)

	return es.SendEmail([]string{email, es.cfg.Email.SupportEmail}, subject, emailBody)
}

// SendPaymentLinkEmail sends a bilingual email with the Tikkie payment link
func (es *EmailService) SendPaymentLinkEmail(email, name, orderNumber, paymentLink string) error {
	emailBody := fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<style>
				body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
				.container { max-width: 600px; margin: 0 auto; padding: 20px; }
				.header { background-color: #4CAF50; color: white; padding: 20px; text-align: center; }
				.content { padding: 20px; background-color: #f9f9f9; }
				.button { display: inline-block; padding: 15px 30px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px; margin: 20px 0; }
				.footer { text-align: center; padding: 20px; color: #666; font-size: 12px; }
				.divider { margin: 30px 0; border-top: 2px solid #ddd; }
			</style>
		</head>
		<body>
			<div class="container">
				<!-- Dutch Version -->
				<div class="header">
					<h1>Je betaallink is klaar!</h1>
				</div>
				<div class="content">
					<p>Beste %s,</p>
					<p>Je Tikkie betaallink voor bestelling <strong>%s</strong> is klaar!</p>

					<p style="text-align: center;">
						<a href="%s" class="button">Betaal via Tikkie</a>
					</p>

					<p>Of kopieer deze link naar je browser:</p>
					<p style="word-break: break-all;">%s</p>

					<p>Zodra de betaling is ontvangen, ga ik direct aan de slag met je bestelling!</p>

					<p>Vragen? Neem contact met ons op via %s</p>
				</div>

				<div class="divider"></div>

				<!-- English Version -->
				<div class="header">
					<h1>Your payment link is ready!</h1>
				</div>
				<div class="content">
					<p>Dear %s,</p>
					<p>Your Tikkie payment link for order <strong>%s</strong> is ready!</p>

					<p style="text-align: center;">
						<a href="%s" class="button">Pay via Tikkie</a>
					</p>

					<p>Or copy this link to your browser:</p>
					<p style="word-break: break-all;">%s</p>

					<p>Once the payment is recieved, I will immediately start preparing your order!</p>

					<p>Questions? Contact me at %s</p>
				</div>

				<div class="footer">
					<p>MamaBloemetjes | Fresh Flowers Delivered with Love</p>
				</div>
			</div>
		</body>
		</html>
	`, name, orderNumber, paymentLink, paymentLink, es.cfg.Email.SupportEmail,
		name, orderNumber, paymentLink, paymentLink, es.cfg.Email.SupportEmail)

	subject := fmt.Sprintf("Betaallink voor bestelling %s / Payment link for order %s", orderNumber, orderNumber)

	return es.SendEmail([]string{email}, subject, emailBody)
}
