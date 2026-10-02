# Implementation Plan: Mobile Profile & Register Flow

## Overview

Implement a complete user registration and profile management system for PresensiGo. The implementation spans backend profile service endpoints (registration, profile CRUD, password change) and mobile UI screens (registration form, profile display/edit, password change dialog). All components integrate with existing P1 validation, error handling, and authentication infrastructure. 8 tasks across 4 implementation phases: backend models/service, API endpoints, mobile UI, and testing.

## Implementation Plan

### Phase 1: Backend Foundation (Task 1-2)
- Enhance User model with profile fields (phone, emergency contact, address)
- Create ProfileService usecase with core business logic
- Implement validation helpers (email, password, phone)

### Phase 2: Backend Endpoints (Task 3-4)
- Implement POST `/api/auth/register` with duplicate email handling
- Implement GET/PUT `/api/profile` for profile CRUD
- Implement PUT `/api/profile/password` with current password verification

### Phase 3: Mobile UI (Task 5-6)
- Create RegisterScreen with validation feedback and error handling
- Create ProfileScreen with edit mode and password change dialog
- Implement ProfileService (API client) for mobile

### Phase 4: Testing & Verification (Task 7-8)
- Write backend unit tests (validation, hashing, repo mocks)
- Write integration tests (API endpoints, error scenarios)
- Write mobile UI tests (form validation, API client)
- Manual testing: complete user journey registration → profile → edit → password change
- Build verification and final commit

### Estimated Timeline
- Phase 1: 1 day
- Phase 2: 1 day
- Phase 3: 2 days
- Phase 4: 1 day
- **Total: 5 days**

## Task Dependency Graph

```json
{
  "waves": [
    { "wave": 1, "tasks": ["1"] },
    { "wave": 2, "tasks": ["2"] },
    { "wave": 3, "tasks": ["3", "4"] },
    { "wave": 4, "tasks": ["5", "6"] },
    { "wave": 5, "tasks": ["7", "8"] }
  ]
}
```

**Wave Explanation:**
- **Wave 1:** Task 1 (User model) independent, foundation for all others
- **Wave 2:** Task 2 (ProfileService) depends on Task 1
- **Wave 3:** Tasks 3, 4 (Endpoints) depend on Task 2, can run parallel
- **Wave 4:** Tasks 5, 6 (Mobile UI) depend on Tasks 3, 4, can run parallel
- **Wave 5:** Tasks 7, 8 (Testing) depend on all prior tasks, can run parallel

## Tasks

- [x] 1. Enhance User Model and Create ProfileService Usecase
  - **Subtasks:**
    - [ ] 1.1 Add fields to User model: phone, emergencyContact (name+phone), address, profilePictureUrl, termsAcceptedAt
    - [ ] 1.2 Create validation helper functions: ValidateEmail, ValidatePassword, ValidatePhone, ValidateName, HashPassword, VerifyPassword
    - [ ] 1.3 Create ProfileUsecaseIface interface with methods: RegisterUser, GetUserProfile, UpdateUserProfile, ChangePassword
    - [ ] 1.4 Implement ProfileUsecase struct with userRepo dependency injection
    - [ ] 1.5 Implement RegisterUser method: validate fields, check email uniqueness, hash password, create user, return JWT
    - [ ] 1.6 Implement GetUserProfile method: fetch user by ID, return sanitized object
    - [ ] 1.7 Implement UpdateUserProfile method: validate fields, update atomically, return updated user
    - [ ] 1.8 Implement ChangePassword method: verify current password, hash new password, update in DB
  - **Files Created/Modified:**
    - `backend/internal/model/user.go` (modified - add fields)
    - `backend/internal/usecase/profile_usecase.go` (new)
  - **Acceptance Criteria:**
    - User struct includes all profile fields (phone, emergency contact, address)
    - All validation functions work correctly (tested in Task 7)
    - Password hashing uses bcrypt with cost 12
    - ProfileUsecase methods implement correct business logic
    - No passwords stored/returned in responses
    - Compile without errors

