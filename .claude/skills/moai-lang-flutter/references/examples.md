# Flutter/Dart Examples

Production-ready code examples for Flutter cross-platform development.

## Complete Feature: User Authentication

### Full Authentication Flow with Riverpod

```dart
// lib/features/auth/domain/entities/user.dart
class User {
  final String id;
  final String email;
  final String name;
  final String? avatarUrl;

  const User({
    required this.id,
    required this.email,
    required this.name,
    this.avatarUrl,
  });
}

class AuthTokens {
  final String accessToken;
  final String refreshToken;
  final DateTime expiresAt;

  const AuthTokens({
    required this.accessToken,
    required this.refreshToken,
    required this.expiresAt,
  });

  bool get isExpired => DateTime.now().isAfter(expiresAt);
}

// lib/features/auth/domain/errors/auth_error.dart
sealed class AuthError implements Exception {
  const AuthError();

  String get message => switch (this) {
    InvalidCredentialsError() => 'Invalid email or password',
    NetworkError(:final cause) => 'Network error: $cause',
    TokenExpiredError() => 'Session expired. Please login again',
    UnauthorizedError() => 'Unauthorized access',
    UnknownError(:final message) => message,
  };
}

class InvalidCredentialsError extends AuthError {
  const InvalidCredentialsError();
}

class NetworkError extends AuthError {
  final Object cause;
  const NetworkError(this.cause);
}

class TokenExpiredError extends AuthError {
  const TokenExpiredError();
}

class UnauthorizedError extends AuthError {
  const UnauthorizedError();
}

class UnknownError extends AuthError {
  @override
  final String message;
  const UnknownError(this.message);
}

// lib/features/auth/domain/repositories/auth_repository.dart
abstract class AuthRepository {
  Future<User> login(String email, String password);
  Future<void> logout();
  Future<User?> restoreSession();
  Stream<User?> watchAuthState();
}

// lib/features/auth/data/models/user_model.dart
class UserModel extends User {
  const UserModel({
    required super.id,
    required super.email,
    required super.name,
    super.avatarUrl,
  });

  factory UserModel.fromJson(Map<String, dynamic> json) => UserModel(
    id: json['id'] as String,
    email: json['email'] as String,
    name: json['name'] as String,
    avatarUrl: json['avatar_url'] as String?,
  );

  Map<String, dynamic> toJson() => {
    'id': id,
    'email': email,
    'name': name,
    if (avatarUrl != null) 'avatar_url': avatarUrl,
  };
}

class AuthTokensModel extends AuthTokens {
  const AuthTokensModel({
    required super.accessToken,
    required super.refreshToken,
    required super.expiresAt,
  });

  factory AuthTokensModel.fromJson(Map<String, dynamic> json) => AuthTokensModel(
    accessToken: json['access_token'] as String,
    refreshToken: json['refresh_token'] as String,
    expiresAt: DateTime.parse(json['expires_at'] as String),
  );

  Map<String, dynamic> toJson() => {
    'access_token': accessToken,
    'refresh_token': refreshToken,
    'expires_at': expiresAt.toIso8601String(),
  };
}

// lib/features/auth/data/datasources/auth_api.dart
class AuthApi {
  final Dio _dio;

  AuthApi(this._dio);

  Future<AuthTokensModel> login(String email, String password) async {
    final response = await _dio.post('/auth/login', data: {
      'email': email,
      'password': password,
    });
    return AuthTokensModel.fromJson(response.data);
  }

  Future<AuthTokensModel> refreshToken(String refreshToken) async {
    final response = await _dio.post('/auth/refresh', data: {
      'refresh_token': refreshToken,
    });
    return AuthTokensModel.fromJson(response.data);
  }

  Future<UserModel> getUser(String accessToken) async {
    final response = await _dio.get(
      '/auth/me',
      options: Options(headers: {'Authorization': 'Bearer $accessToken'}),
    );
    return UserModel.fromJson(response.data);
  }

  Future<void> logout(String accessToken) async {
    await _dio.post(
      '/auth/logout',
      options: Options(headers: {'Authorization': 'Bearer $accessToken'}),
    );
  }
}

// lib/features/auth/data/datasources/secure_storage.dart
class SecureStorage {
  static const _tokensKey = 'auth_tokens';
  final FlutterSecureStorage _storage;

  SecureStorage(this._storage);

  Future<AuthTokensModel?> getTokens() async {
    final json = await _storage.read(key: _tokensKey);
    if (json == null) return null;
    return AuthTokensModel.fromJson(jsonDecode(json));
  }

  Future<void> saveTokens(AuthTokensModel tokens) async {
    await _storage.write(
      key: _tokensKey,
      value: jsonEncode(tokens.toJson()),
    );
  }

  Future<void> clearTokens() async {
    await _storage.delete(key: _tokensKey);
  }
}

// lib/features/auth/data/repositories/auth_repository_impl.dart
class AuthRepositoryImpl implements AuthRepository {
  final AuthApi _api;
  final SecureStorage _storage;
  final _authStateController = StreamController<User?>.broadcast();

  AuthRepositoryImpl(this._api, this._storage);

  @override
  Stream<User?> watchAuthState() => _authStateController.stream;

  @override
  Future<User> login(String email, String password) async {
    try {
      final tokens = await _api.login(email, password);
      await _storage.saveTokens(tokens);

      final user = await _api.getUser(tokens.accessToken);
      _authStateController.add(user);

      return user;
    } on DioException catch (e) {
      throw _mapDioError(e);
    }
  }

  @override
  Future<void> logout() async {
    try {
      final tokens = await _storage.getTokens();
      if (tokens != null) {
        await _api.logout(tokens.accessToken).catchError((_) {});
      }
    } finally {
      await _storage.clearTokens();
      _authStateController.add(null);
    }
  }

  @override
  Future<User?> restoreSession() async {
    final tokens = await _storage.getTokens();
    if (tokens == null) return null;

    try {
      AuthTokensModel activeTokens = tokens;

      if (tokens.isExpired) {
        activeTokens = await _api.refreshToken(tokens.refreshToken);
        await _storage.saveTokens(activeTokens);
      }

      final user = await _api.getUser(activeTokens.accessToken);
      _authStateController.add(user);
      return user;
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        await _storage.clearTokens();
        _authStateController.add(null);
        return null;
      }
      throw _mapDioError(e);
    }
  }

  AuthError _mapDioError(DioException e) {
    if (e.type == DioExceptionType.connectionError ||
        e.type == DioExceptionType.connectionTimeout) {
      return NetworkError(e);
    }

    final statusCode = e.response?.statusCode;
    return switch (statusCode) {
      401 => const InvalidCredentialsError(),
      403 => const UnauthorizedError(),
      _ => UnknownError(e.message ?? 'Unknown error'),
    };
  }
}

// lib/features/auth/presentation/providers/auth_provider.dart
part 'auth_provider.g.dart';

@riverpod
AuthRepository authRepository(Ref ref) {
  return AuthRepositoryImpl(
    ref.read(authApiProvider),
    ref.read(secureStorageProvider),
  );
}

@riverpod
Stream<User?> authState(Ref ref) {
  return ref.watch(authRepositoryProvider).watchAuthState();
}

@riverpod
class AuthController extends _$AuthController {
  @override
  FutureOr<User?> build() async {
    return ref.read(authRepositoryProvider).restoreSession();
  }

  Future<void> login(String email, String password) async {
    state = const AsyncLoading();
    state = await AsyncValue.guard(
      () => ref.read(authRepositoryProvider).login(email, password),
    );
  }

  Future<void> logout() async {
    await ref.read(authRepositoryProvider).logout();
    state = const AsyncData(null);
  }
}

// lib/features/auth/presentation/screens/login_screen.dart
class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _formKey = GlobalKey<FormState>();
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  bool _obscurePassword = true;

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _handleLogin() async {
    if (!_formKey.currentState!.validate()) return;

    await ref.read(authControllerProvider.notifier).login(
      _emailController.text.trim(),
      _passwordController.text,
    );
  }

  @override
  Widget build(BuildContext context) {
    final authState = ref.watch(authControllerProvider);
    final theme = Theme.of(context);

    // Listen for auth state changes
    ref.listen(authControllerProvider, (prev, next) {
      next.whenOrNull(
        data: (user) {
          if (user != null) {
            context.go('/home');
          }
        },
        error: (error, _) {
          final message = error is AuthError
              ? error.message
              : 'An unexpected error occurred';

          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(message),
              backgroundColor: theme.colorScheme.error,
            ),
          );
        },
      );
    });

    return Scaffold(
      body: SafeArea(
        child: SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const SizedBox(height: 48),

                // Header
                Text(
                  'Welcome Back',
                  style: theme.textTheme.headlineLarge,
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 8),
                Text(
                  'Sign in to continue',
                  style: theme.textTheme.bodyLarge?.copyWith(
                    color: theme.colorScheme.onSurfaceVariant,
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 48),

                // Email field
                TextFormField(
                  controller: _emailController,
                  keyboardType: TextInputType.emailAdddess,
                  textInputAction: TextInputAction.next,
                  decoration: const InputDecoration(
                    labelText: 'Email',
                    prefixIcon: Icon(Icons.email_outlined),
                    border: OutlineInputBorder(),
                  ),
                  validator: (value) {
                    if (value == null || value.isEmpty) {
                      return 'Please enter your email';
                    }
                    if (!value.contains('@') || !value.contains('.')) {
                      return 'Please enter a valid email';
                    }
                    return null;
                  },
                ),
                const SizedBox(height: 16),

                // Password field
                TextFormField(
                  controller: _passwordController,
                  obscureText: _obscurePassword,
                  textInputAction: TextInputAction.done,
                  onFieldSubmitted: (_) => _handleLogin(),
                  decoration: InputDecoration(
                    labelText: 'Password',
                    prefixIcon: const Icon(Icons.lock_outlined),
                    suffixIcon: IconButton(
                      icon: Icon(
                        _obscurePassword
                            ? Icons.visibility_outlined
                            : Icons.visibility_off_outlined,
                      ),
                      onPressed: () {
                        setState(() => _obscurePassword = !_obscurePassword);
                      },
                    ),
                    border: const OutlineInputBorder(),
                  ),
                  validator: (value) {
                    if (value == null || value.isEmpty) {
                      return 'Please enter your password';
                    }
                    if (value.length < 6) {
                      return 'Password must be at least 6 characters';
                    }
                    return null;
                  },
                ),
                const SizedBox(height: 8),

                // Forgot password
                Align(
                  alignment: Alignment.centerRight,
                  child: TextButton(
                    onPressed: () => context.push('/forgot-password'),
                    child: const Text('Forgot Password?'),
                  ),
                ),
                const SizedBox(height: 24),

                // Login button
                FilledButton(
                  onPressed: authState.isLoading ? null : _handleLogin,
                  style: FilledButton.styleFrom(
                    minimumSize: const Size.fromHeight(56),
                  ),
                  child: authState.isLoading
                      ? const SizedBox(
                          height: 24,
                          width: 24,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Sign In'),
                ),
                const SizedBox(height: 24),

                // Register link
                Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      "Don't have an account?",
                      style: theme.textTheme.bodyMedium,
                    ),
                    TextButton(
                      onPressed: () => context.push('/register'),
                      child: const Text('Sign Up'),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
```

