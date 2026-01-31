package worker

import (
	"context"
	"jobjet/metrics"

	"github.com/redis/go-redis/v9"
)

// UpdateQueueMetrics updates all queue length gauge metrics from Redis
func UpdateQueueMetrics(ctx context.Context, rdb *redis.Client) {
	jobLen, _ := rdb.LLen(ctx, "job_queue").Result()
	procLen, _ := rdb.LLen(ctx, "processing_queue").Result()
	dlqLen, _ := rdb.LLen(ctx, "dlq_queue").Result()
	retryLen, _ := rdb.ZCard(ctx, "retry_zset").Result()

	metrics.JobQueueLength.Set(float64(jobLen))
	metrics.ProcessingQueueLength.Set(float64(procLen))
	metrics.DLQLength.Set(float64(dlqLen))
	metrics.RetryQueueLength.Set(float64(retryLen))
}
