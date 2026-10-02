package model

import (
	"time"

	"github.com/google/uuid"
)

type AttendanceFilter struct {
	Limit      int
	Offset     int
	DateFrom   *time.Time
	DateTo     *time.Time
	Status     string
	LocationID *uuid.UUID
	UserID     *uuid.UUID
}

type AttendanceListItem struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	UserName     string     `json:"user_name"`
	LocationID   uuid.UUID  `json:"location_id"`
	LocationName string     `json:"location_name"`
	CheckInTime  *time.Time `json:"check_in_time,omitempty"`
	CheckOutTime *time.Time `json:"check_out_time,omitempty"`
	Status       string     `json:"status"`
	IsLate       bool       `json:"is_late"`
	Notes        *string    `json:"notes,omitempty"`
}

type PaginatedAttendances struct {
	Items  []AttendanceListItem `json:"items"`
	Total  int                  `json:"total"`
	Limit  int                  `json:"limit"`
	Offset int                  `json:"offset"`
}

type AdminUserUpdate struct {
	Role        *string `json:"role,omitempty" validate:"omitempty,oneof=admin employee"`
	ResetDevice bool    `json:"reset_device,omitempty"`
}

type WorkSchedule struct {
	ID              uuid.UUID  `json:"id"`
	LocationID      *uuid.UUID `json:"location_id,omitempty"`
	UserID          *uuid.UUID `json:"user_id,omitempty"`
	DayOfWeek       int        `json:"day_of_week" validate:"gte=0,lte=6"`
	StartTime       string     `json:"start_time" validate:"required,datetime=15:04"`
	EndTime         string     `json:"end_time" validate:"required,datetime=15:04"`
	LateAfterMinute int        `json:"late_after_minutes" validate:"gte=0,lte=1440"`
	Active          bool       `json:"active"`
}

type LeaveRequest struct {
	ID         uuid.UUID  `json:"id"`
	UserID     uuid.UUID  `json:"user_id"`
	UserName   string     `json:"user_name,omitempty"`
	StartDate  time.Time  `json:"start_date"`
	EndDate    time.Time  `json:"end_date"`
	Type       string     `json:"type" validate:"required,oneof=leave sick permission"`
	Reason     string     `json:"reason" validate:"required,max=500"`
	Status     string     `json:"status"`
	ReviewedBy *uuid.UUID `json:"reviewed_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type CreateLeaveRequest struct {
	StartDate string `json:"start_date" validate:"required,datetime=2006-01-02"`
	EndDate   string `json:"end_date" validate:"required,datetime=2006-01-02"`
	Type      string `json:"type" validate:"required,oneof=leave sick permission"`
	Reason    string `json:"reason" validate:"required,max=500"`
}

type ReviewLeaveRequest struct {
	Status string `json:"status" validate:"required,oneof=approved rejected"`
}
