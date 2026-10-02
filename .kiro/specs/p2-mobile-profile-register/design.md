# Design Document: Mobile Profile & Register Flow

## Overview

Design and implement a complete user registration and profile management system for PresensiGo. The system includes a registration flow with comprehensive validation, a profile management screen for viewing and editing user information, backend endpoints with security best practices, and integration with existing authentication infrastructure. The architecture leverages existing validation patterns, error handling middleware, and database infrastructure from P1 milestones.

## Architecture

### System Components

```
┌─────────────────┐         ┌──────────────────┐         ┌──────────────────┐
│  Flutter Mobile │         │  Backend (Go)    │         │  Database (SQL)  │
├─────────────────┤         ├──────────────────┤         ├──────────────────┤
│ RegisterScreen  │ POST    │ /api/auth/register           │ users table      │
│ ProfileScreen   │ GET/PUT │ /api/profile               │ profile fields   │
│ FormValidation  │ PUT     │ /api/profile/password      │ password_hash    │
│ ProfileService  │         │                            │ timestamps       │
└─────────────────┘         └──────────────────┘         └──────────────────┘
        ║                            ║                           ║
        ╚════════════════════════════╩═══════════════════════════╝
                        Secure API Calls (JWT)
```

### Data Flow: Registration

```
User fills form
  ↓
RegisterScreen validates locally
  ├─ Email format check
  ├─ Password strength check
  ├─ Password confirmation match
  ├─ Phone format (if provided)
  ├─ Name length (2-255)
  ├─ Terms checkbox
  └─ Show validation errors in real-time
  ↓
User taps Register button
  ↓
POST /api/auth/register with JSON
  ├─ Backend validation middleware
  ├─ Validate all fields (format, length, constraints)
  ├─ Hash password (bcrypt, cost: 12)
  ├─ Check email uniqueness
  ├─ Create user in database
  ├─ Generate JWT token
  └─ Return 201 Created with token + user object
  ↓
Store JWT in SecureStorage/SharedPreferences
  ↓
Auto-login, navigate to ProfileScreen
  ↓
Display user profile with success message
```

### Data Flow: Profile Edit

```
User on ProfileScreen
  ├─ GET /api/profile (fetch current data)
  ├─ Display user info in read-only mode
  ├─ Cache locally in SharedPreferences
  └─ Show loading state during fetch
  ↓
User taps Edit button
  ├─ Switch to edit mode
  ├─ Populate form fields from cached data
  ├─ Enable input fields
  ├─ Show Save/Cancel buttons
  └─ Enable real-time validation
  ↓
User edits fields
  ├─ Show validation feedback for each field
  ├─ Disable Save until valid
  └─ Show error messages below fields
  ↓
User taps Save button
  ├─ Validate all fields locally
  ├─ Show loading indicator
  └─ PUT /api/profile with updated fields
  ↓
Backend validates and updates
  ├─ Validation middleware
  ├─ Update user record atomically
  ├─ Return 200 OK with updated object
  └─ Return error with field details if validation fails
  ↓
ProfileScreen updates UI
  ├─ Display success message (toast)
  ├─ Update cached data
  ├─ Switch to read-only mode
  └─ Refresh profile display
```

### Data Flow: Password Change

```
User taps Change Password button
  ↓
ShowDialog(ChangePasswordForm)
  ├─ Current password field (masked)
  ├─ New password field (masked, show strength)
  ├─ Confirm password field (masked)
  ├─ Real-time validation
  └─ Change Password button (disabled until valid)
  ↓
User enters credentials and taps button
  ├─ Validate locally
  ├─ New password matches confirmation
  ├─ New password meets strength requirements
  └─ Show loading indicator
  ↓
PUT /api/profile/password with JSON
  ├─ {currentPassword, newPassword, confirmPassword}
  ├─ Backend middleware: validate format, strength
  ├─ Verify current password matches stored hash (constant-time comparison)
  ├─ Hash new password (bcrypt)
  ├─ Update password_hash in database
  ├─ Return 200 OK with success message
  └─ Return 400/401 with generic error if validation fails
  ↓
Mobile displays success message
  ├─ Toast: "Password changed successfully"
  ├─ Dismiss dialog
  └─ Return to profile screen
```

