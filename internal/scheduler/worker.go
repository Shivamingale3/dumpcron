package scheduler

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Shivamingale3/dumpcron/internal/drivers"
	"github.com/Shivamingale3/dumpcron/internal/events"
	"github.com/Shivamingale3/dumpcron/internal/retention"
)

type Worker struct {
	queue   <-chan WorkItem
	drivers map[string]drivers.Driver
	wg      sync.WaitGroup
}

func NewWorker(queue <-chan WorkItem, drvs map[string]drivers.Driver) *Worker {
	return &Worker{
		queue:   queue,
		drivers: drvs,
	}
}

func (w *Worker) Start() {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		for item := range w.queue {
			w.processJob(item)
		}
	}()
}

func (w *Worker) Wait() {
	w.wg.Wait()
}

func (w *Worker) processJob(item WorkItem) {
	events.EmitJobStarted(item.Job.Name)
	hadFailures := false

	drv, ok := w.drivers[item.Job.Type]
	if !ok {
		events.EmitJobFailed(item.Job.Name)
		return
	}

	for _, dbName := range item.Job.Databases {
		now := time.Now()
		timestamp := now.Format("2006-01-02_15-04")
		ext := drv.Extension()
		filename := fmt.Sprintf("%s_%s.%s.zst", dbName, timestamp, ext)

		dir := filepath.Join(item.Config.BackupRoot, item.Job.Type)
		if err := os.MkdirAll(dir, 0755); err != nil {
			events.EmitDatabaseBackupFailed(item.Job.Name, dbName, fmt.Sprintf("mkdir %s: %v", dir, err))
			hadFailures = true
			continue
		}

		outputPath := filepath.Join(dir, filename)

		ctx := context.Background()
		if err := drv.Dump(ctx, item.Job, dbName, outputPath); err != nil {
			events.EmitDatabaseBackupFailed(item.Job.Name, dbName, err.Error())
			hadFailures = true
		}
	}

	if hadFailures {
		events.EmitJobFailed(item.Job.Name)
	} else {
		events.EmitJobCompleted(item.Job.Name)
	}

	if err := retention.Cleanup(item.Config.BackupRoot, item.Job.Type, item.Config.RetentionDays); err != nil {
		fmt.Printf("retention cleanup error: %v\n", err)
	}
}
