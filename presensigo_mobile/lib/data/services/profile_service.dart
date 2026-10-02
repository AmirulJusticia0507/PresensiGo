import 'dart:async';
import 'dart:convert';

import 'package:http/http.dart' as http;
import 'package:shared_preferences/shared_preferences.dart';

import '../../core/constants/api_constants.dart';
import '../../core/errors/api_exception.dart';
import '../models/profile_model.dart';
import 'session_manager.dart';

/// ProfileService handles all profile-related API calls
class ProfileService {
  static final ProfileService _instance = ProfileService._internal();
  final http.Client _client;
  final SessionManager Function() _sessionManagerFactory;

  // Cache key and duration
  static const String _cacheKey = 'profile_cache';
  static const Duration _cacheDuration = Duration(hours: 1);
  static const Duration _timeout = Duration(seconds: 15);

  factory ProfileService() {
    return _instance;
  }

  ProfileService._internal()
      : _client = http.Client(),
        _sessionManagerFactory = _defaultSessionManagerFactory;

  /// Constructor used by tests to inject a mock HTTP client.
  ProfileService.withClient(
    http.Client client, {
    SessionManager Function()? sessionManagerFactory,
  })  : _client = client,
        _sessionManagerFactory =
            sessionManagerFactory ?? _defaultSessionManagerFactory;

  static SessionManager _defaultSessionManagerFactory() => SessionManager();

  SessionManager get _session => _sessionManagerFactory();

  /// Get cached profile or null if expired
  Future<ProfileModel?> _getCachedProfile() async {
    final prefs = await SharedPreferences.getInstance();
    final cachedJson = prefs.getString(_cacheKey);
    if (cachedJson == null) return null;

    final cacheTimestamp =
        prefs.getInt('${_cacheKey}_timestamp') ?? 0;
    final now = DateTime.now().millisecondsSinceEpoch;
    final age = Duration(milliseconds: now - cacheTimestamp);

    // If cache is expired, return null
    if (age > _cacheDuration) {
      return null;
    }

    try {
      final json = jsonDecode(cachedJson) as Map<String, dynamic>;
      return ProfileModel.fromJson(json);
    } catch (_) {
      return null;
    }
  }

