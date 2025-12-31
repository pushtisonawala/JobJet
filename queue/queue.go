package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
	"jobqueue/models"
)

var ctx = context.Background()

type RedisQueue struct {
	client *redis.Client
}

func NewRedisQueue() *RedisQueue {
	rdx := redis.NewClient(&redis.Options{
    Addr: "redis:6379", 
	})

	pong, err := rdx.Ping(ctx).Result()
	if err != nil {
		panic("REDIS CONNECTION FAILED: " + err.Error())
	}
	fmt.Println("✅ Redis connected:", pong)

	return &RedisQueue{
		client: rdx,
	}
}
func (q *RedisQueue) Push(job models.Jobs) error {
	// 1️⃣ Log the job you're about to push
	fmt.Println("📤 Attempting to push job to Redis:", job.ID)

	// 2️⃣ Convert job struct to JSON
	data, err := json.Marshal(job)
	if err != nil {
		fmt.Println("❌ Failed to marshal job:", err)
		return err
	}

	// 3️⃣ Push job to Redis list
	err = q.client.LPush(ctx, "job_queue", data).Err()
	if err != nil {
		fmt.Println("❌ Failed to push job to Redis:", err)
		return err
	}

	// 4️⃣ Success log
	fmt.Println("✅ Job pushed successfully:", string(data))
	return nil
}
