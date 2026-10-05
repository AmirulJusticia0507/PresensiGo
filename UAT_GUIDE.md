# PresensiGo UAT Guide — Local Physical Device Testing

## Prerequisites

### Hardware
- Laptop/PC running Windows with Docker Desktop installed
- Android physical device (Android 8.0+) on the same WiFi network
- USB cable or ADB over WiFi enabled

### Software
- Docker Desktop running
- Go 1.21+ installed
- Flutter SDK installed
- Android Studio or ADB tools

---

## Step 1: Find Your Local IP Address

Run this in Command Prompt:
```
ipconfig
```
Look for **IPv4 Address** under your WiFi adapter (e.g., `192.168.1.100`).

---

## Step 2: Start Backend Services

```bash
# From project root
docker compose up -d postgres redis minio

# Wait for services to be healthy
docker compose ps

# Expected output: all services "healthy"
```

---

## Step 3: Run Database Migrations

```bash
cd backend
go run cmd/api/main.go
```

The server will auto-apply migrations on startup. Wait for:
```
✓ Connected to database
Server starting on port 8088
```

Leave this terminal open.

---

## Step 4: Seed Admin User

In a new terminal:
```bash
cd backend
go run cmd/seed/main.go
```

This creates:
- Admin: `admin@presensigo.local` / `admin123`
- Employee: `employee@presensigo.local` / `employee123`

---

## Step 5: Build & Install APK on Android Device

Replace `192.168.1.100` with your actual local IP from Step 1:

```bash
cd presensigo_mobile

flutter build apk --debug \
  --dart-define=API_BASE_URL=http://192.168.1.100:8088/api

# Install to connected device
flutter install
```

Or run directly on connected device:
```bash
flutter run --dart-define=API_BASE_URL=http://192.168.1.100:8088/api
```

---

## Step 6: UAT Test Cases

### 🔐 Auth Flow

| # | Test Case | Expected | Pass/Fail |
|---|-----------|----------|-----------|
| 1 | Register new user with valid data | Success, navigate to home | |
| 2 | Register with existing email | Error: "Email already registered" | |
| 3 | Register with weak password | Error: password requirements shown | |
| 4 | Login with admin credentials | Success | |
| 5 | Login with wrong password | Error message shown | |
| 6 | Logout | Redirect to login screen | |
| 7 | Re-open app after logout | Shows login screen (not auto-login) | |
| 8 | Re-open app while logged in | Shows splash → home (auto-login) | |

### 👤 Profile Flow

| # | Test Case | Expected | Pass/Fail |
|---|-----------|----------|-----------|
| 9 | View profile | All fields displayed | |
| 10 | Edit name & phone | Success, fields updated | |
| 11 | Edit with invalid phone | Error: "Invalid phone format" | |
| 12 | Change password (correct current) | Success message | |
| 13 | Change password (wrong current) | Generic error (no detail leak) | |
| 14 | Change password (weak new password) | Error: requirements shown | |

### 📍 Attendance Flow

| # | Test Case | Expected | Pass/Fail |
|---|-----------|----------|-----------|
| 15 | Check-in within geofence | Success | |
| 16 | Check-in outside geofence | Rejected with location error | |
| 17 | Check-in with mock GPS | Rejected: "Mock location detected" | |
| 18 | View today's attendance | Card shows check-in time | |
| 19 | Check-out | Success, time recorded | |
| 20 | Attempt check-in twice same day | Rejected: already checked in | |

### 📋 History Flow

| # | Test Case | Expected | Pass/Fail |
|---|-----------|----------|-----------|
| 21 | View attendance history | List of past attendances | |
| 22 | Filter by status "late" | Only late records shown | |
| 23 | Pull to refresh | List reloads | |
| 24 | Load more (if >20 records) | More records append | |

### 🏖️ Leave Flow

| # | Test Case | Expected | Pass/Fail |
|---|-----------|----------|-----------|
| 25 | Request sick leave | Submitted, status: pending | |
| 26 | View my leave requests | List of submitted requests | |

### 🛡️ Admin Flow (login as admin first)

| # | Test Case | Expected | Pass/Fail |
|---|-----------|----------|-----------|
| 27 | Open Admin Dashboard | 5 tabs visible | |
| 28 | View all attendances | All employee attendance listed | |
| 29 | Filter attendance by status | Filtered results shown | |
| 30 | Export CSV | CSV data copied to clipboard | |
| 31 | View all users | User list with role & email | |
| 32 | Toggle employee ↔ admin role | Role changes | |
| 33 | Reset device binding | Device UUID cleared | |
| 34 | View/create locations | Location appears in list | |
| 35 | Create work schedule | Schedule saved | |
| 36 | Approve leave request | Status changes to "approved" | |
| 37 | Reject leave request | Status changes to "rejected" | |

### 🔒 Security / Edge Cases

| # | Test Case | Expected | Pass/Fail |
|---|-----------|----------|-----------|
| 38 | Register 4x in 1 minute | 4th attempt: HTTP 429 rate limit | |
| 39 | Access /api/profile without token | 401 Unauthorized | |
| 40 | Employee access admin dashboard | 403 Forbidden on admin API calls | |
| 41 | Kill app + reopen with valid token | Auto-login to home screen | |
| 42 | Let token expire (change JWT_EXPIRE_HOUR=0.001 in .env) | Redirect to login | |

---

## Step 7: Verify Backend Health

```bash
curl http://localhost:8088/health
# Expected: {"status":"ok"}

curl http://localhost:8088/health/ready
# Expected: {"ready":true}
```

---

## Step 8: Check Logs

```bash
# Backend logs (in the terminal running go run)
# Look for: request IDs, rate limit hits, auth failures

# Docker service logs
docker compose logs redis
docker compose logs postgres
```

---

## Known Limitations for Local UAT

- **AI face recognition**: Requires `docker compose --profile full up` — AI model download takes ~5 min on first run
- **MinIO selfie storage**: Works but selfie URLs are local only (`localhost:9000`)
- **Offline sync**: Requires disabling WiFi then re-enabling — test manually
- **Push notifications**: Not implemented in MVP

---

## Troubleshooting

| Problem | Solution |
|---------|----------|
| App can't connect to backend | Check IP in `--dart-define`, ensure same WiFi network |
| "Connection refused" | Ensure `go run cmd/api/main.go` is running |
| "Database not ready" | Run `docker compose ps` — wait for postgres to be healthy |
| APK build fails | Run `flutter pub get` first |
| Device not detected | Enable USB debugging in Developer Options |
| Face enrollment fails | AI service not running — run `docker compose --profile full up -d ai-service` |

---

## Sign-Off Checklist

Before declaring UAT complete:

- [ ] All 42 test cases above executed
- [ ] No crashes or unhandled exceptions observed
- [ ] Logs show no unexpected errors (only expected 400/401/403/429)
- [ ] APK installs and runs on physical device without issues
- [ ] Admin flow fully functional
- [ ] Security test cases (38-42) pass
