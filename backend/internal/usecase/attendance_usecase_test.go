package usecase

import (
	"encoding/base64"
	"strings"
	"testing"
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

func TestDecodeSelfieRejectsInvalidFormatAndSize(t *testing.T) {
	if _, _, _, err := decodeSelfie(base64.StdEncoding.EncodeToString([]byte("GIF89a"))); err == nil {
		t.Fatal("expected unsupported format to fail")
	}

	oversized := base64.StdEncoding.EncodeToString([]byte(strings.Repeat("x", maxSelfieSize+1)))
	if _, _, _, err := decodeSelfie(oversized); err == nil {
		t.Fatal("expected oversized image to fail")
	}
}
