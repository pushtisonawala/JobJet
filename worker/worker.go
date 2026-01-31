package worker

import (
	"context"
	"jobjet/logger"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
)

type Worker struct {
	Ctx        context.Context
	Cancel     context.CancelFunc
	MaxWorkers int
	Rdb        *redis.Client
	wg         sync.WaitGroup
	Paused     bool
}

func NewWorker(maxWorkers int, rdb *redis.Client) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		Ctx:        ctx,
		Cancel:     cancel,
		MaxWorkers: maxWorkers,
		Rdb:        rdb,
	}
}

// Pause the worker
func (w *Worker) Pause() {
	w.Paused = true
	logger.Log.Info("Worker paused")
}

// Resume the worker
func (w *Worker) Resume() {
	w.Paused = false
	logger.Log.Info("Worker resumed")
}

func (w *Worker) Start(processFunc func(string) error) {
	defer w.wg.Wait()

	// OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		logger.Log.Info("Shutdown signal received")
		w.Cancel()
	}()

	sem := make(chan struct{}, w.MaxWorkers)

	for {
		select {
		case <-w.Ctx.Done():
			logger.Log.Info("Worker shutting down")
			return
		default:
		}

		// If paused, skip fetching jobs
		if w.Paused {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		job, err := w.Rdb.BRPopLPush(
			w.Ctx,
			"job_queue",
			"processing_queue",
			0,
		).Result()

		if err != nil {
			time.Sleep(500 * time.Millisecond)
			continue
		}

		// update queue gauges ONCE per pop
		UpdateQueueMetrics(w.Ctx, w.Rdb)

		sem <- struct{}{}
		w.wg.Add(1)

		go func(j string) {
			defer func() {
				<-sem
				w.wg.Done()
			}()
			if err := processFunc(j); err != nil {
				logger.Log.Error("Job processing failed", "error", err)
			}
		}(job)
	}
}
