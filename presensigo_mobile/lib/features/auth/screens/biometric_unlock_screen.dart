import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/biometric_service.dart';
import '../../../data/services/session_manager.dart';

class BiometricUnlockScreen extends StatefulWidget {
  const BiometricUnlockScreen({super.key});

  @override
  State<BiometricUnlockScreen> createState() => _BiometricUnlockScreenState();
}

class _BiometricUnlockScreenState extends State<BiometricUnlockScreen> {
  late final SessionManager _sessionManager;
  int _retryCount = 0;
  static const _maxRetries = 3;
  String _biometricName = 'Biometric';
  bool _isAuthenticating = true;

  @override
  void initState() {
    super.initState();
    _sessionManager = SessionManager();
    _loadBiometricName();
    _startBiometricPrompt();
  }

  Future<void> _loadBiometricName() async {
    try {
      final biometrics = await BiometricService.getAvailableBiometrics();
      final name = BiometricService.getBiometricName(biometrics);
      setState(() {
        _biometricName = name;
      });
    } catch (e) {
      // Use default name if error occurs
    }
  }

  Future<void> _startBiometricPrompt() async {
    setState(() => _isAuthenticating = true);

    try {
      final isAuthenticated = await BiometricService.authenticate(
        reason: 'Authenticate to unlock PresensiGo',
      );

      if (isAuthenticated) {
        // Verify token is still valid
        final isValid = await _sessionManager.isSessionValid();
        if (isValid) {
          // Proceed to attendance
          if (mounted) {
            Navigator.of(context).pushReplacementNamed('/attendance');
          }
        } else {
          // Token expired, redirect to login
          if (mounted) {
            _showSessionExpiredDialog();
          }
        }
      } else {
        // Biometric failed, increment retry
        _retryCount++;

        if (_retryCount >= _maxRetries) {
          // Max retries reached, fallback to password login
          if (mounted) {
            Navigator.of(context).pushReplacementNamed('/login');
          }
        } else {
          // Allow retry
          if (mounted) {
            setState(() => _isAuthenticating = false);
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(
                content: Text(
                    'Biometric failed. Retries: $_retryCount/$_maxRetries'),
                duration: const Duration(seconds: 2),
              ),
            );
            await Future.delayed(const Duration(seconds: 1));
            _startBiometricPrompt();
          }
        }
      }
    } catch (e) {
      // Biometric error (sensor offline, etc.)
      if (mounted) {
        setState(() => _isAuthenticating = false);
        _showBiometricErrorDialog();
      }
    }
  }

  void _showSessionExpiredDialog() {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => AlertDialog(
        title: const Text('Session Expired'),
        content: const Text('Your session has expired. Please log in again.'),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.of(context).pushReplacementNamed('/login');
            },
            child: const Text('OK'),
          ),
        ],
      ),
    );
  }

  void _showBiometricErrorDialog() {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Biometric Error'),
        content: const Text(
            'Biometric authentication failed. Please use password login.'),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.of(context).pushReplacementNamed('/login');
            },
            child: const Text('Login with Password'),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Unlock with Biometric'),
        centerTitle: true,
        elevation: 0,
      ),
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            // Biometric icon
            Container(
              padding: const EdgeInsets.all(24),
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: AppTheme.primaryColor.withValues(alpha: 0.1),
              ),
              child: Icon(
                _biometricName == 'Face ID'
                    ? Icons.face_rounded
                    : Icons.fingerprint,
                size: 80,
                color: AppTheme.primaryColor,
              ),
            ),
            const SizedBox(height: 32),
            Text(
              'Unlocking with $_biometricName',
              style: const TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.w600,
              ),
            ),
            const SizedBox(height: 16),
            if (_isAuthenticating) ...[
              const SizedBox(height: 8),
              const Text(
                'Authenticating...',
                style: TextStyle(
                  fontSize: 14,
                  color: Colors.grey,
                ),
              ),
              const SizedBox(height: 24),
              const CircularProgressIndicator(),
            ] else ...[
              Text(
                'Attempt $_retryCount of $_maxRetries',
                style: TextStyle(
                  fontSize: 12,
                  color: Colors.grey[600],
                ),
              ),
              const SizedBox(height: 32),
              ElevatedButton.icon(
                onPressed: _startBiometricPrompt,
                icon: const Icon(Icons.refresh),
                label: const Text('Try Again'),
              ),
            ],
            const SizedBox(height: 48),
            TextButton.icon(
              onPressed: () {
                Navigator.of(context).pushReplacementNamed('/login');
              },
              icon: const Icon(Icons.password),
              label: const Text('Use Password Login'),
            ),
          ],
        ),
      ),
    );
  }
}
