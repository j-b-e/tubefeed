package provider

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type ProviderNewSourceFn func(url string, logger *slog.Logger) (SourceProvider, error)

// SourceProvider can handle MEdia of a domain
type SourceProvider interface {
	LoadMetadata(ctx context.Context) (*SourceMeta, error)         // Provider starts requesting metadata
	Download(ctx context.Context, id uuid.UUID, path string) error // Provider downloads Source to path atomically
	//DownloadStream(id uuid.UUID) (io.Reader, error) // Download Source and return a Reader
	Url() string // Url to specific Source
}

type SourceMeta struct {
	ID          uuid.UUID // for later when id is removed from Download() signature
	ProviderID  string    // Provider Internal ID
	Title       string
	Channel     string
	Length      time.Duration
	Description string
	URL         string
}