## Components and Interfaces

### Mobile Components (Flutter)

#### 1. RegisterScreen Widget
- **File:** `presensigo_mobile/lib/features/auth/screens/register_screen.dart`
- **Purpose:** User registration form with validation
- **Widgets Used:**
  - TextFormField: email, password, confirmPassword, name, phone
  - CheckboxListTile: terms acceptance
  - ElevatedButton: register action
  - ProgressIndicator: loading state
  - ValidationFeedback: real-time validation messages
- **State Management:** Provider (if available) or StatefulWidget
- **Methods:**
  - `_validateEmail()`: Format validation, show feedback
  - `_validatePassword()`: Strength check, show requirements
  - `_validateConfirmPassword()`: Match check
  - `_validatePhone()`: Format validation (optional)
  - `_validateName()`: Length check
  - `_onRegister()`: Call API service, handle response
  - `_showValidationError()`: Display error messages
  - `_showError()`: Display error dialog/toast

#### 2. ProfileScreen Widget
- **File:** `presensigo_mobile/lib/features/profile/screens/profile_screen.dart`
- **Purpose:** Display and edit user profile
- **Widgets Used:**
  - ProfileHeader: User name, picture, email
  - ListTile: Profile fields (phone, address, etc.)
  - FloatingActionButton: Edit button
  - Dialog: Password change form
  - Form/FormFields: Edit mode form
  - Loading states and error handling
- **State Management:** StatefulWidget or Provider
- **Methods:**
  - `_loadProfile()`: GET `/api/profile`, cache locally
  - `_onEdit()`: Enter edit mode, populate form
  - `_onSave()`: Validate, PUT `/api/profile`
  - `_onCancel()`: Discard changes, exit edit mode
  - `_onChangePassword()`: Show password change dialog
  - `_submitPasswordChange()`: PUT `/api/profile/password`
  - `_uploadProfilePicture()`: Multipart upload (optional)

#### 3. ProfileService (API Client)
- **File:** `presensigo_mobile/lib/data/services/profile_service.dart`
- **Purpose:** Handle all profile-related API calls
- **Methods:**
  ```dart
  register(email, password, confirmPassword, name, phone, termsAccepted)
    → Future<LoginResponse> (includes JWT token)
  
  getProfile()
    → Future<User>
  
  updateProfile(name, phone, emergencyContact, address)
    → Future<User>
  
  changePassword(currentPassword, newPassword, confirmPassword)
    → Future<ChangePasswordResponse>
  
  uploadProfilePicture(imageFile)
    → Future<String> (new picture URL)
  ```
- **Error Handling:**
  - Catch HTTP exceptions (400, 401, 409, 500)
  - Parse error JSON to extract field-level details
  - Throw custom exceptions with user-friendly messages
  - Handle network timeouts gracefully

#### 4. FormValidation Helper
- **File:** `presensigo_mobile/lib/core/utils/form_validator.dart`
- **Purpose:** Centralized validation logic (DRY)
- **Methods:**
  ```dart
  validateEmail(String email) → {valid: bool, error?: String}
  validatePassword(String password) → {valid: bool, strength: 0-4, errors: []}
  validateConfirmPassword(String pwd, String confirm) → {valid: bool, error?: String}
  validatePhone(String phone) → {valid: bool, error?: String}
  validateName(String name) → {valid: bool, error?: String}
  validateEmergencyContact(String name, String phone) → {valid: bool, error?: String}
  validateAddress(String address) → {valid: bool, error?: String}
  ```

### Backend Components (Go)

