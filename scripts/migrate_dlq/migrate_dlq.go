package main

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// migrate_dlq.go - One-time migration script to move jobs from "dlq" to "dlq_queue"
func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Check both queues
	oldDLQLen, err := rdb.LLen(ctx, "dlq").Result()
	if err != nil {
		log.Fatalf("Failed to check old DLQ: %v", err)
	}

	newDLQLen, err := rdb.LLen(ctx, "dlq_queue").Result()
	if err != nil {
		log.Fatalf("Failed to check new DLQ: %v", err)
	}

	fmt.Printf("Old 'dlq' queue length: %d\n", oldDLQLen)
	fmt.Printf("New 'dlq_queue' length: %d\n", newDLQLen)

	if oldDLQLen == 0 {
		fmt.Println("No jobs to migrate from old 'dlq' queue")
		return
	}

	fmt.Printf("Migrating %d jobs from 'dlq' to 'dlq_queue'...\n", oldDLQLen)

	// Move all jobs from dlq to dlq_queue
	migrated := 0
	for {
		job, err := rdb.RPop(ctx, "dlq").Result()
		if err != nil {
			break // Queue is empty
		}

		rdb.LPush(ctx, "dlq_queue", job)
		migrated++
	}

	fmt.Printf("Successfully migrated %d jobs\n", migrated)

	// Verify
	finalOldLen, _ := rdb.LLen(ctx, "dlq").Result()
	finalNewLen, _ := rdb.LLen(ctx, "dlq_queue").Result()

	fmt.Printf("\nAfter migration:\n")
	fmt.Printf("Old 'dlq' queue length: %d\n", finalOldLen)
	fmt.Printf("New 'dlq_queue' length: %d\n", finalNewLen)
}
