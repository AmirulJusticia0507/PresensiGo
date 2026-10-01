package usecase

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/PresensiGo/backend/internal/auth"
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
