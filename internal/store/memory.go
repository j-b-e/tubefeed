package store

import (
	"context"
	"fmt"
	"tubefeed/internal/models"

	"github.com/google/uuid"
)

type memory struct {
	items    map[uuid.UUID]*models.Request
	playlist map[uuid.UUID]models.Playlist
}

// NewMemoryStore initializes an in-memory store
func NewMemoryStore() Store {
	m := &memory{
		items:    make(map[uuid.UUID]*models.Request),
		playlist: make(map[uuid.UUID]models.Playlist),
	}
	_ = m.CreatePlaylist(context.Background(), uuid.MustParse(models.Default_playlist_id), models.Default_playlist_name)
	return m
}

func (m *memory) Close() error {
	return nil
}

func (m *memory) LoadDatabase(_ context.Context) (items []models.Request, err error) {
	for _, item := range m.items {
		items = append(items, *item)
	}
	return items, nil
}

func (m *memory) LoadFromPlaylist(_ context.Context, playlist uuid.UUID) (items []models.Request, err error) {
	for _, item := range m.items {
		if item.Playlist == playlist {
			items = append(items, *item)
		}
	}
	return items, nil
}

func (m *memory) CheckforDuplicate(_ context.Context, sourceurl string, _ uuid.UUID) (bool, error) {
	for _, item := range m.items {
		if item.SourceURL == sourceurl {
			return true, nil
		}
	}
	return false, nil
}

func (m *memory) SetStatus(_ context.Context, id uuid.UUID, status models.Status) error {
	if item, ok := m.items[id]; ok {
		item.Status = status
		m.items[id] = item
		return nil
	}
	return fmt.Errorf("item not found")

}

func (m *memory) UpdateItem(_ context.Context, item models.Request) error {
	m.items[item.ID] = &item
	return nil
}

func (m *memory) GetItem(_ context.Context, id uuid.UUID) (models.Request, error) {
	if item, ok := m.items[id]; ok {
		return *item, nil
	}
	return models.Request{}, fmt.Errorf("item not found")
}

func (m *memory) DeleteItem(_ context.Context, id uuid.UUID) error {
	delete(m.items, id)
	return nil
}

func (m *memory) CreatePlaylist(_ context.Context, id uuid.UUID, name string) error {
	m.playlist[id] = models.Playlist{
		ID:   id,
		Name: name,
	}
	return nil
}

func (m *memory) GetPlaylist(_ context.Context, id uuid.UUID) (string, error) {
	if playlist, ok := m.playlist[id]; ok {
		return playlist.Name, nil
	}
	return "", fmt.Errorf("playlist not found")
}

func (m *memory) DeletePlaylist(_ context.Context, id uuid.UUID) error {
	delete(m.playlist, id)
	return nil
}

func (m *memory) UpdatePlaylist(ctx context.Context, id uuid.UUID, name string) error {
	return m.CreatePlaylist(ctx, id, name)
}

func (m *memory) ListPlaylist(_ context.Context) ([]models.Playlist, error) {
	var playlists []models.Playlist
	for _, playlist := range m.playlist {
		playlists = append(playlists, playlist)
	}
	return playlists, nil
}
