package store

import (
	"context"
	"tubefeed/internal/models"

	"github.com/google/uuid"
)

// Store defines the interface for database operations
type Store interface {
	LoadDatabase(ctx context.Context) (items []models.Request, err error)
	LoadFromPlaylist(ctx context.Context, playlist uuid.UUID) ([]models.Request, error)
	GetPlaylistName(ctx context.Context, id uuid.UUID) (string, error)
	CheckforDuplicate(ctx context.Context, url string, playlist uuid.UUID) (bool, error)
	SetStatus(ctx context.Context, id uuid.UUID, status models.Status) error
	SaveItemMetadata(ctx context.Context, item models.Request, playlist uuid.UUID, status models.Status) error
	GetItem(ctx context.Context, id uuid.UUID) (models.Request, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error
	Close() error // Close the database connection
}
