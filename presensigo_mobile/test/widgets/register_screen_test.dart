import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:presensigo_app/core/constants/api_constants.dart';
import 'package:presensigo_app/data/services/profile_service.dart';
import 'package:presensigo_app/data/services/session_manager.dart';
import 'package:presensigo_app/features/auth/screens/register_screen.dart';

/// SessionManager stand-in so tests never touch the secure-storage channel.
class FakeSessionManager extends SessionManager {
  FakeSessionManager({this.token});

  String? token;

  @override
  Future<String?> getToken() async => token;

  @override
  Future<void> saveToken(String value) async => token = value;

  @override
  Future<void> logout() async => token = null;

  @override
  Future<void> handleUnauthorized() => logout();
}

Map<String, dynamic> _userJson() => {
      'id': 'user-1',
      'name': 'John Doe',
      'email': 'john@example.com',
      'phone': null,
      'emergency_contact_name': null,
      'emergency_contact_phone': null,
      'address': null,
      'profile_picture_url': null,
      'face_enrolled': false,
      'face_enrolled_at': null,
      'created_at': '2026-01-01T00:00:00Z',
      'updated_at': '2026-01-01T00:00:00Z',
    };

/// Builds the screen under test with a service backed by [handler].
({ProfileService service, List<http.Request> requests}) buildUnderTest(
  FutureOr<http.Response> Function(http.Request request) handler,
) {
  final requests = <http.Request>[];
  final client = MockClient((request) async {
    requests.add(request);
    return handler(request);
  });
  return (
    service: ProfileService.withClient(
      client,
      sessionManagerFactory: () => FakeSessionManager(token: 'token'),
    ),
    requests: requests,
  );
}

http.Response _created() => http.Response(
      jsonEncode({
        'message': 'registration successful',
        'token': 'jwt-123',
        'user': _userJson(),
      }),
      201,
    );

/// Scrolls [finder] into view before tapping it, since the form lives inside a
/// SingleChildScrollView that does not fit the default 800x600 test viewport.
Future<void> tapVisible(WidgetTester tester, Finder finder) async {
  await tester.ensureVisible(finder);
  await tester.pumpAndSettle();
  await tester.tap(finder);
  await tester.pump();
}

/// Pumps a bounded number of frames.
///
/// [WidgetTester.pumpAndSettle] cannot be used after a successful registration
/// because the screen navigates to AttendanceScreen, which runs a continuous
/// animation and therefore never settles.
Future<void> settleBriefly(WidgetTester tester, {int frames = 10}) async {
  for (var i = 0; i < frames; i++) {
    await tester.pump(const Duration(milliseconds: 100));
  }
}

/// Fills every field with valid data and accepts the terms.
Future<void> fillValidForm(WidgetTester tester) async {
  await tester.enterText(
    find.byKey(RegisterScreen.emailFieldKey),
    'john@example.com',
  );
  await tester.enterText(find.byKey(RegisterScreen.nameFieldKey), 'John Doe');
  await tester.enterText(
    find.byKey(RegisterScreen.passwordFieldKey),
    'Password123!',
  );
  await tester.enterText(
    find.byKey(RegisterScreen.confirmPasswordFieldKey),
    'Password123!',
  );
  await tapVisible(tester, find.byKey(RegisterScreen.termsCheckboxKey));
}

