# Requirements Document

## Introduction

This specification addresses a critical gap in the PresensiGo backend: input validation, error handling, and response hygiene. Currently, the API accepts invalid data (out-of-range coordinates, oversized files) and may leak sensitive information through error responses (database schema, query details, stack traces). Additionally, CORS is misconfigured for production, and request logging lacks structure for debugging.

This milestone ensures all API inputs are validated against domain constraints, all errors are sanitized and consistently formatted, CORS is environment-aware, and logging provides traceability via request IDs.

## Requirements

### R1: Input Validation with struct Tags

**Requirement:** All request models must declare validation constraints via struct tags that are actively enforced.

- Latitude must be in range [-90, 90]
- Longitude must be in range [-180, 180]
- Radius must be greater than 0
- Selfie/image files must be <= 5MB in size
- Embedding vectors must have length > 0

**Rationale:** Domain constraints prevent invalid data from entering the system. Struct tags provide a declarative, maintainable specification of rules.

### R2: Domain Validation Layer

**Requirement:** Implement a centralized validation layer that checks domain constraints before business logic executes.

- Validate latitude/longitude on location creation/update
- Validate radius on geofence queries
- Validate file size and format on selfie upload
- Validate embedding vector length on face enrollment

**Rationale:** Centralizing validation improves consistency, reduces duplication, and makes rules auditable.

### R3: Sanitized Error Responses

**Requirement:** All error responses must be safe for client consumption; never leak internal details.

- Never include database schema, table names, or column names in error messages
- Never include SQL queries or query fragments
- Never include stack traces or Go package paths
- Never include file system paths (except safe, intended ones like S3 object keys)
- All errors must use generic client-friendly messages (e.g., "Invalid location coordinates")

**Rationale:** Information leakage can enable attackers to craft more effective exploits. Users don't need implementation details; they need actionable guidance.

### R4: Consistent Error Response Format

**Requirement:** All error responses must include `Content-Type: application/json` and follow a consistent structure.

- Status code must reflect the error class: 400 for validation, 401 for auth, 403 for authz, 500 for server errors
- Response body must be valid JSON with `error` or `errors` field
- Include `requestID` for tracing (see R5)

**Rationale:** Consistent format allows clients to parse errors predictably. `Content-Type: application/json` prevents MIME-sniffing attacks.

### R5: Structured Logging with Request IDs

**Requirement:** All requests must be assigned a unique request ID; all logs must include this ID for debugging end-to-end request flow.

- Generate request ID on entry (e.g., UUID or X-Request-ID header if provided)
- Inject request ID into context and pass to all downstream functions
- Include request ID in all log messages and error responses
- Log authentication failures (invalid token, no token) with request ID
- Log validation errors with request ID and field name

**Rationale:** Request IDs allow tracing a single user action through distributed logs. Logging auth/validation failures helps detect attacks and debug user issues.

### R6: Environment-Aware CORS Configuration

**Requirement:** Replace hardcoded `origin: *` with environment-specific allowed origins.

- Development: allow `http://localhost:*`, `http://10.0.2.2:*` (Android emulator), `http://127.0.0.1:*`
- Staging: allow explicit staging frontend domain (e.g., `https://staging-app.example.com`)
- Production: allow explicit production frontend domain (e.g., `https://app.example.com`)
- Never use `origin: *` with `AllowCredentials: true` (insecure)
- Load allowed origins from environment variable or config file

**Rationale:** CORS `*` with credentials is a security anti-pattern. Environment-specific lists prevent cross-origin attacks and support multi-environment deployments.

## Success Criteria

1. ✅ All request models have validation tags for domain constraints
2. ✅ Request validator middleware rejects invalid data with HTTP 400 and sanitized message
3. ✅ All error responses include `Content-Type: application/json`
4. ✅ All error responses omit internal details (schema, stack traces, paths)
5. ✅ All requests receive a unique request ID in X-Request-ID response header
6. ✅ Logs include request ID for authentication and validation failures
7. ✅ CORS configuration is loaded from environment (not hardcoded `*`)
8. ✅ `go test ./...` passes with validation tests covering valid/invalid inputs
9. ✅ Integration test verifies SQL injection attempt → safe 400 error (no schema leakage)
10. ✅ Integration test verifies stack trace absent from error response

## Glossary

- **Domain Validation:** Checking that inputs satisfy business rules (lat/lng ranges, positive radius, file size limits).
- **Sanitized:** Error messages contain only information safe for clients; no internal implementation details.
- **Request ID:** Unique identifier assigned to each HTTP request, used for logging and tracing.
- **CORS:** Cross-Origin Resource Sharing; HTTP headers controlling which origins can access the API.
- **Struct Tags:** Go metadata annotations on struct fields (e.g., `validate:"min=0,max=180"`).
