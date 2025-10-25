package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	Default_playlist = "0353a984-1b53-4e45-bd2c-d4b5da90850f"
)

type Status string

var (
	StatusNew     Status = "New"
	StatusMeta    Status = "FetchingMeta"
	StatusLoading Status = "Downloading"
	StatusReady   Status = "Available"
	StatusError   Status = "Error"
)

type Request struct {
	ID       uuid.UUID
	Title    string
	Channel  string
	URL      string
	Playlist uuid.UUID
	Length   time.Duration
	Progress int
	Done     bool
	Error    *string
	Status   Status
}
