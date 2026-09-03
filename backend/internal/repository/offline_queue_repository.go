package repository

import (
	"database/sql"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/PresensiGo/backend/internal/model"
)

type OfflineQueueRepository struct {
	db *sql.DB
}

func NewOfflineQueueRepository(db *sql.DB) *OfflineQueueRepository {
	return &OfflineQueueRepository{db: db}
}

func (r *OfflineQueueRepository) Create(payload *model.OfflinePayload) error {
	query := `
		INSERT INTO offline_queue (id, user_id, action_type, payload, hmac_signature, device_timestamp, synced, sync_attempts)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.Exec(query,
		payload.ID, payload.UserID, payload.ActionType, payload.Payload,
		payload.HMACSignature, payload.DeviceTimestamp, payload.Synced, payload.SyncAttempts,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *OfflineQueueRepository) GetUnsynced(userID uuid.UUID) ([]model.OfflinePayload, error) {
	query := `
		SELECT id, user_id, action_type, payload, hmac_signature, device_timestamp, synced, sync_attempts, created_at
		FROM offline_queue WHERE user_id = $1 AND synced = false`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payloads []model.OfflinePayload
	for rows.Next() {
		var p model.OfflinePayload
		var payloadJSON []byte
		var hmacSig sql.NullString

		err := rows.Scan(
			&p.ID, &p.UserID, &p.ActionType, &payloadJSON,
			&hmacSig, &p.DeviceTimestamp, &p.Synced, &p.SyncAttempts, &p.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if hmacSig.Valid {
			p.HMACSignature = hmacSig.String
		}

		if payloadJSON != nil {
			json.Unmarshal(payloadJSON, &p.Payload)
		}

		payloads = append(payloads, p)
	}

	return payloads, nil
}

func (r *OfflineQueueRepository) MarkSynced(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, id := range ids {
		_, err := tx.Exec(`UPDATE offline_queue SET synced = true, sync_attempts = sync_attempts + 1 WHERE id = $1`, id)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}