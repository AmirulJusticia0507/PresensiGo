class ProfileModel {
  final String id;
  final String name;
  final String email;
  final String? phone;
  final String? emergencyContactName;
  final String? emergencyContactPhone;
  final String? address;
  final String? profilePictureUrl;
  final bool faceEnrolled;
  final DateTime? faceEnrolledAt;
  final DateTime createdAt;
  final DateTime updatedAt;

  ProfileModel({
    required this.id,
    required this.name,
    required this.email,
    this.phone,
    this.emergencyContactName,
    this.emergencyContactPhone,
    this.address,
    this.profilePictureUrl,
    this.faceEnrolled = false,
    this.faceEnrolledAt,
    required this.createdAt,
    required this.updatedAt,
  });

  factory ProfileModel.fromJson(Map<String, dynamic> json) {
    return ProfileModel(
      id: json['id'],
      name: json['name'],
      email: json['email'],
      phone: json['phone'],
      emergencyContactName: json['emergency_contact_name'],
      emergencyContactPhone: json['emergency_contact_phone'],
      address: json['address'],
      profilePictureUrl: json['profile_picture_url'],
      faceEnrolled: json['face_enrolled'] ?? false,
      faceEnrolledAt: json['face_enrolled_at'] != null
          ? DateTime.parse(json['face_enrolled_at'])
          : null,
      createdAt: DateTime.parse(json['created_at']),
      updatedAt: DateTime.parse(json['updated_at']),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'email': email,
      'phone': phone,
      'emergency_contact_name': emergencyContactName,
      'emergency_contact_phone': emergencyContactPhone,
      'address': address,
      'profile_picture_url': profilePictureUrl,
      'face_enrolled': faceEnrolled,
      'face_enrolled_at': faceEnrolledAt?.toIso8601String(),
      'created_at': createdAt.toIso8601String(),
      'updated_at': updatedAt.toIso8601String(),
    };
  }

  /// Create a copy with updated fields
  ProfileModel copyWith({
    String? id,
    String? name,
    String? email,
    String? phone,
    String? emergencyContactName,
    String? emergencyContactPhone,
    String? address,
    String? profilePictureUrl,
    bool? faceEnrolled,
    DateTime? faceEnrolledAt,
    DateTime? createdAt,
    DateTime? updatedAt,
  }) {
    return ProfileModel(
      id: id ?? this.id,
      name: name ?? this.name,
      email: email ?? this.email,
      phone: phone ?? this.phone,
      emergencyContactName: emergencyContactName ?? this.emergencyContactName,
      emergencyContactPhone:
          emergencyContactPhone ?? this.emergencyContactPhone,
      address: address ?? this.address,
      profilePictureUrl: profilePictureUrl ?? this.profilePictureUrl,
      faceEnrolled: faceEnrolled ?? this.faceEnrolled,
      faceEnrolledAt: faceEnrolledAt ?? this.faceEnrolledAt,
      createdAt: createdAt ?? this.createdAt,
      updatedAt: updatedAt ?? this.updatedAt,
    );
  }
}
