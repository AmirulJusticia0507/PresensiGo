package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Attendance struct {
	ID               uuid.UUID      `json:"id" db:"id"`
	UserID           uuid.UUID      `json:"user_id" db:"user_id"`
	LocationID       uuid.UUID      `json:"location_id" db:"location_id"`
	CheckInTime      *time.Time     `json:"check_in_time,omitempty" db:"check_in_time"`
	CheckOutTime     *time.Time     `json:"check_out_time,omitempty" db:"check_out_time"`
	CheckInLocation  []float64      `json:"check_in_location,omitempty" db:"check_in_location"`
	CheckOutLocation []float64      `json:"check_out_location,omitempty" db:"check_out_location"`
	SelfieURL        *string        `json:"selfie_url,omitempty" db:"selfie_url"`
	Status           string         `json:"status" db:"status"`
	IsLate           bool           `json:"is_late" db:"is_late"`
	Notes            *string        `json:"notes,omitempty" db:"notes"`
	DeviceUUID       string         `json:"device_uuid" db:"device_uuid"`
	HMACSignature    string         `json:"hmac_signature" db:"hmac_signature"`
	Synced           bool           `json:"synced" db:"synced"`
	CreatedAt        time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at" db:"updated_at"`
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
	Latitude    float64 `json:"latitude" validate:"required"`
	Longitude   float64 `json:"longitude" validate:"required"`
	DeviceUUID  string  `json:"device_uuid" validate:"required"`
	HMACSig     string  `json:"hmac_signature" validate:"required"`
	SelfieData  string  `json:"selfie_data"`
}

type CheckOutRequest struct {
	Latitude    float64 `json:"latitude" validate:"required"`
	Longitude   float64 `json:"longitude" validate:"required"`
	DeviceUUID  string  `json:"device_uuid" validate:"required"`
	HMACSig     string  `json:"hmac_signature" validate:"required"`
}

type AttendanceResponse struct {
	Attendance
	UserName   string `json:"user_name"`
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
	UserID          uuid.UUID `json:"user_id"`
	ActionType      string    `json:"action_type"`
	Latitude        float64   `json:"latitude"`
	Longitude       float64   `json:"longitude"`
	DeviceTimestamp time.Time `json:"device_timestamp"`
	SelfieData      string    `json:"selfie_data,omitempty"`
}

type SyncRequest struct {
	Payloads      []OfflinePayload `json:"payloads" validate:"required"`
	DeviceUUID    string           `json:"device_uuid" validate:"required"`
	HMACSignatures pq.StringArray  `json:"hmac_signatures" validate:"required"`
}
