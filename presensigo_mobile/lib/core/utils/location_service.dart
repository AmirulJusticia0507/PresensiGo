import 'package:geolocator/geolocator.dart';

class LocationService {
  static Future<bool> checkPermission() async {
    bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
    if (!serviceEnabled) {
      return false;
    }

    LocationPermission permission = await Geolocator.checkPermission();
    if (permission == LocationPermission.denied) {
      permission = await Geolocator.requestPermission();
      if (permission == LocationPermission.denied) {
        return false;
      }
    }

    if (permission == LocationPermission.deniedForever) {
      return false;
    }

    return true;
  }

  static Future<Position?> getCurrentLocation() async {
    final hasPermission = await checkPermission();
    if (!hasPermission) return null;

    return await Geolocator.getCurrentPosition(
      desiredAccuracy: LocationAccuracy.high,
    );
  }

  /// Whether [position] came from a mock/simulated location provider.
  ///
  /// Android reports the mock-provider flag from the platform. On iOS 15 and
  /// newer this reflects `CLLocation.isSimulatedBySoftware`. Older platforms
  /// always report false, so treat it as a best-effort signal only.
  static bool isMockPosition(Position position) => position.isMocked;

  /// Signal strength below this many meters is too coarse to trust for a
  /// geofenced check-in.
  static const double minimumAccuracyMeters = 50;

  /// Whether [position] is accurate enough to record a check-in.
  static bool isAcceptableAccuracy(Position position) =>
      position.accuracy > 0 && position.accuracy <= minimumAccuracyMeters;

  /// Returns a human readable reason to refuse a check-in, or null when the
  /// fix looks trustworthy.
  static String? rejectionReason(Position position) {
    if (isMockPosition(position)) {
      return 'Mock location is enabled. Turn off your mock location app and try again.';
    }
    if (!isAcceptableAccuracy(position)) {
      return 'GPS accuracy is too low (${position.accuracy.toStringAsFixed(0)} m). '
          'Move to an open area and try again.';
    }
    return null;
  }

  static double calculateDistance(
    double lat1,
    double lon1,
    double lat2,
    double lon2,
  ) {
    return Geolocator.distanceBetween(lat1, lon1, lat2, lon2);
  }
}