- [x] 2. Enhance User Repository with Profile Methods
  - **Subtasks:**
    - [ ] 2.1 Add GetUserByEmail method to UserRepository
    - [ ] 2.2 Add CreateUser method with all fields (email, name, phone, password_hash, etc.)
    - [ ] 2.3 Add UpdateUser method for profile updates
    - [ ] 2.4 Add UpdatePassword method for password changes
    - [ ] 2.5 Ensure email index exists in database for uniqueness check
    - [ ] 2.6 Add database migration for new user fields (phone, emergency_contact_*, address, profile_picture_url)
    - [ ] 2.7 Test repo methods with test database
  - **Files Created/Modified:**
    - `backend/internal/repository/user_repository.go` (enhanced)
    - `backend/migrations/004_profile_fields.sql` (new migration)
  - **Acceptance Criteria:**
    - GetUserByEmail returns user or error correctly
    - CreateUser inserts all fields atomically
    - UpdateUser updates fields atomically
    - UpdatePassword updates hash safely
    - Email index prevents duplicates
    - Migration runs without errors
    - All repo methods tested

- [x] 3. Implement Backend Register and Profile Endpoints
  - **Subtasks:**
    - [ ] 3.1 Create Register handler: POST `/api/auth/register`, accept email/password/confirmPassword/name/phone/termsAccepted
    - [ ] 3.2 Validate request in handler: use validation middleware
    - [ ] 3.3 Call ProfileUsecase.RegisterUser
    - [ ] 3.4 Return 201 Created with JWT token and user object
    - [ ] 3.5 Handle errors: 400 validation, 409 duplicate email, 500 server errors
    - [ ] 3.6 Enhance existing GetProfile handler: return complete user object with all fields
    - [ ] 3.7 Create UpdateProfile handler: PUT `/api/profile`, require auth middleware
    - [ ] 3.8 Validate request fields, call ProfileUsecase.UpdateUserProfile
    - [ ] 3.9 Return 200 OK with updated user object or 400/401 on error
    - [ ] 3.10 Wire handlers to routes in main.go (apply rate limiting for register)
  - **Files Created/Modified:**
    - `backend/internal/delivery/http/handler.go` (add Register, UpdateProfile methods)
    - `backend/cmd/api/main.go` (wire routes with rate limiting)
  - **Acceptance Criteria:**
    - POST `/api/auth/register` creates user and returns JWT
    - Duplicate email returns 409 Conflict
    - Invalid email/password returns 400 with field details
    - PUT `/api/profile` requires auth and updates user
    - GET `/api/profile` returns complete profile (already existed, enhanced)
    - Rate limiting applied to register (3/minute per IP)
    - All errors sanitized (no schema leaks)
    - Endpoints compile and start without errors

