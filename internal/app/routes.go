package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"tubefeed/internal/meta"
	"tubefeed/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	// Mutex to handle concurrent access during file download
	downloadMutex        sync.Mutex
	downloadAudioIDMutex = make(map[string]*sync.RWMutex)
	downloadInProgress   sync.Map
)

// GET /
func (a App) rootHandler(c *gin.Context) {
	ctx := c.Request.Context()
	videometa, err := a.loadVideoMeta(ctx)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.HTML(http.StatusOK, "index.html", gin.H{
		"playlist": nil,
		"audio":    videometa,
	})
}

// POST /new
func (a App) newRequestHandler(c *gin.Context) {
	ctx := c.Request.Context()
	videoURL := c.PostForm("youtube_url")
	if videoURL == "" {
		err := fmt.Errorf("no url provided")
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	duplicate, err := a.Db.CheckforDuplicate(ctx, videoURL, uuid.MustParse(models.Default_playlist))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	if duplicate {
		c.JSON(http.StatusConflict, gin.H{"conflict": "Audio already present"})
		return
	}

	// send request to worker
	request := models.Request{
		ID:       uuid.New(),
		URL:      videoURL,
		Playlist: uuid.MustParse(models.Default_playlist),
		Progress: 0,
		Done:     false,
		Title:    "unknown",
		Status:   models.StatusNew,
		Error:    nil,
	}

	select {
	case a.request <- request:
		c.Status(http.StatusAccepted)
		return
	default:
		c.AbortWithStatusJSON(http.StatusTooManyRequests, `{"error": "too many requests"`)
	}
}

// GET /audio/:id
func (a App) audioIDhandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	// Delete the video from the database
	err = a.deleteVideo(ctx, id)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

func (a App) handlecontent(c *gin.Context) {
	ctx := c.Request.Context()
	tabID := c.Param("id")
	if tabID == "" || tabID == "1" {
		videometa, err := a.Db.LoadDatabase(ctx)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		if tabID == "" {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"Videos": videometa,
				"tab":    1,
			})
		} else {
			c.HTML(http.StatusOK, "tabcontent.html", gin.H{
				"Videos": videometa,
				"tab":    1,
			})
		}

	} else {
		tabIDi, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		videometa, err := a.loadVideoMeta(ctx)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		c.HTML(http.StatusOK, "tabcontent.html", gin.H{
			"Videos": videometa,
			"tab":    tabIDi,
		})
	}
}

func (a App) loadVideoMeta(ctx context.Context) ([]meta.Source, error) {
	videos, err := a.Db.LoadDatabase(ctx)
	if err != nil {
		return nil, err
	}
	return videos, nil
}

func (a App) streamAudio(c *gin.Context) {
	ctx := c.Request.Context()
	audioID := c.Param("id")
	audioUUID, err := uuid.Parse(audioID)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
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
		log.Printf("download of id %s in progress\n", audioID)
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
		video, err := a.Db.GetItem(ctx, audioUUID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		go func() {
			defer audioMutex.Unlock()
			err := video.Download(audioFilePath)
			if err != nil {
				log.Println(err)
				return
			}
		}()
	}
	c.JSON(http.StatusProcessing, gin.H{"msg": "Audio is processing"})
}

// Deletes a video by ID from the database
func (a App) deleteVideo(ctx context.Context, id uuid.UUID) error {
	err := a.Db.DeleteItem(ctx, id)
	if err != nil {
		return err
	}
	file := fmt.Sprintf("%s/%s.mp3", a.config.AudioPath, id)
	err = os.Remove(file)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, fs.ErrNotExist)
}
