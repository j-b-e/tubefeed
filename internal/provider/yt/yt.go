package yt

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
	"tubefeed/internal/provider"

	"github.com/google/uuid"
)

type yt struct {
	ytid string // if set assumes videometadata is fully refreshed from yt
}

var (
	ErrYoutube = errors.New("youtube error")
)

// New implements ProviderNewVideoFn
func New(url string) (provider.VideoProvider, error) {
	if !strings.Contains(url, "youtube.com") {
		return nil, fmt.Errorf("%w: not a youtube url: %s", ErrYoutube, url)
	}
	ytid, err := extractVideoID(url)
	if err != nil {
		return nil, fmt.Errorf("%w: not a youtube url: %s", ErrYoutube, url)
	}
	return &yt{ytid: ytid}, nil
}

func url(ytid string) string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", ytid)
}

func (y *yt) Url() string {
	return url(y.ytid)
}

func (y *yt) Download(id uuid.UUID, path string) error {
	_, err := os.Stat(path)
	if err != nil {
		return err
	}
	log.Printf("⏳ yt: Starting Download: %s", path)
	cmd := exec.Command(
		"yt-dlp",
		"--quiet",
		"--extract-audio",
		"--audio-format", "mp3",
		"-P", path,
		"-P", "temp:.cache",
		"-o", id.String(),
		y.Url(),
	)
	log.Printf("⏳ yt: running cmd:  %s\n", cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: failed cmd %s: %v: %s", ErrYoutube, cmd, err, out)
	}
	log.Printf("✅ yt: finished Download: %s - %s", id, y.Url())
	return nil
}

// Refreshes YouTube video metadata
func (y *yt) LoadMetadata() (*provider.VideoMeta, error) {
	var err error

	cmd := exec.Command("yt-dlp", "--quiet", "--skip-download", "--dump-json", y.Url())
	log.Printf("⏳ running cmd:  %s\n", cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%w: failed cmd %s: %v: %s", ErrYoutube, cmd, err, out)
	}
	var result map[string]any
	err = json.Unmarshal([]byte(out), &result)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrYoutube, err)
	}

	if result["id"] != y.ytid {
		return nil, fmt.Errorf("%w: video id from result didnt match", ErrYoutube)
	}
	meta := provider.VideoMeta{
		Title:   result["title"].(string),
		Channel: result["uploader"].(string),
		Length:  time.Duration(int(result["duration"].(float64))) * time.Second,
		URL:     y.Url(),
	}

	return &meta, nil
}

// Extracts video ID from the provided YouTube URL
func extractVideoID(url string) (string, error) {
	if strings.Contains(url, "v=") {
		parts := strings.Split(url, "v=")
		return strings.Split(parts[1], "&")[0], nil
	}
	return "", fmt.Errorf("%w: no video URL", ErrYoutube)
}
