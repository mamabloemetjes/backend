package structs

import "time"

type Config struct {
	Server     *ServerConfig     `validate:"required"`
	Cors       *CorsConfig       `validate:"required"`
	Database   *DatabaseConfig   `validate:"required"`
	Auth       *AuthConfig       `validate:"required"`
	Cache      *CacheConfig      `validate:"required"`
	RateLimit  *RateLimitConfig  `validate:"required"`
	Email      *EmailConfig      `validate:"required"`
	Encryption *EncryptionConfig `validate:"required"`
}

type ServerConfig struct {
	AppName           string        `validate:"required,min=2,max=100"`                // Mamabloemetjes
	Environment       string        `validate:"required,oneof=development production"` // development, production
	Port              string        `validate:"required,min=4,max=10"`                 // :8082
	LogLevel          string        `validate:"required,oneof=debug info warn error"`  // debug, info, warn, error
	ServerURL         string        `validate:"required,url"`                          // Base URL of the server
	FrontendURL       string        `validate:"required,url"`                          // Base URL of the frontend
	ReadTimeout       time.Duration `validate:"required,min=1s"`                       // in seconds
	WriteTimeout      time.Duration `validate:"required,min=1s"`                       // in seconds
	IdleTimeout       time.Duration `validate:"required,min=1s"`                       // in seconds
	ReadHeaderTimeout time.Duration `validate:"required,min=1s"`                       // in seconds
	MaxHeaderBytes    int           `validate:"required,min=1024"`                     // in bytes
}

type CorsConfig struct {
	AllowedOrigins   []string `validate:"required,min=1,dive,required"`
	AllowedMethods   []string `validate:"required,min=1,dive,required"`
	AllowedHeaders   []string `validate:"required,min=1,dive,required"`
	ExposedHeaders   []string `validate:"omitempty,dive,required"`
	AllowCredentials bool
	MaxAge           int `validate:"required,min=0"` // in seconds
}

type DatabaseConfig struct {
	Host         string        `validate:"required,min=1,max=255"`
	Port         int           `validate:"required,min=1,max=65535"`
	User         string        `validate:"required,min=1,max=100"`
	Password     string        `validate:"required,min=1"`
	Name         string        `validate:"required,min=1,max=100"`
	MaxConns     int           `validate:"required,min=1"`
	MinConns     int           `validate:"required,min=0"`
	MaxLifetime  time.Duration `validate:"required,min=1s"` // in seconds
	MaxIdleTime  time.Duration `validate:"required,min=1s"` // in seconds
	ReadTimeout  time.Duration `validate:"required,min=1s"` // in seconds
	WriteTimeout time.Duration `validate:"required,min=1s"` // in seconds
}

type AuthConfig struct {
	AccessTokenSecret  string        `validate:"required,min=32"`
	AccessTokenExpiry  time.Duration `validate:"required,min=1m"`
	RefreshTokenSecret string        `validate:"required,min=32"`
	RefreshTokenExpiry time.Duration `validate:"required,min=1m"`
	CacheUserTTL       time.Duration `validate:"required,min=1s"`
	BlacklistCacheTTL  time.Duration `validate:"required,min=1s"`
}

type CacheConfig struct {
	Address         string        `validate:"required,min=1,max=255"`
	Username        string        `validate:"omitempty,max=100"`
	Password        string        `validate:"omitempty"`
	DB              int           `validate:"min=0,max=15"`
	PoolSize        int           `validate:"required,min=1"`
	MinIdleConns    int           `validate:"required,min=0"`
	MaxIdleConns    int           `validate:"required,min=0"`
	PoolTimeout     time.Duration `validate:"required,min=1s"`
	IdleTimeout     time.Duration `validate:"required,min=1s"`
	DialTimeout     time.Duration `validate:"required,min=1s"`
	ReadTimeout     time.Duration `validate:"required,min=1s"`
	WriteTimeout    time.Duration `validate:"required,min=1s"`
	MaxRetries      int           `validate:"required,min=0"`
	MinRetryBackoff time.Duration `validate:"required,min=1ms"`
	MaxRetryBackoff time.Duration `validate:"required,min=1ms"`
	ProductListTTL  time.Duration `validate:"required,min=1s"`
	ProductCountTTL time.Duration `validate:"required,min=1s"`
}

type RateLimitConfig struct {
	// General API endpoints
	GeneralLimit  int           `validate:"required,min=1"`
	GeneralWindow time.Duration `validate:"required,min=1s"`

	// Auth endpoints (login, register) - stricter limits
	AuthLimit  int           `validate:"required,min=1"`
	AuthWindow time.Duration `validate:"required,min=1s"`

	// Expensive endpoints (search, list) - moderate limits
	ExpensiveLimit  int           `validate:"required,min=1"`
	ExpensiveWindow time.Duration `validate:"required,min=1s"`

	// Admin endpoints - moderate limits
	AdminLimit  int           `validate:"required,min=1"`
	AdminWindow time.Duration `validate:"required,min=1s"`

	// Enable/disable rate limiting
	Enabled bool
}

type EmailConfig struct {
	ApiKey                  string        `validate:"required,min=10"`
	From                    string        `validate:"required,email"`
	VerificationTokenExpiry time.Duration `validate:"required,min=1m"`
	OrderConfirmationFrom   string        `validate:"required,email"` // Email address for order confirmations
	SupportEmail            string        `validate:"required,email"` // Support email to show in order emails
}

type EncryptionConfig struct {
	Key string `validate:"required,len=32"` // AES-256 encryption key (32 bytes)
}
