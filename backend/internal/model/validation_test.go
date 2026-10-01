package model

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

var v = validator.New()

// TestCreateLocationRequest_ValidInput verifies that valid location data passes validation
func TestCreateLocationRequest_ValidInput(t *testing.T) {
	req := CreateLocationRequest{
		Name:         "Office",
		Latitude:     6.2,
		Longitude:    106.8,
		RadiusMeters: 50,
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected valid location to pass validation: %v", err)
	}
}

// TestCreateLocationRequest_MissingName verifies that missing name is rejected
func TestCreateLocationRequest_MissingName(t *testing.T) {
	req := CreateLocationRequest{
		Name:         "",
		Latitude:     6.2,
		Longitude:    106.8,
		RadiusMeters: 50,
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for missing name")
	}
}

// TestCreateLocationRequest_LatitudeOutOfRange_High verifies latitude > 90 is rejected
func TestCreateLocationRequest_LatitudeOutOfRange_High(t *testing.T) {
	req := CreateLocationRequest{
		Name:         "Office",
		Latitude:     91,
		Longitude:    106.8,
		RadiusMeters: 50,
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for latitude > 90")
	}
}

// TestCreateLocationRequest_LatitudeOutOfRange_Low verifies latitude < -90 is rejected
func TestCreateLocationRequest_LatitudeOutOfRange_Low(t *testing.T) {
	req := CreateLocationRequest{
		Name:         "Office",
		Latitude:     -91,
		Longitude:    106.8,
		RadiusMeters: 50,
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for latitude < -90")
	}
}

// TestCreateLocationRequest_LatitudeEdgeCase_Minimum verifies latitude = -90 passes
func TestCreateLocationRequest_LatitudeEdgeCase_Minimum(t *testing.T) {
	req := CreateLocationRequest{
		Name:         "South Pole",
		Latitude:     -90,
		Longitude:    0,
		RadiusMeters: 50,
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected latitude = -90 to pass: %v", err)
	}
}

// TestCreateLocationRequest_LatitudeEdgeCase_Maximum verifies latitude = 90 passes
func TestCreateLocationRequest_LatitudeEdgeCase_Maximum(t *testing.T) {
	req := CreateLocationRequest{
		Name:         "North Pole",
		Latitude:     90,
		Longitude:    0,
		RadiusMeters: 50,
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected latitude = 90 to pass: %v", err)
	}
}

// TestCreateLocationRequest_LongitudeOutOfRange_High verifies longitude > 180 is rejected
func TestCreateLocationRequest_LongitudeOutOfRange_High(t *testing.T) {
	req := CreateLocationRequest{
		Name:         "Office",
		Latitude:     6.2,
		Longitude:    181,
		RadiusMeters: 50,
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for longitude > 180")
	}
}

// TestCreateLocationRequest_LongitudeOutOfRange_Low verifies longitude < -180 is rejected
func TestCreateLocationRequest_LongitudeOutOfRange_Low(t *testing.T) {
	req := CreateLocationRequest{
		Name:         "Office",
		Latitude:     6.2,
		Longitude:    -181,
		RadiusMeters: 50,
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for longitude < -180")
	}
}

// TestCreateLocationRequest_LongitudeEdgeCase_Minimum verifies longitude = -180 passes
func TestCreateLocationRequest_LongitudeEdgeCase_Minimum(t *testing.T) {
	req := CreateLocationRequest{
		Name:         "Office",
		Latitude:     6.2,
		Longitude:    -180,
		RadiusMeters: 50,
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected longitude = -180 to pass: %v", err)
	}
}

// TestCreateLocationRequest_LongitudeEdgeCase_Maximum verifies longitude = 180 passes
func TestCreateLocationRequest_LongitudeEdgeCase_Maximum(t *testing.T) {
	req := CreateLocationRequest{
		Name:         "Office",
		Latitude:     6.2,
		Longitude:    180,
		RadiusMeters: 50,
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected longitude = 180 to pass: %v", err)
	}
}

// TestCreateLocationRequest_RadiusZero verifies radius = 0 is rejected
func TestCreateLocationRequest_RadiusZero(t *testing.T) {
	req := CreateLocationRequest{
		Name:         "Office",
		Latitude:     6.2,
		Longitude:    106.8,
		RadiusMeters: 0,
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for radius = 0")
	}
}

