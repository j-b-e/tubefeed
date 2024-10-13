package app

import (
	"fmt"
	"log"
	"tubefeed/internal/config"
	"tubefeed/internal/db"
	"tubefeed/internal/meta/worker"
	"tubefeed/internal/rss"

	"github.com/gin-gonic/gin"

	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	config      *config.Config
	rss         *rss.RSS
	ExternalURL string
	Db          *db.Database
	worker      worker.Worker
}

func Setup() App {
	c := config.Load()

	return App{
		config: c,
		rss:    rss.NewRSS(c.ExternalURL),
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

	var closeworker func()
	a.worker, closeworker = worker.CreateWorkers(a.config.Workers, a.Db, a.config.AudioPath)
	defer closeworker()

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

	r.GET("/tab", a.tablist)
	r.GET("/tab/:id", a.tablist)
	r.PATCH("/tab/:id", a.patchtab)
	r.DELETE("/tab/:id", a.deleteTab)
	r.GET("/tab/edit/:id", a.edittab)
	r.POST("/tab", a.createtab)
	r.POST("/tab/:id", a.createtab) // Id just sets the active tab not new tabid

	return r.Run(fmt.Sprintf(":%s", a.config.ListenPort))
}
