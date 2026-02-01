package worker

import (
	"context"
	"encoding/json"
	"jobjet/db"
	"jobjet/logger"
	"jobjet/metrics"
	"jobjet/models"
	"jobjet/utils"
	"math"
	"time"

	"go.opentelemetry.io/otel/codes"

	gootel "go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
)

func ProcessJob(ctx context.Context, rdb *redis.Client, jobData string) error {
	// 🆕 Add tracing span
	tracer := gootel.Tracer("jobqueue-worker")
	ctx, span := tracer.Start(ctx, "ProcessJob")
	defer span.End()

	var job models.Jobs

	logger.Log.Info("ProcessJob ENTERED", "raw", jobData)

	defer func() {
		if r := recover(); r != nil {
			logger.Log.Error("Panic recovered in ProcessJob", "panic", r, "job", jobData)
			rdb.LPush(ctx, "dlq_queue", jobData)
			metrics.JobsFailed.Inc()
			metrics.JobsDLQ.Inc()
		}
		rdb.LRem(ctx, "processing_queue", 1, jobData)
		UpdateQueueMetrics(ctx, rdb)
	}()

	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		logger.Log.Error("Invalid job JSON", "error", err)

		rdb.LPush(ctx, "dlq_queue", jobData)

		metrics.JobsFailed.Inc()
		metrics.JobsDLQ.Inc()
		return err
	}

	// 🆕 Add job metadata to span
	span.SetAttributes(
		attribute.String("job.id", job.ID),
		attribute.String("job.type", job.Type),
	)

	// Add production-grade attributes
	span.SetAttributes(
		attribute.String("job.status", job.Status),
		attribute.Int("job.retry_count", job.RetryCount),
		attribute.Int("job.max_retry", job.MaxRetry),
	)
	if !job.CreatedAt.IsZero() {
		span.SetAttributes(attribute.String("job.created_at", job.CreatedAt.Format(time.RFC3339)))
		span.SetAttributes(attribute.Int64("job.queue_wait_ms", time.Since(job.CreatedAt).Milliseconds()))
	}

	if job.Type == "sendEmail" {
		span.SetAttributes(
			attribute.String("email.recipient", job.Payload.To),
			attribute.String("email.subject", job.Payload.Subject),
		)
	}

	// Execution timing
	workStart := time.Now()
	_, workSpan := tracer.Start(ctx, "ExecuteHandler")
	workSpan.SetAttributes(attribute.String("handler", job.Type))

	logger.Log.Info("Processing job", "jobID", job.ID)
	coll := db.Client.Database("jobqueue").Collection("jobs")

	var handlerErr error
	if job.Type == "email" {
		handlerErr = utils.SendEmail(job.Payload.To, job.Payload.Subject, job.Payload.Body)
		if handlerErr != nil {
			logger.Log.Error("Email failed", "error", handlerErr)
			// Move to DLQ on email failure
			data, _ := json.Marshal(job)
			rdb.LPush(ctx, "dlq_queue", data)
			coll.UpdateOne(ctx,
				bson.M{"id": job.ID},
				bson.M{"$set": bson.M{"status": "dlq", "error": handlerErr.Error(), "updated_at": time.Now()}},
			)
			metrics.JobsFailed.Inc()
			metrics.JobsDLQ.Inc()
		}
		if handlerErr == nil {
			coll.UpdateOne(ctx,
				bson.M{"id": job.ID},
				bson.M{"$set": bson.M{"status": "completed", "updated_at": time.Now()}},
			)
			metrics.JobsProcessed.Inc()
			logger.Log.Info("Job completed", "jobID", job.ID)
		}
	}
	workDuration := time.Since(workStart).Milliseconds()
	workSpan.SetAttributes(attribute.Int64("handler.execution_ms", workDuration))
	workSpan.End()

	var totalTime int64
	if !job.CreatedAt.IsZero() {
		totalTime = time.Since(job.CreatedAt).Milliseconds()
	}

	if handlerErr == nil {
		span.SetStatus(codes.Ok, "Job completed successfully")
		span.SetAttributes(
			attribute.Bool("job.succeeded", true),
			attribute.Int64("job.total_duration_ms", totalTime),
		)
		return nil
	} else if handlerErr != nil {
		span.RecordError(handlerErr)
		span.SetStatus(codes.Error, "Job processing failed")
		span.SetAttributes(attribute.Bool("job.failed", true))
		return handlerErr
	}

	// ----------------- FAIL JOB -----------------
	if job.Type == "fail" {
		if job.RetryCount < job.MaxRetry {
			job.RetryCount++

			retryAt := time.Now().
				Add(time.Duration(math.Pow(2, float64(job.RetryCount))) * time.Second).
				Unix()

			data, _ := json.Marshal(job)

			rdb.ZAdd(ctx, "retry_zset", redis.Z{
				Score:  float64(retryAt),
				Member: data,
			})

			coll.UpdateOne(ctx,
				bson.M{"id": job.ID},
				bson.M{"$set": bson.M{
					"status":      "retrying",
					"retry_count": job.RetryCount,
					"updated_at":  time.Now(),
				}},
			)

			metrics.JobsRetried.Inc()
			logger.Log.Info("Job scheduled for retry", "jobID", job.ID)
			return nil
		}

		data, _ := json.Marshal(job)
		rdb.LPush(ctx, "dlq_queue", data)

		coll.UpdateOne(ctx,
			bson.M{"id": job.ID},
			bson.M{"$set": bson.M{"status": "dlq", "updated_at": time.Now()}},
		)

		metrics.JobsFailed.Inc()
		metrics.JobsDLQ.Inc()
		logger.Log.Info("Job moved to DLQ", "jobID", job.ID)
		return nil
	}

	logger.Log.Error("Unknown job type", "type", job.Type)

	data, _ := json.Marshal(job)
	rdb.LPush(ctx, "dlq_queue", data)

	coll.UpdateOne(ctx,
		bson.M{"id": job.ID},
		bson.M{"$set": bson.M{
			"status":     "dlq",
			"error":      "unknown job type",
			"updated_at": time.Now(),
		}},
	)

	metrics.JobsFailed.Inc()
	metrics.JobsDLQ.Inc()
	return nil
}
