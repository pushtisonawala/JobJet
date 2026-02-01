package main

import (
	"context"
	"jobjet/controllers"
	"jobjet/db"
	"jobjet/logger"
	"jobjet/metrics"
	"jobjet/worker"
	"net/http"
	"os"

	"jobjet/otel"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger.Log.Info("Worker starting...")

	// 🆕 Add tracing
	shutdown, err := otel.InitTracer("jobqueue-worker")
	if err == nil && shutdown != nil {
		defer shutdown(context.Background())
		log.Println("✅ Worker tracing initialized")
	}

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

	w.Start(func(ctx context.Context, job string) error {
		return worker.ProcessJob(ctx, rdb, job)
	})
}
