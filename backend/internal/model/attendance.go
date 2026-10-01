package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type Attendance struct {
	ID                     uuid.UUID  `json:"id" db:"id"`
	UserID                 uuid.UUID  `json:"user_id" db:"user_id"`
	LocationID             uuid.UUID  `json:"location_id" db:"location_id"`
	CheckInTime            *time.Time `json:"check_in_time,omitempty" db:"check_in_time"`
	CheckOutTime           *time.Time `json:"check_out_time,omitempty" db:"check_out_time"`
	CheckInLocation        []float64  `json:"check_in_location,omitempty" db:"check_in_location"`
	CheckOutLocation       []float64  `json:"check_out_location,omitempty" db:"check_out_location"`
	SelfieURL              *string    `json:"selfie_url,omitempty" db:"selfie_url"`
	Status                 string     `json:"status" db:"status"`
	IsLate                 bool       `json:"is_late" db:"is_late"`
	Notes                  *string    `json:"notes,omitempty" db:"notes"`
	DeviceUUID             string     `json:"device_uuid" db:"device_uuid"`
	HMACSignature          string     `json:"hmac_signature" db:"hmac_signature"`
	Synced                 bool       `json:"synced" db:"synced"`
	CreatedAt              time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at" db:"updated_at"`
	CheckInIdempotencyKey  *uuid.UUID `json:"check_in_idempotency_key,omitempty" db:"check_in_idempotency_key"`
	CheckOutIdempotencyKey *uuid.UUID `json:"check_out_idempotency_key,omitempty" db:"check_out_idempotency_key"`
}

type Location struct {
	ID           uuid.UUID `json:"id" db:"id"`
	Name         string    `json:"name" db:"name"`
	Address      *string   `json:"address,omitempty" db:"address"`
	Latitude     float64   `json:"latitude" db:"latitude"`
	Longitude    float64   `json:"longitude" db:"longitude"`
	RadiusMeters int       `json:"radius_meters" db:"radius_meters"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

type CheckInRequest struct {
	// Latitude in decimal degrees, range: -90 to 90
	Latitude float64 `json:"latitude" validate:"required,min=-90,max=90"`
	// Longitude in decimal degrees, range: -180 to 180
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
	// DeviceUUID must be a valid UUID
	DeviceUUID string `json:"device_uuid" validate:"required,uuid"`
	// Timestamp in seconds since epoch
	Timestamp int64 `json:"timestamp" validate:"required"`
	// HMAC signature for request authentication
	HMACSig string `json:"hmac_signature" validate:"required"`
	// Optional base64-encoded selfie image data
	SelfieData string `json:"selfie_data"`
	// Short-lived signed challenge issued by the backend.
	LivenessChallenge string    `json:"liveness_challenge" validate:"required"`
	LivenessToken     string    `json:"liveness_token" validate:"required"`
	IdempotencyKey    uuid.UUID `json:"idempotency_key" validate:"required"`
}

type FaceChallengeResponse struct {
	Challenge string    `json:"challenge"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type CheckOutRequest struct {
	// Latitude in decimal degrees, range: -90 to 90
	Latitude float64 `json:"latitude" validate:"required,min=-90,max=90"`
	// Longitude in decimal degrees, range: -180 to 180
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
	// DeviceUUID must be a valid UUID
	DeviceUUID string `json:"device_uuid" validate:"required,uuid"`
	// Timestamp in seconds since epoch
	Timestamp int64 `json:"timestamp" validate:"required"`
	// HMAC signature for request authentication
	HMACSig        string    `json:"hmac_signature" validate:"required"`
	IdempotencyKey uuid.UUID `json:"idempotency_key" validate:"required"`
}

type AttendanceResponse struct {
	Attendance
	UserName     string `json:"user_name"`
	LocationName string `json:"location_name"`
}

type HistoryQuery struct {
	UserID    uuid.UUID
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int
	Offset    int
}

type OfflinePayload struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	ActionType      string    `json:"action_type"`
	Payload         string    `json:"payload,omitempty"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	DeviceTimestamp time.Time `json:"device_timestamp"`
	SelfieData      string    `json:"selfie_data,omitempty"`
	Synced          bool      `json:"synced"`
	SyncAttempts    int       `json:"sync_attempts"`
	HMACSignature   string    `json:"-"`
	CreatedAt       time.Time `json:"created_at"`
}

type SyncAction struct {
	IdempotencyKey uuid.UUID       `json:"idempotency_key" validate:"required"`
	ActionType     string          `json:"action_type" validate:"required,oneof=check_in check_out"`
	Payload        json.RawMessage `json:"payload" validate:"required"`
}

type SyncRequest struct {
	Actions []SyncAction `json:"actions" validate:"required,min=1,max=20,dive"`
}

type SyncResult struct {
	IdempotencyKey uuid.UUID   `json:"idempotency_key"`
	Status         string      `json:"status"`
	Attendance     *Attendance `json:"attendance,omitempty"`
	Error          string      `json:"error,omitempty"`
}

type SyncStatusResponse struct {
	PendingCount int        `json:"pending_count"`
	StuckCount   int        `json:"stuck_count"`
	LastSyncAt   *time.Time `json:"last_sync_at,omitempty"`
}

// CreateLocationRequest contains fields for creating a new location with geofence
type CreateLocationRequest struct {
	// Location name, max 255 characters
	Name string `json:"name" validate:"required,max=255"`
	// Optional address description
	Address *string `json:"address,omitempty" validate:"omitempty,max=500"`
	// Latitude in decimal degrees, range: -90 to 90
	Latitude float64 `json:"latitude" validate:"required,min=-90,max=90"`
	// Longitude in decimal degrees, range: -180 to 180
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
	// Radius in meters, must be greater than 0
	RadiusMeters int `json:"radius_meters" validate:"required,gt=0"`
}

// UpdateLocationRequest contains fields for updating an existing location
type UpdateLocationRequest struct {
	// Location name, max 255 characters
	Name *string `json:"name,omitempty" validate:"omitempty,max=255"`
	// Optional address description
	Address *string `json:"address,omitempty" validate:"omitempty,max=500"`
	// Latitude in decimal degrees, range: -90 to 90
	Latitude *float64 `json:"latitude,omitempty" validate:"omitempty,min=-90,max=90"`
	// Longitude in decimal degrees, range: -180 to 180
	Longitude *float64 `json:"longitude,omitempty" validate:"omitempty,min=-180,max=180"`
	// Radius in meters, must be greater than 0
	RadiusMeters *int `json:"radius_meters,omitempty" validate:"omitempty,gt=0"`
}

// UploadSelfieRequest contains selfie image metadata for validation
type UploadSelfieRequest struct {
	// File size in bytes, max 5MB (5242880 bytes)
	FileSize int64 `json:"file_size" validate:"required,gt=0,max=5242880"`
	// Image format: jpg, jpeg, or png
	Format string `json:"format" validate:"required,oneof=jpg jpeg png"`
}

// UpdateEmbeddingRequest contains a face embedding vector for update
type UpdateEmbeddingRequest struct {
	// Face embedding vector, must have at least 1 element
	// Typically 128 or 512-dimensional float32 vector
	Embedding []float32 `json:"embedding" validate:"required,min=1"`
}
