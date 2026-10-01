# Implementation Plan

## Overview

This milestone implements comprehensive input validation, error handling, and security controls for the PresensiGo backend. It adds validation middleware, centralized error handling, structured request tracking, and environment-aware CORS configuration to protect against invalid inputs and security threats. The implementation includes both unit and integration tests to verify all validation rules work correctly across request types.

## Task Dependency Graph

Sequential flow for tasks 1-5 (each depends on the previous). Tasks 6 and 7 are parallel and can be executed after task 5 completes.

```json
{
  "waves": [
    { "wave": 1, "tasks": ["Task 1"] },
    { "wave": 2, "tasks": ["Task 2"] },
    { "wave": 3, "tasks": ["Task 3"] },
    { "wave": 4, "tasks": ["Task 4"] },
    { "wave": 5, "tasks": ["Task 5"] },
    { "wave": 6, "tasks": ["Task 6", "Task 7"] }
  ]
}
```

## Tasks

- [ ] 1. Add validate tags to request models
  - Add `validate` struct tags to CreateLocationRequest, CheckInRequest, UpdateLocationRequest, UploadSelfieRequest, UpdateEmbeddingRequest
  - Document validation rules inline (e.g., latitude -90..90, radius > 0, file size <= 5MB)
  - Ensure tags use `github.com/go-playground/validator/v10` syntax

- [ ] 2. Implement validation middleware
  - Create `backend/internal/delivery/http/middleware/validation.go`
  - Parse request body and validate against struct tags + domain rules
  - Return HTTP 400 with sanitized error details on validation failure
  - Include requestID in error response
  - Middleware must not leak internal details (no SQL, schema, stack traces)

- [ ] 3. Create error handler middleware
  - Create `backend/internal/delivery/http/middleware/error_handler.go`
  - Implement `HandleError(w http.ResponseWriter, err error, requestID string)`
  - Map error types to HTTP status codes (validation→400, auth→401/403, server→500, etc.)
  - Sanitize error messages: remove database schema, query details, stack traces
  - Ensure all responses include `Content-Type: application/json` and requestID field

- [ ] 4. Add structured logging
  - Create `backend/internal/delivery/http/middleware/request_id.go`
  - Generate unique request ID (UUID) on entry; accept X-Request-ID header if provided
  - Inject request ID into request context
  - Add X-Request-ID to response headers
  - Update handler logging to include request ID for auth failures and validation errors
  - Log format: `[requestID] event_description`

- [ ] 5. Fix CORS configuration
  - Create `backend/internal/config/cors.go`
  - Implement `LoadCORSConfig(env string)` to return environment-specific allowed origins
  - Development: allow localhost, 127.0.0.1, 10.0.2.2 (Android emulator)
  - Staging: allow explicit staging domain
  - Production: allow explicit production domain
  - Never use `origin: *` with `AllowCredentials: true`
  - Load ENVIRONMENT variable in main.go and apply CORS config

- [ ] 6. Write validation unit tests
  - Test CreateLocationRequest with valid/invalid latitude (edge cases: -90, 0, 90, 91, -91)
  - Test CreateLocationRequest with valid/invalid longitude (edge cases: -180, 0, 180, 181, -181)
  - Test CreateLocationRequest with valid/invalid radius (0, 1, -1, 999999)
  - Test UploadSelfieRequest with file size (0 bytes, 5MB exact, 5MB+1, format validation)
  - Test UpdateEmbeddingRequest with embedding vector length (empty, 1 element, many elements)
  - All tests should verify that validation passes/fails as expected
  - File: `backend/internal/model/validation_test.go`

- [ ] 7. Write integration tests
  - Test POST /api/locations with invalid latitude → HTTP 400 with sanitized error (no schema leak)
  - Test POST /api/check-in with SQL injection in deviceID → HTTP 400, no SQL in response
  - Test auth error response → HTTP 401 with Content-Type: application/json
  - Test validation error response → includes requestID field and X-Request-ID header
  - Test CORS preflight → origin header matches environment config (not *)
  - Test error response sanitization → no stack trace, no database query details
  - File: `backend/internal/delivery/http/handler_test.go` (add new tests)

## Notes

- All validation must follow the `github.com/go-playground/validator/v10` package conventions
- Error responses must be consistent across all endpoints (use centralized error handler)
- Request IDs should be logged at the start and end of request processing for tracing
- CORS configuration should be environment-aware and never expose internal details
