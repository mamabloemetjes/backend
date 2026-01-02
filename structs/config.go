package structs

import "time"

type Config struct {
	Server    *ServerConfig
	Cors      *CorsConfig
	Database  *DatabaseConfig
	Auth      *AuthConfig
	Cache     *CacheConfig
	RateLimit *RateLimitConfig
	Email     *EmailConfig
}

type ServerConfig struct {
	AppName        string        // Mamabloemetjes
	Environment    string        // development, production
	Port           string        // :8082
	LogLevel       string        // debug, info, warn, error
	ServerURL      string        // Base URL of the server
	FrontendURL    string        // Base URL of the frontend
	ReadTimeout    time.Duration // in seconds
	WriteTimeout   time.Duration // in seconds
	IdleTimeout    time.Duration // in seconds
	MaxHeaderBytes int           // in bytes
}

type CorsConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int // in seconds
}

type DatabaseConfig struct {
	Host         string
	Port         int
	User         string
	Password     string
	Name         string
	MaxConns     int
	MinConns     int
	MaxLifetime  time.Duration // in seconds
	MaxIdleTime  time.Duration // in seconds
	ReadTimeout  time.Duration // in seconds
	WriteTimeout time.Duration // in seconds
}

type AuthConfig struct {
	AccessTokenSecret  string
	AccessTokenExpiry  time.Duration
	RefreshTokenSecret string
	RefreshTokenExpiry time.Duration
	CacheUserTTL       time.Duration
	BlacklistCacheTTL  time.Duration
}

type CacheConfig struct {
	Address         string
	Username        string
	Password        string
	DB              int
	PoolSize        int
	MinIdleConns    int
	MaxIdleConns    int
	PoolTimeout     time.Duration
	IdleTimeout     time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration
	ProductListTTL  time.Duration
	ProductCountTTL time.Duration
}

type RateLimitConfig struct {
	// General API endpoints
	GeneralLimit  int
	GeneralWindow time.Duration

	// Auth endpoints (login, register) - stricter limits
	AuthLimit  int
	AuthWindow time.Duration

	// Expensive endpoints (search, list) - moderate limits
	ExpensiveLimit  int
	ExpensiveWindow time.Duration

	// Admin endpoints - moderate limits
	AdminLimit  int
	AdminWindow time.Duration

	// Enable/disable rate limiting
	Enabled bool
}

type EmailConfig struct {
	ApiKey                  string
	From                    string
	VerificationTokenExpiry time.Duration
}
