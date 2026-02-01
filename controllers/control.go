package controllers

import (
	"jobjet/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"jobjet/models"
	"jobjet/queue"
)

func CreateJobs(c *gin.Context) {

	var req struct {
		Type    string              `json:"type"`
		Payload models.EmailPayload `json:"payload"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	job := models.Jobs{
		ID:        uuid.New().String(),
		Type:      req.Type,
		Payload:   req.Payload,
		Status:    "pending",
		MaxRetry:  3,
		CreatedAt: time.Now(),
	}

	q := queue.NewRedisQueue()
	logger.Log.Info("Controller reached", "job_id", job.ID)

	err := q.Push(c.Request.Context(), job)
	if err != nil {
		logger.Log.Error("Job push failed", "job_id", job.ID, "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue job"})
		return
	}

	if q.Client() != nil && q.Client().Options() != nil {
		addr := q.Client().Options().Addr
		jobLen, _ := q.Client().LLen(c.Request.Context(), "job_queue").Result()
		logger.Log.Info("Job pushed; redis info", "job_id", job.ID, "redis_addr", addr, "job_queue_len", jobLen)

		c.JSON(http.StatusAccepted, gin.H{
			"message":       "job queued",
			"job_id":        job.ID,
			"redis_addr":    addr,
			"job_queue_len": jobLen,
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "job queued",
		"job_id":  job.ID,
	})
}
