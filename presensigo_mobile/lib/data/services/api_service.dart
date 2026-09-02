import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';
import '../../core/constants/api_constants.dart';
import '../models/user_model.dart';
import '../models/attendance_model.dart';

class ApiService {
  static String? _token;
  static UserModel? _currentUser;

  static UserModel? get currentUser => _currentUser;

  static Future<String?> getToken() async {
    if (_token != null) return _token;
    final prefs = await SharedPreferences.getInstance();
    _token = prefs.getString('token');
    return _token;
  }

  static Future<void> _saveToken(String token) async {
    _token = token;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('token', token);
  }

  static Future<void> _clearToken() async {
    _token = null;
    _currentUser = null;
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
      body: jsonEncode({
        'name': name,
        'email': email,
        'password': password,
      }),
    );

    final data = jsonDecode(response.body);
    if (response.statusCode == 201) {
      return {'success': true};
    }
    return {'success': false, 'message': data['error'] ?? 'Registration failed'};
  }

  static Future<void> logout() async {
    await _clearToken();
  }

  static Future<AttendanceModel?> getTodayAttendance() async {
    final response = await http.get(
      Uri.parse('${ApiConstants.baseUrl}${ApiConstants.attendanceToday}'),
      headers: await _headers(),
    );

    if (response.statusCode == 200) {
      return AttendanceModel.fromJson(jsonDecode(response.body));
    }
    return null;
  }

  static Future<List<AttendanceModel>> getHistory({int limit = 10, int offset = 0}) async {
    final response = await http.get(
      Uri.parse('${ApiConstants.baseUrl}${ApiConstants.attendanceHistory}?limit=$limit&offset=$offset'),
      headers: await _headers(),
    );

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
    required String hmacSignature,
    String? selfieData,
  }) async {
    final response = await http.post(
      Uri.parse('${ApiConstants.baseUrl}${ApiConstants.attendanceCheckIn}'),
      headers: await _headers(),
      body: jsonEncode({
        'latitude': latitude,
        'longitude': longitude,
        'device_uuid': deviceUuid,
        'hmac_signature': hmacSignature,
        if (selfieData != null) 'selfie_data': selfieData,
      }),
    );

    final data = jsonDecode(response.body);
    if (response.statusCode == 200) {
      return {'success': true, 'attendance': AttendanceModel.fromJson(data)};
    }
    return {'success': false, 'message': data['error'] ?? 'Check-in failed'};
  }

  static Future<Map<String, dynamic>> checkOut({
    required double latitude,
    required double longitude,
    required String deviceUuid,
    required String hmacSignature,
  }) async {
    final response = await http.post(
      Uri.parse('${ApiConstants.baseUrl}${ApiConstants.attendanceCheckOut}'),
      headers: await _headers(),
      body: jsonEncode({
        'latitude': latitude,
        'longitude': longitude,
        'device_uuid': deviceUuid,
        'hmac_signature': hmacSignature,
      }),
    );

    final data = jsonDecode(response.body);
    if (response.statusCode == 200) {
      return {'success': true, 'attendance': AttendanceModel.fromJson(data)};
    }
    return {'success': false, 'message': data['error'] ?? 'Check-out failed'};
  }

  static Future<List<LocationModel>> getLocations() async {
    final response = await http.get(
      Uri.parse('${ApiConstants.baseUrl}${ApiConstants.locations}'),
      headers: await _headers(),
    );

    if (response.statusCode == 200) {
      final List data = jsonDecode(response.body);
      return data.map((e) => LocationModel.fromJson(e)).toList();
    }
    return [];
  }
}
