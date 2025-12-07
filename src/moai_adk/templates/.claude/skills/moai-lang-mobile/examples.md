# Mobile Development Examples

Production-ready code examples for iOS, Android, and Flutter development.

## Complete Feature: User Authentication

### Swift 6 (iOS) - Full Authentication Flow

```swift
// MARK: - Domain Layer

/// Authentication errors with typed throws
enum AuthError: Error, LocalizedError {
    case invalidCredentials
    case networkError(underlying: Error)
    case tokenExpired
    case unauthorized

    var errorDescription: String? {
        switch self {
        case .invalidCredentials: return "Invalid email or password"
        case .networkError(let error): return "Network error: \(error.localizedDescription)"
        case .tokenExpired: return "Session expired. Please login again"
        case .unauthorized: return "Unauthorized access"
        }
    }
}

struct User: Codable, Identifiable, Sendable {
    let id: String
    let email: String
    let name: String
    let avatarURL: URL?
}

struct AuthTokens: Codable, Sendable {
    let accessToken: String
    let refreshToken: String
    let expiresAt: Date
}

// MARK: - Data Layer

protocol AuthAPIProtocol: Sendable {
    func login(email: String, password: String) async throws(AuthError) -> AuthTokens
    func refreshToken(_ token: String) async throws(AuthError) -> AuthTokens
    func fetchUser(token: String) async throws(AuthError) -> User
    func logout(token: String) async throws(AuthError)
}

actor AuthAPI: AuthAPIProtocol {
    private let session: URLSession
    private let baseURL: URL

    init(baseURL: URL, session: URLSession = .shared) {
        self.baseURL = baseURL
        self.session = session
    }

    func login(email: String, password: String) async throws(AuthError) -> AuthTokens {
        let request = try makeRequest(
            path: "/auth/login",
            method: "POST",
            body: ["email": email, "password": password]
        )

        do {
            let (data, response) = try await session.data(for: request)
            guard let httpResponse = response as? HTTPURLResponse else {
                throw AuthError.networkError(underlying: URLError(.badServerResponse))
            }

            switch httpResponse.statusCode {
            case 200:
                return try JSONDecoder().decode(AuthTokens.self, from: data)
            case 401:
                throw AuthError.invalidCredentials
            default:
                throw AuthError.networkError(underlying: URLError(.badServerResponse))
            }
        } catch let error as AuthError {
            throw error
        } catch {
            throw AuthError.networkError(underlying: error)
        }
    }

    func refreshToken(_ token: String) async throws(AuthError) -> AuthTokens {
        var request = try makeRequest(path: "/auth/refresh", method: "POST")
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        do {
            let (data, response) = try await session.data(for: request)
            guard let httpResponse = response as? HTTPURLResponse,
                  httpResponse.statusCode == 200 else {
                throw AuthError.tokenExpired
            }
            return try JSONDecoder().decode(AuthTokens.self, from: data)
        } catch let error as AuthError {
            throw error
        } catch {
            throw AuthError.networkError(underlying: error)
        }
    }

    func fetchUser(token: String) async throws(AuthError) -> User {
        var request = try makeRequest(path: "/auth/me", method: "GET")
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        do {
            let (data, response) = try await session.data(for: request)
            guard let httpResponse = response as? HTTPURLResponse else {
                throw AuthError.networkError(underlying: URLError(.badServerResponse))
            }

            switch httpResponse.statusCode {
            case 200:
                return try JSONDecoder().decode(User.self, from: data)
            case 401:
                throw AuthError.unauthorized
            default:
                throw AuthError.networkError(underlying: URLError(.badServerResponse))
            }
        } catch let error as AuthError {
            throw error
        } catch {
            throw AuthError.networkError(underlying: error)
        }
    }

    func logout(token: String) async throws(AuthError) {
        var request = try makeRequest(path: "/auth/logout", method: "POST")
        request.setValue("Bearer \(token)", forHTTPHeaderField: "Authorization")

        do {
            let (_, response) = try await session.data(for: request)
            guard let httpResponse = response as? HTTPURLResponse,
                  httpResponse.statusCode == 200 else {
                throw AuthError.networkError(underlying: URLError(.badServerResponse))
            }
        } catch let error as AuthError {
            throw error
        } catch {
            throw AuthError.networkError(underlying: error)
        }
    }

    private func makeRequest(path: String, method: String, body: [String: Any]? = nil) throws -> URLRequest {
        var request = URLRequest(url: baseURL.appendingPathComponent(path))
        request.httpMethod = method
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")

        if let body = body {
            request.httpBody = try JSONSerialization.data(withJSONObject: body)
        }

        return request
    }
}

// MARK: - Presentation Layer

@Observable
@MainActor
final class AuthViewModel {
    private let api: AuthAPIProtocol
    private let keychain: KeychainProtocol

    var user: User?
    var isLoading = false
    var error: AuthError?
    var isAuthenticated: Bool { user != nil }

    init(api: AuthAPIProtocol, keychain: KeychainProtocol) {
        self.api = api
        self.keychain = keychain
    }

    func login(email: String, password: String) async {
        isLoading = true
        error = nil
        defer { isLoading = false }

        do {
            let tokens = try await api.login(email: email, password: password)
            try keychain.save(tokens: tokens)

            let user = try await api.fetchUser(token: tokens.accessToken)
            self.user = user
        } catch {
            self.error = error
        }
    }

    func restoreSession() async {
        guard let tokens = try? keychain.loadTokens(),
              tokens.expiresAt > Date() else {
            return
        }

        do {
            user = try await api.fetchUser(token: tokens.accessToken)
        } catch AuthError.unauthorized, AuthError.tokenExpired {
            do {
                let newTokens = try await api.refreshToken(tokens.refreshToken)
                try keychain.save(tokens: newTokens)
                user = try await api.fetchUser(token: newTokens.accessToken)
            } catch {
                try? keychain.deleteTokens()
            }
        } catch {
            self.error = error
        }
    }

    func logout() async {
        guard let tokens = try? keychain.loadTokens() else { return }

        do {
            try await api.logout(token: tokens.accessToken)
        } catch {
            // Log error but continue with local logout
        }

        try? keychain.deleteTokens()
        user = nil
    }
}

// MARK: - SwiftUI Views

struct LoginView: View {
    @Environment(AuthViewModel.self) private var viewModel
    @State private var email = ""
    @State private var password = ""

    var body: some View {
        NavigationStack {
            Form {
                Section {
                    TextField("Email", text: $email)
                        .textContentType(.emailAddress)
                        .keyboardType(.emailAddress)
                        .autocapitalization(.none)

                    SecureField("Password", text: $password)
                        .textContentType(.password)
                }

                if let error = viewModel.error {
                    Section {
                        Text(error.localizedDescription)
                            .foregroundColor(.red)
                    }
                }

                Section {
                    Button {
                        Task { await viewModel.login(email: email, password: password) }
                    } label: {
                        if viewModel.isLoading {
                            ProgressView()
                                .frame(maxWidth: .infinity)
                        } else {
                            Text("Sign In")
                                .frame(maxWidth: .infinity)
                        }
                    }
                    .disabled(email.isEmpty || password.isEmpty || viewModel.isLoading)
                }
            }
            .navigationTitle("Login")
        }
    }
}
```

