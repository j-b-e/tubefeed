// interfaces/registry.go
package registry

import (
	"tubefeed/internal/provider"
	"tubefeed/internal/provider/yt"
)

var registry = map[string]provider.ProviderNewSourceFn{
	"youtube.com": yt.New,
	"youtu.be":    yt.New,
}

func Get(name string) provider.ProviderNewSourceFn {
	if fn, ok := registry[name]; ok {
		return fn
	}
	return nil
}
