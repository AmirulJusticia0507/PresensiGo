import 'dart:convert';
import 'package:crypto/crypto.dart';
import 'package:uuid/uuid.dart';

class CryptoHelper {
  static String generateHMAC(Map<String, dynamic> payload, String secret) {
    final bytes = utf8.encode(json.encode(payload));
    final key = utf8.encode(secret);
    final hmacSha256 = Hmac(sha256, key);
    final digest = hmacSha256.convert(bytes);
    return digest.toString();
  }

  static String generateDeviceId() {
    return const Uuid().v4();
  }
}
