package controllers

import (
	"net/http"
	"time"
    "fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"jobqueue/models"
	"jobqueue/queue"
)

func CreateJobs(c *gin.Context) {

	var req struct {
		Type    string `json:"type"`
		Payload map[string]interface{} `json:"payload"`
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
		CreatedAt: time.Now(),
	}

	q := queue.NewRedisQueue()

fmt.Println("Controller reached, job ID:", job.ID)
	if err := q.Push(job); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to queue job",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "job queued",
		"job_id":  job.ID,
	})
}