### Kotlin 2.0 (Android) - Full Authentication Flow

```kotlin
// Domain Layer
data class User(
    val id: String,
    val email: String,
    val name: String,
    val avatarUrl: String?
)

data class AuthTokens(
    val accessToken: String,
    val refreshToken: String,
    val expiresAt: Long
)

sealed class AuthError : Exception() {
    data object InvalidCredentials : AuthError()
    data class NetworkError(override val cause: Throwable) : AuthError()
    data object TokenExpired : AuthError()
    data object Unauthorized : AuthError()
}

// Repository Interface
interface AuthRepository {
    suspend fun login(email: String, password: String): Result<User>
    suspend fun logout(): Result<Unit>
    suspend fun refreshSession(): Result<User>
    fun observeAuthState(): Flow<User?>
}

// Repository Implementation
class AuthRepositoryImpl @Inject constructor(
    private val authApi: AuthApi,
    private val tokenStorage: TokenStorage,
    private val userDao: UserDao
) : AuthRepository {

    private val _authState = MutableStateFlow<User?>(null)

    override fun observeAuthState(): Flow<User?> = _authState.asStateFlow()

    override suspend fun login(email: String, password: String): Result<User> {
        return runCatching {
            val tokens = authApi.login(LoginRequest(email, password))
            tokenStorage.saveTokens(tokens)

            val user = authApi.getUser("Bearer ${tokens.accessToken}")
            userDao.insertUser(user.toEntity())
            _authState.value = user

            user
        }.recoverCatching { error ->
            when {
                error is HttpException && error.code() == 401 ->
                    throw AuthError.InvalidCredentials
                error is IOException ->
                    throw AuthError.NetworkError(error)
                else -> throw error
            }
        }
    }

    override suspend fun logout(): Result<Unit> {
        return runCatching {
            tokenStorage.getTokens()?.let { tokens ->
                runCatching { authApi.logout("Bearer ${tokens.accessToken}") }
            }
            tokenStorage.clearTokens()
            userDao.clearUser()
            _authState.value = null
        }
    }

    override suspend fun refreshSession(): Result<User> {
        return runCatching {
            val tokens = tokenStorage.getTokens() ?: throw AuthError.Unauthorized

            if (tokens.expiresAt < System.currentTimeMillis()) {
                val newTokens = authApi.refreshToken(
                    "Bearer ${tokens.refreshToken}"
                )
                tokenStorage.saveTokens(newTokens)

                val user = authApi.getUser("Bearer ${newTokens.accessToken}")
                userDao.insertUser(user.toEntity())
                _authState.value = user
                user
            } else {
                val user = authApi.getUser("Bearer ${tokens.accessToken}")
                _authState.value = user
                user
            }
        }.recoverCatching { error ->
            when (error) {
                is HttpException -> when (error.code()) {
                    401 -> {
                        tokenStorage.clearTokens()
                        throw AuthError.TokenExpired
                    }
                    else -> throw AuthError.NetworkError(error)
                }
                is IOException -> throw AuthError.NetworkError(error)
                else -> throw error
            }
        }
    }
}

// ViewModel
@HiltViewModel
class AuthViewModel @Inject constructor(
    private val authRepository: AuthRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow<AuthUiState>(AuthUiState.Initial)
    val uiState: StateFlow<AuthUiState> = _uiState.asStateFlow()

    val isAuthenticated: StateFlow<Boolean> = authRepository.observeAuthState()
        .map { it != null }
        .stateIn(viewModelScope, SharingStarted.WhileSubscribed(5000), false)

    init {
        restoreSession()
    }

    private fun restoreSession() {
        viewModelScope.launch {
            _uiState.value = AuthUiState.Loading
            authRepository.refreshSession()
                .onSuccess { user ->
                    _uiState.value = AuthUiState.Authenticated(user)
                }
                .onFailure {
                    _uiState.value = AuthUiState.Initial
                }
        }
    }

    fun login(email: String, password: String) {
        viewModelScope.launch {
            _uiState.value = AuthUiState.Loading
            authRepository.login(email, password)
                .onSuccess { user ->
                    _uiState.value = AuthUiState.Authenticated(user)
                }
                .onFailure { error ->
                    _uiState.value = AuthUiState.Error(
                        when (error) {
                            is AuthError.InvalidCredentials -> "Invalid email or password"
                            is AuthError.NetworkError -> "Network error. Please try again."
                            else -> "An unexpected error occurred"
                        }
                    )
                }
        }
    }

    fun logout() {
        viewModelScope.launch {
            authRepository.logout()
            _uiState.value = AuthUiState.Initial
        }
    }
}

sealed interface AuthUiState {
    data object Initial : AuthUiState
    data object Loading : AuthUiState
    data class Authenticated(val user: User) : AuthUiState
    data class Error(val message: String) : AuthUiState
}

// Compose UI
@Composable
fun LoginScreen(
    viewModel: AuthViewModel = hiltViewModel(),
    onNavigateToHome: () -> Unit
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()

    var email by rememberSaveable { mutableStateOf("") }
    var password by rememberSaveable { mutableStateOf("") }

    LaunchedEffect(uiState) {
        if (uiState is AuthUiState.Authenticated) {
            onNavigateToHome()
        }
    }

    Scaffold { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp),
            verticalArrangement = Arrangement.Center
        ) {
            Text(
                text = "Welcome Back",
                style = MaterialTheme.typography.headlineLarge
            )

            Spacer(modifier = Modifier.height(32.dp))

            OutlinedTextField(
                value = email,
                onValueChange = { email = it },
                label = { Text("Email") },
                keyboardOptions = KeyboardOptions(
                    keyboardType = KeyboardType.Email,
                    imeAction = ImeAction.Next
                ),
                modifier = Modifier.fillMaxWidth()
            )

            Spacer(modifier = Modifier.height(16.dp))

            OutlinedTextField(
                value = password,
                onValueChange = { password = it },
                label = { Text("Password") },
                visualTransformation = PasswordVisualTransformation(),
                keyboardOptions = KeyboardOptions(
                    keyboardType = KeyboardType.Password,
                    imeAction = ImeAction.Done
                ),
                modifier = Modifier.fillMaxWidth()
            )

            if (uiState is AuthUiState.Error) {
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = (uiState as AuthUiState.Error).message,
                    color = MaterialTheme.colorScheme.error,
                    style = MaterialTheme.typography.bodySmall
                )
            }

            Spacer(modifier = Modifier.height(24.dp))

            Button(
                onClick = { viewModel.login(email, password) },
                enabled = email.isNotBlank() && password.isNotBlank() &&
                        uiState !is AuthUiState.Loading,
                modifier = Modifier.fillMaxWidth()
            ) {
                if (uiState is AuthUiState.Loading) {
                    CircularProgressIndicator(
                        modifier = Modifier.size(24.dp),
                        color = MaterialTheme.colorScheme.onPrimary
                    )
                } else {
                    Text("Sign In")
                }
            }
        }
    }
}
```

