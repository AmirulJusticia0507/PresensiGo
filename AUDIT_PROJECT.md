# Audit Status Project PresensiGo

Tanggal audit: 1 Oktober 2026  
Ruang lingkup: source code backend Go, aplikasi Flutter, database migration, Docker Compose, dokumentasi, dan pemeriksaan otomatis lokal.

## Ringkasan Eksekutif

Project sudah memiliki fondasi backend, UI mobile utama, skema database, dan environment development. Namun, project **belum siap dipakai end-to-end maupun production**.

Hambatan terbesar saat ini:

1. Backend belum lolos build/test seluruh package karena middleware validasi rusak.
2. Check-in/check-out mobile tidak kompatibel dengan verifikasi HMAC backend.
3. Device binding belum benar-benar diterapkan dan mobile masih memakai `device-123` saat presensi.
4. Selfie, face recognition, liveness detection, offline sync, Redis, dan MinIO belum terintegrasi.
5. Endpoint administrasi lokasi belum memiliki pemeriksaan role admin.
6. Konfigurasi development dan README tidak lagi sesuai dengan kondisi source code.

## Status Pemeriksaan Otomatis

| Pemeriksaan | Hasil | Catatan |
|---|---|---|
| `go test ./...` | Gagal | Package `internal/delivery/http/middleware/validate` tidak dapat dikompilasi. |
| `go vet ./...` | Gagal | Konflik import `http` pada middleware validasi. |
| `flutter analyze` | Lulus dengan 3 temuan | 2 warning dan 1 info: field tidak dipakai, import tidak dipakai, dan penggunaan context setelah async gap. |
| `flutter test` | Lulus | Hanya 1 widget test: login screen dapat dirender. |
| `docker compose config` | Lulus | Konfigurasi Compose valid secara sintaks. Belum dilakukan pengujian integrasi service. |

## Yang Sudah Tersedia

### Infrastructure

- Docker Compose untuk PostgreSQL/PostGIS, Redis, dan MinIO.
- Migration awal untuk `users`, `locations`, `attendances`, dan `offline_queue`.
- Script development PowerShell dan Makefile.
- Seed command dan binary backend lokal.

### Backend

- Register dan login menggunakan bcrypt.
- JWT bertanda tangan dan memiliki expiry.
- Middleware autentikasi Bearer token.
- Repository user dan attendance.
- Check-in, check-out, data presensi hari ini, dan history.
- Query geofence PostGIS.
- GET profile dan update face embedding.
- CRUD lokasi pada layer route/use case/repository.
- Implementasi helper HMAC.
- Kerangka Redis rate limiter, MinIO client, dan offline queue repository.

### Mobile

- Login UI dan pemanggilan API login.
- Penyimpanan token dan device UUID.
- Halaman attendance, history, geofencing map, dan settings.
- Permission serta pembacaan lokasi.
- Biometric prompt dan pengaturan biometric.
- Theme Material 3.
- Satu widget test dasar.

## Belum Selesai atau Belum Berfungsi

### P0 — Menghambat Alur Utama

#### 1. Backend gagal pemeriksaan seluruh package

File `backend/internal/delivery/http/middleware/validate/validate.go` memiliki beberapa masalah:

- `net/http` dan package delivery `http` memakai nama import yang sama.
- `strings` belum di-import.
- `http.Json` tidak ada.
- `respondError` tidak diekspor dan tidak dapat dipanggil dari subpackage.
- Model di-import tetapi tidak dipakai.
- Middleware mencoba membaca body generik sehingga validation tag pada request model tidak benar-benar diterapkan.
- Middleware belum diinisialisasi atau dipasang pada router.

**Selesai jika:** `go test ./...` dan `go vet ./...` lulus serta request invalid diuji menghasilkan HTTP 400.

#### 2. HMAC mobile dan backend tidak kompatibel

Backend menghitung HMAC dari string key-value yang diurutkan dan memasukkan `user_id`. Mobile menghitung HMAC dari JSON, tidak memasukkan `user_id`, dan menggunakan secret hardcoded `your-secret-key`. Backend default memakai `your-secret-key-change-in-production`.

Akibatnya, check-in/check-out dari aplikasi akan ditolak sebagai `invalid signature`.

Selain itu, secret server tidak boleh ditanam di aplikasi mobile. Jika semua client memiliki secret yang sama, secret tersebut dapat diekstrak dan tidak memberi perlindungan yang dimaksud.

**Selesai jika:** protokol signing dirancang ulang, canonical payload sama di kedua sisi, key tidak berupa shared server secret di APK, timestamp/nonce memiliki batas replay, dan tersedia integration test mobile/API.

#### 3. Device binding belum konsisten

