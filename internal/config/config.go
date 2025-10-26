package config

import (
	"os"
	"strconv"
	"strings"
	"time"
	"tubefeed/internal/store"
)

type Config struct {
	ListenPort     string
	AudioPath      string
	ExternalURL    string
	Workers        int
	ReportInterval time.Duration
	LogLevel       string
	Store          store.Store
}

// Load configuration from env variables or sets defaults
func Load() *Config {
	workers, err := strconv.Atoi(GetEnvOrDefault("WORKERS", "2"))
	if err != nil {
		panic(err)
	}
	dbstore := store.NewMemoryStore()
	//dbstore, err := store.NewSqliteDb(GetEnvOrDefault("DB_PATH", "./config/tubefeed.db"))

	return &Config{
		ListenPort:     GetEnvOrDefault("LISTEN_PORT", "8091"),
		AudioPath:      GetEnvOrDefault("AUDIO_PATH", "./audio/"),
		ExternalURL:    GetEnvOrDefault("EXTERNAL_URL", "localhost"),
		Workers:        workers,
		ReportInterval: 3 * time.Second, // for sse
		LogLevel:       strings.ToLower(GetEnvOrDefault("LOG_LEVEL", "info")),
		Store:          dbstore,
	}
}

// GetEnvOrDefault gets an env variable or returns a default value
func GetEnvOrDefault(key, def string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return def
}
