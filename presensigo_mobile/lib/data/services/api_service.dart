import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

import '../../core/constants/api_constants.dart';
import '../models/user_model.dart';
import '../models/attendance_model.dart';
import 'secure_storage_service.dart';
import 'session_manager.dart';

class ApiService {
  static String? _token;
  static UserModel? _currentUser;
  static final SecureStorageService _secureStorage = SecureStorageService();
  static final SessionManager _sessionManager = SessionManager();

  static UserModel? get currentUser => _currentUser;

  static Future<String?> getToken() async {
    if (_token != null) return _token;
    // First try to get from secure storage
    _token = await _secureStorage.getToken();
    if (_token != null) return _token;

    // Fallback to SharedPreferences for migration
    final prefs = await SharedPreferences.getInstance();
    _token = prefs.getString('token');
    return _token;
  }

  static Future<void> _saveToken(String token) async {
    _token = token;
    // Save to secure storage instead of SharedPreferences
    await _secureStorage.saveToken(token);
  }

  static Future<void> _clearToken() async {
    _token = null;
    _currentUser = null;
    // Clear from secure storage
    await _secureStorage.clearAllCredentials();
    // Also clear from SharedPreferences for migration safety
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('token');
  }

  static Future<Map<String, String>> _headers() async {
    final token = await getToken();
    return {
      'Content-Type': 'application/json',
      if (token != null) 'Authorization': 'Bearer $token',
    };
  }

  /// Get a valid token or null if expired/missing
  /// Handles automatic logout on session expiry
  static Future<String?> getValidToken() async {
    if (await _sessionManager.isSessionValid()) {
      return await _secureStorage.getToken();
    }
    // Token invalid or expired - clear it and notify
    await _sessionManager.handleUnauthorized();
    await _clearToken();
    return null;
  }

  static Future<Map<String, dynamic>> login({
    required String email,
    required String password,
    required String deviceUuid,
  }) async {
    final response = await http.post(
      Uri.parse('${ApiConstants.baseUrl}${ApiConstants.authLogin}'),
      headers: await _headers(),
      body: jsonEncode({
        'email': email,
        'password': password,
        'device_uuid': deviceUuid,
      }),
    );

    final data = jsonDecode(response.body);
    if (response.statusCode == 200) {
      await _saveToken(data['token']);
      _currentUser = UserModel.fromJson(data['user']);
      return {'success': true, 'user': _currentUser};
    }
    return {'success': false, 'message': data['error'] ?? 'Login failed'};
  }

  static Future<Map<String, dynamic>> register({
    required String name,
    required String email,
    required String password,
  }) async {
    final response = await http.post(
      Uri.parse('${ApiConstants.baseUrl}${ApiConstants.authRegister}'),
      headers: await _headers(),
      body: jsonEncode({'name': name, 'email': email, 'password': password}),
    );

    final data = jsonDecode(response.body);
    if (response.statusCode == 201) {
      return {'success': true};
    }
    return {
      'success': false,
      'message': data['error'] ?? 'Registration failed',
    };
  }

  static Future<void> logout() async {
    // Optional: Notify backend of logout (for audit/revocation)
    try {
      await http.post(
        Uri.parse('${ApiConstants.baseUrl}/api/auth/logout'),
        headers: await _headers(),
      );
    } catch (e) {
      // Ignore errors; local logout always succeeds
    }

    // Clear all credentials and session state
    await _sessionManager.logout();
    await _clearToken();
  }

  /// Handle 401 responses and return false if session expired
  static Future<bool> _handleResponse(http.Response response) async {
    if (response.statusCode == 401) {
      // Token expired or invalid
      await _sessionManager.handleUnauthorized();
      await _clearToken();
      return false;
    }
    return true;
  }

  static Future<AttendanceModel?> getTodayAttendance() async {
    final response = await http.get(
      Uri.parse('${ApiConstants.baseUrl}${ApiConstants.attendanceToday}'),
      headers: await _headers(),
    );

    if (!await _handleResponse(response)) return null;

    if (response.statusCode == 200) {
      return AttendanceModel.fromJson(jsonDecode(response.body));
    }
    return null;
  }

