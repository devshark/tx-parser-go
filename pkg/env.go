package pkg

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func GetEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
func GetEnvInt64(key string, defaultValue int64) int64 {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	parse, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return parse
}

func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	parse, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return parse
}

func GetEnvValues(key string) []string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return []string{}
	}

	return strings.Split(value, ",")
}
