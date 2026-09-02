package usecase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/PresensiGo/backend/internal/config"
	"github.com/PresensiGo/backend/internal/model"
	"github.com/PresensiGo/backend/internal/repository"
)

type AttendanceUsecase struct {
	attRepo  *repository.AttendanceRepository
	config   *config.Config
}

func NewAttendanceUsecase(attRepo *repository.AttendanceRepository, cfg *config.Config) *AttendanceUsecase {
	return &AttendanceUsecase{
		attRepo: attRepo,
		config:  cfg,
	}
}

func (u *AttendanceUsecase) CheckIn(userID uuid.UUID, req *model.CheckInRequest) (*model.Attendance, error) {
	// Verify HMAC signature
	if !u.verifyHMAC(req, req.HMACSig) {
		return nil, errors.New("invalid signature")
	}

	// Find nearest location
	location, err := u.attRepo.FindNearestLocation(req.Latitude, req.Longitude)
	if err != nil {
		return nil, errors.New("no location found nearby")
	}

	// Check geofence
	inside, err := u.attRepo.CheckGeofence(location.ID, req.Latitude, req.Longitude)
	if err != nil {
		return nil, err
	}
	if !inside {
		return nil, errors.New("you are outside the geofence radius")
	}

	// Check if already checked in today
	existing, _ := u.attRepo.FindTodayByUser(userID)
	if existing != nil && existing.CheckOutTime == nil {
		return nil, errors.New("already checked in today")
	}

	now := time.Now()
	isLate := now.Hour() >= 9 // Assume late after 9 AM

	att := &model.Attendance{
		ID:              uuid.New(),
		UserID:          userID,
		LocationID:      location.ID,
		CheckInTime:     &now,
		CheckInLocation: []float64{req.Latitude, req.Longitude},
		Status:          "present",
		IsLate:          isLate,
		DeviceUUID:      req.DeviceUUID,
		HMACSignature:   req.HMACSig,
		Synced:          true,
	}

	if req.SelfieData != "" {
		// TODO: Upload selfie to MinIO and set URL
		selfieURL := "selfies/" + att.ID.String() + ".jpg"
		att.SelfieURL = &selfieURL
	}

	if err := u.attRepo.CreateCheckIn(att); err != nil {
		return nil, err
	}

	return att, nil
}

func (u *AttendanceUsecase) CheckOut(userID uuid.UUID, req *model.CheckOutRequest) (*model.Attendance, error) {
	// Verify HMAC signature
	if !u.verifyHMAC(req, req.HMACSig) {
		return nil, errors.New("invalid signature")
	}

	// Find today's attendance
	att, err := u.attRepo.FindTodayByUser(userID)
	if err != nil {
		return nil, errors.New("no check-in record found today")
	}
	if att.CheckOutTime != nil {
		return nil, errors.New("already checked out today")
	}

	// Verify device
	if att.DeviceUUID != req.DeviceUUID {
		return nil, errors.New("device mismatch")
	}

	now := time.Now()
	att.CheckOutTime = &now
	att.CheckOutLocation = []float64{req.Latitude, req.Longitude}

	if err := u.attRepo.CreateCheckOut(att); err != nil {
		return nil, err
	}

	return att, nil
}

func (u *AttendanceUsecase) GetHistory(userID uuid.UUID, limit, offset int) ([]model.AttendanceResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	return u.attRepo.GetHistory(userID, limit, offset)
}

func (u *AttendanceUsecase) GetTodayAttendance(userID uuid.UUID) (*model.Attendance, error) {
	return u.attRepo.FindTodayByUser(userID)
}

func (u *AttendanceUsecase) GetLocations() ([]model.Location, error) {
	return u.attRepo.GetLocations()
}

func (u *AttendanceUsecase) verifyHMAC(payload interface{}, signature string) bool {
	// Simplified HMAC verification
	// In production, serialize payload properly
	data := "attendance-payload" // TODO: serialize actual payload
	secret := []byte(u.config.JWT.Secret)

	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(data))
	expectedMAC := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(signature), []byte(expectedMAC))
}
