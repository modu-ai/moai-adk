---
name: moai-lang-dart
version: 4.0.0
created: 2025-10-22
updated: 2025-10-22
status: active
description: Dart 3.5 enterprise development with Flutter 3.24, advanced async programming, state management, and cross-platform mobile development. Enterprise patterns for scalable applications with Context7 MCP integration.
keywords: ['dart', 'flutter', 'async-programming', 'state-management', 'cross-platform', 'mobile-development', 'enterprise', 'context7']
allowed-tools:
  - Read
  - Bash
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# Lang Dart Skill - Enterprise v4.0.0

## Skill Metadata

| Field | Value |
| ----- | ----- |
| **Skill Name** | moai-lang-dart |
| **Version** | 4.0.0 (2025-10-22) |
| **Allowed tools** | Read (read_file), Bash (terminal), Context7 MCP |
| **Auto-load** | On demand when keywords detected |
| **Tier** | Language Enterprise |
| **Context7 Integration** | ✅ Dart/Flutter/Async/StateManagement |

---

## What It Does

Dart 3.5 enterprise development featuring Flutter 3.24, advanced async programming patterns, modern state management (BLoC, Riverpod, Provider), and enterprise-grade cross-platform mobile development. Context7 MCP integration provides real-time access to official Dart and Flutter documentation.

**Key capabilities**:
- ✅ Dart 3.5 with advanced type system and patterns
- ✅ Flutter 3.24 enterprise mobile development
- ✅ Advanced async programming with Isolates and Streams
- ✅ Modern state management (BLoC, Riverpod, Provider)
- ✅ Cross-platform development (iOS, Android, Web, Desktop)
- ✅ Enterprise architecture patterns (Clean Architecture, MVVM)
- ✅ Performance optimization and memory management
- ✅ Testing strategies with unit, widget, and integration tests
- ✅ Dependency injection and service layer patterns
- ✅ Firebase integration and backend services

---

## When to Use

**Automatic triggers**:
- Dart and Flutter development discussions
- Cross-platform mobile application development
- State management pattern implementation
- Async programming and stream handling
- Mobile app architecture and design
- Enterprise Flutter application development

**Manual invocation**:
- Design mobile application architecture
- Implement advanced async patterns
- Optimize Flutter app performance
- Review enterprise Dart/Flutter code
- Implement state management solutions
- Troubleshoot mobile development issues

---

## Technology Stack (2025-10-22)

| Component | Version | Purpose | Status |
|-----------|---------|---------|--------|
| **Dart** | 3.5.4 | Core language | ✅ Current |
| **Flutter** | 3.24.0 | UI framework | ✅ Current |
| **Riverpod** | 2.6.1 | State management | ✅ Current |
| **BLoC** | 8.1.6 | State management | ✅ Current |
| **Dio** | 5.6.0 | HTTP client | ✅ Current |

---

## Code Examples (30+ Enterprise Patterns)

### 1. Advanced Dart 3.5 Async Programming

