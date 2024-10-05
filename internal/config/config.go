package config

import (
	"os"
	"tubefeed/internal/provider"
	"tubefeed/internal/provider/yt"
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
		ListenPort: GetEnvOrDefault("LISTEN_PORT", "8091"),
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

func SetupVideoProviders() (p *provider.Provider) {
	p = &provider.Provider{}
	p.List = make(provider.VideoProviderList)
	p.RegisterProvider([]string{"youtube.com"}, yt.New)
	p.RegisterProvider([]string{"www.youtube.com"}, yt.New)
	return p
}
