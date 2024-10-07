package app

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"tubefeed/internal/config"
	"tubefeed/internal/db"
	"tubefeed/internal/provider"
	"tubefeed/internal/rss"
	"tubefeed/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	_ "github.com/mattn/go-sqlite3"
)

var (
	// Mutex to handle concurrent access during file download
	downloadMutex        sync.Mutex
	downloadAudioIDMutex = make(map[string]*sync.RWMutex)
	downloadInProgress   sync.Map
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

	a.Db.CreateTable()

	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.Static("/static", "./static")

	r.GET("/", func(c *gin.Context) {
		tabs, err := a.Db.LoadTabs()
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		videometa, err := a.loadVideoMeta(1)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		c.HTML(http.StatusOK, "index.html", gin.H{
			"tab":    1,
			"Tabs":   tabs,
			"Videos": videometa,
		})
	})

	// Add a new video by fetching its metadata
	r.POST("/audio", func(c *gin.Context) {
		videoURL := c.PostForm("youtube_url")
		tabid, err := strconv.Atoi(c.PostForm("tab"))
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		domain, err := utils.ExtractDomain(videoURL)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		providerSetup, ok := a.videoProvider.List[domain]
		if !ok {
			resp := fmt.Sprintf("No provider for %s found", domain)
			log.Println(resp)
			c.AbortWithStatusJSON(http.StatusInternalServerError, resp)
			return
		}
		vid, err := providerSetup(provider.VideoMetadata{VideoID: uuid.New(), URL: videoURL})
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		duplicate, err := a.Db.CheckforDuplicate(vid, tabid)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		if duplicate {
			c.JSON(http.StatusConflict, gin.H{"conflict": "Audio already present"})
			return
		}
		videoMetadata, err := vid.LoadMetadata()
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		// Save the metadata to the database
		a.Db.SaveVideoMetadata(*videoMetadata)

		// Reload the page with the updated video list
		videometa, err := a.loadVideoMeta(tabid)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		c.HTML(http.StatusOK, "video_list.html", gin.H{
			"Videos": videometa,
		})
	})

	// Stream or download audio route
	r.GET("/audio/:id", a.streamAudio)
	// Route to delete a video by ID
	r.DELETE("/audio/:id", func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, err)
			return
		}

		// Delete the video from the database
		err = a.deleteVideo(id)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		c.Status(http.StatusOK)
	})

	r.GET("/rss/:id", func(c *gin.Context) {
		// Fetch all videos from the database
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		videos, err := a.Db.LoadDatabase(id)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		tabs, err := a.Db.LoadTabs()
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
	})

	r.GET("/tab", a.tablist)
	r.GET("/tab/:id", a.handletab)
	r.PATCH("/tab/:id", a.patchtab)
	r.GET("/tab/edit/:id", a.edittab)

	return r.Run(fmt.Sprintf(":%s", a.config.ListenPort))
}

func (a App) streamAudio(c *gin.Context) {
	audioID := c.Param("id")
	audioUUID, err := uuid.Parse(audioID)
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusBadRequest, err)
		return
	}
	audioFilePath := filepath.Join(a.config.AudioPath, fmt.Sprintf("%s.mp3", audioID))

	// Check if the file exists
	if fileExists(audioFilePath) {
		if _, ok := c.GetQuery("download"); ok {
			c.Header("Content-Type", "application/octet-stream")
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.mp3", audioID))
			c.File(audioFilePath)
			return
		}
		if _, ok := c.GetQuery("check"); ok {
			c.Status(http.StatusOK)
			return
		}
		c.Header("Content-Type", "audio/mpeg")
		c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s.mp3", audioID))
		c.File(audioFilePath)
		return
	}

	// Check if a download is already in progress
	if inProgress, _ := downloadInProgress.Load(audioID); inProgress == true {
		// Return early with a message indicating the download is in progress
		log.Printf("download of id %s in progress\n", audioID)
		c.JSON(http.StatusProcessing, gin.H{"message": "Audio download in progress, please try again later"})
		return
	}

	downloadInProgress.Store(audioID, true)
	defer downloadInProgress.Delete(audioID)

	// File does not exist, attempt to download it
	downloadMutex.Lock()
	// Ensure that we have a mutex for the audioID
	if _, ok := downloadAudioIDMutex[audioID]; !ok {
		downloadAudioIDMutex[audioID] = &sync.RWMutex{}
	}
	audioMutex := downloadAudioIDMutex[audioID]
	defer downloadMutex.Unlock()

	audioMutex.Lock()

	if !fileExists(audioFilePath) {
		video, err := a.Db.GetVideo(audioUUID)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		_, err = video.LoadMetadata()
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}

		go func() {
			defer audioMutex.Unlock()
			err := video.Download(audioFilePath)
			if err != nil {
				log.Println(err)
				return
			}
		}()
	}
	c.JSON(http.StatusProcessing, gin.H{"msg": "Audio is processing"})
}

