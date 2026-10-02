import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../../../core/constants/api_constants.dart';
import '../../../data/services/api_service.dart';

class AdminDashboardScreen extends StatefulWidget {
  const AdminDashboardScreen({super.key});

  @override
  State<AdminDashboardScreen> createState() => _AdminDashboardScreenState();
}

class _AdminDashboardScreenState extends State<AdminDashboardScreen> {
  int _tab = 0;
  bool _loading = true;
  List<dynamic> _items = [];

  static const _titles = [
    'Attendances',
    'Users',
    'Locations',
    'Schedules',
    'Leaves',
  ];

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _loading = true);
    try {
      final paths = [
        ApiConstants.adminAttendances,
        ApiConstants.adminUsers,
        ApiConstants.locations,
        ApiConstants.adminSchedules,
        ApiConstants.adminLeaves,
      ];
      final data = await ApiService.adminRequest(paths[_tab]);
      _items = data is Map<String, dynamic>
          ? (data['items'] as List? ?? [])
          : data as List;
    } catch (_) {
      _items = [];
    }
    if (mounted) setState(() => _loading = false);
  }

  Future<void> _resetDevice(Map<String, dynamic> user) async {
    await ApiService.adminRequest(
      '${ApiConstants.adminUsers}/${user['id']}',
      method: 'PATCH',
      body: {'reset_device': true},
    );
    if (mounted) {
      ScaffoldMessenger.of(context)
          .showSnackBar(const SnackBar(content: Text('Device binding reset')));
    }
    _load();
  }

  Future<void> _reviewLeave(Map<String, dynamic> leave, String status) async {
    await ApiService.adminRequest(
      '${ApiConstants.adminLeaves}/${leave['id']}',
      method: 'PATCH',
      body: {'status': status},
    );
    _load();
  }

  Future<void> _exportCsv() async {
    final csv = await ApiService.exportAttendancesCsv();
    await Clipboard.setData(ClipboardData(text: csv));
    if (mounted) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('CSV copied to clipboard')));
    }
  }

  Future<void> _changeRole(Map<String, dynamic> user) async {
    await ApiService.adminRequest(
      '${ApiConstants.adminUsers}/${user['id']}',
      method: 'PATCH',
      body: {'role': user['role'] == 'admin' ? 'employee' : 'admin'},
    );
    _load();
  }

  Future<void> _delete(String path) async {
    await ApiService.adminRequest(path, method: 'DELETE');
    _load();
  }

  Future<void> _createLocation() async {
    final name = TextEditingController();
    final address = TextEditingController();
    final latitude = TextEditingController();
    final longitude = TextEditingController();
    await showDialog<void>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('New location'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: name,
              decoration: const InputDecoration(labelText: 'Name'),
            ),
            TextField(
              controller: address,
              decoration: const InputDecoration(labelText: 'Address'),
            ),
            TextField(
              controller: latitude,
              decoration: const InputDecoration(labelText: 'Latitude'),
            ),
            TextField(
              controller: longitude,
              decoration: const InputDecoration(labelText: 'Longitude'),
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
              await ApiService.adminRequest(
                ApiConstants.locations,
                method: 'POST',
                body: {
                  'name': name.text,
                  'address': address.text,
                  'latitude': double.tryParse(latitude.text),
                  'longitude': double.tryParse(longitude.text),
                  'radius_meters': 50,
                },
              );
              if (ctx.mounted) Navigator.pop(ctx);
            },
            child: const Text('Save'),
          ),
        ],
      ),
    );
    _load();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Admin Dashboard'),
        actions: [
          if (_tab == 0)
            IconButton(
              icon: const Icon(Icons.download),
              tooltip: 'Copy CSV',
              onPressed: _exportCsv,
            ),
        ],
      ),
      floatingActionButton: _tab == 2
          ? FloatingActionButton(
              onPressed: _createLocation,
              child: const Icon(Icons.add_location_alt),
            )
          : null,
      body: Column(
        children: [
          SizedBox(
            height: 54,
            child: ListView.builder(
              scrollDirection: Axis.horizontal,
              itemCount: _titles.length,
              itemBuilder: (_, i) => Padding(
                padding: const EdgeInsets.all(6),
                child: ChoiceChip(
                  label: Text(_titles[i]),
                  selected: _tab == i,
                  onSelected: (_) {
                    setState(() => _tab = i);
                    _load();
                  },
                ),
              ),
            ),
          ),
          Expanded(
            child: _loading
                ? const Center(child: CircularProgressIndicator())
                : RefreshIndicator(
                    onRefresh: _load,
                    child: ListView.builder(
                      itemCount: _items.length,
                      itemBuilder: (_, index) => _tile(
                        Map<String, dynamic>.from(_items[index] as Map),
                      ),
                    ),
                  ),
          ),
        ],
      ),
    );
  }

  Widget _tile(Map<String, dynamic> item) {
    switch (_tab) {
      case 0:
        return ListTile(
          leading: const Icon(Icons.fingerprint),
          title: Text(item['user_name'] ?? '-'),
          subtitle: Text(
            '${item['location_name'] ?? '-'} • ${item['status'] ?? '-'}',
          ),
        );
      case 1:
        return ListTile(
          leading: const Icon(Icons.person),
          title: Text(item['name'] ?? '-'),
          subtitle: Text('${item['email']} • ${item['role']}'),
          trailing: PopupMenuButton<String>(
            onSelected: (value) {
              if (value == 'reset') _resetDevice(item);
              if (value == 'role') _changeRole(item);
              if (value == 'delete') {
                _delete('${ApiConstants.adminUsers}/${item['id']}');
              }
            },
            itemBuilder: (_) => const [
              PopupMenuItem(value: 'reset', child: Text('Reset device')),
              PopupMenuItem(value: 'role', child: Text('Toggle role')),
              PopupMenuItem(value: 'delete', child: Text('Delete')),
            ],
          ),
        );
      case 2:
        return ListTile(
          leading: const Icon(Icons.location_on),
          title: Text(item['name'] ?? '-'),
          subtitle: Text(item['address'] ?? '-'),
          trailing: IconButton(
            icon: const Icon(Icons.delete_outline),
            onPressed: () => _delete('${ApiConstants.locations}/${item['id']}'),
          ),
        );
      case 3:
        return ListTile(
          leading: const Icon(Icons.schedule),
          title: Text(
            'Day ${item['day_of_week']} • ${item['start_time']}–${item['end_time']}',
          ),
          subtitle: Text('Late tolerance: ${item['late_after_minutes']} min'),
        );
      default:
        return ListTile(
          leading: const Icon(Icons.event_available),
          title: Text('${item['user_name']} • ${item['type']}'),
          subtitle: Text(
            '${item['start_date']} – ${item['end_date']}\n${item['status']}',
          ),
          isThreeLine: true,
          trailing: item['status'] == 'pending'
              ? PopupMenuButton<String>(
                  onSelected: (v) => _reviewLeave(item, v),
                  itemBuilder: (_) => const [
                    PopupMenuItem(value: 'approved', child: Text('Approve')),
                    PopupMenuItem(value: 'rejected', child: Text('Reject')),
                  ],
                )
              : null,
        );
    }
  }
}
