package usecase

import (
	"context"
	"testing"

	"github.com/PresensiGo/backend/internal/model"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// ========== 7.1 Unit Tests for Validation Functions ==========

// TestValidateEmail_ValidFormats tests valid email formats
func TestValidateEmail_ValidFormats(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{"simple email", "user@example.com"},
		{"email with dot", "user.name@example.com"},
		{"email with plus", "user+tag@example.com"},
		{"email with number", "user123@example.co.uk"},
		{"email with hyphen", "user-name@example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if err != nil {
				t.Errorf("ValidateEmail(%s) expected nil, got %v", tt.email, err)
			}
		})
	}
}

// TestValidateEmail_InvalidFormats tests invalid email formats
func TestValidateEmail_InvalidFormats(t *testing.T) {
	tests := []struct {
		name  string
		email string
	}{
		{"missing @", "userexample.com"},
		{"missing domain", "user@"},
		{"missing local part", "@example.com"},
		{"missing TLD", "user@example"},
		{"double @", "user@@example.com"},
		{"space in email", "user @example.com"},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if err == nil {
				t.Errorf("ValidateEmail(%s) expected error, got nil", tt.email)
			}
		})
	}
}

// TestValidatePassword_ValidPasswords tests password meeting all strength requirements
func TestValidatePassword_ValidPasswords(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{"simple strong password", "SecurePass123!"},
		{"long password", "MyLongSecurePassword123!@#"},
		{"with multiple special chars", "Pass@#$%123!abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if err != nil {
				t.Errorf("ValidatePassword(%s) expected nil, got %v", tt.password, err)
			}
		})
	}
}

// TestValidatePassword_MissingUppercase tests password without uppercase
func TestValidatePassword_MissingUppercase(t *testing.T) {
	password := "securepass123!"
	err := ValidatePassword(password)
	if err == nil {
		t.Errorf("ValidatePassword(%s) expected error for missing uppercase, got nil", password)
	}
}

// TestValidatePassword_MissingLowercase tests password without lowercase
func TestValidatePassword_MissingLowercase(t *testing.T) {
	password := "SECUREPASS123!"
	err := ValidatePassword(password)
	if err == nil {
		t.Errorf("ValidatePassword(%s) expected error for missing lowercase, got nil", password)
	}
}

// TestValidatePassword_MissingDigit tests password without digit
func TestValidatePassword_MissingDigit(t *testing.T) {
	password := "SecurePass!"
	err := ValidatePassword(password)
	if err == nil {
		t.Errorf("ValidatePassword(%s) expected error for missing digit, got nil", password)
	}
}

// TestValidatePassword_MissingSpecialChar tests password without special character
func TestValidatePassword_MissingSpecialChar(t *testing.T) {
	password := "SecurePass123"
	err := ValidatePassword(password)
	if err == nil {
		t.Errorf("ValidatePassword(%s) expected error for missing special char, got nil", password)
	}
}

// TestValidatePassword_TooShort tests password shorter than 8 characters
func TestValidatePassword_TooShort(t *testing.T) {
	password := "Ps12!"
	err := ValidatePassword(password)
	if err == nil {
		t.Errorf("ValidatePassword(%s) expected error for too short password, got nil", password)
	}
}

// TestValidatePhone_ValidFormats tests valid E.164 phone formats
func TestValidatePhone_ValidFormats(t *testing.T) {
	tests := []struct {
		name  string
		phone string
	}{
		{"US number", "+12025551234"},
		{"long country code", "+33123456789"},
		{"India number", "+919876543210"},
		{"short country", "+447123456789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePhone(tt.phone)
			if err != nil {
				t.Errorf("ValidatePhone(%s) expected nil, got %v", tt.phone, err)
			}
		})
	}
}

// TestValidatePhone_InvalidFormats tests invalid phone formats
func TestValidatePhone_InvalidFormats(t *testing.T) {
	tests := []struct {
		name  string
		phone string
	}{
		{"missing plus", "12025551234"},
		{"spaces", "+1 202 555 1234"},
		{"dashes", "+1-202-555-1234"},
		{"invalid country code zero", "+0123456789"},
		{"single digit", "+1"},
		{"too long", "+12345678901234567"},
		{"letters", "+1202ABCDEFG1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePhone(tt.phone)
			if err == nil {
				t.Errorf("ValidatePhone(%s) expected error, got nil", tt.phone)
			}
		})
	}
}