#### 1. ProfileService (Usecase Layer)
- **File:** `backend/internal/usecase/profile_usecase.go`
- **Purpose:** Business logic for profile operations
- **Interfaces:**
  ```go
  type ProfileUsecaseIface interface {
    RegisterUser(ctx context.Context, req RegisterUserRequest) (*User, string, error)
    GetUserProfile(ctx context.Context, userID uuid.UUID) (*User, error)
    UpdateUserProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*User, error)
    ChangePassword(ctx context.Context, userID uuid.UUID, req ChangePasswordRequest) error
  }
  
  type ProfileUsecase struct {
    userRepo      repository.UserRepositoryIface
    // logger, config, etc.
  }
  ```
- **Methods:**
  - `RegisterUser()`: Create user, hash password, return JWT
  - `GetUserProfile()`: Fetch user by ID
  - `UpdateUserProfile()`: Validate and update profile fields
  - `ChangePassword()`: Verify old password, hash new password
  - `ValidateEmail()`: Check format and uniqueness
  - `ValidatePassword()`: Check strength requirements
  - `HashPassword()`: bcrypt hashing (cost 12)
  - `VerifyPassword()`: Compare password with hash (constant-time)

#### 2. API Handler Methods (HTTP Layer)
- **File:** `backend/internal/delivery/http/handler.go`
- **Purpose:** HTTP request/response handling
- **Methods (to add):**
  - `Register(w, r)`: POST `/api/auth/register`
  - `GetProfile(w, r)`: GET `/api/profile` (already exists, enhance)
  - `UpdateProfile(w, r)`: PUT `/api/profile` (new)
  - `ChangePassword(w, r)`: PUT `/api/profile/password` (new)
- **Request/Response Types:**
  ```go
  type RegisterRequest struct {
    Email            string `json:"email"`
    Password         string `json:"password"`
    ConfirmPassword  string `json:"confirm_password"`
    Name             string `json:"name"`
    Phone            *string `json:"phone"`
    TermsAccepted    bool `json:"terms_accepted"`
  }
  
  type UpdateProfileRequest struct {
    Name             string `json:"name"`
    Phone            *string `json:"phone"`
    EmergencyContact *struct{Name, Phone string} `json:"emergency_contact"`
    Address          *string `json:"address"`
  }
  
  type ChangePasswordRequest struct {
    CurrentPassword  string `json:"current_password"`
    NewPassword      string `json:"new_password"`
    ConfirmPassword  string `json:"confirm_password"`
  }
  ```

#### 3. User Model Enhancements
- **File:** `backend/internal/model/user.go`
- **Purpose:** User data model with profile fields
- **Additions:**
  ```go
  type User struct {
    ID                  uuid.UUID
    Email               string
    Name                string
    PasswordHash        string (private, never exposed)
    Phone               *string
    EmergencyContact    *struct{Name, Phone string}
    Address             *string
    ProfilePictureUrl   *string
    FaceEmbedding       []byte (private)
    FaceSimilarityThreshold float64
    FaceEnrolledAt      *time.Time
    CreatedAt           time.Time
    UpdatedAt           time.Time
  }
  ```

#### 4. User Repository Enhancements
- **File:** `backend/internal/repository/user_repository.go`
- **Purpose:** Database access for user operations
- **Methods (to add/enhance):**
  - `CreateUser(ctx, user) error`: Insert user with all fields
  - `GetUserByEmail(ctx, email) (*User, error)`: Check email uniqueness
  - `UpdateUser(ctx, user) error`: Update profile fields
  - `UpdatePassword(ctx, userID, passwordHash) error`: Update password hash
  - Ensure proper atomic operations and error handling

#### 5. Validation Middleware Enhancements
- **File:** `backend/internal/delivery/http/middleware/validation.go`
- **Purpose:** Input validation for profile endpoints
- **Validation Rules:**
  - Email: RFC 5322 format using `validator` package
  - Password: Strength check (8+ chars, upper, lower, digit, special)
  - Phone: E.164 format or local format validation
  - Name: 2-255 characters, no injection attempts
  - Addresses: 5-500 characters, sanitized
  - Terms: Boolean flag check

