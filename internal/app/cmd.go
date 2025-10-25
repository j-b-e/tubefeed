package app

import (
	"context"
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
	version     string
	request     chan models.Request
	report      chan models.Request
	requests    []models.Request
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

	a.Db, err = db.NewDatabase(a.config.DbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err = a.Db.Close()
		if err != nil {
			panic(err)
		}
	}()

	a.request = make(chan models.Request)
	a.report = make(chan models.Request)
	defer close(a.request)
	err = worker.CreateWorkers(a.config.Workers, a.Db, a.config.AudioPath, a.request, a.report)
	if err != nil {
		panic(err)
	}
	go a.reportworker()

	a.requests, err = a.Db.LoadDatabase(context.Background())
	if err != nil {
		panic(err)
	}

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.Static("/static", "./static")

	r.GET("/", a.getRootHandler)

	// Add a new item
	r.POST("/audio", a.newRequestHandler)
	// Stream or download audio route
	r.GET("/audio/:id", a.streamAudio)
	// Route to delete an item by ID
	r.DELETE("/audio/:id", a.deleteAudioHandler)
	//r.PATCH("/audio/:id", a.patchAudioHandler)

	// Playlists
	// r.POST("/playlist", a.createPlaylistHandler)
	// r.GET("/playlist/:id", a.getPlaylistHandler)
	// Route to delete an item by ID
	// r.DELETE("/playlist/:id", a.deletePlaylistHandler)
	// r.PATCH("/playlist/:id", a.patchPlaylistHandler)

	r.GET("/rss/:id", a.getRSSHandler)

	r.GET("/version", func(c *gin.Context) {
		json := []byte(`{"version": "` + a.version + `" }`)
		c.Data(http.StatusOK, gin.MIMEJSON, json)
	})

	// SSE Endpoint
	r.GET("/events", a.eventsHandler)

	return r.Run(fmt.Sprintf(":%s", a.config.ListenPort))
}