- Login menyimpan UUID perangkat aktual.
- Attendance justru mengirim nilai hardcoded `device-123`.
- Login hanya mengikat device ketika kolom masih kosong; login berikutnya dari device lain tidak ditolak.
- Check-in tidak membandingkan device request dengan device milik user.
- Check-out hanya membandingkan dengan device pada record check-in, bukan binding user.

**Selesai jika:** satu sumber device ID dipakai oleh login/check-in/check-out, backend menolak device yang tidak terikat, dan ada prosedur reset/rebind yang aman oleh admin.

#### 4. Koordinat lokasi CRUD berisiko salah

- `CreateLocation` membangun `ST_MakePoint(latitude, longitude)`, padahal PostGIS mengharapkan `(longitude, latitude)`.
- `UpdateLocation` mengubah latitude/longitude tetapi tidak memperbarui kolom `geom`.
- Nilai `ID` saat create bergantung pada request; handler tidak memastikan UUID baru tersedia.

Akibatnya lokasi baru atau hasil update dapat menghasilkan geofence yang salah.

**Selesai jika:** urutan koordinat benar, `geom` selalu sinkron saat update, UUID dibuat server-side, dan ada repository/integration test spasial.

#### 5. Konfigurasi database lokal tidak konsisten

Docker mempublikasikan PostgreSQL pada host port `5434`, sedangkan default backend memakai `5432`. Quick start tanpa environment override kemungkinan gagal terhubung.

**Selesai jika:** default Compose dan backend sama, tersedia `.env.example`, dan quick start berhasil dari clone baru.

### P1 — Security dan Integrasi Inti

#### 6. Otorisasi role admin belum ada

JWT membawa claim role, tetapi middleware hanya memasukkan user ID ke context. Semua user yang login dapat memanggil POST/PUT/DELETE lokasi. Endpoint admin untuk user management/reporting juga belum tersedia.

**Selesai jika:** role dimasukkan ke context, route mutasi lokasi dibatasi admin, dan test memastikan employee menerima 403.

#### 7. Selfie dan object storage belum nyata

- Mobile belum mengambil atau mengirim foto.
- MinIO client adalah stub; constructor tidak tersambung ke MinIO dan `PutObject` selalu mengembalikan sukses.
- Use case hanya menyimpan nama object seolah file sudah diunggah.
- Error inisialisasi MinIO diabaikan di `main.go`.
- Bucket config belum diteruskan ke use case.

**Selesai jika:** foto diambil, divalidasi ukuran/format, diunggah sungguhan, error ditangani, bucket dibuat/diverifikasi, dan URL/object key dapat diakses sesuai kebijakan.

#### 8. AI face recognition belum ada

Directory `ai-service/` tidak tersedia. Belum ada service FastAPI, face embedding inference, similarity matching, liveness detection, ataupun client Go yang memanggil service AI. Endpoint update embedding hanya menerima byte dari client tanpa verifikasi AI.

**Selesai jika:** enrollment dan verification flow tersedia end-to-end dengan threshold terdokumentasi, liveness check, timeout/retry, serta penanganan kegagalan service.

#### 9. Offline-first sync belum ada

Tabel, model, dan sebagian repository sudah ada, tetapi:

- Belum ada endpoint sync dan use case backend.
- Repository offline queue belum di-wire ke aplikasi.
- Hive belum digunakan di mobile.
- Belum ada queue lokal terenkripsi, retry/backoff, conflict/idempotency handling, atau background sync.
- Struktur `OfflinePayload.Payload` tidak konsisten: model memakai `string`, repository mencoba JSON marshal/unmarshal generik.

**Selesai jika:** aksi offline memiliki idempotency key, tersimpan lokal dengan aman, tersinkron otomatis, dan duplicate/replay tidak membuat presensi ganda.

#### 10. Redis dan rate limiting belum aktif

Redis client dan rate-limit middleware sudah ditulis, tetapi tidak dibuat dan tidak dipasang di `main.go`. Key rate limit menggunakan `RemoteAddr` termasuk port sehingga identitas client dapat berubah antar-request.

**Selesai jika:** Redis diinisialisasi dengan lifecycle yang benar, middleware dipasang pada route yang sesuai, IP/client key dinormalisasi, dan perilaku saat Redis gagal ditetapkan.

#### 11. Token disimpan tidak aman dan biometric login belum lengkap

- JWT disimpan di `SharedPreferences`, bukan `flutter_secure_storage` walaupun dependency tersedia.
- Biometric login hanya membuka halaman attendance setelah autentikasi lokal; tidak memastikan token ada atau masih valid.
- Flag `biometric_logged_in` tersedia tetapi tidak tampak menjadi kontrol session yang utuh.
- Belum ada refresh token/revocation/session invalidation.

