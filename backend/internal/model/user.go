package model

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                      uuid.UUID  `json:"id" db:"id"`
	Name                    string     `json:"name" db:"name"`
	Email                   string     `json:"email" db:"email"`
	PasswordHash            string     `json:"-" db:"password_hash"`
	Role                    string     `json:"role" db:"role"`
	DeviceUUID              *string    `json:"device_uuid,omitempty" db:"device_uuid"`
	Phone                   *string    `json:"phone,omitempty" db:"phone"`
	EmergencyContactName    *string    `json:"emergency_contact_name,omitempty" db:"emergency_contact_name"`
	EmergencyContactPhone   *string    `json:"emergency_contact_phone,omitempty" db:"emergency_contact_phone"`
	Address                 *string    `json:"address,omitempty" db:"address"`
	ProfilePictureUrl       *string    `json:"profile_picture_url,omitempty" db:"profile_picture_url"`
	TermsAcceptedAt         *time.Time `json:"terms_accepted_at,omitempty" db:"terms_accepted_at"`
	FaceEmbedding           []byte     `json:"-" db:"face_embedding"`
	FaceSimilarityThreshold float64    `json:"face_similarity_threshold" db:"face_similarity_threshold"`
	FaceEnrolledAt          *time.Time `json:"face_enrolled_at,omitempty" db:"face_enrolled_at"`
	CreatedAt               time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt               time.Time  `json:"updated_at" db:"updated_at"`
}

type RegisterRequest struct {
	// User's full name, required
	Name string `json:"name" validate:"required,max=255"`
	// User's email address, must be valid email format
	Email string `json:"email" validate:"required,email,max=255"`
	// Password must be at least 6 characters
	Password string `json:"password" validate:"required,min=6"`
}

type LoginRequest struct {
	// User's email address
	Email string `json:"email" validate:"required,email,max=255"`
	// User's password
	Password string `json:"password" validate:"required"`
	// Device UUID must be a valid UUID format
	DeviceUUID string `json:"device_uuid" validate:"required,uuid"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UpdateFaceRequest struct {
	// Face embedding vector data, must be non-empty
	// Typically 128 or 512-dimensional float32 vector
	FaceEmbedding []byte `json:"face_embedding" validate:"required,min=1"`
}

type EnrollFaceRequest struct {
	Selfies []string `json:"selfies" validate:"required,min=2,max=3,dive,required"`
}
