package app

import (
	"fmt"
	"net/http"
	"sync"
	"tubefeed/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/pingcap/log"
)

var (
	clients   = map[chan string]bool{}
	clientsMu sync.Mutex
)

func (a App) reportworker() {
	for msg := range a.report {
		a.broadcastProgress(&msg)
	}
}

func (a App) broadcastProgress(r *models.Request) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	var errs string
	if r.Error == nil {
		errs = ""
	} else {
		errs = *r.Error
	}
	msg := fmt.Sprintf(
		`{"id": "%s", "progress": %d, "done": %v, "status": "%s", "error": %q}`,
		r.ID, r.Progress, r.Done, r.Status, errs)
	for ch := range clients {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (a App) eventsHandler(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")

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

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		http.Error(c.Writer, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case msg := <-messageChan:
			_, err := fmt.Fprintf(c.Writer, "data: %s\n\n", msg)
			if err != nil {
				log.Error(err.Error())
			}
			flusher.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}