// TestCreateLocationRequest_RadiusNegative verifies negative radius is rejected
func TestCreateLocationRequest_RadiusNegative(t *testing.T) {
	req := CreateLocationRequest{
		Name:         "Office",
		Latitude:     6.2,
		Longitude:    106.8,
		RadiusMeters: -5,
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for negative radius")
	}
}

// TestCreateLocationRequest_RadiusOne verifies radius = 1 passes
func TestCreateLocationRequest_RadiusOne(t *testing.T) {
	req := CreateLocationRequest{
		Name:         "Office",
		Latitude:     6.2,
		Longitude:    106.8,
		RadiusMeters: 1,
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected radius = 1 to pass: %v", err)
	}
}

// TestCheckInRequest_ValidInput verifies valid check-in request passes validation
func TestCheckInRequest_ValidInput(t *testing.T) {
	req := CheckInRequest{
		Latitude:   6.2,
		Longitude:  106.8,
		DeviceUUID: "123e4567-e89b-12d3-a456-426614174000",
		Timestamp:  1234567890,
		HMACSig:    "signature_data",
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected valid check-in to pass validation: %v", err)
	}
}

// TestCheckInRequest_LatitudeOutOfRange verifies latitude validation in check-in
func TestCheckInRequest_LatitudeOutOfRange(t *testing.T) {
	req := CheckInRequest{
		Latitude:   100,
		Longitude:  106.8,
		DeviceUUID: "123e4567-e89b-12d3-a456-426614174000",
		Timestamp:  1234567890,
		HMACSig:    "signature_data",
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for invalid latitude")
	}
}

// TestCheckInRequest_InvalidDeviceUUID verifies UUID validation
func TestCheckInRequest_InvalidDeviceUUID(t *testing.T) {
	req := CheckInRequest{
		Latitude:   6.2,
		Longitude:  106.8,
		DeviceUUID: "not-a-uuid",
		Timestamp:  1234567890,
		HMACSig:    "signature_data",
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for invalid UUID")
	}
}

// TestUploadSelfieRequest_ValidInput verifies valid selfie request passes validation
func TestUploadSelfieRequest_ValidInput(t *testing.T) {
	req := UploadSelfieRequest{
		FileSize: 1000000, // 1MB
		Format:   "jpg",
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected valid selfie request to pass validation: %v", err)
	}
}

// TestUploadSelfieRequest_FileSize5MB verifies 5MB file is accepted
func TestUploadSelfieRequest_FileSize5MB(t *testing.T) {
	req := UploadSelfieRequest{
		FileSize: 5242880, // Exactly 5MB
		Format:   "jpeg",
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected 5MB file to pass validation: %v", err)
	}
}

// TestUploadSelfieRequest_FileSizeExceeded verifies file > 5MB is rejected
func TestUploadSelfieRequest_FileSizeExceeded(t *testing.T) {
	req := UploadSelfieRequest{
		FileSize: 5242881, // 5MB + 1 byte
		Format:   "png",
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for file > 5MB")
	}
}

// TestUploadSelfieRequest_FileZeroBytes verifies 0-byte file is rejected
func TestUploadSelfieRequest_FileZeroBytes(t *testing.T) {
	req := UploadSelfieRequest{
		FileSize: 0,
		Format:   "jpg",
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for 0-byte file")
	}
}

// TestUploadSelfieRequest_InvalidFormat verifies invalid format is rejected
func TestUploadSelfieRequest_InvalidFormat(t *testing.T) {
	req := UploadSelfieRequest{
		FileSize: 1000000,
		Format:   "gif",
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for invalid format")
	}
}

// TestUploadSelfieRequest_JPEGFormat verifies jpeg format is accepted
func TestUploadSelfieRequest_JPEGFormat(t *testing.T) {
	req := UploadSelfieRequest{
		FileSize: 1000000,
		Format:   "jpeg",
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected jpeg format to pass: %v", err)
	}
}

// TestUploadSelfieRequest_PNGFormat verifies png format is accepted
func TestUploadSelfieRequest_PNGFormat(t *testing.T) {
	req := UploadSelfieRequest{
		FileSize: 1000000,
		Format:   "png",
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected png format to pass: %v", err)
	}
}

// TestUpdateEmbeddingRequest_ValidInput verifies valid embedding request passes validation
func TestUpdateEmbeddingRequest_ValidInput(t *testing.T) {
	req := UpdateEmbeddingRequest{
		Embedding: []float32{0.1, 0.2, 0.3, 0.4},
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected valid embedding to pass validation: %v", err)
	}
}

