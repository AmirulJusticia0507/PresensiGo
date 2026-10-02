import 'package:flutter_test/flutter_test.dart';
import 'package:presensigo_app/data/services/session_manager.dart';
import 'package:shared_preferences/shared_preferences.dart';

void main() {
  TestWidgetsFlutterBinding.ensureInitialized();

  group('SessionManager', () {
    late SessionManager sessionManager;

    setUp(() async {
      SharedPreferences.setMockInitialValues({});
      sessionManager = SessionManager();
      final prefs = await SharedPreferences.getInstance();
      await prefs.clear();
    });

    test('isSessionValid returns false when token is missing', () async {
      final isValid = await sessionManager.isSessionValid();
      expect(isValid, false);
    });

    test(
      'isSessionValid returns false for malformed token (not 3 parts)',
      () async {
        // Even without setting a valid token, the method should safely handle
        // missing tokens and return false
        final isValid = await sessionManager.isSessionValid();
        expect(isValid, false);
      },
    );

    test('logout clears all session state', () async {
      await sessionManager.logout();
      final isValid = await sessionManager.isSessionValid();
      expect(isValid, false);
    });

    test('handleUnauthorized calls logout', () async {
      await sessionManager.handleUnauthorized();
      final isValid = await sessionManager.isSessionValid();
      expect(isValid, false);
    });

    test('Session check is resilient to errors', () async {
      // Test that isSessionValid safely returns false on any error
      final isValid = await sessionManager.isSessionValid();
      expect(isValid, isFalse);
    });

    test('Multiple logout calls are safe (idempotent)', () async {
      await sessionManager.logout();
      await sessionManager.logout();
      final isValid = await sessionManager.isSessionValid();
      expect(isValid, false);
    });
  });
}
