import 'package:flutter/material.dart';

import '../../../core/errors/api_exception.dart';
import '../../../core/theme/app_theme.dart';
import '../../../core/utils/form_validator.dart';
import '../../../data/services/profile_service.dart';
import '../../attendance/screens/attendance_screen.dart';

class RegisterScreen extends StatefulWidget {
  const RegisterScreen({super.key, this.service});

  /// Injected in tests; defaults to the shared [ProfileService] singleton.
  final ProfileService? service;

  static const Key emailFieldKey = Key('register_field_email');
  static const Key nameFieldKey = Key('register_field_name');
  static const Key phoneFieldKey = Key('register_field_phone');
  static const Key passwordFieldKey = Key('register_field_password');
  static const Key confirmPasswordFieldKey =
      Key('register_field_confirm_password');
  static const Key termsCheckboxKey = Key('register_terms_checkbox');
  static const Key submitButtonKey = Key('register_submit_button');
  static const Key strengthIndicatorKey = Key('register_strength_indicator');

  @override
  State<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends State<RegisterScreen> {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _confirmPasswordController = TextEditingController();
  final _nameController = TextEditingController();
  final _phoneController = TextEditingController();

  bool _obscurePassword = true;
  bool _obscureConfirmPassword = true;
  bool _termsAccepted = false;
  bool _isLoading = false;

  // Validation state
  String? _emailError;
  String? _passwordError;
  String? _confirmPasswordError;
  String? _nameError;
  String? _phoneError;
  int _passwordStrength = 0;

  // API error handling
  final Map<String, String> _fieldErrors = {};
  String? _generalError;

  @override
  void initState() {
    super.initState();
    _emailController.addListener(_validateEmail);
    _passwordController.addListener(_validatePassword);
    _confirmPasswordController.addListener(_validateConfirmPassword);
    _nameController.addListener(_validateName);
    _phoneController.addListener(_validatePhone);
  }

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    _confirmPasswordController.dispose();
    _nameController.dispose();
    _phoneController.dispose();
    super.dispose();
  }

  void _validateEmail() {
    final result = FormValidator.validateEmail(_emailController.text);
    setState(() {
      _emailError = result.error;
      _fieldErrors.remove('email');
    });
  }

  void _validatePassword() {
    final result = FormValidator.validatePassword(_passwordController.text);
    setState(() {
      _passwordError = result.error;
      _passwordStrength = result.strength;
      _fieldErrors.remove('password');
    });
    // Also validate confirm password if it's not empty
    if (_confirmPasswordController.text.isNotEmpty) {
      _validateConfirmPassword();
    }
  }

  void _validateConfirmPassword() {
    final result = FormValidator.validateConfirmPassword(
      _passwordController.text,
      _confirmPasswordController.text,
    );
    setState(() {
      _confirmPasswordError = result.error;
      _fieldErrors.remove('confirm_password');
    });
  }

  void _validateName() {
    final result = FormValidator.validateName(_nameController.text);
    setState(() {
      _nameError = result.error;
      _fieldErrors.remove('name');
    });
  }

  void _validatePhone() {
    final result = FormValidator.validatePhone(_phoneController.text);
    setState(() {
      _phoneError = result.error;
      _fieldErrors.remove('phone');
    });
  }

  bool _isFormValid() {
    return FormValidator.isFormValid(
      email: _emailController.text,
      password: _passwordController.text,
      confirmPassword: _confirmPasswordController.text,
      name: _nameController.text,
      phone: _phoneController.text.isEmpty ? null : _phoneController.text,
      termsAccepted: _termsAccepted,
    );
  }

