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

// ensure yt implements VideoProvider
var _ provider.VideoProvider = &yt{}

type yt struct {
	meta provider.VideoMetadata
	ytid string // if set assumes videometadata is fully refreshed from yt
}

var (
	ErrYoutube = errors.New("youtube error")
)

// New implements ProviderNewVideoFn
func New(vm provider.VideoMetadata) (provider.VideoProvider, error) {
	if !strings.Contains(vm.URL, "youtube.com") {
		return nil, fmt.Errorf("%w: not a youtube url: %s", ErrYoutube, vm.URL)
	}
	ytid, err := extractVideoID(vm.URL)
	if err != nil {
		return nil, fmt.Errorf("%w: not a youtube url: %s", ErrYoutube, vm.URL)
	}
	vm.URL = url(ytid)
	return &yt{meta: vm}, nil
}

func (y *yt) SetMetadata(meta *provider.VideoMetadata) {
	ytid, err := extractVideoID(meta.URL)
	if err != nil {
		log.Printf("%v: %v", ErrYoutube, err)
		return
	}
	y.ytid = ytid
	y.meta = *meta
}

func (y *yt) New(url string) (uuid.UUID, error) {
	if !strings.Contains(url, "youtube.com") {
		return uuid.Nil, fmt.Errorf("%w: not a youtube url: %s", ErrYoutube, url)
	}
	ytid, err := extractVideoID(url)
	if err != nil {
		return uuid.Nil, err
	}
	y.ytid = ytid
	y.meta.VideoID = uuid.New()
	y.meta.URL = url
	return y.meta.VideoID, nil
}

func url(ytid string) string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", ytid)
}

func (y *yt) Url() string {
	return url(y.ytid)
}

func (y *yt) Download(path string) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	log.Printf("⏳ Starting Download: %s", path)
	cmd := exec.Command("yt-dlp", "--quiet", "--extract-audio", "--audio-format", "mp3", "-o", path, y.Url())
	log.Printf("⏳ running cmd:  %s\n", cmd)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("%w: failed cmd %s: %v", ErrYoutube, cmd, err)
	}
	log.Printf("✅ finished Download: %s", path)
	return nil
}

// Refreshes YouTube video metadata
func (y *yt) LoadMetadata() (*provider.VideoMetadata, error) {
	var err error
	if y.ytid != "" {
		return &y.meta, nil
	}

	y.ytid, err = extractVideoID(y.meta.URL)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("yt-dlp", "--quiet", "--skip-download", "--dump-json", y.Url())
	log.Printf("⏳ running cmd:  %s\n", cmd)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%w: failed cmd %s: %v", ErrYoutube, cmd, err)
	}
	var result map[string]any
	err = json.Unmarshal([]byte(out), &result)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrYoutube, err)
	}

	if result["id"] != y.ytid {
		return nil, fmt.Errorf("%w: video id from result didnt match", ErrYoutube)
	}
	y.meta.Title = result["title"].(string)
	y.meta.Channel = result["uploader"].(string)

	y.meta.Length = time.Duration(int(result["duration"].(float64))) * time.Second
	if y.meta.URL == "" {
		y.meta.URL = y.Url()
	}
	if y.meta.Status == "" {
		y.meta.Status = "unknwown"
	}
	return &y.meta, nil
}

// Extracts video ID from the provided YouTube URL
func extractVideoID(url string) (string, error) {
	if strings.Contains(url, "v=") {
		parts := strings.Split(url, "v=")
		return strings.Split(parts[1], "&")[0], nil
	}
	return "", fmt.Errorf("%w: no video URL", ErrYoutube)
}
