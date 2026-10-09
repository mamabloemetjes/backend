package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/MonkyMars/gecho"
)

// getRateLimitForEndpoint determines which rate limit to apply based on config
func (mw *Middleware) getRateLimitForEndpoint(path, method string) (int, time.Duration) {

	// Order creation - very strict limit (1 per 30 minutes)
	if strings.HasPrefix(path, "/orders/create") && method == http.MethodPost {
		return 1, 30 * time.Minute
	}

	// Auth endpoints - strictest limits
	if strings.HasPrefix(path, "/auth/login") ||
		strings.HasPrefix(path, "/auth/register") ||
		strings.HasPrefix(path, "/auth/logout") ||
		strings.HasPrefix(path, "/auth/refresh") {
		return mw.cfg.RateLimit.AuthLimit, mw.cfg.RateLimit.AuthWindow
	}

	// Admin endpoints
	if strings.HasPrefix(path, "/admin") {
		return mw.cfg.RateLimit.AdminLimit, mw.cfg.RateLimit.AdminWindow
	}

	// Expensive read operations
	if method == http.MethodGet && (strings.Contains(path, "/products") ||
		strings.Contains(path, "/search")) {
		return mw.cfg.RateLimit.ExpensiveLimit, mw.cfg.RateLimit.ExpensiveWindow
	}

	// Default limit for everything else
	return mw.cfg.RateLimit.GeneralLimit, mw.cfg.RateLimit.GeneralWindow
}

// getClientIP extracts the client IP supplied by Cloudflare Tunnel.
func (mw *Middleware) getClientIP(r *http.Request) string {
	if ip := net.ParseIP(strings.TrimSpace(r.Header.Get("CF-Connecting-IP"))); ip != nil {
		return ip.String()
	}

	// When the request did not come through Cloudflare, use the actual peer
	// address rather than trusting client-controlled forwarding headers.
	if host, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		if ip := net.ParseIP(host); ip != nil {
			return ip.String()
		}
	}

	if ip := net.ParseIP(strings.TrimSpace(r.RemoteAddr)); ip != nil {
		return ip.String()
	}

	return strings.TrimSpace(r.RemoteAddr)
}

// normalizeRateLimitEndpoint groups equivalent routes into bounded key buckets.
func (mw *Middleware) normalizeRateLimitEndpoint(endpoint string) string {
	// Normalize endpoint to group similar requests
	// This prevents cache key explosion
	normalizedEndpoint := endpoint

	// Remove trailing slashes
	normalizedEndpoint = strings.TrimSuffix(normalizedEndpoint, "/")

	// Group dynamic routes by their base path
	// e.g., /products/123 -> /products/:id
	if strings.HasPrefix(normalizedEndpoint, "/products/") && !strings.HasSuffix(normalizedEndpoint, "/products") {
		parts := strings.Split(normalizedEndpoint, "/")
		if len(parts) > 2 {
			normalizedEndpoint = "/products/:id"
		}
	}

	return normalizedEndpoint
}

