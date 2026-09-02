import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong2/latlong.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/location_service.dart';
import '../../../data/services/api_service.dart';
import '../../../data/models/attendance_model.dart';

class GeofencingScreen extends StatefulWidget {
  final Function(LatLng position)? onLocationConfirmed;

  const GeofencingScreen({super.key, this.onLocationConfirmed});

  @override
  State<GeofencingScreen> createState() => _GeofencingScreenState();
}

class _GeofencingScreenState extends State<GeofencingScreen> {
  final MapController _mapController = MapController();
  LatLng? _currentPosition;
  LatLng? _officeLocation;
  double _officeRadius = 50;
  double? _distanceToOffice;
  bool _isInsideGeofence = false;
  bool _isLoading = true;
  String _statusMessage = 'Fetching location...';

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    // Get current location
    final position = await LocationService.getCurrentLocation();
    if (position != null) {
      _currentPosition = LatLng(position.latitude, position.longitude);
    }

    // Get office locations from API
    final locations = await ApiService.getLocations();
    if (locations.isNotEmpty) {
      _officeLocation = LatLng(locations[0].latitude, locations[0].longitude);
      _officeRadius = locations[0].radiusMeters.toDouble();
    } else {
      // Default office location (Jakarta)
      _officeLocation = const LatLng(-6.2088, 106.8456);
    }

    // Calculate distance
    if (_currentPosition != null && _officeLocation != null) {
      _distanceToOffice = LocationService.calculateDistance(
        _currentPosition!.latitude,
        _currentPosition!.longitude,
        _officeLocation!.latitude,
        _officeLocation!.longitude,
      );
      _isInsideGeofence = _distanceToOffice! <= _officeRadius;
      _statusMessage = _isInsideGeofence
          ? 'You are inside the geofence'
          : 'You are outside the geofence';
    } else {
      _statusMessage = 'Could not determine location';
    }

    setState(() => _isLoading = false);

