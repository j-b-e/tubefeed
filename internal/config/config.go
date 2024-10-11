package config

import (
	"os"
)

type Config struct {
	ListenPort  string
	AudioPath   string
	DbPath      string
	ExternalURL string
}

func Load() *Config {
	return &Config{
		ListenPort:  GetEnvOrDefault("LISTEN_PORT", "8091"),
		AudioPath:   GetEnvOrDefault("AUDIO_PATH", "./audio/"),
		DbPath:      "./config/tubefeed.db",
		ExternalURL: GetEnvOrDefault("EXTERNAL_URL", "localhost"),
	}
}

func GetEnvOrDefault(key, def string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return def
}
