package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Job struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	RetryCount int       `json:"retry_count"`
	MaxRetry   int       `json:"max_retry"`
	CreatedAt  time.Time `json:"createdat"`
	Payload    Payload   `json:"payload"`
}

type Payload struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

func main() {
	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})

	// Add email job (should succeed - metrics.JobsProcessed.Inc())
	emailJob := Job{
		ID: uuid.NewString(), Type: "email", Status: "pending",
		RetryCount: 0, MaxRetry: 3, CreatedAt: time.Now(),
		Payload: Payload{To: "test@example.com", Subject: "SUCCESS TEST", Body: "This should work"},
	}
	data, _ := json.Marshal(emailJob)
	rdb.LPush(ctx, "job_queue", data)
	fmt.Println("Added email job (should increment jobs_processed_total)")

	// Add fail job (should retry 3x then DLQ - metrics.JobsRetried.Inc() x3, then JobsDLQ.Inc())
	failJob := Job{
		ID: uuid.NewString(), Type: "fail", Status: "pending",
		RetryCount: 0, MaxRetry: 3, CreatedAt: time.Now(),
		Payload: Payload{To: "test@example.com", Subject: "FAIL TEST", Body: "This will fail"},
	}
	data, _ = json.Marshal(failJob)
	rdb.LPush(ctx, "job_queue", data)
	fmt.Println("Added fail job (should retry 3x then DLQ)")

	fmt.Println("\n✅ Wait ~20 seconds, then check http://localhost:2113/metrics")
	fmt.Println("Expected: jobs_processed_total=1, jobs_retried_total=3, jobs_dlq_total=1, jobs_failed_total=1")
}
