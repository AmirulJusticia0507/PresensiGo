import 'package:flutter/material.dart';
import 'package:geolocator/geolocator.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/location_service.dart';
import '../../../core/utils/crypto_helper.dart';
import '../../../data/services/api_service.dart';
import '../../history/screens/history_screen.dart';

class AttendanceScreen extends StatefulWidget {
  const AttendanceScreen({super.key});

  @override
  State<AttendanceScreen> createState() => _AttendanceScreenState();
}

class _AttendanceScreenState extends State<AttendanceScreen> {
  Position? _currentPosition;
  bool _isCheckedIn = false;
  bool _isLoading = true;
  bool _isProcessing = false;

  @override
  void initState() {
    super.initState();
    _loadData();
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
    if (_currentPosition == null) {
      _showError('Location not available');
      return;
    }

    setState(() => _isProcessing = true);

    final payload = {
      'latitude': _currentPosition!.latitude,
      'longitude': _currentPosition!.longitude,
      'device_uuid': 'device-123',
    };

    final hmac = CryptoHelper.generateHMAC(payload, 'your-secret-key');

    final result = await ApiService.checkIn(
      latitude: _currentPosition!.latitude,
      longitude: _currentPosition!.longitude,
      deviceUuid: 'device-123',
      hmacSignature: hmac,
    );

    setState(() => _isProcessing = false);

    if (result['success'] == true) {
      setState(() => _isCheckedIn = true);
      _showSuccess('Check-in successful!');
    } else {
      _showError(result['message']);
    }
  }

  Future<void> _checkOut() async {
    if (_currentPosition == null) {
      _showError('Location not available');
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
      _showSuccess('Check-out successful!');
    } else {
      _showError(result['message']);
    }
  }

  void _showError(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message), backgroundColor: AppTheme.errorColor),
    );
  }

  void _showSuccess(String message) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(message), backgroundColor: AppTheme.secondaryColor),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('PresensiGo'),
        actions: [
          IconButton(
            icon: const Icon(Icons.history),
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (_) => const HistoryScreen()),
              );
            },
          ),
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              await ApiService.logout();
              if (mounted) {
                Navigator.pushReplacementNamed(context, '/login');
              }
            },
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: _loadData,
              child: SingleChildScrollView(
                physics: const AlwaysScrollableScrollPhysics(),
                padding: const EdgeInsets.all(24),
                child: Column(
                  children: [
                    _buildLocationCard(),
                    const SizedBox(height: 24),
                    _buildAttendanceButton(),
                    const SizedBox(height: 24),
                    _buildStatusCard(),
                  ],
                ),
              ),
            ),
    );
  }

  Widget _buildLocationCard() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            const Icon(Icons.location_on, size: 48, color: AppTheme.primaryColor),
            const SizedBox(height: 12),
            Text(
              _currentPosition != null
                  ? 'Lat: ${_currentPosition!.latitude.toStringAsFixed(6)}\nLng: ${_currentPosition!.longitude.toStringAsFixed(6)}'
                  : 'Location not available',
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.bodyLarge,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildAttendanceButton() {
    return SizedBox(
      width: 200,
      height: 200,
      child: ElevatedButton(
        onPressed: _isProcessing ? null : (_isCheckedIn ? _checkOut : _checkIn),
        style: ElevatedButton.styleFrom(
          backgroundColor: _isCheckedIn ? AppTheme.errorColor : AppTheme.secondaryColor,
          shape: const CircleBorder(),
          padding: EdgeInsets.zero,
        ),
        child: _isProcessing
            ? const CircularProgressIndicator(color: Colors.white)
            : Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    _isCheckedIn ? Icons.logout : Icons.login,
                    size: 64,
                    color: Colors.white,
                  ),
                  const SizedBox(height: 8),
                  Text(
                    _isCheckedIn ? 'Check Out' : 'Check In',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ],
              ),
      ),
    );
  }

  Widget _buildStatusCard() {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            Text(
              'Today\'s Status',
              style: Theme.of(context).textTheme.titleLarge?.copyWith(
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 12),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                const Text('Status:'),
                Chip(
                  label: Text(_isCheckedIn ? 'Checked In' : 'Not Checked In'),
                  backgroundColor: _isCheckedIn
                      ? AppTheme.secondaryColor.withOpacity(0.1)
                      : AppTheme.errorColor.withOpacity(0.1),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