  /// Cache profile locally
  Future<void> cacheProfile(ProfileModel profile) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_cacheKey, jsonEncode(profile.toJson()));
    await prefs.setInt(
      '${_cacheKey}_timestamp',
      DateTime.now().millisecondsSinceEpoch,
    );
  }

  /// Clear cached profile
  Future<void> clearCache() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_cacheKey);
    await prefs.remove('${_cacheKey}_timestamp');
  }

  /// Prepare headers with JWT token
  Future<Map<String, String>> _headers() async {
    final token = await _session.getToken();
    return {
      'Content-Type': 'application/json',
      if (token != null) 'Authorization': 'Bearer $token',
    };
  }

  /// GET /api/profile - Load profile with caching (6.4)
  /// Returns cached profile if available, otherwise fetches from server
  /// Handles 401 unauthorized errors
  Future<({ProfileModel? profile, String? error})> getProfile({
    bool forceRefresh = false,
  }) async {
    try {
      // Try cache first if not forcing refresh
      if (!forceRefresh) {
        final cached = await _getCachedProfile();
        if (cached != null) {
          return (profile: cached, error: null);
        }
      }

      // Fetch from server
      final response = await _client
          .get(
            Uri.parse('${ApiConstants.baseUrl}${ApiConstants.profile}'),
            headers: await _headers(),
          )
          .timeout(_timeout);

      // Handle 401 unauthorized
      if (response.statusCode == 401) {
        await _session.handleUnauthorized();
        return (
          profile: null,
          error: 'Session expired, please log in again'
        );
      }

      // Handle 200 success
      if (response.statusCode == 200) {
        try {
          final json = jsonDecode(response.body) as Map<String, dynamic>;
          final profile = ProfileModel.fromJson(json);
          await cacheProfile(profile);
          return (profile: profile, error: null);
        } catch (_) {
          return (profile: null, error: 'Invalid response format');
        }
      }

      // Handle other errors
      final errorMsg =
          _parseErrorMessage(response.body) ?? 'Failed to load profile';
      return (profile: null, error: errorMsg);
    } on TimeoutException {
      return (profile: null, error: 'Request timed out. Please try again.');
    } catch (_) {
      return (profile: null, error: 'Network error. Check your connection.');
    }
  }

  /// PUT /api/profile - Update profile with selected fields (6.5)
  /// Only sends fields that were changed
  /// Handles 400 validation errors with field details
  /// Handles 401 unauthorized
  /// Handles 409 duplicate email
  Future<({ProfileModel? profile, Map<String, String>? fieldErrors, String? error})>
      updateProfile({
    required String name,
    String? phone,
    String? emergencyContactName,
    String? emergencyContactPhone,
    String? address,
  }) async {
    try {
      // Build request with only provided fields
      final body = <String, dynamic>{
        'name': name,
        'phone': ?phone,
        'emergency_contact_name': ?emergencyContactName,
        'emergency_contact_phone': ?emergencyContactPhone,
        'address': ?address,
      };

      final response = await _client
          .put(
            Uri.parse('${ApiConstants.baseUrl}${ApiConstants.profile}'),
            headers: await _headers(),
            body: jsonEncode(body),
          )
          .timeout(_timeout);

      // Handle 401 unauthorized
      if (response.statusCode == 401) {
        await _session.handleUnauthorized();
        return (
          profile: null,
          fieldErrors: null,
          error: 'Session expired, please log in again'
        );
      }

      // Handle 400 validation errors with field details
      if (response.statusCode == 400) {
        final fieldErrors = parseFieldErrors(response.body);
        if (fieldErrors.isNotEmpty) {
          return (profile: null, fieldErrors: fieldErrors, error: null);
        }
        final errorMsg =
            _parseErrorMessage(response.body) ?? 'Validation failed';
        return (profile: null, fieldErrors: null, error: errorMsg);
      }

      // Handle 409 duplicate email
      if (response.statusCode == 409) {
        final fieldErrors = parseFieldErrors(response.body);
        if (fieldErrors.isEmpty) {
          fieldErrors['email'] = 'Email already registered';
        }
        return (profile: null, fieldErrors: fieldErrors, error: null);
      }

      // Handle 200 success
      if (response.statusCode == 200) {
        try {
          final json = jsonDecode(response.body) as Map<String, dynamic>;
          final profile = ProfileModel.fromJson(json);
          await cacheProfile(profile);
          return (profile: profile, fieldErrors: null, error: null);
        } catch (_) {
          return (
            profile: null,
            fieldErrors: null,
            error: 'Invalid response format'
          );
        }
      }

      // Handle other errors
      final errorMsg =
          _parseErrorMessage(response.body) ?? 'Failed to update profile';
      return (profile: null, fieldErrors: null, error: errorMsg);
    } on TimeoutException {
      return (
        profile: null,
        fieldErrors: null,
        error: 'Request timed out. Please try again.'
      );
    } catch (_) {
      return (
        profile: null,
        fieldErrors: null,
        error: 'Network error. Check your connection.'
      );
    }
  }

  /// PUT /api/profile/password - Change password (6.6)
  /// Verifies current password before accepting new
  /// Returns generic error on failure (don't reveal if wrong password vs weak password)
  /// Handles 401 unauthorized
  Future<({bool success, String? error})> changePassword({
    required String currentPassword,
    required String newPassword,
    required String confirmPassword,
  }) async {
    try {
      final body = jsonEncode({
        'current_password': currentPassword,
        'new_password': newPassword,
        'confirm_password': confirmPassword,
      });

      final response = await _client
          .put(
            Uri.parse('${ApiConstants.baseUrl}${ApiConstants.profilePassword}'),
            headers: await _headers(),
            body: body,
          )
          .timeout(_timeout);

      // Handle 401 unauthorized
      if (response.statusCode == 401) {
        await _session.handleUnauthorized();
        return (success: false, error: 'Session expired, please log in again');
      }

      // Handle 200 success
      if (response.statusCode == 200) {
        // Clear cache to force refresh on next profile fetch
        await clearCache();
        return (success: true, error: null);
      }

      // Handle 400 bad request - generic error for security
      if (response.statusCode == 400) {
        // Don't reveal specific reason (security best practice)
        return (
          success: false,
          error: 'Password change failed. Please check your current password.'
        );
      }

      // Handle other errors - generic message
      return (success: false, error: 'Password change failed. Please try again.');
    } on TimeoutException {
      return (success: false, error: 'Request timed out. Please try again.');
    } catch (_) {
      return (
        success: false,
        error: 'Network error. Check your connection.'
      );
    }
  }

  /// POST /api/auth/register - Create an account (5.10)
  ///
  /// On success the returned JWT is persisted in secure storage so the user is
  /// signed in immediately. Throws [ApiException] for validation/conflict
  /// responses and for transport failures, so the caller can render
  /// field-level errors for validation problems.
  Future<ProfileModel> register({
    required String email,
    required String password,
    required String confirmPassword,
    required String name,
    String? phone,
    required bool termsAccepted,
  }) async {
    late final http.Response response;
    try {
      response = await _client
          .post(
            Uri.parse('${ApiConstants.baseUrl}${ApiConstants.authRegister}'),
            headers: const {'Content-Type': 'application/json'},
            body: jsonEncode({
              'email': email.trim().toLowerCase(),
              'password': password,
              'confirm_password': confirmPassword,
              'name': name.trim(),
              if (phone != null && phone.trim().isNotEmpty) 'phone': phone.trim(),
              'terms_accepted': termsAccepted,
            }),
          )
          .timeout(_timeout);
    } on TimeoutException {
      throw ApiException(
        message: 'Registration timed out. Please try again.',
        statusCode: 408,
      );
    } catch (_) {
      throw ApiException(
        message: 'Unable to reach the server. Check your connection.',
        statusCode: 0,
      );
    }

    if (response.statusCode == 201) {
      final Map<String, dynamic> json;
      try {
        json = jsonDecode(response.body) as Map<String, dynamic>;
      } catch (_) {
        throw const ApiException(
          message: 'Invalid response format',
          statusCode: 502,
        );
      }

      final token = json['token'] as String?;
      if (token == null || token.isEmpty) {
        throw const ApiException(
          message: 'Registration response did not include a session token',
          statusCode: 502,
        );
      }

      await _session.saveToken(token);

      final userJson = (json['user'] ?? json) as Map<String, dynamic>;
      final profile = ProfileModel.fromJson(userJson);
      await cacheProfile(profile);

      return profile;
    }

    throw apiExceptionFromBody(response.statusCode, response.body);
  }

  /// Build an [ApiException] from a backend error body.
  static ApiException apiExceptionFromBody(int statusCode, String body) {
    var message = 'Request failed. Please try again.';
    List<Map<String, String?>>? details;

    try {
      final json = jsonDecode(body) as Map<String, dynamic>;

      final error = json['error'];
      if (error is String && error.isNotEmpty) {
        message = error;
      }

      final rawDetails = json['details'];
      if (rawDetails is List) {
        final parsed = rawDetails
            .whereType<Map<String, dynamic>>()
            .map((detail) => {
                  'field': detail['field'] as String?,
                  'reason': detail['reason'] as String?,
                })
            .where((detail) => detail['field'] != null)
            .toList();
        details = parsed.isEmpty ? null : parsed;
      }
    } catch (_) {
      // Non-JSON or unexpected body: keep the generic message.
    }

    if (statusCode == 409) {
      message = 'Email already registered';
      details = null;
    }

    return ApiException(
      message: message,
      statusCode: statusCode,
      details: details,
    );
  }

  /// Parse error message from API response
  static String? _parseErrorMessage(String body) {
    try {
      final json = jsonDecode(body) as Map<String, dynamic>;
      return json['error'] as String?;
    } catch (_) {
      return null;
    }
  }

  /// Parse field-level errors from API response
  /// Expected format: {"errors": {"field_name": "error message"}} or {"field_name": "error message"}
  static Map<String, String> parseFieldErrors(String body) {
    final fieldErrors = <String, String>{};
    try {
      final json = jsonDecode(body) as Map<String, dynamic>;

      // Try errors object first
      if (json.containsKey('errors')) {
        final errors = json['errors'];
        if (errors is Map<String, dynamic>) {
          errors.forEach((key, value) {
            if (value is String) {
              fieldErrors[key] = value;
            }
          });
        }
      } else {
        // Try flat structure
        json.forEach((key, value) {
          if (key != 'error' && value is String) {
            fieldErrors[key] = value;
          }
        });
      }
    } catch (_) {
      // Parse error, return empty map
    }
    return fieldErrors;
  }
}
