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
	}, func() { sqlite.Close() }, nil
}

// Fetches all video providers from the database
func (db *Database) LoadDatabase(ctx context.Context, tab int) ([]meta.Video, error) {
	rows, err := db.queries.LoadDatabase(ctx, sql.NullInt64{Int64: int64(tab), Valid: true})
	if err != nil {
		return nil, dbErr(err)
	}
	var videos []meta.Video
	for _, row := range rows {
		videomd := provider.VideoMeta{
			Length:      time.Duration(row.Length) * time.Second,
			ID:          uuid.MustParse(row.Uuid),
			URL:         row.Url,
			Channel:     row.Channel,
			Title:       row.Title,
			Description: "",
		}
		video := meta.Video{
			Meta: videomd,
		}
		videos = append(videos, video)
	}

	return videos, nil
}

func (db *Database) GetVideo(ctx context.Context, id uuid.UUID) (meta.Video, error) {

	row, err := db.queries.GetVideo(ctx, id.String())
	if err != nil {
		return meta.Video{}, dbErr(err)
	}

	videomd := provider.VideoMeta{
		ID:      id,
		Title:   row.Title,
		Length:  time.Duration(row.Length) * time.Second,
		Channel: row.Channel,
		URL:     row.Url,
	}

	video := meta.Video{
		Meta: videomd,
	}

	return video, nil

}

// Saves video metadata to the database
func (db *Database) SaveVideoMetadata(ctx context.Context, video meta.Video, tabid int) error {
	err := db.queries.SaveMetadata(
		ctx,
		sqlc.SaveMetadataParams{
			Uuid:    video.Meta.ID.String(),
			Title:   video.Meta.Title,
			Channel: video.Meta.Channel,
			Length:  int64(video.Meta.Length.Seconds()),
			Url:     video.Meta.URL,
			Tabid:   sql.NullInt64{Int64: int64(tabid), Valid: true},
		})
	if err != nil {
		return dbErr(err)
	}
	return nil
}

func (db *Database) CheckforDuplicate(ctx context.Context, video meta.Video, tabid int) (bool, error) {

	count, err := db.queries.CountDuplicate(
		ctx,
		sqlc.CountDuplicateParams{
			Url:   video.Meta.URL,
			Tabid: sql.NullInt64{Int64: int64(tabid), Valid: true},
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

func (db *Database) LoadTabs(ctx context.Context) (map[int]string, error) {
	rows, err := db.queries.LoadTabs(ctx)
	if err != nil {
		return nil, dbErr(err)
	}
	tabs := make(map[int]string)
	for _, row := range rows {
		tabs[int(row.ID)] = row.Name
	}
	return tabs, nil
}
func (db *Database) ChangeTabName(ctx context.Context, id int, name string) error {
	err := db.queries.ChangeTabName(
		ctx,
		sqlc.ChangeTabNameParams{
			Name: name,
			ID:   int64(id),
		})
	if err != nil {
		return dbErr(err)
	}
	return nil
}

func (db *Database) AddTab(ctx context.Context, name string) error {

	tabid, err := db.queries.GetLastTabId(ctx)
	if err != nil {
		return dbErr(err)
	}
	err = db.queries.AddTab(
		ctx,
		sqlc.AddTabParams{
			ID:   int64(tabid + 1),
			Name: name,
		},
	)
	if err != nil {
		return dbErr(err)
	}
	return nil
}

func (db *Database) DeleteTab(ctx context.Context, id int) error {
	err := db.queries.DeleteTab(ctx, int64(id))
	if err != nil {
		return dbErr(err)
	}
	err = db.queries.DeleteVideosFromTab(
		ctx,
		sql.NullInt64{
			Int64: int64(id),
			Valid: true,
		})
	if err != nil {
		return dbErr(err)
	}
	return nil
}
