package lib

import (
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"mamabloemetjes_server/config"
)

var (
	cookieDomainOnce sync.Once
	cookieDomain     string
)

// resolveCookieDomain parses config.Server.ServerURL once and caches the
// bare hostname to use as the cookie Domain attribute. Fails fast if the
// value is missing or malformed, rather than silently issuing cookies
// with an empty/wrong domain.
func resolveCookieDomain() string {
	cookieDomainOnce.Do(func() {
		cfg := config.GetConfig()

		u, err := url.Parse(cfg.Server.ServerURL)
		if err != nil || u.Hostname() == "" {
			log.Fatalf("lib: invalid config.Server.ServerURL for cookie domain: %q (err=%v)", cfg.Server.ServerURL, err)
		}

		cookieDomain = u.Hostname()
	})

	return cookieDomain
}

// cookieSecuritySettings returns the SameSite/Secure/Domain values to use
// based on environment. In production, cookies are set with SameSite=None
// and Secure=true, scoped to the API's own host (host-only cookie) since
// the frontend and API are on different registrable domains.
func cookieSecuritySettings() (sameSite http.SameSite, secure bool, domain string) {
	if !config.IsProduction() {
		return http.SameSiteLaxMode, false, ""
	}

	return http.SameSiteNoneMode, true, resolveCookieDomain()
}

// SetCookie sets a secure, HttpOnly cookie for authentication/session usage
func SetCookie(key, val string, expiry time.Time, w http.ResponseWriter) {
	sameSite, secure, domain := cookieSecuritySettings()

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
	sameSite, secure, domain := cookieSecuritySettings()

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
	sameSite, secure, domain := cookieSecuritySettings()

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
