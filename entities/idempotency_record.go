package entities

import "time"

type IdempotencyRecord struct {
	Key        string    `json:"key" gorm:"primaryKey;size:255"`
	Response   string    `json:"response" gorm:"type:text;not null"`
	StatusCode int       `json:"statusCode" gorm:"column:status_code;not null"`
	CreatedAt  time.Time `json:"createdAt" gorm:"column:created_at"`
}
