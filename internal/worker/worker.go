package worker

import (
	"context"
	"fmt"
	"log"

	"tubefeed/internal/db"
	"tubefeed/internal/meta"
	"tubefeed/internal/models"

	"github.com/google/uuid"
)

type Worker struct {
	req  chan models.Request
	db   *db.Database
	path string
}

func CreateWorkers(count int, db *db.Database, path string, req chan models.Request) (w Worker) {

	w = Worker{req: req}
	for i := range count {
		go w.start(i)
	}
	return Worker{req: req, db: db, path: path}
}

func (w *Worker) handleError(ctx context.Context, workerID int, videoID uuid.UUID, err error) {
	log.Printf("Error(worker %d): %v", workerID, err)
	err2 := w.db.SetStatus(ctx, videoID, meta.StatusError)
	if err2 != nil {
		log.Printf("Error(worker %d): %v", workerID, err)
	}
}

func (w *Worker) start(id int) {
	log.Printf("worker %d started.", id)
	ctx := context.Background()
	for r := range w.req {
		log.Printf("worker %d started job %s", id, r.Source.Meta.Title)
		// save id & url to db -> StatusNew
		err := r.Source.LoadMeta()
		if err != nil {
			w.handleError(ctx, id, r.Source.ID, err)
			continue
		}
		// save meta to db -> StateMeta
		err = w.db.SaveVideoMetadata(ctx, r.Source, r.Playlist, meta.StatusMeta)
		if err != nil {
			w.handleError(ctx, id, r.Source.ID, err)
			continue
		}
		// download & extract audio -> StateLoading
		err = w.db.SetStatus(ctx, r.Source.ID, meta.StatusLoading)
		if err != nil {
			w.handleError(ctx, id, r.Source.ID, err)
			continue
		}
		err = r.Source.Download(w.path)
		if err != nil {
			w.handleError(ctx, id, r.Source.ID, err)
			continue
		}
		// complete -> StatusReady
		err = w.db.SetStatus(ctx, r.Source.ID, meta.StatusReady)
		if err != nil {
			w.handleError(ctx, id, r.Source.ID, err)
		}
	}
	log.Printf("worker %d stopped.", id)
}

func (w *Worker) Download(src meta.Source) error {
	req := models.Request{Source: src, ID: uuid.New(), Playlist: uuid.MustParse(models.Default_playlist), Progress: 0, Done: false}
	select {
	case w.req <- req:
		log.Printf("Work Queued: %s", src.Meta.URL)
		return nil
	default:
		return fmt.Errorf("Worker Queue is full")
	}
}
