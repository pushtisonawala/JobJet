package models

import "time"

type Job struct {
	ID         string    `bson:"_id"`
	Type       string    `bson:"type"`
	Status     string    `bson:"status"`
	RetryCount int       `bson:"retryCount"`
	MaxRetry   int       `bson:"maxRetry"`
	CreatedAt  time.Time `bson:"createdAt"`
	UpdatedAt  time.Time `bson:"updatedAt"`
}
