# Design Document

## Overview

This design implements comprehensive input validation, error handling, and response hygiene for the PresensiGo API. The solution consists of:

1. **Validation Tags & Validators:** Struct tags on request models with centralized validation logic
2. **Validation Middleware:** HTTP middleware that intercepts requests, validates, and returns sanitized 400 errors
3. **Error Handler:** Centralized error mapping that sanitizes messages and ensures consistent JSON responses
4. **Structured Logging:** Request ID generation and injection into all log messages
5. **CORS Configuration:** Environment-aware allowed origins loaded from config

## Architecture

```
HTTP Request
    ↓
RequestID Middleware (generates X-Request-ID)
    ↓
CORS Middleware (environment-aware allowed origins)
    ↓
Validation Middleware (struct tags + domain rules)
    ├─ Invalid? → Error Handler → 400 JSON + requestID
    ↓
Auth Middleware (existing)
    ├─ Unauthorized? → Error Handler → 401 JSON + requestID
    ↓
Route Handler (business logic)
    ├─ Error? → Error Handler → 5xx JSON + requestID (sanitized)
    ↓
Response (JSON + Content-Type + X-Request-ID header)
```

## Components and Interfaces

### 1. Request Models with Validation Tags

**File:** `backend/internal/model/validation.go` (new)

```go
type CreateLocationRequest struct {
    Name      string  `json:"name" validate:"required,max=255"`
    Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
    Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
    Radius    int32   `json:"radius" validate:"required,gt=0"`
}

type CheckInRequest struct {
    Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
    Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
    DeviceID  string  `json:"device_id" validate:"required,uuid"`
    Signature string  `json:"signature" validate:"required"`
}

type UploadSelfieRequest struct {
    FileSize int64  `json:"file_size" validate:"required,gt=0,max=5242880"` // 5MB
    Format   string `json:"format" validate:"required,oneof=jpg jpeg png"`
}

type UpdateEmbeddingRequest struct {
    Embedding []float32 `json:"embedding" validate:"required,min=1"` // length > 0
}
```

**Validation Rules:**
- Use `github.com/go-playground/validator/v10` for tag-based validation
- Custom validators for domain rules: `validateLatitude`, `validateLongitude`, `validateRadius`
- File size checks via separate validation function (not MIME parsing; actual payload size)

### 2. Validation Middleware

**File:** `backend/internal/delivery/http/middleware/validation.go` (new or replace broken file)

```go
func ValidationMiddleware(validator *validator.Validate) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Middleware validates POST/PUT request body
            // If invalid, calls ErrorHandler to return sanitized 400
            // Otherwise passes to next handler
        })
    }
}

func (h *HTTPHandler) validateRequest(req interface{}) error {
    // Validates struct and returns sanitized error on failure
    // Error message: "Invalid location: latitude must be between -90 and 90"
}
```

**Error Response on Validation Failure:**
```json
{
  "error": "Validation failed",
  "details": [
    "latitude: must be between -90 and 90",
    "radius: must be greater than 0"
  ],
  "requestID": "req_12345678"
}
```

### 3. Error Handler

**File:** `backend/internal/delivery/http/middleware/error_handler.go` (new)

```go
type ErrorResponse struct {
    Error     string        `json:"error"`
    Details   []string      `json:"details,omitempty"`
    RequestID string        `json:"requestID"`
    StatusCode int          `json:"statusCode"`
}

func HandleError(w http.ResponseWriter, err error, requestID string) {
    // Sanitizes error message, removes internal details
    // Always returns JSON with Content-Type: application/json
    // Maps error type to HTTP status code
}

func sanitizeError(err error) string {
    // Removes: database schema, SQL queries, stack traces, file paths
    // Maps internal errors to client-friendly messages
    // Examples:
    //   sql.ErrNoRows → "Record not found"
    //   context.DeadlineExceeded → "Request timeout"
    //   ParseError → "Invalid request format"
}
```

**Error Response Examples:**

Invalid Input:
```json
{
  "error": "Validation failed",
  "details": ["latitude: out of range"],
  "requestID": "req_abc123",
  "statusCode": 400
}
```

