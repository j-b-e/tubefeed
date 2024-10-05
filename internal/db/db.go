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
	log.Println("Creating Tables")
	db.createTabTable()
	db.createVideoTable()
}

func (db *Database) createVideoTable() {
	query := `
	CREATE TABLE IF NOT EXISTS videos (
		uuid  	 	TEXT PRIMARY KEY,
		title      	TEXT NOT NULL,
		channel    	TEXT NOT NULL,
		status    	TEXT NOT NULL,
		length     	INTEGER NOT NULL,  -- length is in seconds
		size		INTEGER,  -- size is in bytes
		url        	TEXT NOT NULL,
		tab			INTEGER,
		FOREIGN KEY(tab) REFERENCES tabs(id)
	);`
	_, err := db.handler.Exec(query)
	if err != nil {
		log.Fatal(dbErr(err))
	}
}

func (db *Database) createTabTable() {
	query := `
	CREATE TABLE IF NOT EXISTS tabs (
		id  	 	INTEGER PRIMARY KEY,
		name      	TEXT NOT NULL
	);
	INSERT OR IGNORE INTO tabs (id,  name) VALUES (1, "Tab 1");`
	_, err := db.handler.Exec(query)
	if err != nil {
		log.Fatal(dbErr(err))
	}
}

// Fetches all video providers from the database
func (db *Database) LoadDatabase(tab int) ([]provider.VideoProvider, error) {
	query := `SELECT uuid, title, channel, status, length, url FROM videos WHERE tab=(?)`
	rows, err := db.handler.Query(query, tab)
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
	query := `INSERT INTO videos (uuid, title, channel, status, length, url, tab) VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := db.handler.Exec(query, video.VideoID, video.Title, video.Channel, video.Status, len, video.URL, 1) // TODO: use real tabid
	if err != nil {
		log.Fatal(dbErr(err))
	}
}

func (db *Database) Close() {
	db.handler.Close()
}

func (db *Database) CheckforDuplicate(video provider.VideoProvider, tabid int) (bool, error) {
	query := `SELECT count(*) FROM videos WHERE url = (?) and tab = (?)`
	rows, err := db.handler.Query(query, video.Url(), tabid)
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

func (db *Database) LoadTabs() (map[int]string, error) {
	query := `SELECT * FROM tabs`
	rows, err := db.handler.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()
	var id int
	var name string
	tabs := make(map[int]string)
	for rows.Next() {
		err = rows.Scan(&id, &name)
		if err != nil {
			return nil, err
		}
		tabs[id] = name
	}
	return tabs, nil
}
func (db *Database) ChangeTabName(id int, name string) error {
	query := `UPDATE tabs SET name=(?) WHERE id=(?)`
	_, err := db.handler.Exec(query, name, id)
	if err != nil {
		return err
	}
	return nil
}
