package worker

import (
	"context"
	"fmt"
	"jobqueue/logger"
	"time"

	"github.com/redis/go-redis/v9"
)

func RetryScheduler(ctx context.Context, rdb *redis.Client) {
	logger.Log.Info("Retry scheduler started")

	for {
		select {
		case <-ctx.Done():
			logger.Log.Info("Retry scheduler stopped")
			return
		default:
		}

		now := float64(time.Now().Unix())

		jobs, err := rdb.ZRangeByScore(ctx, "retry_zset", &redis.ZRangeBy{
			Min: "0",
			Max: fmt.Sprintf("%f", now),
		}).Result()

		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}

		for _, j := range jobs {
			rdb.LPush(ctx, "job_queue", j)
			rdb.ZRem(ctx, "retry_zset", j)
			// ❌ DO NOT increment JobsRetried here
		}

		UpdateQueueMetrics(ctx, rdb)
		time.Sleep(1 * time.Second)
	}
}
