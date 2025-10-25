package yt

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"time"
	"tubefeed/internal/provider"

	"github.com/google/uuid"
)

type yt struct {
	ytid   string // if set assumes videometadata is fully refreshed from yt
	logger *slog.Logger
}

var (
	ErrYoutube = errors.New("youtube error")
)

// New implements ProviderNewVideoFn
func New(url string, logger *slog.Logger) (provider.SourceProvider, error) {
	if !strings.Contains(url, "youtube.com") && !strings.Contains(url, "youtu.be") {
		return nil, fmt.Errorf("%w: not a youtube url: %s", ErrYoutube, url)
	}
	ytid, err := extractVideoID(url)
	if err != nil {
		return nil, fmt.Errorf("%w: not a youtube url: %s", ErrYoutube, url)
	}
	return &yt{ytid: ytid, logger: logger.WithGroup("provider").With("name", "youtube")}, nil
}

func url(ytid string) string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", ytid)
}

func (y *yt) Url() string {
	return url(y.ytid)
}

func (y *yt) Download(id uuid.UUID, path string) error {
	start := time.Now()
	_, err := os.Stat(path)
	if err != nil {
		return err
	}
	y.logger.Info(fmt.Sprintf("⏳ Starting Download of %s", y.Url()))
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
	y.logger.Info(fmt.Sprintf("⏳ Running cmd %s", cmd))
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: failed cmd %s: %v: %s", ErrYoutube, cmd, err, out)
	}
	y.logger.Info(fmt.Sprintf("✅ Finished Download of %s", y.Url()), "download.time", time.Since(start).String())
	return nil
}

// Refreshes YouTube video metadata
func (y *yt) LoadMetadata() (*provider.SourceMeta, error) {
	var err error

	cmd := exec.Command("yt-dlp", "--quiet", "--skip-download", "--dump-json", y.Url())
	y.logger.Info(fmt.Sprintf("⏳ running cmd: %s", cmd))
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
	meta := provider.SourceMeta{
		ProviderID: y.ytid,
		Title:      result["title"].(string),
		Channel:    result["uploader"].(string),
		Length:     time.Duration(int(result["duration"].(float64))) * time.Second,
		URL:        y.Url(),
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
