package provider

import (
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type ProviderNewSourceFn func(url string, logger *slog.Logger) (SourceProvider, error)

// SourceProvider can handle Videos of a domain
type SourceProvider interface {
	LoadMetadata() (*SourceMeta, error)           // Provider starts requesting metadata
	Download(id uuid.UUID, basepath string) error // Provider must download audio atomicly to Path
	Url() string                                  // Url to specific Source
}

type SourceMeta struct {
	ProviderID  string // Provider Internal ID
	Title       string
	Channel     string
	Length      time.Duration
	Description string
	URL         string
}
