package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MonkyMars/gecho"
)

func TestCSRFMiddlewareRequiresMatchingCookieAndHeader(t *testing.T) {
	logger := gecho.NewLogger(gecho.NewConfig(gecho.WithLogLevel(gecho.LogLevelError)))
	mw := (&Middleware{logger: logger}).CSRFMiddleware()
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	tests := []struct {
		name       string
		cookie     string
		header     string
		wantStatus int
	}{
		{name: "missing token", wantStatus: http.StatusForbidden},
		{name: "mismatch", cookie: "cookie-token", header: "header-token", wantStatus: http.StatusForbidden},
		{name: "valid", cookie: "same-token", header: "same-token", wantStatus: http.StatusNoContent},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/orders/create", nil)
			if test.cookie != "" {
				request.AddCookie(&http.Cookie{Name: "csrf", Value: test.cookie})
			}
			if test.header != "" {
				request.Header.Set("X-CSRF-Token", test.header)
			}
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
		})
	}
}