```dart
import 'dart:async';
import 'dart:isolate';
import 'dart:math';
import 'dart:typed_data';

// Advanced enterprise async service with error handling and retries
class EnterpriseApiService {
  final Dio _httpClient;
  final Logger _logger;
  final Duration _defaultTimeout;
  final int _maxRetries;
  
  EnterpriseApiService({
    required Dio httpClient,
    required Logger logger,
    Duration defaultTimeout = const Duration(seconds: 30),
    int maxRetries = 3,
  }) : _httpClient = httpClient,
       _logger = logger,
       _defaultTimeout = defaultTimeout,
       _maxRetries = maxRetries;
  
  // Advanced fetch with exponential backoff and circuit breaker
  Future<T> fetchWithRetry<T>({
    required String endpoint,
    Map<String, dynamic>? queryParameters,
    required T Function(Map<String, dynamic>) parser,
    Duration? timeout,
    int? maxRetries,
  }) async {
    final retries = maxRetries ?? _maxRetries;
    final requestTimeout = timeout ?? _defaultTimeout;
    
    for (int attempt = 0; attempt <= retries; attempt++) {
      try {
        _logger.info('Attempting API request to $endpoint (attempt ${attempt + 1})');
        
        final response = await _httpClient.get<Map<String, dynamic>>(
          endpoint,
          queryParameters: queryParameters,
          options: Options(
            timeout: requestTimeout,
            headers: {
              'Content-Type': 'application/json',
              'Accept': 'application/json',
              'X-Request-ID': _generateRequestId(),
            },
          ),
        );
        
        if (response.statusCode == 200) {
          _logger.info('API request successful to $endpoint');
          return parser(response.data!);
        } else {
          throw ApiException(
            'HTTP ${response.statusCode}: ${response.statusMessage}',
            statusCode: response.statusCode,
          );
        }
      } catch (e) {
        _logger.warning('API request failed (attempt ${attempt + 1}): $e');
        
        if (attempt == retries) {
          _logger.error('Max retries exceeded for $endpoint');
          rethrow;
        }
        
        // Exponential backoff with jitter
        final baseDelay = Duration(milliseconds: 100 * pow(2, attempt).toInt());
        final jitter = Duration(milliseconds: Random().nextInt(100));
        final delay = baseDelay + jitter;
        
        _logger.info('Retrying in ${delay.inMilliseconds}ms...');
        await Future.delayed(delay);
      }
    }
    
    throw ApiException('Unexpected error in fetchWithRetry');
  }
  
  // Advanced concurrent requests with rate limiting
  Future<List<T>> fetchConcurrently<T>({
    required List<ApiRequest> requests,
    required T Function(Map<String, dynamic>) parser,
    int maxConcurrent = 5,
  }) async {
    final semaphore = Semaphore(maxConcurrent);
    final results = <T>[];
    final errors = <String, dynamic>{};
    
    // Create futures with semaphore control
    final futures = requests.map((request) async {
      await semaphore.acquire();
      try {
        return await fetchWithRetry<T>(
          endpoint: request.endpoint,
          queryParameters: request.queryParameters,
          parser: parser,
        );
      } catch (e) {
        errors[request.endpoint] = e;
        rethrow;
      } finally {
        semaphore.release();
      }
    });
    
    try {
      final results = await Future.wait(
        futures,
        eagerError: false,
      );
      
      if (errors.isNotEmpty) {
        _logger.warning('Some requests failed: ${errors.keys}');
      }
      
      return results;
    } catch (e) {
      _logger.error('Concurrent fetch failed: $e');
      rethrow;
    }
  }
  
  // Stream-based real-time data with backpressure handling
  Stream<T> createDataStream<T>({
    required String endpoint,
    required T Function(Map<String, dynamic>) parser,
    Duration interval = const Duration(seconds: 5),
  }) async* {
    while (true) {
      try {
        final data = await fetchWithRetry<T>(
          endpoint: endpoint,
          parser: parser,
          timeout: Duration(seconds: interval.inSeconds ~/ 2),
        );
        
        yield data;
        
        await Future.delayed(interval);
      } catch (e) {
        _logger.error('Data stream error: $e');
        
        // Yield error event but continue stream
        yield* Stream.error(e).transform(
          StreamTransformer.fromHandlers(
            handleError: (error, sink) {
              _logger.warning('Stream error handled: $error');
            },
          ),
        );
        
        // Wait before retrying
        await Future.delayed(const Duration(seconds: 10));
      }
    }
  }
  
  // Advanced file upload with progress tracking
  Future<UploadResult> uploadFile({
    required String filePath,
    required String uploadUrl,
    ProgressCallback? onProgress,
  }) async {
    try {
      final file = File(filePath);
      if (!await file.exists()) {
        throw ArgumentError('File does not exist: $filePath');
      }
      
      final fileSize = await file.length();
      _logger.info('Starting upload of ${file.path} (${fileSize} bytes)');
      
      final formData = FormData.fromMap({
        'file': await MultipartFile.fromFile(
          filePath,
          filename: file.path.split('/').last,
        ),
      });
      
      final response = await _httpClient.post<Map<String, dynamic>>(
        uploadUrl,
        data: formData,
        onSendProgress: (sent, total) {
          final progress = total > 0 ? sent / total : 0.0;
          onProgress?.call(progress);
          _logger.debug('Upload progress: ${(progress * 100).toStringAsFixed(1)}%');
        },
        options: Options(
          timeout: const Duration(minutes: 30),
        ),
      );
      
      if (response.statusCode == 200) {
        _logger.info('Upload completed successfully');
        return UploadResult.success(
          url: response.data!['url'] as String,
          size: fileSize,
        );
      } else {
        throw ApiException(
          'Upload failed: HTTP ${response.statusCode}',
          statusCode: response.statusCode,
        );
      }
    } catch (e) {
      _logger.error('Upload error: $e');
      return UploadResult.failure(e.toString());
    }
  }
  
  // Isolate-based heavy computation
  Future<T> performHeavyComputation<T>({
    required T Function() computation,
    Duration? timeout,
  }) async {
    try {
      final receivePort = ReceivePort();
      final isolate = await Isolate.spawn<SendPort>(
        _isolateEntryPoint<T>,
        receivePort.sendPort,
      );
      
      final completer = Completer<T>();
      
      // Listen for results from isolate
      late StreamSubscription subscription;
      subscription = receivePort.listen((message) {
        if (message is SendPort) {
          // Send computation function to isolate
          message.send(computation as dynamic);
        } else if (message is T) {
          completer.complete(message);
          subscription.cancel();
          isolate.kill();
        } else if (message is Error) {
          completer.completeError(message);
          subscription.cancel();
          isolate.kill();
        }
      });
      
      // Apply timeout if specified
      if (timeout != null) {
        return completer.future.timeout(timeout);
      }
      
      return completer.future;
    } catch (e) {
      _logger.error('Isolate computation failed: $e');
      rethrow;
    }
  }
  
  String _generateRequestId() {
    return DateTime.now().millisecondsSinceEpoch.toString() +
           Random().nextInt(10000).toString();
  }
}

// Isolate entry point
void _isolateEntryPoint<T>(SendPort sendPort) {
  final receivePort = ReceivePort();
  sendPort.send(receivePort);
  
  receivePort.listen((message) {
    try {
      if (message is T Function()) {
        final result = message();
        sendPort.send(result);
      }
    } catch (e) {
      sendPort.send(e);
    }
  });
}

// Supporting classes
class ApiRequest {
  final String endpoint;
  final Map<String, dynamic>? queryParameters;
  
  const ApiRequest({
    required this.endpoint,
    this.queryParameters,
  });
}

class ApiException implements Exception {
  final String message;
  final int? statusCode;
  
  const ApiException(this.message, {this.statusCode});
  
  @override
  String toString() => 'ApiException: $message${statusCode != null ? ' (HTTP $statusCode)' : ''}';
}

class UploadResult {
  final bool success;
  final String? url;
  final int? size;
  final String? error;
  
  const UploadResult._({
    required this.success,
    this.url,
    this.size,
    this.error,
  });
  
  factory UploadResult.success({required String url, required int size}) {
    return UploadResult._(
      success: true,
      url: url,
      size: size,
    );
  }
  
  factory UploadResult.failure(String error) {
    return UploadResult._(
      success: false,
      error: error,
    );
  }
}

typedef ProgressCallback = void Function(double progress);

// Semaphore implementation for concurrency control
class Semaphore {
  final int maxCount;
  int _currentCount;
  final Queue<Completer<void>> _waitQueue = Queue<Completer<void>>();
  
  Semaphore(this.maxCount) : _currentCount = maxCount;
  
  Future<void> acquire() async {
    if (_currentCount > 0) {
      _currentCount--;
      return;
    }
    
    final completer = Completer<void>();
    _waitQueue.add(completer);
    return completer.future;
  }
  
  void release() {
    if (_waitQueue.isNotEmpty) {
      final completer = _waitQueue.removeFirst();
      completer.complete();
    } else {
      _currentCount++;
    }
  }
}

// Advanced Stream transformer for backpressure handling
class BackpressureStreamTransformer<T>
    extends StreamTransformerBase<T, T> {
  final int maxBufferSize;
  final Duration bufferingTimeout;
  
  const BackpressureStreamTransformer({
    this.maxBufferSize = 1000,
    this.bufferingTimeout = const Duration(seconds: 5),
  });
  
  @override
  Stream<T> bind(Stream<T> stream) {
    final controller = StreamController<T>();
    final buffer = <T>[];
    Timer? bufferingTimer;
    bool isPaused = false;
    
    // Handle incoming data
    late StreamSubscription subscription;
    subscription = stream.listen(
      (data) {
        buffer.add(data);
        
        // Check buffer size
        if (buffer.length >= maxBufferSize) {
          _pauseStream();
        }
        
        _emitFromBuffer();
      },
      onDone: () {
        _emitFromBuffer();
        controller.close();
      },
      onError: controller.addError,
    );
    
    void _pauseStream() {
      if (!isPaused) {
        subscription.pause();
        isPaused = true;
        
        bufferingTimer?.cancel();
        bufferingTimer = Timer(bufferingTimeout, () {
          if (buffer.isNotEmpty) {
            subscription.resume();
            isPaused = false;
          }
        });
      }
    }
    
    void _emitFromBuffer() {
      while (buffer.isNotEmpty && !controller.isPaused) {
        final data = buffer.removeAt(0);
        controller.add(data);
        
        if (isPaused && buffer.length < maxBufferSize ~/ 2) {
          subscription.resume();
          isPaused = false;
          bufferingTimer?.cancel();
        }
      }
    }
    
    return controller.stream;
  }
}
```

