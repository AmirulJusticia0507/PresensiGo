import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../../data/services/session_manager.dart';

class SplashScreen extends StatefulWidget {
  const SplashScreen({super.key});

  @override
  State<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends State<SplashScreen> {
  late final SessionManager _sessionManager;

  @override
  void initState() {
    super.initState();
    _sessionManager = SessionManager();
    _checkSession();
  }

  Future<void> _checkSession() async {
    // Wait a bit for app to fully initialize
    await Future.delayed(const Duration(milliseconds: 500));

    try {
      if (await _sessionManager.isSessionValid()) {
        // Session valid, check if biometric should be used
        final prefs = await SharedPreferences.getInstance();
        final useBiometric = prefs.getBool('use_biometric') ?? false;

        if (useBiometric) {
          // Navigate to biometric unlock screen
          if (mounted) {
            Navigator.of(context).pushReplacementNamed('/biometric-unlock');
          }
        } else {
          // Navigate to last screen (attendance)
          if (mounted) {
            Navigator.of(context).pushReplacementNamed('/attendance');
          }
        }
      } else {
        // No valid session, go to login
        if (mounted) {
          Navigator.of(context).pushReplacementNamed('/login');
        }
      }
    } catch (e) {
      // On error, redirect to login for security
      if (mounted) {
        Navigator.of(context).pushReplacementNamed('/login');
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            // App logo or splash image
            const Icon(Icons.badge, size: 80, color: Colors.blue),
            const SizedBox(height: 24),
            const Text(
              'PresensiGo',
              style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 24),
            const CircularProgressIndicator(),
            const SizedBox(height: 16),
            const Text('Loading...', style: TextStyle(fontSize: 12)),
          ],
        ),
      ),
    );
  }
}
