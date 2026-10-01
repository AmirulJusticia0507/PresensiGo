# Requirements Document

## Introduction

Package `internal/delivery/http/middleware/validate` tidak bisa dikompilasi sehingga seluruh backend gagal `go test ./...` dan `go vet ./...`. Middleware validasi perlu diperbaiki dan dipasang ke router agar request invalid menghasilkan HTTP 400.

## Requirements

### 1. Fix Compile Errors di validate.go

**User Story:** Sebagai developer, saya ingin backend bisa dikompilasi penuh sehingga semua package bisa ditest.

#### Acceptance Criteria

- 1.1: File `validate.go` tidak memiliki import conflict antara `net/http` stdlib dan package delivery `http` internal.
- 1.2: Package `strings` di-import karena digunakan.
- 1.3: Tidak ada penggunaan `http.Json` yang tidak ada — gunakan `encoding/json` standard library.
- 1.4: Fungsi `respondError` dipanggil dari package yang benar (bukan cross-package call ke unexported function).
- 1.5: Import `model` dihapus jika tidak dipakai.
- 1.6: `go vet ./...` lulus tanpa error pada seluruh package.
- 1.7: `go build ./...` lulus tanpa error.

### 2. Refactor Pendekatan Validasi

**User Story:** Sebagai developer, saya ingin validation tag pada request model benar-benar diterapkan, bukan membaca body sebagai `interface{}` generik.

#### Acceptance Criteria

- 2.1: Middleware validasi tidak mencoba decode body generik karena validation tag pada `interface{}` tidak bekerja.
- 2.2: Validasi request body dilakukan per-handler dengan decode ke struct konkret, lalu `validate.Struct()` dipanggil setelah decode.
- 2.3: Helper fungsi `validateRequest(w, r, dst interface{}) error` tersedia di package `http` delivery untuk decode + validate sekaligus.
- 2.4: Handler Register, Login, CheckIn, CheckOut menggunakan helper `validateRequest`.
- 2.5: Jika validasi gagal, response HTTP 400 dikembalikan dengan body JSON `{"error": "<pesan validasi>"}`.

### 3. Middleware Dipasang ke Router

**User Story:** Sebagai operator, saya ingin middleware yang sudah dibuat benar-benar aktif melindungi endpoint.

#### Acceptance Criteria

- 3.1: `InitValidation()` tidak lagi dipanggil di `main.go` karena validator diinisialisasi di package level.
- 3.2: Route handler di `handler.go` mengembalikan 400 untuk request body yang tidak valid (field required kosong, format email salah, dll).

### 4. Test untuk Validasi Request

**User Story:** Sebagai developer, saya ingin ada test yang memastikan endpoint menolak request invalid dengan HTTP 400.

#### Acceptance Criteria

- 4.1: Ada file test yang menguji endpoint register dengan payload kosong dan mendapat 400.
- 4.2: Ada test yang menguji login dengan email format salah dan mendapat 400.
- 4.3: `go test ./...` lulus (semua test pass).
- 4.4: Test tidak memerlukan koneksi database — gunakan mock atau httptest.

## Glossary

- **validate middleware**: Middleware yang memvalidasi request body sebelum diteruskan ke handler.
- **validation tag**: Tag struct Go seperti `validate:"required,email"` yang dibaca oleh library `go-playground/validator`.
- **httptest**: Package standard library Go untuk testing HTTP handler tanpa server nyata.