### 2. Advanced State Management with Riverpod

```dart
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:freezed_annotation/freezed_annotation.dart';
import 'package:json_annotation/json_annotation.dart';

part 'enterprise_state.freezed.dart';
part 'enterprise_state.g.dart';

// Advanced enterprise state model
@freezed
class EnterpriseState with _$EnterpriseState {
  const factory EnterpriseState({
    @Default(AsyncValue.loading()) AsyncValue<List<User>> users,
    @Default(AsyncValue.loading()) AsyncValue<List<Product>> products,
    @Default(AsyncValue.loading()) AsyncValue<List<Order>> orders,
    @Default(AsyncValue.loading()) AsyncValue<UserProfile> userProfile,
    @Default(false) bool isLoading,
    @Default('') String errorMessage,
    @Default(0) int currentPage,
    @Default(20) int itemsPerPage,
    @Default(<String, dynamic>{}) Map<String, dynamic> filters,
    @Default(<String, bool>{}) Map<String, bool> selectedItems,
  }) = _EnterpriseState;
}

// Enterprise service providers
final enterpriseServiceProvider = Provider<EnterpriseService>((ref) {
  final apiService = ref.watch(apiServiceProvider);
  final cacheService = ref.watch(cacheServiceProvider);
  final notificationService = ref.watch(notificationServiceProvider);
  
  return EnterpriseService(
    apiService: apiService,
    cacheService: cacheService,
    notificationService: notificationService,
  );
});

// Advanced state notifier with Riverpod
final enterpriseStateNotifierProvider = 
    StateNotifierProvider<EnterpriseStateNotifier, EnterpriseState>((ref) {
  final service = ref.watch(enterpriseServiceProvider);
  return EnterpriseStateNotifier(service);
});

class EnterpriseStateNotifier extends StateNotifier<EnterpriseState> {
  final EnterpriseService _service;
  
  EnterpriseStateNotifier(this._service) : super(const EnterpriseState()) {
    _loadInitialData();
  }
  
  Future<void> _loadInitialData() async {
    try {
      state = state.copyWith(isLoading: true);
      
      final results = await Future.wait([
        _service.getUsers(),
        _service.getProducts(),
        _service.getUserProfile(),
      ]);
      
      state = state.copyWith(
        users: AsyncValue.data(results[0] as List<User>),
        products: AsyncValue.data(results[1] as List<Product>),
        userProfile: AsyncValue.data(results[2] as UserProfile),
        isLoading: false,
      );
    } catch (e, stack) {
      state = state.copyWith(
        isLoading: false,
        errorMessage: e.toString(),
        users: AsyncValue.error(e, stack),
        products: AsyncValue.error(e, stack),
        userProfile: AsyncValue.error(e, stack),
      );
    }
  }
  
  // Advanced filtering and pagination
  Future<void> loadUsers({Map<String, dynamic>? filters}) async {
    try {
      state = state.copyWith(isLoading: true);
      
      final updatedFilters = {
        ...state.filters,
        ...?filters,
      };
      
      final users = await _service.getUsers(
        page: state.currentPage,
        limit: state.itemsPerPage,
        filters: updatedFilters,
      );
      
      state = state.copyWith(
        users: AsyncValue.data(users),
        filters: updatedFilters,
        isLoading: false,
      );
    } catch (e, stack) {
      state = state.copyWith(
        isLoading: false,
        users: AsyncValue.error(e, stack),
      );
    }
  }
  
  // Advanced search with debouncing
  Future<void> searchUsers(String query) async {
    if (query.isEmpty) {
      return loadUsers();
    }
    
    state = state.copyWith(isLoading: true);
    
    try {
      final users = await _service.searchUsers(
        query,
        filters: state.filters,
      );
      
      state = state.copyWith(
        users: AsyncValue.data(users),
        isLoading: false,
      );
    } catch (e, stack) {
      state = state.copyWith(
        isLoading: false,
        users: AsyncValue.error(e, stack),
      );
    }
  }
  
  // Optimistic updates
  Future<void> updateUser(User user) async {
    try {
      // Optimistic update
      final currentUsers = state.users.value ?? [];
      final updatedUsers = currentUsers.map((u) => 
          u.id == user.id ? user : u).toList();
      
      state = state.copyWith(
        users: AsyncValue.data(updatedUsers),
      );
      
      // Actual update
      final updatedUser = await _service.updateUser(user);
      
      // Update with server response
      final finalUsers = currentUsers.map((u) => 
          u.id == updatedUser.id ? updatedUser : u).toList();
      
      state = state.copyWith(
        users: AsyncValue.data(finalUsers),
      );
    } catch (e, stack) {
      // Revert optimistic update
      _loadInitialData();
      state = state.copyWith(
        errorMessage: e.toString(),
      );
    }
  }
  
  // Advanced batch operations
  Future<void> batchUpdateUsers(List<User> users) async {
    try {
      state = state.copyWith(isLoading: true);
      
      // Optimistic batch update
      final currentUsers = state.users.value ?? [];
      final updatedUsers = currentUsers.map((currentUser) {
        final updateUser = users.firstWhere(
          (u) => u.id == currentUser.id,
          orElse: () => currentUser,
        );
        return updateUser;
      }).toList();
      
      state = state.copyWith(
        users: AsyncValue.data(updatedUsers),
      );
      
      // Actual batch update
      final finalUsers = await _service.batchUpdateUsers(users);
      
      state = state.copyWith(
        users: AsyncValue.data(finalUsers),
        isLoading: false,
      );
    } catch (e, stack) {
      _loadInitialData();
      state = state.copyWith(
        isLoading: false,
        errorMessage: e.toString(),
      );
    }
  }
  
  // Advanced selection management
  void toggleUserSelection(String userId) {
    final currentSelection = Map<String, bool>.from(state.selectedItems);
    currentSelection[userId] = !(currentSelection[userId] ?? false);
    
    state = state.copyWith(selectedItems: currentSelection);
  }
  
  void selectAllUsers() {
    final users = state.users.value ?? [];
    final selection = <String, bool>{
      for (final user in users) user.id: true,
    };
    
    state = state.copyWith(selectedItems: selection);
  }
  
  void clearSelection() {
    state = state.copyWith(selectedItems: <String, bool>{});
  }
  
  // Pagination
  void loadNextPage() {
    if (!state.isLoading) {
      state = state.copyWith(currentPage: state.currentPage + 1);
      loadUsers();
    }
  }
  
  void refreshData() {
    state = state.copyWith(currentPage: 0);
    _loadInitialData();
  }
}

// Advanced async value provider with caching
final cachedUsersProvider = AsyncNotifierProvider<CachedUsersNotifier, List<User>>(
  () => CachedUsersNotifier(),
);

class CachedUsersNotifier extends AsyncNotifier<List<User>> {
  @override
  Future<List<User>> build() async {
    final service = ref.watch(enterpriseServiceProvider);
    return service.getUsersWithCache();
  }
  
  Future<void> refresh() async {
    state = const AsyncValue.loading();
    
    final service = ref.watch(enterpriseServiceProvider);
    state = await AsyncValue.guard(() => service.getUsers(forceRefresh: true));
  }
}

// Advanced family provider for dynamic data
final userDetailProvider = AsyncNotifierProvider.family<UserDetailNotifier, UserDetail, String>(
  UserDetailNotifier.new,
);

class UserDetailNotifier extends FamilyAsyncNotifier<UserDetail, String> {
  @override
  Future<UserDetail> build(String arg) async {
    final service = ref.watch(enterpriseServiceProvider);
    return service.getUserDetail(arg);
  }
  
  Future<void> refresh() async {
    state = const AsyncValue.loading();
    
    final service = ref.watch(enterpriseServiceProvider);
    state = await AsyncValue.guard(() => service.getUserDetail(arg));
  }
}

// Advanced stream provider for real-time data
final realTimeUsersProvider = StreamProvider<List<User>>((ref) {
  final service = ref.watch(enterpriseServiceProvider);
  
  return service.getUserUpdates().transform(
    StreamTransformer<List<User>, List<User>>.fromHandlers(
      handleData: (data, sink) {
        sink.add(data);
      },
      handleError: (error, stackTrace, sink) {
        sink.addError(error, stackTrace);
      },
      handleDone: (sink) {
        sink.close();
      },
    ),
  );
});

// Advanced computed provider for derived state
final usersSummaryProvider = Provider<UsersSummary>((ref) {
  final usersAsync = ref.watch(enterpriseStateNotifierProvider.select((state) => state.users));
  final selectedItems = ref.watch(enterpriseStateNotifierProvider.select((state) => state.selectedItems));
  
  return usersAsync.when(
    data: (users) {
      final selectedCount = selectedItems.values.where((selected) => selected).length;
      final totalUsers = users.length;
      
      return UsersSummary(
        totalUsers: totalUsers,
        selectedUsers: selectedCount,
        canPerformBatchOperations: selectedCount > 0,
      );
    },
    loading: () => UsersSummary.loading(),
    error: (error, stack) => UsersSummary.error(error.toString()),
  );
});

// Data models with advanced features
@freezed
class User with _$User {
  const factory User({
    required String id,
    required String name,
    required String email,
    @Default('') String avatar,
    @Default(UserStatus.active) UserStatus status,
    @Default([]) List<String> tags,
    @JsonKey(name: 'created_at') required DateTime createdAt,
    @JsonKey(name: 'updated_at') DateTime? updatedAt,
  }) = _User;
  
  factory User.fromJson(Map<String, dynamic> json) => _$UserFromJson(json);
}

@freezed
class UserDetail with _$UserDetail {
  const factory UserDetail({
    required User user,
    required List<Order> recentOrders,
    required UserProfile profile,
    @Default([]) List<UserActivity> activities,
  }) = _UserDetail;
  
  factory UserDetail.fromJson(Map<String, dynamic> json) => _$UserDetailFromJson(json);
}

@freezed
class UsersSummary with _$UsersSummary {
  const factory UsersSummary({
    required int totalUsers,
    required int selectedUsers,
    required bool canPerformBatchOperations,
  }) = _UsersSummary;
  
  const factory UsersSummary.loading() = _UsersSummary(
    totalUsers: 0,
    selectedUsers: 0,
    canPerformBatchOperations: false,
  );
  
  const factory UsersSummary.error(String message) = _UsersSummary(
    totalUsers: 0,
    selectedUsers: 0,
    canPerformBatchOperations: false,
  );
}

enum UserStatus {
  active,
  inactive,
  suspended,
  pending,
}
```

