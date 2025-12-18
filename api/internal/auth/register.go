package auth

import (
	"errors"
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (ar *AuthRoutesManager) HandleRegister(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[structs.RegisterRequest](r)
	if err != nil {
		ar.logger.Warn("Failed to extract request body", gecho.Field("error", err))
		gecho.BadRequest(w, gecho.WithMessage("Please check your registration information and try again"), gecho.Send())
		return
	}

	if body.Email == "" || body.Password == "" || body.Username == "" {
		ar.logger.Warn("Missing required fields in registration", gecho.Field("body", body))
		gecho.BadRequest(w, gecho.WithMessage("Email, username, and password are required"), gecho.Send())
		return
	}

	user, err := ar.authService.Register(body)
	if err != nil {
		if errors.Is(err, lib.ErrConflict) {
			ar.logger.Warn("User already exists", gecho.Field("email", body.Email), gecho.Field("username", body.Username))
			gecho.Conflict(w, gecho.WithMessage("This email or username is already registered"), gecho.Send())
			return
		}
		ar.logger.Error("Failed to create user", gecho.Field("error", err))
		gecho.InternalServerError(w, gecho.WithMessage("Unable to create account. Please try again"), gecho.Send())
		return
	}

	// clear password from user
	user.PasswordHash = ""

	go func() {
		// Send verification email
		result, err := ar.emailService.SendVerificationEmail(user)
		if err != nil {
			ar.logger.Error("Failed to send verification email", gecho.Field("error", err), gecho.Field("user_id", user.Id))
			return
		}
		ar.logger.Debug("Verification email sent", gecho.Field("email_verification_id", result.Id), gecho.Field("user_id", user.Id))
	}()

	gecho.Success(w,
		gecho.WithMessage("User registered successfully"),
		gecho.Send(),
	)
}
