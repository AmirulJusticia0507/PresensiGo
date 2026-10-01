package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/validator/v10"
)

// TestRequestIDMiddleware_GeneratesRequestID verifies that middleware generates request ID when not provided
func TestRequestIDMiddleware_GeneratesRequestID(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		if requestID == "" || requestID == "unknown" {
			t.Error("expected request ID to be generated")
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestIDMiddleware(nextHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Check that X-Request-ID header is set
	if w.Header().Get("X-Request-ID") == "" {
		t.Error("expected X-Request-ID header to be set")
	}
}

// TestRequestIDMiddleware_UsesProvidedRequestID verifies that middleware uses X-Request-ID header if provided
func TestRequestIDMiddleware_UsesProvidedRequestID(t *testing.T) {
	providedID := "req_test_123"
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := GetRequestID(r.Context())
		if requestID != providedID {
			t.Errorf("expected %s, got %s", providedID, requestID)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := RequestIDMiddleware(nextHandler)
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", providedID)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if w.Header().Get("X-Request-ID") != providedID {
		t.Errorf("expected X-Request-ID header to be %s", providedID)
	}
}

// TestSanitizeValidationErrors_RequiredField tests that required field error is sanitized
func TestSanitizeValidationErrors_RequiredField(t *testing.T) {
	v := validator.New()

	type TestStruct struct {
		Name string `validate:"required"`
	}

	err := v.Struct(&TestStruct{})
	if err == nil {
		t.Fatal("expected validation error")
	}

	details := SanitizeValidationErrors(err)
	if len(details) == 0 {
		t.Fatal("expected details")
	}

	// Should contain "is required" message
	found := false
	for _, detail := range details {
		if detail == "Name is required" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'Name is required' in details, got %v", details)
	}
}

// TestSanitizeValidationErrors_MinLatitude tests that latitude validation is sanitized properly
func TestSanitizeValidationErrors_MinLatitude(t *testing.T) {
	v := validator.New()

	type LocationRequest struct {
		Latitude float64 `validate:"required,min=-90,max=90"`
	}

	err := v.Struct(&LocationRequest{Latitude: -91})
	if err == nil {
		t.Fatal("expected validation error")
	}

	details := SanitizeValidationErrors(err)
	if len(details) == 0 {
		t.Fatal("expected details")
	}

	// Should contain sanitized geographic range message
	found := false
	for _, detail := range details {
		// Error should not expose validator constraints, only user-friendly message
		if detail == "Latitude must be within valid geographic range" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected sanitized geographic range message, got %v", details)
	}
}

// TestSanitizeValidationErrors_Longitude tests longitude validation sanitization
func TestSanitizeValidationErrors_Longitude(t *testing.T) {
	v := validator.New()

	type LocationRequest struct {
		Longitude float64 `validate:"required,min=-180,max=180"`
	}

	err := v.Struct(&LocationRequest{Longitude: 181})
	if err == nil {
		t.Fatal("expected validation error")
	}

	details := SanitizeValidationErrors(err)
	if len(details) == 0 {
		t.Fatal("expected details")
	}

	// Should contain sanitized geographic range message
	found := false
	for _, detail := range details {
		if detail == "Longitude must be within valid geographic range" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected sanitized geographic range message for longitude, got %v", details)
	}
}

// TestSanitizeValidationErrors_UUID tests UUID validation sanitization
func TestSanitizeValidationErrors_UUID(t *testing.T) {
	v := validator.New()

	type CheckInRequest struct {
		DeviceUUID string `validate:"required,uuid"`
	}

	err := v.Struct(&CheckInRequest{DeviceUUID: "not-a-uuid"})
	if err == nil {
		t.Fatal("expected validation error")
	}

	details := SanitizeValidationErrors(err)
	if len(details) == 0 {
		t.Fatal("expected details")
	}

	// Should contain UUID message
	found := false
	for _, detail := range details {
		if detail == "DeviceUUID must be a valid UUID" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'must be a valid UUID' message, got %v", details)
	}
}

// TestSanitizeValidationErrors_NoLeakingSchema tests that database schema is not leaked
func TestSanitizeValidationErrors_NoLeakingSchema(t *testing.T) {
	v := validator.New()

	type LocationRequest struct {
		RadiusMeters int `validate:"gt=0"`
	}

	err := v.Struct(&LocationRequest{RadiusMeters: -5})
	if err == nil {
		t.Fatal("expected validation error")
	}

	details := SanitizeValidationErrors(err)
	detailsStr := ""
	for _, d := range details {
		detailsStr += d
	}

	// Ensure no schema details leaked
	if containsAny(detailsStr, []string{"column", "table", "schema", "database"}) {
		t.Errorf("error message contains schema details: %s", detailsStr)
	}
}

// TestSanitizeValidationErrors_OneOf tests oneof validation sanitization
func TestSanitizeValidationErrors_OneOf(t *testing.T) {
	v := validator.New()

	type SelfieRequest struct {
		Format string `validate:"required,oneof=jpg jpeg png"`
	}

	err := v.Struct(&SelfieRequest{Format: "gif"})
	if err == nil {
		t.Fatal("expected validation error")
	}

	details := SanitizeValidationErrors(err)
	if len(details) == 0 {
		t.Fatal("expected details")
	}

	// Should contain oneof message with allowed values
	found := false
	for _, detail := range details {
		if detail == "Format must be one of: jpg jpeg png" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected 'must be one of' message, got %v", details)
	}
}

// TestValidationMiddleware_SkipsGETRequests verifies that GET requests are not validated
func TestValidationMiddleware_SkipsGETRequests(t *testing.T) {
	nextHandlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	v := validator.New()
	middleware := ValidationMiddleware(v)(nextHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if !nextHandlerCalled {
		t.Error("expected next handler to be called for GET request")
	}
}

// TestValidationMiddleware_ValidatesPOST verifies that POST requests body is parsed
func TestValidationMiddleware_ValidatesPOST(t *testing.T) {
	nextHandlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	v := validator.New()
	middleware := ValidationMiddleware(v)(nextHandler)

	validJSON := `{"name": "Test"}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(validJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if !nextHandlerCalled {
		t.Error("expected next handler to be called for valid POST request")
	}
}

// TestValidationMiddleware_RejectsInvalidJSON verifies that malformed JSON returns 400
func TestValidationMiddleware_RejectsInvalidJSON(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	v := validator.New()
	middleware := ValidationMiddleware(v)(nextHandler)

	invalidJSON := `{invalid json`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Add request ID to context for middleware to use
	ctx := context.WithValue(req.Context(), RequestIDKey, "test_req_123")
	req = req.WithContext(ctx)

	middleware.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d", w.Code)
	}

	// Verify response includes requestID
	var response map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if _, ok := response["requestID"]; !ok {
		t.Error("expected requestID in error response")
	}

	if response["error"] != "Invalid request format" {
		t.Errorf("expected 'Invalid request format' error message")
	}
}

// TestValidationMiddleware_SkipsHealthEndpoints verifies health checks are not validated
func TestValidationMiddleware_SkipsHealthEndpoints(t *testing.T) {
	nextHandlerCalled := false
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextHandlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	v := validator.New()
	middleware := ValidationMiddleware(v)(nextHandler)

	// Test /health endpoint
	req := httptest.NewRequest("POST", "/health", bytes.NewBufferString("invalid"))
	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if !nextHandlerCalled {
		t.Error("expected next handler to be called for health endpoint")
	}

	nextHandlerCalled = false

	// Test /health/ready endpoint
	req = httptest.NewRequest("POST", "/health/ready", bytes.NewBufferString("invalid"))
	w = httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	if !nextHandlerCalled {
		t.Error("expected next handler to be called for health/ready endpoint")
	}
}

// TestValidationMiddleware_BodyRereadable verifies that middleware doesn't consume body for downstream handlers
func TestValidationMiddleware_BodyRereadable(t *testing.T) {
	bodyContentRead := ""
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		bodyContentRead = string(buf[:n])
		w.WriteHeader(http.StatusOK)
	})

	v := validator.New()
	middleware := ValidationMiddleware(v)(nextHandler)

	testBody := `{"name": "Test"}`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(testBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	if bodyContentRead != testBody {
		t.Errorf("expected body to be readable downstream, got %q", bodyContentRead)
	}
}

// TestValidationMiddleware_ResponseIncludesContentType verifies error response has correct Content-Type
func TestValidationMiddleware_ResponseIncludesContentType(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	v := validator.New()
	middleware := ValidationMiddleware(v)(nextHandler)

	invalidJSON := `{bad json`
	req := httptest.NewRequest("POST", "/test", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	// Add request ID to context
	ctx := context.WithValue(req.Context(), RequestIDKey, "test_req_123")
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	middleware.ServeHTTP(w, req)

	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type: application/json, got %s", contentType)
	}
}

// Helper function to check if any strings are contained in the target string
func containsAny(target string, substrings []string) bool {
	for _, substr := range substrings {
		if bytes.Contains([]byte(target), []byte(substr)) {
			return true
		}
	}
	return false
}
