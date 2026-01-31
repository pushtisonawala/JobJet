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

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
)

func ProcessJob(ctx context.Context, rdb *redis.Client, jobData string) error {
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

	logger.Log.Info("Processing job", "jobID", job.ID)
	coll := db.Client.Database("jobqueue").Collection("jobs")

	if job.Type == "email" {
		if err := utils.SendEmail(job.Payload.To, job.Payload.Subject, job.Payload.Body); err != nil {
			logger.Log.Error("Email failed", "error", err)

			// Move to DLQ on email failure
			data, _ := json.Marshal(job)
			rdb.LPush(ctx, "dlq_queue", data)

			coll.UpdateOne(ctx,
				bson.M{"id": job.ID},
				bson.M{"$set": bson.M{"status": "dlq", "error": err.Error(), "updated_at": time.Now()}},
			)

			metrics.JobsFailed.Inc()
			metrics.JobsDLQ.Inc()
			return err
		}

		coll.UpdateOne(ctx,
			bson.M{"id": job.ID},
			bson.M{"$set": bson.M{"status": "completed", "updated_at": time.Now()}},
		)
		metrics.JobsProcessed.Inc()
		logger.Log.Info("Job completed", "jobID", job.ID)
		return nil
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
