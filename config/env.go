package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func getEnvAsString(key string, defaultVal string) string {
	if value, exists := lookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if valueStr, exists := lookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return defaultVal
}

func getEnvAsTimeDuration(key string, defaultVal time.Duration) time.Duration {
	if valueStr, exists := lookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return time.Duration(value)
		}
	}
	return defaultVal
}

func getEnvAsBool(key string, defaultVal bool) bool {
	if valueStr, exists := lookupEnv(key); exists {
		if value, err := strconv.ParseBool(valueStr); err == nil {
			return value
		}
	}
	return defaultVal
}

func getEnvAsSlice(key string, defaultVal []string) []string {
	if valueStr, exists := lookupEnv(key); exists {
		// Split by comma and trim whitespace
		parts := strings.Split(valueStr, ",")
		result := make([]string, 0, len(parts))
		for _, v := range parts {
			trimmed := strings.TrimSpace(v)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return defaultVal
}

func lookupEnv(key string) (string, bool) {
	return os.LookupEnv(key)
}
