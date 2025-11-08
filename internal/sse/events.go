package sse

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type client struct {
	id string
	ch chan string
}

// Manager provides the Interface to send server-sent-events
type Manager interface {
	Broadcast(msg string)
	Handler(c *gin.Context)
}

type sse struct {
	clients   sync.Map
	broadcast chan string
	logger    *slog.Logger
}

// NewManager creates a new SSE Manager
func NewManager(logger *slog.Logger) Manager {
	manager := &sse{
		logger:    logger,
		broadcast: make(chan string),
	}
	go manager.broadcastWorker()
	return manager
}

func (m *sse) addClient(id string) <-chan string {
	ch := make(chan string)
	client := client{
		id: id,
		ch: ch,
	}
	m.clients.Store(id, client)
	m.logger.Info("added client", "id", id)
	return ch
}

func (m *sse) removeClient(id string) {
	cl, loaded := m.clients.LoadAndDelete(id)
	if loaded {
		close(cl.(client).ch)
	}
	m.logger.Debug("removed client", "id", cl.(client).id)
}

func (m *sse) Broadcast(msg string) {
	m.broadcast <- msg
}

func (m *sse) broadcastWorker() {
	var msg string
	m.logger.Info("started sse broadcastworker")
	for msg = range m.broadcast {
		m.clients.Range(func(id, cl any) bool {
			cl.(client).ch <- msg
			return true
		})
	}
}

func (m *sse) Handler(c *gin.Context) {
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

	id := uuid.New().String()
	ch := m.addClient(id)
	defer m.removeClient(id)

	keepalive := time.NewTicker(time.Second * 30)
	defer keepalive.Stop()
	c.Stream(func(w io.Writer) bool {
		select {
		case msg := <-ch:
			c.SSEvent("", msg)
		case <-keepalive.C:
			_, _ = fmt.Fprintf(w, ": keepalive\n\n")
		}
		return true
	})
}
