package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetRequestID_ReturnsUnknownWhenNotInContext verifies default return value
func TestGetRequestID_ReturnsUnknownWhenNotInContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	requestID := GetRequestID(req.Context())

	if requestID != "unknown" {
		t.Errorf("expected 'unknown', got %s", requestID)
	}
}

// TestRequestIDMiddleware_StartsWithReqPrefix verifies generated IDs follow naming convention
func TestRequestIDMiddleware_StartsWithReqPrefix(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		if len(requestID) < 4 || requestID[:4] != "req_" {
			t.Errorf("expected request ID to start with 'req_', got %s", requestID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestIDMiddleware(nextHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)
}

// TestRequestIDMiddleware_PassesIDToDownstream verifies request ID flows through chain
func TestRequestIDMiddleware_PassesIDToDownstream(t *testing.T) {
	expectedID := "req_abc-123-def"
	capturedID := ""

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestIDMiddleware(nextHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", expectedID)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if capturedID != expectedID {
		t.Errorf("expected ID to be %s, got %s", expectedID, capturedID)
	}
}
