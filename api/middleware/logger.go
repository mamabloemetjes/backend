package middleware

import (
	"net/http"

	"github.com/MonkyMars/gecho"
)

func (mw *Middleware) SetupLoggerMiddleware() func(http.Handler) http.Handler {
	return gecho.Handlers.CreateLoggingMiddleware(mw.logger)
}
