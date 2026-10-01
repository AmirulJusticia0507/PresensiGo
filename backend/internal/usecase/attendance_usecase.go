package usecase

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/PresensiGo/backend/internal/ai"
	"github.com/PresensiGo/backend/internal/auth"
	"github.com/PresensiGo/backend/internal/config"
	"github.com/PresensiGo/backend/internal/model"
	"github.com/PresensiGo/backend/internal/repository"
	"github.com/PresensiGo/backend/internal/storage"
)

type AttendanceUsecase struct {
	attRepo       *repository.AttendanceRepository
	userRepo      *repository.UserRepository
	offlineRepo   *repository.OfflineQueueRepository
	config        *config.Config
	minio         *storage.Client
	faceAI        *ai.Client
}

const maxSelfieSize = 1024 * 1024

func NewAttendanceUsecase(attRepo *repository.AttendanceRepository, userRepo *repository.UserRepository, offlineRepo *repository.OfflineQueueRepository, cfg *config.Config, minio *storage.Client, faceAI *ai.Client) *AttendanceUsecase {
	return &AttendanceUsecase{
		attRepo:     attRepo,
		userRepo:    userRepo,
		offlineRepo: offlineRepo,
		config:      cfg,
		minio:       minio,
		faceAI:      faceAI,
	}
}

func (u *AttendanceUsecase) CheckIn(userID uuid.UUID, req *model.CheckInRequest) (*model.Attendance, error) {
	return u.checkIn(userID, req, false)
}

