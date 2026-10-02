import 'package:flutter_test/flutter_test.dart';
import 'package:presensigo_app/data/models/profile_model.dart';

void main() {
  group('ProfileModel', () {
    group('fromJson', () {
      test('creates ProfileModel from valid JSON', () {
        final json = {
          'id': 'user123',
          'name': 'John Doe',
          'email': 'john@example.com',
          'phone': '+12025551234',
          'emergency_contact_name': 'Jane Doe',
          'emergency_contact_phone': '+12025559999',
          'address': '123 Main St',
          'profile_picture_url': null,
          'face_enrolled': true,
          'face_enrolled_at': '2024-01-15T10:30:00Z',
          'created_at': '2024-01-01T00:00:00Z',
          'updated_at': '2024-01-15T10:30:00Z',
        };

        final profile = ProfileModel.fromJson(json);

        expect(profile.id, 'user123');
        expect(profile.name, 'John Doe');
        expect(profile.email, 'john@example.com');
        expect(profile.phone, '+12025551234');
        expect(profile.emergencyContactName, 'Jane Doe');
        expect(profile.emergencyContactPhone, '+12025559999');
        expect(profile.address, '123 Main St');
        expect(profile.faceEnrolled, true);
        expect(profile.faceEnrolledAt, isNotNull);
      });

      test('handles missing optional fields', () {
        final json = {
          'id': 'user123',
          'name': 'John Doe',
          'email': 'john@example.com',
          'phone': null,
          'emergency_contact_name': null,
          'emergency_contact_phone': null,
          'address': null,
          'profile_picture_url': null,
          'face_enrolled': false,
          'face_enrolled_at': null,
          'created_at': '2024-01-01T00:00:00Z',
          'updated_at': '2024-01-15T10:30:00Z',
        };

        final profile = ProfileModel.fromJson(json);

        expect(profile.phone, null);
        expect(profile.emergencyContactName, null);
        expect(profile.emergencyContactPhone, null);
        expect(profile.address, null);
        expect(profile.faceEnrolled, false);
        expect(profile.faceEnrolledAt, null);
      });
    });

    group('toJson', () {
      test('converts ProfileModel to JSON', () {
        final profile = ProfileModel(
          id: 'user123',
          name: 'John Doe',
          email: 'john@example.com',
          phone: '+12025551234',
          emergencyContactName: 'Jane Doe',
          emergencyContactPhone: '+12025559999',
          address: '123 Main St',
          profilePictureUrl: null,
          faceEnrolled: true,
          faceEnrolledAt: DateTime.parse('2024-01-15T10:30:00Z'),
          createdAt: DateTime.parse('2024-01-01T00:00:00Z'),
          updatedAt: DateTime.parse('2024-01-15T10:30:00Z'),
        );

        final json = profile.toJson();

        expect(json['id'], 'user123');
        expect(json['name'], 'John Doe');
        expect(json['email'], 'john@example.com');
        expect(json['phone'], '+12025551234');
        expect(json['emergency_contact_name'], 'Jane Doe');
        expect(json['emergency_contact_phone'], '+12025559999');
        expect(json['address'], '123 Main St');
        expect(json['face_enrolled'], true);
      });
    });

    group('copyWith', () {
      test('creates copy with updated fields', () {
        final original = ProfileModel(
          id: 'user123',
          name: 'John Doe',
          email: 'john@example.com',
          phone: '+12025551234',
          emergencyContactName: 'Jane Doe',
          emergencyContactPhone: '+12025559999',
          address: '123 Main St',
          profilePictureUrl: null,
          faceEnrolled: false,
          faceEnrolledAt: null,
          createdAt: DateTime.parse('2024-01-01T00:00:00Z'),
          updatedAt: DateTime.parse('2024-01-15T10:30:00Z'),
        );

        final updated = original.copyWith(
          name: 'Jane Doe',
          phone: '+12025559999',
        );

        expect(updated.name, 'Jane Doe');
        expect(updated.phone, '+12025559999');
        expect(updated.email, 'john@example.com'); // Unchanged
        expect(updated.emergencyContactName, 'Jane Doe'); // Unchanged
      });

      test('preserves all fields when none provided', () {
        final profile = ProfileModel(
          id: 'user123',
          name: 'John Doe',
          email: 'john@example.com',
          phone: null,
          emergencyContactName: null,
          emergencyContactPhone: null,
          address: null,
          profilePictureUrl: null,
          faceEnrolled: false,
          faceEnrolledAt: null,
          createdAt: DateTime.parse('2024-01-01T00:00:00Z'),
          updatedAt: DateTime.parse('2024-01-15T10:30:00Z'),
        );

        final copy = profile.copyWith();

        expect(copy.id, profile.id);
        expect(copy.name, profile.name);
        expect(copy.email, profile.email);
        expect(copy.phone, profile.phone);
      });
    });

    group('JSON round-trip', () {
      test('survives to JSON and back', () {
        final original = ProfileModel(
          id: 'user123',
          name: 'John Doe',
          email: 'john@example.com',
          phone: '+12025551234',
          emergencyContactName: 'Jane Doe',
          emergencyContactPhone: '+12025559999',
          address: '123 Main St',
          profilePictureUrl: null,
          faceEnrolled: true,
          faceEnrolledAt: DateTime.parse('2024-01-15T10:30:00Z'),
          createdAt: DateTime.parse('2024-01-01T00:00:00Z'),
          updatedAt: DateTime.parse('2024-01-15T10:30:00Z'),
        );

        final json = original.toJson();
        final restored = ProfileModel.fromJson(json);

        expect(restored.id, original.id);
        expect(restored.name, original.name);
        expect(restored.email, original.email);
        expect(restored.phone, original.phone);
        expect(restored.emergencyContactName, original.emergencyContactName);
        expect(restored.emergencyContactPhone,
            original.emergencyContactPhone);
        expect(restored.address, original.address);
        expect(restored.faceEnrolled, original.faceEnrolled);
      });
    });
  });
}
