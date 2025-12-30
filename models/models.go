package models
import "time"
type Jobs struct{
	ID string `json:"id"`
	Type string `json:"type"`
	Payload map[string]interface{} `json:"payload"`
		Status string `json:"status"`
		RetryCount int `json:"retry_count"`
        MaxRetry   int `json:"max_retry"`

				CreatedAt time.Time `json:"createdat"`



}