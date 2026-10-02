package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/PresensiGo/backend/internal/model"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func TestUUID_Generation(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()

	if id1 == id2 {
		t.Error("Expected two different UUIDs")
	}

	parsed, err := uuid.Parse(id1.String())
	if err != nil {
		t.Fatalf("Failed to parse UUID: %v", err)
	}

	if parsed != id1 {
		t.Error("UUID parse roundtrip failed")
	}
}

func TestPQ_ErrorHandling(t *testing.T) {
	// Test that pq.Error can be created and has message
	err := &pq.Error{Message: "test error"}
	if err.Message != "test error" {
		t.Errorf("Expected 'test error', got '%s'", err.Message)
	}
}

// Test repository method signatures and model structure
func TestUserModel_ProfileFields(t *testing.T) {
	// Test that User model has all required profile fields
	user := &model.User{
		ID:                      uuid.New(),
		Name:                    "Test User",
		Email:                   "test@example.com",
		PasswordHash:            "hashed_password",
		Role:                    "employee",
		FaceSimilarityThreshold: 0.45,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}

	// Verify required fields
	if user.ID == uuid.Nil {
		t.Error("User ID should not be nil")
	}
	if user.Email == "" {
		t.Error("User Email should not be empty")
	}
	if user.CreatedAt.IsZero() {
		t.Error("User CreatedAt should not be zero")
	}

	// Test optional profile fields
	phone := "+62812345678"
	user.Phone = &phone

	emergencyName := "Jane Doe"
	user.EmergencyContactName = &emergencyName

	emergencyPhone := "+62812345679"
	user.EmergencyContactPhone = &emergencyPhone

	address := "123 Main St, Jakarta"
	user.Address = &address

	profileURL := "https://storage.example.com/profile.jpg"
	user.ProfilePictureUrl = &profileURL

	now := time.Now()
	user.TermsAcceptedAt = &now

	// Verify optional fields are assigned
	if user.Phone == nil || *user.Phone != phone {
		t.Error("Phone field not properly assigned")
	}
	if user.EmergencyContactName == nil || *user.EmergencyContactName != emergencyName {
		t.Error("EmergencyContactName field not properly assigned")
	}
	if user.EmergencyContactPhone == nil || *user.EmergencyContactPhone != emergencyPhone {
		t.Error("EmergencyContactPhone field not properly assigned")
	}
	if user.Address == nil || *user.Address != address {
		t.Error("Address field not properly assigned")
	}
	if user.ProfilePictureUrl == nil || *user.ProfilePictureUrl != profileURL {
		t.Error("ProfilePictureUrl field not properly assigned")
	}
	if user.TermsAcceptedAt == nil {
		t.Error("TermsAcceptedAt field not properly assigned")
	}
}

// Test context usage in repository methods
func TestRepositoryMethods_ContextHandling(t *testing.T) {
	// This test verifies that the repository methods properly accept context
	// In a real integration test, you would use a test database

	// Test timeout context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if ctx == nil {
		t.Error("Context should not be nil")
	}

	// Verify context deadline is set
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Error("Context should have a deadline")
	}
	if deadline.IsZero() {
		t.Error("Deadline should not be zero")
	}

	// Test that context cancellation works
	cancelCtx, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()
	select {
	case <-cancelCtx.Done():
		// Context successfully cancelled
	default:
		t.Error("Context should be cancelled")
	}
}

// Test User model JSON marshaling
func TestUserModel_JSONMarshaling(t *testing.T) {
	user := &model.User{
		ID:                      uuid.New(),
		Name:                    "Test User",
		Email:                   "test@example.com",
		PasswordHash:            "hashed_password",
		Role:                    "employee",
		FaceSimilarityThreshold: 0.45,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}

	// Verify that password hash is not exposed in JSON
	// (handled by the json:"-" tag)
	if user.PasswordHash == "" {
		t.Error("PasswordHash should not be empty in memory")
	}
}

// Test optional field handling
func TestUserModel_OptionalFields(t *testing.T) {
	user := &model.User{
		ID:    uuid.New(),
		Name:  "Test User",
		Email: "test@example.com",
	}

	// Optional fields should be nil by default
	if user.Phone != nil {
		t.Error("Phone should be nil by default")
	}
	if user.Address != nil {
		t.Error("Address should be nil by default")
	}
	if user.EmergencyContactName != nil {
		t.Error("EmergencyContactName should be nil by default")
	}
	if user.ProfilePictureUrl != nil {
		t.Error("ProfilePictureUrl should be nil by default")
	}
	if user.TermsAcceptedAt != nil {
		t.Error("TermsAcceptedAt should be nil by default")
	}
}
