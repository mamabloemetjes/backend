package config

import (
	"mamabloemetjes_server/structs"
	"sync"
	"time"
)

var (
	configInstance *structs.Config
	configOnce     sync.Once
)

func GetConfig() *structs.Config {
	configOnce.Do(func() {
		configInstance = &structs.Config{
			Server: &structs.ServerConfig{
				AppName:        getEnvAsString("APP_NAME", "Mamabloemetjes_no_env"),
				Environment:    getEnvAsString("APP_ENV", "development"),
				Port:           getEnvAsString("APP_PORT", ":8082"),
				LogLevel:       getEnvAsString("APP_LOG_LEVEL", "info"),
				ServerURL:      getEnvAsString("APP_SERVER_URL", "http://localhost:8082"),
				FrontendURL:    getEnvAsString("APP_FRONTEND_URL", "http://localhost:3000"),
				ReadTimeout:    getEnvAsTimeDuration("SERVER_READ_TIME_OUT", 15*time.Second),
				WriteTimeout:   getEnvAsTimeDuration("SERVER_WRITE_TIME_OUT", 15*time.Second),
				IdleTimeout:    getEnvAsTimeDuration("SERVER_IDLE_TIME_OUT", 60*time.Second),
				MaxHeaderBytes: getEnvAsInt("SERVER_MAX_HEADER_BYTES", 1<<20), // 1 MB
			},
			Cors: &structs.CorsConfig{
				AllowedOrigins:   getEnvAsSlice("CORS_ALLOW_ORIGINS", []string{"localhost", "http://localhost:3000"}),
				AllowedMethods:   getEnvAsSlice("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
				AllowedHeaders:   getEnvAsSlice("CORS_ALLOW_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token"}),
				AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", false),
				ExposedHeaders:   getEnvAsSlice("CORS_EXPOSED_HEADERS", []string{"Content-Length", "Authorization"}),
				MaxAge:           getEnvAsInt("CORS_MAX_AGE", 600),
			},
			Database: &structs.DatabaseConfig{
				Host:         getEnvAsString("DB_HOST", "localhost"),
				Port:         getEnvAsInt("DB_PORT", 5432),
				User:         getEnvAsString("DB_USER", "postgres"),
				Password:     getEnvAsString("DB_PASSWORD", "password"),
				Name:         getEnvAsString("DB_NAME", "mamabloemetjes_db"),
				MaxConns:     getEnvAsInt("DB_MAX_CONNS", 10),
				MinConns:     getEnvAsInt("DB_MIN_CONNS", 2),
				MaxLifetime:  getEnvAsTimeDuration("DB_MAX_LIFETIME", 30*time.Minute),
				MaxIdleTime:  getEnvAsTimeDuration("DB_MAX_IDLE_TIME", 5*time.Minute),
				ReadTimeout:  getEnvAsTimeDuration("DB_READ_TIMEOUT", 5*time.Second),
				WriteTimeout: getEnvAsTimeDuration("DB_WRITE_TIMEOUT", 5*time.Second),
			},
			Auth: &structs.AuthConfig{
				AccessTokenSecret:  getEnvAsString("AUTH_ACCESS_TOKEN_SECRET", "default_access_secret"),
				AccessTokenExpiry:  getEnvAsTimeDuration("AUTH_ACCESS_TOKEN_EXPIRY", 15*time.Minute),
				RefreshTokenSecret: getEnvAsString("AUTH_REFRESH_TOKEN_SECRET", "default_refresh_secret"),
				RefreshTokenExpiry: getEnvAsTimeDuration("AUTH_REFRESH_TOKEN_EXPIRY", 7*24*time.Hour),
				CacheUserTTL:       getEnvAsTimeDuration("AUTH_CACHE_USER_TTL", 30*time.Minute),
				BlacklistCacheTTL:  getEnvAsTimeDuration("AUTH_BLACKLIST_CACHE_TTL", 7*24*time.Hour),
			},
			Cache: &structs.CacheConfig{
				Address:         getEnvAsString("CACHE_ADDRESS", "localhost:6379"),
				Username:        getEnvAsString("CACHE_USERNAME", ""),
				Password:        getEnvAsString("CACHE_PASSWORD", ""),
				DB:              getEnvAsInt("CACHE_DB", 0),
				PoolSize:        getEnvAsInt("CACHE_POOL_SIZE", 10),
				MinIdleConns:    getEnvAsInt("CACHE_MIN_IDLE_CONNS", 2),
				MaxIdleConns:    getEnvAsInt("CACHE_MAX_IDLE_CONNS", 5),
				PoolTimeout:     getEnvAsTimeDuration("CACHE_POOL_TIMEOUT", 30*time.Second),
				IdleTimeout:     getEnvAsTimeDuration("CACHE_IDLE_TIMEOUT", 5*time.Minute),
				DialTimeout:     getEnvAsTimeDuration("CACHE_DIAL_TIMEOUT", 5*time.Second),
				ReadTimeout:     getEnvAsTimeDuration("CACHE_READ_TIMEOUT", 3*time.Second),
				WriteTimeout:    getEnvAsTimeDuration("CACHE_WRITE_TIMEOUT", 3*time.Second),
				MaxRetries:      getEnvAsInt("CACHE_MAX_RETRIES", 3),
				MinRetryBackoff: getEnvAsTimeDuration("CACHE_MIN_RETRY_BACKOFF", 8*time.Millisecond),
				MaxRetryBackoff: getEnvAsTimeDuration("CACHE_MAX_RETRY_BACKOFF", 512*time.Millisecond),
				ProductListTTL:  getEnvAsTimeDuration("CACHE_PRODUCT_LIST_TTL", 5*time.Minute),
				ProductCountTTL: getEnvAsTimeDuration("CACHE_PRODUCT_COUNT_TTL", 10*time.Minute),
			},
			RateLimit: &structs.RateLimitConfig{
				Enabled:         getEnvAsBool("RATE_LIMIT_ENABLED", true),
				GeneralLimit:    getEnvAsInt("RATE_LIMIT_GENERAL_LIMIT", 100),
				GeneralWindow:   getEnvAsTimeDuration("RATE_LIMIT_GENERAL_WINDOW", 1*time.Minute),
				AuthLimit:       getEnvAsInt("RATE_LIMIT_AUTH_LIMIT", 5),
				AuthWindow:      getEnvAsTimeDuration("RATE_LIMIT_AUTH_WINDOW", 15*time.Minute),
				ExpensiveLimit:  getEnvAsInt("RATE_LIMIT_EXPENSIVE_LIMIT", 30),
				ExpensiveWindow: getEnvAsTimeDuration("RATE_LIMIT_EXPENSIVE_WINDOW", 1*time.Minute),
				AdminLimit:      getEnvAsInt("RATE_LIMIT_ADMIN_LIMIT", 50),
				AdminWindow:     getEnvAsTimeDuration("RATE_LIMIT_ADMIN_WINDOW", 1*time.Minute),
			},
			Email: &structs.EmailConfig{
				ApiKey:                  getEnvAsString("EMAIL_API_KEY", "no_api_key"),
				From:                    getEnvAsString("EMAIl_ADDRESS", "no_email"),
				SupportEmail:            getEnvAsString("EMAIL_SUPPORT_ADDRESS", "no_support_email"),
				VerificationTokenExpiry: getEnvAsTimeDuration("EMAIL_VERIFICATION_TOKEN_EXPIRY", 15*time.Minute),
			},
			Encryption: &structs.EncryptionConfig{
				Key: getEnvAsString("ENCRYPTION_KEY", ""),
			},
		}
	})
	return configInstance
}

func GetLogLevel() string {
	return GetConfig().Server.LogLevel
}

func IsProduction() bool {
	return GetConfig().Server.Environment == "production"
}
