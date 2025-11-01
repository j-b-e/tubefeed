package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
	"tubefeed/internal/models"

	"tubefeed/internal/sqlc"

	"github.com/google/uuid"
	_ "modernc.org/sqlite" // load SQLite driver
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

// NewSqliteDb initializes the database at the given path
func NewSqliteDb(path string) (Store, error) {
	conn, err := sql.Open("sqlite", path)
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

	db := &Database{
		queries: sqlc.New(conn),
		conn:    conn,
	}
	err = db.queries.WithTx(tx).CreatePlaylist(
		ctx,
		sqlc.CreatePlaylistParams{
			ID:   uuid.MustParse(models.Default_playlist_id),
			Name: models.Default_playlist_name,
		},
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
	rows, err := db.queries.ListAudio(ctx)
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
	rows, err := db.queries.LoadAudioByPlaylist(ctx, playlist)
	if err != nil {
		return nil, dbErr(err)
	}
	var items []models.Request
	for _, row := range rows {
		request := models.Request{
			ID:        row.ID,
			Length:    time.Duration(row.Length.Int64) * time.Second,
			SourceURL: row.SourceUrl,
			Channel:   row.Channel,
			Title:     row.Title,
			Playlist:  playlist,
			Progress:  100,
			Done:      true,
			Error:     nil,
			Status:    models.StatusReady,
		}
		items = append(items, request)
	}
	return items, nil
}

// GetPlaylistName retrieves the name of a playlist by its ID
func (db *Database) GetPlaylistName(ctx context.Context, id uuid.UUID) (string, error) {
	playlist, err := db.queries.GetPlaylist(ctx, id)
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
		ID:        id,
		Title:     row.Title,
		Length:    time.Duration(row.Length.Int64) * time.Second,
		Channel:   row.Channel,
		SourceURL: row.SourceUrl,
		Playlist:  row.PlaylistID,
		Progress:  100,
		Done:      true,
		Error:     nil,
		Status:    models.StatusReady,
	}

	return item, nil
}

// UpdateItem writes item metadata to the database
func (db *Database) UpdateItem(
	ctx context.Context,
	item models.Request,
) error {
	err := db.queries.CreateAudio(ctx, sqlc.CreateAudioParams{
		ID:         item.ID,
		Title:      item.Title,
		Channel:    item.Channel,
		Status:     string(item.Status),
		Length:     sql.NullInt64{Int64: int64(item.Length.Seconds())},
		SourceUrl:  item.SourceURL,
		PlaylistID: item.Playlist,
		UpdatedAt:  sql.NullTime{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return dbErr(err)
	}
	return nil
}

// CheckforDuplicate returns false if the item is already in the playlist
func (db *Database) CheckforDuplicate(ctx context.Context, sourceurl string, playlist uuid.UUID) (bool, error) {
	count, err := db.queries.CountDuplicate(
		ctx,
		sqlc.CountDuplicateParams{
			SourceUrl:  sourceurl,
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

// CreatePlaylist adds a new playlist to the database
func (db *Database) CreatePlaylist(ctx context.Context, id uuid.UUID, name string) error {
	err := db.queries.CreatePlaylist(ctx, sqlc.CreatePlaylistParams{ID: id, Name: name})
	if err != nil {
		return dbErr(err)
	}
	return nil
}

// GetPlaylist retrieves a playlist name by its ID
func (db *Database) GetPlaylist(ctx context.Context, id uuid.UUID) (string, error) {
	playlist, err := db.queries.GetPlaylist(ctx, id)
	if err != nil {
		return "", dbErr(err)
	}
	return playlist.Name, nil
}

// DeletePlaylist removes a playlist from the database
func (db *Database) DeletePlaylist(ctx context.Context, id uuid.UUID) error {
	err := db.queries.DeletePlaylist(ctx, id)
	if err != nil {
		return dbErr(err)
	}
	return nil
}

// UpdatePlaylist updates the name of a playlist
func (db *Database) UpdatePlaylist(ctx context.Context, id uuid.UUID, name string) error {
	err := db.queries.UpdatePlaylist(ctx, sqlc.UpdatePlaylistParams{
		ID:   id,
		Name: name,
	})
	if err != nil {
		return dbErr(err)
	}
	return nil
}

// ListPlaylist returns all playlists from the database
func (db *Database) ListPlaylist(context.Context) ([]models.Playlist, error) {
	panic("not implemented") // TODO: Implement
}

func (db *Database) CreateItem(ctx context.Context, item models.Request) error {
	err := db.queries.CreateAudio(ctx, sqlc.CreateAudioParams{
		ID:         item.ID,
		Title:      item.Title,
		Channel:    item.Channel,
		Length:     sql.NullInt64{Int64: 0},
		Size:       sql.NullInt64{Int64: 0},
		SourceUrl:  item.SourceURL,
		Status:     string(item.Status),
		ProviderID: uuid.UUID{},
		PlaylistID: item.Playlist,
		UpdatedAt:  sql.NullTime{Time: time.Now()},
	})
	if err != nil {
		return err
	}
	return nil
}
