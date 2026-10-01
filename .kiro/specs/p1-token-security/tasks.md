# P1 #2: Secure Token Storage & Session Lifecycle — Tasks

## Overview
Migrate token storage from SharedPreferences to flutter_secure_storage, implement session lifecycle, handle token expiry, and secure biometric login. 6 main implementation tasks.

## Implementation Plan

### Phase 1: Secure Storage Foundation (Task 1-2)
- Create SecureStorageService with encrypted storage
- Implement SessionManager with token expiry validation

### Phase 2: Service Integration (Task 3-4)
- Update AuthService to use secure storage
- Add API interceptor for 401 handling

### Phase 3: UI & Testing (Task 5-6)
- Create SplashScreen and BiometricUnlockScreen
- Write unit/integration tests and commit

### Estimated Timeline
- Phase 1: 1 day
- Phase 2: 1 day
- Phase 3: 1 day
- **Total: 3 days**

## Task Dependency Graph

```json
{
  "waves": [
    { "wave": 1, "tasks": ["1", "2"] },
    { "wave": 2, "tasks": ["3"] },
    { "wave": 3, "tasks": ["4"] },
    { "wave": 4, "tasks": ["5"] },
    { "wave": 5, "tasks": ["6"] }
  ]
}
```

**Wave Explanation:**
- **Wave 1:** Task 1 (SecureStorageService) and Task 2 (SessionManager) can run in parallel
- **Wave 2:** Task 3 (AuthService) depends on Task 1 & 2
- **Wave 3:** Task 4 (API Interceptor) depends on Task 3
- **Wave 4:** Task 5 (SplashScreen & BiometricUnlockScreen) depends on Task 2 & 4
- **Wave 5:** Task 6 (Testing & Commit) depends on all prior tasks (1-5)

## Tasks

- [ ] 1. Create SecureStorageService with encryption
  - Create `presensigo_mobile/lib/data/services/secure_storage_service.dart`
  - Implement 8 methods: `saveToken()`, `getToken()`, `deleteToken()`, `saveDeviceId()`, `getDeviceId()`, `setBiometricEnabled()`, `isBiometricEnabled()`, `clearAllCredentials()`
  - Configure FlutterSecureStorage with Android KeyStore (RSA_ECB_OAEPwithSHA_256, AES_GCM_NoPadding) and iOS Keychain (first_this_device_this_device_only)
  - Add error handling for storage failures
  - **Acceptance Criteria:**
    - SecureStorageService created with all 8 methods
    - FlutterSecureStorage platform options configured
    - No token data in logs or console
    - clearAllCredentials() atomically removes all keys

- [ ] 2. Implement SessionManager for token validation
  - Create `presensigo_mobile/lib/data/services/session_manager.dart`
  - Implement `isSessionValid()`: decode JWT, extract `exp` claim, compare with current time + 60s buffer
  - Implement `logout()`: call `clearAllCredentials()` and clear SharedPreferences
  - Implement `handleUnauthorized()`: call logout() and prepare for redirect
  - Implement `_decodeJwt()`: Base64 decode with padding normalization
  - Add unit tests for token expiry detection
  - **Acceptance Criteria:**
    - SessionManager created with all methods
    - JWT decoding handles Base64 padding correctly
    - isSessionValid() returns false for expired tokens (60s buffer)
    - isSessionValid() returns false for missing tokens
    - logout() clears all session state
    - Unit tests pass for expiry detection

- [ ] 3. Update AuthService with secure storage integration
  - Modify `presensigo_mobile/lib/data/services/api_service.dart` (or auth service equivalent)
  - Add `_sessionManager` and `_secureStorage` instances
  - Update `login()` to save token via `_secureStorage.saveToken()`
  - Add `getValidToken()`: check session validity, return token if valid, else trigger logout
  - Update `logout()` to call `_sessionManager.logout()`
  - Add optional POST `/api/auth/logout` call with error handling
  - **Acceptance Criteria:**
    - AuthService uses SecureStorageService for token operations
    - login() saves token to secure storage
    - logout() clears all credentials
    - getValidToken() validates session before returning
    - API tests verify login/logout cycle

- [ ] 4. Add API interceptor for 401 response handling
  - Modify `presensigo_mobile/lib/data/services/api_service.dart` (or Dio interceptor)
  - Add `_handleResponse()` helper to detect 401 status
  - On 401: call `_sessionManager.handleUnauthorized()`, clear token, redirect to login
  - Show user snackbar: "Session expired, please log in again"
  - Apply to all API methods: `getTodayAttendance()`, `getHistory()`, `checkIn()`, `checkOut()`, `getLocations()`
  - Prevent infinite redirect loops
  - **Acceptance Criteria:**
    - 401 responses trigger session logout
    - 401 responses redirect UI to login
    - User sees "Session expired" message
    - No infinite redirect loops
    - Integration test: API 401 → logout → redirect

