package middleware

import (
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

			// Log all cookies received for debugging
			allCookies := r.Cookies()
			cookieNames := make([]string, len(allCookies))
			for i, c := range allCookies {
				cookieNames[i] = c.Name
			}
			mw.logger.Debug("CSRF check - cookies received", gecho.Field("path", r.URL.Path), gecho.Field("cookies", cookieNames))

			cookie, err := r.Cookie("csrf")
			if err != nil {
				mw.logger.Warn("CSRF cookie missing", gecho.Field("error", err), gecho.Field("path", r.URL.Path), gecho.Field("all_cookies", cookieNames))
				http.Error(w, "csrf missing", http.StatusForbidden)
				return
			}

			token := r.Header.Get("X-CSRF-Token")
			if token == "" {
				mw.logger.Warn("CSRF header missing", gecho.Field("path", r.URL.Path), gecho.Field("cookie_value", cookie.Value[:min(10, len(cookie.Value))]))
				http.Error(w, "invalid csrf token", http.StatusForbidden)
				return
			}

			if token != cookie.Value {
				mw.logger.Warn("CSRF token mismatch",
					gecho.Field("path", r.URL.Path),
					gecho.Field("header_full", token),
					gecho.Field("cookie_full", cookie.Value),
					gecho.Field("header_len", len(token)),
					gecho.Field("cookie_len", len(cookie.Value)),
					gecho.Field("match", token == cookie.Value),
				)
				mw.logger.Warn("CSRF mismatch details",
					gecho.Field("header_bytes", []byte(token)),
					gecho.Field("cookie_bytes", []byte(cookie.Value)),
				)
				http.Error(w, "invalid csrf token", http.StatusForbidden)
				return
			}

			mw.logger.Info("CSRF token valid", gecho.Field("path", r.URL.Path))

			next.ServeHTTP(w, r)
		})
	}
}
