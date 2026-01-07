package lib

import (
	"mamabloemetjes_server/config"
	"net/http"
	"time"
)

func SetCookie(key, val string, expiry time.Time, w http.ResponseWriter) {
	isProduction := config.IsProduction()
	sameSite := http.SameSiteLaxMode
	if isProduction {
		sameSite = http.SameSiteStrictMode
	}
	cookie := &http.Cookie{
		Name:     key,
		Value:    val,
		Expires:  expiry,
		Secure:   isProduction,
		Path:     "/",
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

func ClearCookie(key string, w http.ResponseWriter) {
	isProduction := config.IsProduction()
	sameSite := http.SameSiteLaxMode
	if isProduction {
		sameSite = http.SameSiteStrictMode
	}
	cookie := &http.Cookie{
		Name:     key,
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		MaxAge:   -1,
		Secure:   isProduction,
		Path:     "/",
		SameSite: sameSite,
		HttpOnly: true,
	}
	http.SetCookie(w, cookie)
}

// SetCSRFCookie sets a CSRF token cookie that is readable by JavaScript
func SetCSRFCookie(val string, expiry time.Time, w http.ResponseWriter) {
	isProduction := config.IsProduction()
	// Use SameSiteNoneMode for cross-origin requests with Secure flag
	// In development, use SameSiteLaxMode without Secure to allow localhost
	sameSite := http.SameSiteLaxMode
	if isProduction {
		sameSite = http.SameSiteNoneMode
	}

	cookie := &http.Cookie{
		Name:     CSRFCookieName,
		Value:    val,
		Expires:  expiry,
		MaxAge:   int(time.Until(expiry).Seconds()),
		Secure:   isProduction,
		Path:     "/",
		SameSite: sameSite,
		HttpOnly: false, // Must be readable by JavaScript
	}
	http.SetCookie(w, cookie)
}
