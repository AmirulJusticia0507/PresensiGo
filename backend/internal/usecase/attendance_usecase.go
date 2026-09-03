package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/PresensiGo/backend/internal/auth"
	"github.com/PresensiGo/backend/internal/config"
	"github.com/PresensiGo/backend/internal/model"
	"github.com/PresensiGo/backend/internal/repository"
	"github.com/PresensiGo/backend/internal/storage"
)

type AttendanceUsecase struct {
	attRepo *repository.AttendanceRepository
	config  *config.Config
	minio   *storage.Client
}

func NewAttendanceUsecase(attRepo *repository.AttendanceRepository, cfg *config.Config, minio *storage.Client) *AttendanceUsecase {
	return &AttendanceUsecase{
		attRepo: attRepo,
		config:  cfg,
		minio:   minio,
	}
}

func (u *AttendanceUsecase) CheckIn(userID uuid.UUID, req *model.CheckInRequest) (*model.Attendance, error) {
	payload := map[string]interface{}{
		"user_id":     userID.String(),
		"latitude":    req.Latitude,
		"longitude":   req.Longitude,
		"device_uuid": req.DeviceUUID,
	}
	if !auth.VerifyHMAC(payload, req.HMACSig, u.config.JWT.Secret) {
		return nil, errors.New("invalid signature")
	}

	location, err := u.attRepo.FindNearestLocation(req.Latitude, req.Longitude)
	if err != nil {
		return nil, errors.New("no location found nearby")
	}

	inside, err := u.attRepo.CheckGeofence(location.ID, req.Latitude, req.Longitude)
	if err != nil {
		return nil, err
	}
	if !inside {
		return nil, errors.New("you are outside the geofence radius")
	}

	existing, _ := u.attRepo.FindTodayByUser(userID)
	if existing != nil && existing.CheckOutTime == nil {
		return nil, errors.New("already checked in today")
	}

	now := time.Now()
	isLate := now.Hour() >= 9

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
		// Upload selfie to MinIO (or generate URL if MinIO not available)
		objectName := "selfies/" + att.ID.String() + ".jpg"
		if u.minio != nil {
			// TODO: Actual MinIO upload when client is properly configured
			_ = u.minio
		}
		att.SelfieURL = &objectName
	}

	if err := u.attRepo.CreateCheckIn(att); err != nil {
		return nil, err
	}

	return att, nil
}

func (u *AttendanceUsecase) CheckOut(userID uuid.UUID, req *model.CheckOutRequest) (*model.Attendance, error) {
	payload := map[string]interface{}{
		"user_id":     userID.String(),
		"latitude":    req.Latitude,
		"longitude":   req.Longitude,
		"device_uuid": req.DeviceUUID,
	}
	if !auth.VerifyHMAC(payload, req.HMACSig, u.config.JWT.Secret) {
		return nil, errors.New("invalid signature")
	}

	att, err := u.attRepo.FindTodayByUser(userID)
	if err != nil {
		return nil, errors.New("no check-in record found today")
	}
	if att.CheckOutTime != nil {
		return nil, errors.New("already checked out today")
	}

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