# Structured Logging Strategy

## Overview

This document describes the structured logging implementation for the PresensiGo API. All logging uses a consistent format that includes request IDs for end-to-end traceability.

## Log Format

All logs follow this format:
```
[requestID] event_description
```

Examples:
- `[req_12345678-abcd-1234-abcd-123456789012] Login successful for user 550e8400-e29b-41d4-a716-446655440000`
- `[req_87654321-dcba-4321-dcba-987654321098] Login failed: invalid credentials for email user@example.com`

## Request ID Generation

- **Source:** Generated automatically by `RequestIDMiddleware` in `middleware/request_id.go`
- **Format:** `req_` prefix followed by UUID (e.g., `req_550e8400-e29b-41d4-a716-446655440000`)
- **Header Override:** Can be overridden via `X-Request-ID` request header if provided
- **Response Header:** Always included in `X-Request-ID` response header

## Log Points

### Authentication Failures

All authentication failures are logged with the request ID:

- **Missing Authorization Header:**
  ```
  [requestID] Authentication failed: missing authorization header for POST /api/attendance/check-in
  ```

- **Invalid Authorization Format:**
  ```
  [requestID] Authentication failed: invalid authorization format for POST /api/attendance/check-in
  ```

- **Invalid or Expired Token:**
  ```
  [requestID] Authentication failed: invalid or expired token for POST /api/attendance/check-in
  ```

### Validation Errors

All validation errors are logged during parsing/decoding:

- **Invalid JSON:**
  ```
  [requestID] Invalid JSON format: invalid character '}' looking for beginning of value
  ```

- **Failed to Read Body:**
  ```
  [requestID] Failed to read request body: request body too large
  ```

- **Validation Failures:** Logged in handler.validateRequest():
  ```
  [requestID] Validation failed: latitude out of range
  ```

### Authorization Failures

- **Admin-Only Endpoints:**
  ```
  [requestID] Authorization failed: admin role required
  ```

- **Unauthorized Attempts:**
  ```
  [requestID] Unauthorized check-in attempt
  ```

### Login/Registration

- **Registration Success:**
  ```
  [requestID] Registration successful for user 550e8400-e29b-41d4-a716-446655440000
  ```

- **Registration Failure:**
  ```
  [requestID] Registration failed for email user@example.com: user already exists
  ```

- **Login Success:**
  ```
  [requestID] Login successful for user 550e8400-e29b-41d4-a716-446655440000
  ```

- **Login Failure:**
  ```
  [requestID] Login failed: invalid credentials for email user@example.com
  ```

### Attendance Operations

- **Check-In Success:**
  ```
  [requestID] Check-in successful for user 550e8400-e29b-41d4-a716-446655440000 at location
  ```

- **Check-In Failure:**
  ```
  [requestID] Check-in failed for user 550e8400-e29b-41d4-a716-446655440000: outside geofence
  ```

- **Check-Out Success:**
  ```
  [requestID] Check-out successful for user 550e8400-e29b-41d4-a716-446655440000
  ```

- **Check-Out Failure:**
  ```
  [requestID] Check-out failed for user 550e8400-e29b-41d4-a716-446655440000: check-in not found
  ```

- **Get Today's Attendance Failure:**
  ```
  [requestID] GetTodayAttendance failed for user 550e8400-e29b-41d4-a716-446655440000: no record found
  ```

- **Get History Failure:**
  ```
  [requestID] GetHistory failed for user 550e8400-e29b-41d4-a716-446655440000: database error
  ```

### Face Enrollment

- **Face Enrollment Success:**
  ```
  [requestID] Face enrollment successful for user 550e8400-e29b-41d4-a716-446655440000 with 3 samples
  ```

- **Face Enrollment Failure:**
  ```
  [requestID] Face enrollment failed for user 550e8400-e29b-41d4-a716-446655440000: processing error
  ```

- **Get Face Challenge Failure:**
  ```
  [requestID] GetFaceChallenge failed for user 550e8400-e29b-41d4-a716-446655440000: challenge generation error
  ```

### Location Management

- **Location Creation Success:**
  ```
  [requestID] Location created successfully with ID 550e8400-e29b-41d4-a716-446655440000
  ```

- **Location Creation Failure:**
  ```
  [requestID] CreateLocation failed: coordinates out of range
  ```

- **Location Update Success:**
  ```
  [requestID] Location updated successfully with ID 550e8400-e29b-41d4-a716-446655440000
  ```

- **Location Update Failure:**
  ```
  [requestID] UpdateLocation failed for ID 550e8400-e29b-41d4-a716-446655440000: not found
  ```

- **Location Deletion Success:**
  ```
  [requestID] Location deleted successfully with ID 550e8400-e29b-41d4-a716-446655440000
  ```

- **Location Deletion Failure:**
  ```
  [requestID] DeleteLocation failed for ID 550e8400-e29b-41d4-a716-446655440000: still in use
  ```

- **Get Locations Failure:**
  ```
  [requestID] GetLocations failed: database error
  ```

### Profile Operations

- **Get Profile Failure:**
  ```
  [requestID] GetProfile failed for user 550e8400-e29b-41d4-a716-446655440000: not found
  ```

- **Update Face Embedding Success:**
  ```
  [requestID] Face embedding updated successfully for user 550e8400-e29b-41d4-a716-446655440000
  ```

- **Update Face Embedding Failure:**
  ```
  [requestID] UpdateFaceEmbedding failed for user 550e8400-e29b-41d4-a716-446655440000: invalid format
  ```

### Health Checks

- **Liveness Check:**
  ```
  [requestID] Health check request from 127.0.0.1:54321
  ```

- **Readiness Check Success:**
  ```
  [requestID] Health readiness: OK (DB: true, Redis: true)
  ```

- **Readiness Check Failure:**
  ```
  [requestID] Health readiness: NOT READY (DB: false, Redis: true)
  ```

## Integration with Error Handler

The error handler (`middleware/error_handler.go`) ensures:
- All error responses include the `requestID` field
- Error responses have `Content-Type: application/json` header
- Error messages are sanitized (no internal details leaked)
- Request ID is logged and included in all responses

## Tracing a Request

To trace a single user action through logs:

1. Observe the `X-Request-ID` header in the HTTP response
2. Search logs for `[req_<id>]` to find all related log entries
3. The sequence of logs shows the request flow through middleware and handlers

Example:
```
[req_12345678] Health check request from 127.0.0.1
[req_12345678] Authentication failed: missing authorization header for POST /api/attendance/check-in
[req_12345678] Invalid JSON format: unexpected EOF
[req_12345678] Validation failed: latitude out of range
```

## Sanitization Requirements

Log messages NEVER include:
- Database schema names or column names
- SQL queries or query fragments
- Stack traces or Go package paths
- File system paths (except safe, intended ones)
- User passwords or tokens
- Raw error details from external services

## Configuration

No special configuration is required. Logging is enabled by default using Go's standard `log` package.

For production deployments, consider:
- Redirecting stdout/stderr to a logging service (e.g., ELK, Datadog)
- Using structured logging libraries (e.g., `uber-go/zap`) for more sophisticated log management
- Setting up log retention policies
