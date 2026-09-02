package usecase

import (
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/PresensiGo/backend/internal/config"
	"github.com/PresensiGo/backend/internal/model"
	"github.com/PresensiGo/backend/internal/repository"
)

type AuthUsecase struct {
	userRepo *repository.UserRepository
	config   *config.Config
}

func NewAuthUsecase(userRepo *repository.UserRepository, cfg *config.Config) *AuthUsecase {
	return &AuthUsecase{
		userRepo: userRepo,
		config:   cfg,
	}
}

func (u *AuthUsecase) Register(req *model.RegisterRequest) (*model.User, error) {
	existing, _ := u.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		ID:           uuid.New(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         "employee",
	}

	if err := u.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *AuthUsecase) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	user, err := u.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Device binding - auto-bind on first login, skip check for now
	if user.DeviceUUID == nil {
		_ = u.userRepo.UpdateDeviceUUID(user.ID, req.DeviceUUID)
		user.DeviceUUID = &req.DeviceUUID
	}

	// Generate JWT (simplified - in production use jwt-go library)
	token := generateToken(user.ID, u.config.JWT.Secret, u.config.JWT.ExpireHour)

	return &model.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (u *AuthUsecase) GetByID(id uuid.UUID) (*model.User, error) {
	return u.userRepo.FindByID(id)
}

func (u *AuthUsecase) UpdateFaceEmbedding(userID uuid.UUID, embedding []byte) error {
	return u.userRepo.UpdateFaceEmbedding(userID, embedding)
}

func generateToken(userID uuid.UUID, secret string, expireHour int) string {
	return userID.String()
}
