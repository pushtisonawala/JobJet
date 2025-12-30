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

func main() {
	fmt.Println("👷 Worker started")

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	for {
		// Fetch job from job_queue and move to processing_queue atomically
		result, err := rdb.BRPopLPush(ctx, "job_queue", "processing_queue", 0).Result()
		if err != nil {
			fmt.Println("Error fetching job:", err)
			continue
		}

		var job models.Jobs
		json.Unmarshal([]byte(result), &job)

		fmt.Println("⚙️ Processing job:", job.ID)

		// Simulate job failure for demonstration
		isFailed := job.Type == "fail"

		if isFailed {
			if job.RetryCount < job.MaxRetry {
				job.RetryCount++
				fmt.Printf("🔄 Retry #%d for job: %s\n", job.RetryCount, job.ID)

				delay := time.Duration(math.Pow(2, float64(job.RetryCount))) * time.Second
				time.Sleep(delay)

				// Push back to job_queue for retry
				data, _ := json.Marshal(job)
				rdb.LPush(ctx, "job_queue", data)
			} else {
				fmt.Println("☠️ Max retries reached, moving job to DLQ:", job.ID)
				data, _ := json.Marshal(job)
				rdb.LPush(ctx, "dlq", data) // Dead Letter Queue
			}

			// Remove from processing queue
			rdb.LRem(ctx, "processing_queue", 1, result)
			continue
		}

		// Simulate processing time
		time.Sleep(2 * time.Second)

		// Job completed successfully
		rdb.LRem(ctx, "processing_queue", 1, result)
		fmt.Println("✅ Job completed:", job.ID)
	}
}
