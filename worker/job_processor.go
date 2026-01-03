package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"jobqueue/db"
	"jobqueue/logger"
	"jobqueue/metrics"
	"jobqueue/models"
	"jobqueue/utils"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
)

func ProcessJob(ctx context.Context, rdb *redis.Client, jobData string) {
	var job models.Jobs
	err := json.Unmarshal([]byte(jobData), &job)
	if err != nil {
		logger.Log.Error(" Invalid job JSON:","error",err)
		rdb.LRem(ctx, "processing_queue", 1, jobData)
		return
	}

	fmt.Println(" Processing job:", job.ID)
	coll := db.Client.Database("jobqueue").Collection("jobs")

	if job.Type == "email" {
		err := utils.SendEmail(
			job.Payload.To,
			job.Payload.Subject,
			job.Payload.Body,
		)
		if err != nil {
			logger.Log.Error("Email failed:","error", err)
		} else {
			logger.Log.Info("Email sent successfully", "to", job.Payload.To)
		}

		_, err = coll.UpdateOne(ctx,
			bson.M{"id": job.ID},
			bson.M{"$set": bson.M{"status": "completed", "updated_at": time.Now()}},
		)
		if err != nil {
			logger.Log.Error(" Mongo update failed:","error", err)
		}

		rdb.LRem(ctx, "processing_queue", 1, jobData)
		fmt.Println("Job completed:", job.ID)
		metrics.JobsProcessed.Inc()
		metrics.ProcessingQueueLength.Dec()
		return
	}

	if job.Type == "fail" {
		if job.RetryCount < job.MaxRetry {
			err = utils.SendEmail(
				job.Payload.To,
				fmt.Sprintf("Retrying Job %s", job.ID),
				fmt.Sprintf("Job %s failed. Retry attempt #%d", job.ID, job.RetryCount+1),
			)
			if err != nil {
				fmt.Println("Email failed:", err)
			}
			metrics.JobsFailed.Inc()
			metrics.JobsRetried.Inc()

			job.RetryCount++
			retryAt := time.Now().Add(time.Duration(math.Pow(2, float64(job.RetryCount))) * time.Second).Unix()
			data, _ := json.Marshal(job)
			rdb.ZAdd(ctx, "retry_zset", redis.Z{Score: float64(retryAt), Member: data})
			metrics.RetryQueueLength.Inc()


			_, err = coll.UpdateOne(ctx,
				bson.M{"id": job.ID},
				bson.M{"$set": bson.M{"status": "retrying", "retry_count": job.RetryCount, "updated_at": time.Now()}},
			)
			if err != nil {
				logger.Log.Error("Mongo update failed:","error", err)
			}
		} else {
			logger.Log.Info("☠️ Max retries reached → DLQ:", "jobID", job.ID)
			data, _ := json.Marshal(job)
			rdb.LPush(ctx, "dlq", data)
			_, err = coll.UpdateOne(ctx,
				bson.M{"id": job.ID},
				bson.M{"$set": bson.M{"status": "failed", "updated_at": time.Now()}},
			)
			metrics.DLQLength.Inc()
			metrics.DLQLength.Inc()
			if err != nil {
				logger.Log.Error(" update failed:","error", err)
			}
		}
		rdb.LRem(ctx, "processing_queue", 1, jobData)
		return
	}

	logger.Log.Info("Unknown job type:", "type", job.Type)
	rdb.LRem(ctx, "processing_queue", 1, jobData)
}
