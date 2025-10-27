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
	CheckforDuplicate(ctx context.Context, sourceurl string, playlist uuid.UUID) (bool, error)
	UpdateItem(ctx context.Context, item models.Request) error
	GetItem(ctx context.Context, id uuid.UUID) (models.Request, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error
	Close() error // Close the database connection

	CreatePlaylist(ctx context.Context, id uuid.UUID, name string) error
	GetPlaylist(ctx context.Context, id uuid.UUID) (string, error)
	DeletePlaylist(ctx context.Context, id uuid.UUID) error
	UpdatePlaylist(ctx context.Context, id uuid.UUID, name string) error
	ListPlaylist(ctx context.Context) ([]models.Playlist, error)
}
