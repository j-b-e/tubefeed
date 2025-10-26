package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"tubefeed/internal/config"
	"tubefeed/internal/models"
	"tubefeed/internal/rss"
	"tubefeed/internal/store"
	"tubefeed/internal/worker"

	"github.com/gin-gonic/gin"
)

// App defintion
type App struct {
	config      *config.Config
	rss         *rss.RSS
	ExternalURL string
	Store       store.Store
	version     string
	request     chan models.Request
	report      chan models.Request
	logger      *slog.Logger
	checkMu     *sync.Mutex
}

// Setup initializes the app with the given version
func Setup(version string) App {
	c := config.Load()

	return App{
		config:  c,
		rss:     rss.NewRSS(c.ExternalURL),
		version: version,
		checkMu: new(sync.Mutex),
	}
}

// Run main app
func (a App) Run() (err error) {
	ctx := context.Background()
	a.Store = a.config.Store
	defer func() {
		err = a.Store.Close()
		if err != nil {
			panic(err)
		}
	}()

	a.logger = a.createLogger()
	a.request = make(chan models.Request)
	a.report = make(chan models.Request)
	defer close(a.request)
	err = worker.CreateWorkers(
		ctx,
		a.config.Workers,
		a.Store,
		a.config.AudioPath,
		a.request,
		a.report,
		a.logger.WithGroup("worker"),
	)
	if err != nil {
		panic(err)
	}
	go a.reportworker(a.logger.WithGroup("reportworker"))

	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	api := r.Group("/api/v1")
	api.GET("/", a.apiGetRootHandler)
	api.POST("/audio", a.newRequestHandler)
	api.GET("/audio/:id", a.streamAudio)

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
