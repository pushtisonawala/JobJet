package models
import "time"
type EmailPayload struct{
	To string `json:"to"`
	Subject string `json:"subject"`
	Body string `json:"body"`

}
type Jobs struct{
	ID string `json:"id"`
	Type string `json:"type"`
		Status string `json:"status"`
		RetryCount int `json:"retry_count"`
        MaxRetry   int `json:"max_retry"`

				CreatedAt time.Time `json:"createdat"`
				Payload EmailPayload `json:"payload"`



}