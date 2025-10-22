package app

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /rss/:id
func (a App) rssHandler(c *gin.Context) {
	ctx := c.Request.Context()
	// Fetch all videos from the database
	playlistID, err := uuid.Parse(c.Param("playlist"))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	videos, err := a.Db.LoadPlaylist(ctx, playlistID)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	// Generate Podcast RSS feed with the video metadata
	rssfeed, err := a.rss.GeneratePodcastFeed(videos, playlistID)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.Data(http.StatusOK, "application/xml", []byte(rssfeed))
}
