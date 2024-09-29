package video

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"tubefeed/internal/config"
	"tubefeed/internal/yt"
)

// VideoMetadata holds the data retrieved from YouTube API
type VideoMetadata struct {
	VideoID string
	Title   string
	Channel string
	Length  int
	Status  string
}

func DownloadAudioFile(id string) error {

	path := filepath.Join(config.AudioPath, fmt.Sprintf("%s.mp3", id))
	cachepath := fmt.Sprintf("./.cache/%s.mp3", id)
	log.Printf("Starting Download: %s", path)
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	cmd := exec.Command("yt-dlp", "--quiet", "--extract-audio", "--audio-format", "mp3", "-o", cachepath, yt.Yturl(id))
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
