package provider

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SourceProvider can handle MEdia of a domain
type SourceProvider interface {
	LoadMetadata(ctx context.Context) (*SourceMeta, error) // Provider starts requesting metadata
	// Provider downloads Source to path atomically. It must close progress channel.
	Download(ctx context.Context, id uuid.UUID, path string, progress chan<- int) error
	//DownloadStream(id uuid.UUID) (io.Reader, error) // Download Source and return a Reader
	URL() string // Url to specific Source
}

type SourceMeta struct {
	ID           uuid.UUID // for later when id is removed from Download() signature
	ProviderID   string    // Provider Internal ID
	Title        string
	Channel      string
	Length       time.Duration
	Description  string
	URL          string
	ThumbnailURL string
}
