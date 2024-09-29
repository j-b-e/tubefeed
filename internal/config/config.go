package config

import (
	"os"
)

type Config struct {
	ListenPort  string
	AudioPath   string
	DbPath      string
	Hostname    string
	ExternalURL string
}

func Load() *Config {
	return &Config{

		ListenPort: GetEnvOrDefault("LISTE_PORT", "8091"),
		AudioPath:  GetEnvOrDefault("AUDIO_PATH", "./audio/"),
		DbPath:     "./config/tubefeed.db",
		Hostname:   GetEnvOrDefault("HOSTNAME", "localhost"),
	}
}

func GetEnvOrDefault(key, def string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return def
}
