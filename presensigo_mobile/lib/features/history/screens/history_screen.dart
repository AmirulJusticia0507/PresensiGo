import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import '../../../core/theme/app_theme.dart';
import '../../../data/services/api_service.dart';
import '../../../data/models/attendance_model.dart';

class HistoryScreen extends StatefulWidget {
  const HistoryScreen({super.key});

  @override
  State<HistoryScreen> createState() => _HistoryScreenState();
}

class _HistoryScreenState extends State<HistoryScreen> {
  List<AttendanceModel> _history = [];
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadHistory();
  }

  Future<void> _loadHistory() async {
    setState(() => _isLoading = true);
    final history = await ApiService.getHistory(limit: 20);
    setState(() {
      _history = history;
      _isLoading = false;
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Attendance History'),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _history.isEmpty
              ? const Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Icon(Icons.history, size: 64, color: AppTheme.textSecondary),
                      SizedBox(height: 16),
                      Text('No attendance records yet'),
                    ],
                  ),
                )
              : RefreshIndicator(
                  onRefresh: _loadHistory,
                  child: ListView.builder(
                    padding: const EdgeInsets.all(16),
                    itemCount: _history.length,
                    itemBuilder: (context, index) {
                      final attendance = _history[index];
                      return _buildAttendanceCard(attendance);
                    },
                  ),
                ),
    );
  }

  Widget _buildAttendanceCard(AttendanceModel attendance) {
    final dateFormat = DateFormat('dd MMM yyyy');
    final timeFormat = DateFormat('HH:mm');

    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  attendance.locationName ?? 'Unknown Location',
                  style: Theme.of(context).textTheme.titleMedium?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
                ),
                Chip(
                  label: Text(attendance.status.toUpperCase()),
                  backgroundColor: attendance.status == 'present'
                      ? AppTheme.secondaryColor.withOpacity(0.1)
                      : AppTheme.errorColor.withOpacity(0.1),
                ),
              ],
            ),
            const SizedBox(height: 12),
            if (attendance.checkInTime != null)
              Row(
                children: [
                  const Icon(Icons.login, size: 16, color: AppTheme.secondaryColor),
                  const SizedBox(width: 8),
                  Text('Check In: ${timeFormat.format(attendance.checkInTime!)}'),
                  const Spacer(),
                  Text(dateFormat.format(attendance.checkInTime!)),
                ],
              ),
            if (attendance.checkOutTime != null) ...[
              const SizedBox(height: 8),
              Row(
                children: [
                  const Icon(Icons.logout, size: 16, color: AppTheme.errorColor),
                  const SizedBox(width: 8),
                  Text('Check Out: ${timeFormat.format(attendance.checkOutTime!)}'),
                ],
              ),
            ],
            if (attendance.isLate) ...[
              const SizedBox(height: 8),
              const Row(
                children: [
                  Icon(Icons.warning, size: 16, color: AppTheme.errorColor),
                  SizedBox(width: 8),
                  Text('Late', style: TextStyle(color: AppTheme.errorColor)),
                ],
              ),
            ],
          ],
        ),
      ),
    );
  }
}
