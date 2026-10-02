import 'package:flutter_test/flutter_test.dart';
import 'package:presensigo_app/core/utils/form_validator.dart';

void main() {
  group('FormValidator - Email Validation (8.1)', () {
    test('validateEmail accepts valid format', () {
      final result = FormValidator.validateEmail('user@example.com');
      expect(result.isValid, true);
      expect(result.errorMessage, isNull);
    });

    test('validateEmail accepts valid format with subdomain', () {
      final result = FormValidator.validateEmail('user@mail.example.com');
      expect(result.isValid, true);
    });

    test('validateEmail accepts numbers in local part', () {
      final result = FormValidator.validateEmail('user123@example.com');
      expect(result.isValid, true);
    });

    test('validateEmail rejects invalid format - no at sign', () {
      final result = FormValidator.validateEmail('userexample.com');
      expect(result.isValid, false);
      expect(result.errorMessage, isNotNull);
    });

    test('validateEmail rejects invalid format - no domain', () {
      final result = FormValidator.validateEmail('user@');
      expect(result.isValid, false);
    });

    test('validateEmail rejects invalid format - no extension', () {
      final result = FormValidator.validateEmail('user@example');
      expect(result.isValid, false);
    });

    test('validateEmail rejects empty string', () {
      final result = FormValidator.validateEmail('');
      expect(result.isValid, false);
      expect(result.errorMessage, 'Email is required');
    });

    test('validateEmail rejects null', () {
      final result = FormValidator.validateEmail(null);
      expect(result.isValid, false);
    });

    test('validateEmail handles whitespace', () {
      final result = FormValidator.validateEmail('  user@example.com  ');
      expect(result.isValid, true);
    });

    test('validateEmail rejects special characters in local part', () {
      final result = FormValidator.validateEmail('user<>@example.com');
      expect(result.isValid, false);
    });
  });

  group('FormValidator - Password Validation (8.1)', () {
    test('validatePassword accepts strong password', () {
      final result = FormValidator.validatePassword('StrongPass123!');
      expect(result.isValid, true);
      expect(result.errorMessage, isNull);
      expect(result.strength, greaterThan(0));
    });

    test('validatePassword accepts password with multiple special chars', () {
      final result = FormValidator.validatePassword('Pass@word#123');
      expect(result.isValid, true);
    });

    test('validatePassword rejects password without uppercase', () {
      final result = FormValidator.validatePassword('strongpass123!');
      expect(result.isValid, false);
      expect(result.errorMessage, isNotNull);
      expect(result.failedRequirements, contains('uppercase'));
    });

    test('validatePassword rejects password without lowercase', () {
      final result = FormValidator.validatePassword('STRONGPASS123!');
      expect(result.isValid, false);
      expect(result.failedRequirements, contains('lowercase'));
    });

    test('validatePassword rejects password without number', () {
      final result = FormValidator.validatePassword('StrongPass!');
      expect(result.isValid, false);
      expect(result.failedRequirements, contains('digit'));
    });

    test('validatePassword rejects password without special character', () {
      final result = FormValidator.validatePassword('StrongPass123');
      expect(result.isValid, false);
      expect(result.failedRequirements, contains('special'));
    });

    test('validatePassword rejects password too short', () {
      final result = FormValidator.validatePassword('Pass1!');
      expect(result.isValid, false);
      expect(result.errorMessage, contains('at least'));
    });

    test('validatePassword rejects empty password', () {
      final result = FormValidator.validatePassword('');
      expect(result.isValid, false);
      expect(result.errorMessage, 'Password is required');
    });

    test('validatePassword rejects null password', () {
      final result = FormValidator.validatePassword(null);
      expect(result.isValid, false);
    });

    test('validatePassword with optional=true accepts empty', () {
      final result = FormValidator.validatePassword('', optional: true);
      expect(result.isValid, true);
    });

    test('validatePassword with optional=true accepts null', () {
      final result = FormValidator.validatePassword(null, optional: true);
      expect(result.isValid, true);
    });

    test('validatePassword minimum length exactly', () {
      final result = FormValidator.validatePassword('Pass1!!!');
      expect(result.isValid, true);
    });

    test('validatePassword accepts all required character types', () {
      final result = FormValidator.validatePassword('aB1!xxxx');
      expect(result.isValid, true);
    });

    test('validatePassword returns strength level', () {
      final result = FormValidator.validatePassword('StrongPass123!');
      expect(result.strength, greaterThan(0));
    });
  });

  group('FormValidator - Phone Validation (8.1)', () {
    test('validatePhone accepts valid E.164 format', () {
      final result = FormValidator.validatePhone('+12025551234');
      expect(result.isValid, true);
      expect(result.errorMessage, isNull);
    });

    test('validatePhone accepts valid E.164 with different countries', () {
      final result = FormValidator.validatePhone('+33123456789');
      expect(result.isValid, true);
    });

    test('validatePhone accepts phone with many digits', () {
      final result = FormValidator.validatePhone('+1234567890123456');
      expect(result.isValid, true);
    });

    test('validatePhone rejects without plus sign', () {
      final result = FormValidator.validatePhone('12025551234');
      expect(result.isValid, false);
    });

    test('validatePhone rejects invalid format', () {
      final result = FormValidator.validatePhone('+1 (202) 555-1234');
      expect(result.isValid, false);
    });

    test('validatePhone returns valid for empty string (optional)', () {
      final result = FormValidator.validatePhone('');
      expect(result.isValid, true);
    });

    test('validatePhone returns valid for null (optional)', () {
      final result = FormValidator.validatePhone(null);
      expect(result.isValid, true);
    });

    test('validatePhone rejects country code 0', () {
      final result = FormValidator.validatePhone('+0234567890');
      expect(result.isValid, false);
    });

    test('validatePhone rejects single digit after plus', () {
      final result = FormValidator.validatePhone('+1');
      expect(result.isValid, false);
    });

    test('validatePhone handles whitespace with trimming', () {
      final result = FormValidator.validatePhone('  +12025551234  ');
      expect(result.isValid, true);
    });
  });

  group('FormValidator - Name Validation (8.1)', () {
    test('validateName accepts valid name', () {
      final result = FormValidator.validateName('John Doe');
      expect(result.isValid, true);
      expect(result.errorMessage, isNull);
    });

    test('validateName accepts minimum length', () {
      final result = FormValidator.validateName('Jo');
      expect(result.isValid, true);
    });

    test('validateName rejects name too short', () {
      final result = FormValidator.validateName('J');
      expect(result.isValid, false);
      expect(result.errorMessage, contains('at least 2'));
    });

    test('validateName accepts maximum length', () {
      final result = FormValidator.validateName('A' * 255);
      expect(result.isValid, true);
    });

    test('validateName rejects name too long', () {
      final result = FormValidator.validateName('A' * 256);
      expect(result.isValid, false);
      expect(result.errorMessage, contains('not exceed 255'));
    });

    test('validateName rejects empty string', () {
      final result = FormValidator.validateName('');
      expect(result.isValid, false);
      expect(result.errorMessage, 'Name is required');
    });

    test('validateName rejects null without optional flag', () {
      final result = FormValidator.validateName(null);
      expect(result.isValid, false);
    });

    test('validateName returns valid for null with optional=true', () {
      final result = FormValidator.validateName(null, optional: true);
      expect(result.isValid, true);
    });

    test('validateName returns valid for empty with optional=true', () {
      final result = FormValidator.validateName('', optional: true);
      expect(result.isValid, true);
    });

    test('validateName trims whitespace', () {
      final result = FormValidator.validateName('  John Doe  ');
      expect(result.isValid, true);
    });

    test('validateName accepts names with special characters', () {
      final result = FormValidator.validateName("O'Brien-Smith");
      expect(result.isValid, true);
    });
  });

  group('FormValidator - Address Validation (8.1)', () {
    test('validateAddress accepts valid address', () {
      final result = FormValidator.validateAddress('123 Main Street');
      expect(result.isValid, true);
      expect(result.errorMessage, isNull);
    });

    test('validateAddress accepts minimum length', () {
      final result = FormValidator.validateAddress('12345');
      expect(result.isValid, true);
    });

    test('validateAddress rejects too short', () {
      final result = FormValidator.validateAddress('1234');
      expect(result.isValid, false);
      expect(result.errorMessage, contains('at least 5'));
    });

    test('validateAddress accepts maximum length', () {
      final result = FormValidator.validateAddress('A' * 500);
      expect(result.isValid, true);
    });

    test('validateAddress rejects too long', () {
      final result = FormValidator.validateAddress('A' * 501);
      expect(result.isValid, false);
      expect(result.errorMessage, contains('not exceed 500'));
    });

    test('validateAddress returns valid for empty (optional)', () {
      final result = FormValidator.validateAddress('');
      expect(result.isValid, true);
    });

    test('validateAddress returns valid for null (optional)', () {
      final result = FormValidator.validateAddress(null);
      expect(result.isValid, true);
    });

    test('validateAddress trims whitespace', () {
      final result = FormValidator.validateAddress('  123 Main St  ');
      expect(result.isValid, true);
    });
  });

  group('FormValidator - Emergency Contact Validation (8.1)', () {
    test('validateEmergencyContact accepts both empty (optional)', () {
      final result = FormValidator.validateEmergencyContact(null, null);
      expect(result.isValid, true);
    });

    test('validateEmergencyContact accepts valid name and phone', () {
      final result =
          FormValidator.validateEmergencyContact('Jane Doe', '+12025551234');
      expect(result.isValid, true);
      expect(result.errorMessage, isNull);
    });

    test('validateEmergencyContact accepts only valid name', () {
      final result = FormValidator.validateEmergencyContact('Jane Doe', null);
      expect(result.isValid, true);
    });

    test('validateEmergencyContact accepts only valid phone', () {
      final result = FormValidator.validateEmergencyContact(null, '+12025551234');
      expect(result.isValid, true);
    });

    test('validateEmergencyContact rejects invalid name', () {
      final result = FormValidator.validateEmergencyContact('J', '+12025551234');
      expect(result.isValid, false);
    });

    test('validateEmergencyContact rejects invalid phone', () {
      final result = FormValidator.validateEmergencyContact('Jane Doe', '12025551234');
      expect(result.isValid, false);
    });

    test('validateEmergencyContact accepts empty strings (optional)', () {
      final result = FormValidator.validateEmergencyContact('', '');
      expect(result.isValid, true);
    });
  });

  group('FormValidator - Password Match Validation (8.1)', () {
    test('validatePasswordMatch accepts matching passwords', () {
      final result = FormValidator.validatePasswordMatch('Pass123!', 'Pass123!');
      expect(result.isValid, true);
      expect(result.errorMessage, isNull);
    });

    test('validatePasswordMatch rejects non-matching passwords', () {
      final result = FormValidator.validatePasswordMatch('Pass123!', 'Pass456!');
      expect(result.isValid, false);
      expect(result.errorMessage, contains('do not match'));
    });

    test('validatePasswordMatch rejects null first password', () {
      final result = FormValidator.validatePasswordMatch(null, 'Pass123!');
      expect(result.isValid, false);
    });

    test('validatePasswordMatch rejects null second password', () {
      final result = FormValidator.validatePasswordMatch('Pass123!', null);
      expect(result.isValid, false);
    });

    test('validatePasswordMatch rejects both null', () {
      final result = FormValidator.validatePasswordMatch(null, null);
      expect(result.isValid, false);
    });

    test('validatePasswordMatch is case sensitive', () {
      final result = FormValidator.validatePasswordMatch('pass123!', 'Pass123!');
      expect(result.isValid, false);
    });

    test('validateConfirmPassword (alias) works correctly', () {
      final result = FormValidator.validateConfirmPassword('Pass123!', 'Pass123!');
      expect(result.isValid, true);
    });
  });

  group('FormValidator - Password Strength Levels (8.1)', () {
    test('getPasswordStrength returns none for empty', () {
      final strength = FormValidator.getPasswordStrength('');
      expect(strength, PasswordStrength.none);
    });

    test('getPasswordStrength returns none for null', () {
      final strength = FormValidator.getPasswordStrength(null);
      expect(strength, PasswordStrength.none);
    });

    test('getPasswordStrength returns weak for minimal strength', () {
      final strength = FormValidator.getPasswordStrength('abcdef');
      expect(strength, PasswordStrength.weak);
    });

    test('getPasswordStrength returns weak for password with only lowercase', () {
      final strength = FormValidator.getPasswordStrength('abcdefghij');
      expect(strength, PasswordStrength.weak);
    });

    test('getPasswordStrength returns medium for password with 2 requirements', () {
      final strength = FormValidator.getPasswordStrength('Abcdefgh');
      expect(strength, PasswordStrength.medium);
    });

    test('getPasswordStrength returns medium for password with 3 requirements', () {
      final strength = FormValidator.getPasswordStrength('Abcdefgh123');
      expect(strength, PasswordStrength.medium);
    });

    test('getPasswordStrength returns strong for password with 4+ requirements', () {
      final strength = FormValidator.getPasswordStrength('Abcdefgh123!');
      expect(strength, PasswordStrength.strong);
    });

    test('getPasswordStrength returns strong for password with all 5 requirements', () {
      final strength = FormValidator.getPasswordStrength('StrongPass123!');
      expect(strength, PasswordStrength.strong);
    });

    test('PasswordStrength enum has correct labels', () {
      expect(PasswordStrength.none.label, 'None');
      expect(PasswordStrength.weak.label, 'Weak');
      expect(PasswordStrength.medium.label, 'Medium');
      expect(PasswordStrength.strong.label, 'Strong');
    });

    test('PasswordStrength enum has color values', () {
      expect(PasswordStrength.none.colorValue, isNotNull);
      expect(PasswordStrength.weak.colorValue, isNotNull);
      expect(PasswordStrength.medium.colorValue, isNotNull);
      expect(PasswordStrength.strong.colorValue, isNotNull);
    });
  });
}
