package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"tubefeed/internal/meta"
	"tubefeed/internal/provider"

	"tubefeed/internal/sqlc"

	"github.com/google/uuid"
)

var ErrDatabase = errors.New("sqlite database error")

func dbErr(s any) error {
	return fmt.Errorf("%w: %v", ErrDatabase, s)
}

type Database struct {
	queries *sqlc.Queries
}

func NewDatabase(path string) (db *Database, close func(), err error) {
	sqlite, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, nil, dbErr(err)
	}
	_, err = sqlite.Exec(sqlc.Schema)
	if err != nil {
		return nil, nil, dbErr(err)
	}
	return &Database{
		queries: sqlc.New(sqlite),
	}, func() { _ = sqlite.Close() }, nil
}

func (db *Database) LoadDatabase(ctx context.Context) ([]meta.Source, error) {
	rows, err := db.queries.LoadDatabase(ctx)
	if err != nil {
		return nil, dbErr(err)
	}
	var audios []meta.Source
	for _, row := range rows {
		audiomd := provider.SourceMeta{
			Length:      time.Duration(row.Length.Int64) * time.Second,
			URL:         row.Url,
			Channel:     row.Channel,
			Title:       row.Title,
			Description: "",
		}
		video := meta.Source{
			ID:     uuid.MustParse(row.Uuid),
			Meta:   audiomd,
			Status: meta.Status(row.Status),
		}
		audios = append(audios, video)
	}

	return audios, nil
}

func (db *Database) LoadPlaylist(ctx context.Context, playlist uuid.UUID) ([]meta.Source, error) {
	rows, err := db.queries.LoadPlaylist(ctx, sql.NullString{String: playlist.String()})
	if err != nil {
		return nil, dbErr(err)
	}
	var audios []meta.Source
	for _, row := range rows {
		audiomd := provider.SourceMeta{
			Length:      time.Duration(row.Length.Int64) * time.Second,
			URL:         row.Url,
			Channel:     row.Channel,
			Title:       row.Title,
			Description: "",
		}
		video := meta.Source{
			ID:     uuid.MustParse(row.Uuid),
			Meta:   audiomd,
			Status: meta.Status(row.Status),
		}
		audios = append(audios, video)
	}
	return audios, nil
}

func (db *Database) GetVideo(ctx context.Context, id uuid.UUID) (meta.Source, error) {

	row, err := db.queries.GetVideo(ctx, id.String())
	if err != nil {
		return meta.Source{}, dbErr(err)
	}

	videomd := provider.SourceMeta{

		Title:   row.Title,
		Length:  time.Duration(row.Length.Int64) * time.Second,
		Channel: row.Channel,
		URL:     row.Url,
	}

	video := meta.Source{
		ID:     id,
		Meta:   videomd,
		Status: meta.Status(row.Status),
	}

	return video, nil
}

// Saves video metadata to the database
func (db *Database) SaveVideoMetadata(ctx context.Context, video meta.Source, playlist uuid.UUID, status meta.Status) error {
	err := db.queries.SaveMetadata(
		ctx,
		sqlc.SaveMetadataParams{
			Uuid:       video.ID.String(),
			Title:      video.Meta.Title,
			Channel:    video.Meta.Channel,
			Status:     string(status),
			Length:     sql.NullInt64{Int64: int64(video.Meta.Length.Seconds())},
			Url:        video.Meta.URL,
			PlaylistID: sql.NullString{String: playlist.String()},
		},
	)
	if err != nil {
		return dbErr(err)
	}
	return nil
}

func (db *Database) CheckforDuplicate(ctx context.Context, video meta.Source, playlist uuid.UUID) (bool, error) {

	count, err := db.queries.CountDuplicate(
		ctx,
		sqlc.CountDuplicateParams{
			Url:        video.Meta.URL,
			PlaylistID: sql.NullString{String: playlist.String(), Valid: true},
		})
	if err != nil {
		return false, dbErr(err)
	}

	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (db *Database) DeleteVideo(ctx context.Context, id uuid.UUID) error {
	err := db.queries.DeleteVideo(ctx, id.String())
	if err != nil {
		return dbErr(err)
	}
	return nil
}

func (db *Database) SetStatus(ctx context.Context, id uuid.UUID, status meta.Status) error {
	err := db.queries.SetStatus(
		ctx,
		sqlc.SetStatusParams{
			Status: string(status),
			Uuid:   id.String(),
		},
	)
	if err != nil {
		return dbErr(err)
	}
	return nil
}
