package queue

import (
	"context"
	"encoding/json"
"jobqueue/logger"
	"jobqueue/models"

	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue() *RedisQueue {
	rdx := redis.NewClient(&redis.Options{
		Addr: "localhost:6379", // Use localhost for local development
	})

	pong, err := rdx.Ping(ctx).Result()
	if err != nil {
		panic("REDIS CONNECTION FAILED: " + err.Error())
	}
	logger.Log.Info("✅ Redis connected", "pong", pong)

	return &RedisQueue{
		client: rdx,
	}
}
func (q *RedisQueue) Push(job models.Jobs) error {
	// 1️⃣ Log the job you're about to push
	logger.Log.Info("Attempting to push job to Redis:","jobID", job.ID)

	// 2️⃣ Convert job struct to JSON
	data, err := json.Marshal(job)
	if err != nil {
		logger.Log.Error(" Failed to marshal job:","error", err)
		return err
	}

	// 3️⃣ Push job to Redis list
	err = q.client.LPush(ctx, "job_queue", data).Err()
	if err != nil {
		logger.Log.Error("Failed to push job to Redis:","error", err)
		return err
	}

	logger.Log.Info("Job pushed successfully", "jobData", string(data))
	return nil
}
