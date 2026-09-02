
# Smart Attendance System

Sistem presensi modern berbasis geofencing dan face recognition real-time dengan kemampuan offline-first sync.

## Tech Stack Utama

| Layer                       | Teknologi                      | Alasan & Fungsi Utama                                                                                                                         |
| :-------------------------- | :----------------------------- | :-------------------------------------------------------------------------------------------------------------------------------------------- |
| **Frontend / Mobile** | Flutter (Dart)                 | Performa native,*cross-platform* (Android/iOS), serta dukungan plugin Geolocation & Kamera yang sangat stabil.                              |
| **Backend API**       | Go (Golang - Fiber / Gin)      | *Concurrency* tinggi (Goroutines) untuk menangani ribuan *request clock-in* bersamaan di jam sibuk dengan penggunaan memori yang minimal. |
| **Database**          | PostgreSQL + PostGIS           | Pengolahan data relasional tangguh yang dilengkapi ekstensi spatial query (`ST_DWithin`) untuk validasi radius geofencing secara akurat.    |
| **AI Inference**      | Python (FastAPI + InsightFace) | Microservice terpisah khusus pemrosesan liveness detection (anti-spoofing) dan pencocokan*face embedding*.                                  |
| **Cache & Broker**    | Redis                          | Caching session token, rate limiting API, dan*message queue* untuk proses async.                                                            |
| **Storage**           | MinIO / AWS S3                 | Penyimpanan objek (bukti foto selfie presensi).                                                                                               |

---

## Modul & Fitur Utama

### 1. Geofencing & Spatial Radius

Validasi lokasi presensi berbasis koordinat GPS kantor/site.

* **Mekanisme**: Go API memanggil fungsi PostGIS `ST_DWithin` untuk menghitung apakah koordinat pengguna berada dalam radius (misal: <= 50 meter) dari titik pusat kantor.
* **Security**: Deteksi *Mock Location* / *Fake GPS* langsung di layer native Flutter sebelum payload dikirim ke backend.

### 2. Face Recognition & Liveness Detection

Verifikasi identitas wajah beserta validasi objek asli (mencegah pemalsuan menggunakan foto/video).

* **Mekanisme**:
  1. Flutter App mengambil foto selfie dan mengecek pergerakan dasar (liveness check di client).
  2. Gambar dikirim ke Go Backend, lalu diteruskan (*gRPC/HTTP*) ke AI Service (Python).
  3. AI Engine mengubah wajah menjadi *512-d vector embedding* dan menghitung *Cosine Similarity* terhadap data master wajah.

### 3. Offline-First Sync Engine

Memungkinkan presensi tetap berjalan di area tanpa koneksi internet (*blankspot*).

* **Mekanisme**:
  * Data presensi disimpan sementara di penyimpanan lokal HP (Isar / Hive DB) secara terenkripsi.
  * Payload diberi stempel waktu dan ditandatangani kriptografi (HMAC-SHA256) untuk mencegah manipulasi jam HP oleh pengguna.
  * *Background service* di Flutter otomatis melakukan sinkronisasi ke Go API begitu koneksi internet terhubung kembali.

### 4. Device Binding & Fraud Prevention

Penguncian akun untuk mencegah penitipan presensi atau gonta-ganti perangkat.

* **Mekanisme**:
  * Setiap akun dikunci pada 1 `Device UUID` terdaftar.
  * Cross-check opsional dengan Mac Address / BSSID router WiFi kantor.

---

## Struktur Repositori

```text
├── mobile/                  # Flutter App Project
│   ├── lib/
│   │   ├── core/            # Location, Camera & Crypto Services
│   │   ├── features/        # Attendance, Profile, History
│   │   └── data/            # Local Storage (Hive/Isar) & Sync Engine
├── backend/                 # Go API Engine (Fiber / Gin)
│   ├── cmd/api/             # Entry point
│   ├── internal/
│   │   ├── delivery/        # HTTP Handlers / Middleware
│   │   ├── usecase/         # Business Logic (Attendance, Geofence)
│   │   └── repository/      # PostgreSQL (PostGIS) Queries
├── ai-service/              # Python FastAPI (Face Recognition Microservice)
├── docker-compose.yml       # Local Dev Stack (PostgreSQL/PostGIS, Redis, MinIO)
└── README.md
```



Alur Data Presensi (Sequence Flow)
Plaintext
[Flutter App]
     │
     ├── 1. Ambil GPS & Foto Wajah (Cek Mock GPS)
     ├── 2. Kirim Payload + Device UUID + HMAC Signature
     │
     ▼
[Go Backend API]
     │
     ├── 3. Validasi HMAC & Device Binding
     ├── 4. Spatial Query ke PostgreSQL (`ST_DWithin`) ──► [PostgreSQL/PostGIS]
     │
     ▼
[AI Service (Python)]
     │
     ├── 5. Liveness Check & Face Matching
     │
     ▼
[Go Backend API]
     │
     ├── 6. Simpan Foto ke MinIO / S3
     ├── 7. Insert Attendance Record ────────────────────► [PostgreSQL]
     └── 8. Invalidate Cache ────────────────────────────► [Redis]
Quick Start (Local Setup)

1. Jalankan Dependencies (Database & Storage)
   Bash
   docker-compose up -d
2. Jalankan Backend (Go)
   Bash
   cd backend
   go mod download
   go run cmd/api/main.go
3. Jalankan AI Microservice
   Bash
   cd ai-service
   pip install -r requirements.txt
   uvicorn main:app --port 8001
4. Jalankan Mobile App (Flutter)
   Bash
   cd mobile
   flutter pub get
   flutter run
