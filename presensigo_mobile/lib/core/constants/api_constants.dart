class ApiConstants {
  /// Base URL including the `/api` prefix, so paths below are appended directly.
  /// Override per build with:
  /// `--dart-define=API_BASE_URL=https://staging-api.example.com/api`.
  static const String baseUrl = String.fromEnvironment(
    'API_BASE_URL',
    defaultValue: 'http://10.0.2.2:8080/api',
  );
  static const String authLogin = '/auth/login';
  static const String authRegister = '/auth/register';
  static const String authLogout = '/auth/logout';
  static const String attendanceCheckIn = '/attendance/check-in';
  static const String attendanceCheckOut = '/attendance/check-out';
  static const String attendanceToday = '/attendance/today';
  static const String attendanceHistory = '/attendance/history';
  static const String attendanceSync = '/attendance/sync';
  static const String locations = '/locations';
  static const String faceChallenge = '/face/challenge';
  static const String faceEnrollment = '/profile/face-enrollment';
  static const String profile = '/profile';
  static const String profilePassword = '/profile/password';
  static const String locationAttempts = '/security/location-attempts';
  static const String locationAlerts = '/admin/security/location-alerts';
  static const String attendanceHistoryPage = '/attendance/history/page';
  static const String leaves = '/leaves';
  static const String adminAttendances = '/admin/attendances';
  static const String adminAttendanceCsv = '/admin/attendances.csv';
  static const String adminUsers = '/admin/users';
  static const String adminSchedules = '/admin/schedules';
  static const String adminLeaves = '/admin/leaves';
}
