import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'package:image_picker/image_picker.dart';
import 'package:latlong2/latlong.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/utils/location_service.dart';
import '../../../core/utils/crypto_helper.dart';
import '../../../data/services/api_service.dart';
import '../../history/screens/history_screen.dart';
import '../../settings/screens/settings_screen.dart';
import '../../geofencing/screens/geofencing_screen.dart';

class AttendanceScreen extends StatefulWidget {
  const AttendanceScreen({super.key});

  @override
  State<AttendanceScreen> createState() => _AttendanceScreenState();
}

class _AttendanceScreenState extends State<AttendanceScreen> {
  static const int _maxSelfieBytes = 1024 * 1024;

  Position? _currentPosition;
  bool _isCheckedIn = false;
  bool _isLoading = true;
  bool _isProcessing = false;
  String _deviceUuid = '';

  @override
  void initState() {
    super.initState();
    _loadDeviceUuid();
    _loadData();
  }

  Future<void> _loadDeviceUuid() async {
    final prefs = await SharedPreferences.getInstance();
    setState(() {
      _deviceUuid = prefs.getString('device_uuid') ?? '';
    });
  }

  Future<void> _loadData() async {
    await _getCurrentLocation();
    await _checkTodayAttendance();
    setState(() => _isLoading = false);
  }

  Future<void> _getCurrentLocation() async {
    _currentPosition = await LocationService.getCurrentLocation();
  }

  Future<void> _checkTodayAttendance() async {
    final attendance = await ApiService.getTodayAttendance();
    if (attendance != null && attendance.checkOutTime == null) {
      setState(() => _isCheckedIn = true);
    }
  }

