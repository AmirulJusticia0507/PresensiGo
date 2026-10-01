# P1 #2: Secure Token Storage & Session Lifecycle

## Overview
Migrate JWT token storage from insecure `SharedPreferences` to encrypted `flutter_secure_storage`, implement proper session lifecycle management, handle token expiry with automatic redirect to login, implement secure logout, and fix biometric login to validate token freshness.

## Requirements

### 1. Secure Token Storage
1.1 Replace `SharedPreferences` with `flutter_secure_storage` for JWT storage
1.2 Ensure `flutter_secure_storage` dependency is added (already present in pubspec.yaml)
1.3 Migrate existing token save/load logic to use secure storage
1.4 Tokens are never logged or printed to console
1.5 Token deletion is secure (zeroed before removal, not just soft delete)

### 2. Session Lifecycle Management
2.1 Implement session validity check on app resume (in main.dart or splash screen)
2.2 Detect token expiry: extract `exp` claim and compare with current time
2.3 If token expired or missing, redirect to login screen automatically
2.4 If token valid, allow resume to last screen (attendance or profile)
2.5 Session state must be cleared completely on logout

### 3. Token Expiry & Refresh Flow
3.1 On API response with 401 Unauthorized, treat as token expired
3.2 Option A (Simple): Clear all credentials, redirect to login immediately
3.3 Option B (Advanced): Implement refresh token endpoint if backend supports it
3.4 For MVP, implement Option A (immediate redirect to login)
3.5 Ensure user is notified before logout due to expiry

### 4. Logout Implementation
4.1 Logout button clears JWT from secure storage
4.2 Logout button clears device UUID if stored
4.3 Logout button clears biometric flag (`biometric_logged_in`)
4.4 Logout button clears all other session state (user profile, preferences)
4.5 Logout redirects to login screen
4.6 All credential files must be removed before navigation (no race condition)

### 5. Biometric Login Security
5.1 Biometric unlock is only a local authentication, not a login
5.2 Before granting biometric access, verify stored token is present and NOT expired
5.3 If token missing or expired, do not allow biometric unlock; redirect to login
5.4 Biometric flag is cleared on logout
5.5 Biometric prompt must occur each time user opens app (not just on first install)
5.6 Handle case where user disables biometric in device settings (graceful fallback to password)

### 6. Error Handling & User Feedback
6.1 Token expiry is user-friendly: "Session expired, please log in again" notification
6.2 Network errors during session check are handled gracefully (retry or fall back to login)
6.3 If secure storage access fails, treat as logout (security-first)
6.4 Clear error messages for biometric failures (sensor offline, biometric invalid, etc.)

## Success Criteria
- Token stored in secure storage, not `SharedPreferences`
- App resume checks token expiry and redirects to login if expired
- Logout clears all credentials and biometric state completely
- Biometric unlock validates token freshness before granting access
- All 401 responses redirect to login after clearing credentials
- User can re-login and continue workflow without data corruption
- No token leaks in logs or console output