### Dart 3.5 / Flutter 3.24 - Full Authentication Flow

```dart
// Domain Layer
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

  factory User.fromJson(Map<String, dynamic> json) => User(
    id: json['id'] as String,
    email: json['email'] as String,
    name: json['name'] as String,
    avatarUrl: json['avatar_url'] as String?,
  );
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

  factory AuthTokens.fromJson(Map<String, dynamic> json) => AuthTokens(
    accessToken: json['access_token'] as String,
    refreshToken: json['refresh_token'] as String,
    expiresAt: DateTime.parse(json['expires_at'] as String),
  );

  bool get isExpired => DateTime.now().isAfter(expiresAt);
}

sealed class AuthError implements Exception {
  const AuthError();
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

// Data Layer
abstract class AuthRepository {
  Future<User> login(String email, String password);
  Future<void> logout();
  Future<User?> restoreSession();
  Stream<User?> watchAuthState();
}

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
      if (e.response?.statusCode == 401) {
        throw const InvalidCredentialsError();
      }
      throw NetworkError(e);
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
      AuthTokens activeTokens = tokens;

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
      throw NetworkError(e);
    }
  }
}

// Riverpod Providers
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

// Flutter UI
class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _emailController = TextEditingController();
  final _passwordController = TextEditingController();
  final _formKey = GlobalKey<FormState>();

  @override
  void dispose() {
    _emailController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _handleLogin() async {
    if (!_formKey.currentState!.validate()) return;

    await ref.read(authControllerProvider.notifier).login(
      _emailController.text,
      _passwordController.text,
    );
  }

  @override
  Widget build(BuildContext context) {
    final authState = ref.watch(authControllerProvider);

    ref.listen(authControllerProvider, (_, state) {
      state.whenOrNull(
        data: (user) {
          if (user != null) {
            context.go('/home');
          }
        },
        error: (error, _) {
          final message = switch (error) {
            InvalidCredentialsError() => 'Invalid email or password',
            NetworkError() => 'Network error. Please try again.',
            _ => 'An unexpected error occurred',
          };
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(message)),
          );
        },
      );
    });

    return Scaffold(
      body: SafeArea(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Form(
            key: _formKey,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                Text(
                  'Welcome Back',
                  style: Theme.of(context).textTheme.headlineLarge,
                ),
                const SizedBox(height: 32),
                TextFormField(
                  controller: _emailController,
                  decoration: const InputDecoration(
                    labelText: 'Email',
                    border: OutlineInputBorder(),
                  ),
                  keyboardType: TextInputType.emailAddress,
                  validator: (value) {
                    if (value == null || value.isEmpty) {
                      return 'Please enter your email';
                    }
                    if (!value.contains('@')) {
                      return 'Please enter a valid email';
                    }
                    return null;
                  },
                ),
                const SizedBox(height: 16),
                TextFormField(
                  controller: _passwordController,
                  decoration: const InputDecoration(
                    labelText: 'Password',
                    border: OutlineInputBorder(),
                  ),
                  obscureText: true,
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
                const SizedBox(height: 24),
                FilledButton(
                  onPressed: authState.isLoading ? null : _handleLogin,
                  child: authState.isLoading
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Sign In'),
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

## Network Layer Examples

### Swift URLSession with Async/Await

```swift
actor NetworkClient {
    private let session: URLSession
    private let baseURL: URL
    private let decoder: JSONDecoder

    init(baseURL: URL, session: URLSession = .shared) {
        self.baseURL = baseURL
        self.session = session
        self.decoder = JSONDecoder()
        decoder.keyDecodingStrategy = .convertFromSnakeCase
    }

    func get<T: Decodable>(_ path: String, query: [String: String] = [:]) async throws -> T {
        var components = URLComponents(url: baseURL.appendingPathComponent(path), resolvingAgainstBaseURL: true)!
        components.queryItems = query.map { URLQueryItem(name: $0.key, value: $0.value) }

        let (data, response) = try await session.data(from: components.url!)
        try validateResponse(response)
        return try decoder.decode(T.self, from: data)
    }

    func post<T: Decodable, B: Encodable>(_ path: String, body: B) async throws -> T {
        var request = URLRequest(url: baseURL.appendingPathComponent(path))
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.httpBody = try JSONEncoder().encode(body)

        let (data, response) = try await session.data(for: request)
        try validateResponse(response)
        return try decoder.decode(T.self, from: data)
    }

    private func validateResponse(_ response: URLResponse) throws {
        guard let httpResponse = response as? HTTPURLResponse else {
            throw NetworkError.invalidResponse
        }
        guard 200..<300 ~= httpResponse.statusCode else {
            throw NetworkError.statusCode(httpResponse.statusCode)
        }
    }
}
```

### Kotlin Retrofit with Coroutines

```kotlin
interface ApiService {
    @GET("users/{id}")
    suspend fun getUser(@Path("id") id: String): User

