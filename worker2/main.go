package main

import (
"jobqueue/logger"
    "jobqueue/db"
    "jobqueue/worker"
    "github.com/redis/go-redis/v9"
)

func main() {
    logger.Log.Info("Worker started with concurrency")
    db.ConnectMongo() 
    rdb := redis.NewClient(&redis.Options{
        Addr: "localhost:6379",
    })
    maxWorkers := 5
    w := worker.NewWorker(maxWorkers, rdb)
    go worker.RetryScheduler(w.Ctx, rdb, &w.ShuttingDown)
    w.Start(func(jobData string) {
        worker.ProcessJob(w.Ctx, rdb, jobData)
    })
}