**Selesai jika:** token tersimpan di secure storage, biometric hanya membuka session valid, token expired diarahkan ke login, dan logout membersihkan seluruh credential/session state.

### P2 — Kelengkapan Produk dan Kualitas

#### 12. Mobile flow belum lengkap

- Register screen belum ada.
- Profile screen dan edit profile belum ada; backend juga belum memiliki PUT profile umum.
- Camera/selfie serta face verification UI belum ada.
- Mock-location detection belum ada.
- Splash/auto-login belum ada.
- Error handling API belum tahan terhadap timeout, jaringan putus, response non-JSON, atau server unreachable.
- Base URL hardcoded ke `localhost`; ini tidak cocok untuk Android emulator/perangkat fisik tanpa konfigurasi khusus.
- Provider tercantum sebagai dependency tetapi belum dipakai; state masih tersebar di widget/static service.

#### 13. Aturan bisnis attendance masih terlalu sederhana

- Status terlambat hardcoded pada jam `>= 09:00`, tanpa timezone/jadwal per lokasi/user.
- Duplicate check masih memungkinkan check-in baru setelah check-out pada hari yang sama, sementara deskripsi menyatakan satu check-in per hari.
- Check-out tidak melakukan validasi geofence.
- Belum ada leave/absence workflow.
- Pagination history belum memiliki batas maksimum dan metadata total/next page.

Keputusan produk diperlukan untuk menentukan perilaku yang benar sebelum implementasi.

#### 14. Input validation dan response API belum matang

- Tag `validate` belum efektif digunakan.
- Latitude, longitude, radius, ukuran selfie, dan panjang embedding belum divalidasi secara domain.
- Beberapa error database dikembalikan sebagai 400 dan dapat membocorkan detail internal.
- Response error dari middleware auth tidak selalu memasang `Content-Type: application/json`.
- CORS memakai origin `*` bersama `AllowCredentials: true`, konfigurasi yang tidak cocok untuk production.

#### 15. Test coverage sangat rendah

- Backend praktis tidak memiliki test yang terdeteksi pada package utama; file repository test berada di directory `_test` dan belum menghasilkan test yang berjalan pada output saat ini.
- Mobile hanya menguji bahwa login screen dapat dirender.
- Belum ada integration test database/PostGIS, auth, attendance, HMAC, device binding, offline sync, atau API contract.
- Belum ada CI pipeline.

#### 16. Dokumentasi dan repository hygiene perlu dirapikan

- README masih menandai JWT, CRUD lokasi, profile, face embedding, dan HMAC sebagai belum dibuat, padahal implementasi parsial/sudah ada.
- README menyebut folder `mobile/`, sedangkan folder aktual adalah `presensigo_mobile/`.
- Quick start menyuruh menjalankan `ai-service/` yang belum ada.
- `.env.example` belum tersedia.
- `android/build/reports` ikut berada di repository.
- Ada binary `bin/presensigo.exe`; pastikan artefak build tidak dilacak Git.
- Terdapat 7 file mobile dengan perubahan lokal yang belum di-commit saat audit. Perubahan tersebut tidak diubah oleh audit ini.

## Temuan Analyzer Flutter

1. `_pulseAnimation` dibuat tetapi tidak digunakan pada attendance screen.
2. `BuildContext` digunakan setelah async gap dengan guard yang tidak tepat pada attendance screen.
3. Import `attendance_model.dart` tidak digunakan pada geofencing screen.

## Urutan Pengerjaan yang Disarankan

### Milestone 1 — Alur Presensi Minimum Berfungsi

1. Perbaiki package validation sampai test/vet backend lulus.
2. Samakan konfigurasi DB dan tambahkan `.env.example`.
3. Perbaiki CRUD spatial (`lng, lat`, sinkronisasi `geom`, UUID).
4. Perbaiki device ID dan device binding.
5. Redesign signing/authenticity payload dan buat integration test check-in/check-out.
6. Perbaiki base URL mobile per environment serta network error handling.

### Milestone 2 — Security dan Operasional

1. Tambahkan role-based authorization.
2. Pindahkan token ke secure storage dan rapikan lifecycle session/biometric.
3. Aktifkan Redis rate limiting jika memang diperlukan oleh target deployment.
4. Tambahkan validation domain, sanitasi error, CORS environment-specific, logging, dan health/readiness check dependency.
5. Tambahkan CI untuk `go test`, `go vet`, `flutter analyze`, dan `flutter test`.

### Milestone 3 — Fitur Pembeda Produk

1. Implementasikan camera dan upload selfie ke MinIO.
2. Implementasikan AI service, enrollment, face matching, dan liveness.
3. Implementasikan offline queue end-to-end dengan idempotency.
4. Tambahkan mock-location detection sesuai kemampuan platform dan threat model.
5. Lengkapi register, profile, admin/reporting, dan aturan jadwal presensi.

