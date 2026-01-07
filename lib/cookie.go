package lib

import (
	"mamabloemetjes_server/config"
	"net/http"
	"time"
)

// SetCookie sets a secure, HttpOnly cookie for authentication/session usage
func SetCookie(key, val string, expiry time.Time, w http.ResponseWriter) {
	isProduction := config.IsProduction()

	sameSite := http.SameSiteLaxMode
	secure := false
	domain := ""

	if isProduction {
		// Required for cross-subdomain cookies (www <-> api)
		sameSite = http.SameSiteNoneMode
		secure = true
		domain = ".roosvansharon.nl"
	}

	cookie := &http.Cookie{
		Name:     key,
		Value:    val,
		Expires:  expiry,
		Path:     "/",
		Domain:   domain,
		Secure:   secure,
		SameSite: sameSite,
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
}

func GetCookieValue(key string, r *http.Request) (string, error) {
	cookie, err := r.Cookie(key)
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

// ClearCookie removes the cookie from the browser
func ClearCookie(key string, w http.ResponseWriter) {
	isProduction := config.IsProduction()

	sameSite := http.SameSiteLaxMode
	secure := false
	domain := ""

	if isProduction {
		sameSite = http.SameSiteNoneMode
		secure = true
		domain = ".roosvansharon.nl"
	}

	cookie := &http.Cookie{
		Name:     key,
		Value:    "",
		Path:     "/",
		Domain:   domain,
		Expires:  time.Now().Add(-time.Hour),
		MaxAge:   -1,
		Secure:   secure,
		SameSite: sameSite,
		HttpOnly: true,
	}

	http.SetCookie(w, cookie)
}

// SetCSRFCookie sets a CSRF token cookie that must be readable by JavaScript
func SetCSRFCookie(val string, expiry time.Time, w http.ResponseWriter) {
	isProduction := config.IsProduction()

	sameSite := http.SameSiteLaxMode
	secure := false
	domain := ""

	if isProduction {
		// CSRF cookie must be sent cross-subdomain
		sameSite = http.SameSiteNoneMode
		secure = true
		domain = ".roosvansharon.nl"
	}

	cookie := &http.Cookie{
		Name:     CSRFCookieName,
		Value:    val,
		Expires:  expiry,
		MaxAge:   int(time.Until(expiry).Seconds()),
		Path:     "/",
		Domain:   domain,
		Secure:   secure,
		SameSite: sameSite,
		HttpOnly: false, // Must be readable by JS
	}

	http.SetCookie(w, cookie)
}