- [x] 4. Implement Password Change Endpoint
  - **Subtasks:**
    - [ ] 4.1 Create ChangePassword handler: PUT `/api/profile/password`, require auth
    - [ ] 4.2 Accept currentPassword, newPassword, confirmPassword
    - [ ] 4.3 Validate new password meets strength requirements (8+, upper, lower, digit, special)
    - [ ] 4.4 Validate confirmation matches new password
    - [ ] 4.5 Call ProfileUsecase.ChangePassword with current password verification
    - [ ] 4.6 Return 200 OK with success message or 400/401 on error
    - [ ] 4.7 Wire handler to route with auth middleware
    - [ ] 4.8 Return generic error messages (don't reveal if current password wrong vs new password weak)
  - **Files Created/Modified:**
    - `backend/internal/delivery/http/handler.go` (add ChangePassword method)
    - `backend/cmd/api/main.go` (wire route)
  - **Acceptance Criteria:**
    - PUT `/api/profile/password` verifies current password before updating
    - New password must meet strength requirements
    - Incorrect current password returns 401 with generic message
    - Weak new password returns 400 with generic message
    - Successful change returns 200 OK
    - Error messages don't leak information about why validation failed
    - Endpoint requires auth middleware
    - Compile without errors

- [x] 5. Implement Mobile RegisterScreen
  - **Subtasks:**
    - [ ] 5.1 Create RegisterScreen widget: `presensigo_mobile/lib/features/auth/screens/register_screen.dart`
    - [ ] 5.2 Design form with TextFormFields: email, password, confirmPassword, name, phone
    - [ ] 5.3 Add CheckboxListTile for terms acceptance
    - [ ] 5.4 Implement real-time email validation with visual feedback
    - [ ] 5.5 Implement password strength indicator (show requirements: 8+, upper, lower, digit, special)
    - [ ] 5.6 Implement confirm password validation (must match)
    - [ ] 5.7 Implement phone validation (optional, E.164 format)
    - [ ] 5.8 Implement name validation (2-255 characters)
    - [ ] 5.9 Disable Register button until all validations pass
    - [ ] 5.10 Call ProfileService.register on Register button tap
    - [ ] 5.11 Show loading indicator during registration
    - [ ] 5.12 Handle errors: show field-level errors from API, show 409 duplicate email error
    - [ ] 5.13 Store JWT token in SharedPreferences on success
    - [ ] 5.14 Navigate to ProfileScreen after successful registration
  - **Files Created/Modified:**
    - `presensigo_mobile/lib/features/auth/screens/register_screen.dart` (new)
    - `presensigo_mobile/lib/features/auth/models/register_request.dart` (new, if needed)
  - **Acceptance Criteria:**
    - Form displays all fields with proper labels
    - Real-time validation shows errors below fields
    - Password strength indicator shows all 4 requirements
    - Register button disabled until form valid
    - Loading state shows during API call
    - Successful registration: JWT stored, navigate to ProfileScreen
    - API errors displayed correctly (field-level for 400, general for 409)
    - No passwords stored in SharedPreferences (only JWT)
    - UI matches design system

- [x] 6. Implement Mobile ProfileScreen and ProfileService
  - **Subtasks:**
    - [ ] 6.1 Create ProfileScreen widget: `presensigo_mobile/lib/features/profile/screens/profile_screen.dart`
    - [ ] 6.2 Create ProfileService API client: `presensigo_mobile/lib/data/services/profile_service.dart`
    - [ ] 6.3 Create FormValidator helper: `presensigo_mobile/lib/core/utils/form_validator.dart`
    - [ ] 6.4 Implement ProfileService.getProfile() method: GET `/api/profile`, cache in SharedPreferences
    - [ ] 6.5 Implement ProfileService.updateProfile() method: PUT `/api/profile` with edited fields
    - [ ] 6.6 Implement ProfileService.changePassword() method: PUT `/api/profile/password`
    - [ ] 6.7 Design ProfileScreen with read-only display: name, email, phone, emergency contact, address
    - [ ] 6.8 Load profile on screen open, show loading state
    - [ ] 6.9 Add Edit button (pencil icon) to enter edit mode
    - [ ] 6.10 Edit mode: populate form from cached data, enable input fields
    - [ ] 6.11 Edit mode: real-time validation for each field
    - [ ] 6.12 Edit mode: Save button (submit PUT request), Cancel button (discard changes)
    - [ ] 6.13 Add "Change Password" button → show dialog with current, new, confirm password fields
    - [ ] 6.14 Password change dialog: validate strength, match confirmation, call updatePassword
    - [ ] 6.15 Handle all errors: show field-level errors, show generic messages for auth errors
    - [ ] 6.16 Implement pull-to-refresh to reload profile
    - [ ] 6.17 FormValidator: validateEmail, validatePassword, validatePhone, validateName, validateEmergencyContact, validateAddress
  - **Files Created/Modified:**
    - `presensigo_mobile/lib/features/profile/screens/profile_screen.dart` (new)
    - `presensigo_mobile/lib/data/services/profile_service.dart` (new)
    - `presensigo_mobile/lib/core/utils/form_validator.dart` (new)
  - **Acceptance Criteria:**
    - ProfileScreen loads profile on open (GET request)
    - Profile displays all fields read-only initially
    - Edit button switches to edit mode
    - Edit form validates all fields in real-time
    - Save submits PUT request and updates display
    - Cancel discards changes without saving
    - Change password dialog validates strength and match
    - All API errors handled: field-level for validation (400), generic for auth (401), duplicate (409)
    - Profile data cached locally
    - Pull-to-refresh reloads profile
    - UI matches design system
    - No crashes on error responses

- [ ] 7. Write Backend Tests (Unit + Integration)
  - **Subtasks:**
    - [ ] 7.1 Write unit tests for validation functions: ValidateEmail (valid/invalid formats), ValidatePassword (strength check), ValidatePhone (format check), ValidateName (length check)
    - [ ] 7.2 Write unit tests for password hashing: HashPassword creates hash, VerifyPassword checks correctly, wrong password fails
    - [ ] 7.3 Write unit tests for ProfileUsecase methods with mock repository
    - [ ] 7.4 Write integration tests for POST `/api/auth/register`: success with all fields, 400 validation errors, 409 duplicate email
    - [ ] 7.5 Write integration tests for GET `/api/profile`: success returns user, 401 without token, 404 user not found
    - [ ] 7.6 Write integration tests for PUT `/api/profile`: success updates fields, 400 validation, 401 without token
    - [ ] 7.7 Write integration tests for PUT `/api/profile/password`: success, wrong current password (401), weak new password (400)
    - [ ] 7.8 Write integration tests for rate limiting: register endpoint returns 429 after 3 requests in 1 minute
    - [ ] 7.9 Run all tests: `go test ./... -v` from backend directory
    - [ ] 7.10 Run linter: `go vet ./...` and `go fmt` check
  - **Files Created/Modified:**
    - `backend/internal/usecase/profile_usecase_test.go` (new)
    - `backend/internal/repository/_test/user_repository_test.go` (enhanced)
    - `backend/internal/delivery/http/handler_test.go` (enhanced)
  - **Acceptance Criteria:**
    - All validation unit tests pass
    - All password hashing tests pass (hash correct, verify works, wrong password rejected)
    - All ProfileUsecase tests pass (business logic correct)
    - All integration tests pass (API endpoints work correctly, errors handled)
    - Rate limiting test passes (429 after limit)
    - `go test ./...` shows 100% tests passing
    - `go vet ./...` shows no issues
    - `go fmt` check passes
    - All tests use proper error messages (not expose internals)

- [ ] 8. Write Mobile Tests and Final Verification
  - **Subtasks:**
    - [ ] 8.1 Write unit tests for FormValidator: ValidateEmail, ValidatePassword, ValidatePhone, ValidateName (valid/invalid cases)
    - [ ] 8.2 Write widget tests for RegisterScreen: form displays, validation feedback shows, button disabled until valid, loading state
    - [ ] 8.3 Write widget tests for ProfileScreen: profile loads, edit mode works, save/cancel buttons, password change dialog
    - [ ] 8.4 Write unit tests for ProfileService: mock HTTP, test getProfile, updateProfile, changePassword success and errors
    - [ ] 8.5 Manual testing: register new user (valid and invalid inputs), log in, view profile, edit fields, change password
    - [ ] 8.6 Manual testing: verify API errors show correctly (field-level for 400, generic for 401/409)
    - [ ] 8.7 Manual testing: verify no crashes on network errors or 500 server errors
    - [ ] 8.8 Build verification: `flutter build apk --debug` or `flutter build ios` (or run on emulator)
    - [ ] 8.9 Run linter: `flutter analyze` (no errors, minimal warnings)
    - [ ] 8.10 Commit all changes: stage files, write commit message, push to branch
  - **Files Created/Modified:**
    - `presensigo_mobile/lib/features/auth/screens/register_screen_test.dart` (new)
    - `presensigo_mobile/lib/features/profile/screens/profile_screen_test.dart` (new)
    - `presensigo_mobile/lib/core/utils/form_validator_test.dart` (new)
    - `presensigo_mobile/lib/data/services/profile_service_test.dart` (new)
  - **Acceptance Criteria:**
    - All FormValidator unit tests pass
    - RegisterScreen widget tests pass (form, validation, button state, loading)
    - ProfileScreen widget tests pass (load, edit, save, password change)
    - ProfileService tests pass (API calls, error handling)
    - Manual testing passes: user journey from registration to profile edit
    - No runtime crashes
    - `flutter analyze` passes (no errors)
    - `flutter build` succeeds (debug or emulator)
    - All files committed with clear commit message
    - Branch ready for PR/review

## Implementation Notes

### Backend

#### Password Hashing
Use bcrypt with cost 12 (balance between security and performance):
```go
import "golang.org/x/crypto/bcrypt"

hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
// Store hash, never store password

// Later: verify
err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(inputPassword))
if err != nil {
  // Wrong password
}
```

#### Email Validation
Use validator package for RFC 5322 compliance:
```go
import "github.com/go-playground/validator/v10"

validate := validator.New()
err := validate.Var(email, "required,email")
```

#### Phone Validation
Use libphonenumber-go for E.164 format:
```go
import phonenumber "github.com/ttacon/libphonenumber-go/phonenumber"

number, _ := phonenumber.Parse(phone, "US")
if !phonenumber.IsValidNumber(number) {
  return errors.New("invalid phone format")
}
```

#### Database Queries
Use parameterized queries to prevent SQL injection:
```go
// Good: parameterized
rows, err := db.Query("SELECT * FROM users WHERE email = $1", email)

// Bad: string concatenation (never do this)
rows, err := db.Query(fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email))
```

### Mobile

#### Form Validation
Create reusable FormValidator helper to avoid duplicating validation logic:
```dart
class FormValidator {
  static ValidationResult validateEmail(String email) => ...
  static ValidationResult validatePassword(String password) => ...
  // etc.
}

// In widget:
final emailResult = FormValidator.validateEmail(_emailController.text);
setState(() => _emailError = emailResult.error);
```

#### API Error Handling
Parse JSON error responses to extract field-level details:
```dart
try {
  await api.register(...);
} catch (e) {
  if (e is ApiException) {
    if (e.statusCode == 400) {
      // Parse details, show field errors
      for (var detail in e.details) {
        _errors[detail['field']] = detail['reason'];
      }
    } else if (e.statusCode == 409) {
      // Email already exists
      _errors['email'] = 'Email already registered';
    }
  }
}
```

#### Secure Token Storage
Store JWT token in SharedPreferences (or SecureStorage in production):
```dart
final prefs = await SharedPreferences.getInstance();
await prefs.setString('jwt_token', token);

// Later: retrieve
String? token = prefs.getString('jwt_token');
```

### Testing

#### Backend Integration Tests
Use test database or in-memory database:
```go
// In test setup
db, _ := sqlx.Open("postgres", "user=test dbname=presensi_test ...")
defer db.Close()

// Run migrations
// Execute test cases
// Verify results
```

#### Mobile Widget Tests
Use Flutter's testing utilities:
```dart
testWidgets('RegisterScreen validates email', (WidgetTester tester) async {
  await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));
  
  await tester.enterText(find.byType(TextField).first, 'invalid-email');
  await tester.pumpWidget(const SizedBox()); // Trigger validation
  
  expect(find.text('Invalid email format'), findsOneWidget);
});
```

### Database Migrations

Create migration for profile fields:
```sql
-- File: backend/migrations/004_profile_fields.sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);
ALTER TABLE users ADD COLUMN IF NOT EXISTS emergency_contact_name VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS emergency_contact_phone VARCHAR(20);
ALTER TABLE users ADD COLUMN IF NOT EXISTS address TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS profile_picture_url VARCHAR(500);
ALTER TABLE users ADD COLUMN IF NOT EXISTS terms_accepted_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
```

### Rate Limiting Configuration

Register endpoint: 3 requests per minute per IP
```go
// In main.go during route setup
authGroup := r.NewRoute().Subrouter()
authGroup.Use(rateLimiter.RateLimitMiddleware("register"))
authGroup.HandleFunc("/api/auth/register", handler.Register).Methods("POST")
```

### Error Response Format

Always return consistent JSON error format:
```json
{
  "error": "Validation failed",
  "details": [
    {"field": "email", "reason": "Invalid email format"}
  ],
  "request_id": "abc123def456..."
}
```

### Security Checklist

- [x] Passwords hashed with bcrypt (cost 12)
- [x] Email validated for format and uniqueness
- [x] Password strength enforced (8+, upper, lower, digit, special)
- [x] Phone validated for E.164 format
- [x] Error messages sanitized (no schema leaks)
- [x] Rate limiting on public endpoints
- [x] JWT token validation on protected endpoints
- [x] Constant-time password comparison (bcrypt handles this)
- [x] No passwords in logs or error messages
- [x] All inputs validated before database write
- [x] Atomic transactions for critical operations

## Notes

### Reusing P1 Infrastructure
- Use existing validation middleware for request validation
- Use existing error handler middleware for consistent error formatting
- Use existing auth middleware for JWT token validation
- Use existing UserRepository pattern
- Use existing error wrapping and logging patterns

### Future Enhancements (P3)
- Email verification via link
- Profile picture upload to MinIO
- Password reset via email
- Token refresh mechanism
- Session invalidation on password change
- Account deletion
- Audit logging for security events
- Email notifications
