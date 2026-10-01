import 'package:flutter/material.dart';
import 'core/theme/app_theme.dart';
import 'features/auth/screens/login_screen.dart';
import 'features/auth/screens/splash_screen.dart';
import 'features/auth/screens/biometric_unlock_screen.dart';
import 'features/attendance/screens/attendance_screen.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();
  runApp(const PresensiGoApp());
}

class PresensiGoApp extends StatelessWidget {
  const PresensiGoApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'PresensiGo',
      debugShowCheckedModeBanner: false,
      theme: AppTheme.lightTheme,
      initialRoute: '/splash',
      routes: {
        '/splash': (context) => const SplashScreen(),
        '/login': (context) => const LoginScreen(),
        '/attendance': (context) => const AttendanceScreen(),
        '/biometric-unlock': (context) => const BiometricUnlockScreen(),
      },
    );
  }
}
