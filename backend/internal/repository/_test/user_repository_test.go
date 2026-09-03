package repository_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/PresensiGo/backend/internal/model"
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