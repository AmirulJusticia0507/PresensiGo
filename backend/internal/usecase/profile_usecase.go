package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"
	"unicode"

	"github.com/PresensiGo/backend/internal/model"
	"github.com/PresensiGo/backend/internal/repository"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ProfileUsecaseIface defines the profile operations interface
type ProfileUsecaseIface interface {
	RegisterUser(ctx context.Context, req *RegisterUserRequest) (*model.User, string, error)
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*model.User, error)
	UpdateUserProfile(ctx context.Context, userID uuid.UUID, req *UpdateProfileRequest) (*model.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error
}

// Request models
type RegisterUserRequest struct {
	Email           string  `json:"email" validate:"required,email"`
	Password        string  `json:"password" validate:"required"`
	ConfirmPassword string  `json:"confirm_password" validate:"required"`
	Name            string  `json:"name" validate:"required,min=2,max=255"`
	Phone           *string `json:"phone"`
	TermsAccepted   bool    `json:"terms_accepted" validate:"required"`
}

type UpdateProfileRequest struct {
	Name                  string  `json:"name" validate:"required,min=2,max=255"`
	Phone                 *string `json:"phone"`
	EmergencyContactName  *string `json:"emergency_contact_name"`
	EmergencyContactPhone *string `json:"emergency_contact_phone"`
	Address               *string `json:"address"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

// ProfileUsecase implements profile business logic
type ProfileUsecase struct {
	userRepo repository.UserRepositoryIface
}

// NewProfileUsecase creates a new profile usecase instance
func NewProfileUsecase(userRepo repository.UserRepositoryIface) *ProfileUsecase {
	return &ProfileUsecase{
		userRepo: userRepo,
	}
}

// Validation Constants
const (
	minPasswordLength = 8
	maxPasswordLength = 255
	bcryptCost        = 12
)

// Validation error messages (user-friendly)
const (
	ErrInvalidEmailFormat   = "Invalid email format"
	ErrEmailAlreadyExists   = "Email already registered"
	ErrWeakPassword         = "Password must be at least 8 characters with uppercase, lowercase, number, and special character"
	ErrPasswordMismatch     = "Password confirmation does not match"
	ErrInvalidPhone         = "Invalid phone format (expected: +1234567890)"
	ErrInvalidName          = "Name must be 2-255 characters"
	ErrInvalidEmergencyName = "Emergency contact name must be 2-255 characters"
	ErrInvalidAddress       = "Address must be 5-500 characters"
	ErrInvalidCredentials   = "Invalid credentials"
	ErrUserNotFound         = "User not found"
	ErrTermsNotAccepted     = "Terms and conditions must be accepted"
)

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	// Simple email regex (RFC 5322 simplified)
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New(ErrInvalidEmailFormat)
	}
	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < minPasswordLength {
		return errors.New(ErrWeakPassword)
	}

	hasUpper := false
	hasLower := false
	hasDigit := false
	hasSpecial := false

	specialChars := "!@#$%^&*"

	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
		} else if unicode.IsLower(char) {
			hasLower = true
		} else if unicode.IsDigit(char) {
			hasDigit = true
		} else if regexp.MustCompile(fmt.Sprintf("[%s]", regexp.QuoteMeta(specialChars))).MatchString(string(char)) {
			hasSpecial = true
		}
	}

	if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
		return errors.New(ErrWeakPassword)
	}

	return nil
}

// ValidatePhone validates phone format (E.164)
func ValidatePhone(phone string) error {
	if phone == "" {
		return nil // Phone is optional
	}

	// E.164 format: +[country code][number]
	phoneRegex := regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	if !phoneRegex.MatchString(phone) {
		return errors.New(ErrInvalidPhone)
	}
	return nil
}

// ValidateName validates name length
func ValidateName(name string) error {
	if len(name) < 2 || len(name) > 255 {
		return errors.New(ErrInvalidName)
	}
	return nil
}

// ValidateAddress validates address length
func ValidateAddress(address string) error {
	if address == "" {
		return nil // Address is optional
	}
	if len(address) < 5 || len(address) > 500 {
		return errors.New(ErrInvalidAddress)
	}
	return nil
}

// ValidateEmergencyContact validates emergency contact fields
func ValidateEmergencyContact(name *string, phone *string) error {
	if name == nil && phone == nil {
		return nil // Both optional
	}

	if name != nil && (len(*name) < 2 || len(*name) > 255) {
		return errors.New(ErrInvalidEmergencyName)
	}

	if phone != nil && *phone != "" {
		if err := ValidatePhone(*phone); err != nil {
			return err
		}
	}

	return nil
}

// HashPassword hashes a password using bcrypt with cost 12
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword verifies a password against a hash (constant-time comparison)
func VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

// RegisterUser creates a new user with validated credentials
func (uc *ProfileUsecase) RegisterUser(ctx context.Context, req *RegisterUserRequest) (*model.User, string, error) {
	// Validate email format
	if err := ValidateEmail(req.Email); err != nil {
		return nil, "", err
	}

	// Check if email already exists
	existingUser, err := uc.userRepo.GetUserByEmail(ctx, req.Email)
	if err == nil && existingUser != nil {
		return nil, "", errors.New(ErrEmailAlreadyExists)
	}

	// Validate password strength
	if err := ValidatePassword(req.Password); err != nil {
		return nil, "", err
	}

	// Validate password confirmation matches
	if req.Password != req.ConfirmPassword {
		return nil, "", errors.New(ErrPasswordMismatch)
	}

	// Validate name
	if err := ValidateName(req.Name); err != nil {
		return nil, "", err
	}

	// Validate phone if provided
	if req.Phone != nil {
		if err := ValidatePhone(*req.Phone); err != nil {
			return nil, "", err
		}
	}

	// Validate terms accepted
	if !req.TermsAccepted {
		return nil, "", errors.New(ErrTermsNotAccepted)
	}

	// Hash password
	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, "", fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user model
	now := getCurrentTimestamp()
	user := &model.User{
		ID:              uuid.New(),
		Email:           req.Email,
		Name:            req.Name,
		PasswordHash:    passwordHash,
		Phone:           req.Phone,
		TermsAcceptedAt: &now,
		Role:            "employee",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Store user in repository
	if err := uc.userRepo.CreateUser(ctx, user); err != nil {
		return nil, "", fmt.Errorf("failed to create user: %w", err)
	}

	// Generate JWT token (placeholder - actual implementation in AuthUsecase)
	// For now, return empty token - this will be handled by separate auth layer
	token := ""

	return user, token, nil
}

// GetUserProfile retrieves a user's profile
func (uc *ProfileUsecase) GetUserProfile(ctx context.Context, userID uuid.UUID) (*model.User, error) {
	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New(ErrUserNotFound)
	}
	if user == nil {
		return nil, errors.New(ErrUserNotFound)
	}

	// Ensure password hash is not exposed
	user.PasswordHash = ""

	return user, nil
}

// UpdateUserProfile updates a user's profile information
func (uc *ProfileUsecase) UpdateUserProfile(ctx context.Context, userID uuid.UUID, req *UpdateProfileRequest) (*model.User, error) {
	// Validate all fields
	if err := ValidateName(req.Name); err != nil {
		return nil, err
	}

	if req.Phone != nil {
		if err := ValidatePhone(*req.Phone); err != nil {
			return nil, err
		}
	}

	if err := ValidateEmergencyContact(req.EmergencyContactName, req.EmergencyContactPhone); err != nil {
		return nil, err
	}

	// Validate address if provided
	if req.Address != nil {
		if err := ValidateAddress(*req.Address); err != nil {
			return nil, err
		}
	}

	// Fetch existing user
	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New(ErrUserNotFound)
	}
	if user == nil {
		return nil, errors.New(ErrUserNotFound)
	}

	// Update fields
	user.Name = req.Name
	user.Phone = req.Phone
	user.EmergencyContactName = req.EmergencyContactName
	user.EmergencyContactPhone = req.EmergencyContactPhone
	user.Address = req.Address

	// Update in repository
	if err := uc.userRepo.UpdateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Ensure password hash is not exposed
	user.PasswordHash = ""

	return user, nil
}

// ChangePassword changes a user's password
func (uc *ProfileUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, req *ChangePasswordRequest) error {
	// Fetch user
	user, err := uc.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New(ErrInvalidCredentials)
	}
	if user == nil {
		return errors.New(ErrInvalidCredentials)
	}

	// Verify current password
	if err := VerifyPassword(user.PasswordHash, req.CurrentPassword); err != nil {
		return errors.New(ErrInvalidCredentials)
	}

	// Validate new password strength
	if err := ValidatePassword(req.NewPassword); err != nil {
		return err
	}

	// Validate new password confirmation matches
	if req.NewPassword != req.ConfirmPassword {
		return errors.New(ErrPasswordMismatch)
	}

	// Hash new password
	newPasswordHash, err := HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password in repository
	if err := uc.userRepo.UpdatePassword(ctx, userID, newPasswordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// Helper function to get current timestamp
func getCurrentTimestamp() time.Time {
	return time.Now().UTC()
}
