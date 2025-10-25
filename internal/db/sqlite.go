package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"tubefeed/internal/models"

	"tubefeed/internal/sqlc"

	"github.com/google/uuid"
)

// ErrDatabase is returned for any database related error
var ErrDatabase = errors.New("sqlite database error")

func dbErr(s any) error {
	return fmt.Errorf("%w: %v", ErrDatabase, s)
}

// Database wraps the sqlc generated queries and the database connection
type Database struct {
	queries *sqlc.Queries
	conn    *sql.DB
}

// NewDatabase initializes the database at the given path
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

// Close the database connection
func (db *Database) Close() error {
	return db.conn.Close()
}

// LoadDatabase loads all items from the database
func (db *Database) LoadDatabase(ctx context.Context) (items []models.Request, err error) {
	rows, err := db.queries.LoadDatabase(ctx)
	if err != nil {
		return nil, dbErr(err)
	}
	for _, row := range rows {
		item := models.Request{
			ID:       row.ID,
			Title:    row.Title,
			Playlist: row.PlaylistID,
			Progress: 100,
			Done:     true,
			Error:    nil,
			Status:   models.Status(row.Status),
		}
		items = append(items, item)
	}

	return items, nil
}

// LoadFromPlaylist loads all items from a specific playlist
func (db *Database) LoadFromPlaylist(ctx context.Context, playlist uuid.UUID) ([]models.Request, error) {
	rows, err := db.queries.LoadAudioFromPlaylist(ctx, playlist)
	if err != nil {
		return nil, dbErr(err)
	}
	var items []models.Request
	for _, row := range rows {
		request := models.Request{
			ID:       row.ID,
			Length:   time.Duration(row.Length.Int64) * time.Second,
			URL:      row.Url,
			Channel:  row.Channel,
			Title:    row.Title,
			Playlist: playlist,
			Progress: 100,
			Done:     true,
			Error:    nil,
			Status:   models.StatusReady,
		}
		items = append(items, request)
	}
	return items, nil
}

// GetPlaylistName retrieves the name of a playlist by its ID
func (db *Database) GetPlaylistName(ctx context.Context, id uuid.UUID) (string, error) {
	playlist, err := db.queries.LoadPlaylist(ctx, id)
	if err != nil {
		return "", dbErr(err)
	}
	return playlist.Name, nil
}

// GetItem retrieves a single item by its ID
func (db *Database) GetItem(ctx context.Context, id uuid.UUID) (models.Request, error) {
	row, err := db.queries.GetAudio(ctx, id)
	if err != nil {
		return models.Request{}, dbErr(err)
	}

	item := models.Request{
		ID:       id,
		Title:    row.Title,
		Length:   time.Duration(row.Length.Int64) * time.Second,
		Channel:  row.Channel,
		URL:      row.Url,
		Playlist: row.PlaylistID,
		Progress: 100,
		Done:     true,
		Error:    nil,
		Status:   models.StatusReady,
	}

	return item, nil
}

// SaveItemMetadata writes item metadata to the database
func (db *Database) SaveItemMetadata(
	ctx context.Context,
	item models.Request,
	playlist uuid.UUID,
	status models.Status,
) error {
	err := db.queries.SaveMetadata(
		ctx,
		sqlc.SaveMetadataParams{
			ID:         item.ID,
			Title:      item.Title,
			Channel:    item.Channel,
			Status:     string(status),
			Length:     sql.NullInt64{Int64: int64(item.Length.Seconds())},
			Url:        item.URL,
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
