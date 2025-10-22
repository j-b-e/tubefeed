package app

import (
	"fmt"
	"log"
	"net/http"
	"tubefeed/internal/config"
	"tubefeed/internal/db"
	"tubefeed/internal/models"
	"tubefeed/internal/rss"
	"tubefeed/internal/worker"

	"github.com/gin-gonic/gin"

	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	config      *config.Config
	rss         *rss.RSS
	ExternalURL string
	Db          *db.Database
	worker      worker.Worker
	version     string
}

func Setup(version string) App {
	c := config.Load()

	return App{
		config:  c,
		rss:     rss.NewRSS(c.ExternalURL),
		version: version,
	}
}

// Run main app
func (a App) Run() (err error) {

	var closedb func()
	a.Db, closedb, err = db.NewDatabase(a.config.DbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer closedb()

	req := make(chan models.Request, 20)
	defer close(req)
	a.worker = worker.CreateWorkers(a.config.Workers, a.Db, a.config.AudioPath, req)

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.Static("/static", "./static")

	r.GET("/", a.rootHandler)

	// Add a new video by fetching its metadata
	r.POST("/audio", a.audioHandler)
	// status audio route
	r.GET("/audio/status/:id", a.statusAudio)
	// Stream or download audio route
	r.GET("/audio/:id", a.streamAudio)

	// Route to delete a video by ID
	r.DELETE("/audio/:id", a.audioIDhandler)

	r.GET("/rss/:id", a.rssHandler)

	r.GET("/content/:id", a.handlecontent)

	r.GET("/version", func(c *gin.Context) {
		json := []byte(`{"version": "` + a.version + `" }`)
		c.Data(http.StatusOK, gin.MIMEJSON, json)
	})

	// SSE Endpoint
	r.GET("/events", a.eventsHandler)

	return r.Run(fmt.Sprintf(":%s", a.config.ListenPort))
}
