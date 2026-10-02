import 'package:flutter/material.dart';

import '../../../core/theme/app_theme.dart';
import '../../../core/utils/form_validator.dart';
import '../../../data/models/profile_model.dart';
import '../../../data/services/profile_service.dart';

class ProfileScreen extends StatefulWidget {
  const ProfileScreen({super.key, this.service});

  /// Injected in tests; defaults to the shared [ProfileService] singleton.
  final ProfileService? service;

  static const Key nameFieldKey = Key('profile_field_name');
  static const Key emailFieldKey = Key('profile_field_email');
  static const Key phoneFieldKey = Key('profile_field_phone');
  static const Key emergencyNameFieldKey = Key('profile_field_emergency_name');
  static const Key emergencyPhoneFieldKey = Key('profile_field_emergency_phone');
  static const Key addressFieldKey = Key('profile_field_address');
  static const Key editButtonKey = Key('profile_edit_button');
  static const Key saveButtonKey = Key('profile_save_button');
  static const Key cancelButtonKey = Key('profile_cancel_button');
  static const Key changePasswordButtonKey = Key('profile_change_password_button');
  static const Key loadingIndicatorKey = Key('profile_loading_indicator');
  static const Key errorRetryButtonKey = Key('profile_error_retry_button');

  @override
  State<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends State<ProfileScreen> {
  late final ProfileService _profileService;
  final _formKey = GlobalKey<FormState>();

  ProfileModel? _profile;
  bool _isLoading = true;
  bool _isEditing = false;
  bool _isSaving = false;
  String? _error;

  // Form field values for editing
  late TextEditingController _nameController;
  late TextEditingController _emailController;
  late TextEditingController _phoneController;
  late TextEditingController _emergencyContactNameController;
  late TextEditingController _emergencyContactPhoneController;
  late TextEditingController _addressController;

  // Field validation errors
  final Map<String, String> _fieldErrors = {};

  @override
  void initState() {
    super.initState();
    _profileService = widget.service ?? ProfileService();
    _initializeControllers();
    _loadProfile();
  }

  void _initializeControllers() {
    _nameController = TextEditingController();
    _emailController = TextEditingController();
    _phoneController = TextEditingController();
    _emergencyContactNameController = TextEditingController();
    _emergencyContactPhoneController = TextEditingController();
    _addressController = TextEditingController();
  }

  @override
  void dispose() {
    _nameController.dispose();
    _emailController.dispose();
    _phoneController.dispose();
    _emergencyContactNameController.dispose();
    _emergencyContactPhoneController.dispose();
    _addressController.dispose();
    super.dispose();
  }

  Future<void> _loadProfile() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    final result = await _profileService.getProfile();

    if (!mounted) return;

    if (result.profile != null) {
      setState(() {
        _profile = result.profile;
        _populateForm();
        _isLoading = false;
      });
    } else {
      setState(() {
        _error = result.error ?? 'Failed to load profile';
        _isLoading = false;
      });
    }
  }

  void _populateForm() {
    if (_profile == null) return;

    _nameController.text = _profile!.name;
    _emailController.text = _profile!.email;
    _phoneController.text = _profile!.phone ?? '';
    _emergencyContactNameController.text =
        _profile!.emergencyContactName ?? '';
    _emergencyContactPhoneController.text =
        _profile!.emergencyContactPhone ?? '';
    _addressController.text = _profile!.address ?? '';
  }

  void _enterEditMode() {
    setState(() {
      _isEditing = true;
      _fieldErrors.clear();
    });
  }

  void _cancelEdit() {
    setState(() {
      _isEditing = false;
      _fieldErrors.clear();
      _populateForm();
    });
  }

  Future<void> _saveProfile() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    setState(() {
      _isSaving = true;
      _fieldErrors.clear();
    });

    final result = await _profileService.updateProfile(
      name: _nameController.text.trim(),
      phone: _phoneController.text.trim().isNotEmpty
          ? _phoneController.text.trim()
          : null,
      emergencyContactName:
          _emergencyContactNameController.text.trim().isNotEmpty
              ? _emergencyContactNameController.text.trim()
              : null,
      emergencyContactPhone:
          _emergencyContactPhoneController.text.trim().isNotEmpty
              ? _emergencyContactPhoneController.text.trim()
              : null,
      address: _addressController.text.trim().isNotEmpty
          ? _addressController.text.trim()
          : null,
    );

    if (!mounted) return;

