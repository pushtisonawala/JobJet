package worker

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func RestoreProcessingJobs(ctx context.Context, rdb *redis.Client) {
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
