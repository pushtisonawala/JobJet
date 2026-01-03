package main

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// check_queues.go - Check all queue lengths
func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	jobLen, _ := rdb.LLen(ctx, "job_queue").Result()
	procLen, _ := rdb.LLen(ctx, "processing_queue").Result()
	dlqLen, _ := rdb.LLen(ctx, "dlq_queue").Result()
	oldDLQLen, _ := rdb.LLen(ctx, "dlq").Result()
	retryLen, _ := rdb.ZCard(ctx, "retry_zset").Result()

	fmt.Println("=== Queue Lengths ===")
	fmt.Printf("job_queue:         %d\n", jobLen)
	fmt.Printf("processing_queue:  %d\n", procLen)
	fmt.Printf("dlq_queue:         %d\n", dlqLen)
	fmt.Printf("dlq (old):         %d\n", oldDLQLen)
	fmt.Printf("retry_zset:        %d\n", retryLen)
	fmt.Println()

	// Sample a few jobs from dlq_queue if any
	if dlqLen > 0 {
		fmt.Println("=== Sample DLQ Jobs (first 3) ===")
		jobs, err := rdb.LRange(ctx, "dlq_queue", 0, 2).Result()
		if err != nil {
			log.Fatalf("Failed to get DLQ jobs: %v", err)
		}
		for i, job := range jobs {
			fmt.Printf("Job %d: %s\n", i+1, job)
		}
	}

	// Sample processing queue
	if procLen > 0 {
		fmt.Println("\n=== Processing Queue Jobs ===")
		jobs, err := rdb.LRange(ctx, "processing_queue", 0, -1).Result()
		if err != nil {
			log.Fatalf("Failed to get processing jobs: %v", err)
		}
		for i, job := range jobs {
			fmt.Printf("Job %d: %s\n", i+1, job)
		}
	}
}