#### 6. Error Handler Middleware (Existing)
- **File:** `backend/internal/delivery/http/middleware/error_handler.go`
- **Purpose:** Centralized error response formatting
- **Usage in Profile Endpoints:**
  - Return specific error messages (field + reason) for validation
  - Return generic messages for auth/credentials ("Invalid credentials")
  - Include request ID for debugging
  - Never expose schema details or raw database errors

## Data Models

### Database Schema Additions

```sql
-- Enhanced users table
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);
ALTER TABLE users ADD COLUMN IF NOT EXISTS emergency_contact_name VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS emergency_contact_phone VARCHAR(20);
ALTER TABLE users ADD COLUMN IF NOT EXISTS address TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS profile_picture_url VARCHAR(500);
ALTER TABLE users ADD COLUMN IF NOT EXISTS terms_accepted_at TIMESTAMP;

-- Optional: Audit log for security events
CREATE TABLE IF NOT EXISTS user_audit_log (
  id SERIAL PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  event_type VARCHAR(50), -- 'password_changed', 'profile_updated', 'login'
  details JSONB,
  created_at TIMESTAMP DEFAULT NOW()
);
```

### API Request/Response Examples

#### Register Request
```json
POST /api/auth/register
{
  "email": "user@example.com",
  "password": "SecureP@ssw0rd",
  "confirm_password": "SecureP@ssw0rd",
  "name": "John Doe",
  "phone": "+1234567890",
  "terms_accepted": true
}
```

#### Register Response (201 Created)
```json
{
  "token": "eyJhbGc...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com",
    "name": "John Doe",
    "phone": "+1234567890",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
}
```

#### Register Error Response (400 Bad Request)
```json
{
  "error": "Validation failed",
  "details": [
    {"field": "email", "reason": "Invalid email format"},
    {"field": "password", "reason": "Password must contain uppercase letter"}
  ]
}
```

#### Register Error Response (409 Conflict)
```json
{
  "error": "Email already registered"
}
```

#### Get Profile Response (200 OK)
```json
GET /api/profile
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "email": "user@example.com",
  "name": "John Doe",
  "phone": "+1234567890",
  "emergency_contact": {
    "name": "Jane Doe",
    "phone": "+0987654321"
  },
  "address": "123 Main St, City, State 12345",
  "profile_picture_url": "https://storage.example.com/profiles/...",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:35:00Z"
}
```

#### Update Profile Request
```json
PUT /api/profile
{
  "name": "John Doe",
  "phone": "+1234567890",
  "emergency_contact": {
    "name": "Jane Doe",
    "phone": "+0987654321"
  },
  "address": "123 Main St, City, State 12345"
}
```

#### Change Password Request
```json
PUT /api/profile/password
{
  "current_password": "OldSecureP@ss1",
  "new_password": "NewSecureP@ss2",
  "confirm_password": "NewSecureP@ss2"
}
```

#### Change Password Response (200 OK)
```json
{
  "message": "Password changed successfully"
}
```

#### Change Password Error (401 Unauthorized)
```json
{
  "error": "Invalid credentials"
}
```

## Validation Rules

### Email Validation
- **Format:** RFC 5322 compliant using regex or validator package
- **Uniqueness:** Database check before creating user
- **Backend:** Use existing validation middleware
- **Error Messages:**
  - "Invalid email format" (400)
  - "Email already registered" (409)

### Password Validation
- **Minimum Length:** 8 characters
- **Uppercase:** At least one A-Z
- **Lowercase:** At least one a-z
- **Digit:** At least one 0-9
- **Special Character:** At least one of !@#$%^&*
- **Backend:** Validate in ProfileUsecase before hashing
- **Error Message:** "Password must be 8+ characters with uppercase, lowercase, number, and special character"

### Phone Validation
- **Optional:** Can be null/empty
- **E.164 Format:** `+` followed by country code and number (e.g., `+1234567890`)
- **Local Format:** Allow local patterns with country inference (future enhancement)
- **Backend:** Use libphonenumber-go or similar package
- **Error Message:** "Invalid phone format (expected: +country-code-number)"

