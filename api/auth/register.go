package auth

import (
	"net/http"

	"github.com/MonkyMars/gecho"

	"github.com/mamabloemetjes/backend/api/response"
	"github.com/mamabloemetjes/backend/lib"
	"github.com/mamabloemetjes/backend/structs"
)

func (ar *AuthRoutesManager) HandleRegister(w http.ResponseWriter, r *http.Request) {
	body, err := lib.ExtractAndValidateBody[structs.RegisterRequest](r)
	if err != nil {
		ar.logger.Warn("Failed to extract and validate request body", gecho.Field("error", err))
		response.Send(gecho.BadRequest(w, gecho.WithMessage("error.auth.checkRegistrationInformation")))
		return
	}

	user, err := ar.authService.Register(body)
	if err != nil {
		// Get user-friendly message from error
		userMessage := lib.GetUserMessage(err)

		// Unique violations return 409 Conflict (already logged as warn in service)
		if lib.IsUniqueViolation(err) {
			response.Send(gecho.Conflict(w, gecho.WithMessage(userMessage)))
			return
		}
		// Other database errors return 500 (already logged as error in service)
		response.Send(gecho.InternalServerError(w, gecho.WithMessage(userMessage)))
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
	response.Send(gecho.Success(w,
		gecho.WithMessage("success.auth.userRegistered"),
	))
}
