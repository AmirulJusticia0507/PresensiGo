import 'dart:async';
import 'dart:convert';
import 'dart:math';

import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:flutter/foundation.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'package:uuid/uuid.dart';

import 'api_service.dart';
import 'secure_storage_service.dart';

class OfflineQueueService {
  OfflineQueueService._();
  static final instance = OfflineQueueService._();

  static const _boxName = 'encrypted_attendance_queue';
  final pendingCount = ValueNotifier<int>(0);
  final stuckCount = ValueNotifier<int>(0);
  final _storage = SecureStorageService();
  Box<dynamic>? _box;
  StreamSubscription<List<ConnectivityResult>>? _connectivitySubscription;
  bool _syncing = false;

  Future<void> initialize() async {
    await Hive.initFlutter();
    var encodedKey = await _storage.getOfflineQueueKey();
    if (encodedKey == null) {
      final random = Random.secure();
      encodedKey = base64Encode(
        List<int>.generate(32, (_) => random.nextInt(256)),
      );
      await _storage.saveOfflineQueueKey(encodedKey);
    }
    _box = await Hive.openBox<dynamic>(
      _boxName,
      encryptionCipher: HiveAesCipher(base64Decode(encodedKey)),
    );
    _refreshCounts();
    _connectivitySubscription = Connectivity().onConnectivityChanged.listen((
      results,
    ) {
      if (results.any((result) => result != ConnectivityResult.none)) {
        unawaited(syncPending());
      }
    });
    unawaited(syncPending());
  }

  Future<String> enqueue(
    String actionType,
    Map<String, dynamic> payload, {
    String? idempotencyKey,
  }) async {
    final key = idempotencyKey ?? const Uuid().v4();
    await _box!.put(key, {
      'idempotency_key': key,
      'action_type': actionType,
      'payload': payload,
      'created_at': DateTime.now().toUtc().toIso8601String(),
      'attempts': 0,
      'next_retry_at': DateTime.now().toUtc().toIso8601String(),
      'last_error': null,
    });
    _refreshCounts();
    return key;
  }

  bool hasPending(String actionType) =>
      _box?.values.any(
        (value) => (value as Map)['action_type'] == actionType,
      ) ??
      false;

  Future<void> syncPending() async {
    if (_syncing || _box == null || _box!.isEmpty) return;
    _syncing = true;
    try {
      final entries = _box!.toMap().entries.toList()
        ..sort(
          (a, b) => (a.value['created_at'] as String).compareTo(
            b.value['created_at'] as String,
          ),
        );
      for (final entry in entries) {
        final item = Map<String, dynamic>.from(entry.value as Map);
        final nextRetry = DateTime.parse(item['next_retry_at'] as String);
        if (DateTime.now().toUtc().isBefore(nextRetry)) continue;

        final response = await ApiService.syncAttendance([item]);
        if (response['success'] == true) {
          final results = response['results'] as List<dynamic>;
          final status = (results.first as Map<String, dynamic>)['status'];
          if (status == 'synced' || status == 'duplicate') {
            await _box!.delete(entry.key);
            _refreshCounts();
            continue;
          }
        }

        final attempts = (item['attempts'] as int) + 1;
        final delayMinutes = min(1 << min(attempts, 6), 60);
        item['attempts'] = attempts;
        item['next_retry_at'] = DateTime.now()
            .toUtc()
            .add(Duration(minutes: delayMinutes))
            .toIso8601String();
        item['last_error'] = response['message'] ?? 'Sync rejected';
        await _box!.put(entry.key, item);
        break; // Preserve FIFO ordering.
      }
    } finally {
      _syncing = false;
      _refreshCounts();
    }
  }

  Future<void> dispose() async {
    await _connectivitySubscription?.cancel();
  }

  void _refreshCounts() {
    pendingCount.value = _box?.length ?? 0;
    stuckCount.value = _box?.values
            .where((value) => ((value as Map)['attempts'] as int? ?? 0) >= 5)
            .length ??
        0;
  }
}
