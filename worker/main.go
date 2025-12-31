package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
	"jobqueue/models"
)

var ctx = context.Background()

const maxWorkers = 5

func main() {
	fmt.Println("👷 Worker started with concurrency")

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Semaphore to limit concurrent jobs
	semaphore := make(chan struct{}, maxWorkers)

	for {
		// Fetch job atomically
		result, err := rdb.BRPopLPush(ctx, "job_queue", "processing_queue", 0).Result()
		if err != nil {
			fmt.Println("Error fetching job:", err)
			continue
		}

		// Acquire worker slot
		semaphore <- struct{}{}

		// Process job concurrently
		go func(jobData string) {
			defer func() {
				<-semaphore // Release slot
			}()

			processJob(rdb, jobData)
		}(result)
	}
}

func processJob(rdb *redis.Client, result string) {
	var job models.Jobs
	json.Unmarshal([]byte(result), &job)

	fmt.Println("⚙️ Processing job:", job.ID)

	// Simulate failure
	if job.Type == "fail" {
		if job.RetryCount < job.MaxRetry {
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
		} else {
			fmt.Println("☠️ Max retries reached → DLQ:", job.ID)
			data, _ := json.Marshal(job)
			rdb.LPush(ctx, "dlq", data)
		}

		// Acknowledge
		rdb.LRem(ctx, "processing_queue", 1, result)
		return
	}

	// Simulate work
	time.Sleep(2 * time.Second)

	// Success
	rdb.LRem(ctx, "processing_queue", 1, result)
	fmt.Println("✅ Job completed:", job.ID)
}
