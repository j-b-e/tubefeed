package config

import (
	"os"
	"strconv"
)

type Config struct {
	ListenPort  string
	AudioPath   string
	DbPath      string
	ExternalURL string
	Workers     int
}

func Load() *Config {
	workers, err := strconv.Atoi(GetEnvOrDefault("WORKERS", "10"))
	if err != nil {
		panic(err)
	}
	return &Config{
		ListenPort:  GetEnvOrDefault("LISTEN_PORT", "8091"),
		AudioPath:   GetEnvOrDefault("AUDIO_PATH", "./audio/"),
		DbPath:      "./config/tubefeed.db",
		ExternalURL: GetEnvOrDefault("EXTERNAL_URL", "localhost"),
		Workers:     workers,
	}
}

func GetEnvOrDefault(key, def string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return def
}
