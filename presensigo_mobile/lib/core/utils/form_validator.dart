/// Validation result with optional error message and metadata
class ValidationResult {
  final bool isValid;
  final String? errorMessage;
  final int strength; // For password strength (0-4)
  final List<String> failedRequirements; // For password requirements that failed

  ValidationResult({
    required this.isValid,
    this.errorMessage,
    this.strength = 0,
    this.failedRequirements = const [],
  });

  /// Convenience getter for backward compatibility
  String? get error => errorMessage;

  factory ValidationResult.valid() => ValidationResult(isValid: true);

  factory ValidationResult.invalid(String message) =>
      ValidationResult(isValid: false, errorMessage: message);

  factory ValidationResult.withStrength({
    required bool isValid,
    String? errorMessage,
    required int strength,
    required List<String> failedRequirements,
  }) =>
      ValidationResult(
        isValid: isValid,
        errorMessage: errorMessage,
        strength: strength,
        failedRequirements: failedRequirements,
      );
}

/// Helper class for form field validation
class FormValidator {
  // Validation constants
  static const int minNameLength = 2;
  static const int maxNameLength = 255;
  static const int minPasswordLength = 8;
  static const int maxPasswordLength = 255;
  static const int minAddressLength = 5;
  static const int maxAddressLength = 500;

  // Error messages
  static const String errorEmailInvalid =
      'Invalid email format (e.g., user@example.com)';
  static const String errorEmailEmpty = 'Email is required';
  static const String errorPasswordEmpty = 'Password is required';
  static const String errorPasswordShort =
      'Password must be at least $minPasswordLength characters';
  static const String errorPasswordWeak =
      'Password must contain uppercase, lowercase, number, and special character (!@#\$%^&*)';
  static const String errorNameShort = 'Name must be at least 2 characters';
  static const String errorNameLong = 'Name must not exceed 255 characters';
  static const String errorNameEmpty = 'Name is required';
  static const String errorPhoneInvalid =
      'Invalid phone format (e.g., +1234567890)';
  static const String errorAddressShort =
      'Address must be at least 5 characters';
  static const String errorAddressLong =
      'Address must not exceed 500 characters';
  static const String errorEmergencyContactNameInvalid =
      'Emergency contact name must be 2-255 characters';

  /// Validates email format
  /// Returns valid if string is empty (for optional fields)
  static ValidationResult validateEmail(String? value) {
    if (value == null || value.isEmpty) {
      return ValidationResult.invalid(errorEmailEmpty);
    }

    final emailRegex = RegExp(
      r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$',
    );

    if (!emailRegex.hasMatch(value.trim())) {
      return ValidationResult.invalid(errorEmailInvalid);
    }

    return ValidationResult.valid();
  }

  /// Validates password strength
  /// Requires: uppercase, lowercase, number, special character
  /// Returns valid if string is empty (for optional fields like current password in change context)
  /// Returns ValidationResult with strength and failedRequirements for UI feedback
  static ValidationResult validatePassword(String? value,
      {bool optional = false}) {
    if (value == null || value.isEmpty) {
      if (optional) {
        return ValidationResult.valid();
      }
      return ValidationResult.invalid(errorPasswordEmpty);
    }

    final hasUppercase = value.contains(RegExp(r'[A-Z]'));
    final hasLowercase = value.contains(RegExp(r'[a-z]'));
    final hasDigit = value.contains(RegExp(r'[0-9]'));
    final hasSpecial = value.contains(RegExp(r'[!@#$%^&*]'));
    final isLongEnough = value.length >= minPasswordLength;

    // Calculate strength
    int strength = 0;
    final failed = <String>[];

    if (hasUppercase) {
      strength++;
    } else {
      failed.add('uppercase');
    }

    if (hasLowercase) {
      strength++;
    } else {
      failed.add('lowercase');
    }

    if (hasDigit) {
      strength++;
    } else {
      failed.add('digit');
    }

    if (hasSpecial) {
      strength++;
    } else {
      failed.add('special');
    }

    if (isLongEnough) {
      strength++;
    } else {
      failed.add('length');
    }

    // Check if password is valid
    if (value.length < minPasswordLength) {
      return ValidationResult.withStrength(
        isValid: false,
        errorMessage: errorPasswordShort,
        strength: strength,
        failedRequirements: failed,
      );
    }

    if (!hasUppercase || !hasLowercase || !hasDigit || !hasSpecial) {
      return ValidationResult.withStrength(
        isValid: false,
        errorMessage: errorPasswordWeak,
        strength: strength,
        failedRequirements: failed,
      );
    }

    return ValidationResult.withStrength(
      isValid: true,
      strength: strength,
      failedRequirements: [],
    );
  }

  /// Validates phone number in E.164 format
  /// Returns valid if string is empty (optional field)
  static ValidationResult validatePhone(String? value) {
    if (value == null || value.isEmpty) {
      return ValidationResult.valid(); // Optional field
    }

    final phoneRegex = RegExp(r'^\+[1-9]\d{1,14}$');

    if (!phoneRegex.hasMatch(value.trim())) {
      return ValidationResult.invalid(errorPhoneInvalid);
    }

    return ValidationResult.valid();
  }

