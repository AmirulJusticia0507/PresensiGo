/// Thrown when an API call returns a non-success HTTP status.
///
/// Carries the status code and, for validation failures, the field-level
/// details returned by the backend so the UI can render them inline.
class ApiException implements Exception {
  /// Human readable, already-sanitized message safe to show to the user.
  final String message;

  /// HTTP status code returned by the backend.
  final int statusCode;

  /// Field-level validation details, each with `field` and `reason` keys.
  final List<Map<String, String?>>? details;

  const ApiException({
    required this.message,
    required this.statusCode,
    this.details,
  });

  /// 400 with field-level details attached.
  factory ApiException.validation(
    String message,
    List<Map<String, String?>> details,
  ) =>
      ApiException(message: message, statusCode: 400, details: details);

  /// 409 for duplicate resources such as an already registered email.
  factory ApiException.conflict(String message) =>
      ApiException(message: message, statusCode: 409);

  /// 401 when the session is missing, invalid, or expired.
  factory ApiException.unauthorized([String? message]) =>
      ApiException(message: message ?? 'Unauthorized', statusCode: 401);

  @override
  String toString() => 'ApiException($statusCode): $message';
}
