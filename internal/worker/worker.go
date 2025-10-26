// Package worker implements the worker pool to process download requests
package worker

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"tubefeed/internal/config"
	"tubefeed/internal/downloader"
	"tubefeed/internal/models"
	"tubefeed/internal/store"
	"tubefeed/internal/utils"
)

type WorkerManager struct {
	workers []Worker
}

// Worker processes download requests
type Worker struct {
	id             int
	report         chan<- models.Request // report channel to broadcast sse
	request        <-chan models.Request
	store          store.Store
	path           string
	reportInterval time.Duration
	logger         *slog.Logger
}

// CreateWorkers starts all configured workers
func CreateWorkers(
	count int,
	db store.Store,
	path string,
	req <-chan models.Request,
	report chan<- models.Request,
	logger *slog.Logger,
) error {

	w := WorkerManager{}
	for id := range count {
		worker := Worker{
			id:             id,
			report:         report,
			request:        req,
			store:          db,
			path:           path,
			reportInterval: config.Load().ReportInterval,
			logger:         logger.With("id", id),
		}
		w.workers = append(w.workers, worker)
		go worker.start()
	}
	return nil
}

func (w *Worker) handleError(ctx context.Context, item *models.Request, logger *slog.Logger, err error) {
	item.Error = utils.StringToPointer(err.Error())
	item.Status = models.StatusError
	logger.ErrorContext(ctx, fmt.Sprintf("Error: %v, Request: %#v", err, item))
	dberr := w.store.SetStatus(ctx, item.ID, item.Status)
	if dberr != nil {
		logger.ErrorContext(ctx, fmt.Sprintf("Error: %v", dberr))
	}
	// TODO: Cleanup old files
	w.report <- *item
}

func (w *Worker) start() {
	w.logger.Info("worker started")
	bctx := context.Background()
	for item := range w.request {
		ctx, cancel := context.WithTimeout(bctx, time.Duration(time.Hour))
		wlog := w.logger.With("item", item.ID)
		ticker := time.NewTicker(w.reportInterval)
		done := make(chan struct{})
		// 1. Ticker to report progress
		go func() {
			for {
				select {
				case <-done:
					wlog.DebugContext(ctx, "Ticker chan Done")
					ticker.Stop()
					w.report <- item
					return
				case <-ctx.Done():
					wlog.DebugContext(ctx, fmt.Sprintf("Ticker context Done - %v", ctx.Err()))
					ticker.Stop()
					w.report <- item
					return
				case <-ticker.C:
					w.report <- item
				}
			}
		}()
		// 2. Handle the request
		func(item *models.Request) {
			wlog.Info(fmt.Sprintf("started job for %q", item.URL))
			var err error
			defer func() {
				close(done)
				cancel()
				if err != nil {
					w.handleError(bctx, item, wlog, err) // use bctx to handle err outside of worker context
				}
			}()

			var source downloader.Source
			source, err = downloader.NewSource(item.ID, item.URL, wlog)
			if err != nil {
				return
			}
			item.Status = models.StatusNew
			// save id & url to db -> StatusNew
			err = source.LoadMeta(item)
			if err != nil {
				return
			}
			// save meta to db -> StateMeta
			item.Status = models.StatusMeta
			err = w.store.SaveItemMetadata(ctx, *item, item.Playlist, models.StatusMeta)
			if err != nil {
				return
			}
			// download & extract audio -> StateLoading
			item.Status = models.StatusLoading
			err = w.store.SetStatus(ctx, item.ID, models.StatusLoading)
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
			err = w.store.SetStatus(ctx, item.ID, models.StatusReady)
			if err != nil {
				return
			}
		}(&item)
	}
	w.logger.Info("worker stopped")
}