// TestUpdateEmbeddingRequest_SingleElement verifies embedding with 1 element passes
func TestUpdateEmbeddingRequest_SingleElement(t *testing.T) {
	req := UpdateEmbeddingRequest{
		Embedding: []float32{0.5},
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected single element embedding to pass: %v", err)
	}
}

// TestUpdateEmbeddingRequest_EmptyEmbedding verifies empty embedding is rejected
func TestUpdateEmbeddingRequest_EmptyEmbedding(t *testing.T) {
	req := UpdateEmbeddingRequest{
		Embedding: []float32{},
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for empty embedding")
	}
}

// TestUpdateEmbeddingRequest_ManyElements verifies large embedding vectors pass
func TestUpdateEmbeddingRequest_ManyElements(t *testing.T) {
	// Create 512-dimensional embedding
	embedding := make([]float32, 512)
	for i := range embedding {
		embedding[i] = 0.5
	}

	req := UpdateEmbeddingRequest{
		Embedding: embedding,
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected large embedding to pass: %v", err)
	}
}

// TestCheckOutRequest_ValidInput verifies valid check-out request passes validation
func TestCheckOutRequest_ValidInput(t *testing.T) {
	req := CheckOutRequest{
		Latitude:   6.2,
		Longitude:  106.8,
		DeviceUUID: "123e4567-e89b-12d3-a456-426614174000",
		Timestamp:  1234567890,
		HMACSig:    "signature_data",
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected valid check-out to pass validation: %v", err)
	}
}

// TestCheckOutRequest_LongitudeOutOfRange verifies longitude validation in check-out
func TestCheckOutRequest_LongitudeOutOfRange(t *testing.T) {
	req := CheckOutRequest{
		Latitude:   6.2,
		Longitude:  200,
		DeviceUUID: "123e4567-e89b-12d3-a456-426614174000",
		Timestamp:  1234567890,
		HMACSig:    "signature_data",
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for longitude out of range")
	}
}

// TestLoginRequest_ValidInput verifies valid login request passes validation
func TestLoginRequest_ValidInput(t *testing.T) {
	req := LoginRequest{
		Email:      "user@example.com",
		Password:   "securePassword123",
		DeviceUUID: "123e4567-e89b-12d3-a456-426614174000",
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected valid login to pass validation: %v", err)
	}
}

// TestLoginRequest_InvalidEmail verifies invalid email is rejected
func TestLoginRequest_InvalidEmail(t *testing.T) {
	req := LoginRequest{
		Email:      "not-an-email",
		Password:   "securePassword123",
		DeviceUUID: "123e4567-e89b-12d3-a456-426614174000",
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for invalid email")
	}
}

// TestLoginRequest_InvalidDeviceUUID verifies invalid device UUID is rejected
func TestLoginRequest_InvalidDeviceUUID(t *testing.T) {
	req := LoginRequest{
		Email:      "user@example.com",
		Password:   "securePassword123",
		DeviceUUID: "invalid-uuid",
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for invalid device UUID")
	}
}

// TestRegisterRequest_ValidInput verifies valid register request passes validation
func TestRegisterRequest_ValidInput(t *testing.T) {
	req := RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "securePassword123",
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected valid register to pass validation: %v", err)
	}
}

// TestRegisterRequest_PasswordTooShort verifies password < 6 chars is rejected
func TestRegisterRequest_PasswordTooShort(t *testing.T) {
	req := RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "pass",
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for short password")
	}
}

// TestRegisterRequest_PasswordMinimum verifies password = 6 chars passes
func TestRegisterRequest_PasswordMinimum(t *testing.T) {
	req := RegisterRequest{
		Name:     "John Doe",
		Email:    "john@example.com",
		Password: "pass12",
	}

	if err := v.Struct(req); err != nil {
		t.Errorf("expected 6-char password to pass: %v", err)
	}
}

// TestRegisterRequest_InvalidEmail verifies invalid email is rejected
func TestRegisterRequest_InvalidEmail(t *testing.T) {
	req := RegisterRequest{
		Name:     "John Doe",
		Email:    "not-an-email",
		Password: "securePassword123",
	}

	if err := v.Struct(req); err == nil {
		t.Error("expected validation error for invalid email")
	}
}
