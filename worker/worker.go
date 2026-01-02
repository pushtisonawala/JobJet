package worker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/redis/go-redis/v9"
)

type Worker struct {
	Ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	semaphore    chan struct{}
	Rdb          *redis.Client
	ShuttingDown bool
}

func NewWorker(maxWorkers int, rdb *redis.Client) *Worker {
	ctx, cancel := context.WithCancel(context.Background())
	return &Worker{
		Ctx:          ctx,
		cancel:       cancel,
		wg:           sync.WaitGroup{},
		semaphore:    make(chan struct{}, maxWorkers),
		Rdb:          rdb,
		ShuttingDown: false,
	}
}

func (w *Worker) Start(mainLoop func(jobData string)) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Shutdown signal received")
		w.ShuttingDown = true
		w.cancel()
		w.wg.Wait()
		fmt.Println("Worker exited cleanly")
		os.Exit(0)
	}()

	for !w.ShuttingDown {
		result, err := w.Rdb.BRPopLPush(w.Ctx, "job_queue", "processing_queue", 0).Result()
		if err != nil {
			if err == context.Canceled {
				continue
			}
			fmt.Println("❌ Error fetching job:", err)
			continue
		}

		w.semaphore <- struct{}{}
		w.wg.Add(1)

		go func(jobData string) {
			defer func() {
				<-w.semaphore
				w.wg.Done()
			}()
			mainLoop(jobData)
		}(result)
	}
}