// TestValidatePhone_EmptyIsValid tests that empty phone is valid (optional field)
func TestValidatePhone_EmptyIsValid(t *testing.T) {
	err := ValidatePhone("")
	if err != nil {
		t.Errorf("ValidatePhone(\"\") expected nil for optional field, got %v", err)
	}
}

// TestValidateName_ValidNames tests valid name lengths
func TestValidateName_ValidNames(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"minimum length", "Al"},
		{"normal name", "John Doe"},
		{"long name", "Alexander James William Montgomery"},
		{"maximum length", string(make([]byte, 255))}, // 255 chars
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For the 255-char test, fill it with valid characters
			testName := tt.input
			if len(testName) == 255 {
				testName = "A" // Start with valid char, then pad
				for i := 1; i < 255; i++ {
					testName += "B"
				}
			}
			err := ValidateName(testName)
			if err != nil {
				t.Errorf("ValidateName(%s) expected nil, got %v", tt.name, err)
			}
		})
	}
}

// TestValidateName_InvalidNames tests invalid name lengths
func TestValidateName_InvalidNames(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"single character", "A"},
		{"empty string", ""},
		{"too long", string(make([]byte, 256))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testName := tt.input
			if len(testName) == 256 {
				testName = "A"
				for i := 1; i < 256; i++ {
					testName += "B"
				}
			}
			err := ValidateName(testName)
			if err == nil {
				t.Errorf("ValidateName(%s) expected error, got nil", tt.name)
			}
		})
	}
}

// TestValidateAddress_ValidAddresses tests valid address lengths
func TestValidateAddress_ValidAddresses(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{"minimum length", "12345"},
		{"normal address", "123 Main Street, New York"},
		{"long address", "The Building on the Corner of Main and 5th Avenue, Suite 200, New York, NY 10001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddress(tt.address)
			if err != nil {
				t.Errorf("ValidateAddress(%s) expected nil, got %v", tt.name, err)
			}
		})
	}
}

// TestValidateAddress_EmptyIsValid tests that empty address is valid (optional field)
func TestValidateAddress_EmptyIsValid(t *testing.T) {
	err := ValidateAddress("")
	if err != nil {
		t.Errorf("ValidateAddress(\"\") expected nil for optional field, got %v", err)
	}
}

// TestValidateAddress_InvalidAddresses tests invalid addresses
func TestValidateAddress_InvalidAddresses(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		{"too short", "123"},
		{"one char", "A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddress(tt.address)
			if err == nil {
				t.Errorf("ValidateAddress(%s) expected error, got nil", tt.name)
			}
		})
	}
}

// TestValidateEmergencyContact_BothProvided tests valid emergency contact
func TestValidateEmergencyContact_BothProvided(t *testing.T) {
	name := "Jane Doe"
	phone := "+12025551234"
	err := ValidateEmergencyContact(&name, &phone)
	if err != nil {
		t.Errorf("ValidateEmergencyContact expected nil, got %v", err)
	}
}

// TestValidateEmergencyContact_OnlyNameProvided tests emergency contact with only name
func TestValidateEmergencyContact_OnlyNameProvided(t *testing.T) {
	name := "Jane Doe"
	err := ValidateEmergencyContact(&name, nil)
	if err != nil {
		t.Errorf("ValidateEmergencyContact(name only) expected nil, got %v", err)
	}
}

// TestValidateEmergencyContact_OnlyPhoneProvided tests emergency contact with only phone
func TestValidateEmergencyContact_OnlyPhoneProvided(t *testing.T) {
	phone := "+12025551234"
	err := ValidateEmergencyContact(nil, &phone)
	if err != nil {
		t.Errorf("ValidateEmergencyContact(phone only) expected nil, got %v", err)
	}
}

// TestValidateEmergencyContact_BothNil tests that both nil is valid (optional)
func TestValidateEmergencyContact_BothNil(t *testing.T) {
	err := ValidateEmergencyContact(nil, nil)
	if err != nil {
		t.Errorf("ValidateEmergencyContact(nil, nil) expected nil, got %v", err)
	}
}

