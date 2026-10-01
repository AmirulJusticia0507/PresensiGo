# Requirements Document

## Introduction

Konfigurasi default database di `config.go` masih menggunakan port `5432`, sedangkan Docker Compose mempublikasikan PostgreSQL ke host port `5434`. Developer yang menjalankan backend tanpa file `.env` akan gagal terhubung ke database. Selain itu, README tidak mencerminkan struktur dan status aktual project, dan `backend/.env` harus dipastikan tidak ter-track Git.

## Requirements

### 1. Sinkronkan Default Port Database

**User Story:** Sebagai developer baru, saya ingin bisa menjalankan backend dari clone baru tanpa konfigurasi manual tambahan.

#### Acceptance Criteria

- 1.1: Default fallback `DB_PORT` di `config.go` diubah dari `5432` menjadi `5434` agar sesuai dengan Docker Compose.
- 1.2: Nilai default lain di `config.go` (host, user, password, name) konsisten dengan nilai di `.env.example`.
- 1.3: `go build ./...` tetap lulus setelah perubahan.

### 2. Pastikan backend/.env Tidak Ter-track Git

**User Story:** Sebagai developer, saya ingin file `.env` yang berisi secret tidak ter-commit ke repository.

#### Acceptance Criteria

- 2.1: `backend/.env` ditambahkan ke `.gitignore` root atau `backend/.gitignore` sehingga tidak ter-track.
- 2.2: Jika `backend/.env` sudah ter-track di Git, file tersebut di-untrack dengan `git rm --cached`.
- 2.3: `backend/.env.example` tetap tersedia sebagai template.

### 3. Update README Agar Sesuai Kondisi Aktual

**User Story:** Sebagai developer baru, saya ingin README menggambarkan struktur dan kemampuan aktual project sehingga saya tidak dikelirukan.

#### Acceptance Criteria

- 3.1: Referensi folder `mobile/` di README diubah menjadi `presensigo_mobile/`.
- 3.2: Bagian Quick Start tidak lagi menyuruh menjalankan `ai-service/` yang belum ada — langkah tersebut diberi catatan bahwa AI service belum tersedia.
- 3.3: Checklist backend diperbarui: JWT authentication, HMAC, CRUD lokasi, profile endpoint, face embedding endpoint, dan input validation ditandai `[x]` karena sudah ada implementasinya.
- 3.4: Struktur repositori di README menggunakan nama folder aktual `presensigo_mobile/` bukan `mobile/`.
- 3.5: Tech stack backend dikoreksi dari "Fiber / Gin" menjadi "net/http + gorilla/mux" sesuai implementasi aktual.

## Glossary

- **fallback value**: Nilai default yang digunakan `config.go` jika environment variable tidak di-set.
- **untrack**: Menghapus file dari Git index tanpa menghapus file dari filesystem, menggunakan `git rm --cached`.
