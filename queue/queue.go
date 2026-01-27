package queue

import (
	"context"
	"encoding/json"
	"jobjet/logger"
	"jobjet/metrics"
	"jobjet/models"
	"os"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue() *RedisQueue {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "127.0.0.1:6379"
	}
	rdx := redis.NewClient(&redis.Options{
		Addr: redisAddr,
		DB:   0,
	})

	pong, err := rdx.Ping(ctx).Result()
	if err != nil {
		panic("REDIS CONNECTION FAILED: " + err.Error())
	}
	logger.Log.Info("✅ Redis connected", "pong", pong, "addr", redisAddr)

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

func (q *RedisQueue) Client() *redis.Client {
	return q.client
}
