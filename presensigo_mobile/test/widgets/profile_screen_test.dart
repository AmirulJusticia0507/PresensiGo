import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:presensigo_app/core/constants/api_constants.dart';
import 'package:presensigo_app/data/services/profile_service.dart';
import 'package:presensigo_app/data/services/session_manager.dart';
import 'package:presensigo_app/features/profile/screens/profile_screen.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// SessionManager stand-in so tests never touch the secure-storage channel.
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

Map<String, dynamic> _profileJson({
  String id = 'user-1',
  String name = 'John Doe',
  String email = 'john@example.com',
  String? phone = '+12025551234',
  String? emergencyContactName = 'Jane Doe',
  String? emergencyContactPhone = '+12025559999',
  String? address = '123 Main Street',
}) =>
    {
      'id': id,
      'name': name,
      'email': email,
      'phone': phone,
      'emergency_contact_name': emergencyContactName,
      'emergency_contact_phone': emergencyContactPhone,
      'address': address,
      'profile_picture_url': null,
      'face_enrolled': false,
      'face_enrolled_at': null,
      'created_at': '2026-01-01T00:00:00Z',
      'updated_at': '2026-01-02T00:00:00Z',
    };

/// Routes requests by `"<METHOD> <path suffix>"`, e.g. `"PUT /profile/password"`.
///
/// Lookup is done from the longest suffix to the shortest so that
/// "/profile/password" wins over "/profile".
({ProfileService service, List<http.Request> requests}) buildUnderTest({
  Map<String, http.Response Function(http.Request request)>? routes,
}) {
  final requests = <http.Request>[];
  final client = MockClient((request) async {
    requests.add(request);

    final table = routes ?? const <String, http.Response Function(http.Request)>{};
    final path = request.url.path;
    for (final suffix in ['/profile/password', '/profile']) {
      if (path.endsWith(suffix)) {
        final handler = table['${request.method} $suffix'];
        if (handler != null) {
          return handler(request);
        }
      }
    }

    if (request.method == 'GET') {
      return http.Response(
        jsonEncode(_profileJson()),
        200,
        headers: {'content-type': 'application/json'},
      );
    }
    return http.Response(jsonEncode(_profileJson()), 200);
  });

  return (
    service: ProfileService.withClient(
      client,
      sessionManagerFactory: () => FakeSessionManager(token: 'token'),
    ),
    requests: requests,
  );
}

Widget wrap(ProfileService service) =>
    MaterialApp(home: ProfileScreen(service: service));

/// Pumps a bounded number of frames.
///
/// [WidgetTester.pumpAndSettle] cannot be used while the screen shows a
/// CircularProgressIndicator because that widget animates forever.
Future<void> settleBriefly(WidgetTester tester, {int frames = 12}) async {
  for (var i = 0; i < frames; i++) {
    await tester.pump(const Duration(milliseconds: 100));
  }
}

Future<void> tapVisible(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pump();
  await tester.tap(finder);
  await tester.pump();
}

