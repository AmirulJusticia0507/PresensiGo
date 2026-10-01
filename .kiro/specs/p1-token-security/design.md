# P1 #2: Secure Token Storage & Session Lifecycle — Design

## Architecture Overview

### Session State Model
```
User
├─ JWT Token (secure storage)
│  ├─ Header: alg, typ
│  ├─ Payload: user_id, role, exp, iat
│  └─ Signature
├─ Device UUID (secure storage)
├─ Biometric Enabled Flag (SharedPreferences or secure storage)
└─ Last Valid Session Time (local timestamp)
```

### Session Lifecycle State Machine
```
START
  ↓
[Login Screen] ← (No token)
  ├─ Email/password input
  ├─ POST /api/auth/login
  ├─ Save JWT to secure storage
  └─ → [Attendance Screen]
  
[Resume from Background]
  ├─ Check token existence & expiry
  ├─ If expired or missing → [Login Screen]
  ├─ If valid → [Last Screen]
  └─ (Optional: Biometric unlock if enabled and token fresh)
  
[Biometric Unlock]
  ├─ Read token from secure storage
  ├─ Verify NOT expired
  ├─ Show biometric prompt
  ├─ If match → [Attendance Screen]
  ├─ If fail → show retry option or redirect to password login
  └─ Max retries 3 → fallback to password login
  
[Logout]
  ├─ POST /api/auth/logout (optional, server-side revocation)
  ├─ Clear JWT from secure storage
  ├─ Clear device UUID
  ├─ Clear biometric flag
  ├─ Clear all user session state
  └─ → [Login Screen]
  
[Token Expired via 401]
  ├─ Detect 401 response from API
  ├─ Clear JWT from secure storage
  ├─ Clear biometric flag
  ├─ Show snackbar: "Session expired, please log in again"
  └─ → [Login Screen]
```

## Technical Implementation

### 1. Secure Storage Service (Flutter)

File: `presensigo_mobile/lib/services/secure_storage_service.dart`

```dart
class SecureStorageService {
  static const _tokenKey = 'jwt_token';
  static const _deviceIdKey = 'device_uuid';
  static const _biometricKey = 'biometric_enabled';
  
  final _storage = FlutterSecureStorage(
    aOptions: AndroidOptions(
      keyCipherAlgorithm: KeyCipherAlgorithm.RSA_ECB_OAEPwithSHA_256andMGF1Padding,
      storageCipherAlgorithm: StorageCipherAlgorithm.AES_GCM_NoPadding,
    ),
  );
  
  Future<void> saveToken(String token) async {
    await _storage.write(key: _tokenKey, value: token);
  }
  
  Future<String?> getToken() async {
    return await _storage.read(key: _tokenKey);
  }
  
  Future<void> deleteToken() async {
    await _storage.delete(key: _tokenKey);
  }
  
  Future<void> saveDeviceId(String deviceId) async {
    await _storage.write(key: _deviceIdKey, value: deviceId);
  }
  
  Future<String?> getDeviceId() async {
    return await _storage.read(key: _deviceIdKey);
  }
  
  Future<void> setBiometricEnabled(bool enabled) async {
    await _storage.write(key: _biometricKey, value: enabled.toString());
  }
  
  Future<bool> isBiometricEnabled() async {
    final val = await _storage.read(key: _biometricKey);
    return val == 'true';
  }
  
  Future<void> clearAllCredentials() async {
    await deleteToken();
    await _storage.delete(key: _deviceIdKey);
    await _storage.delete(key: _biometricKey);
  }
}
```

### 2. Session Manager Service (Flutter)

File: `presensigo_mobile/lib/services/session_manager.dart`

