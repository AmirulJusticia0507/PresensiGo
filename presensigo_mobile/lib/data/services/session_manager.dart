import 'dart:convert';
import 'package:shared_preferences/shared_preferences.dart';
import 'secure_storage_service.dart';

class SessionManager {
  late final SecureStorageService _secureStorage;

  SessionManager() {
    _secureStorage = SecureStorageService();
  }

  /// Check if session is still valid (token exists and not expired)
  /// Returns true only if token exists and is not expired (with 60s buffer)
  Future<bool> isSessionValid() async {
    try {
      final token = await _secureStorage.getToken();
      if (token == null) return false;

      // Decode JWT and check expiry
      final parts = token.split('.');
      if (parts.length != 3) return false;

      final payload = _decodeJwt(parts[1]);
      final exp = payload['exp'] as int?;

      if (exp == null) return false;

      final expiryTime = DateTime.fromMillisecondsSinceEpoch(exp * 1000);
      final now = DateTime.now();

      // Add 60 second buffer to avoid edge cases
      return now.isBefore(expiryTime.subtract(const Duration(seconds: 60)));
    } catch (e) {
      return false;
    }
  }

  /// Logout: clear all credentials and session state
  Future<void> logout() async {
    try {
      await _secureStorage.clearAllCredentials();

      // Also clear any other session state from SharedPreferences if needed
      final prefs = await SharedPreferences.getInstance();
      await Future.wait([
        prefs.remove('last_screen'),
        prefs.remove('user_profile'),
        prefs.remove('use_biometric'),
      ]);
    } catch (e) {
      // Ensure logout succeeds even if there are errors
      // This is a security-first approach
    }
  }

  /// Handle 401 response: session expired
  /// Clears all credentials and prepares for login redirect
  Future<void> handleUnauthorized() async {
    await logout();
    // Notify UI to redirect to login via event stream if implemented
  }

  /// Decode JWT payload from Base64
  /// Handles padding normalization for URL-safe Base64
  Map<String, dynamic> _decodeJwt(String payload) {
    try {
      // Add padding if needed
      String normalized = payload.replaceAll('-', '+').replaceAll('_', '/');
      while (normalized.length % 4 != 0) {
        normalized += '=';
      }

      final decoded = utf8.decode(base64Url.decode(normalized));
      return jsonDecode(decoded) as Map<String, dynamic>;
    } catch (e) {
      return {};
    }
  }
}
