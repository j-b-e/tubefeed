package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
	"tubefeed/internal/config"
	"tubefeed/internal/provider"
	"tubefeed/internal/utils"

	"github.com/google/uuid"
)

var ErrDatabase = errors.New("database error")

func dbErr(s any) error {
	return fmt.Errorf("%w: %v", ErrDatabase, s)
}

type Database struct {
	handler  *sql.DB
	provider *provider.Provider
}

func NewDatabase(path string) (*Database, error) {
	ret, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	return &Database{
		handler:  ret,
		provider: config.SetupVideoProviders(),
	}, nil
}

func (db *Database) CreateTable() {
	query := `
	CREATE TABLE IF NOT EXISTS videos (
		uuid  	 	TEXT PRIMARY KEY,
		title      	TEXT NOT NULL,
		channel    	TEXT NOT NULL,
		status    	TEXT NOT NULL,
		length     	INTEGER NOT NULL,  -- length is in seconds
		url        	TEXT NOT NULL
	);`
	_, err := db.handler.Exec(query)
	if err != nil {
		log.Fatal(dbErr(err))
	}
}

// Fetches all video providers from the database
func (db *Database) LoadDatabase() ([]provider.VideoProvider, error) {
	query := `SELECT uuid, title, channel, status, length, url FROM videos`
	rows, err := db.handler.Query(query)
	if err != nil {
		log.Println(dbErr(err))
		return nil, err
	}
	defer rows.Close()

	var videos []provider.VideoProvider
	for rows.Next() {
		var videomd provider.VideoMetadata
		var id string
		var len int
		err := rows.Scan(&id, &videomd.Title, &videomd.Channel, &videomd.Status, &len, &videomd.URL)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		videomd.Length = time.Duration(len) * time.Second
		videomd.VideoID = uuid.MustParse(id)
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

func (db *Database) GetVideo(id uuid.UUID) (provider.VideoProvider, error) {
	query := `SELECT title, channel, status, length, url FROM videos WHERE uuid = (?)`
	rows, err := db.handler.Query(query, id.String())
	if err != nil {
		log.Println(err)
		return nil, err
	}
	videomd := provider.VideoMetadata{
		VideoID: id,
	}
	if !rows.Next() {
		return nil, ErrDatabase
	}
	var len int
	err = rows.Scan(&videomd.Title, &videomd.Channel, &videomd.Status, &len, &videomd.URL)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	videomd.Length = time.Duration(len) * time.Second
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
func (db *Database) SaveVideoMetadata(video provider.VideoMetadata) {
	len := video.Length.Seconds()
	query := `INSERT INTO videos (uuid, title, channel, status, length, url) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := db.handler.Exec(query, video.VideoID, video.Title, video.Channel, video.Status, len, video.URL)
	if err != nil {
		log.Fatal(dbErr(err))
	}
}

func (db *Database) Close() {
	db.handler.Close()
}

func (db *Database) CheckforDuplicate(video provider.VideoProvider) (bool, error) {
	query := `SELECT count(*) FROM videos WHERE url = (?)`
	rows, err := db.handler.Query(query, video.Url())
	if err != nil {
		log.Println(err)
		return false, err
	}
	defer rows.Close()

	var length int
	rows.Next()
	err = rows.Scan(&length)
	if err != nil {
		return true, err
	}
	if length == 0 {
		return false, nil
	}
	return true, nil
}

func (db *Database) Delete(id uuid.UUID) error {
	query := `DELETE FROM videos WHERE uuid = ?`
	_, err := db.handler.Exec(query, id)
	if err != nil {
		return err
	}
	return nil
}
