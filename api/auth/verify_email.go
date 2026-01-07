package auth

import (
	"fmt"
	"net/http"

	"github.com/MonkyMars/gecho"
	"github.com/google/uuid"
)

// HandleVerifyEmail handles email verification requests and redirects to the frontend.
func (ar *AuthRoutesManager) HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	// Get the token from the query parameters and user id
	params := r.URL.Query()
	token := params.Get("token")
	userID := params.Get("user_id")

	if token == "" || userID == "" {
		gecho.BadRequest(w, gecho.WithMessage("error.auth.missingTokenOrUserId"), gecho.Send())
		return
	}

	// Parse string to uuid
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		ar.logger.Warn("Invalid user_id format", gecho.Field("error", err), gecho.Field("user_id", userID))
		gecho.BadRequest(w, gecho.WithMessage("error.auth.invalidUserIdFormat"), gecho.Send())
		return
	}

	// Verify the email
	err = ar.authService.VerifyEmail(userUUID, token)
	if err != nil {
		ar.logger.Warn("Email verification failed", gecho.Field("error", err), gecho.Field("user_id", userID))
		// Redirect to frontend with failure
		http.Redirect(w, r, getRedirectURL(ar.cfg.Server.FrontendURL, "err"), http.StatusSeeOther)
		return
	}

	ar.logger.Info("Email verified successfully", gecho.Field("user_id", userID))

	// Redirect to frontend with success (user needs to log in manually)
	http.Redirect(w, r, getRedirectURL(ar.cfg.Server.FrontendURL, "ok"), http.StatusSeeOther)
}

func getRedirectURL(cfgURL, status string) string {
	url := fmt.Sprintf("%s/email-verified?status=%s", cfgURL, status)
	return url
}
