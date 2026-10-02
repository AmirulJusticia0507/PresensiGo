import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:presensigo_app/core/constants/api_constants.dart';
import 'package:presensigo_app/data/services/profile_service.dart';
import 'package:presensigo_app/data/services/session_manager.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// In-memory [SessionManager] so tests do not depend on the platform channel
/// implemented by flutter_secure_storage.
class FakeSessionManager extends SessionManager {
  FakeSessionManager({this.token});

  String? token;
  bool didLogout = false;

  @override
  Future<String?> getToken() async => token;

  @override
  Future<void> saveToken(String value) async => token = value;

  @override
  Future<void> logout() async {
    didLogout = true;
    token = null;
  }

  @override
  Future<void> handleUnauthorized() => logout();
}

/// Captures the request a [MockClient] handler saw so assertions can verify
/// method, URL, and body.
class _CapturedRequest {
  _CapturedRequest(this.request);

  final http.Request request;

  String get body => request.body;
  String get url => request.url.toString();
  String get method => request.method;

  Map<String, dynamic> get json =>
      jsonDecode(request.body) as Map<String, dynamic>;

  String? header(String name) => request.headers[name];
}

Map<String, dynamic> _profileJson({
  String id = 'user-1',
  String name = 'John Doe',
  String email = 'john@example.com',
  String? phone = '+12025551234',
}) =>
    {
      'id': id,
      'name': name,
      'email': email,
      'phone': phone,
      'emergency_contact_name': null,
      'emergency_contact_phone': null,
      'address': null,
      'profile_picture_url': null,
      'face_enrolled': false,
      'face_enrolled_at': null,
      'created_at': '2026-01-01T00:00:00Z',
      'updated_at': '2026-01-02T00:00:00Z',
    };