  static Future<List<AttendanceModel>> getHistory({
    int limit = 10,
    int offset = 0,
  }) async {
    final response = await http.get(
      Uri.parse(
        '${ApiConstants.baseUrl}${ApiConstants.attendanceHistory}?limit=$limit&offset=$offset',
      ),
      headers: await _headers(),
    );

    if (!await _handleResponse(response)) return [];

    if (response.statusCode == 200) {
      final List data = jsonDecode(response.body);
      return data.map((e) => AttendanceModel.fromJson(e)).toList();
    }
    return [];
  }

  static Future<Map<String, dynamic>> checkIn({
    required double latitude,
    required double longitude,
    required String deviceUuid,
    required int timestamp,
    required String hmacSignature,
    required String selfieData,
    required String livenessChallenge,
    required String livenessToken,
    required String idempotencyKey,
  }) async {
    final uri = Uri.parse(
      '${ApiConstants.baseUrl}${ApiConstants.attendanceCheckIn}',
    );
    final body = jsonEncode({
      'latitude': latitude,
      'longitude': longitude,
      'device_uuid': deviceUuid,
      'timestamp': timestamp,
      'hmac_signature': hmacSignature,
      'selfie_data': selfieData,
      'liveness_challenge': livenessChallenge,
      'liveness_token': livenessToken,
      'idempotency_key': idempotencyKey,
    });

    http.Response? response;
    Object? lastError;
    for (var attempt = 0; attempt < 3; attempt++) {
      try {
        response = await http
            .post(uri, headers: await _headers(), body: body)
            .timeout(const Duration(seconds: 30));
        if (response.statusCode < 500) break;
      } catch (error) {
        lastError = error;
      }

      if (attempt < 2) {
        await Future<void>.delayed(Duration(seconds: 1 << attempt));
      }
    }

    if (response == null) {
      return {
        'success': false,
        'retryable': true,
        'message': 'Unable to upload selfie. Please try again. ($lastError)',
      };
    }

    if (!await _handleResponse(response)) {
      return {
        'success': false,
        'message': 'Session expired, please log in again',
      };
    }

    final data = jsonDecode(response.body) as Map<String, dynamic>;
    if (response.statusCode == 200) {
      return {'success': true, 'attendance': AttendanceModel.fromJson(data)};
    }
    return {
      'success': false,
      'retryable': response.statusCode >= 500,
      'message': data['error'] ?? 'Check-in failed',
    };
  }

  static Future<Map<String, dynamic>> getFaceChallenge() async {
    final response = await http.post(
      Uri.parse('${ApiConstants.baseUrl}${ApiConstants.faceChallenge}'),
      headers: await _headers(),
    );
    if (!await _handleResponse(response)) {
      return {'success': false, 'message': 'Session expired'};
    }
    final data = jsonDecode(response.body) as Map<String, dynamic>;
    if (response.statusCode == 200) return {'success': true, ...data};
    return {
      'success': false,
      'message': data['error'] ?? 'Unable to create liveness challenge',
    };
  }

  static Future<Map<String, dynamic>> enrollFace(List<String> selfies) async {
    try {
      final response = await http
          .post(
            Uri.parse('${ApiConstants.baseUrl}${ApiConstants.faceEnrollment}'),
            headers: await _headers(),
            body: jsonEncode({'selfies': selfies}),
          )
          .timeout(const Duration(seconds: 60));
      if (!await _handleResponse(response)) {
        return {'success': false, 'message': 'Session expired'};
      }
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      if (response.statusCode == 200) return {'success': true, ...data};
      return {
        'success': false,
        'message': data['error'] ?? 'Face enrollment failed',
      };
    } catch (_) {
      return {
        'success': false,
        'message': 'Face service is unavailable. Please try again.',
      };
    }
  }