```dart
class SessionManager {
  final _secureStorage = SecureStorageService();
  
  /// Check if session is still valid (token exists and not expired)
  Future<bool> isSessionValid() async {
    final token = await _secureStorage.getToken();
    if (token == null) return false;
    
    // Decode JWT and check expiry
    try {
      final parts = token.split('.');
      if (parts.length != 3) return false;
      
      final payload = _decodeJwt(parts[1]);
      final exp = payload['exp'] as int?;
      
      if (exp == null) return false;
      
      final expiryTime = DateTime.fromMillisecondsSinceEpoch(exp * 1000);
      final now = DateTime.now();
      
      // Add 60 second buffer to avoid edge cases
      return now.isBefore(expiryTime.subtract(Duration(seconds: 60)));
    } catch (e) {
      return false;
    }
  }
  
  /// Logout: clear all credentials and session state
  Future<void> logout() async {
    await _secureStorage.clearAllCredentials();
    // Also clear any other session state from SharedPreferences if needed
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('last_screen');
    await prefs.remove('user_profile');
  }
  
  /// Handle 401 response: session expired
  Future<void> handleUnauthorized() async {
    await logout();
    // Notify UI to redirect to login
    // This can be done via event stream, GetX, Provider, etc.
  }
  
  Map<String, dynamic> _decodeJwt(String payload) {
    // Add padding if needed
    String normalized = payload.replaceAll('-', '+').replaceAll('_', '/');
    while (normalized.length % 4 != 0) {
      normalized += '=';
    }
    
    final decoded = utf8.decode(base64Url.decode(normalized));
    return jsonDecode(decoded) as Map<String, dynamic>;
  }
}
```

### 3. Auth Service Update (Flutter)

Modify `presensigo_mobile/lib/services/auth_service.dart`:

```dart
class AuthService {
  final _sessionManager = SessionManager();
  final _secureStorage = SecureStorageService();
  
  Future<bool> login(String email, String password) async {
    try {
      final response = await apiClient.post(
        '/api/auth/login',
        data: {'email': email, 'password': password},
      );
      
      final token = response.data['token'] as String?;
      if (token != null) {
        await _secureStorage.saveToken(token);
        return true;
      }
      return false;
    } catch (e) {
      // Handle login error
      return false;
    }
  }
  
  Future<void> logout() async {
    // Optional: Notify backend of logout (for audit/revocation)
    try {
      await apiClient.post('/api/auth/logout');
    } catch (e) {
      // Ignore errors; local logout always succeeds
    }
    
    await _sessionManager.logout();
  }
  
  Future<String?> getValidToken() async {
    if (await _sessionManager.isSessionValid()) {
      return await _secureStorage.getToken();
    }
    // Token invalid or expired
    await _sessionManager.handleUnauthorized();
    return null;
  }
}
```

### 4. API Interceptor Update (Flutter)

```dart
class ApiInterceptor extends Interceptor {
  final _authService = AuthService();
  
  @override
  void onResponse(Response response, ResponseInterceptorHandler handler) {
    if (response.statusCode == 401) {
      // Token expired or invalid
      _authService.handleUnauthorized();
      // Optionally emit event to redirect to login
    }
    handler.next(response);
  }
}
```

### 5. Splash/Auto-Login Screen (Flutter)

File: `presensigo_mobile/lib/screens/splash_screen.dart`

```dart
class SplashScreen extends StatefulWidget {
  @override
  State<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends State<SplashScreen> {
  final _sessionManager = SessionManager();
  
  @override
  void initState() {
    super.initState();
    _checkSession();
  }
  
  Future<void> _checkSession() async {
    // Wait a bit for app to fully initialize
    await Future.delayed(Duration(milliseconds: 500));
    
    if (await _sessionManager.isSessionValid()) {
      // Session valid, check if biometric should be used
      final prefs = await SharedPreferences.getInstance();
      final useBiometric = prefs.getBool('use_biometric') ?? false;
      
      if (useBiometric) {
        // Navigate to biometric unlock screen
        if (mounted) {
          Navigator.of(context).pushReplacementNamed('/biometric-unlock');
        }
      } else {
        // Navigate to last screen (attendance)
        if (mounted) {
          Navigator.of(context).pushReplacementNamed('/attendance');
        }
      }
    } else {
      // No valid session, go to login
      if (mounted) {
        Navigator.of(context).pushReplacementNamed('/login');
      }
    }
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: CircularProgressIndicator(),
      ),
    );
  }
}
```

### 6. Biometric Unlock Screen (Flutter)

File: `presensigo_mobile/lib/screens/biometric_unlock_screen.dart`

