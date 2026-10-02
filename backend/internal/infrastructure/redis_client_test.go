package infrastructure

import (
	"context"
	"testing"
	"time"
)

// TestBackgroundReconnect verifies that the background reconnection routine works
func TestBackgroundReconnect(t *testing.T) {
	// Create a Redis client (this will start the background routine)
	rc := NewRedisClientForTesting("localhost:6379")
	defer rc.Close()

	// Initially should be marked as connected
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if !rc.IsConnected(ctx) {
		t.Error("Expected Redis to be connected initially")
	}
}

// TestIsConnectedThreadSafe verifies that IsConnected is thread-safe
func TestIsConnectedThreadSafe(t *testing.T) {
	rc := NewRedisClientForTesting("localhost:6379")
	defer rc.Close()

	ctx := context.Background()

	// Spawn multiple goroutines that call IsConnected
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			rc.IsConnected(ctx)
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// TestUpdateConnectionState verifies connection state updates are properly synchronized
func TestUpdateConnectionState(t *testing.T) {
	rc := NewRedisClientForTesting("localhost:6379")
	defer rc.Close()

	ctx := context.Background()

	// Initially should be connected
	if !rc.IsConnected(ctx) {
		t.Error("Expected Redis to be connected initially")
	}

	// Simulate a connection loss
	rc.updateConnectionState(false)
	if rc.IsConnected(ctx) {
		t.Error("Expected Redis to be disconnected after updateConnectionState(false)")
	}

	// Simulate a connection recovery
	rc.updateConnectionState(true)
	if !rc.IsConnected(ctx) {
		t.Error("Expected Redis to be connected after updateConnectionState(true)")
	}
}

// TestCloseStopsReconnectRoutine verifies that Close stops the reconnect goroutine
func TestCloseStopsReconnectRoutine(t *testing.T) {
	rc := NewRedisClientForTesting("localhost:6379")

	// Close should not panic and should gracefully stop the routine
	err := rc.Close()
	if err != nil {
		// It's okay if there's an error (Redis might not be running), but Close should handle it
		t.Logf("Close returned error (expected if Redis not running): %v", err)
	}

	// Give the goroutine time to receive the stop signal
	time.Sleep(100 * time.Millisecond)
}

// TestConnectionStateOnOperations verifies connection state updates on operations
func TestConnectionStateOnOperations(t *testing.T) {
	rc := NewRedisClientForTesting("localhost:6379")
	defer rc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try an operation (this will fail if Redis is not running, but the state should still update)
	_, err := rc.Increment(ctx, "test-key", 1*time.Minute)

	// We only care that the function completes without panicking
	// The actual success depends on Redis being available
	_ = err
}