/// Builds a [ProfileService] backed by a mock client, recording each request.
({ProfileService service, List<_CapturedRequest> requests, FakeSessionManager session})
    buildService(
  http.Response Function(http.Request request) handler, {
  String? token = 'test-token',
}) {
  final requests = <_CapturedRequest>[];
  final session = FakeSessionManager(token: token);
  final client = MockClient((request) async {
    requests.add(_CapturedRequest(request));
    return handler(request);
  });
  return (
    service: ProfileService.withClient(
      client,
      sessionManagerFactory: () => session,
    ),
    requests: requests,
    session: session,
  );
}

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    final service = ProfileService.withClient(
      MockClient((_) async => http.Response('', 200)),
      sessionManagerFactory: FakeSessionManager.new,
    );
    await service.clearCache();
  });

  group('ProfileService - construction', () {
    test('ProfileService() returns a singleton', () {
      expect(identical(ProfileService(), ProfileService()), isTrue);
    });

    test('ProfileService.withClient creates a distinct instance', () {
      final a = ProfileService.withClient(
        MockClient((_) async => http.Response('', 200)),
        sessionManagerFactory: FakeSessionManager.new,
      );
      final b = ProfileService.withClient(
        MockClient((_) async => http.Response('', 200)),
        sessionManagerFactory: FakeSessionManager.new,
      );
      expect(identical(a, b), isFalse);
    });
  });

  group('getProfile (8.4)', () {
    test('returns the profile on 200 and caches it', () async {
      final harness = buildService((_) => http.Response(
            jsonEncode(_profileJson()),
            200,
            headers: {'content-type': 'application/json'},
          ));

      final result = await harness.service.getProfile();

      expect(result.error, isNull);
      expect(result.profile, isNotNull);
      expect(result.profile!.email, 'john@example.com');
      expect(result.profile!.name, 'John Doe');
      expect(harness.requests.single.method, 'GET');
      expect(harness.requests.single.url,
          '${ApiConstants.baseUrl}${ApiConstants.profile}');
    });

    test('does not repeat the request while the cache is fresh', () async {
      final harness = buildService((_) => http.Response(
            jsonEncode(_profileJson()),
            200,
            headers: {'content-type': 'application/json'},
          ));

      final first = await harness.service.getProfile();
      final second = await harness.service.getProfile();

      expect(first.profile!.email, second.profile!.email);
      expect(harness.requests, hasLength(1),
          reason: 'second call should be served from cache');
    });

    test('forceRefresh bypasses the cache', () async {
      final harness = buildService((_) => http.Response(
            jsonEncode(_profileJson()),
            200,
            headers: {'content-type': 'application/json'},
          ));

      await harness.service.getProfile();
      await harness.service.getProfile(forceRefresh: true);

      expect(harness.requests, hasLength(2));
    });

    test('returns a session-expired error on 401', () async {
      final harness = buildService(
        (_) => http.Response(jsonEncode({'error': 'unauthorized'}), 401),
      );

      final result = await harness.service.getProfile();

      expect(result.profile, isNull);
      expect(result.error, contains('Session expired'));
    });

    test('reports an invalid response format on malformed JSON', () async {
      final harness = buildService(
        (_) => http.Response('not json at all', 200),
      );

      final result = await harness.service.getProfile();

      expect(result.profile, isNull);
      expect(result.error, 'Invalid response format');
    });

    test('surfaces a server error message on 500', () async {
      final harness = buildService(
        (_) => http.Response(jsonEncode({'error': 'internal failure'}), 500),
      );

      final result = await harness.service.getProfile();

      expect(result.profile, isNull);
      expect(result.error, 'internal failure');
    });

    test('returns a network error instead of throwing', () async {
      final client = MockClient((_) async => throw const SocketishError());
      final result = await ProfileService.withClient(client, sessionManagerFactory: FakeSessionManager.new).getProfile();

      expect(result.profile, isNull);
      expect(result.error, contains('Network error'));
    });
  });

  group('updateProfile (8.4)', () {
    test('sends PUT with the edited fields and returns the updated profile',
        () async {
      final harness = buildService((_) => http.Response(
            jsonEncode(_profileJson(name: 'Jane Doe', phone: '+628123456789')),
            200,
            headers: {'content-type': 'application/json'},
          ));

      final result = await harness.service.updateProfile(
        name: 'Jane Doe',
        phone: '+628123456789',
        address: 'Jl. Merdeka 1',
      );

      expect(result.error, isNull);
      expect(result.profile!.name, 'Jane Doe');
      expect(harness.requests.single.method, 'PUT');
      expect(harness.requests.single.json['name'], 'Jane Doe');
      expect(harness.requests.single.json['phone'], '+628123456789');
      expect(harness.requests.single.json['address'], 'Jl. Merdeka 1');
    });

    test('omits optional fields that were not provided', () async {
      final harness = buildService((_) => http.Response(
            jsonEncode(_profileJson()),
            200,
            headers: {'content-type': 'application/json'},
          ));

      await harness.service.updateProfile(name: 'Jane Doe');

      final body = harness.requests.single.json;
      expect(body.containsKey('phone'), isFalse);
      expect(body.containsKey('address'), isFalse);
      expect(body.containsKey('emergency_contact_name'), isFalse);
    });

    test('parses field-level errors on 400', () async {
      final harness = buildService(
        (_) => http.Response(
          jsonEncode({
            'error': 'Validation failed',
            'errors': {'phone': 'Invalid phone format'},
          }),
          400,
        ),
      );

      final result = await harness.service.updateProfile(
        name: 'Jane Doe',
        phone: '12345',
      );

      expect(result.profile, isNull);
      expect(result.error, isNull);
      expect(result.fieldErrors!['phone'], 'Invalid phone format');
    });

    test('maps 409 to a field error on email', () async {
      final harness = buildService(
        (_) => http.Response(jsonEncode({'error': 'Email already registered'}), 409),
      );

      final result = await harness.service.updateProfile(name: 'Jane Doe');

      expect(result.fieldErrors!['email'], 'Email already registered');
    });

    test('returns a session-expired error on 401', () async {
      final harness = buildService(
        (_) => http.Response(jsonEncode({'error': 'unauthorized'}), 401),
      );

      final result = await harness.service.updateProfile(name: 'Jane Doe');

      expect(result.profile, isNull);
      expect(result.error, contains('Session expired'));
    });

    test('returns a network error instead of throwing', () async {
      final client = MockClient((_) async => throw const SocketishError());
      final result =
          await ProfileService.withClient(client, sessionManagerFactory: FakeSessionManager.new).updateProfile(name: 'Jane');

      expect(result.profile, isNull);
      expect(result.error, contains('Network error'));
    });
  });

  group('changePassword (8.4)', () {
    test('returns success on 200 and clears the cached profile', () async {
      final harness = buildService(
        (_) => http.Response(jsonEncode({'message': 'password changed'}), 200),
      );

      final result = await harness.service.changePassword(
        currentPassword: 'OldPass123!',
        newPassword: 'NewPass456!',
        confirmPassword: 'NewPass456!',
      );

      expect(result.success, isTrue);
      expect(result.error, isNull);
      expect(harness.requests.single.method, 'PUT');
      expect(
        harness.requests.single.url,
        '${ApiConstants.baseUrl}${ApiConstants.profilePassword}',
      );
    });

    test('returns a generic error on 400 without revealing the reason', () async {
      final harness = buildService(
        (_) => http.Response(jsonEncode({'error': 'Invalid credentials'}), 400),
      );

      final result = await harness.service.changePassword(
        currentPassword: 'WrongPass123!',
        newPassword: 'NewPass456!',
        confirmPassword: 'NewPass456!',
      );

      expect(result.success, isFalse);
      expect(result.error, isNot(contains('Invalid credentials')));
      expect(result.error, isNot(contains('password must')));
    });

    test('returns a session-expired error on 401', () async {
      final harness = buildService(
        (_) => http.Response(jsonEncode({'error': 'unauthorized'}), 401),
      );

      final result = await harness.service.changePassword(
        currentPassword: 'OldPass123!',
        newPassword: 'NewPass456!',
        confirmPassword: 'NewPass456!',
      );

      expect(result.success, isFalse);
      expect(result.error, contains('Session expired'));
    });

    test('returns a generic error on 500', () async {
      final harness = buildService(
        (_) => http.Response(jsonEncode({'error': 'boom'}), 500),
      );

      final result = await harness.service.changePassword(
        currentPassword: 'OldPass123!',
        newPassword: 'NewPass456!',
        confirmPassword: 'NewPass456!',
      );

      expect(result.success, isFalse);
      expect(result.error, isNotNull);
    });

    test('returns a network error instead of throwing', () async {
      final client = MockClient((_) async => throw const SocketishError());
      final result = await ProfileService.withClient(client, sessionManagerFactory: FakeSessionManager.new).changePassword(
        currentPassword: 'OldPass123!',
        newPassword: 'NewPass456!',
        confirmPassword: 'NewPass456!',
      );

      expect(result.success, isFalse);
      expect(result.error, contains('Network error'));
    });
  });

  group('register (8.4)', () {
    Map<String, dynamic> successBody() => {
          'message': 'registration successful',
          'token': 'jwt-token-123',
          'user': _profileJson(),
        };

    test('posts the registration payload and returns the profile', () async {
      final harness = buildService(
        (_) => http.Response(jsonEncode(successBody()), 201),
      );

      final profile = await harness.service.register(
        email: 'John@Example.com',
        password: 'SecurePass123!',
        confirmPassword: 'SecurePass123!',
        name: ' John Doe ',
        phone: ' +628123456789 ',
        termsAccepted: true,
      );

      expect(profile.email, 'john@example.com');
      expect(harness.requests.single.method, 'POST');
      expect(harness.requests.single.url,
          '${ApiConstants.baseUrl}${ApiConstants.authRegister}');

      final body = harness.requests.single.json;
      expect(body['email'], 'john@example.com', reason: 'email is normalized');
      expect(body['name'], 'John Doe', reason: 'name is trimmed');
      expect(body['phone'], '+628123456789');
      expect(body['terms_accepted'], isTrue);
      expect(body['confirm_password'], 'SecurePass123!');
    });

    test('omits phone when it is empty', () async {
      final harness = buildService(
        (_) => http.Response(jsonEncode(successBody()), 201),
      );

      await harness.service.register(
        email: 'john@example.com',
        password: 'SecurePass123!',
        confirmPassword: 'SecurePass123!',
        name: 'John Doe',
        phone: '   ',
        termsAccepted: true,
      );

      expect(harness.requests.single.json.containsKey('phone'), isFalse);
    });

    test('throws ApiException 409 for a duplicate email', () async {
      final harness = buildService(
        (_) => http.Response(jsonEncode({'error': 'Email already registered'}), 409),
      );

      await expectLater(
        harness.service.register(
          email: 'john@example.com',
          password: 'SecurePass123!',
          confirmPassword: 'SecurePass123!',
          name: 'John Doe',
          termsAccepted: true,
        ),
        throwsA(
          isA<dynamic>()
              .having((e) => e.statusCode, 'statusCode', 409)
              .having((e) => e.message, 'message', 'Email already registered'),
        ),
      );
    });

    test('throws ApiException 400 carrying field-level details', () async {
      final harness = buildService(
        (_) => http.Response(
          jsonEncode({
            'error': 'Validation failed',
            'details': [
              {'field': 'password', 'reason': 'Password must be stronger'},
            ],
          }),
          400,
        ),
      );

      await expectLater(
        harness.service.register(
          email: 'john@example.com',
          password: 'weak',
          confirmPassword: 'weak',
          name: 'John Doe',
          termsAccepted: true,
        ),
        throwsA(
          isA<dynamic>()
              .having((e) => e.statusCode, 'statusCode', 400)
              .having((e) => e.details?.single['field'], 'detail field', 'password'),
        ),
      );
    });

    test('throws when the success body has no token', () async {
      final harness = buildService(
        (_) => http.Response(jsonEncode({'user': _profileJson()}), 201),
      );

      await expectLater(
        harness.service.register(
          email: 'john@example.com',
          password: 'SecurePass123!',
          confirmPassword: 'SecurePass123!',
          name: 'John Doe',
          termsAccepted: true,
        ),
        throwsA(
          isA<dynamic>().having((e) => e.statusCode, 'statusCode', 502),
        ),
      );
    });

    test('throws on a malformed success body', () async {
      final harness = buildService(
        (_) => http.Response('<html>oops</html>', 201),
      );

      await expectLater(
        harness.service.register(
          email: 'john@example.com',
          password: 'SecurePass123!',
          confirmPassword: 'SecurePass123!',
          name: 'John Doe',
          termsAccepted: true,
        ),
        throwsA(
          isA<dynamic>().having((e) => e.statusCode, 'statusCode', 502),
        ),
      );
    });

    test('throws a connection error instead of propagating a raw exception',
        () async {
      final client = MockClient((_) async => throw const SocketishError());

      await expectLater(
        ProfileService.withClient(client, sessionManagerFactory: FakeSessionManager.new).register(
          email: 'john@example.com',
          password: 'SecurePass123!',
          confirmPassword: 'SecurePass123!',
          name: 'John Doe',
          termsAccepted: true,
        ),
        throwsA(
          isA<dynamic>().having((e) => e.statusCode, 'statusCode', 0),
        ),
      );
    });
  });

  group('parseFieldErrors (8.4)', () {
    test('reads a nested errors object', () {
      final parsed = ProfileService.parseFieldErrors(
        jsonEncode({
          'error': 'Validation failed',
          'errors': {'email': 'Invalid email', 'name': 'Too short'},
        }),
      );

      expect(parsed, {'email': 'Invalid email', 'name': 'Too short'});
    });

    test('reads a flat field map and skips the error key', () {
      final parsed = ProfileService.parseFieldErrors(
        jsonEncode({'error': 'Validation failed', 'phone': 'Invalid phone'}),
      );

      expect(parsed, {'phone': 'Invalid phone'});
    });

    test('returns an empty map for a non-JSON body', () {
      expect(ProfileService.parseFieldErrors('not json'), isEmpty);
    });
  });

  group('apiExceptionFromBody (8.4)', () {
    test('maps 409 to a duplicate-email message', () {
      final ex = ProfileService.apiExceptionFromBody(
        409,
        jsonEncode({'error': 'Email already registered'}),
      );

      expect(ex.statusCode, 409);
      expect(ex.message, 'Email already registered');
    });

    test('falls back to a generic message for a non-JSON body', () {
      final ex = ProfileService.apiExceptionFromBody(500, 'oops');

      expect(ex.statusCode, 500);
      expect(ex.message, 'Request failed. Please try again.');
    });

    test('extracts details entries that carry a field', () {
      final ex = ProfileService.apiExceptionFromBody(
        400,
        jsonEncode({
          'error': 'Validation failed',
          'details': [
            {'field': 'email', 'reason': 'Invalid email format'},
            {'reason': 'no field here'},
          ],
        }),
      );

      expect(ex.details, hasLength(1));
      expect(ex.details!.single['field'], 'email');
    });
  });
}

/// Stand-in for a transport failure that is not an [ApiException].
class SocketishError implements Exception {
  const SocketishError();
  @override
  String toString() => 'SocketishError: connection refused';
}