### Name Validation
- **Minimum:** 2 characters
- **Maximum:** 255 characters
- **Allowed Characters:** Letters, spaces, hyphens, apostrophes (no control chars)
- **Error Message:** "Name must be 2-255 characters"

### Emergency Contact
- **Name:** Optional, 2-255 characters if provided
- **Phone:** Optional, E.164 format if provided
- **Validation:** Same as phone validation
- **Error Message:** "Invalid emergency contact format"

### Address Validation
- **Minimum:** 5 characters if provided
- **Maximum:** 500 characters
- **Error Message:** "Address must be 5-500 characters"

## API Endpoints

### 1. POST `/api/auth/register`
- **Authentication:** None (public endpoint)
- **Rate Limiting:** 3 requests per minute per IP
- **Request Body:**
  ```json
  {
    "email": "user@example.com",
    "password": "SecureP@ssw0rd",
    "confirm_password": "SecureP@ssw0rd",
    "name": "John Doe",
    "phone": "+1234567890",
    "terms_accepted": true
  }
  ```
- **Success Response:** 201 Created
  - JWT token for immediate login
  - User object (excluding sensitive fields)
- **Error Responses:**
  - 400 Bad Request: Validation errors (field-level details)
  - 409 Conflict: Email already registered
  - 429 Too Many Requests: Rate limit exceeded
  - 500 Internal Server Error: Server error

### 2. GET `/api/profile`
- **Authentication:** Required (JWT token)
- **Rate Limiting:** Default limit (100 requests/minute per IP)
- **Query Parameters:** None
- **Success Response:** 200 OK
  - User profile object (all fields including optional ones)
  - Excludes password_hash
- **Error Responses:**
  - 401 Unauthorized: Missing/invalid token
  - 404 Not Found: User not found
  - 500 Internal Server Error: Server error

### 3. PUT `/api/profile`
- **Authentication:** Required (JWT token)
- **Rate Limiting:** Default limit
- **Request Body:**
  ```json
  {
    "name": "John Doe",
    "phone": "+1234567890",
    "emergency_contact": {"name": "Jane", "phone": "+0987654321"},
    "address": "123 Main St"
  }
  ```
- **Success Response:** 200 OK
  - Updated user profile object
- **Error Responses:**
  - 400 Bad Request: Validation errors
  - 401 Unauthorized: Invalid token
  - 404 Not Found: User not found
  - 500 Internal Server Error: Server error

### 4. PUT `/api/profile/password`
- **Authentication:** Required (JWT token)
- **Rate Limiting:** Stricter limit (10 requests/minute per IP) - optional
- **Request Body:**
  ```json
  {
    "current_password": "OldPassword123!",
    "new_password": "NewSecureP@ss2",
    "confirm_password": "NewSecureP@ss2"
  }
  ```
- **Success Response:** 200 OK
  ```json
  {
    "message": "Password changed successfully"
  }
  ```
- **Error Responses:**
  - 400 Bad Request: Validation failed (generic message)
  - 401 Unauthorized: Current password incorrect or invalid token (generic)
  - 429 Too Many Requests: Rate limit exceeded
  - 500 Internal Server Error: Server error

### 5. POST `/api/profile/picture` (Optional Enhancement)
- **Authentication:** Required
- **Content-Type:** multipart/form-data
- **File Constraints:** JPEG/PNG, max 5MB
- **Success Response:** 200 OK with new picture URL
- **Error Responses:** 400, 401, 413 (Payload Too Large)

### 6. DELETE `/api/profile` (Optional Enhancement)
- **Authentication:** Required
- **Request Body:** `{password: string}`
- **Success Response:** 204 No Content
- **Error Responses:** 400, 401

## Error Handling

### Error Response Format
All errors return JSON with consistent structure:
```json
{
  "error": "Human-readable error message",
  "details": [
    {
      "field": "email",
      "reason": "Email already registered"
    }
  ],
  "request_id": "abc123..." // for debugging
}
```

