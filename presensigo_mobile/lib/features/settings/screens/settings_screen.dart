import 'package:flutter/material.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/biometric_service.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  bool _biometricEnabled = false;
  bool _biometricAvailable = false;
  String _biometricName = 'Biometric';
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadSettings();
  }

  Future<void> _loadSettings() async {
    final available = await BiometricService.isAvailable();
    final enabled = await BiometricService.isEnabled();
    final biometrics = await BiometricService.getAvailableBiometrics();
    final name = BiometricService.getBiometricName(biometrics);

    setState(() {
      _biometricAvailable = available;
      _biometricEnabled = enabled;
      _biometricName = name;
      _isLoading = false;
    });
  }

  Future<void> _toggleBiometric(bool value) async {
    if (value) {
      // Enable biometric - require authentication first
      final authenticated = await BiometricService.authenticate(
        reason: 'Authenticate to enable $_biometricName login',
      );

      if (!authenticated) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Row(
                children: [
                  const Icon(Icons.error_outline, color: Colors.white, size: 20),
                  const SizedBox(width: 8),
                  const Text('Authentication failed'),
                ],
              ),
              backgroundColor: AppTheme.errorColor,
              behavior: SnackBarBehavior.floating,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            ),
          );
        }
        return;
      }
    }

    await BiometricService.setEnabled(value);
    setState(() => _biometricEnabled = value);

    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Row(
            children: [
              Icon(
                value ? Icons.check_circle_outline : Icons.info_outline,
                color: Colors.white,
                size: 20,
              ),
              const SizedBox(width: 8),
              Text(value ? '$_biometricName enabled' : '$_biometricName disabled'),
            ],
          ),
          backgroundColor: value ? AppTheme.secondaryColor : AppTheme.textSecondary,
          behavior: SnackBarBehavior.floating,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.backgroundColor,
      appBar: AppBar(
        title: const Text(
          'Settings',
          style: TextStyle(fontWeight: FontWeight.w700, letterSpacing: -0.5),
        ),
        leading: IconButton(
          icon: Container(
            padding: const EdgeInsets.all(8),
            decoration: BoxDecoration(
              color: AppTheme.primaryColor.withOpacity(0.1),
              borderRadius: BorderRadius.circular(10),
            ),
            child: const Icon(Icons.arrow_back_ios_new, size: 18, color: AppTheme.primaryColor),
          ),
          onPressed: () => Navigator.pop(context),
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : SingleChildScrollView(
              padding: const EdgeInsets.all(20),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Security Section
                  _buildSectionHeader('Security'),
                  const SizedBox(height: 12),
                  _buildBiometricCard(),
                  const SizedBox(height: 24),

                  // About Section
                  _buildSectionHeader('About'),
                  const SizedBox(height: 12),
                  _buildAboutCard(),
                ],
              ),
            ),
    );
  }

  Widget _buildSectionHeader(String title) {
    return Text(
      title,
      style: TextStyle(
        fontSize: 13,
        fontWeight: FontWeight.w600,
        color: AppTheme.textSecondary,
        letterSpacing: 0.5,
      ),
    );
  }

  Widget _buildBiometricCard() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppTheme.borderColor),
        boxShadow: AppShadows.small,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: _biometricAvailable
                      ? AppTheme.primaryColor.withOpacity(0.1)
                      : AppTheme.textMuted.withOpacity(0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Icon(
                  _biometricName == 'Face ID' ? Icons.face_rounded : Icons.fingerprint,
                  size: 24,
                  color: _biometricAvailable ? AppTheme.primaryColor : AppTheme.textMuted,
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      '$_biometricName Login',
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                        color: AppTheme.textPrimary,
                      ),
                    ),
                    const SizedBox(height: 4),
                    Text(
                      _biometricAvailable
                          ? 'Use $_biometricName to sign in quickly'
                          : '$_biometricName not available on this device',
                      style: TextStyle(
                        fontSize: 13,
                        color: _biometricAvailable ? AppTheme.textSecondary : AppTheme.textMuted,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
          if (_biometricAvailable) ...[
            const SizedBox(height: 16),
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: _biometricEnabled
                    ? AppTheme.secondaryColor.withOpacity(0.05)
                    : AppTheme.backgroundColor,
                borderRadius: BorderRadius.circular(12),
                border: Border.all(
                  color: _biometricEnabled
                      ? AppTheme.secondaryColor.withOpacity(0.2)
                      : AppTheme.borderColor,
                ),
              ),
              child: Row(
                children: [
                  Icon(
                    _biometricEnabled ? Icons.check_circle : Icons.radio_button_unchecked,
                    size: 20,
                    color: _biometricEnabled ? AppTheme.secondaryColor : AppTheme.textMuted,
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          _biometricEnabled ? 'Enabled' : 'Disabled',
                          style: TextStyle(
                            fontSize: 14,
                            fontWeight: FontWeight.w500,
                            color: _biometricEnabled
                                ? AppTheme.secondaryColor
                                : AppTheme.textPrimary,
                          ),
                        ),
                        Text(
                          _biometricEnabled
                              ? 'Tap to disable $_biometricName login'
                              : 'Tap to enable $_biometricName login',
                          style: TextStyle(
                            fontSize: 12,
                            color: AppTheme.textSecondary,
                          ),
                        ),
                      ],
                    ),
                  ),
                  Switch(
                    value: _biometricEnabled,
                    onChanged: _toggleBiometric,
                    activeColor: AppTheme.secondaryColor,
                  ),
                ],
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildAboutCard() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: AppTheme.borderColor),
        boxShadow: AppShadows.small,
      ),
      child: Column(
        children: [
          _buildAboutRow(Icons.business_rounded, 'PresensiGo', 'Smart Attendance System'),
          const Divider(height: 24),
          _buildAboutRow(Icons.code_rounded, 'Version', '1.0.0+1'),
          const Divider(height: 24),
          _buildAboutRow(Icons.storage_rounded, 'Backend', 'Go + PostgreSQL'),
        ],
      ),
    );
  }

  Widget _buildAboutRow(IconData icon, String label, String value) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: AppTheme.primaryColor.withOpacity(0.1),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(icon, size: 18, color: AppTheme.primaryColor),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Text(
            label,
            style: const TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.w500,
              color: AppTheme.textPrimary,
            ),
          ),
        ),
        Text(
          value,
          style: TextStyle(
            fontSize: 14,
            color: AppTheme.textSecondary,
          ),
        ),
      ],
    );
  }
}
