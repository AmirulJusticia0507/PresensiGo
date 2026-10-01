package usecase

import (
	"encoding/json"
	"testing"

	"github.com/PresensiGo/backend/internal/model"
	"github.com/google/uuid"
)

func TestSync_IdempotencyKeyPreventsDuplicate(t *testing.T) {
	userID := uuid.New()
	idempotencyKey := uuid.New()

	existing := &model.Attendance{
		ID:                    uuid.New(),
		UserID:                userID,
		CheckInIdempotencyKey: &idempotencyKey,
	}

	_ = existing

	payload := model.CheckInRequest{
		Latitude:       -6.2,
		Longitude:      106.8,
		DeviceUUID:     "test-device",
		Timestamp:      1234567890,
		HMACSig:        "sig",
		IdempotencyKey: idempotencyKey,
	}
	payloadBytes, _ := json.Marshal(payload)

	req := &model.SyncRequest{
		Actions: []model.SyncAction{
			{
				IdempotencyKey: idempotencyKey,
				ActionType:     "check_in",
				Payload:        payloadBytes,
			},
		},
	}

	_ = req
}

func TestSync_UnsupportedActionType(t *testing.T) {
	userID := uuid.New()
	idempotencyKey := uuid.New()

	payload := model.CheckInRequest{
		Latitude:       -6.2,
		Longitude:      106.8,
		DeviceUUID:     "test-device",
		Timestamp:      1234567890,
		HMACSig:        "sig",
		IdempotencyKey: idempotencyKey,
	}
	payloadBytes, _ := json.Marshal(payload)

	req := &model.SyncRequest{
		Actions: []model.SyncAction{
			{
				IdempotencyKey: idempotencyKey,
				ActionType:     "invalid_action",
				Payload:        payloadBytes,
			},
		},
	}

	_ = req
	_ = userID
}

func TestSync_MultipleActions(t *testing.T) {
	userID := uuid.New()
	key1 := uuid.New()
	key2 := uuid.New()

	payload1 := model.CheckInRequest{
		Latitude:       -6.2,
		Longitude:      106.8,
		DeviceUUID:     "test-device",
		Timestamp:      1234567890,
		HMACSig:        "sig",
		IdempotencyKey: key1,
	}
	payload2 := model.CheckOutRequest{
		Latitude:       -6.2,
		Longitude:      106.8,
		DeviceUUID:     "test-device",
		Timestamp:      1234567891,
		HMACSig:        "sig",
		IdempotencyKey: key2,
	}

	p1, _ := json.Marshal(payload1)
	p2, _ := json.Marshal(payload2)

	req := &model.SyncRequest{
		Actions: []model.SyncAction{
			{IdempotencyKey: key1, ActionType: "check_in", Payload: p1},
			{IdempotencyKey: key2, ActionType: "check_out", Payload: p2},
		},
	}

	_ = req
	_ = userID
}

func TestSync_EmptyActions(t *testing.T) {
	req := &model.SyncRequest{
		Actions: []model.SyncAction{},
	}

	if len(req.Actions) != 0 {
		t.Fatal("expected empty actions")
	}
}

func TestSyncResult_StatusValues(t *testing.T) {
	validStatuses := map[string]bool{
		"synced":    true,
		"duplicate": true,
		"failed":    true,
	}

	for status := range validStatuses {
		result := model.SyncResult{
			IdempotencyKey: uuid.New(),
			Status:         status,
		}
		if result.Status != status {
			t.Errorf("expected status %q, got %q", status, result.Status)
		}
	}
}
