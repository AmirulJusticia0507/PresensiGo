package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestFaceChallengeRoundTrip(t *testing.T) {
	userID := uuid.New()
	challenge, token, _, err := NewFaceChallenge(userID, "test-secret")
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyFaceChallenge(token, challenge, userID, "test-secret"); err != nil {
		t.Fatalf("valid challenge rejected: %v", err)
	}
	if err := VerifyFaceChallenge(token, challenge, uuid.New(), "test-secret"); err == nil {
		t.Fatal("challenge should be bound to the user")
	}
	if err := VerifyFaceChallenge(token, "different", userID, "test-secret"); err == nil {
		t.Fatal("challenge should be bound to the requested pose")
	}
}
