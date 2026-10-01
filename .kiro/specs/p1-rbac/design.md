# Design: Role-Based Authorization (RBAC)

## Overview

JWT sudah membawa role. Kita perlu (1) ekstrak role ke context di middleware, (2) implement authorization checks di handler, (3) return 403 untuk unauthorized access, (4) test enforcement.

Tidak ada perubahan database atau model — hanya flow/middleware/handler updates.

## Architecture

Flow request dengan RBAC:

```
Request dengan Bearer token
  ↓
AuthMiddleware:
  - Parse JWT
  - Ekstrak user_id, role
  - Inject ke context
  ↓
Handler:
  - Get user_id, role dari context
  - Check authorization (jika mutasi, cek role == admin)
  - Proceed atau return 403
  ↓
Response
```

## Components and Interfaces

### Middleware: auth.go (updated)

File: `backend/internal/delivery/http/middleware/auth.go`

Tambahan:

```go
type contextKey string

const (
    UserIDKey contextKey = "user_id"
    RoleKey   contextKey = "role"
)

// Dalam AuthMiddleware, tambahan baris setelah extract claims:
ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
ctx = context.WithValue(ctx, RoleKey, claims.Role)
next.ServeHTTP(w, r.WithContext(ctx))

// Helper function baru:
func GetRoleFromContext(ctx context.Context) string {
    if role, ok := ctx.Value(RoleKey).(string); ok {
        return role
    }
    return "employee" // default role jika tidak ada di context
}
```

### Handler Authorization Checks: handler.go (updated)

File: `backend/internal/delivery/http/handler.go`

Tambahan helper:

```go
func requireAdmin(w http.ResponseWriter, r *http.Request) bool {
    role := middleware.GetRoleFromContext(r.Context())
    if role != "admin" {
        respondError(w, http.StatusForbidden, "admin role required")
        return false
    }
    return true
}
```

Update handlers:

```go
func (h *Handler) CreateLocation(w http.ResponseWriter, r *http.Request) {
    if !requireAdmin(w, r) {
        return
    }
    // rest of handler...
}

func (h *Handler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
    if !requireAdmin(w, r) {
        return
    }
    // rest of handler...
}

func (h *Handler) DeleteLocation(w http.ResponseWriter, r *http.Request) {
    if !requireAdmin(w, r) {
        return
    }
    // rest of handler...
}
```

## Data Models

No changes — role sudah di JWT dan model Auth.

## Error Handling

- 403 Forbidden: user terauthentikasi tetapi role tidak authorized
- Response: `{"error": "admin role required"}` (konsisten dengan error lain)
- Logging: log semua authorization failures dengan user_id, role, endpoint untuk audit

## Testing Strategy

Unit tests menggunakan mock context:

```go
func TestCreateLocationAdmin(t *testing.T) {
    ctx := context.WithValue(context.Background(), middleware.RoleKey, "admin")
    req := httptest.NewRequest("POST", "/api/locations", body)
    req = req.WithContext(ctx)
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    // assert w.Code == 200
}

func TestCreateLocationEmployee(t *testing.T) {
    ctx := context.WithValue(context.Background(), middleware.RoleKey, "employee")
    req := httptest.NewRequest("POST", "/api/locations", body)
    req = req.WithContext(ctx)
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    // assert w.Code == 403
}
```

Integration tests menggunakan real login:

```go
func TestRBAC_EmployeeCannotMutateLocation(t *testing.T) {
    // 1. Login sebagai employee → get token
    // 2. POST /locations dengan employee token → expect 403
    // 3. Login sebagai admin → get token
    // 4. POST /locations dengan admin token → expect 201/200
}
```

## Correctness Properties

### Property 1: Role Always in Context
Setiap authenticated request harus memiliki role di context, minimal default "employee".
**Validates: Requirements 1.1, 1.3**

### Property 2: Authorization Enforced at Handler Boundary
Sebelum business logic dieksekusi, handler harus check authorization. Unauthorized request return 403 sebelum state berubah.
**Validates: Requirements 2.1–2.5**

### Property 3: No Privilege Escalation
Employee user tidak bisa self-elevate ke admin. Role diambil dari JWT (server-signed), tidak dari request body.
**Validates: Requirements 2.1, 2.3**