  static Future<Map<String, dynamic>> checkOut({
    required double latitude,
    required double longitude,
    required String deviceUuid,
    required int timestamp,
    required String hmacSignature,
    required String idempotencyKey,
  }) async {
    http.Response response;
    try {
      response = await http
          .post(
            Uri.parse(
              '${ApiConstants.baseUrl}${ApiConstants.attendanceCheckOut}',
            ),
            headers: await _headers(),
            body: jsonEncode({
              'latitude': latitude,
              'longitude': longitude,
              'device_uuid': deviceUuid,
              'timestamp': timestamp,
              'hmac_signature': hmacSignature,
              'idempotency_key': idempotencyKey,
            }),
          )
          .timeout(const Duration(seconds: 15));
    } catch (_) {
      return {
        'success': false,
        'retryable': true,
        'message': 'Network unavailable. Attendance can be queued.',
      };
    }

    if (!await _handleResponse(response)) {
      return {
        'success': false,
        'message': 'Session expired, please log in again',
      };
    }

    final data = jsonDecode(response.body);
    if (response.statusCode == 200) {
      return {'success': true, 'attendance': AttendanceModel.fromJson(data)};
    }
    return {
      'success': false,
      'retryable': response.statusCode >= 500,
      'message': data['error'] ?? 'Check-out failed',
    };
  }

  static Future<Map<String, dynamic>> syncAttendance(
    List<Map<String, dynamic>> items,
  ) async {
    try {
      final actions = items
          .map(
            (item) => {
              'idempotency_key': item['idempotency_key'],
              'action_type': item['action_type'],
              'payload': item['payload'],
            },
          )
          .toList();
      final response = await http
          .post(
            Uri.parse('${ApiConstants.baseUrl}${ApiConstants.attendanceSync}'),
            headers: await _headers(),
            body: jsonEncode({'actions': actions}),
          )
          .timeout(const Duration(seconds: 60));
      if (!await _handleResponse(response)) {
        return {'success': false, 'message': 'Session expired'};
      }
      if (response.statusCode == 200) {
        final data = jsonDecode(response.body) as Map<String, dynamic>;
        return {'success': true, ...data};
      }
      return {
        'success': false,
        'message': 'Sync failed (${response.statusCode})',
      };
    } catch (_) {
      return {'success': false, 'message': 'Network unavailable'};
    }
  }

  static Future<Map<String, dynamic>> getSyncStatus() async {
    final response = await http.get(
      Uri.parse('${ApiConstants.baseUrl}${ApiConstants.attendanceSync}/status'),
      headers: await _headers(),
    );

    if (!await _handleResponse(response)) {
      return {'success': false, 'message': 'Session expired'};
    }

    if (response.statusCode == 200) {
      final data = jsonDecode(response.body) as Map<String, dynamic>;
      return {'success': true, ...data};
    }
    return {'success': false, 'message': 'Failed to get sync status'};
  }

  /// Best-effort security telemetry. A reporting failure must never turn a
  /// locally rejected attendance attempt into an accepted one.
  static Future<void> reportLocationAttempt({
    required String reason,
    required String platform,
    required double latitude,
    required double longitude,
    required double accuracy,
  }) async {
    try {
      await http
          .post(
            Uri.parse(
              '${ApiConstants.baseUrl}${ApiConstants.locationAttempts}',
            ),
            headers: await _headers(),
            body: jsonEncode({
              'reason': reason,
              'platform': platform,
              'latitude': latitude,
              'longitude': longitude,
              'accuracy': accuracy,
            }),
          )
          .timeout(const Duration(seconds: 5));
    } catch (_) {
      // Deliberately ignored: the suspicious action is already blocked.
    }
  }

  static Future<List<LocationModel>> getLocations() async {
    final response = await http.get(
      Uri.parse('${ApiConstants.baseUrl}${ApiConstants.locations}'),
      headers: await _headers(),
    );

    if (!await _handleResponse(response)) return [];

    if (response.statusCode == 200) {
      final List data = jsonDecode(response.body);
      return data.map((e) => LocationModel.fromJson(e)).toList();
    }
    return [];
  }
}