void main() {
  group('RegisterScreen (8.2) - form display', () {
    testWidgets('displays all form fields, terms row and submit button',
        (tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));

      expect(find.text('Email Address'), findsOneWidget);
      expect(find.text('Full Name'), findsOneWidget);
      expect(find.text('Phone Number (Optional)'), findsOneWidget);
      expect(find.text('Password'), findsOneWidget);
      expect(find.text('Confirm Password'), findsOneWidget);
      expect(find.text('I accept the terms and conditions'), findsOneWidget);
      expect(find.byKey(RegisterScreen.termsCheckboxKey), findsOneWidget);
      expect(find.byKey(RegisterScreen.submitButtonKey), findsOneWidget);
      expect(find.text('Already have an account? '), findsOneWidget);
    });

    testWidgets('renders exactly five text fields', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));

      expect(find.byType(TextField), findsNWidgets(5));
    });
  });

  group('RegisterScreen (8.2) - validation feedback', () {
    testWidgets('shows an email error for a malformed address', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));

      await tester.enterText(
        find.byKey(RegisterScreen.emailFieldKey),
        'invalidemail',
      );
      await tester.pump();

      expect(
        find.text('Invalid email format (e.g., user@example.com)'),
        findsOneWidget,
      );
    });

    testWidgets('clears the email error once a valid address is entered',
        (tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));

      final field = find.byKey(RegisterScreen.emailFieldKey);
      await tester.enterText(field, 'invalidemail');
      await tester.pump();
      expect(
        find.text('Invalid email format (e.g., user@example.com)'),
        findsOneWidget,
      );

      await tester.enterText(field, 'john@example.com');
      await tester.pump();
      expect(
        find.text('Invalid email format (e.g., user@example.com)'),
        findsNothing,
      );
    });

    testWidgets('shows a name error for a one-character name', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));

      await tester.enterText(find.byKey(RegisterScreen.nameFieldKey), 'J');
      await tester.pump();

      expect(find.text('Name must be at least 2 characters'), findsOneWidget);
    });

    testWidgets('shows a mismatch error when the confirmation differs',
        (tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));

      await tester.enterText(
        find.byKey(RegisterScreen.passwordFieldKey),
        'Password123!',
      );
      await tester.enterText(
        find.byKey(RegisterScreen.confirmPasswordFieldKey),
        'Different123!',
      );
      await tester.pump();

      expect(find.text('Passwords do not match'), findsOneWidget);
    });

    testWidgets('shows a phone error for a non-E.164 number', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));

      await tester.enterText(find.byKey(RegisterScreen.phoneFieldKey), '0812345');
      await tester.pump();

      expect(
        find.text('Invalid phone format (e.g., +1234567890)'),
        findsOneWidget,
      );
    });

    testWidgets('accepts an empty optional phone without an error',
        (tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));

      expect(
        find.text('Invalid phone format (e.g., +1234567890)'),
        findsNothing,
      );
    });
  });

  group('RegisterScreen (8.2) - password strength indicator', () {
    testWidgets('lists every password requirement', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));

      expect(find.byKey(RegisterScreen.strengthIndicatorKey), findsOneWidget);
      expect(find.text('Password requirements:'), findsOneWidget);
      expect(find.text('At least 8 characters'), findsOneWidget);
      expect(find.text('Uppercase letter (A-Z)'), findsOneWidget);
      expect(find.text('Lowercase letter (a-z)'), findsOneWidget);
      expect(find.text('Number (0-9)'), findsOneWidget);
      expect(find.text(r'Special character (!@#$%^&*)'), findsOneWidget);
    });

    testWidgets('reports Strong once all requirements are met', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));

      await tester.enterText(
        find.byKey(RegisterScreen.passwordFieldKey),
        'Password123!',
      );
      await tester.pump();

      expect(find.text('Strong'), findsOneWidget);
    });

    testWidgets('reports Weak for a short password', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));

      await tester.enterText(find.byKey(RegisterScreen.passwordFieldKey), 'Pass');
      await tester.pump();

      expect(find.text('Weak'), findsOneWidget);
    });

    testWidgets('toggles password visibility', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));

      await tester.enterText(
        find.byKey(RegisterScreen.passwordFieldKey),
        'Password123!',
      );
      await tester.pump();

      expect(find.byIcon(Icons.visibility_off), findsNWidgets(2));

      await tester.tap(find.byIcon(Icons.visibility_off).first);
      await tester.pump();

      expect(find.byIcon(Icons.visibility), findsAtLeastNWidgets(1));
    });
  });

  group('RegisterScreen (8.2) - submit button state', () {
    testWidgets('is disabled on an empty form', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));

      final button =
          tester.widget<ElevatedButton>(find.byKey(RegisterScreen.submitButtonKey));
      expect(button.onPressed, isNull);
    });

    testWidgets('stays disabled while the terms are not accepted',
        (tester) async {
      final harness = buildUnderTest((_) => _created());
      await tester.pumpWidget(
        MaterialApp(home: RegisterScreen(service: harness.service)),
      );

      await tester.enterText(
        find.byKey(RegisterScreen.emailFieldKey),
        'john@example.com',
      );
      await tester.enterText(find.byKey(RegisterScreen.nameFieldKey), 'John Doe');
      await tester.enterText(
        find.byKey(RegisterScreen.passwordFieldKey),
        'Password123!',
      );
      await tester.enterText(
        find.byKey(RegisterScreen.confirmPasswordFieldKey),
        'Password123!',
      );
      await tester.pump();

      final button =
          tester.widget<ElevatedButton>(find.byKey(RegisterScreen.submitButtonKey));
      expect(button.onPressed, isNull,
          reason: 'terms acceptance is required');
    });

    testWidgets('is enabled once the whole form is valid', (tester) async {
      final harness = buildUnderTest((_) => _created());
      await tester.pumpWidget(
        MaterialApp(home: RegisterScreen(service: harness.service)),
      );

      await fillValidForm(tester);

      final button =
          tester.widget<ElevatedButton>(find.byKey(RegisterScreen.submitButtonKey));
      expect(button.onPressed, isNotNull);
    });
  });

  group('RegisterScreen (8.2) - terms checkbox', () {
    testWidgets('toggles via the checkbox', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));

      final finder = find.byKey(RegisterScreen.termsCheckboxKey);
      expect(tester.widget<Checkbox>(finder).value, isFalse);

      await tapVisible(tester, finder);
      expect(tester.widget<Checkbox>(finder).value, isTrue);

      await tapVisible(tester, finder);
      expect(tester.widget<Checkbox>(finder).value, isFalse);
    });

    testWidgets('toggles when the label text is tapped', (tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterScreen()));

      await tapVisible(tester, find.text('I accept the terms and conditions'));

      expect(
        tester
            .widget<Checkbox>(find.byKey(RegisterScreen.termsCheckboxKey))
            .value,
        isTrue,
      );
    });
  });

  group('RegisterScreen (8.2) - submit behaviour', () {
    testWidgets('sends the registration payload', (tester) async {
      final harness = buildUnderTest((_) => _created());
      await tester.pumpWidget(
        MaterialApp(home: RegisterScreen(service: harness.service)),
      );

      await fillValidForm(tester);
      await tapVisible(tester, find.byKey(RegisterScreen.submitButtonKey));
      await settleBriefly(tester);

      final body =
          jsonDecode(harness.requests.single.body) as Map<String, dynamic>;
      expect(body['email'], 'john@example.com');
      expect(body['name'], 'John Doe');
      expect(body['password'], 'Password123!');
      expect(body['confirm_password'], 'Password123!');
      expect(body['terms_accepted'], isTrue);
    });

    testWidgets('shows a loading indicator while the request is in flight',
        (tester) async {
      final completer = Completer<http.Response>();
      final harness = buildUnderTest((_) => completer.future);
      await tester.pumpWidget(
        MaterialApp(home: RegisterScreen(service: harness.service)),
      );

      await fillValidForm(tester);
      await tapVisible(tester, find.byKey(RegisterScreen.submitButtonKey));
      await tester.pump();

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
      final button =
          tester.widget<ElevatedButton>(find.byKey(RegisterScreen.submitButtonKey));
      expect(button.onPressed, isNull, reason: 'cannot double-submit');

      completer.complete(_created());
      await settleBriefly(tester);
    });

    testWidgets('surfaces a 409 duplicate email on the email field',
        (tester) async {
      final harness = buildUnderTest(
        (_) => http.Response(jsonEncode({'error': 'Email already registered'}), 409),
      );
      await tester.pumpWidget(
        MaterialApp(home: RegisterScreen(service: harness.service)),
      );

      await fillValidForm(tester);
      await tapVisible(tester, find.byKey(RegisterScreen.submitButtonKey));
      await settleBriefly(tester);

      expect(find.text('Email already registered'), findsWidgets);
      expect(find.text('This email is already in use'), findsOneWidget);
    });

    testWidgets('maps 400 field details onto the matching fields',
        (tester) async {
      final harness = buildUnderTest(
        (_) => http.Response(
          jsonEncode({
            'error': 'Validation failed',
            'details': [
              {'field': 'phone', 'reason': 'Invalid phone format'},
            ],
          }),
          400,
        ),
      );
      await tester.pumpWidget(
        MaterialApp(home: RegisterScreen(service: harness.service)),
      );

      await fillValidForm(tester);
      await tapVisible(tester, find.byKey(RegisterScreen.submitButtonKey));
      await settleBriefly(tester);

      expect(find.text('Invalid phone format'), findsOneWidget);
    });

    testWidgets('shows a sanitized message on a 500 without leaking the body',
        (tester) async {
      final harness = buildUnderTest(
        (_) => http.Response('internal stack trace and table users', 500),
      );
      await tester.pumpWidget(
        MaterialApp(home: RegisterScreen(service: harness.service)),
      );

      await fillValidForm(tester);
      await tapVisible(tester, find.byKey(RegisterScreen.submitButtonKey));
      await settleBriefly(tester);

      expect(find.text('Request failed. Please try again.'), findsWidgets);
      expect(find.textContaining('stack trace'), findsNothing);
      expect(find.textContaining('table users'), findsNothing);
    });

    testWidgets('does not throw when the network is unreachable',
        (tester) async {
      final client = MockClient((_) async => throw const _Offline());
      final service = ProfileService.withClient(
        client,
        sessionManagerFactory: () => FakeSessionManager(token: 'token'),
      );
      await tester.pumpWidget(MaterialApp(home: RegisterScreen(service: service)));

      await fillValidForm(tester);
      await tapVisible(tester, find.byKey(RegisterScreen.submitButtonKey));
      await settleBriefly(tester);

      expect(tester.takeException(), isNull);
    });

    testWidgets('does not send a request while the form is invalid',
        (tester) async {
      final harness = buildUnderTest((_) => _created());
      await tester.pumpWidget(
        MaterialApp(home: RegisterScreen(service: harness.service)),
      );

      await tester.enterText(
        find.byKey(RegisterScreen.emailFieldKey),
        'not-an-email',
      );
      await tester.pump();

      final button =
          tester.widget<ElevatedButton>(find.byKey(RegisterScreen.submitButtonKey));
      expect(button.onPressed, isNull);
      expect(harness.requests, isEmpty);
    });
  });
}

class _Offline implements Exception {
  const _Offline();
}
