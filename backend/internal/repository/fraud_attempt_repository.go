package repository

import (
	"database/sql"

	"github.com/PresensiGo/backend/internal/model"
	"github.com/google/uuid"
)

type FraudAttemptRepository struct{ db *sql.DB }

func NewFraudAttemptRepository(db *sql.DB) *FraudAttemptRepository {
	return &FraudAttemptRepository{db: db}
}

func (r *FraudAttemptRepository) Create(userID uuid.UUID, req *model.FraudAttemptRequest) (*model.FraudAttempt, error) {
	attempt := &model.FraudAttempt{ID: uuid.New(), UserID: userID, Reason: req.Reason, Platform: req.Platform, Latitude: req.Latitude, Longitude: req.Longitude, Accuracy: req.Accuracy}
	err := r.db.QueryRow(`
		INSERT INTO fraud_attempts (id, user_id, reason, platform, latitude, longitude, accuracy)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING attempted_at`, attempt.ID, attempt.UserID, attempt.Reason, attempt.Platform,
		attempt.Latitude, attempt.Longitude, attempt.Accuracy).Scan(&attempt.AttemptedAt)
	return attempt, err
}

func (r *FraudAttemptRepository) CountRecent(userID uuid.UUID) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM fraud_attempts WHERE user_id = $1 AND reason = 'mock_location' AND attempted_at >= NOW() - INTERVAL '24 hours'`, userID).Scan(&count)
	return count, err
}

func (r *FraudAttemptRepository) ListAlerts(limit, threshold int) ([]model.FraudAlert, error) {
	rows, err := r.db.Query(`
		SELECT f.user_id, u.name, u.email, COUNT(*) AS attempt_count,
		       (ARRAY_AGG(f.reason ORDER BY f.attempted_at DESC))[1] AS last_reason,
		       MAX(f.attempted_at) AS last_attempt
		FROM fraud_attempts f JOIN users u ON u.id = f.user_id
		WHERE f.reason = 'mock_location' AND f.attempted_at >= NOW() - INTERVAL '24 hours'
		GROUP BY f.user_id, u.name, u.email
		HAVING COUNT(*) >= $1
		ORDER BY last_attempt DESC LIMIT $2`, threshold, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	alerts := make([]model.FraudAlert, 0)
	for rows.Next() {
		var alert model.FraudAlert
		if err := rows.Scan(&alert.UserID, &alert.UserName, &alert.UserEmail, &alert.AttemptCount, &alert.LastReason, &alert.LastAttempt); err != nil {
			return nil, err
		}
		alerts = append(alerts, alert)
	}
	return alerts, rows.Err()
}
