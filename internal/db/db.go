package db

import (
	"database/sql"
	"log"
	"tubefeed/internal/video"
)

type Database struct {
	handler *sql.DB
}

func NewDatabase(path string) (*Database, error) {
	ret, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	return &Database{
		handler: ret,
	}, nil
}

// Creates the 'videos' table if it doesn't already exist
func (db *Database) CreateTable() {
	query := `
	CREATE TABLE IF NOT EXISTS videos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		video_id TEXT,
		title TEXT,
		channel TEXT,
		length INT
	);`
	_, err := db.handler.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Fetches all video metadata from the database
func (db *Database) LoadDatabase() []video.VideoMetadata {
	query := `SELECT video_id, title, channel, length FROM videos`
	rows, err := db.handler.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var videos []video.VideoMetadata
	for rows.Next() {
		var video video.VideoMetadata
		err := rows.Scan(&video.VideoID, &video.Title, &video.Channel, &video.Length)
		if err != nil {
			log.Fatal(err)
		}
		video.Status = "unknown"
		videos = append(videos, video)
	}

	return videos
}

// Saves video metadata to the database
func (db *Database) SaveVideoMetadata(video video.VideoMetadata) {
	query := `INSERT INTO videos (title, channel, video_id, length) VALUES (?, ?, ?, ?)`
	_, err := db.handler.Exec(query, video.Title, video.Channel, video.VideoID, video.Length)
	if err != nil {
		log.Fatal(err)
	}
}

func (db *Database) Close() {
	db.handler.Close()
}

func (db *Database) CheckforDuplicate(videoID string) (bool, error) {
	query := `SELECT count(*) FROM videos WHERE video_id = (?)`
	rows, err := db.handler.Query(query, videoID)
	if err != nil {
		log.Fatal(err)
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

func (db *Database) Delete(videoid string) error {
	query := `DELETE FROM videos WHERE video_id = ?`
	_, err := db.handler.Exec(query, videoid)
	if err != nil {
		return err
	}
	return nil
}
