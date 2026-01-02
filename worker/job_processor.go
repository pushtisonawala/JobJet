package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"jobqueue/db"
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
		fmt.Println("❌ Invalid job JSON:", err)
		rdb.LRem(ctx, "processing_queue", 1, jobData)
		return
	}

	fmt.Println("⚙️ Processing job:", job.ID)
	coll := db.Client.Database("jobqueue").Collection("jobs")

	if job.Type == "email" {
		err := utils.SendEmail(
			job.Payload.To,
			job.Payload.Subject,
			job.Payload.Body,
		)
		if err != nil {
			fmt.Println("❌ Email failed:", err)
		} else {
			fmt.Println("📧 Email sent successfully to", job.Payload.To)
		}

		_, err = coll.UpdateOne(ctx,
			bson.M{"id": job.ID},
			bson.M{"$set": bson.M{"status": "completed", "updated_at": time.Now()}},
		)
		if err != nil {
			fmt.Println("❌ Mongo update failed:", err)
		}

		rdb.LRem(ctx, "processing_queue", 1, jobData)
		fmt.Println("✅ Job completed:", job.ID)
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
				fmt.Println("❌ Email failed:", err)
			}

			job.RetryCount++
			retryAt := time.Now().Add(time.Duration(math.Pow(2, float64(job.RetryCount))) * time.Second).Unix()
			data, _ := json.Marshal(job)
			rdb.ZAdd(ctx, "retry_zset", redis.Z{Score: float64(retryAt), Member: data})

			_, err = coll.UpdateOne(ctx,
				bson.M{"id": job.ID},
				bson.M{"$set": bson.M{"status": "retrying", "retry_count": job.RetryCount, "updated_at": time.Now()}},
			)
			if err != nil {
				fmt.Println("❌ Mongo update failed:", err)
			}
		} else {
			fmt.Println("☠️ Max retries reached → DLQ:", job.ID)
			data, _ := json.Marshal(job)
			rdb.LPush(ctx, "dlq", data)
			_, err = coll.UpdateOne(ctx,
				bson.M{"id": job.ID},
				bson.M{"$set": bson.M{"status": "failed", "updated_at": time.Now()}},
			)
			if err != nil {
				fmt.Println("❌ Mongo update failed:", err)
			}
		}
		rdb.LRem(ctx, "processing_queue", 1, jobData)
		return
	}

	fmt.Println("⚠️ Unknown job type:", job.Type)
	rdb.LRem(ctx, "processing_queue", 1, jobData)
}
