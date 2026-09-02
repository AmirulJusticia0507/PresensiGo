class UserModel {
  final String id;
  final String name;
  final String email;
  final String role;
  final String? deviceUuid;

  UserModel({
    required this.id,
    required this.name,
    required this.email,
    required this.role,
    this.deviceUuid,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) {
    return UserModel(
      id: json['id'],
      name: json['name'],
      email: json['email'],
      role: json['role'],
      deviceUuid: json['device_uuid'],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'email': email,
      'role': role,
      'device_uuid': deviceUuid,
    };
  }
}
