import 'package:flutter/material.dart';

import '../../../core/constants/api_constants.dart';
import '../../../data/services/api_service.dart';

class LeaveScreen extends StatefulWidget {
  const LeaveScreen({super.key});
  @override
  State<LeaveScreen> createState() => _LeaveScreenState();
}

class _LeaveScreenState extends State<LeaveScreen> {
  List<dynamic> items = [];
  @override
  void initState() {
    super.initState();
    load();
  }

  Future<void> load() async {
    try {
      items = await ApiService.adminRequest(ApiConstants.leaves) as List;
    } catch (_) {
      items = [];
    }
    if (mounted) setState(() {});
  }

  Future<void> create() async {
    final reason = TextEditingController();
    DateTimeRange? range;
    String type = 'leave';
    await showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setLocal) => AlertDialog(
          title: const Text('Request leave'),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              DropdownButtonFormField(
                initialValue: type,
                items: const [
                  DropdownMenuItem(value: 'leave', child: Text('Leave')),
                  DropdownMenuItem(value: 'sick', child: Text('Sick')),
                  DropdownMenuItem(
                    value: 'permission',
                    child: Text('Permission'),
                  ),
                ],
                onChanged: (v) => type = v!,
              ),
              TextField(
                controller: reason,
                decoration: const InputDecoration(labelText: 'Reason'),
              ),
              TextButton(
                onPressed: () async {
                  final picked = await showDateRangePicker(
                    context: ctx,
                    firstDate: DateTime.now(),
                    lastDate: DateTime.now().add(const Duration(days: 365)),
                  );
                  setLocal(() => range = picked);
                },
                child: Text(
                  range == null
                      ? 'Choose dates'
                      : '${range!.start.toString().substring(0, 10)} – ${range!.end.toString().substring(0, 10)}',
                ),
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx),
              child: const Text('Cancel'),
            ),
            FilledButton(
              onPressed: () async {
                if (range == null || reason.text.trim().isEmpty) return;
                await ApiService.adminRequest(
                  ApiConstants.leaves,
                  method: 'POST',
                  body: {
                    'start_date': range!.start.toString().substring(0, 10),
                    'end_date': range!.end.toString().substring(0, 10),
                    'type': type,
                    'reason': reason.text.trim(),
                  },
                );
                if (ctx.mounted) Navigator.pop(ctx);
              },
              child: const Text('Submit'),
            ),
          ],
        ),
      ),
    );
    load();
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Leave & Absence')),
    floatingActionButton: FloatingActionButton(
      onPressed: create,
      child: const Icon(Icons.add),
    ),
    body: RefreshIndicator(
      onRefresh: load,
      child: ListView.builder(
        itemCount: items.length,
        itemBuilder: (_, i) {
          final x = items[i] as Map<String, dynamic>;
          return ListTile(
            title: Text('${x['type']} • ${x['status']}'),
            subtitle: Text(
              '${x['start_date']} – ${x['end_date']}\n${x['reason']}',
            ),
            isThreeLine: true,
          );
        },
      ),
    ),
  );
}
