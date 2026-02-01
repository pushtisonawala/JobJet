package main

import (
	"context"
	"fmt"
	"jobjet/controllers"
	"jobjet/db"
	"jobjet/logger"
	"jobjet/metrics"
	"jobjet/queue"
	"net/http"
	"time"

	"jobjet/otel"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/gin-gonic/gin"
)

func main() {

	metrics.Init()
	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":2112", nil)
	logger.Log.Info("application starting")

	// 🆕 Add tracing
	shutdown, err := otel.InitTracer("jobqueue-api")
	if err == nil && shutdown != nil {
		defer shutdown(context.Background())
		logger.Log.Info("✅ Tracing initialized")
	}

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
	r.Use(otelgin.Middleware("jobqueue-api"))
	if err := r.SetTrustedProxies([]string{"127.0.0.1"}); err != nil {
		logger.Log.Error("Failed to set trusted proxies", "error", err)
	}
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

	if err := r.Run(":8000"); err != nil {
		logger.Log.Error("Gin server exited", "error", err)
	}
}
