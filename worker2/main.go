package main

import (
	"jobqueue/db"
	"jobqueue/logger"
	"jobqueue/metrics"
	"jobqueue/worker"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger.Log.Info("Worker starting...")

	metrics.Init()
	metrics.WorkerRestarts.Inc()

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2113", nil)

	db.ConnectMongo()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	w := worker.NewWorker(5, rdb)

	go worker.RetryScheduler(w.Ctx, rdb)

	w.Start(func(job string) error {
		return worker.ProcessJob(w.Ctx, rdb, job)
	})
}
