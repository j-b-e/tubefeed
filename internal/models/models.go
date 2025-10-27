package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	Default_playlist_id   = "0353a984-1b53-4e45-bd2c-d4b5da90850f"
	Default_playlist_name = "default"
)

type Status string

// Possible statuses for a Request
var (
	StatusNew       Status = "New"
	StatusMeta      Status = "FetchingMeta"
	StatusLoading   Status = "Downloading"
	StatusReady     Status = "Available"
	StatusError     Status = "Error"
	StatusDuplicate Status = "Duplicate"
	StatusDeleted   Status = "Deleted"
)

// Request represents a media download request
type Request struct {
	ID        uuid.UUID     `json:"id"`
	Title     string        `json:"title"`
	Channel   string        `json:"channel,omitempty"`
	SourceURL string        `json:"source_url"`
	Playlist  uuid.UUID     `json:"playlist_id"`
	Length    time.Duration `json:"length"`
	Progress  int           `json:"progress"`
	Done      bool          `json:"done"`
	Error     *string       `json:"error,omitempty"`
	Status    Status        `json:"status"`
	StreamURL string        `json:"stream_url,omitempty"`
}

// Playlist for collecting audio into lists
type Playlist struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