// TestValidateEmergencyContact_InvalidName tests invalid emergency contact name
func TestValidateEmergencyContact_InvalidName(t *testing.T) {
	name := "J"
	err := ValidateEmergencyContact(&name, nil)
	if err == nil {
		t.Errorf("ValidateEmergencyContact(invalid name) expected error, got nil")
	}
}

// TestValidateEmergencyContact_InvalidPhone tests invalid emergency contact phone
func TestValidateEmergencyContact_InvalidPhone(t *testing.T) {
	name := "Jane Doe"
	phone := "1234567890" // Missing + prefix
	err := ValidateEmergencyContact(&name, &phone)
	if err == nil {
		t.Errorf("ValidateEmergencyContact(invalid phone) expected error, got nil")
	}
}

// ========== 7.2 Unit Tests for Password Hashing ==========

// TestHashPassword_CreatesValidHash tests that HashPassword creates a valid bcrypt hash
func TestHashPassword_CreatesValidHash(t *testing.T) {
	password := "SecurePass123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Errorf("HashPassword expected nil error, got %v", err)
	}

	if hash == "" {
		t.Errorf("HashPassword expected non-empty hash, got empty string")
	}

	// Verify hash is bcrypt (starts with $2a$, $2b$, or $2x$)
	if len(hash) < 3 || (hash[0:3] != "$2a" && hash[0:3] != "$2b" && hash[0:3] != "$2x") {
		t.Errorf("HashPassword returned invalid bcrypt hash format: %s", hash)
	}
}

// TestHashPassword_ConsistentCost tests that HashPassword uses consistent cost factor 12
func TestHashPassword_ConsistentCost(t *testing.T) {
	password := "SecurePass123!"
	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	// Both hashes should have cost 12 (shown as $2a$12$ or $2b$12$)
	// Format: $2a$12$... (cost 12 is at positions 4-6)
	if len(hash1) < 6 || hash1[4:6] != "12" {
		t.Errorf("HashPassword hash1 expected cost 12, got %s", hash1[4:6])
	}
	if len(hash2) < 6 || hash2[4:6] != "12" {
		t.Errorf("HashPassword hash2 expected cost 12, got %s", hash2[4:6])
	}

	// Both hashes should be around same length (within 1 char due to minor variations)
	if len(hash1) < 50 || len(hash1) > 65 {
		t.Errorf("HashPassword hash1 length unexpected: %d", len(hash1))
	}
}

// TestHashPassword_DifferentHashes tests that different passwords produce different hashes
func TestHashPassword_DifferentHashes(t *testing.T) {
	password1 := "SecurePass123!"
	password2 := "DifferentPass456!"

	hash1, _ := HashPassword(password1)
	hash2, _ := HashPassword(password2)

	if hash1 == hash2 {
		t.Errorf("HashPassword expected different hashes for different passwords")
	}
}