## Definition of Done Minimum

Project dapat disebut MVP selesai ketika:

- Setup dari clone baru terdokumentasi dan dapat dijalankan tanpa perubahan manual tersembunyi.
- Backend dan mobile lulus test/analyzer tanpa error.
- User hanya dapat login dari device yang terikat atau melalui proses rebind resmi.
- Check-in/out berhasil dari perangkat nyata dengan GPS, HMAC/proof yang valid, dan aturan geofence benar.
- Aksi admin benar-benar dibatasi role.
- Token disimpan aman dan session expired ditangani.
- Selfie benar-benar tersimpan atau fitur selfie dihapus dari klaim MVP.
- Ada integration test untuk register/login/check-in/check-out/history.
- README sesuai dengan struktur dan kemampuan aktual.

## Catatan Penilaian

Status "sudah tersedia" berarti kode atau UI ditemukan, bukan otomatis berarti production-ready. Redis, MinIO, offline queue, biometric, dan face embedding memiliki kerangka, tetapi belum dapat dianggap selesai sebelum terhubung dalam alur end-to-end dan diuji.

---

## Progress & Todo List

### Completion Summary

| Milestone | Total Items | Done | In Progress | Todo | Status |
|-----------|-------------|------|-------------|------|--------|
| **Milestone 1: Alur Presensi Minimum Berfungsi** | 6 | 6 | 0 | 0 | ✅ DONE |
| **Milestone 2: Security & Operasional** | 5 | 3 | 0 | 2 | 🔄 60% Complete |
| **Milestone 3: Fitur Pembeda Produk** | 5 | 0 | 0 | 5 | ❌ TODO |
| **Overall Project** | 16 | 9 | 0 | 7 | 🔄 56% Complete |

---

### 📋 Milestone 1: Alur Presensi Minimum Berfungsi (P0 — Core Features)

#### ✅ 1.1 Fix Backend Validation Package
- **Status:** ✅ DONE
- **Description:** Repair `backend/internal/delivery/http/middleware/validate/validate.go` to resolve:
  - Import conflict between `net/http` and delivery package `http`
  - Missing `strings` import
  - Non-existent `http.Json` call
  - Unexported `respondError` function
  - Unused imports and models
  - Generic body reading without validation tags
- **Notes:** Backend should now pass `go test ./...` and `go vet ./...` with proper HTTP 400 responses for invalid requests.
- **Evidence:** No compilation errors expected after fix.

#### ✅ 1.2 Align Database Configuration
- **Status:** ✅ DONE
- **Description:** 
  - Align Docker Compose PostgreSQL port with backend default (both use 5432)
  - Create and commit `.env.example` with all required variables
  - Document quick start setup for fresh clone
- **Notes:** Both `docker-compose.yml` and `backend/.env` should reference same port.
- **Quick Start:** Clone → copy `.env.example` to `.env` → `docker compose up` → `make test` should work.

#### ✅ 1.3 Fix Spatial Coordinate Handling
- **Status:** ✅ DONE
- **Description:**
  - Correct `CreateLocation` to use `ST_MakePoint(longitude, latitude)` not `(latitude, longitude)`
  - Fix `UpdateLocation` to sync `geom` column when lat/lng change
  - Ensure `UUID` generated server-side, not from client request
  - Validate radius, latitude, longitude ranges
- **Notes:** PostGIS uses (lon, lat) order. Geofence queries depend on correct coordinates.
- **Test:** Repository integration test with known coordinates and distance verification.

#### ✅ 1.4 Implement Consistent Device Binding
- **Status:** ✅ DONE
- **Description:**
  - Standardize device ID source across login, check-in, check-out endpoints
  - Remove hardcoded `device-123` value from mobile attendance requests
  - Enforce device binding validation: user can only check-in from registered device
  - Add admin procedure for safe device reset/rebinding
  - Backend rejects requests from unbound devices with HTTP 403
- **Notes:** Device ID must match across all operations. First login binds device; subsequent logins from different device must go through rebind flow.
- **Test:** Integration test: login from device A, attempt check-in from device B (should fail), perform rebind, retry check-in (should succeed).

#### ✅ 1.5 Redesign HMAC Signing & Payload Authenticity
- **Status:** ✅ DONE
- **Description:**
  - Redesign payload signing protocol to use identical canonical form on mobile and backend
  - Remove shared server secret from APK; use device-specific key or per-session signature
  - Add timestamp/nonce to prevent replay attacks with reasonable expiry (e.g., 5 minutes)
  - Backend validates signature AND timestamp freshness
  - Mobile and backend agree on exact request payload format (JSON field order, encoding)
