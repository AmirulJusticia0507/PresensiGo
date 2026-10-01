import 'dart:convert';
import 'package:crypto/crypto.dart';
import 'package:uuid/uuid.dart';

class CryptoHelper {
  /// Generate HMAC signature using sorted key-value canonical format
  /// Signs with device UUID as key (not a shared secret)
  static String generateHMAC(Map<String, dynamic> payload, String deviceUuid) {
    // Sort keys alphabetically
    final sortedKeys = payload.keys.toList()..sort();
    
    // Build canonical string: "key1=val1&key2=val2&..."
    final parts = <String>[];
    for (final key in sortedKeys) {
      parts.add('$key=${payload[key]}');
    }
    final canonical = parts.join('&');
    
    // Sign with device UUID as key
    final key = utf8.encode(deviceUuid);
    final bytes = utf8.encode(canonical);
    final hmacSha256 = Hmac(sha256, key);
    final digest = hmacSha256.convert(bytes);
    return digest.toString();
  }

  static String generateDeviceId() {
    return const Uuid().v4();
  }
  
  /// Get current Unix timestamp in seconds
  static int getCurrentTimestamp() {
    return DateTime.now().millisecondsSinceEpoch ~/ 1000;
  }
}
