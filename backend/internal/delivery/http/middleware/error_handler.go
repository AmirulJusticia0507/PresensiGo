package middleware

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// Custom error types for different error categories
type ValidationError struct {
	Message string
	Details []string
}

func (e ValidationError) Error() string {
	return e.Message
}

type AuthenticationError struct {
	Message string
}

func (e AuthenticationError) Error() string {
	return e.Message
}

type AuthorizationError struct {
	Message string
}

func (e AuthorizationError) Error() string {
	return e.Message
}

type NotFoundError struct {
	Message string
}

func (e NotFoundError) Error() string {
	return e.Message
}

type ConflictError struct {
	Message string
}

func (e ConflictError) Error() string {
	return e.Message
}

type ServerError struct {
	Message string
	Err     error
}

func (e ServerError) Error() string {
	return e.Message
}

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Error     string   `json:"error"`
	Details   []string `json:"details,omitempty"`
	RequestID string   `json:"requestID"`
	StatusCode int     `json:"statusCode"`
}

// HandleError sanitizes an error and sends a standardized JSON response
func HandleError(w http.ResponseWriter, err error, requestID string) {
	if requestID == "" {
		requestID = "unknown"
	}

	statusCode, sanitizedMessage, details := classifyAndSanitizeError(err)

	response := ErrorResponse{
		Error:      sanitizedMessage,
		RequestID:  requestID,
		StatusCode: statusCode,
	}

	if len(details) > 0 {
		response.Details = details
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)

	log.Printf("[%s] Error response: status=%d, message=%s", requestID, statusCode, sanitizedMessage)
}

// classifyAndSanitizeError maps error types to HTTP status codes and sanitized messages
func classifyAndSanitizeError(err error) (int, string, []string) {
	if err == nil {
		return http.StatusInternalServerError, "Internal server error", nil
	}

	// Handle custom error types
	switch e := err.(type) {
	case ValidationError:
		return http.StatusBadRequest, e.Message, e.Details

	case AuthenticationError:
		return http.StatusUnauthorized, e.Message, nil

	case AuthorizationError:
		return http.StatusForbidden, e.Message, nil

	case NotFoundError:
		return http.StatusNotFound, e.Message, nil

	case ConflictError:
		return http.StatusConflict, e.Message, nil

	case ServerError:
		// Log the actual error server-side but send generic message to client
		log.Printf("Server error details: %v", e.Err)
		return http.StatusInternalServerError, e.Message, nil
	}

	// Handle standard error types
	errorMsg := err.Error()

	// Database errors
	if err == sql.ErrNoRows {
		return http.StatusNotFound, "Record not found", nil
	}

	// Check for SQL/database error patterns
	if strings.Contains(errorMsg, "duplicate key") || strings.Contains(errorMsg, "UNIQUE constraint") {
		return http.StatusConflict, "Resource already exists", nil
	}

	if strings.Contains(errorMsg, "not found") || strings.Contains(errorMsg, "no rows") {
		return http.StatusNotFound, "Record not found", nil
	}

	// Check for context errors
	if strings.Contains(errorMsg, "context deadline exceeded") || strings.Contains(errorMsg, "timeout") {
		return http.StatusRequestTimeout, "Request timeout", nil
	}

	// Check for JSON/parsing errors
	if strings.Contains(errorMsg, "invalid character") || 
	   strings.Contains(errorMsg, "Unmarshal") || 
	   strings.Contains(errorMsg, "json") {
		return http.StatusBadRequest, "Invalid request format", nil
	}

	// Default to 500 for unclassified errors
	log.Printf("Unclassified error: %v", err)
	return http.StatusInternalServerError, "Internal server error", nil
}

// sanitizeMessage removes sensitive information from error messages
func sanitizeMessage(msg string) string {
	// Remove common sensitive patterns
	sensitive := []string{
		"column", "table", "database", "schema",
		"goroutine", "main.", "runtime/",
		"/home/", "/var/", "C:\\", "D:\\",
		".go:", "panic", "stack trace",
	}

	result := msg
	for _, pattern := range sensitive {
		if strings.Contains(strings.ToLower(result), strings.ToLower(pattern)) {
			return "Internal server error"
		}
	}

	return result
}

// MapErrorToStatus returns the HTTP status code for a given error
func MapErrorToStatus(err error) int {
	statusCode, _, _ := classifyAndSanitizeError(err)
	return statusCode
}
