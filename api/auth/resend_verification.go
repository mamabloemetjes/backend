package auth

import (
	"context"
	"net/http"
	"time"

	"mamabloemetjes_server/database"
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs/tables"

	"github.com/MonkyMars/gecho"
)

type ResendVerificationRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// HandleResendVerification handles requests to resend verification emails
func (ar *AuthRoutesManager) HandleResendVerification(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[ResendVerificationRequest](r)
	if err != nil {
		ar.logger.Warn("Failed to extract and validate request body", gecho.Field("error", err))
		gecho.BadRequest(w, gecho.WithMessage("error.invalidRequest"), gecho.WithData(err), gecho.Send())
		return
	}

	// Find the user by email
	user, err := database.Query[tables.User](ar.authService.GetDB()).
		Where("email", body.Email).
		First(context.Background())
	if err != nil {
		ar.logger.Error("Failed to find user", gecho.Field("error", err), gecho.Field("email", body.Email))
		// Don't reveal if user exists or not for security reasons
		gecho.Success(w, gecho.WithMessage("success.auth.verificationEmailSent"), gecho.Send())
		return
	}

	// If user not found, still return success to prevent email enumeration
	if user == nil {
		ar.logger.Warn("User not found", gecho.Field("email", body.Email))
		gecho.Success(w, gecho.WithMessage("success.auth.verificationEmailSent"), gecho.Send())
		return
	}

	// Check if email is already verified
	if user.EmailVerified {
		ar.logger.Info("Email already verified", gecho.Field("user_id", user.Id))
		gecho.Success(w, gecho.WithMessage("success.auth.emailAlreadyVerified"), gecho.Send())
		return
	}

	// Check for rate limiting - prevent spam (max 1 email per 2 minutes)
	recentVerification, err := database.Query[tables.EmailVerification](ar.authService.GetDB()).
		Where("user_id", user.Id).
		OrderBy("created_at", "DESC").
		First(context.Background())

	if err == nil && recentVerification != nil {
		timeSinceLastEmail := time.Since(recentVerification.CreatedAt)
		if timeSinceLastEmail < 2*time.Minute {
			ar.logger.Warn("Rate limit exceeded for verification email",
				gecho.Field("user_id", user.Id),
				gecho.Field("time_since_last", timeSinceLastEmail))
			gecho.TooManyRequests(w,
				gecho.WithMessage("error.rateLimitExceeded"),
				gecho.WithData(map[string]interface{}{
					"retry_after_seconds": int((2*time.Minute - timeSinceLastEmail).Seconds()),
				}),
				gecho.Send())
			return
		}
	}

	// Delete any existing verification tokens for this user
	_, err = database.Query[tables.EmailVerification](ar.authService.GetDB()).
		Where("user_id", user.Id).
		Delete(context.Background())
	if err != nil {
		ar.logger.Warn("Failed to delete old verification tokens", gecho.Field("error", err), gecho.Field("user_id", user.Id))
		// Continue anyway - this is not critical
	}

	// Send new verification email
	_, err = ar.emailService.SendVerificationEmail(user)
	if err != nil {
		ar.logger.Error("Failed to send verification email", gecho.Field("error", err), gecho.Field("user_id", user.Id))
		gecho.InternalServerError(w, gecho.WithMessage("error.failedToSendEmail"), gecho.Send())
		return
	}

	ar.logger.Info("Verification email resent successfully", gecho.Field("user_id", user.Id))
	gecho.Success(w, gecho.WithMessage("success.auth.verificationEmailSent"), gecho.Send())
}