  Future<void> _register() async {
    // Final validation
    if (!_isFormValid()) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please fill all required fields correctly')),
      );
      return;
    }

    setState(() {
      _isLoading = true;
      _generalError = null;
      _fieldErrors.clear();
    });

    try {
      final profile = await (widget.service ?? ProfileService()).register(
        email: _emailController.text.trim(),
        password: _passwordController.text,
        confirmPassword: _confirmPasswordController.text,
        name: _nameController.text.trim(),
        phone: _phoneController.text.isEmpty
            ? null
            : _phoneController.text.trim(),
        termsAccepted: _termsAccepted,
      );

      if (!mounted) return;

      // Show success message
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Welcome ${profile.name}!'),
          backgroundColor: Colors.green,
        ),
      );

      // Navigate to AttendanceScreen (ProfileScreen will be available from there)
      Navigator.of(context).pushReplacement(
        PageRouteBuilder(
          pageBuilder: (_, animation, __) => const AttendanceScreen(),
          transitionsBuilder: (_, animation, __, child) {
            return FadeTransition(opacity: animation, child: child);
          },
        ),
      );
    } on ApiException catch (e) {
      setState(() {
        _isLoading = false;
        if (e.statusCode == 409) {
          _generalError = 'Email already registered';
          _fieldErrors['email'] = 'This email is already in use';
        } else if (e.statusCode == 400 && e.details != null) {
          // Parse field-level errors
          for (final detail in e.details!) {
            final field = detail['field'] as String?;
            final reason = detail['reason'] as String?;
            if (field != null && reason != null) {
              _fieldErrors[field] = reason;
            }
          }
          if (_fieldErrors.isEmpty) {
            _generalError = e.message;
          }
        } else {
          _generalError = e.message;
        }
      });

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(_generalError ?? 'Registration failed')),
        );
      }
    } catch (e) {
      setState(() {
        _isLoading = false;
        _generalError = 'An unexpected error occurred';
      });

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('An unexpected error occurred')),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        title: const Text('Create Account'),
        elevation: 0,
        backgroundColor: AppTheme.primaryColor,
        foregroundColor: Colors.white,
      ),
      body: SingleChildScrollView(
        child: Padding(
          padding: const EdgeInsets.all(24.0),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const SizedBox(height: 24),
                // Email Field
                _buildTextField(
                  key: RegisterScreen.emailFieldKey,
                  controller: _emailController,
                  label: 'Email Address',
                  hint: 'your@email.com',
                  keyboardType: TextInputType.emailAddress,
                  error: _emailError ?? _fieldErrors['email'],
                  icon: Icons.email_outlined,
                ),
                const SizedBox(height: 20),

                // Name Field
                _buildTextField(
                  key: RegisterScreen.nameFieldKey,
                  controller: _nameController,
                  label: 'Full Name',
                  hint: 'John Doe',
                  error: _nameError ?? _fieldErrors['name'],
                  icon: Icons.person_outlined,
                ),
                const SizedBox(height: 20),

                // Phone Field (Optional)
                _buildTextField(
                  key: RegisterScreen.phoneFieldKey,
                  controller: _phoneController,
                  label: 'Phone Number',
                  hint: '+1234567890',
                  keyboardType: TextInputType.phone,
                  error: _phoneError ?? _fieldErrors['phone'],
                  icon: Icons.phone_outlined,
                  isOptional: true,
                ),
                const SizedBox(height: 20),

                // Password Field
                _buildPasswordField(
                  key: RegisterScreen.passwordFieldKey,
                  controller: _passwordController,
                  label: 'Password',
                  obscureText: _obscurePassword,
                  onToggleVisibility: () {
                    setState(() => _obscurePassword = !_obscurePassword);
                  },
                  error: _passwordError ?? _fieldErrors['password'],
                ),
                const SizedBox(height: 12),

                // Password Strength Indicator
                _buildPasswordStrengthIndicator(),
                const SizedBox(height: 20),

                // Confirm Password Field
                _buildPasswordField(
                  key: RegisterScreen.confirmPasswordFieldKey,
                  controller: _confirmPasswordController,
                  label: 'Confirm Password',
                  obscureText: _obscureConfirmPassword,
                  onToggleVisibility: () {
                    setState(() =>
                        _obscureConfirmPassword = !_obscureConfirmPassword);
                  },
                  error: _confirmPasswordError ?? _fieldErrors['confirm_password'],
                ),
                const SizedBox(height: 24),

                // Terms Checkbox
                _buildTermsCheckbox(),
                const SizedBox(height: 28),

                // General Error Message
                if (_generalError != null)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.red.shade50,
                        border: Border.all(color: Colors.red.shade300),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        _generalError!,
                        style: TextStyle(color: Colors.red.shade700),
                      ),
                    ),
                  ),

                // Register Button
                ElevatedButton(
                  key: RegisterScreen.submitButtonKey,
                  onPressed: _isLoading || !_isFormValid() ? null : _register,
                  style: ElevatedButton.styleFrom(
                    backgroundColor: AppTheme.primaryColor,
                    disabledBackgroundColor: Colors.grey.shade300,
                    padding: const EdgeInsets.symmetric(vertical: 16),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                  child: _isLoading
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(
                            strokeWidth: 2,
                            valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                          ),
                        )
                      : const Text(
                          'Create Account',
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.w600,
                            color: Colors.white,
                          ),
                        ),
                ),
                const SizedBox(height: 16),

                // Login Link
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Text('Already have an account? '),
                    GestureDetector(
                      onTap: () => Navigator.of(context).pop(),
                      child: Text(
                        'Login',
                        style: TextStyle(
                          color: AppTheme.primaryColor,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 24),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    required String hint,
    Key? key,
    TextInputType keyboardType = TextInputType.text,
    String? error,
    IconData? icon,
    bool isOptional = false,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          isOptional ? '$label (Optional)' : label,
          style: const TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w500,
            color: Color(0xFF2C3E50),
          ),
        ),
        const SizedBox(height: 8),
        TextField(
          key: key,
          controller: controller,
          keyboardType: keyboardType,
          decoration: InputDecoration(
            hintText: hint,
            prefixIcon: icon != null ? Icon(icon, color: AppTheme.primaryColor) : null,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
              borderSide: BorderSide(
                color: error != null ? Colors.red : Colors.grey.shade300,
              ),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
              borderSide: BorderSide(
                color: error != null ? Colors.red : Colors.grey.shade300,
              ),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
              borderSide: BorderSide(
                color: error != null ? Colors.red : AppTheme.primaryColor,
                width: 2,
              ),
            ),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 12,
            ),
            errorText: null, // Handled below
          ),
        ),
        if (error != null) ...[
          const SizedBox(height: 6),
          Text(
            error,
            style: const TextStyle(
              fontSize: 12,
              color: Colors.red,
            ),
          ),
        ],
      ],
    );
  }

  Widget _buildPasswordField({
    required TextEditingController controller,
    required String label,
    required bool obscureText,
    required VoidCallback onToggleVisibility,
    Key? key,
    String? error,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: const TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w500,
            color: Color(0xFF2C3E50),
          ),
        ),
        const SizedBox(height: 8),
        TextField(
          key: key,
          controller: controller,
          obscureText: obscureText,
          decoration: InputDecoration(
            hintText: '••••••••',
            prefixIcon: Icon(Icons.lock_outline, color: AppTheme.primaryColor),
            suffixIcon: GestureDetector(
              onTap: onToggleVisibility,
              child: Icon(
                obscureText ? Icons.visibility_off : Icons.visibility,
                color: AppTheme.primaryColor,
              ),
            ),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
              borderSide: BorderSide(
                color: error != null ? Colors.red : Colors.grey.shade300,
              ),
            ),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
              borderSide: BorderSide(
                color: error != null ? Colors.red : Colors.grey.shade300,
              ),
            ),
            focusedBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
              borderSide: BorderSide(
                color: error != null ? Colors.red : AppTheme.primaryColor,
                width: 2,
              ),
            ),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 16,
              vertical: 12,
            ),
          ),
        ),
        if (error != null) ...[
          const SizedBox(height: 6),
          Text(
            error,
            style: const TextStyle(
              fontSize: 12,
              color: Colors.red,
            ),
          ),
        ],
      ],
    );
  }

  Widget _buildPasswordStrengthIndicator() {
    final strength = _passwordStrength;
    final colors = [
      Colors.red,
      Colors.orange,
      Colors.amber,
      Colors.lime,
      Colors.green,
    ];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Expanded(
              child: ClipRRect(
                borderRadius: BorderRadius.circular(4),
                child: LinearProgressIndicator(
                  key: RegisterScreen.strengthIndicatorKey,
                  value: strength / 5,
                  minHeight: 6,
                  backgroundColor: Colors.grey.shade200,
                  valueColor: AlwaysStoppedAnimation<Color>(
                    strength == 0 ? Colors.grey.shade200 : colors[strength - 1],
                  ),
                ),
              ),
            ),
            const SizedBox(width: 12),
            Text(
              strength == 0
                  ? 'No password'
                  : strength == 5
                      ? 'Strong'
                      : 'Weak',
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: strength == 0
                    ? Colors.grey
                    : strength < 4
                        ? Colors.orange
                        : Colors.green,
              ),
            ),
          ],
        ),
        const SizedBox(height: 12),
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Password requirements:',
              style: TextStyle(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: Colors.grey.shade700,
              ),
            ),
            const SizedBox(height: 8),
            ..._buildRequirementsList(),
          ],
        ),
      ],
    );
  }

  List<Widget> _buildRequirementsList() {
    final requirements = [
      ('At least 8 characters', RegExp(r'.{8}').hasMatch(_passwordController.text)),
      ('Uppercase letter (A-Z)', RegExp(r'[A-Z]').hasMatch(_passwordController.text)),
      ('Lowercase letter (a-z)', RegExp(r'[a-z]').hasMatch(_passwordController.text)),
      ('Number (0-9)', RegExp(r'[0-9]').hasMatch(_passwordController.text)),
      ('Special character (!@#\$%^&*)', RegExp(r'[!@#$%^&*]').hasMatch(_passwordController.text)),
    ];

    return requirements.map((req) {
      final isMet = req.$2;
      return Padding(
        padding: const EdgeInsets.only(bottom: 6),
        child: Row(
          children: [
            Icon(
              isMet ? Icons.check_circle : Icons.circle_outlined,
              size: 16,
              color: isMet ? Colors.green : Colors.grey.shade400,
            ),
            const SizedBox(width: 8),
            Text(
              req.$1,
              style: TextStyle(
                fontSize: 12,
                color: isMet ? Colors.green : Colors.grey.shade600,
                decoration: isMet ? TextDecoration.lineThrough : null,
              ),
            ),
          ],
        ),
      );
    }).toList();
  }

  Widget _buildTermsCheckbox() {
    return Row(
      children: [
        Checkbox(
          key: RegisterScreen.termsCheckboxKey,
          value: _termsAccepted,
          onChanged: (value) {
            setState(() {
              _termsAccepted = value ?? false;
            });
          },
          activeColor: AppTheme.primaryColor,
        ),
        Expanded(
          child: GestureDetector(
            onTap: () {
              setState(() {
                _termsAccepted = !_termsAccepted;
              });
            },
            child: Text(
              'I accept the terms and conditions',
              style: TextStyle(
                fontSize: 13,
                color: Colors.grey.shade700,
              ),
            ),
          ),
        ),
      ],
    );
  }
}
