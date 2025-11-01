package ytdlp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"strconv"
	"strings"
	"time"
	"tubefeed/internal/provider"

	"github.com/google/uuid"
)

type ytdlp struct {
	sourceurl string // if set assumes videometadata is fully refreshed from yt
	source    string // name of the source
	logger    *slog.Logger
}

var (
	ErrYtdlp = errors.New("ytdlp error")
)

func init() {
	provider.Register("youtube.com", newytprovider)
	provider.Register("youtu.be", newytprovider)
	provider.Register("archive.org", newarchiveprovider)
}

// newarchiveprovider implements ProviderNewVideoFn
func newarchiveprovider(url string, logger *slog.Logger) (provider.SourceProvider, error) {
	if !strings.Contains(url, "archive.org") {
		return nil, fmt.Errorf("%w: not an archive.org url: %s", ErrYtdlp, url)
	}
	return &ytdlp{
		logger:    logger.WithGroup("provider").With("name", "ytdlp", "source", "archive.org"),
		sourceurl: url,
		source:    "archive.org",
	}, nil
}

// newytprovider implements ProviderNewVideoFn
func newytprovider(url string, logger *slog.Logger) (provider.SourceProvider, error) {
	if !strings.Contains(url, "youtube.com") && !strings.Contains(url, "youtu.be") {
		return nil, fmt.Errorf("%w: not a youtube url: %s", ErrYtdlp, url)
	}
	ytid, err := extractYTVideoID(url)
	if err != nil {
		return nil, fmt.Errorf("%w: not a youtube url: %s", ErrYtdlp, url)
	}
	return &ytdlp{
		sourceurl: yturl(ytid),
		logger:    logger.WithGroup("provider").With("name", "ytdlp", "source", "youtube"),
		source:    "youtube",
	}, nil
}

func yturl(ytid string) string {
	return fmt.Sprintf("https://www.youtube.com/watch?v=%s", ytid)
}

func (y *ytdlp) URL() string {
	return y.sourceurl
}

// func (y *yt) DownloadStream(ctx context.Context) (reader io.Reader, err error) {
// 	start := time.Now()
// 	y.logger.InfoContext(fmt.Sprintf("⏳ Starting Download of %s", y.Url()))
// 	cmd := exec.Command(
// 		"yt-dlp",
// 		"--quiet",
// 		"--extract-audio",
// 		"--audio-format", "mp3",
// 		"-o", "-",
// 		y.Url(),
// 	)

// 	stdout, err := cmd.StdoutPipe()
// 	if err != nil {
// 		return nil, fmt.Errorf("%w: failed stdout pipe for cmd %s: %v", ErrYoutube, cmd, err)
// 	}
// 	stderr, err := cmd.StderrPipe()
// 	if err != nil {
// 		return nil, fmt.Errorf("%w: failed stderr pipe for cmd %s: %v", ErrYoutube, cmd, err)
// 	}
// 	y.logger.DebugContext(fmt.Sprintf("⏳ Running cmd %s", cmd))
// 	err = cmd.Start()
// 	if err != nil {
// 		errOutput, _ := io.ReadAll(stderr)
// 		return nil, fmt.Errorf("%w: failed cmd %s: %v: %s", ErrYoutube, cmd, err, errOutput)
// 	}
// 	if err := cmd.Wait(); err != nil {
// 		errOutput, _ := io.ReadAll(stderr)
// 		return nil, fmt.Errorf("%w: failed cmd %s: %v: %s", ErrYoutube, cmd, err, errOutput)
// 	}
// 	y.logger.InfoContext(fmt.Sprintf("✅ Finished Download of %s", y.Url()), "download.time", time.Since(start).String())
// 	return reader, nil
// }

