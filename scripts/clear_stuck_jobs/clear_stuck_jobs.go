package main

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// clear_stuck_jobs.go - Move stuck jobs from processing_queue back to job_queue
func main() {
	ctx := context.Background()

	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	// Check processing_queue
	procLen, err := rdb.LLen(ctx, "processing_queue").Result()
	if err != nil {
		log.Fatalf("Failed to check processing_queue: %v", err)
	}

	fmt.Printf("Processing queue length: %d\n", procLen)

	if procLen == 0 {
		fmt.Println("No stuck jobs to clear")
		return
	}

	fmt.Printf("Moving %d stuck jobs from processing_queue back to job_queue...\n", procLen)

	// Move all jobs from processing_queue back to job_queue
	moved := 0
	for {
		job, err := rdb.RPop(ctx, "processing_queue").Result()
		if err != nil {
			break // Queue is empty
		}

		rdb.LPush(ctx, "job_queue", job)
		moved++
	}

	fmt.Printf("Successfully moved %d stuck jobs back to job_queue\n", moved)

	// Verify
	finalProcLen, _ := rdb.LLen(ctx, "processing_queue").Result()
	finalJobLen, _ := rdb.LLen(ctx, "job_queue").Result()

	fmt.Printf("\nAfter cleanup:\n")
	fmt.Printf("processing_queue length: %d\n", finalProcLen)
	fmt.Printf("job_queue length: %d\n", finalJobLen)
}