## Network Layer

### Dio Client with Interceptors

```dart
// lib/core/network/api_client.dart
class ApiClient {
  late final Dio _dio;
  final SecureStorage _storage;

  ApiClient({
    required String baseUrl,
    required SecureStorage storage,
  }) : _storage = storage {
    _dio = Dio(BaseOptions(
      baseUrl: baseUrl,
      connectTimeout: const Duration(seconds: 30),
      receiveTimeout: const Duration(seconds: 30),
      headers: {
        'Content-Type': 'application/json',
        'Accept': 'application/json',
      },
    ));

    _dio.interceptors.addAll([
      AuthInterceptor(_storage, _dio),
      LogInterceptor(
        requestBody: true,
        responseBody: true,
        logPrint: (log) => debugPrint(log.toString()),
      ),
      RetryInterceptor(_dio),
    ]);
  }

  Future<T> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
    required T Function(dynamic) fromJson,
  }) async {
    final response = await _dio.get(path, queryParameters: queryParameters);
    return fromJson(response.data);
  }

  Future<T> post<T>(
    String path, {
    dynamic data,
    required T Function(dynamic) fromJson,
  }) async {
    final response = await _dio.post(path, data: data);
    return fromJson(response.data);
  }

  Future<T> put<T>(
    String path, {
    dynamic data,
    required T Function(dynamic) fromJson,
  }) async {
    final response = await _dio.put(path, data: data);
    return fromJson(response.data);
  }

  Future<void> delete(String path) async {
    await _dio.delete(path);
  }
}

// lib/core/network/interceptors/auth_interceptor.dart
class AuthInterceptor extends Interceptor {
  final SecureStorage _storage;
  final Dio _dio;
  bool _isRefreshing = false;
  final _pendingRequests = <({RequestOptions options, ErrorInterceptorHandler handler})>[];

  AuthInterceptor(this._storage, this._dio);

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final tokens = await _storage.getTokens();
    if (tokens != null) {
      options.headers['Authorization'] = 'Bearer ${tokens.accessToken}';
    }
    handler.next(options);
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    if (err.response?.statusCode != 401) {
      return handler.next(err);
    }

    // Queue the request if already refreshing
    if (_isRefreshing) {
      _pendingRequests.add((options: err.requestOptions, handler: handler));
      return;
    }

    _isRefreshing = true;

    try {
      final tokens = await _storage.getTokens();
      if (tokens == null) {
        return handler.next(err);
      }

      // Refresh token
      final response = await _dio.post(
        '/auth/refresh',
        data: {'refresh_token': tokens.refreshToken},
        options: Options(headers: {'Authorization': null}),
      );

      final newTokens = AuthTokensModel.fromJson(response.data);
      await _storage.saveTokens(newTokens);

      // Retry original request
      final retryResponse = await _retryRequest(err.requestOptions, newTokens);
      handler.resolve(retryResponse);

      // Retry pending requests
      for (final pending in _pendingRequests) {
        final response = await _retryRequest(pending.options, newTokens);
        pending.handler.resolve(response);
      }
    } on DioException catch (e) {
      // Refresh failed - clear tokens and reject all pending
      await _storage.clearTokens();
      handler.next(e);

      for (final pending in _pendingRequests) {
        pending.handler.next(e);
      }
    } finally {
      _isRefreshing = false;
      _pendingRequests.clear();
    }
  }

  Future<Response> _retryRequest(
    RequestOptions options,
    AuthTokensModel tokens,
  ) async {
    options.headers['Authorization'] = 'Bearer ${tokens.accessToken}';
    return _dio.fetch(options);
  }
}

// lib/core/network/interceptors/retry_interceptor.dart
class RetryInterceptor extends Interceptor {
  final Dio _dio;
  final int maxRetries;
  final Duration retryDelay;

  RetryInterceptor(
    this._dio, {
    this.maxRetries = 3,
    this.retryDelay = const Duration(seconds: 1),
  });

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    final shouldRetry = _shouldRetry(err);
    final retryCount = err.requestOptions.extra['retryCount'] ?? 0;

    if (!shouldRetry || retryCount >= maxRetries) {
      return handler.next(err);
    }

    await Future.delayed(retryDelay * (retryCount + 1));

    try {
      err.requestOptions.extra['retryCount'] = retryCount + 1;
      final response = await _dio.fetch(err.requestOptions);
      handler.resolve(response);
    } on DioException catch (e) {
      handler.next(e);
    }
  }

  bool _shouldRetry(DioException err) {
    return err.type == DioExceptionType.connectionTimeout ||
        err.type == DioExceptionType.sendTimeout ||
        err.type == DioExceptionType.receiveTimeout ||
        (err.response?.statusCode != null &&
            err.response!.statusCode! >= 500);
  }
}
```

