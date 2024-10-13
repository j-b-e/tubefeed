package worker

import (
	"context"
	"fmt"
	"log"

	"tubefeed/internal/db"
	"tubefeed/internal/meta"

	"github.com/google/uuid"
)

type Request struct {
	Video meta.Video
	Tabid int
}

type Worker struct {
	req  chan Request
	db   *db.Database
	path string
}

func CreateWorkers(count int, db *db.Database, path string) (w Worker, closefn func()) {
	req := make(chan Request, 20)
	w = Worker{req: req}
	for i := range count {
		go w.start(i)
	}
	return Worker{req: req, db: db, path: path}, func() { close(req) }
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
		log.Printf("worker %d started job %s", id, r.Video.Meta.Title)
		// save id & url to db -> StatusNew
		err := r.Video.LoadMeta()
		if err != nil {
			w.handleError(ctx, id, r.Video.ID, err)
			continue
		}
		// save meta to db -> StateMeta
		err = w.db.SaveVideoMetadata(ctx, r.Video, r.Tabid, meta.StatusMeta)
		if err != nil {
			w.handleError(ctx, id, r.Video.ID, err)
			continue
		}
		// download & extract audio -> StateLoading
		err = w.db.SetStatus(ctx, r.Video.ID, meta.StatusLoading)
		if err != nil {
			w.handleError(ctx, id, r.Video.ID, err)
			continue
		}
		err = r.Video.Download(w.path)
		if err != nil {
			w.handleError(ctx, id, r.Video.ID, err)
			continue
		}
		// complete -> StatusReady
		err = w.db.SetStatus(ctx, r.Video.ID, meta.StatusReady)
		if err != nil {
			w.handleError(ctx, id, r.Video.ID, err)
		}
	}
	log.Printf("worker %d stopped.", id)
}

func (w *Worker) Download(video meta.Video, tabid int) error {
	// TODO: check video request is ok?
	select {
	case w.req <- Request{Video: video, Tabid: tabid}:
		log.Printf("Work Queued: %s", video.Meta.URL)
		return nil
	default:
		return fmt.Errorf("Worker Queue is full")
	}
}