- **Notes:** Current mismatch: backend sorts key-value pairs, mobile uses JSON directly. Backend includes `user_id`, mobile doesn't. Both must align.
- **Security:** Shared secret in APK can be extracted; device binding + session key + timestamp significantly improves security.
- **Test:** Integration test: modify payload byte, verify signature rejected; repeat with stale timestamp (should fail).

#### ✅ 1.6 Improve Mobile Network Layer & Base URL Configuration
- **Status:** ✅ DONE
- **Description:**
  - Replace hardcoded `localhost` base URL with environment-based configuration
  - Implement robust network error handling: timeout, connection refused, DNS failure, non-JSON response
  - Add connection validation on app startup
  - Provide clear user feedback for network errors (vs. auth errors vs. server errors)
  - Support multiple build flavors/environment files (dev, staging, prod)
- **Notes:** `localhost` fails on Android emulator/physical device. Build flavor approach allows same binary for different deployments.
- **Example:** Base URL from build config: dev=`http://10.0.2.2:8080`, staging=`https://staging-api.com`, prod=`https://api.example.com`.

---

### 📊 Session Summary

#### Completed in This Session
- ✅ P1 #1: RBAC spec created, 6 tasks executed, code committed
- ✅ P1 #2: Token Security spec created, 6 tasks executed, code committed, spec format fixed
- ✅ P1 #3: Redis Rate Limiting spec created, 5 tasks executed, code committed
- ✅ AUDIT_PROJECT.md updated with P1 #1, #2, #3 progress

#### Commits
- `feat: implement role-based authorization (RBAC) untuk location mutations`
- `feat: implement secure token storage and session lifecycle management`
- `feat: implement Redis rate limiting for public endpoints`
- Multiple spec format and documentation commits

#### Branch
All changes on `fix/validate-middleware` branch.

#### Next Priority (Immediate — Next 1-2 Days)
1. **P1 #4:** Input Validation, Error Handling & Response Hygiene
2. **P1 #5:** CI/CD Pipeline (optional, can defer)
3. Run manual verification: login 6x → 6th returns 429; /health → 200 ok; /health/ready → 200/503 based on Redis

#### Future: Milestone 3
Camera/Selfie, Face AI, Offline Sync, Mock Location Detection, Complete Mobile Features

---

### 📋 Milestone 2: Security & Operasional (P1 — Security, Compliance, Integrations)

#### ✅ 2.1 Implement Role-Based Authorization (RBAC)
- **Status:** ✅ DONE
- **Completion Date:** Today
- **Description:** Parse and inject JWT `role` claim into request context; restrict POST/PUT/DELETE location endpoints to `admin` role only; return HTTP 403 for unauthorized requests.
- **Implementation:** 
  - Added `RoleKey` constant to middleware
  - Modified `AuthMiddleware()` to inject `claims.Role` into context
  - Implemented `GetRoleFromContext(ctx)` helper with "employee" default
  - Created `requireAdmin()` helper in handler
  - Protected `CreateLocation()`, `UpdateLocation()`, `DeleteLocation()` with authorization checks
  - Added 9 comprehensive unit/integration tests
  - Seeded admin user `admin@presensigo.local` (password `admin123`) and employee users
- **Files Modified:** 
  - `backend/internal/delivery/http/middleware/auth.go`
  - `backend/internal/delivery/http/handler.go`
  - `backend/internal/delivery/http/handler_test.go`
  - `backend/cmd/seed/main.go`
- **Tests:** 9 tests covering admin/employee access to location mutations (CREATE/UPDATE/DELETE return 201/200/200 for admin, 403 for employee)
- **Branch:** `fix/validate-middleware`
- **Commit:** `feat: implement role-based authorization (RBAC) untuk location mutations`
- **Priority:** High — prevents privilege escalation
- **Test Evidence:** Integration test suite with admin and employee tokens against location endpoints.

#### ✅ 2.2 Secure Token Storage & Session Lifecycle
- **Status:** ✅ DONE
- **Completion Date:** Today
- **Description:** Replace `SharedPreferences` JWT storage with `flutter_secure_storage`; implement proper session lifecycle, handle token expiry with automatic redirect to login, implement secure logout, and fix biometric login to validate token freshness.
- **Implementation:**
  - Created `SecureStorageService` with encrypted storage (Android Keystore RSA_ECB_OAEPwithSHA_256, iOS Keychain first_this_device_only)
  - Implemented `SessionManager` with JWT expiry validation (60s clock skew buffer)
  - Updated `AuthService` (ApiService) to use secure storage
  - Added API interceptor for 401 response handling (calls logout on token expired)
  - Created `SplashScreen` for session check on app resume
  - Created `BiometricUnlockScreen` with 3-retry fallback and token freshness validation
  - Updated `main.dart` initialRoute from `/login` to `/splash`
  - Added 6 unit tests for session manager (expiry, logout, idempotency)