    if (result.profile != null) {
      setState(() {
        _profile = result.profile;
        _isEditing = false;
        _isSaving = false;
        _fieldErrors.clear();
      });

      _showSuccess('Profile updated successfully');
    } else if (result.fieldErrors != null && result.fieldErrors!.isNotEmpty) {
      setState(() {
        _fieldErrors.addAll(result.fieldErrors!);
        _isSaving = false;
      });
    } else {
      setState(() {
        _isSaving = false;
        _error = result.error ?? 'Failed to update profile';
      });

      _showError(result.error ?? 'Failed to update profile');
    }
  }

  void _showChangePasswordDialog() {
    final currentPasswordController = TextEditingController();
    final newPasswordController = TextEditingController();
    final confirmPasswordController = TextEditingController();
    final formKey = GlobalKey<FormState>();
    bool isSubmitting = false;

    showDialog<void>(
      context: context,
      barrierDismissible: false,
      builder: (context) => StatefulBuilder(
        builder: (context, setState) => AlertDialog(
          title: const Text('Change Password'),
          content: Form(
            key: formKey,
            child: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  TextFormField(
                    controller: currentPasswordController,
                    obscureText: true,
                    decoration: const InputDecoration(
                      labelText: 'Current Password',
                      hintText: 'Enter your current password',
                    ),
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return 'Current password is required';
                      }
                      return null;
                    },
                  ),
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: newPasswordController,
                    obscureText: true,
                    decoration: const InputDecoration(
                      labelText: 'New Password',
                      hintText: 'Enter new password',
                    ),
                    onChanged: (_) => setState(() {}),
                    validator: (value) {
                      final result =
                          FormValidator.validatePassword(value, optional: false);
                      return result.isValid ? null : result.errorMessage;
                    },
                  ),
                  if (newPasswordController.text.isNotEmpty) ...[
                    const SizedBox(height: 8),
                    _buildPasswordStrengthIndicator(newPasswordController.text),
                  ],
                  const SizedBox(height: 16),
                  TextFormField(
                    controller: confirmPasswordController,
                    obscureText: true,
                    decoration: const InputDecoration(
                      labelText: 'Confirm Password',
                      hintText: 'Confirm new password',
                    ),
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return 'Password confirmation is required';
                      }
                      final matchResult = FormValidator.validatePasswordMatch(
                        newPasswordController.text,
                        value,
                      );
                      return matchResult.isValid ? null : matchResult.errorMessage;
                    },
                  ),
                ],
              ),
            ),
          ),
          actions: [
            TextButton(
              onPressed: isSubmitting ? null : () => Navigator.pop(context),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: isSubmitting
                  ? null
                  : () async {
                      if (!formKey.currentState!.validate()) {
                        return;
                      }

                      setState(() => isSubmitting = true);

                      final result =
                          await _profileService.changePassword(
                        currentPassword: currentPasswordController.text,
                        newPassword: newPasswordController.text,
                        confirmPassword: confirmPasswordController.text,
                      );

                      if (!context.mounted) return;

                      Navigator.pop(context);

                      if (result.success) {
                        _showSuccess('Password changed successfully');
                      } else {
                        _showError(
                          result.error ??
                              'Failed to change password',
                        );
                      }
                    },
              child: isSubmitting
                  ? const SizedBox(
                      height: 20,
                      width: 20,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                      ),
                    )
                  : const Text('Change'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildPasswordStrengthIndicator(String password) {
    final strength = FormValidator.getPasswordStrength(password);
    final strengthColor = Color(strength.colorValue);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Text(
              'Strength: ${strength.label}',
              style: TextStyle(
                fontSize: 12,
                color: strengthColor,
                fontWeight: FontWeight.w500,
              ),
            ),
          ],
        ),
        const SizedBox(height: 6),
        ClipRRect(
          borderRadius: BorderRadius.circular(4),
          child: LinearProgressIndicator(
            value: strength.index / PasswordStrength.strong.index,
            minHeight: 4,
            backgroundColor: Colors.grey[300],
            valueColor: AlwaysStoppedAnimation<Color>(strengthColor),
          ),
        ),
      ],
    );
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
            const Icon(
              Icons.check_circle_outline,
              color: Colors.white,
              size: 20,
            ),
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
        leading: _isEditing
            ? IconButton(
                key: ProfileScreen.cancelButtonKey,
                icon: const Icon(Icons.close),
                onPressed: _cancelEdit,
              )
            : null,
        title: const Text('Profile'),
        centerTitle: true,
        actions: [
          if (!_isEditing)
            Container(
              margin: const EdgeInsets.only(right: 12),
              decoration: BoxDecoration(
                color: AppTheme.primaryColor.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(10),
              ),
              child: IconButton(
                key: ProfileScreen.editButtonKey,
                icon: const Icon(
                  Icons.edit_rounded,
                  color: AppTheme.primaryColor,
                ),
                onPressed: _enterEditMode,
              ),
            ),
        ],
      ),
      body: _isLoading
          ? const Center(
              child: CircularProgressIndicator(
                key: ProfileScreen.loadingIndicatorKey,
              ),
            )
          : _error != null
              ? _buildErrorState()
              : RefreshIndicator(
                  onRefresh: () =>
                      _profileService.getProfile(forceRefresh: true).then(
                    (result) {
                      if (result.profile != null) {
                        setState(() {
                          _profile = result.profile;
                          _populateForm();
                        });
                      }
                    },
                  ),
                  color: AppTheme.primaryColor,
                  child: SingleChildScrollView(
                    physics: const AlwaysScrollableScrollPhysics(),
                    padding: const EdgeInsets.all(16),
                    child: _isEditing
                        ? _buildEditForm()
                        : _buildReadOnlyView(),
                  ),
                ),
    );
  }

  Widget _buildErrorState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.error_outline,
            size: 64,
            color: AppTheme.errorColor.withValues(alpha: 0.5),
          ),
          const SizedBox(height: 16),
          Text(
            _error ?? 'Unknown error',
            textAlign: TextAlign.center,
            style: const TextStyle(color: AppTheme.textSecondary),
          ),
          const SizedBox(height: 24),
          ElevatedButton(
            key: ProfileScreen.errorRetryButtonKey,
            onPressed: _loadProfile,
            child: const Text('Retry'),
          ),
        ],
      ),
    );
  }

  Widget _buildReadOnlyView() {
    if (_profile == null) {
      return const Center(child: Text('No profile data'));
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Profile Header Card
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: AppTheme.borderColor),
            boxShadow: AppShadows.small,
          ),
          child: Column(
            children: [
              Container(
                width: 80,
                height: 80,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: AppTheme.primaryColor.withValues(alpha: 0.1),
                ),
                child: Icon(
                  Icons.person_rounded,
                  size: 40,
                  color: AppTheme.primaryColor,
                ),
              ),
              const SizedBox(height: 12),
              Text(
                _profile!.name,
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.w600,
                  color: AppTheme.textPrimary,
                ),
              ),
              const SizedBox(height: 4),
              Text(
                _profile!.email,
                style: const TextStyle(
                  fontSize: 13,
                  color: AppTheme.textSecondary,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 20),

        // Contact Information
        _buildSectionTitle('Contact Information'),
        const SizedBox(height: 12),
        _buildInfoCard('Email', _profile!.email),
        _buildInfoCard('Phone', _profile!.phone ?? 'Not provided'),
        const SizedBox(height: 20),

        // Emergency Contact
        _buildSectionTitle('Emergency Contact'),
        const SizedBox(height: 12),
        _buildInfoCard(
          'Name',
          _profile!.emergencyContactName ?? 'Not provided',
        ),
        _buildInfoCard(
          'Phone',
          _profile!.emergencyContactPhone ?? 'Not provided',
        ),
        const SizedBox(height: 20),

        // Address
        _buildSectionTitle('Address'),
        const SizedBox(height: 12),
        _buildInfoCard('Address', _profile!.address ?? 'Not provided'),
        const SizedBox(height: 24),

        // Face Status
        Container(
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: (_profile!.faceEnrolled
                    ? AppTheme.secondaryColor
                    : AppTheme.textMuted)
                .withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: (_profile!.faceEnrolled
                      ? AppTheme.secondaryColor
                      : AppTheme.textMuted)
                  .withValues(alpha: 0.2),
            ),
          ),
          child: Row(
            children: [
              Icon(
                _profile!.faceEnrolled ? Icons.verified : Icons.info,
                color: _profile!.faceEnrolled
                    ? AppTheme.secondaryColor
                    : AppTheme.textMuted,
                size: 24,
              ),
              const SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      _profile!.faceEnrolled
                          ? 'Face Recognition Enrolled'
                          : 'Face Recognition Not Enrolled',
                      style: TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w600,
                        color: _profile!.faceEnrolled
                            ? AppTheme.secondaryColor
                            : AppTheme.textMuted,
                      ),
                    ),
                    if (_profile!.faceEnrolledAt != null)
                      Text(
                        'Enrolled on ${_formatDate(_profile!.faceEnrolledAt!)}',
                        style: const TextStyle(
                          fontSize: 12,
                          color: AppTheme.textMuted,
                        ),
                      ),
                  ],
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: 20),

        // Change Password Button
        SizedBox(
          width: double.infinity,
          child: OutlinedButton(
            key: ProfileScreen.changePasswordButtonKey,
            onPressed: _showChangePasswordDialog,
            style: OutlinedButton.styleFrom(
              padding: const EdgeInsets.symmetric(vertical: 12),
              side: const BorderSide(color: AppTheme.primaryColor),
            ),
            child: const Text('Change Password'),
          ),
        ),
      ],
    );
  }

  Widget _buildEditForm() {
    return Form(
      key: _formKey,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Name
          _buildFormLabel('Name *'),
          const SizedBox(height: 8),
          _buildTextField(
            key: ProfileScreen.nameFieldKey,
            controller: _nameController,
            label: 'Full Name',
            validator: (value) {
              final result = FormValidator.validateName(value);
              if (!result.isValid) {
                return result.errorMessage;
              }
              return null;
            },
            errorText: _fieldErrors['name'],
          ),
          const SizedBox(height: 16),

          // Email (read-only)
          _buildFormLabel('Email'),
          const SizedBox(height: 8),
          TextField(
            key: ProfileScreen.emailFieldKey,
            controller: _emailController,
            enabled: false,
            decoration: const InputDecoration(
              labelText: 'Email',
              hintText: 'Email cannot be changed',
              filled: true,
              fillColor: Color(0xFFF1F5F9),
            ),
          ),
          const SizedBox(height: 16),

          // Phone
          _buildFormLabel('Phone'),
          const SizedBox(height: 8),
          _buildTextField(
            key: ProfileScreen.phoneFieldKey,
            controller: _phoneController,
            label: 'Phone Number (e.g., +1234567890)',
            hint: '+1234567890',
            validator: (value) {
              final result = FormValidator.validatePhone(value);
              if (!result.isValid) {
                return result.errorMessage;
              }
              return null;
            },
            errorText: _fieldErrors['phone'],
          ),
          const SizedBox(height: 16),

          // Emergency Contact Name
          _buildFormLabel('Emergency Contact Name'),
          const SizedBox(height: 8),
          _buildTextField(
            key: ProfileScreen.emergencyNameFieldKey,
            controller: _emergencyContactNameController,
            label: 'Contact Name',
            validator: (value) {
              final result = FormValidator.validateName(value, optional: true);
              if (!result.isValid) {
                return result.errorMessage;
              }
              return null;
            },
            errorText: _fieldErrors['emergency_contact_name'],
          ),
          const SizedBox(height: 16),

          // Emergency Contact Phone
          _buildFormLabel('Emergency Contact Phone'),
          const SizedBox(height: 8),
          _buildTextField(
            key: ProfileScreen.emergencyPhoneFieldKey,
            controller: _emergencyContactPhoneController,
            label: 'Emergency Phone (e.g., +1234567890)',
            hint: '+1234567890',
            validator: (value) {
              final result = FormValidator.validatePhone(value);
              if (!result.isValid) {
                return result.errorMessage;
              }
              return null;
            },
            errorText: _fieldErrors['emergency_contact_phone'],
          ),
          const SizedBox(height: 16),

          // Address
          _buildFormLabel('Address'),
          const SizedBox(height: 8),
          _buildTextField(
            key: ProfileScreen.addressFieldKey,
            controller: _addressController,
            label: 'Full Address',
            minLines: 2,
            maxLines: 4,
            validator: (value) {
              final result = FormValidator.validateAddress(value);
              if (!result.isValid) {
                return result.errorMessage;
              }
              return null;
            },
            errorText: _fieldErrors['address'],
          ),
          const SizedBox(height: 24),

          // Save Button
          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              key: ProfileScreen.saveButtonKey,
              onPressed: _isSaving ? null : _saveProfile,
              child: _isSaving
                  ? const SizedBox(
                      height: 20,
                      width: 20,
                      child: CircularProgressIndicator(
                        strokeWidth: 2,
                        valueColor:
                            AlwaysStoppedAnimation<Color>(Colors.white),
                      ),
                    )
                  : const Text('Save Changes'),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFormLabel(String label) {
    return Text(
      label,
      style: const TextStyle(
        fontSize: 14,
        fontWeight: FontWeight.w600,
        color: AppTheme.textPrimary,
      ),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    Key? key,
    String? hint,
    int minLines = 1,
    int maxLines = 1,
    String? Function(String?)? validator,
    String? errorText,
  }) {
    return TextFormField(
      key: key,
      controller: controller,
      minLines: minLines,
      maxLines: maxLines,
      decoration: InputDecoration(
        labelText: label,
        hintText: hint,
        errorText: errorText,
      ),
      validator: validator,
    );
  }

  Widget _buildSectionTitle(String title) {
    return Text(
      title,
      style: const TextStyle(
        fontSize: 14,
        fontWeight: FontWeight.w600,
        color: AppTheme.textPrimary,
      ),
    );
  }

  Widget _buildInfoCard(String label, String value) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: AppTheme.borderColor),
      ),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  label,
                  style: const TextStyle(
                    fontSize: 12,
                    color: AppTheme.textSecondary,
                  ),
                ),
                const SizedBox(height: 4),
                Text(
                  value,
                  style: const TextStyle(
                    fontSize: 14,
                    fontWeight: FontWeight.w500,
                    color: AppTheme.textPrimary,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  String _formatDate(DateTime date) {
    return '${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}';
  }
}
