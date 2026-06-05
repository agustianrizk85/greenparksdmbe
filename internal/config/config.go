// Package config loads runtime configuration from environment variables with
// sensible defaults, so the service runs out of the box for local development.
package config

import (
	"os"
	"time"
)

// Config holds the server runtime configuration.
type Config struct {
	Port        string        // HTTP port to listen on
	AllowOrigin string        // CORS allowed origin
	DataPath    string        // JSON file the data is persisted to
	SessionTTL  time.Duration // bearer-token session lifetime
}

// Load reads configuration from the environment, applying defaults.
func Load() Config {
	return Config{
		Port:        getenv("SDM_PORT", "8087"),
		AllowOrigin: getenv("SDM_ALLOW_ORIGIN", "*"),
		DataPath:    getenv("SDM_DATA_PATH", "data/sdm-data.json"),
		SessionTTL:  12 * time.Hour,
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