void main() {
  setUp(() async {
    SharedPreferences.setMockInitialValues({});
    final service = ProfileService.withClient(
      MockClient((_) async => http.Response(jsonEncode(_profileJson()), 200)),
      sessionManagerFactory: FakeSessionManager.new,
    );
    await service.clearCache();
  });

  group('ProfileScreen (8.3) - loading and read-only view', () {
    testWidgets('shows a loading indicator while the profile is loading',
        (tester) async {
      final completer = Completer<http.Response>();
      final client = MockClient((_) => completer.future);
      final service = ProfileService.withClient(
        client,
        sessionManagerFactory: FakeSessionManager.new,
      );

      await tester.pumpWidget(wrap(service));
      await tester.pump();

      expect(find.byKey(ProfileScreen.loadingIndicatorKey), findsOneWidget);

      completer.complete(
        http.Response(jsonEncode(_profileJson()), 200),
      );
      await settleBriefly(tester);
    });

    testWidgets('renders an AppBar and the profile values', (tester) async {
      final harness = buildUnderTest();
      await tester.pumpWidget(wrap(harness.service));
      await settleBriefly(tester);

      expect(find.byType(AppBar), findsOneWidget);
      expect(find.text('Profile'), findsOneWidget);
      expect(find.text('John Doe'), findsWidgets);
      expect(find.text('john@example.com'), findsWidgets);
      expect(find.text('+12025551234'), findsOneWidget);
      expect(find.text('Jane Doe'), findsOneWidget);
      expect(find.text('+12025559999'), findsOneWidget);
      expect(find.text('123 Main Street'), findsOneWidget);
    });

    testWidgets('wraps the body in a RefreshIndicator for pull-to-refresh',
        (tester) async {
      final harness = buildUnderTest();
      await tester.pumpWidget(wrap(harness.service));
      await settleBriefly(tester);

      expect(find.byType(RefreshIndicator), findsOneWidget);
    });

    testWidgets('refresh re-requests the profile with forceRefresh',
        (tester) async {
      final harness = buildUnderTest();
      await tester.pumpWidget(wrap(harness.service));
      await settleBriefly(tester);

      expect(harness.requests.where((r) => r.method == 'GET'), hasLength(1));

      await tester.drag(
        find.byType(SingleChildScrollView).first,
        const Offset(0, 300),
      );
      await settleBriefly(tester);

      expect(
        harness.requests.where((r) => r.method == 'GET').length,
        greaterThan(1),
        reason: 'pull-to-refresh should hit the API again',
      );
    });

    testWidgets('does not render a Form in read-only mode', (tester) async {
      final harness = buildUnderTest();
      await tester.pumpWidget(wrap(harness.service));
      await settleBriefly(tester);

      expect(find.byType(Form), findsNothing);
    });
  });

  group('ProfileScreen (8.3) - edit mode', () {
    testWidgets('edit button switches to a form with populated values',
        (tester) async {
      final harness = buildUnderTest();
      await tester.pumpWidget(wrap(harness.service));
      await settleBriefly(tester);

      await tapVisible(tester, find.byKey(ProfileScreen.editButtonKey));
      await settleBriefly(tester);

      expect(find.byType(Form), findsOneWidget);
      expect(
        tester
            .widget<TextFormField>(find.byKey(ProfileScreen.nameFieldKey))
            .controller
            ?.text,
        'John Doe',
      );
      expect(
        tester
            .widget<TextFormField>(find.byKey(ProfileScreen.phoneFieldKey))
            .controller
            ?.text,
        '+12025551234',
      );
    });

    testWidgets('the edit button is replaced by a cancel affordance',
        (tester) async {
      final harness = buildUnderTest();
      await tester.pumpWidget(wrap(harness.service));
      await settleBriefly(tester);

      await tapVisible(tester, find.byKey(ProfileScreen.editButtonKey));
      await settleBriefly(tester);

      expect(find.byKey(ProfileScreen.editButtonKey), findsNothing);
      expect(find.byKey(ProfileScreen.cancelButtonKey), findsOneWidget);
      expect(find.byKey(ProfileScreen.saveButtonKey), findsOneWidget);
    });

    testWidgets('save submits the edited fields and returns to read-only',
        (tester) async {
      final harness = buildUnderTest(
        routes: {
          'PUT /profile': (_) => http.Response(
                jsonEncode(_profileJson(name: 'Jane Doe')),
                200,
                headers: {'content-type': 'application/json'},
              ),
        },
      );
      await tester.pumpWidget(wrap(harness.service));
      await settleBriefly(tester);

      await tapVisible(tester, find.byKey(ProfileScreen.editButtonKey));
      await settleBriefly(tester);

      await tester.enterText(
        find.byKey(ProfileScreen.nameFieldKey),
        'Jane Doe',
      );
      await settleBriefly(tester);

      await tapVisible(tester, find.byKey(ProfileScreen.saveButtonKey));
      await settleBriefly(tester);

      final put = harness.requests.firstWhere((r) => r.method == 'PUT');
      expect(put.url.path, endsWith(ApiConstants.profile));
      final body = jsonDecode(put.body) as Map<String, dynamic>;
      expect(body['name'], 'Jane Doe');
      expect(body['phone'], '+12025551234');

      expect(find.byKey(ProfileScreen.editButtonKey), findsOneWidget);
      expect(find.text('Jane Doe'), findsWidgets);
    });

    testWidgets('cancel discards edits without calling the API', (tester) async {
      final harness = buildUnderTest();
      await tester.pumpWidget(wrap(harness.service));
      await settleBriefly(tester);

      await tapVisible(tester, find.byKey(ProfileScreen.editButtonKey));
      await settleBriefly(tester);

      await tester.enterText(
        find.byKey(ProfileScreen.nameFieldKey),
        'Discarded Name',
      );
      await settleBriefly(tester);

      await tapVisible(tester, find.byKey(ProfileScreen.cancelButtonKey));
      await settleBriefly(tester);

      expect(harness.requests.where((r) => r.method == 'PUT'), isEmpty);
      expect(find.byKey(ProfileScreen.editButtonKey), findsOneWidget);
      expect(find.text('Discarded Name'), findsNothing);
    });

    testWidgets('the email field is read-only in edit mode', (tester) async {
      final harness = buildUnderTest();
      await tester.pumpWidget(wrap(harness.service));
      await settleBriefly(tester);

      await tapVisible(tester, find.byKey(ProfileScreen.editButtonKey));
      await settleBriefly(tester);

      final email =
          tester.widget<TextField>(find.byKey(ProfileScreen.emailFieldKey));
      expect(email.enabled, isFalse);
      expect(email.controller?.text, 'john@example.com');
    });
  });

  group('ProfileScreen (8.3) - edit mode validation', () {
    Future<void> enterEditMode(WidgetTester tester, ProfileService service) async {
      await tester.pumpWidget(wrap(service));
      await settleBriefly(tester);
      await tapVisible(tester, find.byKey(ProfileScreen.editButtonKey));
      await settleBriefly(tester);
    }

    testWidgets('rejects a name shorter than two characters', (tester) async {
      final harness = buildUnderTest();
      await enterEditMode(tester, harness.service);

      await tester.enterText(find.byKey(ProfileScreen.nameFieldKey), 'J');
      await settleBriefly(tester);
      await tapVisible(tester, find.byKey(ProfileScreen.saveButtonKey));
      await settleBriefly(tester);

      expect(find.text('Name must be at least 2 characters'), findsOneWidget);
      expect(harness.requests.where((r) => r.method == 'PUT'), isEmpty);
    });

    testWidgets('rejects a phone that is not E.164', (tester) async {
      final harness = buildUnderTest();
      await enterEditMode(tester, harness.service);

      await tester.enterText(find.byKey(ProfileScreen.phoneFieldKey), '0812345');
      await settleBriefly(tester);
      await tapVisible(tester, find.byKey(ProfileScreen.saveButtonKey));
      await settleBriefly(tester);

      expect(
        find.text('Invalid phone format (e.g., +1234567890)'),
        findsOneWidget,
      );
      expect(harness.requests.where((r) => r.method == 'PUT'), isEmpty);
    });

    testWidgets('rejects an address shorter than five characters',
        (tester) async {
      final harness = buildUnderTest();
      await enterEditMode(tester, harness.service);

      await tester.enterText(find.byKey(ProfileScreen.addressFieldKey), 'abc');
      await settleBriefly(tester);
      await tapVisible(tester, find.byKey(ProfileScreen.saveButtonKey));
      await settleBriefly(tester);

      expect(find.text('Address must be at least 5 characters'), findsOneWidget);
      expect(harness.requests.where((r) => r.method == 'PUT'), isEmpty);
    });

    testWidgets('accepts empty optional fields', (tester) async {
      final harness = buildUnderTest(
        routes: {
          'PUT /profile': (_) => http.Response(
                jsonEncode(_profileJson(
                  phone: null,
                  emergencyContactName: null,
                  emergencyContactPhone: null,
                  address: null,
                )),
                200,
                headers: {'content-type': 'application/json'},
              ),
        },
      );
      await enterEditMode(tester, harness.service);

      await tester.enterText(find.byKey(ProfileScreen.phoneFieldKey), '');
      await tester.enterText(find.byKey(ProfileScreen.addressFieldKey), '');
      await tester.enterText(
        find.byKey(ProfileScreen.emergencyNameFieldKey),
        '',
      );
      await settleBriefly(tester);

      await tapVisible(tester, find.byKey(ProfileScreen.saveButtonKey));
      await settleBriefly(tester);

      expect(harness.requests.where((r) => r.method == 'PUT'), hasLength(1));
    });

    testWidgets('shows field-level errors returned by the API', (tester) async {
      final harness = buildUnderTest(
        routes: {
          'PUT /profile': (_) => http.Response(
                jsonEncode({
                  'error': 'Validation failed',
                  'errors': {'phone': 'Phone already in use'},
                }),
                400,
              ),
        },
      );
      await enterEditMode(tester, harness.service);

      await tapVisible(tester, find.byKey(ProfileScreen.saveButtonKey));
      await settleBriefly(tester);

      expect(find.text('Phone already in use'), findsOneWidget);
      expect(find.byKey(ProfileScreen.editButtonKey), findsNothing,
          reason: 'should stay in edit mode when the API rejects the update');
    });
  });

  group('ProfileScreen (8.3) - error state', () {
    testWidgets('shows the error and a retry action when loading fails',
        (tester) async {
      final harness = buildUnderTest(
        routes: {
          'GET /profile': (_) => http.Response(
                jsonEncode({'error': 'server on fire'}),
                500,
              ),
        },
      );
      await tester.pumpWidget(wrap(harness.service));
      await settleBriefly(tester);

      expect(find.text('server on fire'), findsOneWidget);
      expect(find.byKey(ProfileScreen.errorRetryButtonKey), findsOneWidget);
    });

    testWidgets('retry re-issues the request', (tester) async {
      var calls = 0;
      final harness = buildUnderTest(
        routes: {
          'GET /profile': (_) {
            calls++;
            if (calls == 1) {
              return http.Response(jsonEncode({'error': 'boom'}), 500);
            }
            return http.Response(
              jsonEncode(_profileJson()),
              200,
              headers: {'content-type': 'application/json'},
            );
          },
        },
      );
      await tester.pumpWidget(wrap(harness.service));
      await settleBriefly(tester);

      expect(find.text('boom'), findsOneWidget);

      await tapVisible(tester, find.byKey(ProfileScreen.errorRetryButtonKey));
      await settleBriefly(tester);

      expect(find.text('boom'), findsNothing);
      expect(find.text('John Doe'), findsWidgets);
    });

    testWidgets('surfaces a network failure without throwing', (tester) async {
      final client = MockClient((_) async => throw const _Offline());
      final service = ProfileService.withClient(
        client,
        sessionManagerFactory: FakeSessionManager.new,
      );

      await tester.pumpWidget(wrap(service));
      await settleBriefly(tester);

      expect(tester.takeException(), isNull);
      expect(find.textContaining('Network error'), findsOneWidget);
    });

    testWidgets('shows a session-expired message on 401', (tester) async {
      final harness = buildUnderTest(
        routes: {
          'GET /profile': (_) => http.Response(jsonEncode({'error': 'unauthorized'}), 401),
        },
      );
      await tester.pumpWidget(wrap(harness.service));
      await settleBriefly(tester);

      expect(find.textContaining('Session expired'), findsOneWidget);
    });
  });

  group('ProfileScreen (8.3) - change password dialog', () {
    Future<void> openDialog(WidgetTester tester, ProfileService service) async {
      await tester.pumpWidget(wrap(service));
      await settleBriefly(tester);
      await tapVisible(tester, find.byKey(ProfileScreen.changePasswordButtonKey));
      await settleBriefly(tester);
    }

    testWidgets('opens a dialog with the three password fields',
        (tester) async {
      final harness = buildUnderTest();
      await openDialog(tester, harness.service);

      expect(find.byType(AlertDialog), findsOneWidget);
      expect(find.text('Current Password'), findsOneWidget);
      expect(find.text('New Password'), findsOneWidget);
      expect(find.text('Confirm Password'), findsOneWidget);
    });

    testWidgets('requires the current password', (tester) async {
      final harness = buildUnderTest();
      await openDialog(tester, harness.service);

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Current Password'),
        '',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'New Password'),
        'NewPass456!',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Confirm Password'),
        'NewPass456!',
      );
      await settleBriefly(tester);

      await tapVisible(tester, find.widgetWithText(FilledButton, 'Change'));
      await settleBriefly(tester);

      expect(find.text('Current password is required'), findsOneWidget);
      expect(harness.requests.where((r) => r.method == 'PUT'), isEmpty);
    });

    testWidgets('rejects a weak new password', (tester) async {
      final harness = buildUnderTest();
      await openDialog(tester, harness.service);

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Current Password'),
        'OldPass123!',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'New Password'),
        'weak',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Confirm Password'),
        'weak',
      );
      await settleBriefly(tester);

      await tapVisible(tester, find.widgetWithText(FilledButton, 'Change'));
      await settleBriefly(tester);

      expect(find.textContaining('Password must be at least 8 characters'),
          findsOneWidget);
      expect(harness.requests.where((r) => r.method == 'PUT'), isEmpty);
    });

    testWidgets('rejects a mismatched confirmation', (tester) async {
      final harness = buildUnderTest();
      await openDialog(tester, harness.service);

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Current Password'),
        'OldPass123!',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'New Password'),
        'NewPass456!',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Confirm Password'),
        'Different789!',
      );
      await settleBriefly(tester);

      await tapVisible(tester, find.widgetWithText(FilledButton, 'Change'));
      await settleBriefly(tester);

      expect(find.text('Passwords do not match'), findsOneWidget);
      expect(harness.requests.where((r) => r.method == 'PUT'), isEmpty);
    });

    testWidgets('shows a strength indicator once a new password is typed',
        (tester) async {
      final harness = buildUnderTest();
      await openDialog(tester, harness.service);

      await tester.enterText(
        find.widgetWithText(TextFormField, 'New Password'),
        'NewPass456!',
      );
      await settleBriefly(tester);

      expect(find.text('Strength: Strong'), findsOneWidget);
    });

    testWidgets('submits a valid change and reports success', (tester) async {
      final harness = buildUnderTest(
        routes: {
          'PUT /profile/password': (_) =>
              http.Response(jsonEncode({'message': 'ok'}), 200),
        },
      );
      await openDialog(tester, harness.service);

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Current Password'),
        'OldPass123!',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'New Password'),
        'NewPass456!',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Confirm Password'),
        'NewPass456!',
      );
      await settleBriefly(tester);

      await tapVisible(tester, find.widgetWithText(FilledButton, 'Change'));
      await settleBriefly(tester);

      final put = harness.requests.firstWhere(
        (r) => r.method == 'PUT' && r.url.path.endsWith('/password'),
      );
      final body = jsonDecode(put.body) as Map<String, dynamic>;
      expect(body['current_password'], 'OldPass123!');
      expect(body['new_password'], 'NewPass456!');

      expect(find.byType(AlertDialog), findsNothing,
          reason: 'dialog should close on success');
    });

    testWidgets('keeps a generic message when the API rejects the change',
        (tester) async {
      final harness = buildUnderTest(
        routes: {
          'PUT /profile/password': (_) =>
              http.Response(jsonEncode({'error': 'Invalid credentials'}), 400),
        },
      );
      await openDialog(tester, harness.service);

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Current Password'),
        'WrongPass123!',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'New Password'),
        'NewPass456!',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Confirm Password'),
        'NewPass456!',
      );
      await settleBriefly(tester);

      await tapVisible(tester, find.widgetWithText(FilledButton, 'Change'));
      await settleBriefly(tester);

      expect(find.textContaining('Invalid credentials'), findsNothing);
      expect(find.textContaining('Password change failed'), findsWidgets);
    });

    testWidgets('cancel closes the dialog without calling the API',
        (tester) async {
      final harness = buildUnderTest();
      await openDialog(tester, harness.service);

      await tapVisible(tester, find.widgetWithText(TextButton, 'Cancel'));
      await settleBriefly(tester);

      expect(find.byType(AlertDialog), findsNothing);
      expect(harness.requests.where((r) => r.method == 'PUT'), isEmpty);
    });
  });
}

class _Offline implements Exception {
  const _Offline();
}
