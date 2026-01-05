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

	// Debug endpoint to show job_queue contents
	r.GET("/debug/jobqueue", func(c *gin.Context) {
		jobs, err := q.Client().LRange(context.Background(), "job_queue", 0, -1).Result()
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"job_queue": jobs, "count": len(jobs)})
	})

	// Debug endpoint to show redis addr, ping and queue lengths
	r.GET("/debug/redis", func(c *gin.Context) {
		addr := ""
		if q.Client() != nil && q.Client().Options() != nil {
			addr = q.Client().Options().Addr
		}
		pong, _ := q.Client().Ping(context.Background()).Result()
		jobLen, _ := q.Client().LLen(context.Background(), "job_queue").Result()
		procLen, _ := q.Client().LLen(context.Background(), "processing_queue").Result()
		c.JSON(200, gin.H{"redis_addr": addr, "ping": pong, "job_queue_len": jobLen, "processing_queue_len": procLen})
	})

	r.Run(":8000")
}