// Deletes a video by ID from the database
func (a App) deleteVideo(id uuid.UUID) error {
	err := a.Db.Delete(id)
	if err != nil {
		return err
	}
	file := fmt.Sprintf("%s/%s.mp3", a.config.AudioPath, id)
	err = os.Remove(file)
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, fs.ErrNotExist)
}

func (a App) handletab(c *gin.Context) {
	tabID := c.Param("id")
	if tabID == "" || tabID == "1" {
		videos, err := a.Db.LoadDatabase(1)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		var videometa []provider.VideoMetadata
		for _, video := range videos {
			meta, err := video.LoadMetadata()
			if err != nil {
				log.Println(err)
				continue
			}
			videometa = append(videometa, *meta)

		}
		if tabID == "" {
			c.HTML(http.StatusOK, "index.html", gin.H{
				"Videos": videometa,
				"tab":    1,
			})
		} else {
			c.HTML(http.StatusOK, "tabcontent.html", gin.H{
				"Videos": videometa,
				"tab":    1,
			})
		}

	} else {
		tabIDi, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		videometa, err := a.loadVideoMeta(tabIDi)
		if err != nil {
			log.Println(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, err)
			return
		}
		c.HTML(http.StatusOK, "tabcontent.html", gin.H{
			"Videos": videometa,
			"tab":    tabIDi,
		})
	}
}

func (a App) loadVideoMeta(tab int) ([]provider.VideoMetadata, error) {
	videos, err := a.Db.LoadDatabase(tab)
	if err != nil {
		return nil, err
	}
	var videometa []provider.VideoMetadata
	for _, v := range videos {
		meta, err := v.LoadMetadata()
		if err != nil {
			log.Println(err)
			continue
		}
		videometa = append(videometa, *meta)
	}
	return videometa, nil
}

// GET /tab/:id
func (a App) edittab(c *gin.Context) {
	tabid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	tabs, err := a.Db.LoadTabs()
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.HTML(http.StatusOK, "tabedit.html", gin.H{"Tab": tabid, "Name": tabs[tabid]})
}

// PATCH /tab/:id
func (a App) patchtab(c *gin.Context) {
	tabid, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}

	newname := c.PostForm("name")
	log.Printf("newname: %s", newname)

	err = a.Db.ChangeTabName(tabid, newname)
	if err != nil {
		// ignore err, old name will be reused
		log.Println(err)
	}
	tabs, err := a.Db.LoadTabs()
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.HTML(http.StatusOK, "tablist.html", gin.H{"Tabs": tabs, "tab": tabid})
}

// GET /tab
func (a App) tablist(c *gin.Context) {
	active, err := strconv.Atoi(c.Query("active"))
	if err != nil {
		active = 1
	}
	tabs, err := a.Db.LoadTabs()
	if err != nil {
		log.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, err)
		return
	}
	c.HTML(http.StatusOK, "tablist.html", gin.H{"Tabs": tabs, "tab": active})
}
