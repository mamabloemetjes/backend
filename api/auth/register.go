package auth

import (
	"mamabloemetjes_server/lib"
	"mamabloemetjes_server/structs"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (ar *AuthRoutesManager) HandleRegister(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[structs.RegisterRequest](r)
	if err != nil {
		ar.logger.Warn("Failed to extract and validate request body", gecho.Field("error", err))
		gecho.BadRequest(w, gecho.WithMessage("error.auth.checkRegistrationInformation"), gecho.WithData(err), gecho.Send())
		return
	}

	user, err := ar.authService.Register(body)
	if err != nil {
		// Get user-friendly message from error
		userMessage := lib.GetUserMessage(err)

		// Unique violations return 409 Conflict (already logged as warn in service)
		if lib.IsUniqueViolation(err) {
			gecho.Conflict(w, gecho.WithMessage(userMessage), gecho.Send())
			return
		}

		// Other database errors return 500 (already logged as error in service)
		gecho.InternalServerError(w, gecho.WithMessage(userMessage), gecho.Send())
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
		gecho.WithMessage("success.auth.userRegistered"),
		gecho.Send(),
	)
}