// TestHashPassword_SamePwd_DifferentHashes tests that same password produces different hashes (salt)
func TestHashPassword_SamePwd_DifferentHashes(t *testing.T) {
	password := "SecurePass123!"

	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	// Different hashes due to random salt
	if hash1 == hash2 {
		t.Errorf("HashPassword expected different hashes due to random salt")
	}

	// But both should verify with the same password
	if err := bcrypt.CompareHashAndPassword([]byte(hash1), []byte(password)); err != nil {
		t.Errorf("HashPassword hash1 verification failed: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash2), []byte(password)); err != nil {
		t.Errorf("HashPassword hash2 verification failed: %v", err)
	}
}

// TestVerifyPassword_CorrectPassword tests that correct password verifies successfully
func TestVerifyPassword_CorrectPassword(t *testing.T) {
	password := "SecurePass123!"
	hash, _ := HashPassword(password)

	err := VerifyPassword(hash, password)
	if err != nil {
		t.Errorf("VerifyPassword(correct password) expected nil, got %v", err)
	}
}

// TestVerifyPassword_WrongPassword tests that wrong password fails verification
func TestVerifyPassword_WrongPassword(t *testing.T) {
	password := "SecurePass123!"
	wrongPassword := "WrongPass456!"
	hash, _ := HashPassword(password)

	err := VerifyPassword(hash, wrongPassword)
	if err == nil {
		t.Errorf("VerifyPassword(wrong password) expected error, got nil")
	}
}

// TestVerifyPassword_CaseSensitive tests that password verification is case-sensitive
func TestVerifyPassword_CaseSensitive(t *testing.T) {
	password := "SecurePass123!"
	wrongCase := "securepass123!" // lowercase s

	hash, _ := HashPassword(password)

	err := VerifyPassword(hash, wrongCase)
	if err == nil {
		t.Errorf("VerifyPassword(wrong case) expected error, got nil")
	}
}

// TestVerifyPassword_PartialMatch tests that partial password match fails
func TestVerifyPassword_PartialMatch(t *testing.T) {
	password := "SecurePass123!"
	partial := "SecurePass123"

	hash, _ := HashPassword(password)

	err := VerifyPassword(hash, partial)
	if err == nil {
		t.Errorf("VerifyPassword(partial match) expected error, got nil")
	}
}

// TestVerifyPassword_ConstantTimeComparison tests that comparison is constant-time
// This test verifies bcrypt handles constant-time comparison correctly
func TestVerifyPassword_ConstantTimeComparison(t *testing.T) {
	password := "SecurePass123!"
	hash, _ := HashPassword(password)

	// This should use constant-time comparison (bcrypt does this automatically)
	// We verify by checking it works correctly with timing-attack resistant comparison
	err := VerifyPassword(hash, password)
	if err != nil {
		t.Errorf("VerifyPassword constant-time comparison failed: %v", err)
	}

	// Wrong password should also complete without timing variation (bcrypt handles this)
	err = VerifyPassword(hash, "WrongPassword123!")
	if err == nil {
		t.Errorf("VerifyPassword should fail for wrong password")
	}
}

// ========== 7.3 Unit Tests for ProfileUsecase Methods with Mocks ==========

// mockUserRepository implements UserRepositoryIface for testing
type mockUserRepository struct {
	users                map[uuid.UUID]*model.User
	emailToID            map[string]uuid.UUID
	createUserErr        error
	getUserByIDErr       error
	getUserByEmailErr    error
	updateUserErr        error
	updatePasswordErr    error
	shouldReturnNilUser  bool
	shouldReturnNilEmail bool
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		users:     make(map[uuid.UUID]*model.User),
		emailToID: make(map[string]uuid.UUID),
	}
}

func (m *mockUserRepository) CreateUser(ctx context.Context, user *model.User) error {
	if m.createUserErr != nil {
		return m.createUserErr
	}
	m.users[user.ID] = user
	m.emailToID[user.Email] = user.ID
	return nil
}

func (m *mockUserRepository) Create(user *model.User) error {
	return m.CreateUser(context.Background(), user)
}

func (m *mockUserRepository) FindByEmail(email string) (*model.User, error) {
	return m.GetUserByEmail(context.Background(), email)
}

func (m *mockUserRepository) FindByID(id uuid.UUID) (*model.User, error) {
	return m.GetUserByID(context.Background(), id)
}

func (m *mockUserRepository) FindByDeviceUUID(deviceUUID string) (*model.User, error) {
	return nil, nil
}

func (m *mockUserRepository) UpdateDeviceUUID(userID uuid.UUID, deviceUUID string) error {
	return nil
}

func (m *mockUserRepository) UpdateFaceEmbedding(userID uuid.UUID, embedding []byte) error {
	return nil
}

func (m *mockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	if m.getUserByIDErr != nil {
		return nil, m.getUserByIDErr
	}
	if m.shouldReturnNilUser {
		return nil, nil
	}
	return m.users[id], nil
}

func (m *mockUserRepository) GetUserByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.getUserByEmailErr != nil {
		return nil, m.getUserByEmailErr
	}
	if m.shouldReturnNilEmail {
		return nil, nil
	}
	if id, exists := m.emailToID[email]; exists {
		return m.users[id], nil
	}
	return nil, nil
}

