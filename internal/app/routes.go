package app

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"tubefeed/internal/downloader"
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
func (a App) getRootHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"playlist": nil,
		"audio":    a.requests,
	})
}

// POST /new
func (a App) newRequestHandler(c *gin.Context) {
	ctx := c.Request.Context()
	logger := a.logger.With("handler", "newRequest")
	videoURL := c.PostForm("media_url")
	if videoURL == "" {
		err := fmt.Errorf("no url provided")
		logger.ErrorContext(ctx, err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	duplicate, err := a.Store.CheckforDuplicate(ctx, videoURL, uuid.MustParse(models.Default_playlist))
	if err != nil {
		logger.Error(err.Error())
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
func (a App) deleteAudioHandler(c *gin.Context) {
	ctx := c.Request.Context()
	logger := a.logger.With("handler", "deleteAudio")
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		logger.Error(err.Error())
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}

	// Delete the video from the database
	err = a.deleteVideo(ctx, id)
	if err != nil {
		logger.Error(err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.Status(http.StatusOK)
}

func (a App) streamAudio(c *gin.Context) {
	ctx := c.Request.Context()
	logger := a.logger.With("handler", "streamAudio")
	audioID := c.Param("id")
	audioUUID, err := uuid.Parse(audioID)
	if err != nil {
		logger.Error(err.Error())
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
		logger.Info(fmt.Sprintf("download of id %s in progress", audioID))
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
		item, err := a.Store.GetItem(ctx, audioUUID)
		if err != nil {
			logger.Error(err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		go func() {
			defer audioMutex.Unlock()
			source, err := downloader.NewSource(item.ID, item.URL, a.logger)
			if err != nil {
				logger.Error(err.Error())
				return
			}
			err = source.Download(audioFilePath)
			if err != nil {
				logger.Error(err.Error())
				return
			}
		}()
	}
	c.JSON(http.StatusProcessing, gin.H{"msg": "Audio is processing"})
}

// Deletes a video by ID from the database
func (a App) deleteVideo(ctx context.Context, id uuid.UUID) error {
	err := a.Store.DeleteItem(ctx, id)
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
