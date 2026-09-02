package repository

import (
	"database/sql"

	"github.com/google/uuid"
	"github.com/PresensiGo/backend/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	query := `
		INSERT INTO users (id, name, email, password_hash, role, device_uuid, face_embedding)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING created_at, updated_at`

	return r.db.QueryRow(query,
		user.ID, user.Name, user.Email, user.PasswordHash,
		user.Role, user.DeviceUUID, user.FaceEmbedding,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, name, email, password_hash, role, device_uuid, face_embedding, created_at, updated_at
		FROM users WHERE email = $1`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.Role, &user.DeviceUUID, &user.FaceEmbedding,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, name, email, password_hash, role, device_uuid, face_embedding, created_at, updated_at
		FROM users WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.Role, &user.DeviceUUID, &user.FaceEmbedding,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) FindByDeviceUUID(deviceUUID string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, name, email, password_hash, role, device_uuid, face_embedding, created_at, updated_at
		FROM users WHERE device_uuid = $1`

	err := r.db.QueryRow(query, deviceUUID).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.Role, &user.DeviceUUID, &user.FaceEmbedding,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *UserRepository) UpdateDeviceUUID(userID uuid.UUID, deviceUUID string) error {
	query := `UPDATE users SET device_uuid = $1 WHERE id = $2`
	_, err := r.db.Exec(query, deviceUUID, userID)
	return err
}

func (r *UserRepository) UpdateFaceEmbedding(userID uuid.UUID, embedding []byte) error {
	query := `UPDATE users SET face_embedding = $1 WHERE id = $2`
	_, err := r.db.Exec(query, embedding, userID)
	return err
}
