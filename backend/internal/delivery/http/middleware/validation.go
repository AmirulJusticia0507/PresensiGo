package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationMiddleware validates JSON request bodies against struct tags
// If validation fails, returns HTTP 400 with sanitized error details
func ValidationMiddleware(validator *validator.Validate) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only validate POST and PUT requests
			if r.Method != http.MethodPost && r.Method != http.MethodPut {
				next.ServeHTTP(w, r)
				return
			}

			// Skip validation for certain endpoints that don't need it
			if shouldSkipValidation(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			// Read the request body
			body, err := io.ReadAll(r.Body)
			if err != nil {
				requestID := GetRequestID(r.Context())
				log.Printf("[%s] Failed to read request body: %v", requestID, err)
				respondValidationError(w, r.Context(), "Invalid request format", nil)
				return
			}

			// Reset body for downstream handlers
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			// Try to parse and validate the body
			// We don't know the exact type at middleware level, so we validate generic JSON structure
			var jsonData interface{}
			if err := json.Unmarshal(body, &jsonData); err != nil {
				requestID := GetRequestID(r.Context())
				log.Printf("[%s] Invalid JSON format: %v", requestID, err)
				respondValidationError(w, r.Context(), "Invalid request format", nil)
				return
			}

			// Continue to next handler
			next.ServeHTTP(w, r)
		})
	}
}

// ValidateRequest validates a struct against validator rules and returns sanitized errors
func ValidateRequest(ctx context.Context, v *validator.Validate, data interface{}) error {
	if err := v.Struct(data); err != nil {
		return err
	}
	return nil
}

// respondValidationError sends a sanitized validation error response
func respondValidationError(w http.ResponseWriter, ctx context.Context, errorMsg string, details []string) {
	requestID := GetRequestID(ctx)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)

	response := map[string]interface{}{
		"error":     errorMsg,
		"requestID": requestID,
	}

	if len(details) > 0 {
		response["details"] = details
	}

	json.NewEncoder(w).Encode(response)
}

// SanitizeValidationErrors converts validator errors to sanitized client-friendly messages
func SanitizeValidationErrors(err error) []string {
	if err == nil {
		return nil
	}

	var details []string
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		// Not a validation error, return generic message
		return []string{"Validation failed"}
	}

	for _, fieldError := range validationErrors {
		fieldName := fieldError.Field()
		tag := fieldError.Tag()

		// Create user-friendly error message based on validation tag
		message := sanitizeFieldError(fieldName, tag, fieldError.Param())
		details = append(details, message)
	}

	return details
}

// sanitizeFieldError creates a sanitized error message for a specific field validation failure
func sanitizeFieldError(fieldName, tag, param string) string {
	normalizedField := strings.ToLower(fieldName)
	switch tag {
	case "required":
		return fieldName + " is required"
	case "email":
		return fieldName + " must be a valid email address"
	case "uuid":
		return fieldName + " must be a valid UUID"
	case "min":
		if strings.Contains(normalizedField, "latitude") || strings.Contains(normalizedField, "longitude") {
			return fieldName + " must be within valid geographic range"
		}
		return fieldName + " must be at least " + param
	case "max":
		if strings.Contains(normalizedField, "latitude") || strings.Contains(normalizedField, "longitude") {
			return fieldName + " must be within valid geographic range"
		}
		return fieldName + " must be at most " + param
	case "gte", "lte":
		if strings.Contains(normalizedField, "latitude") || strings.Contains(normalizedField, "longitude") {
			return fieldName + " must be within valid geographic range"
		}
		return fieldName + " is outside the allowed range"
	case "gt":
		return fieldName + " must be greater than " + param
	case "lt":
		return fieldName + " must be less than " + param
	case "oneof":
		return fieldName + " must be one of: " + param
	case "len":
		return fieldName + " must have length " + param
	default:
		return fieldName + " is invalid"
	}
}

// shouldSkipValidation returns true if the endpoint doesn't need request body validation
func shouldSkipValidation(path string) bool {
	// Skip validation for health checks
	if path == "/health" || path == "/health/ready" {
		return true
	}
	// Skip validation for GET requests (no body to validate)
	// This is handled by method check in middleware, but keeping for clarity
	return false
}
