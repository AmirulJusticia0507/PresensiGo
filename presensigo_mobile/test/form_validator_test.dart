import 'package:flutter_test/flutter_test.dart';
import 'package:presensigo_app/core/utils/form_validator.dart';

void main() {
  group('FormValidator', () {
    group('validateEmail', () {
      test('returns valid for correct email format', () {
        final result = FormValidator.validateEmail('user@example.com');
        expect(result.isValid, true);
        expect(result.errorMessage, null);
      });

      test('returns invalid for empty email', () {
        final result = FormValidator.validateEmail('');
        expect(result.isValid, false);
        expect(result.errorMessage, isNotNull);
      });

      test('returns invalid for email without @', () {
        final result = FormValidator.validateEmail('userexample.com');
        expect(result.isValid, false);
      });

      test('returns invalid for email without domain', () {
        final result = FormValidator.validateEmail('user@');
        expect(result.isValid, false);
      });

      test('returns valid for complex email', () {
        final result = FormValidator.validateEmail('user.name+tag@example.co.uk');
        expect(result.isValid, true);
      });
    });

    group('validatePassword', () {
      test('returns valid for strong password', () {
        final result = FormValidator.validatePassword('StrongPass123!');
        expect(result.isValid, true);
        expect(result.errorMessage, null);
      });

      test('returns invalid for empty password', () {
        final result = FormValidator.validatePassword('');
        expect(result.isValid, false);
      });

      test('returns invalid for password too short', () {
        final result = FormValidator.validatePassword('Short1!');
        expect(result.isValid, false);
      });

      test('returns invalid for password without uppercase', () {
        final result = FormValidator.validatePassword('password123!');
        expect(result.isValid, false);
      });

      test('returns invalid for password without lowercase', () {
        final result = FormValidator.validatePassword('PASSWORD123!');
        expect(result.isValid, false);
      });

      test('returns invalid for password without number', () {
        final result = FormValidator.validatePassword('Password!');
        expect(result.isValid, false);
      });

      test('returns invalid for password without special character', () {
        final result = FormValidator.validatePassword('Password123');
        expect(result.isValid, false);
      });

      test('returns valid when optional and empty', () {
        final result = FormValidator.validatePassword('', optional: true);
        expect(result.isValid, true);
      });

      test('returns valid for password with special chars', () {
        final result = FormValidator.validatePassword('Pass@123');
        expect(result.isValid, true);
      });
    });

    group('validatePhone', () {
      test('returns valid for E.164 format', () {
        final result = FormValidator.validatePhone('+12025551234');
        expect(result.isValid, true);
      });

      test('returns valid for empty phone (optional)', () {
        final result = FormValidator.validatePhone('');
        expect(result.isValid, true);
      });

      test('returns invalid for non-E.164 format', () {
        final result = FormValidator.validatePhone('2025551234');
        expect(result.isValid, false);
      });

      test('returns invalid for phone without +', () {
        final result = FormValidator.validatePhone('12025551234');
        expect(result.isValid, false);
      });

      test('returns valid for international format', () {
        final result = FormValidator.validatePhone('+441234567890');
        expect(result.isValid, true);
      });
    });

    group('validateName', () {
      test('returns valid for normal name', () {
        final result = FormValidator.validateName('John Doe');
        expect(result.isValid, true);
      });

      test('returns invalid for empty name', () {
        final result = FormValidator.validateName('');
        expect(result.isValid, false);
      });

      test('returns invalid for single character', () {
        final result = FormValidator.validateName('J');
        expect(result.isValid, false);
      });

      test('returns valid for 2-character name', () {
        final result = FormValidator.validateName('Jo');
        expect(result.isValid, true);
      });

      test('returns invalid for name exceeding 255 chars', () {
        final longName = 'A' * 256;
        final result = FormValidator.validateName(longName);
        expect(result.isValid, false);
      });

      test('returns valid when optional and empty', () {
        final result = FormValidator.validateName('', optional: true);
        expect(result.isValid, true);
      });

      test('returns valid for 255-character name', () {
        final longName = 'A' * 255;
        final result = FormValidator.validateName(longName);
        expect(result.isValid, true);
      });
    });

    group('validateAddress', () {
      test('returns valid for normal address', () {
        final result = FormValidator.validateAddress('123 Main St, City, State');
        expect(result.isValid, true);
      });

      test('returns valid for empty address (optional)', () {
        final result = FormValidator.validateAddress('');
        expect(result.isValid, true);
      });

      test('returns invalid for address too short', () {
        final result = FormValidator.validateAddress('123 St');
        expect(result.isValid, false);
      });

      test('returns valid for 5-character address', () {
        final result = FormValidator.validateAddress('12345');
        expect(result.isValid, true);
      });

      test('returns invalid for address exceeding 500 chars', () {
        final longAddress = 'A' * 501;
        final result = FormValidator.validateAddress(longAddress);
        expect(result.isValid, false);
      });

      test('returns valid for 500-character address', () {
        final longAddress = 'A' * 500;
        final result = FormValidator.validateAddress(longAddress);
        expect(result.isValid, true);
      });
    });

    group('validateEmergencyContact', () {
      test('returns valid for both empty (optional)', () {
        final result = FormValidator.validateEmergencyContact(null, null);
        expect(result.isValid, true);
      });

      test('returns valid for valid name and empty phone', () {
        final result = FormValidator.validateEmergencyContact('Jane Doe', null);
        expect(result.isValid, true);
      });

      test('returns valid for empty name and valid phone', () {
        final result =
            FormValidator.validateEmergencyContact(null, '+12025551234');
        expect(result.isValid, true);
      });

      test('returns valid for both provided and valid', () {
        final result = FormValidator.validateEmergencyContact(
          'Jane Doe',
          '+12025551234',
        );
        expect(result.isValid, true);
      });

      test('returns invalid for invalid name', () {
        final result =
            FormValidator.validateEmergencyContact('J', '+12025551234');
        expect(result.isValid, false);
      });

      test('returns invalid for invalid phone', () {
        final result = FormValidator.validateEmergencyContact('Jane Doe', '123');
        expect(result.isValid, false);
      });
    });

    group('validatePasswordMatch', () {
      test('returns valid when passwords match', () {
        final result =
            FormValidator.validatePasswordMatch('Password123!', 'Password123!');
        expect(result.isValid, true);
      });

      test('returns invalid when passwords do not match', () {
        final result =
            FormValidator.validatePasswordMatch('Password123!', 'Password456!');
        expect(result.isValid, false);
      });

      test('returns invalid when either is null', () {
        final result = FormValidator.validatePasswordMatch(null, 'Password123!');
        expect(result.isValid, false);
      });
    });

    group('getPasswordStrength', () {
      test('returns weak for weak password', () {
        final strength = FormValidator.getPasswordStrength('weak');
        expect(strength, PasswordStrength.weak);
      });

      test('returns medium for medium password', () {
        final strength = FormValidator.getPasswordStrength('Weak123');
        expect(strength, PasswordStrength.medium);
      });

      test('returns strong for strong password', () {
        final strength = FormValidator.getPasswordStrength('Strong123!');
        expect(strength, PasswordStrength.strong);
      });

      test('returns none for empty password', () {
        final strength = FormValidator.getPasswordStrength('');
        expect(strength, PasswordStrength.none);
      });

      test('has correct label', () {
        expect(PasswordStrength.weak.label, 'Weak');
        expect(PasswordStrength.medium.label, 'Medium');
        expect(PasswordStrength.strong.label, 'Strong');
      });
    });
  });
}
