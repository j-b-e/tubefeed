package main

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

var testDB *sql.DB

// Setup the router and test database before running the tests
func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	// Open the in-memory SQLite database for testing (won't persist on disk)
	testDB, _ = sql.Open("sqlite3", ":memory:")
	createTable() // Create table in the test database

	// Use the same handler setup
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	r.GET("/", func(c *gin.Context) {
		videos := LoadDatabase()
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Videos": videos,
		})
	})

	r.POST("/fetch-video-metadata", func(c *gin.Context) {
		youtubeURL := c.PostForm("youtube_url")
		videoID := extractVideoID(youtubeURL)
		videoMetadata := fetchYouTubeMetadata(videoID)

		// Save the metadata to the test database
		saveVideoMetadata(videoMetadata)

		videos := LoadDatabase()
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Videos": videos,
		})
	})

	r.DELETE("/delete-video/:id", func(c *gin.Context) {
		id := c.Param("id")
		deleteVideo(id)

		videos := LoadDatabase()
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Videos": videos,
		})
	})

	return r
}

// Test extracting YouTube video IDs
func TestExtractVideoID(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		input    string
		expected string
	}{
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ", "dQw4w9WgXcQ"},
		{"https://youtu.be/dQw4w9WgXcQ", ""},
		{"https://www.youtube.com/watch?v=dQw4w9WgXcQ&feature=youtu.be", "dQw4w9WgXcQ"},
	}

	for _, test := range tests {
		result := extractVideoID(test.input)
		assert.Equal(test.expected, result, "they should be equal")
	}
}

// Test adding video metadata via the POST /fetch-video-metadata route
func TestAddVideoMetadata(t *testing.T) {
	assert := assert.New(t)
	router := setupRouter()

	// Simulate form data
	form := "youtube_url=https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	req, _ := http.NewRequest("POST", "/fetch-video-metadata", strings.NewReader(form))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.Contains(w.Body.String(), "dQw4w9WgXcQ") // Check if the video ID is present in the response
}

// Test deleting a video via the DELETE /delete-video/:id route
func TestDeleteVideo(t *testing.T) {
	assert := assert.New(t)
	router := setupRouter()

	// Add a video first
	video := VideoMetadata{
		Title:   "Test Title",
		Channel: "Test Description",
		VideoID: "dQw4w9WgXcQ",
	}
	saveVideoMetadata(video)

	// Delete the video
	req, _ := http.NewRequest("DELETE", "/delete-video/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(http.StatusOK, w.Code)
	assert.NotContains(w.Body.String(), "dQw4w9WgXcQ") // Check if the video ID is no longer present in the response
}

// Test the RSS feed generation (optional)
func TestGeneratePodcastRSSFeed(t *testing.T) {
	assert := assert.New(t)

	// Add some video metadata
	videos := []VideoMetadata{
		{Title: "Video 1", Channel: "Description 1", VideoID: "abc123"},
		{Title: "Video 2", Channel: "Description 2", VideoID: "def456"},
	}

	// Generate RSS feed
	rss := generatePodcastRSSFeed(videos)

	// Check if RSS contains the necessary information
	assert.Contains(rss, "<title>Video 1</title>")
	assert.Contains(rss, "<guid>https://www.youtube.com/watch?v=abc123</guid>")
	assert.Contains(rss, "<enclosure url=\"https://example.com/audio/abc123.mp3\"")
	assert.Contains(rss, "<title>Video 2</title>")
}
