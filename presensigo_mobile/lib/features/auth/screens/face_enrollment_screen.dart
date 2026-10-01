import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:image_picker/image_picker.dart';

import '../../../core/theme/app_theme.dart';
import '../../../data/services/api_service.dart';

class FaceEnrollmentScreen extends StatefulWidget {
  const FaceEnrollmentScreen({super.key});

  @override
  State<FaceEnrollmentScreen> createState() => _FaceEnrollmentScreenState();
}

class _FaceEnrollmentScreenState extends State<FaceEnrollmentScreen> {
  static const _steps = [
    ('Look straight', 'Keep your face centered with even lighting.'),
    ('Turn slightly left', 'Keep both eyes visible.'),
    ('Turn slightly right', 'Keep both eyes visible.'),
  ];

  final List<Uint8List> _photos = [];
  bool _submitting = false;

  Future<void> _capture() async {
    final photo = await ImagePicker().pickImage(
      source: ImageSource.camera,
      preferredCameraDevice: CameraDevice.front,
      maxWidth: 500,
      maxHeight: 500,
      imageQuality: 70,
    );
    if (photo == null) return;
    final bytes = await photo.readAsBytes();
    if (bytes.length > 1024 * 1024 || !_isSupported(bytes)) {
      _message('Use a JPEG/PNG selfie smaller than 1 MB.', error: true);
      return;
    }
    if (!mounted) return;
    setState(() => _photos.add(bytes));
  }

  Future<void> _submit() async {
    setState(() => _submitting = true);
    final result = await ApiService.enrollFace(
      _photos.map(base64Encode).toList(),
    );
    if (!mounted) return;
    setState(() => _submitting = false);
    if (result['success'] == true) {
      await showDialog<void>(
        context: context,
        builder: (context) => AlertDialog(
          title: const Text('Enrollment complete'),
          content: const Text('Your face can now be verified during check-in.'),
          actions: [
            FilledButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Done'),
            ),
          ],
        ),
      );
      if (mounted) Navigator.pop(context, true);
    } else {
      _message(result['message'] as String, error: true);
    }
  }

  bool _isSupported(Uint8List bytes) {
    final jpeg =
        bytes.length >= 3 &&
        bytes[0] == 0xff &&
        bytes[1] == 0xd8 &&
        bytes[2] == 0xff;
    final png =
        bytes.length >= 8 &&
        bytes[0] == 0x89 &&
        bytes[1] == 0x50 &&
        bytes[2] == 0x4e &&
        bytes[3] == 0x47;
    return jpeg || png;
  }

  void _message(String message, {required bool error}) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: error ? AppTheme.errorColor : AppTheme.secondaryColor,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final complete = _photos.length == _steps.length;
    return Scaffold(
      appBar: AppBar(title: const Text('Face enrollment')),
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              const Text(
                'Capture three selfies',
                style: TextStyle(fontSize: 22, fontWeight: FontWeight.w700),
              ),
              const SizedBox(height: 8),
              Text(
                'Remove masks and glasses. Use good lighting and keep only one face in frame.',
                style: TextStyle(color: AppTheme.textSecondary),
              ),
              const SizedBox(height: 24),
              Expanded(
                child: ListView.separated(
                  itemCount: _steps.length,
                  separatorBuilder: (_, _) => const SizedBox(height: 12),
                  itemBuilder: (context, index) {
                    final captured = index < _photos.length;
                    final active = index == _photos.length;
                    return ListTile(
                      tileColor: Colors.white,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(14),
                        side: BorderSide(
                          color: active
                              ? AppTheme.primaryColor
                              : AppTheme.borderColor,
                        ),
                      ),
                      leading: captured
                          ? ClipOval(
                              child: Image.memory(
                                _photos[index],
                                width: 48,
                                height: 48,
                                fit: BoxFit.cover,
                              ),
                            )
                          : CircleAvatar(child: Text('${index + 1}')),
                      title: Text(
                        _steps[index].$1,
                        style: const TextStyle(fontWeight: FontWeight.w600),
                      ),
                      subtitle: Text(_steps[index].$2),
                      trailing: captured
                          ? const Icon(
                              Icons.check_circle,
                              color: AppTheme.secondaryColor,
                            )
                          : null,
                      onTap: active ? _capture : null,
                    );
                  },
                ),
              ),
              if (!complete)
                FilledButton.icon(
                  onPressed: _capture,
                  icon: const Icon(Icons.camera_alt),
                  label: Text(
                    'Capture ${_steps[_photos.length].$1.toLowerCase()}',
                  ),
                )
              else
                FilledButton.icon(
                  onPressed: _submitting ? null : _submit,
                  icon: _submitting
                      ? const SizedBox(
                          width: 18,
                          height: 18,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Icon(Icons.verified_user),
                  label: const Text('Complete enrollment'),
                ),
              if (_photos.isNotEmpty) ...[
                const SizedBox(height: 8),
                TextButton(
                  onPressed: _submitting ? null : () => setState(_photos.clear),
                  child: const Text('Start over'),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}