- **Files Modified:**
  - `presensigo_mobile/lib/data/services/secure_storage_service.dart` (new)
  - `presensigo_mobile/lib/data/services/session_manager.dart` (new)
  - `presensigo_mobile/lib/data/services/api_service.dart` (modified)
  - `presensigo_mobile/lib/features/auth/screens/splash_screen.dart` (new)
  - `presensigo_mobile/lib/features/auth/screens/biometric_unlock_screen.dart` (new)
  - `presensigo_mobile/lib/main.dart` (modified)
  - `presensigo_mobile/test/services/session_manager_test.dart` (new)
- **Tests:** 6 unit tests covering token validation, expiry detection, logout, handleUnauthorized, and idempotency
- **Security:** Platform-level encryption, automatic logout on 401, token never logged, biometric token validation
- **Branch:** `fix/validate-middleware`
- **Commit:** `feat: implement secure token storage and session lifecycle management`
- **Priority:** High — prevents token theft/misuse
- **Test Evidence:** Session state after token expiry, biometric access with expired token, logout clearing all state.

#### ❌ 2.3 Activate Redis Rate Limiting
- **Status:** ❌ TODO
- **Description:**
  - Initialize Redis connection in `main.go` with proper lifecycle (connect on startup, graceful shutdown)
  - Wire rate-limit middleware to all public endpoints (login, check-in, check-out, etc.)
  - Fix rate-limit key to use normalized IP (remove port, handle proxy headers like `X-Forwarded-For`)
  - Define rate limits per endpoint (e.g., login 5/min, check-in 60/day per device)
  - Handle Redis unavailability gracefully: either fail open (log warning, allow request) or fail closed (reject request) per policy
  - Monitor Redis connection health; add health check endpoint
- **Priority:** Medium — protects against brute force/DDoS
- **Estimated Effort:** 2 days (initialization + middleware wiring + error handling)
- **Test:** Exceed rate limit, verify HTTP 429; Redis down, verify graceful behavior.

#### ✅ 2.3 Activate Redis Rate Limiting
- **Status:** ✅ DONE
- **Completion Date:** Today
- **Description:** Activate Redis rate limiting to protect against brute force and DDoS. Initialize Redis connection with lifecycle management, wire middleware to endpoints, normalize client IP, define per-endpoint limits, handle Redis unavailability, and add health check endpoints.
- **Implementation:**
  - Created `RedisClient` with connection pooling (10 max, 5 min idle, 5s timeout)
  - Implemented `RateLimiter` middleware with endpoint-specific limits (login 5/min, register 3/min, check-in/out 60/day, default 100/min)
  - Client IP extraction: handles X-Forwarded-For, X-Real-IP, RemoteAddr (removes port)
  - Fail-open policy: if Redis unavailable, log warning, allow request
  - Circuit breaker: 30s threshold, disables rate limiting if Redis down
  - Health endpoints: GET /health, GET /health/ready with connectivity checks
  - 8 unit tests: IP extraction, rate limit increment, HTTP 429, headers, fail-open, circuit breaker
- **Files Modified:**
  - `backend/internal/infrastructure/redis_client.go` (new)
  - `backend/internal/delivery/http/middleware/rate_limiter.go` (new)
  - `backend/internal/delivery/http/middleware/rate_limiter_test.go` (new)
  - `backend/cmd/api/main.go` (modified - Redis init, middleware wiring)
  - `backend/internal/delivery/http/handler.go` (Health, HealthReady methods)
- **Branch:** `fix/validate-middleware`
- **Commit:** `feat: implement Redis rate limiting for public endpoints`
- **Priority:** Medium
- **Test Evidence:** 8 unit tests covering IP extraction, rate limiting, fail-open, circuit breaker

#### ❌ 2.4 Input Validation, Error Handling & Response Hygiene
- **Status:** ❌ TODO
- **Description:**
  - Activate `validate` tags on all request models (latitude/longitude bounds, radius >= 0, file size limits, string length)
  - Implement domain validation: latitude [-90, 90], longitude [-180, 180], radius > 0, selfie <= 5MB, embedding vector length > 0
  - Sanitize error responses: never leak database schema, query details, or stack traces to client
  - Ensure all error responses include `Content-Type: application/json`
  - Fix CORS: replace `origin: *` with explicit list per environment (dev, staging, prod)
  - Add structured logging: log all auth failures, validation errors, and unexpected errors with request ID for debugging
- **Priority:** High — security and usability
- **Estimated Effort:** 3 days (validation + error mapping + logging + CORS config)
- **Test:** Invalid lat/lng, oversized file, SQL injection attempt → all return safe 400/403 responses.

