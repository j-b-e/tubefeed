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
	version     string
	request     chan models.Request
	report      chan models.Request
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

	a.request = make(chan models.Request, 20)
	a.report = make(chan models.Request)
	defer close(a.request)
	err = worker.CreateWorkers(a.config.Workers, a.Db, a.config.AudioPath, a.request, a.report)
	if err != nil {
		panic(err)
	}
	go a.reportworker()

	r := gin.Default()

	r.LoadHTMLGlob("templates/*")

	r.Static("/static", "./static")

	r.GET("/", a.rootHandler)

	// Add a new item by fetching its metadata
	r.POST("/new", a.newRequestHandler)
	// // status audio route
	// r.GET("/audio/status/:id", a.statusAudio)
	// Stream or download audio route
	r.GET("/audio/:id", a.streamAudio)

	// Route to delete an item by ID
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
