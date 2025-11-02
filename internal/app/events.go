package app

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
	"tubefeed/internal/models"

	"github.com/gin-gonic/gin"
)

var (
	clients   = map[chan string]bool{}
	clientsMu sync.Mutex
)

func (a App) reportworker(logger *slog.Logger) {
	logger.Info("reportworker for SSE started")
	for msg := range a.report {
		a.broadcastProgress(&msg)
	}
}

func (a App) broadcastProgress(r *models.Request) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	msg, _ := json.Marshal(r)
	for ch := range clients {
		select {
		case ch <- string(msg):
		default:
		}
	}
}

func (a App) eventsHandler(c *gin.Context) {
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		_ = c.AbortWithError(
			http.StatusInternalServerError,
			fmt.Errorf("streaming unsupported"),
		)
		return
	}
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Status(http.StatusOK)
	flusher.Flush()

	messageChan := make(chan string)
	clientsMu.Lock()
	clients[messageChan] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, messageChan)
		clientsMu.Unlock()
		close(messageChan)
	}()

	keepalive := time.NewTicker(time.Second * 30)
	defer keepalive.Stop()
	c.Stream(func(w io.Writer) bool {
		select {
		case msg := <-messageChan:
			c.SSEvent("", msg)
		case <-keepalive.C:
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
		}
		return true
	})
}