- [ ] 5. Create SplashScreen and BiometricUnlockScreen
  - Create `presensigo_mobile/lib/features/auth/screens/splash_screen.dart`
    - Check session validity on init
    - Route to BiometricUnlockScreen if biometric enabled
    - Route to AttendanceScreen if biometric disabled + valid token
    - Route to LoginScreen if no valid session
    - 500ms delay to avoid visual jank
  - Create `presensigo_mobile/lib/features/auth/screens/biometric_unlock_screen.dart`
    - Show biometric prompt (Face ID or Fingerprint)
    - Validate token freshness before unlocking
    - Allow max 3 retry attempts
    - On 3rd failure: fallback to LoginScreen
    - Show "Session expired" dialog if token expired
    - Provide "Use Password Login" button
  - Update `presensigo_mobile/lib/main.dart`: change initialRoute to `/splash`
  - **Acceptance Criteria:**
    - SplashScreen checks session validity
    - SplashScreen routes correctly based on session state
    - BiometricUnlockScreen shows biometric prompt
    - BiometricUnlockScreen validates token freshness
    - BiometricUnlockScreen allows 3 retries then fallback
    - User can manually choose password login
    - Integration tests for valid/expired token scenarios

- [ ] 6. Write tests, verify build, and commit
  - Create `presensigo_mobile/test/services/session_manager_test.dart`
    - Unit test: missing token → isSessionValid() = false
    - Unit test: expired token → isSessionValid() = false
    - Unit test: valid token → isSessionValid() = true (60s buffer applied)
    - Unit test: logout → clears all credentials
    - Unit test: handleUnauthorized → calls logout
    - Unit test: logout idempotency (multiple calls safe)
  - Run `flutter analyze` → 0 errors, warnings fixed
  - Run `flutter test` → all tests pass
  - Manual verification:
    - Login with valid credentials → token in secure storage (verify with adb or xcode)
    - Logout → token cleared from storage
    - App resume with valid token → navigates to attendance/biometric
    - App resume with expired token → redirects to login
    - Biometric enabled + valid token → shows biometric unlock → enters attendance
    - Biometric enabled + expired token → shows "Session expired" dialog
    - API returns 401 → session cleared, redirected to login
    - Logout + enable biometric → biometric flag cleared, next app open shows login
  - Commit to `fix/validate-middleware` branch:
    ```bash
    git add presensigo_mobile/lib/data/services/secure_storage_service.dart
    git add presensigo_mobile/lib/data/services/session_manager.dart
    git add presensigo_mobile/lib/data/services/api_service.dart
    git add presensigo_mobile/lib/features/auth/screens/splash_screen.dart
    git add presensigo_mobile/lib/features/auth/screens/biometric_unlock_screen.dart
    git add presensigo_mobile/lib/main.dart
    git add presensigo_mobile/test/services/session_manager_test.dart
    git add .kiro/specs/p1-token-security/
    git commit -m "feat: implement secure token storage and session lifecycle management"
    git push
    ```
  - **Acceptance Criteria:**
    - `flutter analyze` passes
    - `flutter test` passes
    - Manual test flow all steps pass
    - All files committed to branch
    - Commit message clear and descriptive

## Notes

### Security Considerations
- Tokens are stored in platform-level encrypted storage (Keychain/Keystore), not SharedPreferences
- Token expiry validated with 60s clock skew buffer before use
- Automatic logout on 401 responses prevents stale token usage
- Biometric unlock is local auth only; tokens never transmitted to biometric subsystem
- Logout atomically clears all credentials (token, device ID, biometric flag)
- Error handling prioritizes security: on storage access failure, treat as logout

### Design Patterns
- **SecureStorageService**: Single responsibility for persistence to encrypted storage
- **SessionManager**: Business logic for token lifecycle and validation
- **API Service pattern**: Interceptor/helper function for consistent 401 handling
- **Screen routing**: SplashScreen as entry point determines app state on resume
- **Biometric flow**: Local auth only; session/token validation happens after biometric succeeds

### Testing Strategy
- Unit tests focus on SessionManager JWT decoding, expiry validation, and logout
- Integration tests verify end-to-end flows: login → token saved → logout → token cleared
- Manual tests verify platform-specific features: secure storage encryption, biometric prompt
- No mocking of secure storage in tests; use real FlutterSecureStorage for platform-level verification

### Rollout Plan
1. Deploy P1 #2 to staging environment
2. Verify secure storage encryption works on iOS/Android devices
3. Test biometric prompt on physical devices (emulator biometric unreliable)
4. Monitor 401 response handling in logs
5. Gradually roll out to production with feature flag if possible

### Known Limitations
- SharedPreferences migration fallback only; new installs skip SharedPreferences
- Biometric prompt unavailable on emulators; manual testing required on physical devices
- Token refresh not implemented (Option A: immediate logout on expiry); can be added in future
- Device UUID stored in secure storage but not used for session binding (P1 #1 handles binding)
