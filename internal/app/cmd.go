package app

import (
	"encoding/json"
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
	"tubefeed/internal/config"
	"tubefeed/internal/db"
	"tubefeed/internal/rss"
	"tubefeed/internal/video"
	"tubefeed/internal/yt"

	"github.com/gin-gonic/gin"

	_ "github.com/mattn/go-sqlite3"
)

var (
	// Mutex to handle concurrent access during file download
	downloadMutex        sync.Mutex
	downloadAudioIDMutex = make(map[string]*sync.RWMutex)
	downloadInProgress   sync.Map
)

type App struct {
	config      *config.Config
	rss         *rss.RSS
	ExternalURL string
	Db          *db.Database
}

func Setup() App {
	c := config.Load()
	externalUrl := fmt.Sprintf("%s:%s", c.Hostname, c.ListenPort)
	return App{
		config: c,
		rss:    rss.NewRSS(externalUrl),
	}
}

// Run main app
func (a App) Run() (err error) {

	a.Db, err = db.NewDatabase(a.config.DbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer a.Db.Close()

	a.Db.CreateTable()

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		videos := a.Db.LoadDatabase()
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Videos": videos,
		})
	})

	// Add a new video by fetching its metadata
	r.POST("/audio", func(c *gin.Context) {
		youtubeURL := c.PostForm("youtube_url")
		videoID, err := extractVideoID(youtubeURL)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		duplicate, err := a.Db.CheckforDuplicate(videoID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		if duplicate {
			c.JSON(http.StatusConflict, gin.H{"conflict": "Audio already present"})
			return
		}
		videoMetadata, err := fetchYouTubeMetadata(videoID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		// Save the metadata to the database
		a.Db.SaveVideoMetadata(videoMetadata)

		// Reload the page with the updated video list
		videos := a.Db.LoadDatabase()
		c.HTML(http.StatusOK, "video_list.html", gin.H{
			"Videos": videos,
		})
	})

	// Stream or download audio route
	r.GET("/audio/:id", a.streamAudio)
	// Route to delete a video by ID
	r.DELETE("/audio/:id", func(c *gin.Context) {
		id := c.Param("id")

		// Delete the video from the database
		err := a.deleteVideo(id)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		// Reload the page with the updated video list
		videos := a.Db.LoadDatabase()
		c.HTML(http.StatusOK, "video_list.html", gin.H{
			"Videos": videos,
		})
	})

	r.GET("/rss", func(c *gin.Context) {
		// Fetch all videos from the database
		videos := a.Db.LoadDatabase()

		// Generate Podcast RSS feed with the video metadata
		rssfeed := a.rss.GeneratePodcastFeed(videos)

		c.Data(http.StatusOK, "application/xml", []byte(rssfeed))
	})

	// Start the web server
	return r.Run(fmt.Sprintf(":%s", a.config.ListenPort))
}

func (a App) streamAudio(c *gin.Context) {
	audioID := c.Param("id")
	audioFilePath := filepath.Join(a.config.AudioPath, fmt.Sprintf("%s.mp3", audioID))

	// Check if the file exists
	if fileExists(audioFilePath) {
		if _, ok := c.GetQuery("download"); ok {
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.mp3", audioID))
			c.File(audioFilePath)
			return
		}
		if _, ok := c.GetQuery("check"); ok {
			c.Status(http.StatusOK)
			return
		}
		c.Header("Content-Type", "audio/mpeg")
		c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s.mp3", audioID))
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
			file := video.VideoMetadata{
				VideoID:   audioID,
				AudioPath: a.config.AudioPath,
			}
			err := file.Download()
			if err != nil {
				log.Println(err)
				return
			}
		}()
	}
	c.JSON(http.StatusProcessing, gin.H{"msg": "Audio is processing"})
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
func fetchYouTubeMetadata(videoID string) (video video.VideoMetadata, err error) {
	cmd := exec.Command("yt-dlp", "--quiet", "--skip-download", "--dump-json", yt.Yturl(videoID))
	out, err := cmd.Output()
	if err != nil {
		log.Println(err)
		return video, err
	}
	var result map[string]any
	err = json.Unmarshal([]byte(out), &result)
	if err != nil {
		return video, err
	}

	if result["id"] != videoID {
		return video, errors.New("video id from result didnt match")
	}
	video.Title = result["title"].(string)
	video.Channel = result["uploader"].(string)
	video.Length = int(result["duration"].(float64))
	video.VideoID = videoID
	return video, nil
}

// Deletes a video by ID from the database
func (a App) deleteVideo(videoid string) error {
	err := a.Db.Delete(videoid)
	if err != nil {
		return err
	}
	err = os.Remove(fmt.Sprintf("%s/%s.mp3", a.config.AudioPath, videoid))
	return err
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, fs.ErrNotExist)
}
