package provider

import "log/slog"

type ProviderNewSourceFn func(url string, logger *slog.Logger) (SourceProvider, error)

var registry = make(map[string]ProviderNewSourceFn)

// Register provider for a domain
func Register(domain string, p ProviderNewSourceFn) {
	registry[domain] = p
}

func Get(name string) ProviderNewSourceFn {
	if fn, ok := registry[name]; ok {
		return fn
	}
	return nil
}
