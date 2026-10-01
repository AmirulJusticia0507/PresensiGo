# Implementation Plan: Fix Validate Middleware Backend

## Overview

Perbaiki compile error di package validate middleware, refactor validasi ke helper per-handler, dan tambah test untuk memastikan `go build`, `go vet`, dan `go test` lulus.

## Task Dependency Graph

```json
{
  "waves": [
    { "wave": 1, "tasks": ["1"] },
    { "wave": 2, "tasks": ["2"] },
    { "wave": 3, "tasks": ["3"] },
    { "wave": 4, "tasks": ["4"] },
    { "wave": 5, "tasks": ["5"] },
    { "wave": 6, "tasks": ["6"] },
    { "wave": 7, "tasks": ["7"] },
    { "wave": 8, "tasks": ["8"] }
  ]
}
```

Tasks bersifat sekuensial. Task 5 (verifikasi build) harus lulus sebelum membuat test di task 6.

## Tasks

- [ ] 1. Hapus direktori validate yang bermasalah
  - Hapus file `backend/internal/delivery/http/middleware/validate/validate.go`
  - Hapus direktori `backend/internal/delivery/http/middleware/validate/`
  - Verifikasi tidak ada import ke package ini di file lain dengan grep
  - **Files:** `backend/internal/delivery/http/middleware/validate/validate.go`

- [ ] 2. Tambah validator dan helper validateRequest ke handler.go
  - Import `github.com/go-playground/validator/v10` di `handler.go`
  - Tambah package-level variable `var validate = validator.New()`
  - Tambah fungsi `validateRequest(w http.ResponseWriter, r *http.Request, dst interface{}) error`
  - Fungsi melakukan decode JSON body ke dst lalu validate.Struct(dst)
  - Jika decode gagal → respondError 400 "invalid request body"
  - Jika validasi gagal → respondError 400 dengan pesan error dari validator
  - **Files:** `backend/internal/delivery/http/handler.go`

- [ ] 3. Update handler Register dan Login menggunakan validateRequest
  - Ganti decode manual di handler Register dengan `validateRequest(w, r, &req)`
  - Ganti decode manual di handler Login dengan `validateRequest(w, r, &req)`
  - Pastikan handler return lebih awal jika validateRequest mengembalikan error
  - **Files:** `backend/internal/delivery/http/handler.go`

- [ ] 4. Update handler CheckIn dan CheckOut menggunakan validateRequest
  - Ganti decode manual di handler CheckIn dengan `validateRequest(w, r, &req)`
  - Ganti decode manual di handler CheckOut dengan `validateRequest(w, r, &req)`
  - Pastikan handler return lebih awal jika validateRequest mengembalikan error
  - **Files:** `backend/internal/delivery/http/handler.go`

- [ ] 5. Verifikasi go build dan go vet lulus
  - Jalankan `go build ./...` dari direktori `backend/`
  - Jalankan `go vet ./...` dari direktori `backend/`
  - Kedua perintah harus lulus tanpa error
  - **Files:** (verifikasi saja)

- [ ] 6. Buat handler_test.go dengan test validasi request
  - Buat file `backend/internal/delivery/http/handler_test.go`
  - Definisikan mock interface untuk AuthUsecase dan AttendanceUsecase
  - Test: POST /api/auth/register dengan body `{}` harus return 400
  - Test: POST /api/auth/register dengan email format salah harus return 400
  - Test: POST /api/auth/login dengan body `{}` harus return 400
  - Gunakan `net/http/httptest` dan `github.com/gorilla/mux`
  - **Files:** `backend/internal/delivery/http/handler_test.go`

- [ ] 7. Verifikasi go test lulus
  - Jalankan `go test ./...` dari direktori `backend/`
  - Semua test harus lulus
  - **Files:** (verifikasi saja)

- [ ] 8. Commit dan push perubahan
  - `git add` semua file yang berubah
  - Commit dengan pesan: `fix: perbaiki middleware validasi backend agar go build/test/vet lulus`
  - Push ke branch baru `fix/validate-middleware`
  - **Files:** (git operations)

## Notes

- `go-playground/validator` sudah ada di `go.mod` sebagai indirect dependency — tidak perlu `go get`.
- Validator di-reuse sebagai package-level variable karena `validator.New()` mahal dan thread-safe.
- Test di task 6 tidak memerlukan koneksi database (pure httptest mock).
