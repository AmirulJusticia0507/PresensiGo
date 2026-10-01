# Implementation Plan: Role-Based Authorization

## Overview

Implement role extraction ke context, authorization checks di handler, dan comprehensive tests.

## Task Dependency Graph

```json
{
  "waves": [
    { "wave": 1, "tasks": ["1", "2"] },
    { "wave": 2, "tasks": ["3"] },
    { "wave": 3, "tasks": ["4", "5"] },
    { "wave": 4, "tasks": ["6"] }
  ]
}
```

Task 1–2 paralel (middleware + handler), task 3 test, task 4–5 paralel (seed + integration), task 6 verify+commit.

## Tasks

- [ ] 1. Extract role dari JWT ke context di middleware
  - Buka `backend/internal/delivery/http/middleware/auth.go`
  - Tambah constant `RoleKey contextKey = "role"`
  - Di `AuthMiddleware`, setelah extract claims, inject role ke context
  - Tambah helper `GetRoleFromContext(ctx) string` yang return role atau default "employee"
  - Pastikan role dipass ke context sebelum `next.ServeHTTP()`
  - **Files:** `backend/internal/delivery/http/middleware/auth.go`

- [ ] 2. Implement authorization checks di handler location endpoints
  - Buka `backend/internal/delivery/http/handler.go`
  - Tambah helper `requireAdmin(w, r) bool` yang check role dan return 403 jika tidak admin
  - Update `CreateLocation`: tambah `if !requireAdmin(w, r) { return }` di awal
  - Update `UpdateLocation`: sama
  - Update `DeleteLocation`: sama
  - GET endpoints (GetLocations, GetProfile, CheckIn, CheckOut) tidak perlu authorization check (tetap public untuk auth user)
  - **Files:** `backend/internal/delivery/http/handler.go`

- [ ] 3. Buat unit tests untuk RBAC di handler_test.go
  - Tambah test `TestCreateLocationAsAdmin` — mock context dengan role="admin", POST lokasi → expect 201
  - Tambah test `TestCreateLocationAsEmployee` — mock context dengan role="employee", POST lokasi → expect 403
  - Tambah test `TestUpdateLocationAsAdmin` — PUT lokasi → expect 200 (jika berhasil update)
  - Tambah test `TestUpdateLocationAsEmployee` — PUT lokasi → expect 403
  - Tambah test `TestDeleteLocationAsAdmin` — DELETE lokasi → expect 200
  - Tambah test `TestDeleteLocationAsEmployee` — DELETE lokasi → expect 403
  - Tambah test `TestGetLocationsAsEmployee` — GET lokasi → expect 200 (no auth required)
  - **Files:** `backend/internal/delivery/http/handler_test.go`

- [ ] 4. Update seed script dengan admin user
  - Buka `backend/cmd/seed/main.go`
  - Pastikan seed membuat 1 admin user: email=`admin@presensigo.local`, role=`admin`, password=`admin123`
  - Pastikan seed membuat 2+ employee user: email=`employee1@presensigo.local`, `employee2@presensigo.local`, role=`employee`
  - Seed idempotent: check user exists sebelum insert atau use UPSERT
  - **Files:** `backend/cmd/seed/main.go`

- [ ] 5. Buat integration test: login + RBAC
  - Buka `backend/internal/delivery/http/handler_test.go` atau buat file test baru
  - Test flow: login as employee, get token, attempt POST location → expect 403
  - Test flow: login as admin, get token, POST location → expect 201
  - Pastikan integration test menggunakan actual auth flow (not mocked)
  - **Files:** `backend/internal/delivery/http/handler_test.go`

- [ ] 6. Verifikasi build, test, dan commit
  - Jalankan `make seed` atau `go run cmd/seed/main.go` dari `backend/`
  - Jalankan `go build ./...` dari `backend/` → harus lulus
  - Jalankan `go test ./...` dari `backend/` → semua test harus lulus (termasuk RBAC tests)
  - Jalankan `go vet ./...` → harus lulus
  - Commit: `git add backend/.kiro/specs/p1-rbac/ backend/internal/delivery/http/middleware/auth.go backend/internal/delivery/http/handler.go backend/internal/delivery/http/handler_test.go backend/cmd/seed/main.go`
  - Commit message: `feat: implement role-based authorization (RBAC) untuk location mutations`
  - Push ke current branch atau buat PR

## Notes

- Role di JWT sudah di-generate saat login (backend/internal/usecase/auth_usecase.go)
- Context key harus consistent antara middleware dan helper function
- Default role "employee" memastikan backward compatibility jika token lama atau invalid
- Seed script dapat dijalankan multiple times tanpa error (idempotent)
- Authorization check dilakukan sebelum business logic, jadi jika 403, tidak ada side effect