func (u *AttendanceUsecase) checkIn(userID uuid.UUID, req *model.CheckInRequest, offline bool) (*model.Attendance, error) {
	if existing, err := u.attRepo.FindByIdempotencyKey(userID, req.IdempotencyKey); err == nil {
		return existing, nil
	}
	// Build payload for HMAC verification
	payload := map[string]interface{}{
		"device_uuid": req.DeviceUUID,
		"latitude":    fmt.Sprintf("%v", req.Latitude),
		"longitude":   fmt.Sprintf("%v", req.Longitude),
		"timestamp":   fmt.Sprintf("%d", req.Timestamp),
	}

	// Verify HMAC with timestamp check (using device UUID as key)
	if err := verifyAttendanceProof(payload, req.HMACSig, req.DeviceUUID, req.Timestamp, offline); err != nil {
		return nil, err
	}

	// Device binding validation
	user, err := u.userRepo.FindByID(userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if user.DeviceUUID == nil {
		return nil, errors.New("device not bound - please login first")
	}
	if *user.DeviceUUID != req.DeviceUUID {
		return nil, errors.New("device mismatch - unauthorized device")
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

	if lastCheckOut, err := u.attRepo.FindLastCheckOut(userID); err == nil && lastCheckOut.CheckOutTime != nil {
		if len(lastCheckOut.CheckOutLocation) >= 2 {
			lastLat := lastCheckOut.CheckOutLocation[0]
			lastLng := lastCheckOut.CheckOutLocation[1]
			timeDiff := time.Since(*lastCheckOut.CheckOutTime).Hours()
			if timeDiff > 0 && timeDiff < 24 {
				dist := haversine(lastLat, lastLng, req.Latitude, req.Longitude)
				speed := dist / timeDiff
				if speed > 900 {
					return nil, errors.New("velocity anomaly detected - possible mock location")
				}
			}
		}
	}

	if req.SelfieData == "" {
		return nil, errors.New("selfie is required for check-in")
	}
	if !offline {
		if err := auth.VerifyFaceChallenge(req.LivenessToken, req.LivenessChallenge, userID, u.config.JWT.Secret); err != nil {
			return nil, err
		}
	} else if req.LivenessChallenge != "turn_left" && req.LivenessChallenge != "turn_right" {
		return nil, errors.New("invalid offline liveness challenge")
	}
	if len(user.FaceEmbedding) == 0 {
		return nil, errors.New("face is not enrolled; enroll it from Settings before check-in")
	}
	var enrolledEmbedding []float32
	if err := json.Unmarshal(user.FaceEmbedding, &enrolledEmbedding); err != nil || len(enrolledEmbedding) != 512 {
		return nil, errors.New("stored face enrollment is invalid; please enroll again")
	}
	if u.faceAI == nil {
		return nil, errors.New("face recognition service is unavailable")
	}
	verification, err := u.faceAI.Verify(
		context.Background(),
		req.SelfieData,
		enrolledEmbedding,
		req.LivenessChallenge,
		user.FaceSimilarityThreshold,
	)
	if err != nil {
		return nil, err
	}
	if !verification.LivenessPassed {
		return nil, errors.New("liveness check failed; follow the head-turn instruction and retry")
	}
	if !verification.Verified {
		return nil, fmt.Errorf("face verification failed (similarity %.3f, required %.3f)", verification.Similarity, verification.Threshold)
	}

	now := time.Now()
	isLate := now.Hour() >= 9

	att := &model.Attendance{
		ID:                    uuid.New(),
		UserID:                userID,
		LocationID:            location.ID,
		CheckInTime:           &now,
		CheckInLocation:       []float64{req.Latitude, req.Longitude},
		Status:                "present",
		IsLate:                isLate,
		DeviceUUID:            req.DeviceUUID,
		HMACSignature:         req.HMACSig,
		Synced:                true,
		CheckInIdempotencyKey: &req.IdempotencyKey,
	}

	var uploadedObject string
	if req.SelfieData != "" {
		image, contentType, extension, err := decodeSelfie(req.SelfieData)
		if err != nil {
			return nil, err
		}
		if u.minio == nil {
			return nil, errors.New("selfie storage is unavailable")
		}

		uploadedObject = fmt.Sprintf(
			"selfies/%s/%d-%s.%s",
			userID,
			time.Now().UnixMilli(),
			att.ID,
			extension,
		)
		selfieURL, err := u.minio.PutImage(context.Background(), uploadedObject, contentType, image)
		if err != nil {
			return nil, err
		}
		att.SelfieURL = &selfieURL
	}

	if err := u.attRepo.CreateCheckIn(att); err != nil {
		if uploadedObject != "" {
			_ = u.minio.RemoveObject(context.Background(), uploadedObject)
		}
		return nil, err
	}

	return att, nil
}

func (u *AttendanceUsecase) GetFaceChallenge(userID uuid.UUID) (*model.FaceChallengeResponse, error) {
	challenge, token, expiresAt, err := auth.NewFaceChallenge(userID, u.config.JWT.Secret)
	if err != nil {
		return nil, errors.New("failed to create liveness challenge")
	}
	return &model.FaceChallengeResponse{Challenge: challenge, Token: token, ExpiresAt: expiresAt}, nil
}

func decodeSelfie(encoded string) ([]byte, string, string, error) {
	if comma := strings.IndexByte(encoded, ','); strings.HasPrefix(encoded, "data:") && comma >= 0 {
		encoded = encoded[comma+1:]
	}

	image, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, "", "", errors.New("selfie must be valid base64")
	}
	if len(image) == 0 || len(image) > maxSelfieSize {
		return nil, "", "", fmt.Errorf("selfie must be between 1 byte and %d bytes", maxSelfieSize)
	}

	if len(image) >= 3 && image[0] == 0xff && image[1] == 0xd8 && image[2] == 0xff {
		return image, "image/jpeg", "jpg", nil
	}
	pngHeader := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	if len(image) >= len(pngHeader) && string(image[:len(pngHeader)]) == string(pngHeader) {
		return image, "image/png", "png", nil
	}
	return nil, "", "", errors.New("selfie format must be JPEG or PNG")
}

func (u *AttendanceUsecase) CheckOut(userID uuid.UUID, req *model.CheckOutRequest) (*model.Attendance, error) {
	return u.checkOut(userID, req, false)
}

func (u *AttendanceUsecase) checkOut(userID uuid.UUID, req *model.CheckOutRequest, offline bool) (*model.Attendance, error) {
	if existing, err := u.attRepo.FindByIdempotencyKey(userID, req.IdempotencyKey); err == nil {
		return existing, nil
	}
	payload := map[string]interface{}{
		"device_uuid": req.DeviceUUID,
		"latitude":    fmt.Sprintf("%v", req.Latitude),
		"longitude":   fmt.Sprintf("%v", req.Longitude),
		"timestamp":   fmt.Sprintf("%d", req.Timestamp),
	}

	if err := verifyAttendanceProof(payload, req.HMACSig, req.DeviceUUID, req.Timestamp, offline); err != nil {
		return nil, err
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
	att.CheckOutIdempotencyKey = &req.IdempotencyKey

	if err := u.attRepo.CreateCheckOut(att); err != nil {
		return nil, err
	}

	return att, nil
}

func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371.0
	toRad := func(deg float64) float64 { return deg * 3.141592653589793 / 180.0 }
	dLat := toRad(lat2 - lat1)
	dLon := toRad(lon2 - lon1)
	sinDLat := dLat / 2
	sinDLon := dLon / 2
	a := sinDLat*sinDLat + toRad(lat1)*toRad(lat2)*sinDLon*sinDLon
	return 2 * R * 3.141592653589793 / 180.0 * 0.5 * a
}

func verifyAttendanceProof(payload map[string]interface{}, signature, deviceUUID string, timestamp int64, offline bool) error {
	if !offline {
		return auth.VerifyHMACWithTimestamp(payload, signature, deviceUUID, timestamp)
	}
	age := time.Now().Unix() - timestamp
	if age < -300 || age > int64((24*time.Hour).Seconds()) {
		return errors.New("offline action timestamp is outside the 24 hour sync window")
	}
	if !auth.VerifyHMAC(payload, signature, deviceUUID) {
		return errors.New("invalid signature")
	}
	return nil
}

func (u *AttendanceUsecase) Sync(userID uuid.UUID, req *model.SyncRequest) []model.SyncResult {
	results := make([]model.SyncResult, 0, len(req.Actions))
	for _, action := range req.Actions {
		result := model.SyncResult{IdempotencyKey: action.IdempotencyKey}
		if existing, err := u.attRepo.FindByIdempotencyKey(userID, action.IdempotencyKey); err == nil {
			result.Status = "duplicate"
			result.Attendance = existing
			results = append(results, result)
			continue
		}

		if u.offlineRepo != nil {
			var payload model.CheckInRequest
			_ = json.Unmarshal(action.Payload, &payload)
			_ = u.offlineRepo.Create(&model.OfflinePayload{
				ID:              uuid.New(),
				UserID:          userID,
				ActionType:      action.ActionType,
				Payload:         string(action.Payload),
				DeviceTimestamp: time.Unix(payload.Timestamp, 0),
				Synced:          false,
				SyncAttempts:    0,
			})
		}

		var attendance *model.Attendance
		var err error
		switch action.ActionType {
		case "check_in":
			var payload model.CheckInRequest
			if err = json.Unmarshal(action.Payload, &payload); err == nil {
				payload.IdempotencyKey = action.IdempotencyKey
				attendance, err = u.checkIn(userID, &payload, true)
			}
		case "check_out":
			var payload model.CheckOutRequest
			if err = json.Unmarshal(action.Payload, &payload); err == nil {
				payload.IdempotencyKey = action.IdempotencyKey
				attendance, err = u.checkOut(userID, &payload, true)
			}
		default:
			err = errors.New("unsupported offline action")
		}
		if err != nil {
			result.Status = "failed"
			result.Error = err.Error()
		} else {
			result.Status = "synced"
			result.Attendance = attendance
		}
		results = append(results, result)
	}
	return results
}

func (u *AttendanceUsecase) GetSyncStatus(userID uuid.UUID) (*model.SyncStatusResponse, error) {
	if u.offlineRepo == nil {
		return &model.SyncStatusResponse{
			PendingCount: 0,
			StuckCount:   0,
			LastSyncAt:   nil,
		}, nil
	}
	unsynced, err := u.offlineRepo.GetUnsynced(userID)
	if err != nil {
		return nil, err
	}
	pending := len(unsynced)
	stuck := 0
	for _, item := range unsynced {
		if item.SyncAttempts >= 5 {
			stuck++
		}
	}
	return &model.SyncStatusResponse{
		PendingCount: pending,
		StuckCount:   stuck,
	}, nil
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

func (u *AttendanceUsecase) CreateLocation(req *model.Location) error {
	return u.attRepo.CreateLocation(req)
}

func (u *AttendanceUsecase) UpdateLocation(id uuid.UUID, req *model.Location) error {
	return u.attRepo.UpdateLocation(id, req)
}

func (u *AttendanceUsecase) DeleteLocation(id uuid.UUID) error {
	return u.attRepo.DeleteLocation(id)
}
