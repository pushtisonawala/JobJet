package models
import "time"
type Jobs struct{
	ID string `json:"id"`
	Type string `json:"type"`
	Payload map[string]interface{} `json:"payload"`
		Status string `json:"status"`
				CreatedAt time.Time `json:"createdat"`



}