package controllers

import (
	"jobjet/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"jobjet/models"
	"jobjet/queue"
)

func CreateJobs(c *gin.Context) {
	// Create a span for this operation
	tracer := otel.Tracer("jobqueue-controller")
	ctx, span := tracer.Start(c.Request.Context(), "CreateJobs")
	defer span.End()

	var req struct {
		Type    string              `json:"type"`
		Payload models.EmailPayload `json:"payload"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to bind JSON")
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

	// Add span attributes
	span.SetAttributes(
		attribute.String("job.id", job.ID),
		attribute.String("job.type", job.Type),
		attribute.String("job.status", job.Status),
	)

	err := q.Push(ctx, job)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to push job to queue")
		logger.Log.Error("Job push failed", "job_id", job.ID, "error", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue job"})
		return
	}

	if q.Client() != nil && q.Client().Options() != nil {
		addr := q.Client().Options().Addr
		jobLen, _ := q.Client().LLen(ctx, "job_queue").Result()
		logger.Log.Info("Job pushed; redis info", "job_id", job.ID, "redis_addr", addr, "job_queue_len", jobLen)

		span.SetAttributes(
			attribute.String("redis.addr", addr),
			attribute.Int64("queue.length", jobLen),
		)
		span.SetStatus(codes.Ok, "Job successfully queued")

		c.JSON(http.StatusAccepted, gin.H{
			"message":       "job queued",
			"job_id":        job.ID,
			"redis_addr":    addr,
			"job_queue_len": jobLen,
		})
		return
	}

	span.SetStatus(codes.Ok, "Job successfully queued")
	c.JSON(http.StatusAccepted, gin.H{
		"message": "job queued",
		"job_id":  job.ID,
	})
}
