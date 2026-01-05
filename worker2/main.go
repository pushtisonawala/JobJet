package main

import (
	"jobqueue/controllers"
	"jobqueue/db"
	"jobqueue/logger"
	"jobqueue/metrics"
	"jobqueue/worker"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
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

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	w := worker.NewWorker(5, rdb)

	// Admin API endpoints
	adminRouter := gin.Default()
	adminRouter.SetTrustedProxies([]string{"127.0.0.1"})
	adminController := controllers.NewAdminController(rdb, w)
	adminRouter.GET("/admin/status", adminController.Status)
	adminRouter.POST("/admin/pause", adminController.Pause)
	adminRouter.POST("/admin/resume", adminController.Resume)
	go adminRouter.Run(":8001")

	go worker.RetryScheduler(w.Ctx, rdb)

	w.Start(func(job string) error {
		return worker.ProcessJob(w.Ctx, rdb, job)
	})
}