    @POST("users")
    suspend fun createUser(@Body user: CreateUserRequest): User

    @GET("users")
    suspend fun getUsers(
        @Query("page") page: Int,
        @Query("limit") limit: Int
    ): PaginatedResponse<User>
}

class NetworkModule {
    @Provides
    @Singleton
    fun provideOkHttpClient(
        authInterceptor: AuthInterceptor,
        loggingInterceptor: HttpLoggingInterceptor
    ): OkHttpClient {
        return OkHttpClient.Builder()
            .addInterceptor(authInterceptor)
            .addInterceptor(loggingInterceptor)
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(30, TimeUnit.SECONDS)
            .build()
    }

    @Provides
    @Singleton
    fun provideRetrofit(okHttpClient: OkHttpClient): Retrofit {
        return Retrofit.Builder()
            .baseUrl(BuildConfig.API_BASE_URL)
            .client(okHttpClient)
            .addConverterFactory(MoshiConverterFactory.create())
            .build()
    }

    @Provides
    @Singleton
    fun provideApiService(retrofit: Retrofit): ApiService {
        return retrofit.create(ApiService::class.java)
    }
}

class AuthInterceptor @Inject constructor(
    private val tokenStorage: TokenStorage
) : Interceptor {
    override fun intercept(chain: Interceptor.Chain): Response {
        val originalRequest = chain.request()

        val token = runBlocking { tokenStorage.getAccessToken() }
            ?: return chain.proceed(originalRequest)

        val newRequest = originalRequest.newBuilder()
            .header("Authorization", "Bearer $token")
            .build()

        return chain.proceed(newRequest)
    }
}
```

### Flutter Dio Client

```dart
class ApiClient {
  late final Dio _dio;

