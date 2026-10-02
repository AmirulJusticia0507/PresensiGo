package usecase

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/PresensiGo/backend/internal/auth"
	"github.com/PresensiGo/backend/internal/model"
)

func TestDecodeSelfie(t *testing.T) {
	jpeg := []byte{0xff, 0xd8, 0xff, 0xdb}
	data, contentType, extension, err := decodeSelfie(base64.StdEncoding.EncodeToString(jpeg))
	if err != nil {
		t.Fatalf("decodeSelfie returned error: %v", err)
	}
	if string(data) != string(jpeg) || contentType != "image/jpeg" || extension != "jpg" {
		t.Fatalf("unexpected decoded selfie: %v, %q, %q", data, contentType, extension)
	}
}

func TestVerifyOfflineAttendanceProof(t *testing.T) {
	timestamp := time.Now().Add(-time.Hour).Unix()
	payload := map[string]interface{}{
		"device_uuid": "device-id", "latitude": "-6.2",
		"longitude": "106.8", "timestamp": timestamp,
	}
	signature := auth.GenerateHMAC(payload, "device-id")
	if err := verifyAttendanceProof(payload, signature, "device-id", timestamp, true); err != nil {
		t.Fatalf("valid offline proof rejected: %v", err)
	}
	if err := verifyAttendanceProof(payload, "invalid", "device-id", timestamp, true); err == nil {
		t.Fatal("invalid offline signature accepted")
	}
	stale := time.Now().Add(-25 * time.Hour).Unix()
	if err := verifyAttendanceProof(payload, signature, "device-id", stale, true); err == nil {
		t.Fatal("stale offline action accepted")
	}
}

func TestDecodeSelfieRejectsInvalidFormatAndSize(t *testing.T) {
	if _, _, _, err := decodeSelfie(base64.StdEncoding.EncodeToString([]byte("GIF89a"))); err == nil {
		t.Fatal("expected unsupported format to fail")
	}

	oversized := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("x", maxSelfieSize+1)))
	if _, _, _, err := decodeSelfie(oversized); err == nil {
		t.Fatal("expected oversized image to fail")
	}
}

func TestHaversine(t *testing.T) {
	// Jakarta to Bandung is roughly 117 km.
	km := haversine(-6.2088, 106.8456, -6.9175, 107.6191)
	if km < 110 || km > 125 {
		t.Fatalf("expected ~117 km, got %.2f km", km)
	}

	if d := haversine(-6.2088, 106.8456, -6.2088, 106.8456); d != 0 {
		t.Fatalf("expected 0 km for identical coordinates, got %.4f", d)
	}
}

func TestImplausibleTravelSpeedKmh(t *testing.T) {
	t.Run("reports high speed for teleport", func(t *testing.T) {
		fiveMinutesAgo := time.Now().Add(-5 * time.Minute)
		previous := &model.Attendance{
			CheckOutTime:     &fiveMinutesAgo,
			CheckOutLocation: []float64{-6.2088, 106.8456},
		}
		speed, ok := implausibleTravelSpeedKmh(previous, -6.9175, 107.6191)
		if !ok {
			t.Fatal("expected teleport to be flagged")
		}
		if speed <= maxPlausibleSpeedKmh {
			t.Fatalf("expected speed above %.0f km/h, got %.2f", maxPlausibleSpeedKmh, speed)
		}
	})

	t.Run("ignores normal travel speed", func(t *testing.T) {
		twoHoursAgo := time.Now().Add(-2 * time.Hour)
		previous := &model.Attendance{
			CheckOutTime:     &twoHoursAgo,
			CheckOutLocation: []float64{-6.2088, 106.8456},
		}
		if _, ok := implausibleTravelSpeedKmh(previous, -6.25, 106.90); ok {
			t.Fatal("expected short commute to be accepted")
		}
	})

	t.Run("skips when position missing", func(t *testing.T) {
		oneHourAgo := time.Now().Add(-time.Hour)
		previous := &model.Attendance{CheckOutTime: &oneHourAgo}
		if _, ok := implausibleTravelSpeedKmh(previous, -6.9175, 107.6191); ok {
			t.Fatal("expected missing previous position to skip the check")
		}
	})

	t.Run("skips when check-out time is stale", func(t *testing.T) {
		threeDaysAgo := time.Now().Add(-72 * time.Hour)
		previous := &model.Attendance{
			CheckOutTime:     &threeDaysAgo,
			CheckOutLocation: []float64{-6.2088, 106.8456},
		}
		if _, ok := implausibleTravelSpeedKmh(previous, -6.9175, 107.6191); ok {
			t.Fatal("expected stale check-out to skip the check")
		}
	})

	t.Run("skips for non-positive time delta", func(t *testing.T) {
		future := time.Now().Add(time.Minute)
		previous := &model.Attendance{
			CheckOutTime:     &future,
			CheckOutLocation: []float64{-6.2088, 106.8456},
		}
		if _, ok := implausibleTravelSpeedKmh(previous, -6.9175, 107.6191); ok {
			t.Fatal("expected zero elapsed time to skip the check")
		}
	})
}
