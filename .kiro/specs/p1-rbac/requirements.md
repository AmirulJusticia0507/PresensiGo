# Requirements Document

## Introduction

JWT sudah membawa claim `role`, tetapi middleware auth hanya memasukkan `user_id` ke context. Semua user yang login dapat memanggil POST/PUT/DELETE lokasi untuk membuat, mengubah, atau menghapus lokasi. Tidak ada pemeriksaan role admin pada endpoint-endpoint administrasi. Sistem access control belum diterapkan end-to-end.

## Requirements

### 1. Extract Role dari JWT ke Context

**User Story:** Sebagai developer, saya ingin middleware auth menyediakan role di context sehingga handler dapat menggunakannya untuk authorization checks.

#### Acceptance Criteria

- 1.1: Middleware auth mengekstrak claim `role` dari JWT token dan memasukkan ke context dengan key yang konsisten
- 1.2: Handler dapat mengambil role dari context menggunakan helper function `GetRoleFromContext(ctx)`
- 1.3: Jika token tidak memiliki role, gunakan default "employee" (bukan undefined/nil)
- 1.4: `go build ./...` dan `go vet ./...` lulus tanpa error

### 2. Restrict Location Mutations to Admin Role

**User Story:** Sebagai admin, saya ingin hanya admin yang dapat membuat, mengubah, atau menghapus lokasi, sehingga data lokasi tidak bisa diubah secara sembarangan.

#### Acceptance Criteria

- 2.1: POST /api/locations hanya dapat diakses oleh user dengan role "admin"
- 2.2: PUT /api/locations/{id} hanya dapat diakses oleh user dengan role "admin"
- 2.3: DELETE /api/locations/{id} hanya dapat diakses oleh user dengan role "admin"
- 2.4: GET /api/locations (read) dapat diakses oleh semua user terauthentikasi
- 2.5: Non-admin yang mencoba mutasi menerima HTTP 403 Forbidden dengan body `{"error": "admin role required"}`
- 2.6: Response format error konsisten dengan error response lainnya

### 3. Test RBAC Enforcement

**User Story:** Sebagai QA, saya ingin ada test yang memastikan RBAC diterapkan dan tidak bisa di-bypass.

#### Acceptance Criteria

- 3.1: Unit test dengan mock: admin role → 200 pada POST/PUT/DELETE lokasi
- 3.2: Unit test dengan mock: employee role → 403 pada POST/PUT/DELETE lokasi
- 3.3: Integration test: login sebagai employee, attempt PUT lokasi → 403
- 3.4: Integration test: login sebagai admin, attempt PUT lokasi → 200
- 3.5: Test coverage untuk authorization middleware minimal 80%

### 4. Prepare Admin User Data

**User Story:** Sebagai operator, saya ingin data test dengan admin dan employee user tersedia untuk local development dan testing.

#### Acceptance Criteria

- 4.1: Seed script membuat minimal 1 admin user dan 2 employee user
- 4.2: Admin user memiliki email `admin@presensigo.local`, password `admin123` (hanya dev)
- 4.3: Employee user memiliki email `employee1@presensigo.local`, password `employee123`
- 4.4: Seed script dapat dijalankan: `make seed` atau `go run cmd/seed/main.go`
- 4.5: Seed script aman untuk dijalankan berkali-kali (idempotent atau safe re-run)

## Glossary

- **RBAC:** Role-Based Access Control — sistem otorisasi berdasarkan role user
- **Claim:** Informasi dalam JWT token (user_id, role, email, dll)
- **Context:** Struktur Go untuk menyimpan request-scoped data (user_id, role, request_id, dll)
- **403 Forbidden:** HTTP response ketika user terauthentikasi tetapi tidak memiliki permission untuk resource tersebut
