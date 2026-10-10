package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (mw *Middleware) SecurityHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			w.Header().Set("Permissions-Policy", "geolocation=(), camera=()")

			next.ServeHTTP(w, r)
		})
	}
}

func (mw *Middleware) BodyLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

func (mw *Middleware) CSRFMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip CSRF check for GET, HEAD, and OPTIONS requests
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			allCookies := r.Cookies()
			cookieNames := make([]string, len(allCookies))
			for i, c := range allCookies {
				cookieNames[i] = c.Name
			}
			cookie, err := r.Cookie("csrf")
			if err != nil {
				mw.logger.Warn("CSRF cookie missing", gecho.Field("path", r.URL.Path), gecho.Field("cookies", cookieNames))
				gecho.Forbidden(w, gecho.WithMessage("invalid csrf token"), gecho.Send())
				return
			}

			token := r.Header.Get("X-CSRF-Token")
			if token == "" {
				mw.logger.Warn("CSRF header missing", gecho.Field("path", r.URL.Path))
				http.Error(w, "invalid csrf token", http.StatusForbidden)
				return
			}

			if len(token) != len(cookie.Value) ||
				subtle.ConstantTimeCompare([]byte(token), []byte(cookie.Value)) != 1 {
				mw.logger.Warn("CSRF token mismatch",
					gecho.Field("path", r.URL.Path),
					gecho.Field("header_len", len(token)),
					gecho.Field("cookie_len", len(cookie.Value)),
				)
				http.Error(w, "invalid csrf token", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
