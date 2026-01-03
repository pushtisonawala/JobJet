package controllers

import (
	"context"
	"jobqueue/worker"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type AdminController struct {
	Rdb    *redis.Client
	Worker *worker.Worker
}

func NewAdminController(rdb *redis.Client, w *worker.Worker) *AdminController {
	return &AdminController{
		Rdb:    rdb,
		Worker: w,
	}
}

func (ac *AdminController) Status(c *gin.Context) {
	jobLen, _ := ac.Rdb.LLen(context.Background(), "job_queue").Result()
	procLen, _ := ac.Rdb.LLen(context.Background(), "processing_queue").Result()
	retryLen, _ := ac.Rdb.ZCard(context.Background(), "retry_zset").Result()
	dlqLen, _ := ac.Rdb.LLen(context.Background(), "dlq_queue").Result()

	c.JSON(http.StatusOK, gin.H{
		"job_queue":        jobLen,
		"processing_queue": procLen,
		"retry_queue":      retryLen,
		"dlq":              dlqLen,
		"worker_paused":    ac.Worker.Paused,
	})
}

func (ac *AdminController) Pause(c *gin.Context) {
	ac.Worker.Paused = true
	c.JSON(http.StatusOK, gin.H{"message": "Worker paused"})
}

func (ac *AdminController) Resume(c *gin.Context) {
	ac.Worker.Paused = false
	c.JSON(http.StatusOK, gin.H{"message": "Worker resumed"})
}
