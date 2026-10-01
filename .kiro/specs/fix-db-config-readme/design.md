# Design: Fix DB Config dan README

## Overview

Tiga perubahan independen yang saling tidak bergantung: (1) ubah satu baris default port di `config.go`, (2) pastikan `backend/.env` tidak ter-track Git, (3) update README agar akurat.

## Architecture

Tidak ada perubahan arsitektur. Ini adalah perubahan konfigurasi dan dokumentasi murni.

## Components and Interfaces

### config.go (diubah)

File: `backend/internal/config/config.go`

Ubah satu baris pada fungsi `Load()`:
```go
// SEBELUM
Port: getEnvInt("DB_PORT", 5432),

// SESUDAH
Port: getEnvInt("DB_PORT", 5434),
```

### .gitignore (diubah)

File: `.gitignore` di root project

Tambahkan entri `backend/.env` agar file `.env` di dalam subdirektori backend juga ter-ignore. Root `.gitignore` saat ini hanya memiliki `.env` tanpa path prefix, yang mungkin tidak mencakup `backend/.env` tergantung Git version.

### README.md (diubah)

File: `README.md` di root project

Perubahan:
1. Tech stack: "Fiber / Gin" → "net/http + gorilla/mux"
2. Checklist backend: tandai `[x]` untuk JWT auth, HMAC helper, input validation, location CRUD, profile endpoint, face embedding endpoint
3. Referensi path: `mobile/` → `presensigo_mobile/` di bagian Struktur Repositori
4. Quick Start step 3 (ai-service): tambah catatan "(belum tersedia, skip untuk saat ini)"

## Data Models

Tidak ada perubahan data model.

## Error Handling

Tidak ada error handling baru. Perubahan bersifat konfigurasi dan dokumentasi.

## Testing Strategy

Setelah perubahan `config.go`:
- Jalankan `go build ./...` untuk verifikasi tidak ada compile error.
- Jalankan `go test ./...` untuk pastikan tidak ada regresi.

Untuk `.gitignore`:
- Jalankan `git status` untuk verifikasi `backend/.env` tidak muncul sebagai tracked/modified.

## Correctness Properties

### Property 1: Default port sinkron dengan Docker
Default `DB_PORT` di `config.go` harus sama dengan host port yang di-publish Docker Compose untuk PostgreSQL.
**Validates: Requirements 1.1**

### Property 2: Secret tidak ter-track
`backend/.env` tidak boleh muncul di `git status` sebagai tracked file setelah perubahan `.gitignore`.
**Validates: Requirements 2.1, 2.2**

### Property 3: README akurat
Semua path folder dan status fitur di README harus mencerminkan kondisi aktual kode.
**Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5**
