package middleware

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleError_ValidationError_Returns400(t *testing.T) {
	w := httptest.NewRecorder()
	err := ValidationError{
		Message: "Validation failed",
		Details: []string{"latitude: must be between -90 and 90", "radius: must be greater than 0"},
	}

	HandleError(w, err, "req_12345")

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected StatusCode 400, got %d", response.StatusCode)
	}

	if response.Error != "Validation failed" {
		t.Errorf("Expected error 'Validation failed', got '%s'", response.Error)
	}

	if response.RequestID != "req_12345" {
		t.Errorf("Expected requestID 'req_12345', got '%s'", response.RequestID)
	}

	if len(response.Details) != 2 {
		t.Errorf("Expected 2 details, got %d", len(response.Details))
	}
}

func TestHandleError_AuthenticationError_Returns401(t *testing.T) {
	w := httptest.NewRecorder()
	err := AuthenticationError{Message: "Unauthorized"}

	HandleError(w, err, "req_67890")

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	var response ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Error != "Unauthorized" {
		t.Errorf("Expected error 'Unauthorized', got '%s'", response.Error)
	}

	if response.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected StatusCode 401, got %d", response.StatusCode)
	}
}

func TestHandleError_AuthorizationError_Returns403(t *testing.T) {
	w := httptest.NewRecorder()
	err := AuthorizationError{Message: "Insufficient permissions"}

	HandleError(w, err, "req_abc123")

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	var response ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Error != "Insufficient permissions" {
		t.Errorf("Expected error 'Insufficient permissions', got '%s'", response.Error)
	}

	if response.StatusCode != http.StatusForbidden {
		t.Errorf("Expected StatusCode 403, got %d", response.StatusCode)
	}
}

func TestHandleError_NotFoundError_Returns404(t *testing.T) {
	w := httptest.NewRecorder()
	err := NotFoundError{Message: "Record not found"}

	HandleError(w, err, "req_def456")

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Error != "Record not found" {
		t.Errorf("Expected error 'Record not found', got '%s'", response.Error)
	}

	if response.StatusCode != http.StatusNotFound {
		t.Errorf("Expected StatusCode 404, got %d", response.StatusCode)
	}
}

func TestHandleError_ConflictError_Returns409(t *testing.T) {
	w := httptest.NewRecorder()
	err := ConflictError{Message: "Resource already exists"}

	HandleError(w, err, "req_ghi789")

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", w.Code)
	}

	var response ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Error != "Resource already exists" {
		t.Errorf("Expected error 'Resource already exists', got '%s'", response.Error)
	}

	if response.StatusCode != http.StatusConflict {
		t.Errorf("Expected StatusCode 409, got %d", response.StatusCode)
	}
}

func TestHandleError_ServerError_Returns500(t *testing.T) {
	w := httptest.NewRecorder()
	err := ServerError{Message: "Internal server error", Err: sql.ErrConnDone}

	HandleError(w, err, "req_jkl012")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var response ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Error != "Internal server error" {
		t.Errorf("Expected error 'Internal server error', got '%s'", response.Error)
	}

	if response.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected StatusCode 500, got %d", response.StatusCode)
	}
}

func TestHandleError_SQLNoRowsError_Returns404(t *testing.T) {
	w := httptest.NewRecorder()
	err := sql.ErrNoRows

	HandleError(w, err, "req_mno345")

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Error != "Record not found" {
		t.Errorf("Expected error 'Record not found', got '%s'", response.Error)
	}
}

func TestHandleError_DuplicateKeyError_Returns409(t *testing.T) {
	w := httptest.NewRecorder()
	HandleError(w, ConflictError{Message: "Resource already exists"}, "req_test")

	if w.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", w.Code)
	}

	// Verify it doesn't contain the actual error message
	body := w.Body.String()
	if bytes.Contains([]byte(body), []byte("duplicate key")) {
		t.Errorf("Response should not contain database error details")
	}
}

func TestHandleError_AllErrorsIncludeContentType(t *testing.T) {
	errors := []error{
		ValidationError{Message: "Validation failed"},
		AuthenticationError{Message: "Unauthorized"},
		AuthorizationError{Message: "Forbidden"},
		NotFoundError{Message: "Not found"},
		ConflictError{Message: "Conflict"},
		ServerError{Message: "Internal error", Err: sql.ErrConnDone},
	}

	for _, err := range errors {
		w := httptest.NewRecorder()
		HandleError(w, err, "req_test")

		contentType := w.Header().Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type 'application/json', got '%s' for error: %v", contentType, err)
		}
	}
}