### Validation Errors (400 Bad Request)
- Include field-level details showing which field failed and why
- Example: `{field: "password", reason: "Must contain special character"}`
- Never expose regex patterns or internal validation logic

### Authentication Errors (401 Unauthorized)
- Use generic messages: "Invalid credentials"
- Do NOT say "email not found" or "password incorrect"
- Include request_id for security investigation
- Never leak information about other users

### Conflict Errors (409 Conflict)
- "Email already registered" - for duplicate email on registration
- Only used when uniqueness constraints violated

### Rate Limit Errors (429 Too Many Requests)
- Include X-RateLimit-Reset header
- Message: "Too many requests. Try again later."

### Server Errors (500 Internal Server Error)
- Return generic message: "An error occurred"
- Include unique request_id for server-side logging
- Never expose stack traces or database details
- Log actual error server-side for debugging

### Security Best Practices
1. **Sanitize Error Messages:** Never expose database schema, SQL, or infrastructure details
2. **Constant-Time Comparison:** Use constant-time comparison for passwords (not ==)
3. **Generic Auth Failures:** Don't reveal whether email exists or password wrong
4. **Log Security Events:** Log suspicious activity (repeated failures, unusual patterns)
5. **Rate Limiting:** Prevent brute force via rate limiting
6. **HTTPS Only:** Enforce TLS/SSL for all endpoints
7. **Token Expiration:** JWT tokens must have reasonable expiration time
8. **No Token Refresh (P2):** Tokens expire and require re-login (P3 enhancement)

## Implementation Notes

### Reusing P1 Infrastructure

1. **Validation Middleware** (`internal/delivery/http/middleware/validation.go`)
   - Reuse existing validators for field-level validation
   - Enhance with password strength validator
   - Add phone number validator using libphonenumber-go

2. **Error Handler Middleware** (`internal/delivery/http/middleware/error_handler.go`)
   - Use existing error response formatting
   - Ensure consistent HTTP status codes
   - Sanitize error messages (no schema leaks)

3. **Auth Middleware** (`internal/delivery/http/middleware/auth.go`)
   - Reuse JWT validation logic
   - Ensure user context extracted properly
   - Validate token expiration

4. **Database Patterns**
   - Use existing repository pattern for UserRepository
   - Follow existing error handling (wrap errors with context)
   - Use database transactions for atomicity where needed

### Password Hashing Strategy

```go
// In ProfileUsecase
import "golang.org/x/crypto/bcrypt"

// Hash password with cost 12
hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
if err != nil {
  return fmt.Errorf("failed to hash password: %w", err)
}

// Verify password (constant-time comparison)
err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(inputPassword))
if err != nil {
  return errors.New("invalid credentials")
}
```

### Email Validation

```go
import "github.com/go-playground/validator/v10"

validate := validator.New()
err := validate.Var(email, "required,email")
if err != nil {
  return errors.New("invalid email format")
}
```

### Phone Validation

```go
import "github.com/ttacon/libphonenumber-go/phonenumber"

number, _ := phonenumber.Parse(phone, "US") // Auto-detect region
if !phonenumber.IsValidNumber(number) {
  return errors.New("invalid phone format")
}
```

### Testing Strategy

**Backend Tests:**
- Unit tests: Password hashing, validation functions
- Unit tests: ProfileUsecase business logic
- Integration tests: API endpoints with test database
- Integration tests: Error handling scenarios
- Integration tests: Duplicate email handling

**Mobile Tests:**
- Widget tests: Form validation display
- Unit tests: FormValidator helper functions
- Integration tests: API service calls (mock backend)
- UI tests: Profile screen interactions

## Correctness Properties

### Property 1: Email Uniqueness
**Validates: Requirement 1.2**
- For any two users with different IDs, their emails must be different
- Attempting to register with an existing email returns 409 Conflict

