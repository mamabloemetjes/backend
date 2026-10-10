package auth

import (
	"encoding/json"
	"mamabloemetjes_server/lib"
	"net/http"
	"time"
)

// HandleCSRF generates and sets a CSRF token
func (ar *AuthRoutesManager) HandleCSRF(w http.ResponseWriter, r *http.Request) {
	// Generate a new CSRF token
	token, err := lib.GenerateRandomToken()
	if err != nil {
		ar.logger.Error("Failed to generate CSRF token", "error", err)
		http.Error(w, "unable to generate csrf token", http.StatusInternalServerError)
		return
	}

	// Set CSRF cookie with 24 hour expiration
	expiry := time.Now().Add(24 * time.Hour)
	lib.SetCSRFCookie(token, expiry, w)

	// Return the token in the response as well
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"data": map[string]string{"csrf_token": token},
	})
}
