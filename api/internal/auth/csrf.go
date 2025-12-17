package auth

import (
	"mamabloemetjes_server/lib"
	"net/http"
	"time"

	"github.com/MonkyMars/gecho"
)

// HandleCSRF generates and sets a CSRF token
func (ar *AuthRoutesManager) HandleCSRF(w http.ResponseWriter, r *http.Request) {
	// Generate a new CSRF token
	token, err := lib.GenerateCSRFToken()
	if err != nil {
		ar.logger.Error("Failed to generate CSRF token", gecho.Field("error", err))
		gecho.InternalServerError(w,
			gecho.WithMessage("Failed to generate CSRF token"),
			gecho.Send(),
		)
		return
	}

	// Set CSRF cookie with 24 hour expiration
	expiry := time.Now().Add(24 * time.Hour)
	lib.SetCSRFCookie(token, expiry, w)

	ar.logger.Info("CSRF token generated and cookie set",
		gecho.Field("token_length", len(token)),
		gecho.Field("token_preview", token[:min(10, len(token))]),
		gecho.Field("expiry", expiry),
	)

	// Return the token in the response as well
	gecho.Success(w,
		gecho.WithMessage("CSRF token generated"),
		gecho.WithData(map[string]string{
			"csrf_token": token,
		}),
		gecho.Send(),
	)
}