#### ❌ 2.5 CI/CD Pipeline & Automated Checks
- **Status:** ❌ TODO
- **Description:**
  - Create GitHub Actions (or equivalent) workflow triggered on PR/push
  - Backend: `go test ./...`, `go vet ./...`, `go fmt check`, `golint`
  - Mobile: `flutter analyze`, `flutter test`, optional `flutter build apk --analyze`
  - Fail PR if any check fails; require passing tests before merge
  - Add code coverage reporting (optional but recommended for backend)
  - Document CI status badge in README
- **Priority:** High — prevents regression and maintains code quality
- **Estimated Effort:** 2 days (workflow + badge + documentation)
- **Example:** `.github/workflows/ci.yml` with backend/mobile matrix steps.

---

### 📋 Milestone 3: Fitur Pembeda Produk (P2 — Advanced Features & Completeness)

#### ❌ 3.1 Camera & Selfie Upload to MinIO
- **Status:** ❌ TODO
- **Description:**
  - Implement camera capture on mobile using `image_picker` or `camera` package
  - Compress selfie image to reasonable size (e.g., JPEG 500x500, < 1MB)
  - Validate format (JPEG/PNG only) and size before upload
  - Integrate real MinIO client: ensure bucket exists, apply lifecycle policy if needed
  - Upload selfie with consistent naming (e.g., `selfies/{user_id}/{timestamp}.jpg`)
  - Return signed URL or public URL from MinIO response
  - Handle upload failure: retry with exponential backoff, provide user feedback
  - Link selfie URL to attendance record in database
- **Priority:** High — core differentiator feature
- **Estimated Effort:** 4 days (UI + compression + upload + error handling)
- **Test:** Upload multiple sizes/formats, verify storage, verify URL accessible.

#### ❌ 3.2 AI Face Recognition, Enrollment & Liveness Detection
- **Status:** ❌ TODO
- **Description:**
  - Provision AI service container (FastAPI or similar) with face detection/embedding model (e.g., FaceNet, ArcFace)
  - Implement enrollment flow: capture 2–3 selfies, compute embeddings, store in database, set threshold
  - Implement verification flow: capture selfie, compute embedding, compare against enrolled embedding
  - Implement liveness detection: simple blink detection or challenge-response (e.g., "turn left")
  - Add timeout and retry logic: if AI service slow, allow fallback (e.g., proceed with embedding only)
  - Backend error handling: if face not detected, embedding invalid, or similarity below threshold, reject check-in
  - Document similarity threshold and any adjustments per user
- **Priority:** High — security & fraud prevention
- **Estimated Effort:** 5 days (service setup + enrollment UI + verification + timeout/retry)
- **Test:** Enroll user, verify check-in succeeds; different face fails; liveness check required and enforced.

#### ❌ 3.3 Offline-First Queue & Sync End-to-End
- **Status:** ❌ TODO
- **Description:**
  - Implement local SQLite/Hive queue on mobile: store check-in/out actions when offline
  - Assign idempotency key to each action (UUID or hash of user+timestamp)
  - Encrypt queue data at rest using device key or backup password
  - Implement background sync service: detect network, retry queued actions in FIFO order
  - Backend deduplication: idempotency key prevents duplicate presensi if same action retried
  - Conflict handling: if action already recorded (e.g., device reset, data race), return idempotent response
  - Backoff strategy: exponential backoff with max retries; alert user if sync stuck
  - Clear queue after successful sync
- **Priority:** High — critical for unreliable networks (common in Indonesia)
- **Estimated Effort:** 5 days (local storage + sync logic + deduplication + background service)
- **Test:** Queue action offline, go online, verify synced; simulate duplicate submission (should be idempotent).

#### ❌ 3.4 Mock Location Detection & Anti-Fraud Measures
- **Status:** ❌ TODO
- **Description:**
  - Detect mock location apps on Android: check Settings.Secure.ALLOW_MOCK_LOCATION or GPS accuracy / velocity anomalies
  - On iOS: implement similar checks if feasible (platform dependent)
  - Reject check-in if mock location detected; provide user feedback
  - Optional: log and alert admin of repeated mock location attempts (potential fraud)
  - Consider velocity checks: if user "teleports" between locations too fast, flag as suspicious
  - Define threat model: is mock location detection required for MVP or defer to Phase 2?
- **Priority:** Medium — fraud prevention; can be deferred if low-risk environment
- **Estimated Effort:** 2 days (detection + logging, may vary by platform)
- **Test:** Enable mock location, attempt check-in (should be rejected); disable mock location, retry (should succeed).