func (y *ytdlp) Download(ctx context.Context, id uuid.UUID, path string, chanProgress chan<- int) error {
	defer close(chanProgress)
	start := time.Now()

	y.logger.InfoContext(ctx, fmt.Sprintf("⏳ Starting Download of %s", y.URL()))
	cmd := exec.CommandContext(
		ctx,
		"yt-dlp",
		"--quiet",
		//"--limit-rate", "100K",
		"--progress-delta", "5",
		"--progress",
		"--progress-template", "%(progress._percent)d",
		"--extract-audio",
		"--audio-format", "mp3",
		"--playlist-items", "1", // TODO: Support playlist download
		"-P", path,
		"-P", "temp:.cache",
		"-o", id.String(),
		y.URL(),
	)
	y.logger.DebugContext(ctx, fmt.Sprintf("⏳ Running cmd %s", cmd))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("%w: pipe failed cmd %s: %v", ErrYtdlp, cmd, err)
	}
	scanner := bufio.NewScanner(stdout)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// search for '\r'
		for i, b := range data {
			if b == '\r' {
				return i + 1, data[:i], nil
			}
		}
		if atEOF && len(data) > 0 {
			return len(data), data, nil
		}
		return 0, nil, nil
	})
	if err = cmd.Start(); err != nil {
		return fmt.Errorf("%w: start failed cmd %s: %v", ErrYtdlp, cmd, err)
	}
	var progress int
	for scanner.Scan() {
		output := scanner.Text()
		progress, _ = strconv.Atoi(output)
		y.logger.DebugContext(ctx, fmt.Sprintf("command progress output: %s / %d", output, progress))
		select {
		case chanProgress <- progress:
			continue
		case <-ctx.Done():
			y.logger.ErrorContext(ctx, "ctx closed while scanning")
			return fmt.Errorf("ctx closed while scanning")
		}
	}
	if err := scanner.Err(); err != nil {
		y.logger.DebugContext(ctx, fmt.Sprintf("scanner err: %v", err))
		return err
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("%w: wait failed cmd %s: %v", ErrYtdlp, cmd, err)
	}

	y.logger.InfoContext(ctx, fmt.Sprintf("✅ Finished Download of %s", y.URL()), "download.time", time.Since(start).String())
	return nil
}

// Refreshes YouTube video metadata
func (y *ytdlp) LoadMetadata(ctx context.Context) (*provider.SourceMeta, error) {
	var err error
	cmd := exec.CommandContext(
		ctx,
		"yt-dlp",
		"--quiet",
		"--skip-download",
		"--dump-json",
		"--playlist-items", "1", // TODO: Support playlist download
		y.URL(),
	)
	y.logger.DebugContext(ctx, fmt.Sprintf("⏳ running cmd: %s", cmd))
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%w: failed cmd %s: %v: %s", ErrYtdlp, cmd, err, out)
	}
	var result map[string]any
	err = json.Unmarshal([]byte(out), &result)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrYtdlp, err)
	}

	if y.source == "youtube" && yturl(result["id"].(string)) != y.sourceurl {
		return nil, fmt.Errorf("%w: video id from result didnt match", ErrYtdlp)
	}
	uploader, ok := result["uploader"].(string)
	if !ok {
		uploader = "unknown"
	}
	description, ok := result["description"].(string)
	if !ok {
		description = "unknown"
	}
	url, ok := result["url"].(string)
	if !ok {
		if y.source == "youtube" {
			url, ok = result["original_url"].(string)
			if !ok {
				return nil, fmt.Errorf("%w: unable to retrieve url for youtube", ErrYtdlp)
			}
		} else {
			return nil, fmt.Errorf("%w: unable to retrieve url", ErrYtdlp)
		}
	}

	meta := provider.SourceMeta{
		ProviderID:  y.sourceurl,
		Title:       result["title"].(string),
		Channel:     uploader,
		Length:      time.Duration(int(result["duration"].(float64))) * time.Second,
		URL:         url,
		Description: description,
	}
	return &meta, nil
}

// Extracts video ID from the provided YouTube URL
func extractYTVideoID(url string) (string, error) {
	if strings.Contains(url, "v=") {
		parts := strings.Split(url, "v=")
		return strings.Split(parts[1], "&")[0], nil
	}
	return "", fmt.Errorf("%w: no video URL", ErrYtdlp)
}