```dart
class BiometricUnlockScreen extends StatefulWidget {
  @override
  State<BiometricUnlockScreen> createState() => _BiometricUnlockScreenState();
}

class _BiometricUnlockScreenState extends State<BiometricUnlockScreen> {
  final _biometricService = BiometricService();
  final _sessionManager = SessionManager();
  int _retryCount = 0;
  static const _maxRetries = 3;
  
  @override
  void initState() {
    super.initState();
    _startBiometricPrompt();
  }
  
  Future<void> _startBiometricPrompt() async {
    try {
      final isAuthenticated = await _biometricService.authenticate();
      
      if (isAuthenticated) {
        // Verify token is still valid
        final isValid = await _sessionManager.isSessionValid();
        if (isValid) {
          // Proceed to attendance
          if (mounted) {
            Navigator.of(context).pushReplacementNamed('/attendance');
          }
        } else {
          // Token expired, redirect to login
          _showSessionExpiredDialog();
        }
      } else {
        // Biometric failed, increment retry
        _retryCount++;
        
        if (_retryCount >= _maxRetries) {
          // Max retries reached, fallback to password login
          if (mounted) {
            Navigator.of(context).pushReplacementNamed('/login');
          }
        } else {
          // Allow retry
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Biometric failed. Retries: $_retryCount/$_maxRetries')),
          );
          await Future.delayed(Duration(seconds: 1));
          _startBiometricPrompt();
        }
      }
    } catch (e) {
      // Biometric error (sensor offline, etc.)
      _showBiometricErrorDialog();
    }
  }
  
  void _showSessionExpiredDialog() {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        title: Text('Session Expired'),
        content: Text('Your session has expired. Please log in again.'),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.of(context).pushReplacementNamed('/login');
            },
            child: Text('OK'),
          ),
        ],
      ),
    );
  }
  
  void _showBiometricErrorDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('Biometric Error'),
        content: Text('Biometric authentication failed. Please use password login.'),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.of(context).pushReplacementNamed('/login');
            },
            child: Text('Login with Password'),
          ),
        ],
      ),
    );
  }
  
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('Unlock with Biometric')),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.fingerprint, size: 80),
            SizedBox(height: 16),
            Text('Authenticating...'),
            SizedBox(height: 32),
            TextButton(
              onPressed: () {
                Navigator.of(context).pushReplacementNamed('/login');
              },
              child: Text('Use Password Login'),
            ),
          ],
        ),
      ),
    );
  }
}
```

### 7. Backend Changes (Optional but Recommended)

If implementing logout on backend (for audit trail):

```go
// handler.go
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
  userID := getUserIDFromContext(r)
  if userID == uuid.Nil {
    respondError(w, http.StatusUnauthorized, "unauthorized")
    return
  }
  
  // Optional: Log logout event or invalidate token on server
  // For now, just return success (token will be cleared on client)
  
  respondJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// In RegisterRoutes:
r.HandleFunc("/api/auth/logout", h.Logout).Methods("POST")
```

## Data Flow

### Login Flow
```
User inputs credentials
  ↓
POST /api/auth/login (email, password)
  ↓
Backend validates, returns JWT
  ↓
Mobile saves JWT to SecureStorage
  ↓
Navigate to AttendanceScreen
```

### Resume Flow
```
App opened from background
  ↓
SplashScreen checks token validity
  ↓
If valid:
  - Check biometric enabled
  - If yes: show BiometricUnlockScreen
  - If no: navigate to AttendanceScreen
  ↓
If invalid or missing:
  - Navigate to LoginScreen
```

### Logout Flow
```
User taps Logout
  ↓
POST /api/auth/logout (optional)
  ↓
Clear SecureStorage (token, device_id, biometric flag)
  ↓
Navigate to LoginScreen
```

### API 401 Response Flow
```
API request returns 401
  ↓
Interceptor detects 401
  ↓
Call SessionManager.handleUnauthorized()
  ↓
Clear SecureStorage
  ↓
Emit redirect event to LoginScreen
  ↓
Show snackbar: "Session expired, please log in again"
```

## Security Considerations

1. **Secure Storage**: FlutterSecureStorage uses platform-level encryption (Keychain on iOS, Keystore on Android)
2. **Token Expiry**: Always verify `exp` claim; add buffer (60s) for clock skew
3. **Biometric**: Is local auth only; never transmit biometric data to server
4. **Logout**: Clear all state immediately; no async operations that might fail
5. **401 Handling**: Treat any 401 as session loss; don't retry with same token
6. **Logs**: Never log tokens or sensitive data

## Testing Strategy

- Unit test: Token expiry detection with various timestamps
- Unit test: Logout clears all state
- Integration test: Login → check token in SecureStorage → logout → token gone
- Integration test: Biometric unlock validates token expiry
- Integration test: API 401 response triggers logout and redirect
