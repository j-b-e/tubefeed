package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"tubefeed/internal/meta"
	"tubefeed/internal/models"
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
	conn    *sql.DB
}

func NewDatabase(path string) (db *Database, err error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, dbErr(err)
	}

	ctx := context.Background()
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	_, err = tx.ExecContext(ctx, sqlc.Schema)
	if err != nil {
		return nil, dbErr(err)
	}

	db = &Database{
		queries: sqlc.New(conn),
		conn:    conn,
	}
	err = db.queries.WithTx(tx).AddPlaylist(
		ctx,
		sqlc.AddPlaylistParams{ID: uuid.MustParse(models.Default_playlist), Name: "default"},
	)
	if err != nil {
		return nil, dbErr(err)
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (db *Database) Close() error {
	return db.conn.Close()
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
			ID:     row.ID,
			Meta:   audiomd,
			Status: models.Status(row.Status),
		}
		audios = append(audios, video)
	}

	return audios, nil
}

func (db *Database) LoadAudioFromPlaylist(ctx context.Context, playlist uuid.UUID) ([]meta.Source, error) {
	rows, err := db.queries.LoadAudioFromPlaylist(ctx, playlist)
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
			ID:     row.ID,
			Meta:   audiomd,
			Status: models.Status(row.Status),
		}
		audios = append(audios, video)
	}
	return audios, nil
}

func (db *Database) GetPlaylistName(ctx context.Context, id uuid.UUID) (string, error) {
	playlist, err := db.queries.LoadPlaylist(ctx, id)
	if err != nil {
		return "", dbErr(err)
	}
	return playlist.Name, nil
}

func (db *Database) GetItem(ctx context.Context, id uuid.UUID) (meta.Source, error) {

	row, err := db.queries.GetAudio(ctx, id)
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
		Status: models.Status(row.Status),
	}

	return video, nil
}

// SaveItemMetadata writes item metadata to the database
func (db *Database) SaveItemMetadata(
	ctx context.Context,
	video meta.Source,
	playlist uuid.UUID,
	status models.Status,
) error {
	err := db.queries.SaveMetadata(
		ctx,
		sqlc.SaveMetadataParams{
			ID:         video.ID,
			Title:      video.Meta.Title,
			Channel:    video.Meta.Channel,
			Status:     string(status),
			Length:     sql.NullInt64{Int64: int64(video.Meta.Length.Seconds())},
			Url:        video.Meta.URL,
			PlaylistID: playlist,
		},
	)
	if err != nil {
		return dbErr(err)
	}
	return nil
}

// CheckforDuplicate returns false if the item is already in the playlist
func (db *Database) CheckforDuplicate(ctx context.Context, url string, playlist uuid.UUID) (bool, error) {

	count, err := db.queries.CountDuplicate(
		ctx,
		sqlc.CountDuplicateParams{
			Url:        url,
			PlaylistID: playlist,
		})
	if err != nil {
		return false, dbErr(err)
	}

	if count == 0 {
		return false, nil
	}
	return true, nil
}

// DeleteItem from the database
func (db *Database) DeleteItem(ctx context.Context, id uuid.UUID) error {
	err := db.queries.DeleteAudio(ctx, id)
	if err != nil {
		return dbErr(err)
	}
	return nil
}

// SetStatus in the database
func (db *Database) SetStatus(ctx context.Context, id uuid.UUID, status models.Status) error {
	err := db.queries.SetStatus(
		ctx,
		sqlc.SetStatusParams{
			Status: string(status),
			ID:     id,
		},
	)
	if err != nil {
		return dbErr(err)
	}
	return nil
}