## Platform Channels

### Complete Native Bridge

```dart
// lib/core/platform/native_bridge.dart
class NativeBridge {
  static const _methodChannel = MethodChannel('com.example.app/native');
  static const _eventChannel = EventChannel('com.example.app/events');

  // Singleton pattern
  static final NativeBridge _instance = NativeBridge._();
  static NativeBridge get instance => _instance;
  NativeBridge._() {
    _setupMethodCallHandler();
  }

  // Method calls
  Future<String> getPlatformVersion() async {
    try {
      final version = await _methodChannel.invokeMethod<String>('getPlatformVersion');
      return version ?? 'Unknown';
    } on PlatformException catch (e) {
      throw NativeBridgeException('Failed to get platform version: ${e.message}');
    }
  }

  Future<DeviceInfo> getDeviceInfo() async {
    try {
      final result = await _methodChannel.invokeMapMethod<String, dynamic>('getDeviceInfo');
      return DeviceInfo.fromMap(result ?? {});
    } on PlatformException catch (e) {
      throw NativeBridgeException('Failed to get device info: ${e.message}');
    }
  }

  Future<void> shareContent({
    required String text,
    String? title,
    String? url,
  }) async {
    try {
      await _methodChannel.invokeMethod('share', {
        'text': text,
        if (title != null) 'title': title,
        if (url != null) 'url': url,
      });
    } on PlatformException catch (e) {
      throw NativeBridgeException('Failed to share: ${e.message}');
    }
  }

  Future<bool> requestPermission(PermissionType type) async {
    try {
      final result = await _methodChannel.invokeMethod<bool>(
        'requestPermission',
        {'type': type.name},
      );
      return result ?? false;
    } on PlatformException catch (e) {
      throw NativeBridgeException('Failed to request permission: ${e.message}');
    }
  }

  // Event streams
  Stream<BatteryState> watchBatteryState() {
    return _eventChannel.receiveBroadcastStream('battery').map((event) {
      final data = event as Map<dynamic, dynamic>;
      return BatteryState(
        level: data['level'] as int,
        isCharging: data['isCharging'] as bool,
      );
    });
  }

  Stream<ConnectivityState> watchConnectivity() {
    return _eventChannel.receiveBroadcastStream('connectivity').map((event) {
      final data = event as Map<dynamic, dynamic>;
      return ConnectivityState(
        isConnected: data['isConnected'] as bool,
        type: ConnectivityType.values.byName(data['type'] as String),
      );
    });
  }

  // Bidirectional communication
  void _setupMethodCallHandler() {
    _methodChannel.setMethodCallHandler((call) async {
      switch (call.method) {
        case 'onDeepLink':
          final url = call.arguments as String;
          _handleDeepLink(url);
          return true;
        case 'onPushNotification':
          final data = call.arguments as Map<dynamic, dynamic>;
          _handlePushNotification(data.cast<String, dynamic>());
          return true;
        default:
          throw MissingPluginException('Method not implemented: ${call.method}');
      }
    });
  }

  void _handleDeepLink(String url) {
    // Handle deep link
    debugPrint('Received deep link: $url');
  }

  void _handlePushNotification(Map<String, dynamic> data) {
    // Handle push notification
    debugPrint('Received notification: $data');
  }
}

// Models
class DeviceInfo {
  final String platform;
  final String version;
  final String model;
  final String manufacturer;

  const DeviceInfo({
    required this.platform,
    required this.version,
    required this.model,
    required this.manufacturer,
  });

  factory DeviceInfo.fromMap(Map<String, dynamic> map) => DeviceInfo(
    platform: map['platform'] as String? ?? 'unknown',
    version: map['version'] as String? ?? 'unknown',
    model: map['model'] as String? ?? 'unknown',
    manufacturer: map['manufacturer'] as String? ?? 'unknown',
  );
}

class BatteryState {
  final int level;
  final bool isCharging;
  const BatteryState({required this.level, required this.isCharging});
}

class ConnectivityState {
  final bool isConnected;
  final ConnectivityType type;
  const ConnectivityState({required this.isConnected, required this.type});
}

enum ConnectivityType { none, wifi, mobile, ethernet }
enum PermissionType { camera, microphone, location, storage, notifications }

class NativeBridgeException implements Exception {
  final String message;
  const NativeBridgeException(this.message);

  @override
  String toString() => 'NativeBridgeException: $message';
}
```

