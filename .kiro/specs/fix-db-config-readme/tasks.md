# Implementation Plan: Fix DB Config dan README

## Overview

Tiga perbaikan konfigurasi dan dokumentasi: sinkronkan default port DB, amankan .env dari Git, dan update README agar akurat.

## Task Dependency Graph

```json
{
  "waves": [
    { "wave": 1, "tasks": ["1", "2", "3"] },
    { "wave": 2, "tasks": ["4"] },
    { "wave": 3, "tasks": ["5"] }
  ]
}
```

Task 1, 2, dan 3 dapat dikerjakan paralel. Task 4 verifikasi build setelah ketiganya selesai. Task 5 commit dan push.

## Tasks

- [ ] 1. Fix default DB_PORT di config.go
  - Buka `backend/internal/config/config.go`
  - Ubah `getEnvInt("DB_PORT", 5432)` menjadi `getEnvInt("DB_PORT", 5434)`
  - Pastikan nilai default lain konsisten dengan `.env.example`
  - **Files:** `backend/internal/config/config.go`

- [ ] 2. Pastikan backend/.env tidak ter-track Git
  - Periksa apakah `backend/.env` saat ini ter-track di Git dengan `git ls-files backend/.env`
  - Jika ter-track: jalankan `git rm --cached backend/.env`
  - Tambahkan `backend/.env` ke `.gitignore` root (sudah ada entri `.env` tapi pastikan juga cover `backend/.env`)
  - Verifikasi dengan `git status` bahwa file tidak muncul sebagai tracked
  - **Files:** `.gitignore`

- [ ] 3. Update README.md agar akurat
  - Ubah tech stack backend: "Fiber / Gin" → "net/http + gorilla/mux"
  - Update checklist backend — tandai `[x]` untuk: JWT authentication, HMAC-SHA256 verification (helper sudah ada), input validation, location CRUD (POST/PUT/DELETE handler sudah ada), profile endpoint, face embedding endpoint
  - Ubah referensi `mobile/` → `presensigo_mobile/` di bagian Struktur Repositori
  - Tambahkan catatan pada Quick Start step 3 bahwa `ai-service/` belum tersedia: `(opsional — AI service belum tersedia, skip langkah ini)`
  - Update nama folder di struktur repositori dari `mobile/` menjadi `presensigo_mobile/`
  - **Files:** `README.md`

- [ ] 4. Verifikasi go build dan go test lulus
  - Jalankan `go build ./...` dari direktori `backend/`
  - Jalankan `go test ./...` dari direktori `backend/`
  - Kedua perintah harus lulus tanpa error
  - **Files:** (verifikasi saja)

- [ ] 5. Commit dan push perubahan
  - `git add` file yang berubah: `backend/internal/config/config.go`, `.gitignore`, `README.md`
  - Commit dengan pesan: `fix: sinkronkan default DB port, amankan .env, dan update README`
  - Push ke branch `fix/validate-middleware` (branch yang sedang aktif)
  - **Files:** (git operations)

## Notes

- `backend/.env` mungkin sudah ter-ignore oleh entri `.env` di `.gitignore` root — cek dulu dengan `git ls-files` sebelum menambahkan entri baru.
- Perubahan README bersifat deskriptif, tidak mempengaruhi runtime.
- Task ini tidak memerlukan database aktif untuk verifikasi.
