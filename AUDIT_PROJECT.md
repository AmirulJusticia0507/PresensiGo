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