## Testing Examples

### Widget Tests with Riverpod

```dart
// test/features/auth/presentation/screens/login_screen_test.dart
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:mocktail/mocktail.dart';

class MockAuthRepository extends Mock implements AuthRepository {}

void main() {
  late ProviderContainer container;
  late MockAuthRepository mockAuthRepository;

  setUp(() {
    mockAuthRepository = MockAuthRepository();

    // Default mock behavior
    when(() => mockAuthRepository.restoreSession())
        .thenAnswer((_) async => null);
    when(() => mockAuthRepository.watchAuthState())
        .thenAnswer((_) => const Stream.empty());

    container = ProviderContainer(overrides: [
      authRepositoryProvider.overrideWithValue(mockAuthRepository),
    ]);
  });

  tearDown(() => container.dispose());

  Widget buildTestWidget() {
    return UncontrolledProviderScope(
      container: container,
      child: MaterialApp(
        home: const LoginScreen(),
      ),
    );
  }

  group('LoginScreen', () {
    testWidgets('renders email and password fields', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      expect(find.byType(TextFormField), findsNWidgets(2));
      expect(find.text('Email'), findsOneWidget);
      expect(find.text('Password'), findsOneWidget);
    });

    testWidgets('shows validation errors for empty fields', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      await tester.tap(find.byType(FilledButton));
      await tester.pump();

      expect(find.text('Please enter your email'), findsOneWidget);
      expect(find.text('Please enter your password'), findsOneWidget);
    });

    testWidgets('shows validation error for invalid email', (tester) async {
      await tester.pumpWidget(buildTestWidget());

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Email'),
        'invalid-email',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Password'),
        'password123',
      );

      await tester.tap(find.byType(FilledButton));
      await tester.pump();

      expect(find.text('Please enter a valid email'), findsOneWidget);
    });

    testWidgets('calls login when form is valid', (tester) async {
      when(() => mockAuthRepository.login(any(), any()))
          .thenAnswer((_) async => const User(
                id: '1',
                email: 'test@example.com',
                name: 'Test User',
              ));

      await tester.pumpWidget(buildTestWidget());

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Email'),
        'test@example.com',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Password'),
        'password123',
      );

      await tester.tap(find.byType(FilledButton));
      await tester.pump();

      verify(() => mockAuthRepository.login('test@example.com', 'password123'))
          .called(1);
    });

    testWidgets('shows loading indicator during login', (tester) async {
      when(() => mockAuthRepository.login(any(), any()))
          .thenAnswer((_) async {
        await Future.delayed(const Duration(seconds: 2));
        return const User(
          id: '1',
          email: 'test@example.com',
          name: 'Test User',
        );
      });

      await tester.pumpWidget(buildTestWidget());

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Email'),
        'test@example.com',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Password'),
        'password123',
      );

      await tester.tap(find.byType(FilledButton));
      await tester.pump();

      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });

    testWidgets('shows error snackbar on login failure', (tester) async {
      when(() => mockAuthRepository.login(any(), any()))
          .thenThrow(const InvalidCredentialsError());

      await tester.pumpWidget(buildTestWidget());

      await tester.enterText(
        find.widgetWithText(TextFormField, 'Email'),
        'test@example.com',
      );
      await tester.enterText(
        find.widgetWithText(TextFormField, 'Password'),
        'wrongpassword',
      );

      await tester.tap(find.byType(FilledButton));
      await tester.pumpAndSettle();

      expect(find.byType(SnackBar), findsOneWidget);
      expect(find.text('Invalid email or password'), findsOneWidget);
    });
  });
}
```

