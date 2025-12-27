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
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}

	err = q.client.LPush(ctx, "job_queue", data).Err()
	if err != nil {
		fmt.Println("❌ Failed to push job:", err)
		return err
	}

	fmt.Println("✅ Job pushed to Redis:", string(data))
	return nil
}
