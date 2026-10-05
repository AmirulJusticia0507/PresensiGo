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
  int _total = 0;
  String? _attendanceStatus;

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
      final attendancePath = Uri(
        path: ApiConstants.adminAttendances,
        queryParameters: {
          'limit': '50',
          if (_attendanceStatus != null) 'status': _attendanceStatus!,
        },
      ).toString();
      final data = await ApiService.adminRequest(
        _tab == 0 ? attendancePath : paths[_tab],
      );
      _items = data is Map<String, dynamic>
          ? (data['items'] as List? ?? [])
          : data as List;
      _total = data is Map<String, dynamic>
          ? (data['total'] as int? ?? _items.length)
          : _items.length;
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

  Future<bool> _confirm(String message) async {
    return await showDialog<bool>(
          context: context,
          builder: (ctx) => AlertDialog(
            title: const Text('Confirm action'),
            content: Text(message),
            actions: [
              TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Cancel')),
              FilledButton(onPressed: () => Navigator.pop(ctx, true), child: const Text('Continue')),
            ],
          ),
        ) ??
        false;
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

  Future<void> _editLocation(Map<String, dynamic> location) async {
    final name = TextEditingController(text: location['name']?.toString());
    final address = TextEditingController(text: location['address']?.toString());
    final latitude = TextEditingController(text: location['latitude']?.toString());
    final longitude = TextEditingController(text: location['longitude']?.toString());
    final radius = TextEditingController(text: location['radius_meters']?.toString() ?? '50');
    await showDialog<void>(context: context, builder: (ctx) => AlertDialog(
      title: const Text('Edit location'),
      content: SingleChildScrollView(child: Column(mainAxisSize: MainAxisSize.min, children: [
        TextField(controller: name, decoration: const InputDecoration(labelText: 'Name')),
        TextField(controller: address, decoration: const InputDecoration(labelText: 'Address')),
        TextField(controller: latitude, decoration: const InputDecoration(labelText: 'Latitude')),
        TextField(controller: longitude, decoration: const InputDecoration(labelText: 'Longitude')),
        TextField(controller: radius, decoration: const InputDecoration(labelText: 'Radius meters')),
      ])),
      actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel')), FilledButton(onPressed: () async {
        await ApiService.adminRequest('${ApiConstants.locations}/${location['id']}', method: 'PUT', body: {
          'name': name.text, 'address': address.text,
          'latitude': double.tryParse(latitude.text), 'longitude': double.tryParse(longitude.text),
          'radius_meters': int.tryParse(radius.text),
        });
        if (ctx.mounted) Navigator.pop(ctx);
      }, child: const Text('Save'))],
    ));
    _load();
  }

  Future<void> _editSchedule([Map<String, dynamic>? schedule]) async {
    var scope = schedule?['user_id'] != null ? 'user' : 'location';
    final target = TextEditingController(text: (schedule?['user_id'] ?? schedule?['location_id'] ?? '').toString());
    var day = schedule?['day_of_week'] as int? ?? 1;
    final start = TextEditingController(text: (schedule?['start_time'] ?? '08:00').toString().substring(0, 5));
    final end = TextEditingController(text: (schedule?['end_time'] ?? '17:00').toString().substring(0, 5));
    final tolerance = TextEditingController(text: (schedule?['late_after_minutes'] ?? 15).toString());
    var active = schedule?['active'] as bool? ?? true;
    await showDialog<void>(context: context, builder: (ctx) => StatefulBuilder(builder: (ctx, setLocal) => AlertDialog(
      title: Text(schedule == null ? 'New schedule' : 'Edit schedule'),
      content: SingleChildScrollView(child: Column(mainAxisSize: MainAxisSize.min, children: [
        DropdownButtonFormField<String>(initialValue: scope, decoration: const InputDecoration(labelText: 'Scope'), items: const [DropdownMenuItem(value: 'user', child: Text('User')), DropdownMenuItem(value: 'location', child: Text('Location'))], onChanged: schedule == null ? (v) => setLocal(() => scope = v ?? scope) : null),
        TextField(controller: target, enabled: schedule == null, decoration: InputDecoration(labelText: scope == 'user' ? 'User UUID' : 'Location UUID')),
        DropdownButtonFormField<int>(initialValue: day, decoration: const InputDecoration(labelText: 'Day'), items: List.generate(7, (i) => DropdownMenuItem(value: i, child: Text(['Sunday','Monday','Tuesday','Wednesday','Thursday','Friday','Saturday'][i]))), onChanged: (v) => setLocal(() => day = v ?? day)),
        TextField(controller: start, decoration: const InputDecoration(labelText: 'Start (HH:mm)')),
        TextField(controller: end, decoration: const InputDecoration(labelText: 'End (HH:mm)')),
        TextField(controller: tolerance, decoration: const InputDecoration(labelText: 'Late tolerance (minutes)')),
        SwitchListTile(value: active, title: const Text('Active'), onChanged: (v) => setLocal(() => active = v)),
      ])),
      actions: [TextButton(onPressed: () => Navigator.pop(ctx), child: const Text('Cancel')), FilledButton(onPressed: () async {
        await ApiService.adminRequest(ApiConstants.adminSchedules, method: 'POST', body: {
          if (schedule != null) 'id': schedule['id'],
          if (scope == 'user') 'user_id': target.text else 'location_id': target.text,
          'day_of_week': day, 'start_time': start.text, 'end_time': end.text,
          'late_after_minutes': int.tryParse(tolerance.text), 'active': active,
        });
        if (ctx.mounted) Navigator.pop(ctx);
      }, child: const Text('Save'))],
    )));
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
          : _tab == 3
          ? FloatingActionButton(
              onPressed: _editSchedule,
              child: const Icon(Icons.add_alarm),
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
          if (_tab == 0)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12),
              child: Row(children: [
                Expanded(child: DropdownButtonFormField<String?>(
                  initialValue: _attendanceStatus,
                  decoration: const InputDecoration(labelText: 'Attendance status'),
                  items: const [DropdownMenuItem(value: null, child: Text('All')), DropdownMenuItem(value: 'present', child: Text('Present')), DropdownMenuItem(value: 'late', child: Text('Late')), DropdownMenuItem(value: 'leave', child: Text('Leave')), DropdownMenuItem(value: 'absent', child: Text('Absent'))],
                  onChanged: (value) { _attendanceStatus = value; _load(); },
                )),
                const SizedBox(width: 12),
                Text('${_items.length}/$_total'),
              ]),
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
          onTap: () => _editLocation(item),
          trailing: IconButton(
            icon: const Icon(Icons.delete_outline),
            onPressed: () async {
              if (await _confirm('Delete ${item['name']}?')) {
                _delete('${ApiConstants.locations}/${item['id']}');
              }
            },
          ),
        );
      case 3:
        return ListTile(
          leading: const Icon(Icons.schedule),
          title: Text(
            'Day ${item['day_of_week']} • ${item['start_time']}–${item['end_time']}',
          ),
          subtitle: Text('Late tolerance: ${item['late_after_minutes']} min'),
          onTap: () => _editSchedule(item),
          trailing: IconButton(
            icon: const Icon(Icons.delete_outline),
            onPressed: () async {
              if (await _confirm('Delete this schedule?')) {
                _delete('${ApiConstants.adminSchedules}/${item['id']}');
              }
            },
          ),
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