  ApiClient({required String baseUrl, required SecureStorage storage}) {
    _dio = Dio(BaseOptions(
      baseUrl: baseUrl,
      connectTimeout: const Duration(seconds: 30),
      receiveTimeout: const Duration(seconds: 30),
    ));

    _dio.interceptors.addAll([
      AuthInterceptor(storage),
      LogInterceptor(requestBody: true, responseBody: true),
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
}

class AuthInterceptor extends Interceptor {
  final SecureStorage _storage;

  AuthInterceptor(this._storage);

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
    if (err.response?.statusCode == 401) {
      // Attempt token refresh
      try {
        final tokens = await _storage.getTokens();
        if (tokens != null) {
          // Refresh token logic
          final newTokens = await _refreshTokens(tokens.refreshToken);
          await _storage.saveTokens(newTokens);

          // Retry original request
          final options = err.requestOptions;
          options.headers['Authorization'] = 'Bearer ${newTokens.accessToken}';
          final response = await Dio().fetch(options);
          return handler.resolve(response);
        }
      } catch (e) {
        await _storage.clearTokens();
      }
    }
    handler.next(err);
  }

  Future<AuthTokens> _refreshTokens(String refreshToken) async {
    // Implementation
    throw UnimplementedError();
  }
}
```

## Testing Examples

### Swift XCTest with async

```swift
@MainActor
final class AuthViewModelTests: XCTestCase {
    var sut: AuthViewModel!
    var mockAPI: MockAuthAPI!
    var mockKeychain: MockKeychain!

    override func setUp() {
        mockAPI = MockAuthAPI()
        mockKeychain = MockKeychain()
        sut = AuthViewModel(api: mockAPI, keychain: mockKeychain)
    }

    func testLoginSuccess() async {
        // Given
        mockAPI.mockTokens = AuthTokens(
            accessToken: "test-token",
            refreshToken: "refresh-token",
            expiresAt: Date().addingTimeInterval(3600)
        )
        mockAPI.mockUser = User(id: "1", email: "test@example.com", name: "Test User", avatarURL: nil)

        // When
        await sut.login(email: "test@example.com", password: "password123")

        // Then
        XCTAssertNotNil(sut.user)
        XCTAssertEqual(sut.user?.email, "test@example.com")
        XCTAssertNil(sut.error)
        XCTAssertFalse(sut.isLoading)
    }

    func testLoginInvalidCredentials() async {
        // Given
        mockAPI.error = .invalidCredentials

        // When
        await sut.login(email: "test@example.com", password: "wrong")

        // Then
        XCTAssertNil(sut.user)
        XCTAssertEqual(sut.error, .invalidCredentials)
    }
}
```

### Kotlin Coroutines Test

```kotlin
@OptIn(ExperimentalCoroutinesApi::class)
class AuthViewModelTest {
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    private lateinit var viewModel: AuthViewModel
    private val mockRepository = mockk<AuthRepository>()

    @Before
    fun setup() {
        viewModel = AuthViewModel(mockRepository)
    }

    @Test
    fun `login success updates state to authenticated`() = runTest {
        // Given
        val testUser = User("1", "test@example.com", "Test User", null)
        coEvery { mockRepository.login(any(), any()) } returns Result.success(testUser)

        // When
        viewModel.login("test@example.com", "password123")
        advanceUntilIdle()

        // Then
        val state = viewModel.uiState.value
        assertThat(state).isInstanceOf(AuthUiState.Authenticated::class.java)
        assertThat((state as AuthUiState.Authenticated).user).isEqualTo(testUser)
    }

    @Test
    fun `login failure updates state to error`() = runTest {
        // Given
        coEvery { mockRepository.login(any(), any()) } returns
            Result.failure(AuthError.InvalidCredentials)

        // When
        viewModel.login("test@example.com", "wrong")
        advanceUntilIdle()

        // Then
        val state = viewModel.uiState.value
        assertThat(state).isInstanceOf(AuthUiState.Error::class.java)
    }
}
```

### Flutter Widget Test

```dart
void main() {
  group('LoginScreen', () {
    late ProviderContainer container;

    setUp(() {
      container = ProviderContainer(overrides: [
        authControllerProvider.overrideWith(() => MockAuthController()),
      ]);
    });

    tearDown(() {
      container.dispose();
    });

    testWidgets('shows validation errors for empty fields', (tester) async {
      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: const MaterialApp(home: LoginScreen()),
        ),
      );

      await tester.tap(find.byType(FilledButton));
      await tester.pump();

      expect(find.text('Please enter your email'), findsOneWidget);
      expect(find.text('Please enter your password'), findsOneWidget);
    });

    testWidgets('calls login when form is valid', (tester) async {
      final mockController = MockAuthController();
      container = ProviderContainer(overrides: [
        authControllerProvider.overrideWith(() => mockController),
      ]);

      await tester.pumpWidget(
        UncontrolledProviderScope(
          container: container,
          child: const MaterialApp(home: LoginScreen()),
        ),
      );

      await tester.enterText(
        find.byType(TextFormField).first,
        'test@example.com',
      );
      await tester.enterText(
        find.byType(TextFormField).last,
        'password123',
      );
      await tester.tap(find.byType(FilledButton));
      await tester.pump();

      verify(() => mockController.login('test@example.com', 'password123'))
          .called(1);
    });
  });
}
```

---

Version: 1.0.0
Last Updated: 2025-12-07
