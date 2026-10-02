# Requirements Document: Mobile Profile & Register Flow

## Introduction

This specification defines the requirements for implementing a comprehensive user registration and profile management system in PresensiGo. The feature encompasses a mobile-friendly registration flow with email/phone validation and a dedicated profile screen enabling users to view and edit their account information, manage emergency contacts, and change passwords. The backend provides secure REST endpoints with robust validation, error handling, and follows established PresensiGo security patterns. This milestone builds on P1 authentication infrastructure and introduces user profile persistence and self-service account management.

## Overview

Enable users to register for PresensiGo with validated credentials, log in, and manage their profile information through a mobile interface. Provide backend endpoints for registration, profile retrieval, and profile updates with comprehensive validation and error handling. Implement password strength requirements, email uniqueness validation, and optional profile enhancements like profile pictures and emergency contacts.

## Requirements

### R1: User Registration Flow

1.1 **Registration Form (Mobile):** Provide a user-friendly registration screen with the following fields:
   - Email address (required, validated for format and uniqueness)
   - Password (required, validated for strength)
   - Confirm password (required, must match password field)
   - Full name (required, 2-255 characters)
   - Phone number (optional, validated for format)
   - Terms & conditions checkbox (required, must be checked to proceed)

1.2 **Email Validation:** 
   - Format validation: RFC 5322 compliant email format
   - Uniqueness: Backend must reject duplicate emails with 409 Conflict
   - Verification: Optional email verification link (future enhancement)

