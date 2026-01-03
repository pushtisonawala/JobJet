package controllers

import (
	"jobqueue/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"jobqueue/models"
	"jobqueue/queue"
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

	err := q.Push(job)
	if err != nil {
		logger.Log.Error("Job push failed", "job_id", job.ID, "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue job"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "job queued",
		"job_id":  job.ID,
	})
}
