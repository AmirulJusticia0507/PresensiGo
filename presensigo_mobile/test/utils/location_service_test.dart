import 'package:flutter_test/flutter_test.dart';
import 'package:geolocator/geolocator.dart';
import 'package:presensigo_app/core/utils/location_service.dart';

Position buildPosition({
  required double latitude,
  required double longitude,
  required double accuracy,
  bool isMocked = false,
}) {
  return Position(
    latitude: latitude,
    longitude: longitude,
    timestamp: DateTime.now(),
    accuracy: accuracy,
    altitude: 0,
    altitudeAccuracy: 0,
    heading: 0,
    headingAccuracy: 0,
    speed: 0,
    speedAccuracy: 0,
    isMocked: isMocked,
  );
}

void main() {
  group('LocationService.isMockPosition', () {
    test('flags a mocked position', () {
      final position = buildPosition(
        latitude: -6.2088,
        longitude: 106.8456,
        accuracy: 5,
        isMocked: true,
      );

      expect(LocationService.isMockPosition(position), isTrue);
    });

    test('accepts a genuine position', () {
      final position = buildPosition(
        latitude: -6.2088,
        longitude: 106.8456,
        accuracy: 5,
      );

      expect(LocationService.isMockPosition(position), isFalse);
    });
  });

  group('LocationService.isAcceptableAccuracy', () {
    test('rejects a non-positive accuracy', () {
      final position = buildPosition(
        latitude: -6.2088,
        longitude: 106.8456,
        accuracy: 0,
      );

      expect(LocationService.isAcceptableAccuracy(position), isFalse);
    });

    test('rejects accuracy beyond the threshold', () {
      final position = buildPosition(
        latitude: -6.2088,
        longitude: 106.8456,
        accuracy: LocationService.minimumAccuracyMeters + 1,
      );

      expect(LocationService.isAcceptableAccuracy(position), isFalse);
    });

    test('accepts accuracy at the threshold', () {
      final position = buildPosition(
        latitude: -6.2088,
        longitude: 106.8456,
        accuracy: LocationService.minimumAccuracyMeters,
      );

      expect(LocationService.isAcceptableAccuracy(position), isTrue);
    });
  });

  group('LocationService.rejectionReason', () {
    test('explains how to disable mock location', () {
      final position = buildPosition(
        latitude: -6.2088,
        longitude: 106.8456,
        accuracy: 5,
        isMocked: true,
      );

      final reason = LocationService.rejectionReason(position);

      expect(reason, isNotNull);
      expect(reason, contains('Mock location'));
    });

    test('reports low accuracy with the measured value', () {
      final position = buildPosition(
        latitude: -6.2088,
        longitude: 106.8456,
        accuracy: 120,
      );

      final reason = LocationService.rejectionReason(position);

      expect(reason, isNotNull);
      expect(reason, contains('120 m'));
    });

    test('accepts a trustworthy fix', () {
      final position = buildPosition(
        latitude: -6.2088,
        longitude: 106.8456,
        accuracy: 8,
      );

      expect(LocationService.rejectionReason(position), isNull);
    });
  });

  group('LocationService.rejectionCode', () {
    test('returns stable telemetry codes', () {
      expect(
        LocationService.rejectionCode(
          buildPosition(latitude: 0, longitude: 0, accuracy: 5, isMocked: true),
        ),
        'mock_location',
      );
      expect(
        LocationService.rejectionCode(
          buildPosition(latitude: 0, longitude: 0, accuracy: 120),
        ),
        'low_accuracy',
      );
      expect(
        LocationService.rejectionCode(
          buildPosition(latitude: 0, longitude: 0, accuracy: 8),
        ),
        isNull,
      );
    });
  });
}