  Future<void> _checkIn() async {
    // Open geofencing screen first
    final confirmedPosition = await Navigator.push<LatLng>(
      context,
      MaterialPageRoute(builder: (_) => const GeofencingScreen()),
    );

    if (confirmedPosition == null) return; // User cancelled

    final challenge = await ApiService.getFaceChallenge();
    if (challenge['success'] != true) {
      _showError(challenge['message'] as String);
      return;
    }
    if (!mounted) return;
    final challengeName = challenge['challenge'] as String;
    await showDialog<void>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Liveness check'),
        content: Text(
          challengeName == 'turn_left'
              ? 'Turn your head slightly to the left, then take the selfie.'
              : 'Turn your head slightly to the right, then take the selfie.',
        ),
        actions: [
          FilledButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Open camera'),
          ),
        ],
      ),
    );
    final selfieData = await _captureSelfie();
    if (selfieData == null) return;

    setState(() => _isProcessing = true);

    final timestamp = CryptoHelper.getCurrentTimestamp();
    final payload = {
      'device_uuid': _deviceUuid,
      'latitude': confirmedPosition.latitude.toString(),
      'longitude': confirmedPosition.longitude.toString(),
      'timestamp': timestamp.toString(),
    };

    final hmac = CryptoHelper.generateHMAC(payload, _deviceUuid);

    final result = await ApiService.checkIn(
      latitude: confirmedPosition.latitude,
      longitude: confirmedPosition.longitude,
      deviceUuid: _deviceUuid,
      timestamp: timestamp,
      hmacSignature: hmac,
      selfieData: selfieData,
      livenessChallenge: challengeName,
      livenessToken: challenge['token'] as String,
    );

    setState(() => _isProcessing = false);

    if (result['success'] == true) {
      setState(() => _isCheckedIn = true);
      _showSuccess('Check-in successful! Have a great day.');
    } else {
      _showError(result['message']);
    }
  }

  Future<String?> _captureSelfie() async {
    while (mounted) {
      final photo = await ImagePicker().pickImage(
        source: ImageSource.camera,
        preferredCameraDevice: CameraDevice.front,
        maxWidth: 500,
        maxHeight: 500,
        imageQuality: 70,
      );
      if (photo == null) return null;

      final bytes = await photo.readAsBytes();
      if (!_isJpegOrPng(bytes)) {
        _showError('Selfie must be a JPEG or PNG image.');
        continue;
      }
      if (bytes.length > _maxSelfieBytes) {
        _showError('Selfie is still larger than 1 MB. Please retake it.');
        continue;
      }
      if (!mounted) return null;

      final accepted = await showDialog<bool>(
        context: context,
        barrierDismissible: false,
        builder: (context) => AlertDialog(
          title: const Text('Use this selfie?'),
          content: ClipRRect(
            borderRadius: BorderRadius.circular(12),
            child: Image.memory(bytes, height: 280, fit: BoxFit.cover),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Retake'),
            ),
            FilledButton(
              onPressed: () => Navigator.pop(context, true),
              child: const Text('Use selfie'),
            ),
          ],
        ),
      );
      if (accepted == true) return base64Encode(bytes);
    }
    return null;
  }

  bool _isJpegOrPng(Uint8List bytes) {
    final isJpeg =
        bytes.length >= 3 &&
        bytes[0] == 0xff &&
        bytes[1] == 0xd8 &&
        bytes[2] == 0xff;
    final isPng =
        bytes.length >= 8 &&
        bytes[0] == 0x89 &&
        bytes[1] == 0x50 &&
        bytes[2] == 0x4e &&
        bytes[3] == 0x47 &&
        bytes[4] == 0x0d &&
        bytes[5] == 0x0a &&
        bytes[6] == 0x1a &&
        bytes[7] == 0x0a;
    return isJpeg || isPng;
  }

  Future<void> _checkOut() async {
    if (_currentPosition == null) {
      _showError('Location not available. Please enable GPS.');
      return;
    }

    setState(() => _isProcessing = true);

    final timestamp = CryptoHelper.getCurrentTimestamp();
    final payload = {
      'device_uuid': _deviceUuid,
      'latitude': _currentPosition!.latitude.toString(),
      'longitude': _currentPosition!.longitude.toString(),
      'timestamp': timestamp.toString(),
    };

    final hmac = CryptoHelper.generateHMAC(payload, _deviceUuid);

    final result = await ApiService.checkOut(
      latitude: _currentPosition!.latitude,
      longitude: _currentPosition!.longitude,
      deviceUuid: _deviceUuid,
      timestamp: timestamp,
      hmacSignature: hmac,
    );

    setState(() => _isProcessing = false);

    if (result['success'] == true) {
      setState(() => _isCheckedIn = false);
      _showSuccess('Check-out successful! See you tomorrow.');
    } else {
      _showError(result['message']);
    }
  }

  void _showError(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(Icons.error_outline, color: Colors.white, size: 20),
            const SizedBox(width: 8),
            Expanded(child: Text(message)),
          ],
        ),
        backgroundColor: AppTheme.errorColor,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      ),
    );
  }

  void _showSuccess(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const Icon(
              Icons.check_circle_outline,
              color: Colors.white,
              size: 20,
            ),
            const SizedBox(width: 8),
            Expanded(child: Text(message)),
          ],
        ),
        backgroundColor: AppTheme.secondaryColor,
        behavior: SnackBarBehavior.floating,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.backgroundColor,
      appBar: AppBar(
        title: const Text(
          'PresensiGo',
          style: TextStyle(fontWeight: FontWeight.w700, letterSpacing: -0.5),
        ),
        actions: [
          Container(
            margin: const EdgeInsets.only(right: 8),
            decoration: BoxDecoration(
              color: AppTheme.primaryColor.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(10),
            ),
            child: IconButton(
              icon: const Icon(
                Icons.history_rounded,
                color: AppTheme.primaryColor,
              ),
              onPressed: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (_) => const HistoryScreen()),
                );
              },
            ),
          ),
          Container(
            margin: const EdgeInsets.only(right: 8),
            decoration: BoxDecoration(
              color: AppTheme.primaryColor.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(10),
            ),
            child: IconButton(
              icon: const Icon(
                Icons.settings_outlined,
                color: AppTheme.primaryColor,
              ),
              onPressed: () {
                Navigator.push(
                  context,
                  MaterialPageRoute(builder: (_) => const SettingsScreen()),
                );
              },
            ),
          ),
          Container(
            margin: const EdgeInsets.only(right: 12),
            decoration: BoxDecoration(
              color: AppTheme.errorColor.withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(10),
            ),
            child: IconButton(
              icon: const Icon(
                Icons.logout_rounded,
                color: AppTheme.errorColor,
              ),
              onPressed: () async {
                await ApiService.logout();
                if (!context.mounted) return;
                Navigator.pushReplacementNamed(context, '/login');
              },
            ),
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: _loadData,
              color: AppTheme.primaryColor,
              child: SingleChildScrollView(
                physics: const AlwaysScrollableScrollPhysics(),
                padding: const EdgeInsets.all(20),
                child: Column(
                  children: [
                    _buildLocationCard(),
                    const SizedBox(height: 20),
                    _buildAttendanceButton(),
                    const SizedBox(height: 20),
                    _buildStatusCards(),
                  ],
                ),
              ),
            ),
    );
  }

  Widget _buildLocationCard() {
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
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: AppTheme.primaryColor.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Icon(
                  Icons.location_on_rounded,
                  color: AppTheme.primaryColor,
                  size: 24,
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Current Location',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                        color: AppTheme.textPrimary,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      _currentPosition != null
                          ? '${_currentPosition!.latitude.toStringAsFixed(4)}, ${_currentPosition!.longitude.toStringAsFixed(4)}'
                          : 'Fetching location...',
                      style: TextStyle(
                        fontSize: 13,
                        color: _currentPosition != null
                            ? AppTheme.textSecondary
                            : AppTheme.warningColor,
                      ),
                    ),
                  ],
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 10,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: _currentPosition != null
                      ? AppTheme.secondaryColor.withValues(alpha: 0.1)
                      : AppTheme.warningColor.withValues(alpha: 0.1),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Container(
                      width: 8,
                      height: 8,
                      decoration: BoxDecoration(
                        color: _currentPosition != null
                            ? AppTheme.secondaryColor
                            : AppTheme.warningColor,
                        shape: BoxShape.circle,
                      ),
                    ),
                    const SizedBox(width: 6),
                    Text(
                      _currentPosition != null ? 'Active' : 'Waiting',
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.w500,
                        color: _currentPosition != null
                            ? AppTheme.secondaryColor
                            : AppTheme.warningColor,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildAttendanceButton() {
    return GestureDetector(
      onTap: _isProcessing ? null : (_isCheckedIn ? _checkOut : _checkIn),
      child: Container(
        width: 220,
        height: 220,
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          gradient: _isCheckedIn
              ? const LinearGradient(
                  colors: [Color(0xFFEF4444), Color(0xFFDC2626)],
                  begin: Alignment.topLeft,
                  end: Alignment.bottomRight,
                )
              : AppGradients.primaryGradient,
          boxShadow: [
            BoxShadow(
              color:
                  (_isCheckedIn ? AppTheme.errorColor : AppTheme.primaryColor)
                      .withValues(alpha: 0.4),
              blurRadius: 24,
              offset: const Offset(0, 8),
            ),
          ],
        ),
        child: Container(
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            border: Border.all(
              color: Colors.white.withValues(alpha: 0.2),
              width: 3,
            ),
          ),
          child: _isProcessing
              ? const Center(
                  child: CircularProgressIndicator(
                    color: Colors.white,
                    strokeWidth: 3,
                  ),
                )
              : Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(
                      _isCheckedIn ? Icons.logout_rounded : Icons.login_rounded,
                      size: 56,
                      color: Colors.white,
                    ),
                    const SizedBox(height: 8),
                    Text(
                      _isCheckedIn ? 'Check Out' : 'Check In',
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 18,
                        fontWeight: FontWeight.w700,
                        letterSpacing: -0.5,
                      ),
                    ),
                  ],
                ),
        ),
      ),
    );
  }

  Widget _buildStatusCards() {
    return Row(
      children: [
        Expanded(
          child: Container(
            padding: const EdgeInsets.all(16),
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
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color:
                            (_isCheckedIn
                                    ? AppTheme.secondaryColor
                                    : AppTheme.textMuted)
                                .withValues(alpha: 0.1),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Icon(
                        Icons.circle,
                        size: 12,
                        color: _isCheckedIn
                            ? AppTheme.secondaryColor
                            : AppTheme.textMuted,
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      'Status',
                      style: TextStyle(
                        fontSize: 13,
                        color: AppTheme.textSecondary,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Text(
                  _isCheckedIn ? 'Checked In' : 'Not Checked In',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: _isCheckedIn
                        ? AppTheme.secondaryColor
                        : AppTheme.textPrimary,
                  ),
                ),
              ],
            ),
          ),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Container(
            padding: const EdgeInsets.all(16),
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
                      padding: const EdgeInsets.all(8),
                      decoration: BoxDecoration(
                        color: AppTheme.primaryColor.withValues(alpha: 0.1),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: const Icon(
                        Icons.access_time_rounded,
                        size: 16,
                        color: AppTheme.primaryColor,
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      'Time',
                      style: TextStyle(
                        fontSize: 13,
                        color: AppTheme.textSecondary,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Text(
                  TimeOfDay.now().format(context),
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                    color: AppTheme.textPrimary,
                  ),
                ),
              ],
            ),
          ),
        ),
      ],
    );
  }
}
