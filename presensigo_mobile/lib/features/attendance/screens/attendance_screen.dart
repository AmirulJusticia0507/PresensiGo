import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import 'package:latlong2/latlong.dart';
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

class _AttendanceScreenState extends State<AttendanceScreen> with SingleTickerProviderStateMixin {
  Position? _currentPosition;
  bool _isCheckedIn = false;
  bool _isLoading = true;
  bool _isProcessing = false;
  late AnimationController _animController;
  late Animation<double> _pulseAnimation;

  @override
  void initState() {
    super.initState();
    _animController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    )..repeat(reverse: true);
    _pulseAnimation = Tween<double>(begin: 1.0, end: 1.05).animate(
      CurvedAnimation(parent: _animController, curve: Curves.easeInOut),
    );
    _loadData();
  }

  @override
  void dispose() {
    _animController.dispose();
    super.dispose();
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

    setState(() => _isProcessing = true);

    final payload = {
      'latitude': confirmedPosition.latitude,
      'longitude': confirmedPosition.longitude,
      'device_uuid': 'device-123',
    };

    final hmac = CryptoHelper.generateHMAC(payload, 'your-secret-key');

    final result = await ApiService.checkIn(
      latitude: confirmedPosition.latitude,
      longitude: confirmedPosition.longitude,
      deviceUuid: 'device-123',
      hmacSignature: hmac,
    );

    setState(() => _isProcessing = false);

    if (result['success'] == true) {
      setState(() => _isCheckedIn = true);
      _showSuccess('Check-in successful! Have a great day.');
    } else {
      _showError(result['message']);
    }
  }

  Future<void> _checkOut() async {
    if (_currentPosition == null) {
      _showError('Location not available. Please enable GPS.');
      return;
    }

    setState(() => _isProcessing = true);

    final payload = {
      'latitude': _currentPosition!.latitude,
      'longitude': _currentPosition!.longitude,
      'device_uuid': 'device-123',
    };

    final hmac = CryptoHelper.generateHMAC(payload, 'your-secret-key');

    final result = await ApiService.checkOut(
      latitude: _currentPosition!.latitude,
      longitude: _currentPosition!.longitude,
      deviceUuid: 'device-123',
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
            const Icon(Icons.check_circle_outline, color: Colors.white, size: 20),
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
              color: AppTheme.primaryColor.withOpacity(0.1),
              borderRadius: BorderRadius.circular(10),
            ),
            child: IconButton(
              icon: const Icon(Icons.history_rounded, color: AppTheme.primaryColor),
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
              color: AppTheme.primaryColor.withOpacity(0.1),
              borderRadius: BorderRadius.circular(10),
            ),
            child: IconButton(
              icon: const Icon(Icons.settings_outlined, color: AppTheme.primaryColor),
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
              color: AppTheme.errorColor.withOpacity(0.1),
              borderRadius: BorderRadius.circular(10),
            ),
            child: IconButton(
              icon: const Icon(Icons.logout_rounded, color: AppTheme.errorColor),
              onPressed: () async {
                await ApiService.logout();
                if (mounted) {
                  Navigator.pushReplacementNamed(context, '/login');
                }
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
                  color: AppTheme.primaryColor.withOpacity(0.1),
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
                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                decoration: BoxDecoration(
                  color: _currentPosition != null
                      ? AppTheme.secondaryColor.withOpacity(0.1)
                      : AppTheme.warningColor.withOpacity(0.1),
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
                color: (_isCheckedIn ? AppTheme.errorColor : AppTheme.primaryColor)
                    .withOpacity(0.4),
                blurRadius: 24,
                offset: const Offset(0, 8),
              ),
            ],
          ),
          child: Container(
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              border: Border.all(color: Colors.white.withOpacity(0.2), width: 3),
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
                        color: (_isCheckedIn ? AppTheme.secondaryColor : AppTheme.textMuted)
                            .withOpacity(0.1),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Icon(
                        Icons.circle,
                        size: 12,
                        color: _isCheckedIn ? AppTheme.secondaryColor : AppTheme.textMuted,
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
                    color: _isCheckedIn ? AppTheme.secondaryColor : AppTheme.textPrimary,
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
                        color: AppTheme.primaryColor.withOpacity(0.1),
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
