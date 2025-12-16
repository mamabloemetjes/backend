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
				ReadTimeout:    getEnvAsTimeDuration("SERVER_READ_TIME_OUT", 15*time.Second),
				WriteTimeout:   getEnvAsTimeDuration("SERVER_WRITE_TIME_OUT", 15*time.Second),
				IdleTimeout:    getEnvAsTimeDuration("SERVER_IDLE_TIME_OUT", 60*time.Second),
				MaxHeaderBytes: getEnvAsInt("SERVER_MAX_HEADER_BYTES", 1<<20), // 1 MB
			},
			Cors: &structs.CorsConfig{
				AllowOrigins:     getEnvAsSlice("CORS_ALLOW_ORIGINS", []string{"localhost", "http://localhost:3000"}),
				AllowMethods:     getEnvAsSlice("CORS_ALLOW_METHODS", []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
				AllowHeaders:     getEnvAsSlice("CORS_ALLOW_HEADERS", []string{"Origin", "Content-Type", "Accept", "Authorization"}),
				AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", false),
				ExposedHeaders:   getEnvAsSlice("CORS_EXPOSED_HEADERS", []string{"Content-Length", "Authorization"}),
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
			},
		}
	})
	return configInstance
}

func GetLogLevel() string {
	if GetConfig().Server.Environment == "production" {
		return "info"
	}
	return "debug"
}

func IsProduction() bool {
	return GetConfig().Server.Environment == "production"
}
