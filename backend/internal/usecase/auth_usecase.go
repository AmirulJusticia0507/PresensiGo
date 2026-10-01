package usecase

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/PresensiGo/backend/internal/ai"
	"github.com/PresensiGo/backend/internal/auth"
	"github.com/PresensiGo/backend/internal/config"
	"github.com/PresensiGo/backend/internal/model"
	"github.com/PresensiGo/backend/internal/repository"
)

type AuthUsecase struct {
	userRepo   *repository.UserRepository
	config     *config.Config
	jwtService *auth.JWTService
	faceAI     *ai.Client
}

func NewAuthUsecase(userRepo *repository.UserRepository, cfg *config.Config, faceAI *ai.Client) *AuthUsecase {
	return &AuthUsecase{
		userRepo:   userRepo,
		config:     cfg,
		jwtService: auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.ExpireHour),
		faceAI:     faceAI,
	}
}

func (u *AuthUsecase) EnrollFace(ctx context.Context, userID uuid.UUID, selfies []string) error {
	if u.faceAI == nil {
		return errors.New("face recognition service is unavailable")
	}
	embedding, err := u.faceAI.Enroll(ctx, selfies)
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(embedding)
	if err != nil {
		return errors.New("failed to encode face embedding")
	}
	return u.userRepo.UpdateFaceEmbedding(userID, encoded)
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
		ID:                      uuid.New(),
		Name:                    req.Name,
		Email:                   req.Email,
		PasswordHash:            string(hashedPassword),
		Role:                    "employee",
		FaceSimilarityThreshold: u.config.AI.SimilarityThreshold,
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

	// Device binding check
	if user.DeviceUUID == nil {
		// First login - bind this device
		_ = u.userRepo.UpdateDeviceUUID(user.ID, req.DeviceUUID)
		user.DeviceUUID = &req.DeviceUUID
	} else if *user.DeviceUUID != req.DeviceUUID {
		// Device mismatch - reject login
		return nil, errors.New("device not authorized - please contact admin to reset device binding")
	}

	token, err := u.jwtService.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

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
