package constants

const (
	// Redis related constants
	RedisServiceAddr    = "redis-service:6379"
	ClickCountKeyPrefix = "clicks:"

	// Server related constants
	RedirectServerPort = ":8082"
	HealthProbePort    = ":8081"
	MetricsPort        = ":8080"

	// Controller related constants
	ReconcileInterval = 30 // seconds
	ShortPathLength   = 3  // characters
	LeaderElectionID  = "shorturl.tapsi.ir"

	// URL schemes
	SchemeHTTP  = "http"
	SchemeHTTPS = "https"
)
