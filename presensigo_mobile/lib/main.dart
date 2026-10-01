import 'package:flutter/material.dart';

import 'core/theme/app_theme.dart';
import 'features/auth/screens/login_screen.dart';
import 'features/auth/screens/splash_screen.dart';
import 'features/auth/screens/biometric_unlock_screen.dart';
import 'features/attendance/screens/attendance_screen.dart';
import 'data/services/offline_queue_service.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await OfflineQueueService.instance.initialize();
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
