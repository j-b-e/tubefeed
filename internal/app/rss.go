package app

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GET /rss/:id
func (a App) rssHandler(c *gin.Context) {
	ctx := c.Request.Context()
	// Fetch all videos from the database
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	videos, err := a.Db.LoadDatabase(ctx, id)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	tabs, err := a.Db.LoadTabs(ctx)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	tabname, ok := tabs[id]
	if !ok {
		err = fmt.Errorf("failed to get tabname for id %d", id)
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	// Generate Podcast RSS feed with the video metadata
	rssfeed, err := a.rss.GeneratePodcastFeed(videos, tabname)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	c.Data(http.StatusOK, "application/xml", []byte(rssfeed))
}