Unauthorized:
```json
{
  "error": "Unauthorized",
  "requestID": "req_def456",
  "statusCode": 401
}
```

Server Error (sanitized):
```json
{
  "error": "Internal server error",
  "requestID": "req_ghi789",
  "statusCode": 500
}
```

**Never include in response:**
- SQL error details: `duplicate key value violates unique constraint "users_email_key"`
- Stack traces: `goroutine 1 [running]: main.doSomething(...)`
- File paths: `/home/user/backend/internal/repository/...`
- Schema info: `column "user_id" does not exist`

### 4. Request ID Middleware

**File:** `backend/internal/delivery/http/middleware/request_id.go` (new)

```go
func RequestIDMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        requestID := r.Header.Get("X-Request-ID")
        if requestID == "" {
            requestID = generateRequestID() // UUID or similar
        }
        ctx := context.WithValue(r.Context(), "requestID", requestID)
        w.Header().Set("X-Request-ID", requestID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value("requestID").(string); ok {
        return id
    }
    return "unknown"
}
```

**Logging with Request ID:**
```go
// In auth middleware
log.Printf("[%s] Authentication failed: invalid token", GetRequestID(r.Context()))

// In validation middleware
log.Printf("[%s] Validation error: latitude out of range", GetRequestID(r.Context()))

// In handler
log.Printf("[%s] Check-in successful for user %s", GetRequestID(r.Context()), userID)
```

### 5. Structured Logging

**File:** `backend/internal/delivery/http/handler.go` (modified)

Add logging to all critical paths:

```go
func (h *HTTPHandler) Login(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())
    
    // Attempt to parse credentials
    // If invalid JSON:
    log.Printf("[%s] Invalid login request: %v", requestID, err)
    
    // If credentials invalid:
    log.Printf("[%s] Login failed: invalid credentials for user %s", requestID, username)
    
    // If device binding fails:
    log.Printf("[%s] Device binding failed: user %s already bound to device %s", 
        requestID, userID, existingDeviceID)
}

func (h *HTTPHandler) CheckIn(w http.ResponseWriter, r *http.Request) {
    requestID := GetRequestID(r.Context())
    
    // Validation errors already logged by middleware
    
    // Business logic errors:
    log.Printf("[%s] Check-in geofence validation failed: user %s outside radius",
        requestID, userID)
    
    log.Printf("[%s] Check-in successful: user %s at location %s",
        requestID, userID, locationID)
}
```

### 6. CORS Configuration

**File:** `backend/internal/config/cors.go` (new)

```go
type CORSConfig struct {
    AllowedOrigins []string
    AllowedMethods []string
    AllowedHeaders []string
    Credentials    bool
}

func LoadCORSConfig(env string) *CORSConfig {
    switch env {
    case "development":
        return &CORSConfig{
            AllowedOrigins: []string{
                "http://localhost:3000",
                "http://localhost:8080",
                "http://10.0.2.2:3000", // Android emulator
                "http://127.0.0.1:3000",
            },
            AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowedHeaders: []string{"Content-Type", "Authorization"},
            Credentials:    true,
        }
    case "staging":
        return &CORSConfig{
            AllowedOrigins: []string{"https://staging-app.example.com"},
            AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowedHeaders: []string{"Content-Type", "Authorization"},
            Credentials:    true,
        }
    case "production":
        return &CORSConfig{
            AllowedOrigins: []string{"https://app.example.com"},
            AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
            AllowedHeaders: []string{"Content-Type", "Authorization"},
            Credentials:    true,
        }
    default:
        // Fail safe: no CORS allowed if env unknown
        return &CORSConfig{
            AllowedOrigins: []string{},
        }
    }
}
```

**Usage in main.go:**
```go
corsConfig := LoadCORSConfig(os.Getenv("ENVIRONMENT")) // dev, staging, prod
router.Use(cors.New(cors.Config{
    AllowedOrigins: corsConfig.AllowedOrigins,
    AllowedMethods: corsConfig.AllowedMethods,
    AllowedHeaders: corsConfig.AllowedHeaders,
    AllowCredentials: corsConfig.Credentials,
}))
```

