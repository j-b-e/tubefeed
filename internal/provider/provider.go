package provider

import (
	"time"

	"github.com/google/uuid"
)

// VideoProvider can handle Videos of a domain
type VideoProvider interface {
	LoadMetadata() (*VideoMetadata, error) // Provider starts requesting metadata
	SetMetadata(*VideoMetadata)            // Sets Metadata (eg from db)
	Download(path string) error            // Provider must download audio atomicly to Path
	Url() string                           // Url to Website of specific Video
}

type VideoProviderList map[string]ProviderNewVideoFn

type ProviderNewVideoFn func(vm VideoMetadata) (VideoProvider, error)

type Provider struct {
	List VideoProviderList
}

// VideoMetadata holds the data retrieved by Provider
type VideoMetadata struct {
	VideoID uuid.UUID
	Title   string
	Channel string
	Length  time.Duration
	Status  string
	URL     string
}

func (p *Provider) RegisterProvider(domains []string, provider ProviderNewVideoFn) {
	for _, val := range domains {
		p.List[val] = provider
	}
}
