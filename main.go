package main

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
)

const (
	listenPort = 8091
	audioPath  = "./audio/"
	dbPath     = "./config/yt2pod.db"
	hostname   = "luchs"
)

var (
	db *sql.DB
	// Mutex to handle concurrent access during file download
	downloadMutex        sync.Mutex
	externalURL          = fmt.Sprintf("%s:%d", hostname, listenPort)
	yturl                = func(id string) string { return fmt.Sprintf("https://www.youtube.com/watch?v=%s", id) }
	downloadAudioIDMutex = make(map[string]*sync.RWMutex)
	downloadInProgress   sync.Map
)

// VideoMetadata holds the data retrieved from YouTube API
type VideoMetadata struct {
	VideoID string
	Title   string
	Channel string
	Length  int
}

// PodcastRSS defines the structure for the podcast RSS XML feed
type PodcastRSS struct {
	XMLName     xml.Name       `xml:"rss"`
	Version     string         `xml:"version,attr"`
	XmlnsItunes string         `xml:"xmlns:itunes,attr"`
	Channel     PodcastChannel `xml:"channel"`
}

// PodcastChannel is the rss feed
type PodcastChannel struct {
	Title       string        `xml:"title"`
	Link        string        `xml:"link"`
	Description string        `xml:"description"`
	Language    string        `xml:"language"`
	Author      string        `xml:"itunes:author"`
	Image       PodcastImage  `xml:"itunes:image"`
	Items       []PodcastItem `xml:"item"`
}

// PodcastImage for the podcast
type PodcastImage struct {
	Href string `xml:"href,attr"`
}

// PodcastItem is an Item
type PodcastItem struct {
	Title       string           `xml:"title"`
	Description string           `xml:"description"`
	PubDate     string           `xml:"pubDate"`
	Link        string           `xml:"link"`
	GUID        string           `xml:"guid"`
	Enclosure   PodcastEnclosure `xml:"enclosure"`
}

// PodcastEnclosure is enclosure
type PodcastEnclosure struct {
	URL    string `xml:"url,attr"`
	Length string `xml:"length,attr"`
	Type   string `xml:"type,attr"`
}

