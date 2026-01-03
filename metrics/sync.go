package metrics

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func SyncQueueLengths(ctx context.Context, rdb *redis.Client) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			jobQ, _ := rdb.LLen(ctx, "job_queue").Result()
			procQ, _ := rdb.LLen(ctx, "processing_queue").Result()
			dlq, _ := rdb.LLen(ctx, "dlq").Result()
			retryQ, _ := rdb.ZCard(ctx, "retry_zset").Result()

			JobQueueLength.Set(float64(jobQ))
			ProcessingQueueLength.Set(float64(procQ))
			RetryQueueLength.Set(float64(retryQ))
			DLQLength.Set(float64(dlq))

		case <-ctx.Done():
			return
		}
	}
}
