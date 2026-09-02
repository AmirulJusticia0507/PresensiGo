class AttendanceModel {
  final String id;
  final String userId;
  final String locationId;
  final DateTime? checkInTime;
  final DateTime? checkOutTime;
  final List<double>? checkInLocation;
  final List<double>? checkOutLocation;
  final String? selfieUrl;
  final String status;
  final bool isLate;
  final String? notes;
  final String? userName;
  final String? locationName;

  AttendanceModel({
    required this.id,
    required this.userId,
    required this.locationId,
    this.checkInTime,
    this.checkOutTime,
    this.checkInLocation,
    this.checkOutLocation,
    this.selfieUrl,
    required this.status,
    required this.isLate,
    this.notes,
    this.userName,
    this.locationName,
  });

  factory AttendanceModel.fromJson(Map<String, dynamic> json) {
    return AttendanceModel(
      id: json['id'],
      userId: json['user_id'],
      locationId: json['location_id'],
      checkInTime: json['check_in_time'] != null
          ? DateTime.parse(json['check_in_time'])
          : null,
      checkOutTime: json['check_out_time'] != null
          ? DateTime.parse(json['check_out_time'])
          : null,
      checkInLocation: json['check_in_location'] != null
          ? List<double>.from(json['check_in_location'])
          : null,
      checkOutLocation: json['check_out_location'] != null
          ? List<double>.from(json['check_out_location'])
          : null,
      selfieUrl: json['selfie_url'],
      status: json['status'],
      isLate: json['isLate'] ?? false,
      notes: json['notes'],
      userName: json['user_name'],
      locationName: json['location_name'],
    );
  }
}

class LocationModel {
  final String id;
  final String name;
  final String? address;
  final double latitude;
  final double longitude;
  final int radiusMeters;

  LocationModel({
    required this.id,
    required this.name,
    this.address,
    required this.latitude,
    required this.longitude,
    required this.radiusMeters,
  });

  factory LocationModel.fromJson(Map<String, dynamic> json) {
    return LocationModel(
      id: json['id'],
      name: json['name'],
      address: json['address'],
      latitude: json['latitude'].toDouble(),
      longitude: json['longitude'].toDouble(),
      radiusMeters: json['radius_meters'],
    );
  }
}
