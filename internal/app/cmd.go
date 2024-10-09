package app

import (
	"fmt"
	"log"
	"tubefeed/internal/config"
	"tubefeed/internal/db"
	"tubefeed/internal/provider"
	"tubefeed/internal/rss"

	"github.com/gin-gonic/gin"

	_ "github.com/mattn/go-sqlite3"
)

type App struct {
	config        *config.Config
	rss           *rss.RSS
	ExternalURL   string
	Db            *db.Database
	videoProvider *provider.Provider
}

func Setup() App {
	c := config.Load()
	return App{
		config:        c,
		rss:           rss.NewRSS(c.ExternalURL),
		videoProvider: config.SetupVideoProviders(),
	}
}

// Run main app
func (a App) Run() (err error) {

	a.Db, err = db.NewDatabase(a.config.DbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer a.Db.Close()

	err = a.Db.CreateTables()
	if err != nil {
		return err
	}

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.Static("/static", "./static")

	r.GET("/", a.rootHandler)

	// Add a new video by fetching its metadata
	r.POST("/audio", a.audioHandler)
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
