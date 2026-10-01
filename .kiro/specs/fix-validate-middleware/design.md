# Design: Fix Validate Middleware Backend

## Overview

Masalah utama adalah `validate.go` mencoba mengimport package delivery `http` internal yang menyebabkan circular import, menggunakan simbol yang tidak ada (`http.Json`, `respondError`), dan mengimport package yang tidak terpakai. Pendekatan middleware generik yang membaca body sebagai `interface{}` juga tidak dapat menerapkan validation tag Go.

Solusinya: hapus subpackage `validate` yang bermasalah, pindahkan logika validasi ke helper function di package delivery `http`, dan terapkan di tiap handler secara per-struct.

## Architecture

Arsitektur setelah fix mengikuti pola yang sudah ada di `handler.go`:

```
Request → AuthMiddleware → Handler → validateRequest helper → business logic
```

Tidak ada perubahan arsitektur besar. `validateRequest` adalah internal helper di package `http` delivery, bukan middleware HTTP terpisah. Ini karena setiap endpoint membutuhkan validasi struct yang berbeda, sehingga tidak bisa dilakukan secara generik di middleware.

## Components and Interfaces

### Handler (diubah)

File: `backend/internal/delivery/http/handler.go`

Tambahan:
```go
var validate = validator.New()

func validateRequest(w http.ResponseWriter, r *http.Request, dst interface{}) error {
    if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
        respondError(w, http.StatusBadRequest, "invalid request body")
        return err
    }
    if err := validate.Struct(dst); err != nil {
        respondError(w, http.StatusBadRequest, err.Error())
        return err
    }
    return nil
}
```

### validate subpackage (dihapus)

File: `backend/internal/delivery/http/middleware/validate/validate.go`

Dihapus karena circular import dan desain yang tidak bisa dipertahankan.

### Handler Test (baru)

File: `backend/internal/delivery/http/handler_test.go`

Mock interface untuk AuthUsecase dan AttendanceUsecase, test menggunakan `net/http/httptest`.

## Data Models

Tidak ada perubahan data model. Struct request yang ada sudah memiliki validation tag yang benar:

- `RegisterRequest`: `name` (required), `email` (required,email), `password` (required,min=6)
- `LoginRequest`: `email` (required,email), `password` (required), `device_uuid` (required)
- `CheckInRequest`: `latitude` (required), `longitude` (required), `device_uuid` (required), `hmac_signature` (required)
- `CheckOutRequest`: `latitude` (required), `longitude` (required), `device_uuid` (required), `hmac_signature` (required)

## Error Handling

- Decode JSON gagal: HTTP 400 dengan `{"error": "invalid request body"}`
- Validation gagal: HTTP 400 dengan `{"error": "<pesan dari validator>"}` — pesan berisi nama field dan constraint yang gagal
- Handler return lebih awal setelah `validateRequest` mengembalikan error, response sudah ditulis oleh helper

## Testing Strategy

Test menggunakan `net/http/httptest.NewRecorder()` dan `httptest.NewRequest()` sehingga tidak perlu database atau server. Mock usecase diimplementasikan sebagai struct yang memenuhi interface. Skenario test:
1. Register dengan body kosong `{}` → 400
2. Register dengan email tidak valid → 400
3. Login dengan body kosong → 400

## Correctness Properties

### Property 1: No double response write
`validateRequest` tidak pernah menulis response dua kali — caller wajib return setelah error dikembalikan.
**Validates: Requirements 2.5**

### Property 2: Thread-safe validator
Validator di-inisialisasi sekali sebagai package-level variable, thread-safe untuk concurrent request sesuai dokumentasi `go-playground/validator`.
**Validates: Requirements 2.3**

### Property 3: No breaking changes
Response format `{"error": "..."}` konsisten dengan format yang sudah ada di `respondError`, tidak ada breaking change untuk client.
**Validates: Requirements 1.6, 1.7**