### Property 2: Password Strength Enforcement
**Validates: Requirement 1.3**
- Every password stored must satisfy strength requirements (8+, upper, lower, digit, special)
- Weak passwords are rejected at registration and password change

### Property 3: Authenticated Access
**Validates: Requirements 3.2, 3.3, 3.4**
- Profile endpoints require valid JWT token
- User can only access/modify their own profile
- Invalid tokens return 401 Unauthorized

### Property 4: Password Verification
**Validates: Requirement 3.4**
- When changing password, current password must be verified against stored hash
- Incorrect current password returns 401 with generic message
- New password must meet strength requirements

### Property 5: Atomic Profile Updates
**Validates: Requirement 3.3**
- Profile updates are atomic (all or nothing)
- Partial failures don't leave database in inconsistent state
- Validation happens before database write

### Property 6: Error Message Sanitization
**Validates: Requirement 5.2**
- Error responses never expose schema details, SQL, or infrastructure
- Authentication failures use generic messages
- Validation errors are specific (field + reason) without exposure

### Property 7: Rate Limiting on Public Endpoints
**Validates: Requirement 5.4**
- Registration endpoint limited to 3 requests per minute per IP
- When exceeded, returns 429 with X-RateLimit headers
- Other endpoints follow default rate limits

## Security Considerations

### Password Security
- Never log passwords
- Never return passwords in API responses
- Use bcrypt with cost factor 12 (secure, not too slow)
- Implement constant-time comparison for password verification
- Consider password history (P3 enhancement)

### Email Security
- Validate email format before database operations
- Check uniqueness to prevent account enumeration (but safe to say "email exists")
- Consider email verification link (P3 enhancement)

### API Security
- Require HTTPS for all endpoints (enforce in infrastructure)
- Validate all inputs (never trust client)
- Sanitize error messages (no schema leaks)
- Rate limit public endpoints (registration, password change)
- Log security events (failed login attempts, password changes)

### Data Protection
- Never store passwords in plain text
- Encrypt sensitive fields at rest (optional, P3)
- Use HTTPS/TLS for transport security
- Implement request ID for traceability
- Consider audit logging (P3 enhancement)

### Authentication
- Issue JWT tokens with reasonable expiration (e.g., 1 hour)
- Validate token signature on every request
- Extract user ID from token (don't trust client-submitted user_id)
- Implement token refresh mechanism (P3)
- Consider session invalidation on password change

## Performance Considerations

### Database
- Index email column for uniqueness check
- Index user_id for profile queries
- Use connection pooling
- Consider caching frequently accessed profiles (Redis)

### API
- Implement pagination for list endpoints (future)
- Cache validation rules in memory
- Consider CDN for profile pictures (P3)
- Monitor API response times

### Mobile
- Cache profile data locally (SharedPreferences)
- Implement pull-to-refresh for manual sync
- Show loading states during API calls
- Implement exponential backoff for retries
- Consider offline support (P3)

## Deployment Notes

### Environment Configuration
- Database connection string (from .env)
- JWT secret key (secure storage)
- Password hashing cost factor (default: 12)
- Rate limiting configuration
- Email service credentials (for verification, P3)

### Database Migrations
- Run migrations before starting server
- Test rollback scripts
- Schema changes must be backward compatible

### Monitoring & Logging
- Log registration attempts (success/failure)
- Log password changes
- Log failed authentications (with rate limit)
- Monitor API error rates
- Alert on suspicious patterns (multiple failed attempts)

### Security Checklist
- [ ] HTTPS enforced for all endpoints
- [ ] Password hashing uses bcrypt with cost 12
- [ ] Email validation implemented
- [ ] Password strength enforced
- [ ] Rate limiting on public endpoints
- [ ] Error messages sanitized
- [ ] No passwords in logs
- [ ] JWT token validation on protected endpoints
- [ ] Test duplicate email handling
- [ ] Test weak password rejection
- [ ] Test rate limiting
- [ ] Database indexes created
- [ ] Migrations tested
