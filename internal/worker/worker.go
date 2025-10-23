package worker

import (
	"context"
	"log"
	"time"

	"tubefeed/internal/config"
	"tubefeed/internal/db"
	"tubefeed/internal/meta"
	"tubefeed/internal/models"
)

type WorkerManager struct {
	workers []Worker
}

type Worker struct {
	id             int
	report         chan<- models.Request // report channel to broadcast sse
	request        <-chan models.Request
	db             *db.Database
	path           string
	reportInterval time.Duration
}

func CreateWorkers(
	count int,
	db *db.Database,
	path string,
	req <-chan models.Request,
	report chan<- models.Request,
) error {

	w := WorkerManager{}
	for id := range count {
		worker := Worker{
			id:             id,
			report:         report,
			request:        req,
			db:             db,
			path:           path,
			reportInterval: config.Load().ReportInterval,
		}
		w.workers = append(w.workers, worker)
		go worker.start()
	}
	return nil
}

func s2p(s string) *string {
	return &s
}

func (w *Worker) handleError(ctx context.Context, item *models.Request, err error) {
	item.Error = s2p(err.Error())
	item.Status = models.StatusError
	log.Printf("Error(worker %d): %v, Request: %s", w.id, err, item.ID.String())
	dberr := w.db.SetStatus(ctx, item.ID, item.Status)
	if dberr != nil {
		log.Printf("Error(worker %d): %v", w.id, dberr)
	}
	// TODO: Cleanup old files
	w.report <- *item
}

func (w *Worker) start() {
	id := w.id
	log.Printf("worker %d started.", id)
	bctx := context.Background()
	for item := range w.request {
		ctx, cancel := context.WithTimeout(bctx, time.Duration(time.Hour))
		ticker := time.NewTicker(w.reportInterval)
		done := make(chan bool)
		// 1. Ticker to report progress
		go func() {
			for {
				select {
				case <-done:
					log.Printf("%s - Ticker chan Done", item.ID)
					ticker.Stop()
					return
				case <-ctx.Done():
					log.Printf("%s - Ticker context Done", item.ID)
					ticker.Stop()
					return
				case <-ticker.C:
					w.report <- item
				}
			}
		}()
		// 2. Handle the request
		func(item *models.Request) {
			log.Printf("worker %d started job for %s (%s)", id, item.ID, item.URL)
			var err error
			defer func() {
				cancel()
				close(done)
				if err != nil {
					w.handleError(bctx, item, err) // use bctx to handle err outside of worker context
				}
			}()

			var source meta.Source
			source, err = meta.NewSource(item.ID, item.URL)
			if err != nil {
				return
			}
			item.Status = models.StatusNew
			// save id & url to db -> StatusNew
			err = source.LoadMeta()
			if err != nil {
				return
			}
			// save meta to db -> StateMeta
			item.Status = models.StatusMeta
			err = w.db.SaveItemMetadata(ctx, source, item.Playlist, models.StatusMeta)
			if err != nil {
				return
			}
			// download & extract audio -> StateLoading
			item.Status = models.StatusLoading
			err = w.db.SetStatus(ctx, item.ID, models.StatusLoading)
			if err != nil {
				return
			}
			err = source.Download(w.path)
			if err != nil {
				return
			}
			// complete -> StatusReady
			item.Status = models.StatusReady
			item.Done = true
			err = w.db.SetStatus(ctx, item.ID, models.StatusReady)
			if err != nil {
				return
			}
		}(&item)
	}
	log.Printf("worker %d stopped.", id)
}

// func (w *WorkerManager) Download(src meta.Source) error {
// 	req := models.Request{ID: uuid.New(), Playlist: uuid.MustParse(models.Default_playlist), Progress: 0, Done: false}
// 	select {
// 	case w.request <- req:
// 		log.Printf("Work Queued: %s", src.Meta.URL)
// 		return nil
// 	default:
// 		return fmt.Errorf("Worker Queue is full")
// 	}
// }