    // Center map
    if (_officeLocation != null) {
      Future.delayed(const Duration(milliseconds: 300), () {
        _mapController.fitCamera(
          CameraFit.bounds(
            bounds: LatLngBounds.fromPoints([
              _officeLocation!,
              if (_currentPosition != null) _currentPosition!,
            ]),
            padding: const EdgeInsets.all(100),
          ),
        );
      });
    }
  }

  void _confirmLocation() {
    if (_isInsideGeofence && _currentPosition != null) {
      widget.onLocationConfirmed?.call(_currentPosition!);
      Navigator.pop(context, _currentPosition);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppTheme.backgroundColor,
      appBar: AppBar(
        title: const Text(
          'Location Verification',
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
        actions: [
          Container(
            margin: const EdgeInsets.only(right: 12),
            decoration: BoxDecoration(
              color: AppTheme.primaryColor.withOpacity(0.1),
              borderRadius: BorderRadius.circular(10),
            ),
            child: IconButton(
              icon: const Icon(Icons.my_location, color: AppTheme.primaryColor),
              onPressed: _loadData,
            ),
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : Column(
              children: [
                // Map
                Expanded(
                  flex: 3,
                  child: _buildMap(),
                ),
                // Info panel
                Expanded(
                  flex: 2,
                  child: _buildInfoPanel(),
                ),
              ],
            ),
    );
  }

  Widget _buildMap() {
    if (_officeLocation == null) {
      return const Center(child: Text('No office location found'));
    }

    return FlutterMap(
      mapController: _mapController,
      options: MapOptions(
        initialCenter: _officeLocation!,
        initialZoom: 16,
      ),
      children: [
        // Tile layer
        TileLayer(
          urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
          userAgentPackageName: 'com.presensigo.app',
        ),
        // Circle layer for geofence radius
        CircleLayer(
          circles: [
            CircleMarker(
              point: _officeLocation!,
              radius: _officeRadius,
              useRadiusInMeter: true,
              color: _isInsideGeofence
                  ? AppTheme.secondaryColor.withOpacity(0.2)
                  : AppTheme.errorColor.withOpacity(0.2),
              borderColor: _isInsideGeofence
                  ? AppTheme.secondaryColor
                  : AppTheme.errorColor,
              borderStrokeWidth: 2,
            ),
          ],
        ),
        // Marker layer
        MarkerLayer(
          markers: [
            // Office marker
            Marker(
              point: _officeLocation!,
              width: 50,
              height: 50,
              child: Container(
                decoration: BoxDecoration(
                  color: AppTheme.primaryColor,
                  shape: BoxShape.circle,
                  border: Border.all(color: Colors.white, width: 3),
                  boxShadow: [
                    BoxShadow(
                      color: AppTheme.primaryColor.withOpacity(0.3),
                      blurRadius: 10,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: const Icon(Icons.business, color: Colors.white, size: 24),
              ),
            ),
            // Current position marker
            if (_currentPosition != null)
              Marker(
                point: _currentPosition!,
                width: 50,
                height: 50,
                child: Container(
                  decoration: BoxDecoration(
                    color: _isInsideGeofence ? AppTheme.secondaryColor : AppTheme.errorColor,
                    shape: BoxShape.circle,
                    border: Border.all(color: Colors.white, width: 3),
                    boxShadow: [
                      BoxShadow(
                        color: (_isInsideGeofence ? AppTheme.secondaryColor : AppTheme.errorColor)
                            .withOpacity(0.3),
                        blurRadius: 10,
                        offset: const Offset(0, 4),
                      ),
                    ],
                  ),
                  child: const Icon(Icons.person, color: Colors.white, size: 24),
                ),
              ),
          ],
        ),
      ],
    );
  }

  Widget _buildInfoPanel() {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: const BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.vertical(top: Radius.circular(24)),
        boxShadow: [
          BoxShadow(
            color: Colors.black12,
            blurRadius: 20,
            offset: Offset(0, -5),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Status indicator
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: _isInsideGeofence
                  ? AppTheme.secondaryColor.withOpacity(0.1)
                  : AppTheme.errorColor.withOpacity(0.1),
              borderRadius: BorderRadius.circular(16),
              border: Border.all(
                color: _isInsideGeofence
                    ? AppTheme.secondaryColor.withOpacity(0.3)
                    : AppTheme.errorColor.withOpacity(0.3),
              ),
            ),
            child: Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(10),
                  decoration: BoxDecoration(
                    color: _isInsideGeofence ? AppTheme.secondaryColor : AppTheme.errorColor,
                    shape: BoxShape.circle,
                  ),
                  child: Icon(
                    _isInsideGeofence ? Icons.check : Icons.close,
                    color: Colors.white,
                    size: 24,
                  ),
                ),
                const SizedBox(width: 16),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        _isInsideGeofence ? 'Inside Geofence' : 'Outside Geofence',
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.w600,
                          color: _isInsideGeofence ? AppTheme.secondaryColor : AppTheme.errorColor,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        _statusMessage,
                        style: TextStyle(
                          fontSize: 13,
                          color: AppTheme.textSecondary,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: 16),

          // Distance info
          if (_distanceToOffice != null)
            Row(
              children: [
                _buildInfoItem(
                  Icons.straighten,
                  'Distance',
                  '${_distanceToOffice!.toStringAsFixed(0)}m',
                ),
                const SizedBox(width: 16),
                _buildInfoItem(
                  Icons.radio_button_checked,
                  'Radius',
                  '${_officeRadius.toStringAsFixed(0)}m',
                ),
                const SizedBox(width: 16),
                _buildInfoItem(
                  Icons.circle,
                  'Status',
                  _isInsideGeofence ? 'In Range' : 'Out of Range',
                ),
              ],
            ),

          const Spacer(),

          // Confirm button
          SizedBox(
            width: double.infinity,
            height: 52,
            child: ElevatedButton(
              onPressed: _isInsideGeofence ? _confirmLocation : null,
              style: ElevatedButton.styleFrom(
                backgroundColor: _isInsideGeofence
                    ? AppTheme.secondaryColor
                    : AppTheme.textMuted,
                disabledBackgroundColor: AppTheme.textMuted.withOpacity(0.3),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
                elevation: 0,
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    _isInsideGeofence ? Icons.check_circle : Icons.location_off,
                    color: Colors.white,
                    size: 20,
                  ),
                  const SizedBox(width: 8),
                  Text(
                    _isInsideGeofence ? 'Confirm Location' : 'Move Closer to Office',
                    style: const TextStyle(
                      color: Colors.white,
                      fontSize: 16,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildInfoItem(IconData icon, String label, String value) {
    return Expanded(
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: AppTheme.backgroundColor,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          children: [
            Icon(icon, size: 20, color: AppTheme.primaryColor),
            const SizedBox(height: 6),
            Text(
              label,
              style: TextStyle(
                fontSize: 11,
                color: AppTheme.textMuted,
              ),
            ),
            const SizedBox(height: 2),
            Text(
              value,
              style: const TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.w600,
                color: AppTheme.textPrimary,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
