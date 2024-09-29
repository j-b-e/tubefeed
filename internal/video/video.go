package video

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"tubefeed/internal/yt"
)

// VideoMetadata holds the data retrieved from YouTube API
type VideoMetadata struct {
	VideoID   string
	Title     string
	Channel   string
	Length    int
	Status    string
	AudioPath string // path to audio data
}

func (v VideoMetadata) Download() error {

	path := filepath.Join(v.AudioPath, fmt.Sprintf("%s.mp3", v.VideoID))
	cachepath := fmt.Sprintf("./.cache/%s.mp3", v.VideoID)
	log.Printf("Starting Download: %s", path)
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	cmd := exec.Command("yt-dlp", "--quiet", "--extract-audio", "--audio-format", "mp3", "-o", cachepath, yt.Yturl(v.VideoID))
	log.Println(cmd)
	err = cmd.Run()
	if err != nil {
		return err
	}
	err = os.Rename(cachepath, path)
	if err != nil {
		return err
	}
	log.Printf("Finished Download: %s", path)
	return nil
}
