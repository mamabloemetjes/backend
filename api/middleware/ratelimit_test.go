package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetClientIPUsesCloudflareConnectingIP(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/products", nil)
	request.RemoteAddr = "10.0.0.2:1234"
	request.Header.Set("CF-Connecting-IP", "203.0.113.10")

	if got, want := (&Middleware{}).getClientIP(request), "203.0.113.10"; got != want {
		t.Fatalf("client IP = %q, want %q", got, want)
	}
}

func TestGetClientIPDoesNotTrustForwardedHeaders(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/products", nil)
	request.RemoteAddr = "10.0.0.2:1234"
	request.Header.Set("X-Forwarded-For", "203.0.113.10")
	request.Header.Set("X-Real-IP", "203.0.113.11")

	if got, want := (&Middleware{}).getClientIP(request), "10.0.0.2"; got != want {
		t.Fatalf("client IP = %q, want %q", got, want)
	}
}

func TestGetClientIPFallsBackToIPv6Peer(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/products", nil)
	request.RemoteAddr = "[2001:db8::1]:1234"

	if got, want := (&Middleware{}).getClientIP(request), "2001:db8::1"; got != want {
		t.Fatalf("client IP = %q, want %q", got, want)
	}
}

func TestIsCriticalRateLimitedPath(t *testing.T) {
	for _, path := range []string{"/auth/login", "/orders/create", "/admin/products"} {
		if !isCriticalRateLimitedPath(path) {
			t.Errorf("expected %s to be critical", path)
		}
	}
	if isCriticalRateLimitedPath("/products") {
		t.Error("public product reads should fail open")
	}
}
