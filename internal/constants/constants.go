package constants

import (
	"fmt"
	"os"
	"strconv"
)

var (
	// Redis related constants
	RedisServiceHost    = getEnvOrDefault("REDIS_SERVICE_HOST", "urlshortener-redis")
	RedisServicePort    = getEnvOrDefault("REDIS_SERVICE_PORT", "6379")
	RedisServiceAddr    = fmt.Sprintf("%s:%s", RedisServiceHost, RedisServicePort)
	ClickCountKeyPrefix = getEnvOrDefault("CLICK_COUNT_KEY_PREFIX", "clicks:")

	// Controller related constants
	ReconcileInterval = getIntEnvOrDefault("RECONCILE_INTERVAL", 30) // seconds
	ShortPathLength   = getIntEnvOrDefault("SHORT_PATH_LENGTH", 3)   // characters
	LeaderElectionID  = getEnvOrDefault("LEADER_ELECTION_ID", "shorturl.tapsi.ir")
)

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getIntEnvOrDefault(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
