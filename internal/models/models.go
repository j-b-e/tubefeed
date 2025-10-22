package models

import (
	"tubefeed/internal/meta"

	"github.com/google/uuid"
)

const (
	Default_playlist = "0353a984-1b53-4e45-bd2c-d4b5da90850f"
)

type Request struct {
	ID       uuid.UUID
	Source   meta.Source
	Playlist uuid.UUID
	Progress int
	Done     bool
}