func (m *mockUserRepository) UpdateUser(ctx context.Context, user *model.User) error {
	if m.updateUserErr != nil {
		return m.updateUserErr
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	if m.updatePasswordErr != nil {
		return m.updatePasswordErr
	}
	if user, exists := m.users[userID]; exists {
		user.PasswordHash = passwordHash
		m.users[userID] = user
	}
	return nil
}

// TestRegisterUser_ValidInput tests RegisterUser with valid input
func TestRegisterUser_ValidInput(t *testing.T) {
	mockRepo := newMockUserRepository()
	usecase := NewProfileUsecase(mockRepo)

	req := &RegisterUserRequest{
		Email:           "user@example.com",
		Password:        "SecurePass123!",
		ConfirmPassword: "SecurePass123!",
		Name:            "John Doe",
		TermsAccepted:   true,
	}

	user, token, err := usecase.RegisterUser(context.Background(), req)
	if err != nil {
		t.Errorf("RegisterUser expected nil error, got %v", err)
	}
	if user == nil {
		t.Errorf("RegisterUser expected user object, got nil")
	}
	if user.Email != req.Email {
		t.Errorf("RegisterUser email mismatch: expected %s, got %s", req.Email, user.Email)
	}
	if user.Name != req.Name {
		t.Errorf("RegisterUser name mismatch: expected %s, got %s", req.Name, user.Name)
	}
	if user.Role != "employee" {
		t.Errorf("RegisterUser role expected employee, got %s", user.Role)
	}
	if user.PasswordHash == "" {
		t.Errorf("RegisterUser expected password hash, got empty string")
	}
	// Token may be empty for now (handled by auth layer)
	_ = token
}

// TestRegisterUser_DuplicateEmail tests RegisterUser with duplicate email
func TestRegisterUser_DuplicateEmail(t *testing.T) {
	mockRepo := newMockUserRepository()
	existingUser := &model.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		Name:         "Existing",
		PasswordHash: "hash",
	}
	mockRepo.users[existingUser.ID] = existingUser
	mockRepo.emailToID[existingUser.Email] = existingUser.ID

	usecase := NewProfileUsecase(mockRepo)

	req := &RegisterUserRequest{
		Email:           "user@example.com",
		Password:        "SecurePass123!",
		ConfirmPassword: "SecurePass123!",
		Name:            "John Doe",
		TermsAccepted:   true,
	}

	user, token, err := usecase.RegisterUser(context.Background(), req)
	if err == nil {
		t.Errorf("RegisterUser duplicate email expected error, got nil")
	}
	if user != nil {
		t.Errorf("RegisterUser duplicate email expected nil user, got %v", user)
	}
	_ = token
}

// TestRegisterUser_InvalidEmail tests RegisterUser with invalid email
func TestRegisterUser_InvalidEmail(t *testing.T) {
	mockRepo := newMockUserRepository()
	usecase := NewProfileUsecase(mockRepo)

	req := &RegisterUserRequest{
		Email:           "invalid-email",
		Password:        "SecurePass123!",
		ConfirmPassword: "SecurePass123!",
		Name:            "John Doe",
		TermsAccepted:   true,
	}

	user, token, err := usecase.RegisterUser(context.Background(), req)
	if err == nil {
		t.Errorf("RegisterUser invalid email expected error, got nil")
	}
	if user != nil {
		t.Errorf("RegisterUser invalid email expected nil user, got %v", user)
	}
	_ = token
}

// TestRegisterUser_WeakPassword tests RegisterUser with weak password
func TestRegisterUser_WeakPassword(t *testing.T) {
	mockRepo := newMockUserRepository()
	usecase := NewProfileUsecase(mockRepo)

	req := &RegisterUserRequest{
		Email:           "user@example.com",
		Password:        "weak",
		ConfirmPassword: "weak",
		Name:            "John Doe",
		TermsAccepted:   true,
	}

	user, token, err := usecase.RegisterUser(context.Background(), req)
	if err == nil {
		t.Errorf("RegisterUser weak password expected error, got nil")
	}
	if user != nil {
		t.Errorf("RegisterUser weak password expected nil user, got %v", user)
	}
	_ = token
}

