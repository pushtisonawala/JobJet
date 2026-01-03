package queue

import (
	"context"
	"encoding/json"
	"jobqueue/logger"
	"jobqueue/metrics"
	"jobqueue/models"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue() *RedisQueue {
	rdx := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	pong, err := rdx.Ping(ctx).Result()
	if err != nil {
		panic("REDIS CONNECTION FAILED: " + err.Error())
	}
	logger.Log.Info("✅ Redis connected", "pong", pong)

	return &RedisQueue{client: rdx}
}

func (q *RedisQueue) Push(job models.Jobs) error {
	logger.Log.Info("Attempting to push job to Redis:", "jobID", job.ID)

	data, err := json.Marshal(job)
	if err != nil {
		logger.Log.Error("Failed to marshal job:", "error", err)
		return err
	}

	if err := q.client.LPush(ctx, "job_queue", data).Err(); err != nil {
		logger.Log.Error("Failed to push job to Redis:", "error", err)
		return err
	}

	logger.Log.Info("Job pushed successfully", "jobData", string(data))
	metrics.JobsRecieved.Inc()
	return nil
}

// Client returns the underlying Redis client
func (q *RedisQueue) Client() *redis.Client {
	return q.client
}
