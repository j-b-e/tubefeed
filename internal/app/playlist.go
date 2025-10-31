package app

import (
	"io"
	"tubefeed/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (a App) createPlaylistHandler(c *gin.Context) {
	logger := a.logger.With("handler", "createPlaylistHandler")
	ctx := c.Request.Context()
	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		if err == io.EOF {
			c.JSON(400, gin.H{"error": "Request body is empty"})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	uuid, err := uuid.NewV7()
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		c.JSON(500, gin.H{"error": "error with uuid"})
		return
	}
	err = a.Store.CreatePlaylist(ctx, uuid, body.Name)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		c.JSON(500, gin.H{"error": "createPlaylist failed"})
		return
	}
	c.JSON(201, gin.H{"id": uuid.String(), "name": body.Name})
}
func (a App) getPlaylistHandler(c *gin.Context) {
	ctx := c.Request.Context()
	logger := a.logger.With("handler", "getPlaylistHandler")
	id := c.Param("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid UUID"})
		return
	}
	playlist, err := a.Store.GetPlaylist(ctx, uuid)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		c.JSON(500, gin.H{"error": "getPlaylist failed"})
		return
	}
	c.JSON(200, gin.H{"id": id, "name": playlist})
}
func (a App) deletePlaylistHandler(c *gin.Context) {
	ctx := c.Request.Context()
	logger := a.logger.With("handler", "deletePlaylistHandler")
	id := c.Param("id")
	if id == models.Default_playlist_id {
		c.JSON(400, gin.H{"error": "cannot delete default playlist"})
		return
	}
	uuid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid UUID"})
		return
	}
	err = a.Store.DeletePlaylist(ctx, uuid)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		c.JSON(500, gin.H{"error": "deletetPlaylist failed"})
		return
	}
	c.Status(200)
}

func (a App) updatePlaylistHandler(c *gin.Context) {
	ctx := c.Request.Context()
	logger := a.logger.With("handler", "updatePlaylistHandler")
	id := c.Param("id")
	uuid, err := uuid.Parse(id)
	if err != nil {
		c.JSON(400, gin.H{"error": "invalid UUID"})
		return
	}
	var body struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		if err == io.EOF {
			c.JSON(400, gin.H{"error": "Request body is empty"})
			return
		}
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}
	err = a.Store.UpdatePlaylist(ctx, uuid, body.Name)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		c.JSON(500, gin.H{"error": "updatePlaylist failed"})
		return
	}
	c.Status(200)
}

func (a App) listPlaylistHandler(c *gin.Context) {
	ctx := c.Request.Context()
	logger := a.logger.With("handler", "listPlaylistHandler")
	playlists, err := a.Store.ListPlaylist(ctx)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		c.JSON(500, gin.H{"error": "listPlaylist failed"})
		return
	}
	c.JSON(200, playlists)
}