  /// Validates name (2-255 characters)
  /// Returns valid if string is empty (for optional fields)
  static ValidationResult validateName(String? value, {bool optional = false}) {
    if (value == null || value.isEmpty) {
      if (optional) {
        return ValidationResult.valid();
      }
      return ValidationResult.invalid(errorNameEmpty);
    }

    final trimmed = value.trim();

    if (trimmed.length < minNameLength) {
      return ValidationResult.invalid(errorNameShort);
    }

    if (trimmed.length > maxNameLength) {
      return ValidationResult.invalid(errorNameLong);
    }

    return ValidationResult.valid();
  }

  /// Validates address (5-500 characters when provided)
  /// Returns valid if string is empty (optional field)
  static ValidationResult validateAddress(String? value) {
    if (value == null || value.isEmpty) {
      return ValidationResult.valid(); // Optional field
    }

    final trimmed = value.trim();

    if (trimmed.length < minAddressLength) {
      return ValidationResult.invalid(errorAddressShort);
    }

    if (trimmed.length > maxAddressLength) {
      return ValidationResult.invalid(errorAddressLong);
    }

    return ValidationResult.valid();
  }

  /// Validates emergency contact fields (name and phone)
  /// Both are optional, but if provided must meet requirements
  static ValidationResult validateEmergencyContact(
    String? name,
    String? phone,
  ) {
    // If both are empty, it's valid (optional)
    if ((name == null || name.isEmpty) && (phone == null || phone.isEmpty)) {
      return ValidationResult.valid();
    }

    // If name is provided, validate it
    if (name != null && name.isNotEmpty) {
      final nameResult = validateName(name);
      if (!nameResult.isValid) {
        return ValidationResult.invalid(errorEmergencyContactNameInvalid);
      }
    }

    // If phone is provided, validate it
    if (phone != null && phone.isNotEmpty) {
      return validatePhone(phone);
    }

    return ValidationResult.valid();
  }

  /// Validates that two password strings match
  static ValidationResult validatePasswordMatch(String? password1, String? password2) {
    if (password1 == null || password2 == null) {
      return ValidationResult.invalid('Both passwords must be provided');
    }

    if (password1 != password2) {
      return ValidationResult.invalid('Passwords do not match');
    }

    return ValidationResult.valid();
  }

  /// Alias for validatePasswordMatch for backward compatibility
  static ValidationResult validateConfirmPassword(
    String? password,
    String? confirmPassword,
  ) =>
      validatePasswordMatch(password, confirmPassword);

  /// Determines password strength level
  static PasswordStrength getPasswordStrength(String? value) {
    if (value == null || value.isEmpty) {
      return PasswordStrength.none;
    }

    final hasUppercase = value.contains(RegExp(r'[A-Z]'));
    final hasLowercase = value.contains(RegExp(r'[a-z]'));
    final hasDigit = value.contains(RegExp(r'[0-9]'));
    final hasSpecial = value.contains(RegExp(r'[!@#$%^&*]'));
    final isLongEnough = value.length >= minPasswordLength;

    int strength = 0;
    if (hasUppercase) {
      strength++;
    }
    if (hasLowercase) {
      strength++;
    }
    if (hasDigit) {
      strength++;
    }
    if (hasSpecial) {
      strength++;
    }
    if (isLongEnough) {
      strength++;
    }

    // Of five requirements (upper, lower, digit, special, length), meeting only
    // one or two is not strong enough to call "medium".
    if (strength <= 2) {
      return PasswordStrength.weak;
    } else if (strength <= 4) {
      return PasswordStrength.medium;
    } else {
      return PasswordStrength.strong;
    }
  }

  /// Validates the entire registration form
  static bool isFormValid({
    required String email,
    required String password,
    required String confirmPassword,
    required String name,
    String? phone,
    required bool termsAccepted,
  }) {
    if (!termsAccepted) return false;

    final emailResult = validateEmail(email);
    if (!emailResult.isValid) return false;

    final passwordResult = validatePassword(password);
    if (!passwordResult.isValid) return false;

    final confirmResult = validatePasswordMatch(password, confirmPassword);
    if (!confirmResult.isValid) return false;

    final nameResult = validateName(name);
    if (!nameResult.isValid) return false;

    if (phone != null && phone.isNotEmpty) {
      final phoneResult = validatePhone(phone);
      if (!phoneResult.isValid) return false;
    }

    return true;
  }
}

/// Password strength levels
enum PasswordStrength {
  none,
  weak,
  medium,
  strong;

  String get label {
    switch (this) {
      case PasswordStrength.none:
        return 'None';
      case PasswordStrength.weak:
        return 'Weak';
      case PasswordStrength.medium:
        return 'Medium';
      case PasswordStrength.strong:
        return 'Strong';
    }
  }

  int get colorValue {
    switch (this) {
      case PasswordStrength.none:
        return 0xFF94A3B8; // gray
      case PasswordStrength.weak:
        return 0xFFEF4444; // red
      case PasswordStrength.medium:
        return 0xFFF59E0B; // amber
      case PasswordStrength.strong:
        return 0xFF10B981; // emerald
    }
  }
}
