package provider

import (
	"time"

	"github.com/google/uuid"
)

type ProviderNewVideoFn func(url string) (VideoProvider, error)

// VideoProvider can handle Videos of a domain
type VideoProvider interface {
	LoadMetadata() (*VideoMeta, error) // Provider starts requesting metadata
	Download(path string) error        // Provider must download audio atomicly to Path
	Url() string                       // Url to Website of specific Video
}

type VideoMeta struct {
	ID          uuid.UUID
	ProviderID  string // Provider Internal ID
	Title       string
	Channel     string
	Length      time.Duration
	Description string
	URL         string
}
