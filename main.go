package main

import (
	"context"
	"fmt"
	"jobqueue/controllers"
	"jobqueue/db"
	"jobqueue/logger"
	"jobqueue/metrics"
	"jobqueue/queue"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gin-gonic/gin"
)

func main() {

	metrics.Init()
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2112", nil)
	logger.Log.Info("application starting")
	q := queue.NewRedisQueue()
	go func() {
		for {
			jobLen, _ := q.Client().LLen(context.Background(), "job_queue").Result()
			metrics.JobQueueLength.Set(float64(jobLen))
			procLen, _ := q.Client().LLen(context.Background(), "processing_queue").Result()
			metrics.ProcessingQueueLength.Set(float64(procLen))
			retryLen, _ := q.Client().ZCard(context.Background(), "retry_zset").Result()
			metrics.RetryQueueLength.Set(float64(retryLen))
			dlqLen, _ := q.Client().LLen(context.Background(), "dlq_queue").Result()
			metrics.DLQLength.Set(float64(dlqLen))
			time.Sleep(5 * time.Second)
		}
	}()
	r := gin.Default()
	r.SetTrustedProxies([]string{"127.0.0.1"})
	fmt.Println("server started")
	db.ConnectMongo()
	r.POST("/jobs", controllers.CreateJobs)
	r.Run(":8000")
}