### Integration Tests

```dart
// integration_test/auth_flow_test.dart
import 'package:flutter_test/flutter_test.dart';
import 'package:integration_test/integration_test.dart';

void main() {
  IntegrationTestWidgetsFlutterBinding.ensureInitialized();

  group('Authentication Flow', () {
    testWidgets('complete login and logout flow', (tester) async {
      await tester.pumpWidget(const MyApp());
      await tester.pumpAndSettle();

      // Verify we're on login screen
      expect(find.text('Welcome Back'), findsOneWidget);

      // Enter credentials
      await tester.enterText(
        find.byKey(const Key('email_field')),
        'test@example.com',
      );
      await tester.enterText(
        find.byKey(const Key('password_field')),
        'password123',
      );

      // Tap login
      await tester.tap(find.byKey(const Key('login_button')));
      await tester.pumpAndSettle();

      // Verify navigation to home
      expect(find.byType(HomeScreen), findsOneWidget);
      expect(find.text('Test User'), findsOneWidget);

      // Navigate to profile
      await tester.tap(find.byIcon(Icons.person));
      await tester.pumpAndSettle();

      // Tap logout
      await tester.tap(find.text('Logout'));
      await tester.pumpAndSettle();

      // Verify back on login screen
      expect(find.text('Welcome Back'), findsOneWidget);
    });

    testWidgets('session restoration on app restart', (tester) async {
      // First launch - login
      await tester.pumpWidget(const MyApp());
      await tester.pumpAndSettle();

      await tester.enterText(
        find.byKey(const Key('email_field')),
        'test@example.com',
      );
      await tester.enterText(
        find.byKey(const Key('password_field')),
        'password123',
      );
      await tester.tap(find.byKey(const Key('login_button')));
      await tester.pumpAndSettle();

      expect(find.byType(HomeScreen), findsOneWidget);

      // Simulate app restart
      await tester.pumpWidget(const MyApp());
      await tester.pumpAndSettle();

      // Should restore session and show home
      expect(find.byType(HomeScreen), findsOneWidget);
    });
  });
}
```

---

Version: 1.0.0
Last Updated: 2025-12-07
