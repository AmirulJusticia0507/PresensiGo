package model

import (
	"time"

	"github.com/google/uuid"
)

type FraudAttempt struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	UserName    string    `json:"user_name,omitempty"`
	Reason      string    `json:"reason"`
	Platform    string    `json:"platform"`
	Latitude    *float64  `json:"latitude,omitempty"`
	Longitude   *float64  `json:"longitude,omitempty"`
	Accuracy    *float64  `json:"accuracy,omitempty"`
	AttemptedAt time.Time `json:"attempted_at"`
}

type FraudAttemptRequest struct {
	Reason    string   `json:"reason" validate:"required,max=100"`
	Platform  string   `json:"platform" validate:"required,oneof=android ios"`
	Latitude  *float64 `json:"latitude,omitempty" validate:"omitempty,gte=-90,lte=90"`
	Longitude *float64 `json:"longitude,omitempty" validate:"omitempty,gte=-180,lte=180"`
	Accuracy  *float64 `json:"accuracy,omitempty" validate:"omitempty,gte=0,lte=100000"`
}

type FraudAlert struct {
	UserID       uuid.UUID `json:"user_id"`
	UserName     string    `json:"user_name"`
	UserEmail    string    `json:"user_email"`
	AttemptCount int       `json:"attempt_count"`
	LastReason   string    `json:"last_reason"`
	LastAttempt  time.Time `json:"last_attempt"`
}