// RateLimitMiddleware implements sliding window rate limiting with minimal latency
func (mw *Middleware) RateLimitMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip if rate limiting is disabled
			if !mw.cfg.RateLimit.Enabled {
				next.ServeHTTP(w, r)
				return
			}

			// Skip rate limiting for health check
			if r.URL.Path == "/health" || r.URL.Path == "/" {
				next.ServeHTTP(w, r)
				return
			}

			// Extract client IP
			clientIP := mw.getClientIP(r)

			// Get rate limit for this endpoint
			limit, window := mw.getRateLimitForEndpoint(r.URL.Path, r.Method)

			endpoint := mw.normalizeRateLimitEndpoint(r.URL.Path)

			// Increment rate limit counter (synchronous call)
			count, err := mw.cacheService.IncrementRateLimit(clientIP, endpoint, window)
			if err != nil {
				if isCriticalRateLimitedPath(r.URL.Path) {
					mw.logger.Error("Critical rate limit unavailable", gecho.Field("error", err), gecho.Field("endpoint", endpoint))
					http.Error(w, "rate limiting temporarily unavailable", http.StatusServiceUnavailable)
					return
				}

				// Public reads fail open so a Redis outage does not take down the catalogue.
				mw.logger.Warn("Rate limit cache error, allowing request",
					gecho.Field("error", err),
					gecho.Field("ip", clientIP),
					gecho.Field("endpoint", endpoint),
				)
				next.ServeHTTP(w, r)
				return
			}

			// Check if limit exceeded
			if count > limit {
				mw.logger.Warn("Rate limit exceeded",
					gecho.Field("ip", clientIP),
					gecho.Field("endpoint", endpoint),
					gecho.Field("count", count),
					gecho.Field("limit", limit),
				)

				// Add rate limit headers
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
				w.Header().Set("Content-Type", "application/json")

				gecho.TooManyRequests(w,
					gecho.WithMessage("error.rateLimitExceeded"),
					gecho.WithData(map[string]any{"limit": limit, "window": window.String(), "retry_after": int(window.Seconds())}),
					gecho.Send(),
				)
				return
			}

			// Add rate limit headers (informational)
			remaining := max(0, limit-count)
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(window).Unix()))

			// Log if getting close to limit (80% threshold)
			if count > int(float64(limit)*0.8) {
				mw.logger.Debug("Rate limit warning",
					gecho.Field("ip", clientIP),
					gecho.Field("endpoint", endpoint),
					gecho.Field("count", count),
					gecho.Field("limit", limit),
					gecho.Field("remaining", remaining),
				)
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

func isCriticalRateLimitedPath(path string) bool {
	return strings.HasPrefix(path, "/auth/") ||
		strings.HasPrefix(path, "/orders/") ||
		strings.HasPrefix(path, "/admin/")
}

// StrictRateLimitMiddleware is a stricter version that fails closed on cache errors
// Use this for critical endpoints where you prefer to block on cache failure
func (mw *Middleware) StrictRateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := mw.getClientIP(r)
			endpoint := mw.normalizeRateLimitEndpoint(r.URL.Path)

			count, err := mw.cacheService.IncrementRateLimit(clientIP, endpoint, window)
			if err != nil {
				// Fail closed - block request on cache error
				mw.logger.Error("Rate limit cache error, blocking request",
					gecho.Field("error", err),
					gecho.Field("ip", clientIP),
					gecho.Field("endpoint", endpoint),
				)

				gecho.ServiceUnavailable(w,
					gecho.WithMessage("error.serviceUnavailable"),
					gecho.Send(),
				)
				return
			}

			if count > limit {
				mw.logger.Warn("Strict rate limit exceeded",
					gecho.Field("ip", clientIP),
					gecho.Field("endpoint", endpoint),
					gecho.Field("count", count),
					gecho.Field("limit", limit),
				)

				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(window.Seconds())))
				w.Header().Set("Content-Type", "application/json")

				gecho.TooManyRequests(w, gecho.WithMessage("error.rateLimitExceeded.strict"), gecho.Send())
				return
			}

			// Add headers
			remaining := limit - count
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

			next.ServeHTTP(w, r)
		})
	}
}

// IPWhitelistMiddleware allows bypassing rate limits for whitelisted IPs
func (mw *Middleware) IPWhitelistMiddleware(whitelistedIPs []string) func(http.Handler) http.Handler {
	// Convert to map for O(1) lookup
	whitelist := make(map[string]bool)
	for _, ip := range whitelistedIPs {
		whitelist[ip] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := mw.getClientIP(r)

			if whitelist[clientIP] {
				mw.logger.Debug("Bypassing rate limit for whitelisted IP",
					gecho.Field("ip", clientIP),
				)
				next.ServeHTTP(w, r)
				return
			}

			// Not whitelisted, continue with rate limiting
			mw.RateLimitMiddleware()(next).ServeHTTP(w, r)
		})
	}
}