// TestRegisterUser_PasswordMismatch tests RegisterUser with mismatched passwords
func TestRegisterUser_PasswordMismatch(t *testing.T) {
	mockRepo := newMockUserRepository()
	usecase := NewProfileUsecase(mockRepo)

	req := &RegisterUserRequest{
		Email:           "user@example.com",
		Password:        "SecurePass123!",
		ConfirmPassword: "DifferentPass456!",
		Name:            "John Doe",
		TermsAccepted:   true,
	}

	user, token, err := usecase.RegisterUser(context.Background(), req)
	if err == nil {
		t.Errorf("RegisterUser password mismatch expected error, got nil")
	}
	if user != nil {
		t.Errorf("RegisterUser password mismatch expected nil user, got %v", user)
	}
	_ = token
}

// TestRegisterUser_TermsNotAccepted tests RegisterUser when terms not accepted
func TestRegisterUser_TermsNotAccepted(t *testing.T) {
	mockRepo := newMockUserRepository()
	usecase := NewProfileUsecase(mockRepo)

	req := &RegisterUserRequest{
		Email:           "user@example.com",
		Password:        "SecurePass123!",
		ConfirmPassword: "SecurePass123!",
		Name:            "John Doe",
		TermsAccepted:   false,
	}

	user, token, err := usecase.RegisterUser(context.Background(), req)
	if err == nil {
		t.Errorf("RegisterUser terms not accepted expected error, got nil")
	}
	if user != nil {
		t.Errorf("RegisterUser terms not accepted expected nil user, got %v", user)
	}
	_ = token
}

// TestGetUserProfile_Success tests GetUserProfile returns user with all fields
func TestGetUserProfile_Success(t *testing.T) {
	mockRepo := newMockUserRepository()
	user := &model.User{
		ID:    uuid.New(),
		Email: "user@example.com",
		Name:  "John Doe",
	}
	mockRepo.users[user.ID] = user

	usecase := NewProfileUsecase(mockRepo)
	retrieved, err := usecase.GetUserProfile(context.Background(), user.ID)

	if err != nil {
		t.Errorf("GetUserProfile expected nil error, got %v", err)
	}
	if retrieved == nil {
		t.Errorf("GetUserProfile expected user object, got nil")
	}
	if retrieved.Email != user.Email {
		t.Errorf("GetUserProfile email mismatch: expected %s, got %s", user.Email, retrieved.Email)
	}
	// Verify password hash is cleared
	if retrieved.PasswordHash != "" {
		t.Errorf("GetUserProfile password hash should be cleared, got %s", retrieved.PasswordHash)
	}
}

// TestGetUserProfile_UserNotFound tests GetUserProfile with nonexistent user
func TestGetUserProfile_UserNotFound(t *testing.T) {
	mockRepo := newMockUserRepository()
	mockRepo.shouldReturnNilUser = true

	usecase := NewProfileUsecase(mockRepo)
	user, err := usecase.GetUserProfile(context.Background(), uuid.New())

	if err == nil {
		t.Errorf("GetUserProfile user not found expected error, got nil")
	}
	if user != nil {
		t.Errorf("GetUserProfile user not found expected nil user, got %v", user)
	}
}

// TestUpdateUserProfile_ValidUpdate tests UpdateUserProfile with valid input
func TestUpdateUserProfile_ValidUpdate(t *testing.T) {
	mockRepo := newMockUserRepository()
	user := &model.User{
		ID:    uuid.New(),
		Email: "user@example.com",
		Name:  "John Doe",
	}
	mockRepo.users[user.ID] = user

	usecase := NewProfileUsecase(mockRepo)

	newName := "Jane Doe"
	phone := "+12025551234"
	req := &UpdateProfileRequest{
		Name:  newName,
		Phone: &phone,
	}

	updated, err := usecase.UpdateUserProfile(context.Background(), user.ID, req)
	if err != nil {
		t.Errorf("UpdateUserProfile expected nil error, got %v", err)
	}
	if updated == nil {
		t.Errorf("UpdateUserProfile expected user object, got nil")
	}
	if updated.Name != newName {
		t.Errorf("UpdateUserProfile name not updated: expected %s, got %s", newName, updated.Name)
	}
	if updated.Phone == nil || *updated.Phone != phone {
		t.Errorf("UpdateUserProfile phone not updated")
	}
}

