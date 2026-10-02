package repository

import (
	"context"
	"database/sql"

	"github.com/PresensiGo/backend/internal/model"
	"github.com/google/uuid"
)

// UserRepositoryIface defines the interface for user repository operations
type UserRepositoryIface interface {
	Create(user *model.User) error
	FindByEmail(email string) (*model.User, error)
	FindByID(id uuid.UUID) (*model.User, error)
	FindByDeviceUUID(deviceUUID string) (*model.User, error)
	UpdateDeviceUUID(userID uuid.UUID, deviceUUID string) error
	UpdateFaceEmbedding(userID uuid.UUID, embedding []byte) error
	// New methods for profile management
	CreateUser(ctx context.Context, user *model.User) error
	GetUserByEmail(ctx context.Context, email string) (*model.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error)
	UpdateUser(ctx context.Context, user *model.User) error
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
}

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	query := `
		INSERT INTO users (id, name, email, password_hash, role, device_uuid, face_embedding, face_similarity_threshold)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at`

	return r.db.QueryRow(query,
		user.ID, user.Name, user.Email, user.PasswordHash,
		user.Role, user.DeviceUUID, user.FaceEmbedding, user.FaceSimilarityThreshold,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
}

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, name, email, password_hash, role, device_uuid, face_embedding, face_similarity_threshold, face_enrolled_at, created_at, updated_at
		FROM users WHERE email = $1`

	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.Role, &user.DeviceUUID, &user.FaceEmbedding, &user.FaceSimilarityThreshold, &user.FaceEnrolledAt,
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
		SELECT id, name, email, password_hash, role, device_uuid, face_embedding, face_similarity_threshold, face_enrolled_at, created_at, updated_at
		FROM users WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.Role, &user.DeviceUUID, &user.FaceEmbedding, &user.FaceSimilarityThreshold, &user.FaceEnrolledAt,
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
		SELECT id, name, email, password_hash, role, device_uuid, face_embedding, face_similarity_threshold, face_enrolled_at, created_at, updated_at
		FROM users WHERE device_uuid = $1`

	err := r.db.QueryRow(query, deviceUUID).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.Role, &user.DeviceUUID, &user.FaceEmbedding, &user.FaceSimilarityThreshold, &user.FaceEnrolledAt,
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
	query := `UPDATE users SET face_embedding = $1, face_enrolled_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, embedding, userID)
	return err
}

// CreateUser creates a new user with all profile fields
func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (id, name, email, password_hash, role, device_uuid, phone, emergency_contact_name, 
		emergency_contact_phone, address, profile_picture_url, terms_accepted_at, face_embedding, 
		face_similarity_threshold, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		RETURNING created_at, updated_at`

	return r.db.QueryRowContext(ctx, query,
		user.ID, user.Name, user.Email, user.PasswordHash,
		user.Role, user.DeviceUUID, user.Phone, user.EmergencyContactName,
		user.EmergencyContactPhone, user.Address, user.ProfilePictureUrl, user.TermsAcceptedAt,
		user.FaceEmbedding, user.FaceSimilarityThreshold, user.CreatedAt, user.UpdatedAt,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
}

// GetUserByEmail retrieves a user by email (context-aware)
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, name, email, password_hash, role, device_uuid, phone, emergency_contact_name, 
		emergency_contact_phone, address, profile_picture_url, terms_accepted_at, face_embedding, 
		face_similarity_threshold, face_enrolled_at, created_at, updated_at
		FROM users WHERE email = $1`

	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.Role, &user.DeviceUUID, &user.Phone, &user.EmergencyContactName,
		&user.EmergencyContactPhone, &user.Address, &user.ProfilePictureUrl, &user.TermsAcceptedAt,
		&user.FaceEmbedding, &user.FaceSimilarityThreshold, &user.FaceEnrolledAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// GetUserByID retrieves a user by ID (context-aware)
func (r *UserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	user := &model.User{}
	query := `
		SELECT id, name, email, password_hash, role, device_uuid, phone, emergency_contact_name, 
		emergency_contact_phone, address, profile_picture_url, terms_accepted_at, face_embedding, 
		face_similarity_threshold, face_enrolled_at, created_at, updated_at
		FROM users WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash,
		&user.Role, &user.DeviceUUID, &user.Phone, &user.EmergencyContactName,
		&user.EmergencyContactPhone, &user.Address, &user.ProfilePictureUrl, &user.TermsAcceptedAt,
		&user.FaceEmbedding, &user.FaceSimilarityThreshold, &user.FaceEnrolledAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// UpdateUser updates user profile information
func (r *UserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users 
		SET name = $1, phone = $2, emergency_contact_name = $3, emergency_contact_phone = $4, 
		address = $5, profile_picture_url = $6, updated_at = NOW()
		WHERE id = $7
		RETURNING updated_at`

	return r.db.QueryRowContext(ctx, query,
		user.Name, user.Phone, user.EmergencyContactName, user.EmergencyContactPhone,
		user.Address, user.ProfilePictureUrl, user.ID,
	).Scan(&user.UpdatedAt)
}

// UpdatePassword updates user's password hash
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, passwordHash, userID)
	return err
}