**Environment Variable:**
```
ENVIRONMENT=development  # or staging, production
```

## Data Models

### Request Validation Flow

```
POST /api/locations
  ↓
Body parsed to CreateLocationRequest
  ↓
ValidationMiddleware.validateRequest(req)
  ├─ Check struct tags (name: required, max=255)
  ├─ Check domain rules (latitude -90..90, longitude -180..180, radius > 0)
  ├─ Invalid? Return 400 with sanitized details
  ↓ (Valid)
Handler processes request
  ↓
Response 201 Created
```

### Error Type Classification

| Error Type | HTTP Status | Client Message |
|---|---|---|
| Validation Error | 400 | "Validation failed: {field}: {reason}" |
| Malformed JSON | 400 | "Invalid request format" |
| Unauthorized (no token) | 401 | "Unauthorized" |
| Unauthorized (bad token) | 401 | "Invalid token" |
| Forbidden (wrong role) | 403 | "Access denied" |
| Not Found | 404 | "Resource not found" |
| Conflict (unique constraint) | 409 | "Resource already exists" |
| Request Timeout | 408 | "Request timeout" |
| Database Error | 500 | "Internal server error" |
| Service Unavailable | 503 | "Service unavailable" |

## Correctness Properties

**Property 1: All Validation Errors Return HTTP 400**
- **Description:** When a request fails validation (invalid lat/lng, oversized file, etc.), the response status is 400 Bad Request
- **Test:** POST /api/locations with latitude=100 → 400; with latitude=45 → 200 (or 201)

**Property 2: Error Responses Contain No Internal Details**
- **Description:** No error response includes database schema, SQL queries, stack traces, or file paths
- **Test:** Send SQL injection attempt, verify response does not contain table names, column names, or SQL fragments

**Property 3: All Error Responses Include Request ID**
- **Description:** Every error response includes the requestID field and X-Request-ID header
- **Test:** POST invalid request → response includes "requestID" and header "X-Request-ID: req_abc123"

**Property 4: All Errors Include Content-Type: application/json**
- **Description:** Whether error occurs in middleware or handler, response has Content-Type: application/json
- **Test:** Malformed JSON, validation error, auth error → all have Content-Type: application/json

**Property 5: Request ID Persists Through Request Lifecycle**
- **Description:** The same request ID appears in logs, response headers, and error responses for a single request
- **Test:** Log request, generate error, check that all three contain the same ID

**Property 6: CORS Respects Environment Configuration**
- **Description:** Allowed origins depend on ENVIRONMENT; development allows localhost, production allows prod domain only
- **Test:** dev env with prod domain → CORS rejected; dev env with localhost → CORS allowed

## Error Handling

### Validation Errors

When request body fails validation:

1. Validation middleware intercepts request
2. Calls `validateRequest()` on parsed body
3. Validator returns list of failed rules
4. Middleware calls `ErrorHandler` with validation error
5. `ErrorHandler` maps to HTTP 400 and returns sanitized details

### Database Errors

When database operation fails:

1. Handler catches error (sql.ErrNoRows, unique constraint, etc.)
2. Maps to appropriate HTTP status via `sanitizeError()`
3. Calls `ErrorHandler` with mapped status and safe message
4. Client receives 400/404/409/500 with generic message

### Authentication/Authorization Errors

When auth middleware detects invalid token or forbidden access:

1. Middleware logs failure with request ID
2. Calls `ErrorHandler` with 401/403 status
3. Returns JSON with error message (no details needed; client knows auth failed)

### Timeout/Network Errors

When request context exceeds deadline:

1. Handler checks `ctx.Err()` and detects context.DeadlineExceeded
2. Calls `ErrorHandler` with 408 status and "Request timeout" message
3. No stack trace; no database query included

### Unhandled Panics

When handler panics (should not happen, but defense-in-depth):

1. Recovery middleware catches panic
2. Logs full stack trace server-side (not sent to client)
3. Calls `ErrorHandler` with 500 status and generic message
4. Client receives: `{"error": "Internal server error", "requestID": "..."}`
