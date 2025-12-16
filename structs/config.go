package structs

import "time"

type Config struct {
	Server   *ServerConfig
	Cors     *CorsConfig
	Database *DatabaseConfig
	Auth     *AuthConfig
}

type ServerConfig struct {
	AppName        string        // Mamabloemetjes
	Environment    string        // development, production
	Port           string        // :8082
	ReadTimeout    time.Duration // in seconds
	WriteTimeout   time.Duration // in seconds
	IdleTimeout    time.Duration // in seconds
	MaxHeaderBytes int           // in bytes
}

type CorsConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposedHeaders   []string
	AllowCredentials bool
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
}
