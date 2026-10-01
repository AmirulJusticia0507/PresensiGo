import 'package:flutter_secure_storage/flutter_secure_storage.dart';

class SecureStorageService {
  static const _tokenKey = 'jwt_token';
  static const _deviceIdKey = 'device_uuid';
  static const _biometricKey = 'biometric_enabled';
  static const _offlineQueueKey = 'offline_queue_key';

  late final FlutterSecureStorage _storage;

  SecureStorageService() {
    _storage = const FlutterSecureStorage(
      aOptions: AndroidOptions(
        keyCipherAlgorithm:
            KeyCipherAlgorithm.RSA_ECB_OAEPwithSHA_256andMGF1Padding,
        storageCipherAlgorithm: StorageCipherAlgorithm.AES_GCM_NoPadding,
      ),
      iOptions: IOSOptions(
        accessibility: KeychainAccessibility.first_unlock_this_device,
      ),
    );
  }

  /// Save JWT token to secure storage
  Future<void> saveToken(String token) async {
    try {
      await _storage.write(key: _tokenKey, value: token);
    } catch (e) {
      throw Exception('Failed to save token: $e');
    }
  }

  /// Retrieve JWT token from secure storage
  Future<String?> getToken() async {
    try {
      return await _storage.read(key: _tokenKey);
    } catch (e) {
      throw Exception('Failed to retrieve token: $e');
    }
  }

  /// Securely delete JWT token from secure storage
  Future<void> deleteToken() async {
    try {
      await _storage.delete(key: _tokenKey);
    } catch (e) {
      throw Exception('Failed to delete token: $e');
    }
  }

  /// Save device UUID to secure storage
  Future<void> saveDeviceId(String deviceId) async {
    try {
      await _storage.write(key: _deviceIdKey, value: deviceId);
    } catch (e) {
      throw Exception('Failed to save device ID: $e');
    }
  }

  /// Retrieve device UUID from secure storage
  Future<String?> getDeviceId() async {
    try {
      return await _storage.read(key: _deviceIdKey);
    } catch (e) {
      throw Exception('Failed to retrieve device ID: $e');
    }
  }

  /// Set biometric enabled flag in secure storage
  Future<void> setBiometricEnabled(bool enabled) async {
    try {
      await _storage.write(key: _biometricKey, value: enabled.toString());
    } catch (e) {
      throw Exception('Failed to set biometric flag: $e');
    }
  }

  /// Check if biometric is enabled
  Future<bool> isBiometricEnabled() async {
    try {
      final val = await _storage.read(key: _biometricKey);
      return val == 'true';
    } catch (e) {
      return false;
    }
  }

  /// Atomically clear all credentials from secure storage
  Future<void> clearAllCredentials() async {
    try {
      await Future.wait([
        _storage.delete(key: _tokenKey),
        _storage.delete(key: _deviceIdKey),
        _storage.delete(key: _biometricKey),
      ]);
    } catch (e) {
      throw Exception('Failed to clear credentials: $e');
    }
  }

  Future<String?> getOfflineQueueKey() => _storage.read(key: _offlineQueueKey);

  Future<void> saveOfflineQueueKey(String value) =>
      _storage.write(key: _offlineQueueKey, value: value);
}
