package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"

	"jobqueue/db"
	"jobqueue/models"
	"jobqueue/utils"
)

var ctx context.Context
var cancel context.CancelFunc
var wg sync.WaitGroup
var shuttingDown bool

const maxWorkers = 5

func main() {
	fmt.Println("👷 Worker started with concurrency")
	db.ConnectMongo()
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx, cancel = context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Println("Shutdown signal received")
		shuttingDown = true
		cancel()
		restoreProcessingJobs(rdb)
		wg.Wait()
		fmt.Println("Worker exited cleanly")
		os.Exit(0)
	}()

	semaphore := make(chan struct{}, maxWorkers)
	go retryScheduler(rdb)

	for !shuttingDown {
		result, err := rdb.BRPopLPush(ctx, "job_queue", "processing_queue", 0).Result()
		if err != nil {
			if err == context.Canceled {
				continue
			}
			fmt.Println("❌ Error fetching job:", err)
			continue
		}

		semaphore <- struct{}{}
		wg.Add(1)

		go func(jobData string) {
			defer func() {
				<-semaphore
				wg.Done()
			}()

			processJob(rdb, jobData)
		}(result)
	}
}

func retryScheduler(rdb *redis.Client) {
	fmt.Println("⏰ Retry scheduler started")
	for !shuttingDown {
		now := float64(time.Now().Unix())
		jobs, err := rdb.ZRangeByScore(ctx, "retry_zset", &redis.ZRangeBy{
			Min: "0",
			Max: fmt.Sprintf("%f", now),
		}).Result()
		if err != nil {
			fmt.Println("Retry scheduler error:", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, j := range jobs {
			rdb.LPush(ctx, "job_queue", j)
			rdb.ZRem(ctx, "retry_zset", j)
			fmt.Println("Job moved from retry_zset to job_queue:", j)
		}

		time.Sleep(1 * time.Second)
	}
}

func restoreProcessingJobs(rdb *redis.Client) {
	fmt.Println("♻️ Restoring in-flight jobs")
	for {
		job, err := rdb.RPopLPush(ctx, "processing_queue", "job_queue").Result()
		if err == redis.Nil {
			break
		}
		if err != nil {
			fmt.Println("Restore error:", err)
			break
		}
		fmt.Println("↩️ Restored job:", job)
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