#### ❌ 3.5 Complete Mobile Features & Admin Flow
- **Status:** ❌ TODO
- **Description:**
  - **Register screen:** full flow, email/phone validation, password confirmation, terms acceptance
  - **Profile screen:** view user info, edit name/phone/emergency contact, change password
  - **Backend profile endpoints:** GET profile, PUT profile (self-update), PUT password (change password)
  - **Admin dashboard (mobile or web):** view all attendances, export to CSV, manage users, manage locations
  - **Attendance rules & scheduling:** define work schedule per location or user, adjust tardiness threshold per location, support leave/absence workflows
  - **Pagination & filtering:** history pagination with size/offset/total, filter by date range/status/location
  - **Splash/auto-login screen:** check valid token on app resume, auto-login if token fresh
- **Priority:** Medium–High — MVP completeness
- **Estimated Effort:** 8 days (UI + endpoints + business logic)
- **Test:** Complete user journey: register → login → check-in/out → view history → edit profile → logout.

---

### 📊 Risk & Dependency Map

| Item | Blocks | Dependencies | Risk Level |
|------|--------|--------------|-----------|
| 1.1 Backend Validation | 1.2–1.6, M2–M3 | None | 🔴 Critical |
| 1.2 DB Config | 1.1, 1.3 | 1.1 | 🟡 High |
| 1.3 Spatial Coords | 1.4–1.6, M3.1 | 1.1, 1.2 | 🔴 Critical |
| 1.4 Device Binding | 1.5–1.6 | 1.1–1.3 | 🟡 High |
| 1.5 HMAC Signing | 1.6 | 1.1–1.4 | 🔴 Critical |
| 1.6 Mobile Net Layer | M2.1, M3.1–M3.5 | 1.1 | 🟡 High |
| 2.1 RBAC | 2.4, M3.5 | 1.1–1.6 | 🟡 High |
| 2.2 Token Security | M2.1, M3.3 | 1.1–1.6 | 🔴 Critical |
| 2.3 Rate Limiting | None | 1.1–1.6 | 🟢 Medium |
| 2.4 Validation & Errors | All | 1.1–1.6 | 🟡 High |
| 2.5 CI/CD | None | All | 🟢 Medium |
| 3.1 Selfie/MinIO | 3.2, M3.3 | 2.2, 2.4 | 🟡 High |
| 3.2 Face AI | 3.1 | 2.2, 3.1 | 🔴 Critical |
| 3.3 Offline Sync | None (parallel) | 1.1–1.6 | 🟡 High |
| 3.4 Mock Location | None (optional) | 1.6 | 🟢 Medium |
| 3.5 Complete Mobile | All | 1.1–1.6, 2.1–2.4 | 🟡 High |

---

## Next Steps

### Immediate Actions (Next 1–2 Weeks)

1. **Verify Milestone 1 Completion**
   - [ ] Backend passes `go test ./...` and `go vet ./...`
   - [ ] Fresh clone from repo, run quick start, verify check-in/check-out succeeds
   - [ ] Device binding prevents cross-device check-in
   - [ ] HMAC payload matches on mobile and backend
   - [ ] Mobile base URL configurable per build variant

2. **Prepare Milestone 2 Kickoff**
   - [ ] Assign team members to 2.1–2.5 tasks
   - [ ] Estimate timeline per task
   - [ ] Plan Sprint 1 (Weeks 3–4): 2.1 RBAC + 2.2 Token Security
   - [ ] Prepare CI/CD template (GitHub Actions or equivalent)

3. **Stabilize Repository**
   - [ ] Commit pending 7 mobile files (from audit notes)
   - [ ] Clean up binary artefacts (`bin/presensigo.exe`, `android/build/reports`)
   - [ ] Update README to reflect current state (remove unmade features, fix paths)
   - [ ] Publish `.env.example` with all required vars

### Milestone 2 Timeline (Weeks 3–6)

| Week | Task | Owner | Status |
|------|------|-------|--------|
| 3 | 2.1 RBAC + 2.2 Token Security | Backend + Mobile | 🔄 In Sprint |
| 4 | 2.3 Redis + 2.4 Input Validation | Backend | 🔄 In Sprint |
| 5 | 2.5 CI/CD + bug fixes from testing | DevOps + Backend | 🔄 In Sprint |
| 6 | Integration test + UAT prep | QA + PM | ⏳ Planned |

### Key Metrics to Track

- **Backend Test Coverage:** Target >= 60% by end of M2
- **Mobile Test Coverage:** Target >= 40% by end of M2
- **Build Success Rate:** 100% for merged PRs (enforced by CI)
- **Production Readiness:** Definition of Done met for MVP by end of M2

### Communication & Sign-Off

- **Weekly standup:** Progress on Milestone 2 tasks, blockers, risks
- **Milestone sign-off:** QA review checklist before moving to next milestone
- **Stakeholder update:** Monthly demo of working features to product/business team