1.3 **Password Requirements:**
   - Minimum 8 characters
   - Must contain uppercase letter (A-Z)
   - Must contain lowercase letter (a-z)
   - Must contain number (0-9)
   - Must contain special character (!@#$%^&*)
   - Password confirmation must match

1.4 **Phone Number Validation:**
   - Optional field
   - Support E.164 format (international: +1234567890)
   - Support local format with country inference
   - Validate format but not carrier existence

1.5 **Registration Submission:**
   - POST `/api/auth/register` endpoint creates new user
   - Response includes user ID, email, name
   - Auto-login after successful registration (return JWT token)
   - Auto-navigate to profile screen after registration

1.6 **Error Handling (R1):**
   - Invalid email format: 400 Bad Request with field-level error
   - Email already exists: 409 Conflict with clear message
   - Weak password: 400 Bad Request with password policy details
   - Phone format invalid: 400 Bad Request with format hint
   - Password mismatch: Client-side validation before submit
   - Required fields missing: 400 Bad Request with field list

### R2: User Profile Screen

2.1 **Profile Display (Mobile):**
   - Show authenticated user's profile information:
     - Full name
     - Email address (read-only)
     - Phone number
     - Account creation date
     - Last updated date
     - Optional profile picture (if set)
   - Display in read-only mode initially

2.2 **Profile Edit Mode:**
   - Edit button (pencil icon) to enter edit mode
   - Allow editing of:
     - Full name (2-255 characters)
     - Phone number (optional, validate format)
     - Emergency contact name (optional, 2-255 characters)
     - Emergency contact phone (optional, validate E.164 format)
     - Address (optional, 5-500 characters)
     - Profile picture (optional, upload JPEG/PNG, max 5MB)
   - Save button to submit changes
   - Cancel button to discard changes without saving

2.3 **Profile Picture Management:**
   - Upload image from camera or gallery
   - Display preview before upload
   - Compress image (max 5MB, optimized for mobile storage)
   - Store in backend object storage (MinIO)
   - Display profile picture in profile header

2.4 **Change Password Feature:**
   - Change password button opens secure form
   - Require entry of current password (for verification)
   - Require entry of new password (validate strength)
   - Require confirmation of new password
   - Validate new password meets strength requirements
   - Prevent password reuse (optional enhancement)
   - Show password strength indicator during entry

2.5 **Profile Data Persistence:**
   - Load profile data from GET `/api/profile` on screen open
   - Cache profile data locally (SharedPreferences for Flutter)
   - Refresh data on pull-to-refresh gesture
   - Show loading state while fetching data

### R3: Backend Profile Endpoints

3.1 **POST `/api/auth/register` (Public)**
   - Accept JSON: `{email, password, confirmPassword, name, phone, termsAccepted}`
   - Validate all fields per R1 requirements
   - Hash password using bcrypt (cost factor: 12)
   - Create user record in database
   - Return 201 Created with JWT token and user object
   - Return 400 Bad Request if validation fails (with field-level details)
   - Return 409 Conflict if email already exists

3.2 **GET `/api/profile` (Authenticated)**
   - Return JSON: `{id, email, name, phone, emergencyContact*, address*, profilePictureUrl*, createdAt, updatedAt}`
   - Require valid JWT token (401 Unauthorized if missing/invalid)
   - Return user profile for authenticated user only
   - Return 200 OK with user object
   - Return 401 Unauthorized if token invalid/expired
   - Return 404 Not Found if user deleted

3.3 **PUT `/api/profile` (Authenticated)**
   - Accept JSON: `{name, phone, emergencyContact, address, ...}`
   - Require valid JWT token
   - Validate all fields (same rules as edit requirements)
   - Update user record atomically
   - Return 200 OK with updated user object
   - Return 400 Bad Request if validation fails
   - Return 401 Unauthorized if token invalid
   - Return 409 Conflict if trying to update email (not allowed via this endpoint)

3.4 **PUT `/api/profile/password` (Authenticated)**
   - Accept JSON: `{currentPassword, newPassword, confirmPassword}`
   - Require valid JWT token
   - Verify current password matches stored hash
   - Validate new password meets strength requirements
   - Validate confirmation matches new password
   - Update password hash in database
   - Invalidate all active sessions (user must re-login) - optional for P2
   - Return 200 OK with success message
   - Return 400 Bad Request if validation fails (generic message, no password hints)
   - Return 401 Unauthorized if current password incorrect (generic message)
   - Return 401 Unauthorized if token invalid

3.5 **POST `/api/profile/picture` (Authenticated) - Optional Enhancement**
   - Accept multipart form data with image file
   - Validate file type (JPEG/PNG only)
   - Validate file size (max 5MB)
   - Upload to MinIO object storage
   - Update user profile_picture_url
   - Return 200 OK with new picture URL

3.6 **DELETE `/api/profile` (Authenticated) - Optional Enhancement**
   - Require valid JWT token
   - Require password confirmation for security
   - Delete user account and related data
   - Return 204 No Content on success
   - Return 400 Bad Request if password incorrect

### R4: Form Validation (Frontend & Backend)

4.1 **Email Validation:**
   - Frontend: Real-time format validation (visual feedback)
   - Backend: RFC 5322 format, uniqueness check
   - Error message: "Invalid email format" or "Email already registered"

4.2 **Password Validation:**
   - Frontend: Real-time strength indicator showing requirements met
   - Backend: Enforce strength rules (min 8, upper, lower, number, special)
   - Error message: "Password must contain uppercase, lowercase, number, and special character"

4.3 **Phone Validation:**
   - Frontend: Real-time format guidance
   - Backend: Validate E.164 or local format
   - Error message: "Invalid phone format (expected: +1234567890 or local format)"

4.4 **Name Validation:**
   - Frontend: Character count indicator
   - Backend: Min 2, max 255 characters
   - Error message: "Name must be 2-255 characters"

4.5 **Emergency Contact:**
   - Phone: Optional, validate E.164 format if provided
   - Name: Optional, 2-255 characters if provided

4.6 **Address Validation:**
   - Optional field
   - Min 5, max 500 characters if provided

### R5: Error Handling & Security

5.1 **Error Responses (Backend):**
   - 400 Bad Request: Field-level validation errors with specific reasons
   - 401 Unauthorized: Missing/invalid token or incorrect credentials
   - 409 Conflict: Duplicate email registration
   - 500 Internal Server Error: Server-side errors
   - All errors return JSON: `{error: string, details?: [{field, reason}]}`

5.2 **Error Messages (Security):**
   - Never expose database schema details
   - Never reveal query structure
   - Generic messages for authentication failures ("Invalid credentials")
   - Specific messages for validation failures (field + rule)
   - Never leak information about other users' emails
   - For password change: use generic "Invalid credentials" message if wrong current password

5.3 **Sensitive Data Handling:**
   - Passwords never returned in API responses
   - Password hash never exposed
   - Email verification tokens (if used) must be single-use and time-limited
   - API logs must not contain passwords or tokens
   - Client must not store passwords in SharedPreferences (only JWT tokens)

5.4 **Rate Limiting:**
   - Apply rate limiting to registration endpoint (3 attempts per minute per IP)
   - Apply rate limiting to password change endpoint
   - Return 429 Too Many Requests when limit exceeded

5.5 **HTTPS/TLS:**
   - All profile endpoints require HTTPS (enforced by infrastructure)
   - Certificate validation on mobile client
   - No HTTP fallback for production

### R6: Optional Features (Future Enhancements)

6.1 **Profile Picture Management:**
   - Upload profile picture from camera or gallery
   - Store in MinIO object storage
   - Display in profile header and attendance history
   - Compress and optimize for mobile viewing

6.2 **Emergency Contact:**
   - Store emergency contact name and phone
   - Display in profile
   - Use for notifications in future phases

6.3 **Address Storage:**
   - Store user address for context
   - Display in profile
   - Future use: organize attendance by location

6.4 **Account Deletion:**
   - Self-service account deletion with password confirmation
   - Cascade delete related data
   - Send confirmation email before permanent deletion

## Success Criteria

- User can complete registration with valid email, password, name, and phone
- User receives validation errors for invalid email format
- User receives specific validation errors for weak passwords
- User receives 409 error when email already registered
- User can log in immediately after registration
- User can view their profile after login (GET `/api/profile` returns correct data)
- User can edit name, phone, emergency contact, and address
- User can change password with current password verification
- User receives errors for invalid phone format or weak new password
- All backend validation is enforced (not just frontend)
- All error messages are sanitized (no schema leaks)
- Registration rate limiting prevents spam (3/minute per IP)
- All tests pass (unit + integration)
- Build succeeds and app runs without crashes
- Profile data persists across app restarts
- All code follows PresensiGo patterns (validation, error handling, auth middleware)

## Glossary

- **JWT Token:** JSON Web Token used for stateless authentication
- **RFC 5322:** Internet standard for email address format validation
- **E.164 Format:** International telephone numbering format (+country-code-number)
- **Bcrypt:** Password hashing algorithm with salt and configurable cost factor
- **Rate Limiting:** Restricting number of requests per client in a time window
- **HTTPS/TLS:** Encrypted communication protocol for secure data transmission
- **MinIO:** Object storage service compatible with Amazon S3 API
- **Atomic Update:** Database operation that completes entirely or not at all
- **Cascade Delete:** Automatically deleting related data when parent is deleted
- **Session Invalidation:** Revoking authentication tokens to force re-login
- **Salt:** Random data added to passwords before hashing to prevent rainbow tables
- **Cost Factor:** Parameter controlling bcrypt hashing difficulty (recommended: 12)
- **Circuit Breaker:** Pattern to prevent cascading failures when service unavailable
