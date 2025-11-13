package app

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"tubefeed/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (a App) htmxPlaylist(c *gin.Context) {
	ctx := c.Request.Context()
	logger := a.logger.With("handler", "htmxPlaylist")
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	playlist, err := a.Store.LoadFromPlaylist(ctx, id)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.HTML(http.StatusOK, "itemlist.html", gin.H{
		"Items": playlist, "playlist_id": id,
	})
}

// POST /playlist (from form)
func (a App) createPlaylistFromFormHandler(c *gin.Context) {
	logger := a.logger.With("handler", "createPlaylistHandler")
	ctx := c.Request.Context()
	err := c.Request.ParseForm()
	if err != nil {
		logger.ErrorContext(ctx, "error", "msg", err.Error())
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	name, ok := c.Request.PostForm["name"]
	if !ok || len(name) == 0 || name[0] == "" {
		c.JSON(400, gin.H{"error": "invalid name"})
		return
	}
	if err := a.newPlaylist(ctx, name[0]); err != nil {
		logger.ErrorContext(ctx, "error", "msg", err.Error())
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	playlists, err := a.Store.ListPlaylist(ctx)
	if err != nil {
		logger.ErrorContext(ctx, "error", "msg", err.Error())
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.HTML(200, "playlist.html", gin.H{"Playlist": playlists})
}

func (a App) newPlaylist(ctx context.Context, name string) error {
	uuid, err := uuid.NewV7()
	if err != nil {
		a.logger.Error(err.Error())
		return fmt.Errorf("error with uuid")

	}
	err = a.Store.CreatePlaylist(ctx, uuid, name)
	if err != nil {
		a.logger.ErrorContext(ctx, err.Error())
		return fmt.Errorf("createPlaylist failed")
	}
	a.logger.InfoContext(ctx, "new playlist created", "name", name, "id", uuid)
	return nil
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
	if id == "" {
		id = c.Query("id")
	}
	if id == "" {
		c.JSON(400, gin.H{"error": "id missing"})
		return
	}
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
	playlists, err := a.Store.ListPlaylist(ctx)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		c.JSON(500, gin.H{"error": "deletetPlaylist failed"})
		return
	}
	c.HTML(200, "playlist.html", gin.H{"Playlist": playlists})
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
