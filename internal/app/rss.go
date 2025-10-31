package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /rss/:id
func (a App) getRSSHandler(c *gin.Context) {
	ctx := c.Request.Context()
	logger := a.logger.With("handler", "getRSS")
	// Fetch all videos from the database
	playlistID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	audio, err := a.Store.LoadFromPlaylist(ctx, playlistID)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	playlistName, err := a.Store.GetPlaylist(ctx, playlistID)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	// Generate Podcast RSS feed with the video metadata
	rssfeed, err := a.rss.GeneratePodcastFeed(audio, playlistName, playlistName)
	if err != nil {
		logger.ErrorContext(ctx, err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.Data(http.StatusOK, "application/xml", []byte(rssfeed))
}
