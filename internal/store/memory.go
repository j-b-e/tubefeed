package store

import (
	"context"
	"fmt"
	"tubefeed/internal/models"

	"github.com/google/uuid"
)

type memory struct {
	items    map[uuid.UUID]*models.Request
	playlist map[uuid.UUID]string
}

// NewMemoryStore initializes an in-memory store
func NewMemoryStore() Store {
	m := &memory{
		items:    make(map[uuid.UUID]*models.Request),
		playlist: make(map[uuid.UUID]string),
	}
	_ = m.AddPlaylist(context.Background(), uuid.MustParse(models.Default_playlist_id), models.Default_playlist_name)
	return m
}

func (m *memory) AddPlaylist(_ context.Context, id uuid.UUID, name string) error {
	m.playlist[id] = name
	return nil
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

func (m *memory) GetPlaylistName(_ context.Context, id uuid.UUID) (string, error) {
	if playlist, ok := m.playlist[id]; ok {
		return playlist, nil
	}
	return "", fmt.Errorf("playlist not found")
}

func (m *memory) CheckforDuplicate(_ context.Context, url string, _ uuid.UUID) (bool, error) {
	for _, item := range m.items {
		if item.URL == url {
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

func (m *memory) SaveItemMetadata(_ context.Context, item models.Request, playlist uuid.UUID, status models.Status) error {
	item.Playlist = playlist
	item.Status = status
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
