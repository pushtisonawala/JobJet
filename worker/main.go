package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"

	"jobqueue/db"
	"jobqueue/models"
	"jobqueue/utils"
)

var ctx = context.Background()

const maxWorkers = 5

func main() {
	fmt.Println("👷 Worker started with concurrency")
	db.ConnectMongo()
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	semaphore := make(chan struct{}, maxWorkers)

	for {
		result, err := rdb.BRPopLPush(ctx, "job_queue", "processing_queue", 0).Result()
		if err != nil {
			fmt.Println("❌ Error fetching job:", err)
			continue
		}

		semaphore <- struct{}{}

		go func(jobData string) {
			defer func() {
				<-semaphore
			}()

			processJob(rdb, jobData)
		}(result)
	}
}
func processJob(rdb *redis.Client, result string) {
	var job models.Jobs
	err := json.Unmarshal([]byte(result), &job)
	if err != nil {
		fmt.Println("❌ Invalid job JSON:", err)
		rdb.LRem(ctx, "processing_queue", 1, result)
		return
	}

	fmt.Println("⚙️ Processing job:", job.ID)

	coll := db.Client.Database("jobqueue").Collection("jobs")

	// HANDLE EMAIL JOBS
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

		// Mark as completed in Mongo
		_, err = coll.UpdateOne(ctx,
			bson.M{"id": job.ID},
			bson.M{
				"$set": bson.M{
					"status":     "completed",
					"updated_at": time.Now(),
				},
			},
		)
		if err != nil {
			fmt.Println("❌ Mongo update failed:", err)
		}

		rdb.LRem(ctx, "processing_queue", 1, result)
		fmt.Println("✅ Job completed:", job.ID)
		return
	}

	// FAILURE CASE (retry logic)
	if job.Type == "fail" {

		if job.RetryCount < job.MaxRetry {

			// 🔔 Send email for retry
			err = utils.SendEmail(
				job.Payload.To, // use job payload if you want real email
				fmt.Sprintf("Retrying Job %s", job.ID),
				fmt.Sprintf("Job %s failed. Retry attempt #%d", job.ID, job.RetryCount+1),
			)
			if err != nil {
				fmt.Println("❌ Email failed:", err)
			}

			job.RetryCount++
			fmt.Printf("🔄 Retry #%d for job: %s\n", job.RetryCount, job.ID)

			retryAt := time.Now().
				Add(time.Duration(math.Pow(2, float64(job.RetryCount))) * time.Second).
				Unix()

			data, _ := json.Marshal(job)
			rdb.ZAdd(ctx, "retry_zset", redis.Z{
				Score:  float64(retryAt),
				Member: data,
			})

			_, err = coll.UpdateOne(ctx,
				bson.M{"id": job.ID},
				bson.M{
					"$set": bson.M{
						"status":      "retrying",
						"retry_count": job.RetryCount,
						"updated_at":  time.Now(),
					},
				},
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
				bson.M{
					"$set": bson.M{
						"status":     "failed",
						"updated_at": time.Now(),
					},
				},
			)
			if err != nil {
				fmt.Println("❌ Mongo update failed:", err)
			}
		}

		rdb.LRem(ctx, "processing_queue", 1, result)
		return
	}

	fmt.Println("⚠️ Unknown job type:", job.Type)
	rdb.LRem(ctx, "processing_queue", 1, result)
}
