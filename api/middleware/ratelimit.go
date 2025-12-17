package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/MonkyMars/gecho"
)

// getRateLimitForEndpoint determines which rate limit to apply based on config
func (mw *Middleware) getRateLimitForEndpoint(path, method string) (int, time.Duration) {

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

// getClientIP extracts the real client IP from request headers
func (mw *Middleware) getClientIP(r *http.Request) string {
	// Try X-Forwarded-For first (if behind proxy/load balancer)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Try X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	ip := r.RemoteAddr
	// Remove port if present
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}

// generateRateLimitKey creates a unique cache key for rate limiting
func (mw *Middleware) generateRateLimitKey(ip, endpoint string) string {
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

	return fmt.Sprintf("%s:%s", ip, normalizedEndpoint)
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

			// Use endpoint path directly
			endpoint := r.URL.Path

			// Increment rate limit counter (synchronous call)
			count, err := mw.cacheService.IncrementRateLimit(clientIP, endpoint, window)
			if err != nil {
				// Cache error - log and allow request (fail open)
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

				http.Error(w, fmt.Sprintf(`{"message":"Rate limit exceeded. Please try again later.","data":{"limit":%d,"window":"%s","retry_after":%d}}`,
					limit, window.String(), int(window.Seconds())), http.StatusTooManyRequests)
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

// StrictRateLimitMiddleware is a stricter version that fails closed on cache errors
// Use this for critical endpoints where you prefer to block on cache failure
func (mw *Middleware) StrictRateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := mw.getClientIP(r)
			endpoint := r.URL.Path

			count, err := mw.cacheService.IncrementRateLimit(clientIP, endpoint, window)
			if err != nil {
				// Fail closed - block request on cache error
				mw.logger.Error("Rate limit cache error, blocking request",
					gecho.Field("error", err),
					gecho.Field("ip", clientIP),
					gecho.Field("endpoint", endpoint),
				)

				gecho.ServiceUnavailable(w,
					gecho.WithMessage("Service temporarily unavailable"),
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

				http.Error(w, `{"message":"Rate limit exceeded"}`, http.StatusTooManyRequests)
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
