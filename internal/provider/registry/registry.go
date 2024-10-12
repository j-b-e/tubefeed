// interfaces/registry.go
package registry

import (
	"tubefeed/internal/provider"
	"tubefeed/internal/provider/yt"
)

var registry = map[string]provider.ProviderNewVideoFn{
	"youtube.com": yt.New,
}

func Register(name string, fn provider.ProviderNewVideoFn) {
	registry[name] = fn
}

func Get(name string) provider.ProviderNewVideoFn {
	if fn, ok := registry[name]; ok {
		return fn
	}
	return nil
}
