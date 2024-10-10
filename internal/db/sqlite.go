package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
	"tubefeed/internal/config"
	"tubefeed/internal/provider"
	"tubefeed/internal/utils"

	"tubefeed/internal/sqlc"

	"github.com/google/uuid"
)

var ErrDatabase = errors.New("sqlite database error")

func dbErr(s any) error {
	return fmt.Errorf("%w: %v", ErrDatabase, s)
}

type Database struct {
	provider *provider.Provider
	queries  *sqlc.Queries
}

func NewDatabase(path string) (db *Database, close func(), err error) {
	sqlite, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, nil, err
	}
	_, err = sqlite.Exec(sqlc.Schema)
	if err != nil {
		return nil, nil, dbErr(err)
	}
	return &Database{
		provider: config.SetupVideoProviders(),
		queries:  sqlc.New(sqlite),
	}, func() { sqlite.Close() }, nil
}

// Fetches all video providers from the database
func (db *Database) LoadDatabase(ctx context.Context, tab int) ([]provider.VideoProvider, error) {
	rows, err := db.queries.LoadDatabase(ctx, sql.NullInt64{Int64: int64(tab), Valid: true})
	if err != nil {
		return nil, err
	}
	var videos []provider.VideoProvider
	for _, row := range rows {
		videomd := provider.VideoMetadata{
			Length:  time.Duration(row.Length) * time.Second,
			VideoID: uuid.MustParse(row.Uuid),
			URL:     row.Url,
			Channel: row.Channel,
			Status:  row.Status,
			Title:   row.Title,
		}

		domain, err := utils.ExtractDomain(videomd.URL)
		if err != nil {
			return nil, err
		}
		if domain == "" {
			log.Printf("Domain for '%s' is empty\n", videomd.VideoID)
			continue
		}
		providerSetup, ok := db.provider.List[domain]
		if !ok {
			log.Printf("Provider for '%s' not found\n", domain)
			continue
		}
		video, err := providerSetup(videomd)
		if err != nil {
			log.Println(fmt.Errorf("Provider %s returned %w", video.Url(), err))
			continue
		}
		video.SetMetadata(&videomd)
		videos = append(videos, video)
	}

	return videos, nil
}

func (db *Database) GetVideo(ctx context.Context, id uuid.UUID) (provider.VideoProvider, error) {

	row, err := db.queries.GetVideo(ctx, id.String())
	if err != nil {
		return nil, err
	}

	videomd := provider.VideoMetadata{
		VideoID: id,
		Title:   row.Title,
		Length:  time.Duration(row.Length) * time.Second,
		Channel: row.Channel,
		Status:  row.Status,
		URL:     row.Url,
	}

	domain, err := utils.ExtractDomain(videomd.URL)
	if err != nil {
		return nil, err
	}
	provider, ok := db.provider.List[domain]
	if !ok {
		err = fmt.Errorf("%w: Provider for %s not found", ErrDatabase, domain)
		log.Println(err)
		return nil, err
	}
	video, err := provider(videomd)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	_, err = video.LoadMetadata()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return video, nil

}

// Saves video metadata to the database
func (db *Database) SaveVideoMetadata(ctx context.Context, video provider.VideoMetadata, tabid int) error {
	err := db.queries.SaveMetadata(
		ctx,
		sqlc.SaveMetadataParams{
			Uuid:    video.VideoID.String(),
			Title:   video.Title,
			Channel: video.Channel,
			Status:  video.Status,
			Length:  int64(video.Length.Seconds()),
			Url:     video.URL,
			Tabid:   sql.NullInt64{Int64: int64(tabid), Valid: true},
		})
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) CheckforDuplicate(ctx context.Context, video provider.VideoProvider, tabid int) (bool, error) {

	count, err := db.queries.CountDuplicate(
		ctx,
		sqlc.CountDuplicateParams{
			Url:   video.Url(),
			Tabid: sql.NullInt64{Int64: int64(tabid), Valid: true},
		})
	if err != nil {
		return false, err
	}

	if count == 0 {
		return false, nil
	}
	return true, nil
}

func (db *Database) DeleteVideo(ctx context.Context, id uuid.UUID) error {
	err := db.queries.DeleteVideo(ctx, id.String())
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) LoadTabs(ctx context.Context) (map[int]string, error) {
	rows, err := db.queries.LoadTabs(ctx)
	if err != nil {
		return nil, err
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
		return err
	}
	return nil
}

func (db *Database) AddTab(ctx context.Context, name string) error {

	tabid, err := db.queries.GetLastTabId(ctx)
	if err != nil {
		return err
	}
	err = db.queries.AddTab(
		ctx,
		sqlc.AddTabParams{
			ID:   int64(tabid + 1),
			Name: name,
		},
	)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) DeleteTab(ctx context.Context, id int) error {
	err := db.queries.DeleteTab(ctx, int64(id))
	if err != nil {
		return err
	}
	err = db.queries.DeleteVideosFromTab(
		ctx,
		sql.NullInt64{
			Int64: int64(id),
			Valid: true,
		})
	if err != nil {
		return err
	}
	return nil
}
