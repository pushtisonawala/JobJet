package worker

import (
	"context"
	"fmt"
	"jobqueue/logger"
	"time"

	"github.com/redis/go-redis/v9"
)

func RetryScheduler(ctx context.Context, rdb *redis.Client, shuttingDown *bool) {
	logger.Log.Info("Retry scheduler started")
	for !*shuttingDown {
		now := float64(time.Now().Unix())
		jobs, err := rdb.ZRangeByScore(ctx, "retry_zset", &redis.ZRangeBy{
			Min: "0",
			Max: fmt.Sprintf("%f", now),
		}).Result()
		if err != nil {
			logger.Log.Error("Retry scheduler error:", "error", err)
			time.Sleep(2 * time.Second)
			continue
		}

		for _, j := range jobs {
			rdb.LPush(ctx, "job_queue", j)
			rdb.ZRem(ctx, "retry_zset", j)
			logger.Log.Info("Job moved from retry_zset to job_queue", "job", j)
		}

		time.Sleep(1 * time.Second)
	}
}