func main() {
	var err error

	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable()

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	// Render the home page with the list of videos
	r.GET("/", func(c *gin.Context) {
		videos := LoadDatabase()
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Videos": videos,
		})
	})

	// Add a new video by fetching its metadata
	r.POST("/fetch-video-metadata", func(c *gin.Context) {
		youtubeURL := c.PostForm("youtube_url")
		videoID, err := extractVideoID(youtubeURL)
		if err != nil {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		if duplicate, _ := checkforDuplicate(videoID); duplicate == true {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		videoMetadata, err := fetchYouTubeMetadata(videoID)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Save the metadata to the database
		saveVideoMetadata(videoMetadata)

		// Reload the page with the updated video list
		videos := LoadDatabase()
		c.HTML(http.StatusOK, "video_list.html", gin.H{
			"Videos": videos,
		})
	})

	// Route to delete a video by ID
	r.DELETE("/delete-video/:id", func(c *gin.Context) {
		id := c.Param("id")

		// Delete the video from the database
		err := deleteVideo(id)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		// Reload the page with the updated video list
		videos := LoadDatabase()
		c.HTML(http.StatusOK, "video_list.html", gin.H{
			"Videos": videos,
		})
	})

	r.GET("/rss", func(c *gin.Context) {
		// Fetch all videos from the database
		videos := LoadDatabase()

		// Generate Podcast RSS feed with the video metadata
		rss := generatePodcastRSSFeed(videos)

		c.Data(http.StatusOK, "application/xml", []byte(rss))
	})

	// Stream or download audio route
	r.GET("/audio/:id", streamAudio)

	// Start the web server
	r.Run(fmt.Sprintf(":%d", listenPort))
}

func streamAudio(c *gin.Context) {
	audioID := c.Param("id")
	audioFilePath := filepath.Join(audioPath, fmt.Sprintf("%s.mp3", audioID))

	// Check if the file exists
	if fileExists(audioFilePath) {
		if _, ok := c.GetQuery("download"); ok {
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.mp3", audioID))
		} else {
			c.Header("Content-Type", "audio/mpeg")
			c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s.mp3", audioID))
		}
		c.File(audioFilePath)
		return
	}

	// Check if a download is already in progress
	if inProgress, _ := downloadInProgress.Load(audioID); inProgress == true {
		// Return early with a message indicating the download is in progress
		c.JSON(http.StatusProcessing, gin.H{"message": "Audio download in progress, please try again later"})
		return
	}

	downloadInProgress.Store(audioID, true)
	defer downloadInProgress.Delete(audioID)

	// File does not exist, attempt to download it
	downloadMutex.Lock()
	// Ensure that we have a mutex for the audioID
	if _, ok := downloadAudioIDMutex[audioID]; !ok {
		downloadAudioIDMutex[audioID] = &sync.RWMutex{}
	}
	audioMutex := downloadAudioIDMutex[audioID]
	defer downloadMutex.Unlock()

	audioMutex.Lock()

	if !fileExists(audioFilePath) {
		go func() {
			defer audioMutex.Unlock()
			err := downloadAudioFile(audioID)
			if err != nil {
				log.Println(err)
				return
			}
		}()
	}
	c.JSON(http.StatusProcessing, gin.H{"msg": "Audio is processing"})

}

func checkforDuplicate(videoID string) (bool, error) {
	query := `SELECT count(*) FROM videos WHERE video_id = (?)`
	rows, err := db.Query(query, videoID)
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

// Extracts video ID from the provided YouTube URL
func extractVideoID(url string) (string, error) {
	if strings.Contains(url, "v=") {
		parts := strings.Split(url, "v=")
		return strings.Split(parts[1], "&")[0], nil
	}
	return "", errors.New("No video URL")
}

// Fetches YouTube video metadata
func fetchYouTubeMetadata(videoID string) (VideoMetadata, error) {
	cmd := exec.Command("yt-dlp", "--quiet", "--skip-download", "--dump-json", yturl(videoID))
	out, err := cmd.Output()
	if err != nil {
		log.Println(err)
		return VideoMetadata{}, err
	}
	var result map[string]any
	json.Unmarshal([]byte(out), &result)

	if result["id"] != videoID {
		return VideoMetadata{}, errors.New("video id from result didnt match")
	}
	return VideoMetadata{
		Title:   result["title"].(string),
		Channel: result["uploader"].(string),
		Length:  int(result["duration"].(float64)),
		VideoID: videoID,
	}, nil
}

// Saves video metadata to the database
func saveVideoMetadata(video VideoMetadata) {
	query := `INSERT INTO videos (title, channel, video_id, length) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(query, video.Title, video.Channel, video.VideoID, video.Length)
	if err != nil {
		log.Fatal(err)
	}
}

// Fetches all video metadata from the database
func LoadDatabase() []VideoMetadata {
	query := `SELECT video_id, title, channel, length FROM videos`
	rows, err := db.Query(query)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var videos []VideoMetadata
	for rows.Next() {
		var video VideoMetadata
		err := rows.Scan(&video.VideoID, &video.Title, &video.Channel, &video.Length)
		if err != nil {
			log.Fatal(err)
		}
		videos = append(videos, video)
	}

	return videos
}

// Deletes a video by ID from the database
func deleteVideo(videoid string) error {
	query := `DELETE FROM videos WHERE video_id = ?`
	_, err := db.Exec(query, videoid)
	if err != nil {
		return err
	}
	err = os.Remove(fmt.Sprintf("%s/%s.mp3", audioPath, videoid))
	return err
}

// Creates the 'videos' table if it doesn't already exist
func createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS videos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		video_id TEXT,
		title TEXT,
		channel TEXT,
		length INT
	);`
	_, err := db.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}

// Generates a podcast RSS feed with the given metadata
func generatePodcastRSSFeed(videos []VideoMetadata) string {
	channel := PodcastChannel{
		Title:       "yt2pod",
		Link:        externalURL,
		Description: "A collection of YouTube videos as podcast episodes.",
		Language:    "en-us",
		Author:      "yt2pod",
		Image:       PodcastImage{Href: fmt.Sprintf("http://%s/static/podcast-cover.jpg", externalURL)},
	}

	for _, video := range videos {
		// Dynamically generate the full YouTube URL using the video ID
		videoURL := yturl(video.VideoID)
		audioURL := fmt.Sprintf("http://%s/audio/%s", externalURL, video.VideoID) // Stub for audio files

		item := PodcastItem{
			Title:       fmt.Sprintf("%s - %s", video.Channel, video.Title),
			Description: "Dummy Description",
			PubDate:     time.Now().Format("Tue, 15 Sep 2023 19:00:00 GMT"), //"Tue, 15 Sep 2023 19:00:00 GMT",
			Link:        videoURL,
			GUID:        videoURL,
			Enclosure: PodcastEnclosure{
				URL:    audioURL,                        // Replace this with the actual audio file URL
				Length: fmt.Sprintf("%d", video.Length), // Stub for the length of the audio file
				Type:   "audio/mpeg",                    // The type of enclosure
			},
		}
		channel.Items = append(channel.Items, item)
	}

	rss := PodcastRSS{
		Version:     "2.0",
		XmlnsItunes: "http://www.itunes.com/dtds/podcast-1.0.dtd",
		Channel:     channel,
	}

	output, _ := xml.MarshalIndent(rss, "", "  ")
	return xml.Header + string(output)
}

func downloadAudioFile(id string) error {
	path := fmt.Sprintf("%s/%s.mp3", audioPath, id)
	cachepath := fmt.Sprintf("./.cache/%s.mp3", id)
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	cmd := exec.Command("yt-dlp", "--quiet", "--extract-audio", "--audio-format", "mp3", "-o", cachepath, yturl(id))
	log.Println(cmd)
	err = cmd.Run()
	if err != nil {
		return err
	}
	err = os.Rename(cachepath, path)
	if err != nil {
		return err
	}
	return nil
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, fs.ErrNotExist)
}
