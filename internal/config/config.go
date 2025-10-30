package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config stores all app settings
type Config struct {
	ListenPort     string
	AudioPath      string
	ExternalURL    string
	Workers        int
	ReportInterval time.Duration
	LogLevel       string
	DBPath         string
	//Store          store.Store
}

var config *Config

// Load configuration from env variables or sets defaults
func Load() *Config {
	if config != nil {
		return config
	}
	workers, err := strconv.Atoi(GetEnvOrDefault("WORKERS", "2"))
	if err != nil {
		panic(err)
	}

	config = &Config{
		ListenPort:     GetEnvOrDefault("LISTEN_PORT", "8091"),
		AudioPath:      GetEnvOrDefault("AUDIO_PATH", "./audio/"),
		ExternalURL:    GetEnvOrDefault("EXTERNAL_URL", "localhost"),
		Workers:        workers,
		ReportInterval: 3 * time.Second, // for sse
		LogLevel:       strings.ToLower(GetEnvOrDefault("LOG_LEVEL", "info")),
		DBPath:         GetEnvOrDefault("DATABASE_PATH", "./config/tubefeed.db"),
	}
	return config
}

// GetEnvOrDefault gets an env variable or returns a default value
func GetEnvOrDefault(key, def string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return def
}