func TestHandleError_AllErrorsIncludeRequestID(t *testing.T) {
	errors := []error{
		ValidationError{Message: "Validation failed"},
		AuthenticationError{Message: "Unauthorized"},
		NotFoundError{Message: "Not found"},
		ServerError{Message: "Internal error", Err: sql.ErrConnDone},
	}

	requestID := "req_unique_123"

	for _, err := range errors {
		w := httptest.NewRecorder()
		HandleError(w, err, requestID)

		var response ErrorResponse
		json.NewDecoder(w.Body).Decode(&response)

		if response.RequestID != requestID {
			t.Errorf("Expected requestID '%s', got '%s' for error: %v", requestID, response.RequestID, err)
		}
	}
}

func TestHandleError_EmptyRequestID_UsesDefault(t *testing.T) {
	w := httptest.NewRecorder()
	err := ServerError{Message: "Error", Err: nil}

	HandleError(w, err, "")

	var response ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.RequestID != "unknown" {
		t.Errorf("Expected requestID 'unknown' for empty input, got '%s'", response.RequestID)
	}
}

func TestHandleError_ValidationError_WithDetails(t *testing.T) {
	w := httptest.NewRecorder()
	details := []string{"field1: error1", "field2: error2", "field3: error3"}
	err := ValidationError{Message: "Validation failed", Details: details}

	HandleError(w, err, "req_details_test")

	var response ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if len(response.Details) != len(details) {
		t.Errorf("Expected %d details, got %d", len(details), len(response.Details))
	}

	for i, detail := range response.Details {
		if detail != details[i] {
			t.Errorf("Detail mismatch at index %d: expected '%s', got '%s'", i, details[i], detail)
		}
	}
}

func TestHandleError_ValidationError_NoDetails(t *testing.T) {
	w := httptest.NewRecorder()
	err := ValidationError{Message: "Validation failed"}

	HandleError(w, err, "req_no_details")

	var response ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if len(response.Details) != 0 {
		t.Errorf("Expected no details, got %d", len(response.Details))
	}

	// Verify details field is omitted in JSON (not present as null)
	body := w.Body.String()
	if bytes.Contains([]byte(body), []byte("\"details\":null")) {
		t.Errorf("Details field should be omitted if empty, not null")
	}
}

func TestHandleError_UnclassifiedError_Returns500(t *testing.T) {
	w := httptest.NewRecorder()
	err := ServerError{Message: "Internal server error", Err: sql.ErrConnDone}

	HandleError(w, err, "req_unclassified")

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var response ErrorResponse
	json.NewDecoder(w.Body).Decode(&response)

	if response.Error != "Internal server error" {
		t.Errorf("Expected generic error message, got '%s'", response.Error)
	}
}

func TestClassifyAndSanitizeError_ValidationError(t *testing.T) {
	err := ValidationError{
		Message: "Validation failed",
		Details: []string{"lat: out of range"},
	}

	status, msg, details := classifyAndSanitizeError(err)

	if status != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", status)
	}

	if msg != "Validation failed" {
		t.Errorf("Expected message 'Validation failed', got '%s'", msg)
	}

	if len(details) != 1 {
		t.Errorf("Expected 1 detail, got %d", len(details))
	}
}

func TestClassifyAndSanitizeError_DatabaseConnectionError(t *testing.T) {
	err := sql.ErrConnDone

	status, msg, _ := classifyAndSanitizeError(err)

	if status != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", status)
	}

	if msg != "Internal server error" {
		t.Errorf("Expected generic message, got '%s'", msg)
	}
}

func TestMapErrorToStatus(t *testing.T) {
	tests := []struct {
		err      error
		expected int
	}{
		{ValidationError{Message: "test"}, http.StatusBadRequest},
		{AuthenticationError{Message: "test"}, http.StatusUnauthorized},
		{AuthorizationError{Message: "test"}, http.StatusForbidden},
		{NotFoundError{Message: "test"}, http.StatusNotFound},
		{ConflictError{Message: "test"}, http.StatusConflict},
		{ServerError{Message: "test", Err: nil}, http.StatusInternalServerError},
		{sql.ErrNoRows, http.StatusNotFound},
	}

	for _, test := range tests {
		status := MapErrorToStatus(test.err)
		if status != test.expected {
			t.Errorf("Error %v: expected status %d, got %d", test.err, test.expected, status)
		}
	}
}

func TestHandleError_ResponseIsValidJSON(t *testing.T) {
	w := httptest.NewRecorder()
	err := ValidationError{
		Message: "Validation failed",
		Details: []string{"field: error"},
	}

	HandleError(w, err, "req_json_test")

	var response ErrorResponse
	err2 := json.NewDecoder(w.Body).Decode(&response)

	if err2 != nil {
		t.Errorf("Response is not valid JSON: %v", err2)
	}
}
