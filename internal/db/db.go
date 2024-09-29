package db

import (
	"log"
	"tubefeed/internal/config"
	"tubefeed/internal/video"
)

// Creates the 'videos' table if it doesn't already exist
func CreateTable() {
	query := `
	CREATE TABLE IF NOT EXISTS videos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		video_id TEXT,
		title TEXT,
		channel TEXT,
		length INT
	);`
	_, err := config.Db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Fetches all video metadata from the database
func LoadDatabase() []video.VideoMetadata {
	query := `SELECT video_id, title, channel, length FROM videos`
	rows, err := config.Db.Query(query)
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
func SaveVideoMetadata(video video.VideoMetadata) {
	query := `INSERT INTO videos (title, channel, video_id, length) VALUES (?, ?, ?, ?)`
	_, err := config.Db.Exec(query, video.Title, video.Channel, video.VideoID, video.Length)
	if err != nil {
		log.Fatal(err)
	}
}