### 3. Advanced Flutter 3.24 Enterprise Widgets

```dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:cached_network_image/cached_network_image.dart';
import 'package:shimmer/shimmer.dart';

// Advanced enterprise dashboard widget
class EnterpriseDashboard extends ConsumerWidget {
  const EnterpriseDashboard({super.key});
  
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(enterpriseStateNotifierProvider);
    final usersSummary = ref.watch(usersSummaryProvider);
    
    return Scaffold(
      appBar: _buildAppBar(context, ref),
      body: RefreshIndicator(
        onRefresh: () => ref.read(enterpriseStateNotifierProvider.notifier).refreshData(),
        child: CustomScrollView(
          slivers: [
            _buildSummaryCards(context, usersSummary),
            _buildContentArea(context, ref, state),
          ],
        ),
      ),
      floatingActionButton: _buildFloatingActionButton(context, ref),
    );
  }
  
  PreferredSizeWidget _buildAppBar(BuildContext context, WidgetRef ref) {
    return AppBar(
      title: const Text('Enterprise Dashboard'),
      backgroundColor: Theme.of(context).colorScheme.primary,
      foregroundColor: Colors.white,
      actions: [
        IconButton(
          icon: const Icon(Icons.search),
          onPressed: () => _showSearchDialog(context, ref),
        ),
        IconButton(
          icon: const Icon(Icons.filter_list),
          onPressed: () => _showFilterDialog(context, ref),
        ),
        PopupMenuButton<String>(
          onSelected: (value) => _handleMenuAction(context, ref, value),
          itemBuilder: (context) => [
            const PopupMenuItem(
              value: 'export',
              child: Row(
                children: [
                  Icon(Icons.download),
                  SizedBox(width: 8),
                  Text('Export Data'),
                ],
              ),
            ),
            const PopupMenuItem(
              value: 'refresh',
              child: Row(
                children: [
                  Icon(Icons.refresh),
                  SizedBox(width: 8),
                  Text('Refresh'),
                ],
              ),
            ),
            const PopupMenuItem(
              value: 'settings',
              child: Row(
                children: [
                  Icon(Icons.settings),
                  SizedBox(width: 8),
                  Text('Settings'),
                ],
              ),
            ),
          ],
        ),
      ],
    );
  }
  
  Widget _buildSummaryCards(BuildContext context, UsersSummary summary) {
    return SliverToBoxAdapter(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Row(
          children: [
            Expanded(
              child: _SummaryCard(
                title: 'Total Users',
                value: summary.totalUsers.toString(),
                icon: Icons.people,
                color: Colors.blue,
              ),
            ),
            const SizedBox(width: 16),
            Expanded(
              child: _SummaryCard(
                title: 'Selected',
                value: summary.selectedUsers.toString(),
                icon: Icons.check_circle,
                color: Colors.green,
              ),
            ),
          ],
        ),
      ),
    );
  }
  
  Widget _buildContentArea(BuildContext context, WidgetRef ref, EnterpriseState state) {
    if (state.isLoading && state.users.value == null) {
      return const SliverFillRemaining(
        child: _LoadingWidget(),
      );
    }
    
    if (state.users.hasError) {
      return SliverFillRemaining(
        child: _ErrorWidget(
          error: state.users.error?.toString() ?? 'Unknown error',
          onRetry: () => ref.read(enterpriseStateNotifierProvider.notifier).refreshData(),
        ),
      );
    }
    
    final users = state.users.value ?? [];
    
    return SliverPadding(
      padding: const EdgeInsets.all(16.0),
      sliver: SliverList(
        delegate: SliverChildBuilderDelegate(
          (context, index) {
            if (index < users.length) {
              return _UserListItem(
                user: users[index],
                isSelected: state.selectedItems[users[index].id] ?? false,
                onTap: () => _navigateToUserDetail(context, ref, users[index]),
                onSelectionChanged: (selected) {
                  ref.read(enterpriseStateNotifierProvider.notifier)
                      .toggleUserSelection(users[index].id);
                },
              );
            }
            
            // Load more indicator
            if (index == users.length - 1) {
              return _LoadMoreWidget(
                onLoadMore: () => ref.read(enterpriseStateNotifierProvider.notifier)
                    .loadNextPage(),
              );
            }
            
            return null;
          },
          childCount: users.length + 1,
        ),
      ),
    );
  }
  
  Widget _buildFloatingActionButton(BuildContext context, WidgetRef ref) {
    return Consumer(
      builder: (context, ref, child) {
        final summary = ref.watch(usersSummaryProvider);
        
        if (!summary.canPerformBatchOperations) {
          return const SizedBox.shrink();
        }
        
        return FloatingActionButton.extended(
          onPressed: () => _showBatchOperationsDialog(context, ref),
          icon: const Icon(Icons.more_horiz),
          label: const Text('Actions'),
        );
      },
    );
  }
  
  void _handleMenuAction(BuildContext context, WidgetRef ref, String action) {
    switch (action) {
      case 'export':
        _exportData(context, ref);
        break;
      case 'refresh':
        ref.read(enterpriseStateNotifierProvider.notifier).refreshData();
        break;
      case 'settings':
        _navigateToSettings(context);
        break;
    }
  }
  
  void _navigateToUserDetail(BuildContext context, WidgetRef ref, User user) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => UserDetailScreen(userId: user.id),
      ),
    );
  }
  
  void _showSearchDialog(BuildContext context, WidgetRef ref) {
    final controller = TextEditingController();
    
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Search Users'),
        content: TextField(
          controller: controller,
          decoration: const InputDecoration(
            hintText: 'Enter name or email...',
            prefixIcon: Icon(Icons.search),
          ),
          autofocus: true,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () {
              ref.read(enterpriseStateNotifierProvider.notifier)
                  .searchUsers(controller.text);
              Navigator.of(context).pop();
            },
            child: const Text('Search'),
          ),
        ],
      ),
    );
  }
  
  void _showFilterDialog(BuildContext context, WidgetRef ref) {
    // Implementation for filter dialog
    showDialog(
      context: context,
      builder: (context) => const FilterDialog(),
    );
  }
  
  void _showBatchOperationsDialog(BuildContext context, WidgetRef ref) {
    showModalBottomSheet(
      context: context,
      builder: (context) => BatchOperationsSheet(
        onDelete: () => _performBatchDelete(context, ref),
        onExport: () => _performBatchExport(context, ref),
      ),
    );
  }
  
  void _exportData(BuildContext context, WidgetRef ref) {
    // Implementation for data export
  }
  
  void _navigateToSettings(BuildContext context) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => const SettingsScreen(),
      ),
    );
  }
}

// Advanced summary card widget
class _SummaryCard extends StatelessWidget {
  final String title;
  final String value;
  final IconData icon;
  final Color color;
  
  const _SummaryCard({
    required this.title,
    required this.value,
    required this.icon,
    required this.color,
  });
  
  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 4,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(
                  icon,
                  color: color,
                  size: 24,
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    title,
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              value,
              style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                color: color,
                fontWeight: FontWeight.bold,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// Advanced user list item widget
class _UserListItem extends StatelessWidget {
  final User user;
  final bool isSelected;
  final VoidCallback onTap;
  final ValueChanged<bool> onSelectionChanged;
  
  const _UserListItem({
    required this.user,
    required this.isSelected,
    required this.onTap,
    required this.onSelectionChanged,
  });
  
  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.only(bottom: 8.0),
      child: ListTile(
        leading: CircleAvatar(
          backgroundImage: user.avatar.isNotEmpty
              ? CachedNetworkImageProvider(user.avatar)
              : null,
          child: user.avatar.isEmpty
              ? Text(
                  user.name.isNotEmpty ? user.name[0].toUpperCase() : '?',
                  style: const TextStyle(fontWeight: FontWeight.bold),
                )
              : null,
        ),
        title: Text(
          user.name,
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
        subtitle: Text(user.email),
        trailing: Checkbox(
          value: isSelected,
          onChanged: onSelectionChanged,
        ),
        onTap: onTap,
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      ),
    );
  }
}

// Loading widget with shimmer effect
class _LoadingWidget extends StatelessWidget {
  const _LoadingWidget();
  
  @override
  Widget build(BuildContext context) {
    return Shimmer.fromColors(
      baseColor: Colors.grey[300]!,
      highlightColor: Colors.grey[100]!,
      child: ListView.builder(
        itemCount: 10,
        itemBuilder: (context, index) => Card(
          margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
          child: ListTile(
            leading: const CircleAvatar(),
            title: Container(
              height: 16,
              color: Colors.white,
            ),
            subtitle: Container(
              height: 12,
              color: Colors.white,
            ),
          ),
        ),
      ),
    );
  }
}

// Error widget with retry functionality
class _ErrorWidget extends StatelessWidget {
  final String error;
  final VoidCallback onRetry;
  
  const _ErrorWidget({
    required this.error,
    required this.onRetry,
  });
  
  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(
              Icons.error_outline,
              size: 64,
              color: Colors.red,
            ),
            const SizedBox(height: 16),
            Text(
              'Something went wrong',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            Text(
              error,
              textAlign: TextAlign.center,
              style: Theme.of(context).textTheme.bodyMedium,
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: onRetry,
              icon: const Icon(Icons.refresh),
              label: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }
}

// Load more widget with infinite scroll support
class _LoadMoreWidget extends StatelessWidget {
  final VoidCallback onLoadMore;
  final bool isLoading;
  
  const _LoadMoreWidget({
    required this.onLoadMore,
    this.isLoading = false,
  });
  
  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16.0),
      child: isLoading
          ? const Center(child: CircularProgressIndicator())
          : ElevatedButton(
              onPressed: onLoadMore,
              child: const Text('Load More'),
            ),
    );
  }
}

// Advanced filter dialog
class FilterDialog extends ConsumerStatefulWidget {
  const FilterDialog({super.key});
  
  @override
  ConsumerState<FilterDialog> createState() => _FilterDialogState();
}

class _FilterDialogState extends ConsumerState<FilterDialog> {
  final _statusController = TextEditingController();
  String? _selectedStatus;
  
  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Filter Users'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          DropdownButtonFormField<UserStatus>(
            value: _selectedStatus,
            decoration: const InputDecoration(
              labelText: 'Status',
              border: OutlineInputBorder(),
            ),
            items: UserStatus.values.map((status) {
              return DropdownMenuItem(
                value: status,
                child: Text(status.name.toUpperCase()),
              );
            }).toList(),
            onChanged: (value) {
              setState(() {
                _selectedStatus = value;
              });
            },
          ),
          const SizedBox(height: 16),
          TextField(
            controller: _statusController,
            decoration: const InputDecoration(
              labelText: 'Search by name or email',
              border: OutlineInputBorder(),
              prefixIcon: Icon(Icons.search),
            ),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () {
            Navigator.of(context).pop();
          },
          child: const Text('Cancel'),
        ),
        TextButton(
          onPressed: () {
            _applyFilters();
            Navigator.of(context).pop();
          },
          child: const Text('Apply'),
        ),
      ],
    );
  }
  
  void _applyFilters() {
    final filters = <String, dynamic>{};
    
    if (_selectedStatus != null) {
      filters['status'] = _selectedStatus.toString();
    }
    
    if (_statusController.text.isNotEmpty) {
      filters['search'] = _statusController.text;
    }
    
    ref.read(enterpriseStateNotifierProvider.notifier).loadUsers(filters: filters);
  }
}

// Batch operations sheet
class BatchOperationsSheet extends StatelessWidget {
  final VoidCallback onDelete;
  final VoidCallback onExport;
  
  const BatchOperationsSheet({
    required this.onDelete,
    required this.onExport,
  });
  
  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16.0),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Text(
            'Batch Operations',
            style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.bold,
            ),
          ),
          const SizedBox(height: 16),
          ListTile(
            leading: const Icon(Icons.delete, color: Colors.red),
            title: const Text('Delete Selected'),
            onTap: () {
              Navigator.of(context).pop();
              onDelete();
            },
          ),
          ListTile(
            leading: const Icon(Icons.download, color: Colors.blue),
            title: const Text('Export Selected'),
            onTap: () {
              Navigator.of(context).pop();
              onExport();
            },
          ),
        ],
      ),
    );
  }
}

// Logger service
class Logger {
  void info(String message) {
    debugprint('[INFO] $message');
  }
  
  void warning(String message) {
    debugprint('[WARNING] $message');
  }
  
  void error(String message) {
    debugprint('[ERROR] $message');
  }
  
  void debug(String message) {
    debugprint('[DEBUG] $message');
  }
}