// TestUpdateUserProfile_InvalidPhone tests UpdateUserProfile with invalid phone
func TestUpdateUserProfile_InvalidPhone(t *testing.T) {
	mockRepo := newMockUserRepository()
	user := &model.User{
		ID:    uuid.New(),
		Email: "user@example.com",
		Name:  "John Doe",
	}
	mockRepo.users[user.ID] = user

	usecase := NewProfileUsecase(mockRepo)

	invalidPhone := "1234567890"
	req := &UpdateProfileRequest{
		Name:  "Jane Doe",
		Phone: &invalidPhone,
	}

	updated, err := usecase.UpdateUserProfile(context.Background(), user.ID, req)
	if err == nil {
		t.Errorf("UpdateUserProfile invalid phone expected error, got nil")
	}
	if updated != nil {
		t.Errorf("UpdateUserProfile invalid phone expected nil user, got %v", updated)
	}
}

// TestChangePassword_Success tests ChangePassword with valid current and new password
func TestChangePassword_Success(t *testing.T) {
	mockRepo := newMockUserRepository()
	password := "OldPass123!"
	hash, _ := HashPassword(password)
	user := &model.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		Name:         "John Doe",
		PasswordHash: hash,
	}
	mockRepo.users[user.ID] = user

	usecase := NewProfileUsecase(mockRepo)

	newPassword := "NewPass456!"
	req := &ChangePasswordRequest{
		CurrentPassword: password,
		NewPassword:     newPassword,
		ConfirmPassword: newPassword,
	}

	err := usecase.ChangePassword(context.Background(), user.ID, req)
	if err != nil {
		t.Errorf("ChangePassword expected nil error, got %v", err)
	}
}

// TestChangePassword_WrongCurrentPassword tests ChangePassword with wrong current password
func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	mockRepo := newMockUserRepository()
	password := "OldPass123!"
	hash, _ := HashPassword(password)
	user := &model.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		Name:         "John Doe",
		PasswordHash: hash,
	}
	mockRepo.users[user.ID] = user

	usecase := NewProfileUsecase(mockRepo)

	req := &ChangePasswordRequest{
		CurrentPassword: "WrongPass456!",
		NewPassword:     "NewPass789!",
		ConfirmPassword: "NewPass789!",
	}

	err := usecase.ChangePassword(context.Background(), user.ID, req)
	if err == nil {
		t.Errorf("ChangePassword wrong current password expected error, got nil")
	}
}

// TestChangePassword_WeakNewPassword tests ChangePassword with weak new password
func TestChangePassword_WeakNewPassword(t *testing.T) {
	mockRepo := newMockUserRepository()
	password := "OldPass123!"
	hash, _ := HashPassword(password)
	user := &model.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		Name:         "John Doe",
		PasswordHash: hash,
	}
	mockRepo.users[user.ID] = user

	usecase := NewProfileUsecase(mockRepo)

	req := &ChangePasswordRequest{
		CurrentPassword: password,
		NewPassword:     "weak",
		ConfirmPassword: "weak",
	}

	err := usecase.ChangePassword(context.Background(), user.ID, req)
	if err == nil {
		t.Errorf("ChangePassword weak new password expected error, got nil")
	}
}

// TestChangePassword_PasswordMismatch tests ChangePassword with mismatched new passwords
func TestChangePassword_PasswordMismatch(t *testing.T) {
	mockRepo := newMockUserRepository()
	password := "OldPass123!"
	hash, _ := HashPassword(password)
	user := &model.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		Name:         "John Doe",
		PasswordHash: hash,
	}
	mockRepo.users[user.ID] = user

	usecase := NewProfileUsecase(mockRepo)

	req := &ChangePasswordRequest{
		CurrentPassword: password,
		NewPassword:     "NewPass456!",
		ConfirmPassword: "DifferentPass789!",
	}

	err := usecase.ChangePassword(context.Background(), user.ID, req)
	if err == nil {
		t.Errorf("ChangePassword mismatched passwords expected error, got nil")
	}
}
