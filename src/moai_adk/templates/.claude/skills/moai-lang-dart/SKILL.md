---
name: moai-lang-dart
version: 2.0.0
created: 2025-11-06
updated: 2025-11-06
status: active
description: "Dart best practices with Flutter mobile development, async programming, and server-side Dart for 2025"
keywords: [dart, programming, flutter, mobile, async, server-side, development]
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - WebFetch
  - WebSearch
---

# Dart Development Mastery

**Modern Dart Development with 2025 Best Practices**

> Comprehensive Dart development guidance covering Flutter mobile applications, async programming patterns, server-side development, and cross-platform solutions using the latest tools and frameworks.

## What It Does

### Flutter Mobile Development
- **Mobile App Development**: Flutter 3.x with Material Design 3 and Cupertino widgets
- **State Management**: Provider, Riverpod, BLoC patterns for scalable applications
- **Navigation**: Go Router for declarative routing and deep linking
- **Performance**: Widget optimization, lazy loading, memory management

### Server-Side Development
- **Web APIs**: Shelf, Dart Frog, or Aqueduct for backend services
- **Database Integration**: PostgreSQL, MongoDB with async drivers
- **Real-time Communication**: WebSockets, gRPC with Dart's async capabilities
- **Testing**: Unit tests, widget tests, integration tests with Dart test framework

### Cross-Platform Development
- **Flutter for Multiple Platforms**: iOS, Android, Web, Desktop, and Embedded
- **Shared Codebases**: Dart packages for business logic across platforms
- **Platform-Specific APIs**: Method channels for native integration

## When to Use

### Perfect Scenarios
- **Building cross-platform mobile applications with Flutter**
- **Developing scalable Flutter apps with modern state management**
- **Creating server-side APIs with Dart**
- **Implementing real-time applications with WebSockets**
- **Building web applications with Flutter Web**
- **Developing desktop applications with Flutter Desktop**
- **Creating embedded systems applications**

### Common Triggers
- "Create Flutter app"
- "Build Dart web API"
- "Implement Flutter state management"
- "Optimize Flutter performance"
- "Test Flutter application"
- "Dart best practices"

## Tool Version Matrix (2025-11-06)

### Core Dart/Flutter
- **Dart**: 3.5.x (current) / 3.4.x (LTS)
- **Flutter**: 3.24.x (current) / 3.22.x (LTS)
- **Dart SDK**: 3.5.0
- **Package Manager**: pub (built-in)

### Flutter Frameworks
- **Material Design**: 3.x (Material 3)
- **Cupertino Widgets**: iOS 17+ support
- **Go Router**: 13.x - Declarative routing
- **Riverpod**: 2.5.x - Reactive state management
- **BLoC**: 8.1.x - State management library

### Testing Tools
- **Dart Test**: Built-in testing framework
- **Flutter Test**: Widget and integration testing
- **Mockito**: 5.4.x - Mocking framework
- **Golden Tests**: Widget screenshot testing
- **Integration Test**: End-to-end testing

### Development Tools
- **Flutter CLI**: 3.24.x
- **Dart DevTools**: Web-based debugging tools
- **Android Studio**: Dolphin 2024.x
- **VS Code**: Flutter extension

### Backend Tools
- **Shelf**: 1.4.x - Web server framework
- **Dart Frog**: 1.0.x - Server-side framework
- **Aqueduct**: 7.x - Full-stack framework
- **MongoDB Dart Driver**: 4.12.x

## Ecosystem Overview

### Package Management

```bash
# Create new Flutter project
flutter create my_app
flutter create --org com.example --platforms=web,desktop my_app

# Create new Dart project
dart create my_dart_app

# Add dependencies
flutter pub add provider riverpod go_router
flutter pub add dio retrofit json_annotation
dart pub add shelf shelf_router

# Get dependencies
flutter pub get
dart pub get

# Run and build
flutter run
flutter run --release
flutter build apk
flutter build web
dart run
```

### Project Structure (2025 Best Practice)

```
my_flutter_app/
├── lib/
│   ├── main.dart                 # App entry point
│   ├── app.dart                  # App configuration
│   ├── core/                     # Core utilities
│   │   ├── constants/           # App constants
│   │   ├── errors/              # Custom error classes
│   │   ├── extensions/          # Dart extensions
│   │   ├── network/             # Network configuration
│   │   ├── themes/              # App themes
│   │   └── utils/               # Utility functions
│   ├── features/                # Feature modules
│   │   ├── authentication/      # Auth feature
│   │   │   ├── data/            # Data layer (repositories, data sources)
│   │   │   ├── domain/          # Domain layer (entities, use cases)
│   │   │   └── presentation/    # UI layer (pages, widgets, providers)
│   │   ├── user_profile/        # User profile feature
│   │   └── settings/            # Settings feature
│   ├── shared/                  # Shared components
│   │   ├── widgets/             # Reusable widgets
│   │   ├── models/              # Shared data models
│   │   ├── services/            # Shared services
│   │   └── providers/           # Shared providers
│   └── routes/                  # App routes
├── test/                        # Test files
│   ├── unit/                    # Unit tests
│   ├── widget/                  # Widget tests
│   └── integration/             # Integration tests
├── assets/                      # Static assets
├── pubspec.yaml                 # Dependencies
└── analysis_options.yaml        # Linting rules
```

## Modern Development Patterns

### Dart 3.x Language Features

```dart
// Enhanced patterns with records and pattern matching
sealed class NetworkResult<T> {
  const NetworkResult();
}

class Success<T> extends NetworkResult<T> {
  final T data;
  const Success(this.data);
}

class Error<T> extends NetworkResult<T> {
  final String message;
  final Exception? exception;
  const Error(this.message, [this.exception]);
}

class Loading<T> extends NetworkResult<T> {
  const Loading();
}

// Pattern matching with switch expressions
T handleNetworkResult<T>(NetworkResult<T> result) {
  return switch (result) {
    Success(data: final data) => data,
    Error(message: final message) => throw Exception(message),
    Loading() => throw StateError('Still loading'),
  };
}

// Records for lightweight data structures
typedef UserInfo = (String name, int age, String email);

class UserService {
  UserInfo getUserInfo(int id) {
    return ('John Doe', 30, 'john@example.com');
  }
  
  void printUserInfo(UserInfo user) {
    final (name, age, email) = user;
    print('$name, $age years old, email: $email');
  }
}

// Enhanced type inference and const constructors
class AppConfig {
  static const apiBaseUrl = String.fromEnvironment('API_BASE_URL', defaultValue: 'https://api.example.com');
  static const appVersion = String.fromEnvironment('APP_VERSION', defaultValue: '1.0.0');
  static const isDebug = bool.fromEnvironment('DEBUG', defaultValue: false);
}

// Enhanced enums with methods and properties
enum ThemeMode {
  light._('Light Theme', '☀️'),
  dark._('Dark Theme', '🌙'),
  system._('System Theme', '💻');
  
  const ThemeMode._(this.displayName, this.icon);
  
  final String displayName;
  final String icon;
  
  Brightness get brightness => switch (this) {
    ThemeMode.light => Brightness.light,
    ThemeMode.dark => Brightness.dark,
    ThemeMode.system => PlatformDispatcher.instance.platformBrightness,
  };
}

// Extension methods for enhanced APIs
extension StringExtension on String {
  bool get isValidEmail {
    return RegExp(r'^[^@]+@[^@]+\.[^@]+').hasMatch(this);
  }
  
  String get capitalize {
    return '${this[0].toUpperCase()}${substring(1)}';
  }
  
  String truncate(int length, {String suffix = '...'}) {
    if (this.length <= length) return this;
    return '${substring(0, length)}$suffix';
  }
}
```

### Modern Flutter State Management with Riverpod

```dart
// Provider setup with Riverpod 2.x
import 'package:flutter_riverpod/flutter_riverpod.dart';

// Model classes with immutability
@immutable
class User {
  final String id;
  final String name;
  final String email;
  final String avatarUrl;
  
  const User({
    required this.id,
    required this.name,
    required this.email,
    required this.avatarUrl,
  });
  
  User copyWith({
    String? id,
    String? name,
    String? email,
    String? avatarUrl,
  }) {
    return User(
      id: id ?? this.id,
      name: name ?? this.name,
      email: email ?? this.email,
      avatarUrl: avatarUrl ?? this.avatarUrl,
    );
  }
  
  @override
  bool operator ==(Object other) {
    if (identical(this, other)) return true;
    return other is User && 
           other.id == id && 
           other.name == name && 
           other.email == email;
  }
  
  @override
  int get hashCode => id.hashCode ^ name.hashCode ^ email.hashCode;
}

// Repository interface
abstract class UserRepository {
  Future<User> getUser(String userId);
  Future<List<User>> getUsers({int page = 1, int limit = 20});
  Future<User> updateUser(User user);
  Future<void> deleteUser(String userId);
}

// Implementation with HTTP client
class HttpUserRepository implements UserRepository {
  final Dio _dio;
  
  HttpUserRepository(this._dio);
  
  @override
  Future<User> getUser(String userId) async {
    try {
      final response = await _dio.get('/users/$userId');
      return User.fromJson(response.data);
    } on DioException catch (e) {
      throw UserRepositoryException('Failed to get user: $e');
    }
  }
  
  @override
  Future<List<User>> getUsers({int page = 1, int limit = 20}) async {
    try {
      final response = await _dio.get('/users', queryParameters: {
        'page': page,
        'limit': limit,
      });
      
      return (response.data as List)
          .map((json) => User.fromJson(json))
          .toList();
    } on DioException catch (e) {
      throw UserRepositoryException('Failed to get users: $e');
    }
  }
  
  @override
  Future<User> updateUser(User user) async {
    try {
      final response = await _dio.put('/users/${user.id}', data: user.toJson());
      return User.fromJson(response.data);
    } on DioException catch (e) {
      throw UserRepositoryException('Failed to update user: $e');
    }
  }
  
  @override
  Future<void> deleteUser(String userId) async {
    try {
      await _dio.delete('/users/$userId');
    } on DioException catch (e) {
      throw UserRepositoryException('Failed to delete user: $e');
    }
  }
}

// Riverpod providers
final dioProvider = Provider<Dio>((ref) {
  final dio = Dio(BaseOptions(baseUrl: AppConfig.apiBaseUrl));
  dio.interceptors.add(LogInterceptor());
  return dio;
});

final userRepositoryProvider = Provider<UserRepository>((ref) {
  return HttpUserRepository(ref.read(dioProvider));
});

// Async providers for data fetching
final userProvider = FutureProvider.family<User, String>((ref, userId) async {
  final repository = ref.watch(userRepositoryProvider);
  return repository.getUser(userId);
});

final usersProvider = AsyncNotifierProvider<UsersNotifier, List<User>>(UsersNotifier.new);

// Notifier for managing state
class UsersNotifier extends AsyncNotifier<List<User>> {
  int _page = 1;
  final int _limit = 20;
  bool _hasMore = true;
  
  @override
  Future<List<User>> build() async {
    return _loadUsers();
  }
  
  Future<List<User>> _loadUsers() async {
    if (state.isLoading || !_hasMore) return state.value ?? [];
    
    state = const AsyncLoading();
    
    try {
      final repository = ref.read(userRepositoryProvider);
      final newUsers = await repository.getUsers(page: _page, limit: _limit);
      
      if (newUsers.length < _limit) {
        _hasMore = false;
      }
      
      final currentUsers = state.value ?? [];
      final updatedUsers = _page == 1 ? newUsers : [...currentUsers, ...newUsers];
      
      state = AsyncData(updatedUsers);
      _page++;
      
      return updatedUsers;
    } catch (error, stackTrace) {
      state = AsyncError(error, stackTrace);
      rethrow;
    }
  }
  
  Future<void> loadMoreUsers() async {
    await _loadUsers();
  }
  
  Future<void> refresh() async {
    _page = 1;
    _hasMore = true;
    await _loadUsers();
  }
}

// State management with Notifier for single user
final userNotifierProvider = AsyncNotifierProvider.family<UserNotifier, User, String>(UserNotifier.new);

class UserNotifier extends FamilyAsyncNotifier<User, String> {
  @override
  Future<User> build(String arg) async {
    final repository = ref.read(userRepositoryProvider);
    return repository.getUser(arg);
  }
  
  Future<void> updateUser(User user) async {
    state = const AsyncLoading();
    
    try {
      final repository = ref.read(userRepositoryProvider);
      final updatedUser = await repository.updateUser(user);
      state = AsyncData(updatedUser);
    } catch (error, stackTrace) {
      state = AsyncError(error, stackTrace);
    }
  }
}

// App-wide state providers
final appThemeProvider = StateProvider<ThemeMode>((ref) => ThemeMode.system);
final appLocaleProvider = StateProvider<Locale>((ref) => const Locale('en', 'US'));
```

### Modern Flutter UI with Material 3

```dart
// Modern app with Material 3 and go_router
class MyApp extends ConsumerWidget {
  const MyApp({super.key});
  
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final appRouter = ref.watch(goRouterProvider);
    final themeMode = ref.watch(appThemeProvider);
    final appLocale = ref.watch(appLocaleProvider);
    
    return MaterialApp.router(
      title: 'My Flutter App',
      debugShowCheckedModeBanner: false,
      themeMode: themeMode,
      theme: AppTheme.lightTheme,
      darkTheme: AppTheme.darkTheme,
      locale: appLocale,
      supportedLocales: const [
        Locale('en', 'US'),
        Locale('es', 'ES'),
        Locale('fr', 'FR'),
      ],
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      routerConfig: appRouter,
    );
  }
}

// Modern theme configuration
class AppTheme {
  static ThemeData get lightTheme {
    return ThemeData(
      useMaterial3: true,
      colorScheme: ColorScheme.fromSeed(
        seedColor: const Color(0xFF6750A4),
        brightness: Brightness.light,
      ),
      appBarTheme: const AppBarTheme(
        centerTitle: true,
        elevation: 0,
        scrolledUnderElevation: 1,
      ),
      cardTheme: CardTheme(
        elevation: 2,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
        ),
      ),
      elevatedButtonTheme: ElevatedButtonThemeData(
        style: ElevatedButton.styleFrom(
          elevation: 1,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(20),
          ),
          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
        ),
      ),
      inputDecorationTheme: InputDecorationTheme(
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(12),
        ),
        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      ),
    );
  }
  
  static ThemeData get darkTheme {
    return ThemeData(
      useMaterial3: true,
      colorScheme: ColorScheme.fromSeed(
        seedColor: const Color(0xFF6750A4),
        brightness: Brightness.dark,
      ),
      appBarTheme: const AppBarTheme(
        centerTitle: true,
        elevation: 0,
        scrolledUnderElevation: 1,
      ),
      cardTheme: CardTheme(
        elevation: 2,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(16),
        ),
      ),
    );
  }
}

// Modern user list widget with state management
class UserListScreen extends ConsumerWidget {
  const UserListScreen({super.key});
  
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final usersAsync = ref.watch(usersProvider);
    
    return Scaffold(
      appBar: AppBar(
        title: const Text('Users'),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () {
              ref.read(usersProvider.notifier).refresh();
            },
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          await ref.read(usersProvider.notifier).refresh();
        },
        child: usersAsync.when(
          data: (users) {
            if (users.isEmpty) {
              return const EmptyStateWidget(
                message: 'No users found',
                icon: Icons.people_outline,
              );
            }
            
            return NotificationListener<ScrollNotification>(
              onNotification: (scrollInfo) {
                if (scrollInfo.metrics.pixels == scrollInfo.metrics.maxScrollExtent) {
                  ref.read(usersProvider.notifier).loadMoreUsers();
                }
                return false;
              },
              child: ListView.builder(
                padding: const EdgeInsets.all(16),
                itemCount: users.length + 1, // +1 for loading indicator
                itemBuilder: (context, index) {
                  if (index == users.length) {
                    return const LoadingIndicator();
                  }
                  
                  return UserCard(user: users[index]);
                },
              ),
            );
          },
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (error, stack) => ErrorWidget(
            error: error,
            onRetry: () {
              ref.read(usersProvider.notifier).refresh();
            },
          ),
        ),
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () {
          context.go('/add-user');
        },
        icon: const Icon(Icons.add),
        label: const Text('Add User'),
      ),
    );
  }
}

// Modern user card widget
class UserCard extends ConsumerWidget {
  final User user;
  
  const UserCard({
    super.key,
    required this.user,
  });
  
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Card(
      margin: const EdgeInsets.only(bottom: 12),
      child: InkWell(
        onTap: () {
          context.go('/users/${user.id}');
        },
        borderRadius: BorderRadius.circular(16),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              CircleAvatar(
                radius: 28,
                backgroundImage: NetworkImage(user.avatarUrl),
                backgroundColor: Theme.of(context).colorScheme.surfaceVariant,
                child: user.avatarUrl.isEmpty
                    ? Text(
                        user.name.isNotEmpty ? user.name[0].toUpperCase() : '?',
                        style: Theme.of(context).textTheme.titleLarge,
                      )
                    : null,
              ),
              const SizedBox(width: 16),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      user.name,
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 4),
                    Text(
                      user.email,
                      style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: Theme.of(context).colorScheme.onSurfaceVariant,
                      ),
                    ),
                  ],
                ),
              ),
              IconButton(
                icon: const Icon(Icons.more_vert),
                onPressed: () {
                  _showUserMenu(context, ref, user);
                },
              ),
            ],
          ),
        ),
      ),
    );
  }
  
  void _showUserMenu(BuildContext context, WidgetRef ref, User user) {
    showModalBottomSheet(
      context: context,
      builder: (context) {
        return SafeArea(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              ListTile(
                leading: const Icon(Icons.edit),
                title: const Text('Edit User'),
                onTap: () {
                  Navigator.pop(context);
                  context.go('/users/${user.id}/edit');
                },
              ),
              ListTile(
                leading: const Icon(Icons.delete, color: Colors.red),
                title: const Text('Delete User', style: TextStyle(color: Colors.red)),
                onTap: () async {
                  Navigator.pop(context);
                  
                  final confirmed = await _showDeleteConfirmation(context);
                  if (confirmed) {
                    try {
                      await ref.read(userRepositoryProvider).deleteUser(user.id);
                      if (context.mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          const SnackBar(content: Text('User deleted successfully')),
                        );
                      }
                    } catch (error) {
                      if (context.mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(content: Text('Failed to delete user: $error')),
                        );
                      }
                    }
                  }
                },
              ),
            ],
          ),
        );
      },
    );
  }
  
  Future<bool> _showDeleteConfirmation(BuildContext context) async {
    return await showDialog<bool>(
      context: context,
      builder: (context) {
        return AlertDialog(
          title: const Text('Delete User'),
          content: Text('Are you sure you want to delete ${user.name}?'),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context, false),
              child: const Text('Cancel'),
            ),
            TextButton(
              onPressed: () => Navigator.pop(context, true),
              style: TextButton.styleFrom(foregroundColor: Colors.red),
              child: const Text('Delete'),
            ),
          ],
        );
      },
    ) ?? false;
  }
}
```

### Go Router for Navigation

```dart
// Go router configuration
final goRouterProvider = Provider<GoRouter>((ref) {
  return GoRouter(
    initialLocation: '/',
    debugLogDiagnostics: AppConfig.isDebug,
    routes: [
      // Shell route for navigation
      ShellRoute(
        builder: (context, state, child) {
          return MainScaffold(child: child);
        },
        routes: [
          // Home route
          GoRoute(
            path: '/',
            builder: (context, state) => const HomeScreen(),
          ),
          
          // User routes
          GoRoute(
            path: '/users',
            builder: (context, state) => const UserListScreen(),
            routes: [
              GoRoute(
                path: '/:userId',
                builder: (context, state) {
                  final userId = state.pathParameters['userId']!;
                  return UserDetailScreen(userId: userId);
                },
                routes: [
                  GoRoute(
                    path: '/edit',
                    builder: (context, state) {
                      final userId = state.pathParameters['userId']!;
                      return EditUserScreen(userId: userId);
                    },
                  ),
                ],
              ),
            ],
          ),
          
          // Settings route
          GoRoute(
            path: '/settings',
            builder: (context, state) => const SettingsScreen(),
            routes: [
              GoRoute(
                path: '/profile',
                builder: (context, state) => const ProfileSettingsScreen(),
              ),
              GoRoute(
                path: '/appearance',
                builder: (context, state) => const AppearanceSettingsScreen(),
              ),
            ],
          ),
        ],
      ),
      
      // Standalone routes
      GoRoute(
        path: '/add-user',
        builder: (context, state) => const AddUserScreen(),
      ),
      
      GoRoute(
        path: '/login',
        builder: (context, state) => const LoginScreen(),
      ),
    ],
    
    // Error handling
    errorBuilder: (context, state) => ErrorScreen(error: state.error),
    
    // Redirects
    redirect: (context, state) {
      // Example: redirect to login if not authenticated
      final isAuthenticated = true; // Check authentication status
      
      if (!isAuthenticated && !state.location.startsWith('/login')) {
        return '/login';
      }
      
      return null;
    },
  );
});

// Main scaffold with bottom navigation
class MainScaffold extends ConsumerWidget {
  const MainScaffold({
    required this.child,
    super.key,
  });
  
  final Widget child;
  
  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final selectedIndex = ref.watch(bottomNavigationIndexProvider);
    
    return Scaffold(
      body: child,
      bottomNavigationBar: NavigationBar(
        selectedIndex: selectedIndex,
        onDestinationSelected: (index) {
          ref.read(bottomNavigationIndexProvider.notifier).state = index;
          
          switch (index) {
            case 0:
              context.go('/');
              break;
            case 1:
              context.go('/users');
              break;
            case 2:
              context.go('/settings');
              break;
          }
        },
        destinations: const [
          NavigationDestination(
            icon: Icon(Icons.home_outlined),
            selectedIcon: Icon(Icons.home),
            label: 'Home',
          ),
          NavigationDestination(
            icon: Icon(Icons.people_outlined),
            selectedIcon: Icon(Icons.people),
            label: 'Users',
          ),
          NavigationDestination(
            icon: Icon(Icons.settings_outlined),
            selectedIcon: Icon(Icons.settings),
            label: 'Settings',
          ),
        ],
      ),
    );
  }
}

// Provider for bottom navigation state
final bottomNavigationIndexProvider = StateProvider<int>((ref) => 0);
```

## Performance Considerations

### Widget Performance

```dart
// Performance-optimized widgets with const constructors
class OptimizedUserCard extends StatelessWidget {
  final User user;
  final VoidCallback? onTap;
  final VoidCallback? onEdit;
  final VoidCallback? onDelete;
  
  const OptimizedUserCard({
    super.key,
    required this.user,
    this.onTap,
    this.onEdit,
    this.onDelete,
  });
  
  @override
  Widget build(BuildContext context) {
    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(12),
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Row(
            children: [
              // Use Hero widget for smooth avatar transitions
              Hero(
                tag: 'user-avatar-${user.id}',
                child: UserAvatar(
                  imageUrl: user.avatarUrl,
                  name: user.name,
                  size: 48,
                ),
              ),
              const SizedBox(width: 16),
              Expanded(
                child: _buildUserInfo(),
              ),
              _buildActionButtons(),
            ],
          ),
        ),
      ),
    );
  }
  
  Widget _buildUserInfo() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(
          user.name,
          style: const TextStyle(
            fontSize: 16,
            fontWeight: FontWeight.w600,
          ),
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
        const SizedBox(height: 4),
        Text(
          user.email,
          style: const TextStyle(
            fontSize: 14,
            color: Colors.grey,
          ),
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
      ],
    );
  }
  
  Widget _buildActionButtons() {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        if (onEdit != null)
          IconButton(
            icon: const Icon(Icons.edit_outlined),
            onPressed: onEdit,
            visualDensity: VisualDensity.compact,
          ),
        if (onDelete != null)
          IconButton(
            icon: const Icon(Icons.delete_outline),
            onPressed: onDelete,
            visualDensity: VisualDensity.compact,
          ),
      ],
    );
  }
}

// Efficient image loading and caching
class CachedNetworkImage extends StatefulWidget {
  final String imageUrl;
  final double? width;
  final double? height;
  final Widget? placeholder;
  final Widget? errorWidget;
  final BoxFit fit;
  
  const CachedNetworkImage({
    super.key,
    required this.imageUrl,
    this.width,
    this.height,
    this.placeholder,
    this.errorWidget,
    this.fit = BoxFit.cover,
  });
  
  @override
  State<CachedNetworkImage> createState() => _CachedNetworkImageState();
}

class _CachedNetworkImageState extends State<CachedNetworkImage> {
  final Map<String, ui.Image> _imageCache = {};
  bool _isLoading = true;
  bool _hasError = false;
  
  @override
  void initState() {
    super.initState();
    _loadImage();
  }
  
  Future<void> _loadImage() async {
    if (_imageCache.containsKey(widget.imageUrl)) {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
      return;
    }
    
    try {
      final image = await _fetchImage();
      _imageCache[widget.imageUrl] = image;
      
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _isLoading = false;
          _hasError = true;
        });
      }
    }
  }
  
  Future<ui.Image> _fetchImage() async {
    final completer = Completer<ui.Image>();
    final codec = await ui.instantiateImageCodec(
      await NetworkAssetBundle(Uri.parse(widget.imageUrl)).load(widget.imageUrl).then((bytes) => bytes.buffer.asUint8List()),
    );
    final frame = await codec.getNextFrame();
    completer.complete(frame.image);
    return completer.future;
  }
  
  @override
  Widget build(BuildContext context) {
    if (_isLoading) {
      return widget.placeholder ??
          Container(
            width: widget.width,
            height: widget.height,
            color: Colors.grey[200],
            child: const Center(
              child: CircularProgressIndicator(),
            ),
          );
    }
    
    if (_hasError) {
      return widget.errorWidget ??
          Container(
            width: widget.width,
            height: widget.height,
            color: Colors.grey[200],
            child: const Icon(Icons.error),
          );
    }
    
    final image = _imageCache[widget.imageUrl];
    return CustomPaint(
      size: Size(widget.width ?? double.infinity, widget.height ?? double.infinity),
      painter: _ImagePainter(image: image!, fit: widget.fit),
    );
  }
}

class _ImagePainter extends CustomPainter {
  final ui.Image image;
  final BoxFit fit;
  
  _ImagePainter({required this.image, required this.fit});
  
  @override
  void paint(Canvas canvas, Size size) {
    final imageSize = Size(image.width.toDouble(), image.height.toDouble());
    final scales = _calculateScales(imageSize, size);
    
    final paint = Paint()
      ..isAntiAlias = true
      ..filterQuality = FilterQuality.high;
    
    canvas.save();
    
    if (fit == BoxFit.cover) {
      canvas.scale(scales.dx, scales.dy);
      canvas.drawImageRect(
        image,
        Rect.fromLTWH(0, 0, imageSize.width, imageSize.height),
        Rect.fromLTWH(0, 0, size.width / scales.dx, size.height / scales.dy),
        paint,
      );
    }
    
    canvas.restore();
  }
  
  Offset _calculateScales(Size inputSize, Size outputSize) {
    final scaleX = outputSize.width / inputSize.width;
    final scaleY = outputSize.height / inputSize.height;
    return Offset(scaleX, scaleY);
  }
  
  @override
  bool shouldRepaint(covariant _ImagePainter oldDelegate) {
    return image != oldDelegate.image || fit != oldDelegate.fit;
  }
}

// ListView with lazy loading and recycling
class OptimizedListView<T> extends StatelessWidget {
  final List<T> items;
  final Widget Function(BuildContext context, T item, int index) itemBuilder;
  final VoidCallback? onLoadMore;
  final bool hasMore;
  final bool isLoading;
  
  const OptimizedListView({
    super.key,
    required this.items,
    required this.itemBuilder,
    this.onLoadMore,
    this.hasMore = false,
    this.isLoading = false,
  });
  
  @override
  Widget build(BuildContext context) {
    return NotificationListener<ScrollNotification>(
      onNotification: (notification) {
        if (notification is ScrollEndNotification &&
            notification.metrics.extentAfter == 0 &&
            hasMore &&
            !isLoading &&
            onLoadMore != null) {
          onLoadMore!();
        }
        return false;
      },
      child: ListView.builder(
        itemCount: items.length + (hasMore ? 1 : 0),
        itemBuilder: (context, index) {
          if (index == items.length) {
            return const Center(
              child: Padding(
                padding: EdgeInsets.all(16.0),
                child: CircularProgressIndicator(),
              ),
            );
          }
          
          return itemBuilder(context, items[index], index);
        },
      ),
    );
  }
}
```

### Memory Management

```dart
// Efficient memory usage with ImageCache
class ImageCacheManager {
  static final ImageCacheManager _instance = ImageCacheManager._internal();
  factory ImageCacheManager() => _instance;
  ImageCacheManager._internal();
  
  final PaintingBinding _paintingBinding = PaintingBinding.instance;
  final Map<String, ui.Image> _memoryCache = {};
  final int _maxCacheSize = 100 * 1024 * 1024; // 100MB
  int _currentCacheSize = 0;
  
  Future<ui.Image?> getImage(String url) async {
    // Check memory cache first
    if (_memoryCache.containsKey(url)) {
      return _memoryCache[url];
    }
    
    // Check painting binding cache
    final cachedImage = _paintingBinding.imageCache?.image;
    if (cachedImage != null) {
      return cachedImage;
    }
    
    try {
      final image = await _loadImage(url);
      _addToMemoryCache(url, image);
      return image;
    } catch (e) {
      return null;
    }
  }
  
  Future<ui.Image> _loadImage(String url) async {
    final completer = Completer<ui.Image>();
    final codec = await ui.instantiateImageCodec(
      await _fetchImageData(url),
    );
    final frame = await codec.getNextFrame();
    completer.complete(frame.image);
    return completer.future;
  }
  
  Future<Uint8List> _fetchImageData(String url) async {
    final response = await http.get(Uri.parse(url));
    return response.bodyBytes;
  }
  
  void _addToMemoryCache(String url, ui.Image image) {
    final imageSize = image.width * image.height * 4; // 4 bytes per pixel
    
    if (_currentCacheSize + imageSize > _maxCacheSize) {
      _evictLeastRecentlyUsed(imageSize);
    }
    
    _memoryCache[url] = image;
    _currentCacheSize += imageSize;
  }
  
  void _evictLeastRecentlyUsed(int requiredSize) {
    final entries = _memoryCache.entries.toList();
    entries.sort((a, b) => a.key.compareTo(b.key));
    
    int freedSize = 0;
    for (final entry in entries) {
      final imageSize = entry.value.width * entry.value.height * 4;
      _memoryCache.remove(entry.key);
      freedSize += imageSize;
      _currentCacheSize -= imageSize;
      
      if (freedSize >= requiredSize) {
        break;
      }
    }
  }
  
  void clearCache() {
    _memoryCache.clear();
    _currentCacheSize = 0;
    _paintingBinding.imageCache?.clear();
    _paintingBinding.imageCache?.clearLiveImages();
  }
}

// Resource management with automatic cleanup
class ResourceManager {
  final Map<String, StreamSubscription> _subscriptions = {};
  final Map<String, Timer> _timers = {};
  
  StreamSubscription<T>? addSubscription<T>(
    String key,
    StreamSubscription<T> subscription,
  ) {
    _subscriptions[key] = subscription as StreamSubscription;
    return subscription;
  }
  
  Timer? addTimer(String key, Duration duration, VoidCallback callback) {
    final timer = Timer(duration, callback);
    _timers[key] = timer;
    return timer;
  }
  
  void removeSubscription(String key) {
    final subscription = _subscriptions.remove(key);
    subscription?.cancel();
  }
  
  void removeTimer(String key) {
    final timer = _timers.remove(key);
    timer?.cancel();
  }
  
  void dispose() {
    for (final subscription in _subscriptions.values) {
      subscription.cancel();
    }
    _subscriptions.clear();
    
    for (final timer in _timers.values) {
      timer.cancel();
    }
    _timers.clear();
  }
}

// Widget with automatic resource cleanup
class AutoCleanupWidget extends StatefulWidget {
  final Widget child;
  final VoidCallback? onInit;
  final VoidCallback? onDispose;
  
  const AutoCleanupWidget({
    super.key,
    required this.child,
    this.onInit,
    this.onDispose,
  });
  
  @override
  State<AutoCleanupWidget> createState() => _AutoCleanupWidgetState();
}

class _AutoCleanupWidgetState extends State<AutoCleanupWidget> {
  final ResourceManager _resourceManager = ResourceManager();
  
  @override
  void initState() {
    super.initState();
    widget.onInit?.call();
  }
  
  @override
  void dispose() {
    _resourceManager.dispose();
    widget.onDispose?.call();
    super.dispose();
  }
  
  @override
  Widget build(BuildContext context) {
    return widget.child;
  }
}
```

## Testing Strategy

### Unit Testing with Dart Test Framework

```dart
// Unit tests for business logic
void main() {
  group('UserRepository', () {
    late UserRepository userRepository;
    late MockDio mockDio;
    
    setUp(() {
      mockDio = MockDio();
      userRepository = HttpUserRepository(mockDio);
    });
    
    test('should return user when getUser is called with valid ID', () async {
      // Arrange
      final userId = '123';
      final userJson = {
        'id': userId,
        'name': 'John Doe',
        'email': 'john@example.com',
        'avatarUrl': 'https://example.com/avatar.jpg',
      };
      
      when(() => mockDio.get('/users/$userId'))
          .thenAnswer((_) async => Response(data: userJson, statusCode: 200));
      
      // Act
      final result = await userRepository.getUser(userId);
      
      // Assert
      expect(result.id, equals(userId));
      expect(result.name, equals('John Doe'));
      expect(result.email, equals('john@example.com'));
      expect(result.avatarUrl, equals('https://example.com/avatar.jpg'));
      
      verify(() => mockDio.get('/users/$userId')).called(1);
    });
    
    test('should throw UserRepositoryException when API call fails', () async {
      // Arrange
      final userId = '123';
      
      when(() => mockDio.get('/users/$userId'))
          .thenThrow(DioException(requestOptions: RequestOptions(path: '/users/$userId')));
      
      // Act & Assert
      expect(
        () => userRepository.getUser(userId),
        throwsA(isA<UserRepositoryException>()),
      );
      
      verify(() => mockDio.get('/users/$userId')).called(1);
    });
    
    test('should return users list when getUsers is called', () async {
      // Arrange
      final usersJson = [
        {
          'id': '1',
          'name': 'John Doe',
          'email': 'john@example.com',
          'avatarUrl': 'https://example.com/avatar1.jpg',
        },
        {
          'id': '2',
          'name': 'Jane Smith',
          'email': 'jane@example.com',
          'avatarUrl': 'https://example.com/avatar2.jpg',
        },
      ];
      
      when(() => mockDio.get('/users', queryParameters: any(named: 'queryParameters')))
          .thenAnswer((_) async => Response(data: usersJson, statusCode: 200));
      
      // Act
      final result = await userRepository.getUsers();
      
      // Assert
      expect(result, hasLength(2));
      expect(result[0].name, equals('John Doe'));
      expect(result[1].name, equals('Jane Smith'));
      
      verify(() => mockDio.get('/users', queryParameters: any(named: 'queryParameters'))).called(1);
    });
  });
  
  group('User model', () {
    test('should create user with valid data', () {
      // Arrange
      const user = User(
        id: '123',
        name: 'John Doe',
        email: 'john@example.com',
        avatarUrl: 'https://example.com/avatar.jpg',
      );
      
      // Act & Assert
      expect(user.id, equals('123'));
      expect(user.name, equals('John Doe'));
      expect(user.email, equals('john@example.com'));
      expect(user.avatarUrl, equals('https://example.com/avatar.jpg'));
    });
    
    test('should copy user with updated name', () {
      // Arrange
      const originalUser = User(
        id: '123',
        name: 'John Doe',
        email: 'john@example.com',
        avatarUrl: 'https://example.com/avatar.jpg',
      );
      
      // Act
      final updatedUser = originalUser.copyWith(name: 'Jane Doe');
      
      // Assert
      expect(updatedUser.id, equals(originalUser.id));
      expect(updatedUser.name, equals('Jane Doe'));
      expect(updatedUser.email, equals(originalUser.email));
      expect(updatedUser.avatarUrl, equals(originalUser.avatarUrl));
    });
    
    test('should compare users correctly', () {
      // Arrange
      const user1 = User(
        id: '123',
        name: 'John Doe',
        email: 'john@example.com',
        avatarUrl: 'https://example.com/avatar.jpg',
      );
      
      const user2 = User(
        id: '123',
        name: 'John Doe',
        email: 'john@example.com',
        avatarUrl: 'https://example.com/avatar.jpg',
      );
      
      const user3 = User(
        id: '456',
        name: 'John Doe',
        email: 'john@example.com',
        avatarUrl: 'https://example.com/avatar.jpg',
      );
      
      // Act & Assert
      expect(user1, equals(user2));
      expect(user1, isNot(equals(user3)));
      expect(user1.hashCode, equals(user2.hashCode));
      expect(user1.hashCode, isNot(equals(user3.hashCode)));
    });
  });
}

// Test for extensions
void main() {
  group('StringExtension', () {
    test('should validate email correctly', () {
      expect('test@example.com'.isValidEmail, isTrue);
      expect('test.name@example.com'.isValidEmail, isTrue);
      expect('test+tag@example.com'.isValidEmail, isTrue);
      
      expect('invalid-email'.isValidEmail, isFalse);
      expect('@example.com'.isValidEmail, isFalse);
      expect('test@'.isValidEmail, isFalse);
      expect('test@example'.isValidEmail, isFalse);
    });
    
    test('should capitalize first letter correctly', () {
      expect('hello'.capitalize, equals('Hello'));
      expect('WORLD'.capitalize, equals('WORLD'));
      expect(''.capitalize, equals(''));
    });
    
    test('should truncate string correctly', () {
      expect('short'.truncate(10), equals('short'));
      expect('this is a long string'.truncate(10), equals('this is a...'));
      expect('this is a long string'.truncate(10, suffix: ' [more]'), equals('this is a [more]'));
    });
  });
}
```

### Widget Testing

```dart
// Widget tests for UI components
void main() {
  group('UserCard Widget Tests', () {
    testWidgets('should display user information correctly', (WidgetTester tester) async {
      // Arrange
      const user = User(
        id: '123',
        name: 'John Doe',
        email: 'john@example.com',
        avatarUrl: 'https://example.com/avatar.jpg',
      );
      
      // Act
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: UserCard(user: user),
          ),
        ),
      );
      
      // Assert
      expect(find.text('John Doe'), findsOneWidget);
      expect(find.text('john@example.com'), findsOneWidget);
      expect(find.byType(CircleAvatar), findsOneWidget);
    });
    
    testWidgets('should call onTap when card is tapped', (WidgetTester tester) async {
      // Arrange
      const user = User(
        id: '123',
        name: 'John Doe',
        email: 'john@example.com',
        avatarUrl: 'https://example.com/avatar.jpg',
      );
      
      bool wasTapped = false;
      
      // Act
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: UserCard(
              user: user,
              onTap: () => wasTapped = true,
            ),
          ),
        ),
      );
      
      await tester.tap(find.byType(UserCard));
      await tester.pump();
      
      // Assert
      expect(wasTapped, isTrue);
    });
    
    testWidgets('should show edit and delete buttons when callbacks provided', (WidgetTester tester) async {
      // Arrange
      const user = User(
        id: '123',
        name: 'John Doe',
        email: 'john@example.com',
        avatarUrl: 'https://example.com/avatar.jpg',
      );
      
      // Act
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: UserCard(
              user: user,
              onEdit: () {},
              onDelete: () {},
            ),
          ),
        ),
      );
      
      // Assert
      expect(find.byIcon(Icons.edit_outlined), findsOneWidget);
      expect(find.byIcon(Icons.delete_outline), findsOneWidget);
    });
  });
  
  group('UserListScreen Widget Tests', () {
    testWidgets('should show loading indicator initially', (WidgetTester tester) async {
      // Arrange
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            usersProvider.overrideWith((ref) => AsyncValue.loading()),
          ],
          child: MaterialApp(
            home: UserListScreen(),
          ),
        ),
      );
      
      // Act & Assert
      expect(find.byType(CircularProgressIndicator), findsOneWidget);
    });
    
    testWidgets('should show error message when loading fails', (WidgetTester tester) async {
      // Arrange
      const errorMessage = 'Failed to load users';
      
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            usersProvider.overrideWith(
              (ref) => AsyncValue.error(Exception(errorMessage), StackTrace.current),
            ),
          ],
          child: MaterialApp(
            home: UserListScreen(),
          ),
        ),
      );
      
      await tester.pump();
      
      // Act & Assert
      expect(find.text(errorMessage), findsOneWidget);
      expect(find.byType(ElevatedButton), findsOneWidget);
    });
    
    testWidgets('should show users when loading succeeds', (WidgetTester tester) async {
      // Arrange
      final users = [
        const User(
          id: '1',
          name: 'John Doe',
          email: 'john@example.com',
          avatarUrl: 'https://example.com/avatar1.jpg',
        ),
        const User(
          id: '2',
          name: 'Jane Smith',
          email: 'jane@example.com',
          avatarUrl: 'https://example.com/avatar2.jpg',
        ),
      ];
      
      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            usersProvider.overrideWith((ref) => AsyncValue.data(users)),
          ],
          child: MaterialApp(
            home: UserListScreen(),
          ),
        ),
      );
      
      await tester.pump();
      
      // Act & Assert
      expect(find.text('John Doe'), findsOneWidget);
      expect(find.text('Jane Smith'), findsOneWidget);
      expect(find.text('john@example.com'), findsOneWidget);
      expect(find.text('jane@example.com'), findsOneWidget);
      expect(find.byType(UserCard), findsNWidgets(2));
    });
  });
}
```

### Integration Testing

```dart
// Integration tests with Flutter integration_test package
void main() {
  group('User Management Integration Tests', () {
    IntegrationTestWidgetsFlutterBinding.ensureInitialized();
    
    testWidgets('should complete user creation flow', (WidgetTester tester) async {
      // Arrange
      app.main();
      await tester.pumpAndSettle();
      
      // Navigate to add user screen
      await tester.tap(find.byIcon(Icons.add));
      await tester.pumpAndSettle();
      
      // Fill out user form
      await tester.enterText(find.byKey(const Key('name_field')), 'John Doe');
      await tester.enterText(find.byKey(const Key('email_field')), 'john@example.com');
      
      // Submit form
      await tester.tap(find.byKey(const Key('submit_button')));
      await tester.pumpAndSettle();
      
      // Assert user appears in list
      expect(find.text('John Doe'), findsOneWidget);
      expect(find.text('john@example.com'), findsOneWidget);
    });
    
    testWidgets('should navigate to user details and back', (WidgetTester tester) async {
      // Arrange
      app.main();
      await tester.pumpAndSettle();
      
      // Wait for users to load
      await tester.pumpAndSettle(const Duration(seconds: 3));
      
      // Tap on first user
      await tester.tap(find.byType(UserCard).first);
      await tester.pumpAndSettle();
      
      // Assert we're on user details screen
      expect(find.byType(UserDetailScreen), findsOneWidget);
      
      // Navigate back
      await tester.tap(find.byIcon(Icons.arrow_back));
      await tester.pumpAndSettle();
      
      // Assert we're back on user list
      expect(find.byType(UserListScreen), findsOneWidget);
    });
    
    testWidgets('should handle network errors gracefully', (WidgetTester tester) async {
      // Arrange - mock network failure
      setUpAll(() {
        HttpOverrides.global = MockHttpOverrides();
      });
      
      app.main();
      await tester.pumpAndSettle();
      
      // Wait for error to appear
      await tester.pumpAndSettle(const Duration(seconds: 3));
      
      // Assert error message is shown
      expect(find.byType(ErrorWidget), findsOneWidget);
      
      // Tap retry button
      await tester.tap(find.byType(ElevatedButton));
      await tester.pumpAndSettle();
      
      // Verify retry attempts
      expect(find.text('Retrying...'), findsOneWidget);
    });
  });
}

// Mock HTTP overrides for testing
class MockHttpOverrides extends HttpOverrides {
  @override
  HttpClient createHttpClient(SecurityContext? context) {
    return MockHttpClient();
  }
}

class MockHttpClient extends FakeHttpClient {
  @override
  Future<HttpClientRequest> getUrl(Uri url) async {
    if (url.path.contains('/users')) {
      throw HttpClientException('Connection failed');
    }
    return super.getUrl(url);
  }
}

// Golden tests for visual regression
void main() {
  group('UserCard Golden Tests', () {
    testWidgets('should match golden snapshot', (WidgetTester tester) async {
      // Arrange
      const user = User(
        id: '123',
        name: 'John Doe',
        email: 'john@example.com',
        avatarUrl: 'https://example.com/avatar.jpg',
      );
      
      // Act
      await tester.pumpWidget(
        MaterialApp(
          theme: ThemeData.light(),
          home: Scaffold(
            body: UserCard(user: user),
          ),
        ),
      );
      
      await tester.pumpAndSettle();
      
      // Assert
      await expectLater(
        find.byType(UserCard),
        matchesGoldenFile('goldens/user_card.png'),
      );
    });
    
    testWidgets('should match dark mode golden snapshot', (WidgetTester tester) async {
      // Arrange
      const user = User(
        id: '123',
        name: 'John Doe',
        email: 'john@example.com',
        avatarUrl: 'https://example.com/avatar.jpg',
      );
      
      // Act
      await tester.pumpWidget(
        MaterialApp(
          theme: ThemeData.dark(),
          home: Scaffold(
            body: UserCard(user: user),
          ),
        ),
      );
      
      await tester.pumpAndSettle();
      
      // Assert
      await expectLater(
        find.byType(UserCard),
        matchesGoldenFile('goldens/user_card_dark.png'),
      );
    });
  });
}
```

## Security Best Practices

### Input Validation

```dart
// Input validation utilities
class InputValidator {
  static String? validateEmail(String? value) {
    if (value == null || value.isEmpty) {
      return 'Email is required';
    }
    
    if (!value.isValidEmail) {
      return 'Please enter a valid email address';
    }
    
    return null;
  }
  
  static String? validatePassword(String? value) {
    if (value == null || value.isEmpty) {
      return 'Password is required';
    }
    
    if (value.length < 8) {
      return 'Password must be at least 8 characters long';
    }
    
    if (!value.contains(RegExp(r'[A-Z]'))) {
      return 'Password must contain at least one uppercase letter';
    }
    
    if (!value.contains(RegExp(r'[a-z]'))) {
      return 'Password must contain at least one lowercase letter';
    }
    
    if (!value.contains(RegExp(r'[0-9]'))) {
      return 'Password must contain at least one digit';
    }
    
    return null;
  }
  
  static String? validateName(String? value) {
    if (value == null || value.isEmpty) {
      return 'Name is required';
    }
    
    if (value.length < 2) {
      return 'Name must be at least 2 characters long';
    }
    
    if (value.length > 50) {
      return 'Name must be less than 50 characters';
    }
    
    if (!value.contains(RegExp(r'^[a-zA-Z\s]+$'))) {
      return 'Name can only contain letters and spaces';
    }
    
    return null;
  }
}

// Secure storage for sensitive data
class SecureStorage {
  static const _storage = FlutterSecureStorage();
  
  static Future<void> storeToken(String token) async {
    await _storage.write(key: 'auth_token', value: token);
  }
  
  static Future<String?> getToken() async {
    return await _storage.read(key: 'auth_token');
  }
  
  static Future<void> deleteToken() async {
    await _storage.delete(key: 'auth_token');
  }
  
  static Future<void> storeUserCredentials(String username, String password) async {
    await _storage.write(key: 'username', value: username);
    await _storage.write(key: 'password', value: password);
  }
  
  static Future<Map<String, String?>> getUserCredentials() async {
    final username = await _storage.read(key: 'username');
    final password = await _storage.read(key: 'password');
    return {'username': username, 'password': password};
  }
  
  static Future<void> clearAll() async {
    await _storage.deleteAll();
  }
}

// Secure HTTP client with authentication
class SecureHttpClient {
  late final Dio _dio;
  
  SecureHttpClient() {
    _dio = Dio(BaseOptions(
      baseUrl: AppConfig.apiBaseUrl,
      connectTimeout: const Duration(seconds: 30),
      receiveTimeout: const Duration(seconds: 30),
    ));
    
    _setupInterceptors();
  }
  
  void _setupInterceptors() {
    // Authentication interceptor
    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) async {
          final token = await SecureStorage.getToken();
          if (token != null) {
            options.headers['Authorization'] = 'Bearer $token';
          }
          handler.next(options);
        },
        onError: (error, handler) async {
          if (error.response?.statusCode == 401) {
            // Token expired, clear storage and navigate to login
            await SecureStorage.deleteToken();
            // Navigate to login screen
            _navigateToLogin();
          }
          handler.next(error);
        },
      ),
    );
    
    // Logging interceptor (only in debug mode)
    if (AppConfig.isDebug) {
      _dio.interceptors.add(LogInterceptor(
        requestBody: true,
        responseBody: true,
      ));
    }
    
    // Retry interceptor
    _dio.interceptors.add(RetryInterceptor(
      dio: _dio,
      options: const RetryOptions(
        retries: 3,
        retryInterval: Duration(seconds: 1),
      ),
    ));
  }
  
  void _navigateToLogin() {
    // Use navigator key to navigate to login screen
    navigatorKey.currentState?.pushNamedAndRemoveUntil(
      '/login',
      (route) => false,
    );
  }
  
  Future<Response<T>> get<T>(
    String path, {
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    return _dio.get<T>(path, queryParameters: queryParameters, options: options);
  }
  
  Future<Response<T>> post<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    return _dio.post<T>(
      path,
      data: data,
      queryParameters: queryParameters,
      options: options,
    );
  }
  
  Future<Response<T>> put<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    return _dio.put<T>(
      path,
      data: data,
      queryParameters: queryParameters,
      options: options,
    );
  }
  
  Future<Response<T>> delete<T>(
    String path, {
    dynamic data,
    Map<String, dynamic>? queryParameters,
    Options? options,
  }) async {
    return _dio.delete<T>(
      path,
      data: data,
      queryParameters: queryParameters,
      options: options,
    );
  }
}

// Local authentication with biometrics
class BiometricAuth {
  static final LocalAuthentication _auth = LocalAuthentication();
  
  static Future<bool> isAvailable() async {
    final isAvailable = await _auth.canCheckBiometrics;
    final isDeviceSupported = await _auth.isDeviceSupported();
    return isAvailable && isDeviceSupported;
  }
  
  static Future<bool> authenticate() async {
    try {
      final didAuthenticate = await _auth.authenticate(
        localizedReason: 'Please authenticate to access this feature',
        options: const AuthenticationOptions(
          biometricOnly: false,
          useErrorDialogs: true,
          stickyAuth: true,
        ),
      );
      return didAuthenticate;
    } catch (e) {
      return false;
    }
  }
  
  static Future<List<BiometricType>> getAvailableBiometrics() async {
    try {
      return await _auth.getAvailableBiometrics();
    } catch (e) {
      return [];
    }
  }
}
```

### Data Protection

```dart
// Encrypted data model
class SecureUser {
  final String id;
  final String encryptedName;
  final String encryptedEmail;
  
  SecureUser({
    required this.id,
    required this.encryptedName,
    required this.encryptedEmail,
  });
  
  factory SecureUser.fromUser(User user, EncryptionService encryptionService) {
    return SecureUser(
      id: user.id,
      encryptedName: encryptionService.encrypt(user.name),
      encryptedEmail: encryptionService.encrypt(user.email),
    );
  }
  
  User toUser(EncryptionService encryptionService) {
    return User(
      id: id,
      name: encryptionService.decrypt(encryptedName),
      email: encryptionService.decrypt(encryptedEmail),
      avatarUrl: '', // Not encrypted in this example
    );
  }
}

// Encryption service
class EncryptionService {
  static final EncryptionService _instance = EncryptionService._internal();
  factory EncryptionService() => _instance;
  EncryptionService._internal();
  
  late final Encrypter _encrypter;
  late final IV _iv;
  
  Future<void> initialize() async {
    final key = await _getOrCreateKey();
    _encrypter = Encrypter(AES(key));
    _iv = IV.fromLength(16);
  }
  
  Future<Key> _getOrCreateKey() async {
    const storage = FlutterSecureStorage();
    
    String? keyString = await storage.read(key: 'encryption_key');
    if (keyString != null) {
      return Key.fromBase64(keyString);
    }
    
    final key = Key.fromSecureRandom(32);
    await storage.write(key: 'encryption_key', value: key.base64);
    return key;
  }
  
  String encrypt(String plaintext) {
    final encrypted = _encrypter.encrypt(plaintext, iv: _iv);
    return encrypted.base64;
  }
  
  String decrypt(String ciphertext) {
    final encrypted = Encrypted.fromBase64(ciphertext);
    return _encrypter.decrypt(encrypted, iv: _iv);
  }
  
  Future<void> clearKey() async {
    const storage = FlutterSecureStorage();
    await storage.delete(key: 'encryption_key');
  }
}

// API security with rate limiting
class RateLimiter {
  final Map<String, List<DateTime>> _requests = {};
  final int maxRequests;
  final Duration timeWindow;
  
  RateLimiter({
    this.maxRequests = 100,
    this.timeWindow = const Duration(minutes: 1),
  });
  
  bool canMakeRequest(String identifier) {
    final now = DateTime.now();
    final requests = _requests[identifier] ?? [];
    
    // Remove old requests outside the time window
    requests.removeWhere((request) => now.difference(request) > timeWindow);
    
    if (requests.length >= maxRequests) {
      return false;
    }
    
    requests.add(now);
    _requests[identifier] = requests;
    return true;
  }
  
  Duration? getTimeUntilNextRequest(String identifier) {
    final now = DateTime.now();
    final requests = _requests[identifier] ?? [];
    
    if (requests.length < maxRequests) {
      return null;
    }
    
    final oldestRequest = requests.first;
    final timeSinceOldest = now.difference(oldestRequest);
    
    if (timeSinceOldest > timeWindow) {
      return null;
    }
    
    return timeWindow - timeSinceOldest;
  }
}

// Content security for web applications
class ContentSecurityPolicy {
  static String get policy {
    return [
      "default-src 'self'",
      "script-src 'self' 'unsafe-inline' 'unsafe-eval'",
      "style-src 'self' 'unsafe-inline'",
      "img-src 'self' data: https:",
      "font-src 'self' data:",
      "connect-src 'self' https://api.example.com",
      "media-src 'self'",
      "object-src 'none'",
      "base-uri 'self'",
      "form-action 'self'",
      "frame-ancestors 'none'",
      "upgrade-insecure-requests",
    ].join('; ');
  }
}
```

## Integration Patterns

### Server-Side Dart with Shelf

```dart
// Server-side Dart application
import 'dart:io';
import 'dart:convert';
import 'package:shelf/shelf.dart';
import 'package:shelf/shelf_io.dart' as shelf_io;
import 'package:shelf_router/shelf_router.dart';
import 'package:shelf_cors_headers/shelf_cors_headers.dart';
import 'package:shelf_static/shelf_static.dart';

// User service for server-side
class UserService {
  final Map<String, User> _users = {};
  
  UserService() {
    _seedUsers();
  }
  
  void _seedUsers() {
    _users['1'] = const User(
      id: '1',
      name: 'John Doe',
      email: 'john@example.com',
      avatarUrl: 'https://example.com/avatar1.jpg',
    );
    _users['2'] = const User(
      id: '2',
      name: 'Jane Smith',
      email: 'jane@example.com',
      avatarUrl: 'https://example.com/avatar2.jpg',
    );
  }
  
  List<User> getUsers({int page = 1, int limit = 20}) {
    final skip = (page - 1) * limit;
    return _users.values.skip(skip).take(limit).toList();
  }
  
  User? getUser(String id) {
    return _users[id];
  }
  
  User createUser(CreateUserRequest request) {
    final id = (_users.keys.length + 1).toString();
    final user = User(
      id: id,
      name: request.name,
      email: request.email,
      avatarUrl: request.avatarUrl ?? '',
    );
    _users[id] = user;
    return user;
  }
  
  User? updateUser(String id, UpdateUserRequest request) {
    final user = _users[id];
    if (user == null) return null;
    
    final updatedUser = user.copyWith(
      name: request.name ?? user.name,
      email: request.email ?? user.email,
      avatarUrl: request.avatarUrl ?? user.avatarUrl,
    );
    
    _users[id] = updatedUser;
    return updatedUser;
  }
  
  bool deleteUser(String id) {
    return _users.remove(id) != null;
  }
}

// Request handlers
Handler getUsersHandler(UserService userService) {
  return (Request request) async {
    final page = int.tryParse(request.url.queryParameters['page'] ?? '1') ?? 1;
    final limit = int.tryParse(request.url.queryParameters['limit'] ?? '20') ?? 20;
    
    final users = userService.getUsers(page: page, limit: limit);
    
    final response = {
      'users': users.map((u) => u.toJson()).toList(),
      'page': page,
      'limit': limit,
      'total': users.length,
    };
    
    return Response.ok(
      json.encode(response),
      headers: {'content-type': 'application/json'},
    );
  };
}

Handler getUserHandler(UserService userService) {
  return (Request request) {
    final id = request.params['id']!;
    final user = userService.getUser(id);
    
    if (user == null) {
      return Response.notFound(json.encode({'error': 'User not found'}));
    }
    
    return Response.ok(
      json.encode(user.toJson()),
      headers: {'content-type': 'application/json'},
    );
  };
}

Handler createUserHandler(UserService userService) {
  return (Request request) async {
    try {
      final body = await request.readAsString();
      final jsonData = json.decode(body) as Map<String, dynamic>;
      final createRequest = CreateUserRequest.fromJson(jsonData);
      
      final user = userService.createUser(createRequest);
      
      return Response(201,
        body: json.encode(user.toJson()),
        headers: {'content-type': 'application/json'},
      );
    } catch (e) {
      return Response(400,
        body: json.encode({'error': 'Invalid request data: $e'}),
        headers: {'content-type': 'application/json'},
      );
    }
  };
}

// Router setup
Router setupRouter(UserService userService) {
  final router = Router();
  
  // CORS headers
  router.all('/<ignored|.*>', (Request request) {
    return Response.ok(null);
  });
  
  // API routes
  router.get('/users', getUsersHandler(userService));
  router.get('/users/<id>', getUserHandler(userService));
  router.post('/users', createUserHandler(userService));
  router.put('/users/<id>', updateUserHandler(userService));
  router.delete('/users/<id>', deleteUserHandler(userService));
  
  // Static file serving
  router.get('/<.*>', (Request request) {
    final path = request.url.path;
    if (path.startsWith('/api/')) {
      return Response.notFound('Not found');
    }
    
    return staticHandler('web')(request);
  });
  
  return router;
}

// Static file handler
Handler staticHandler(String directory) {
  return createStaticHandler(directory, defaultDocument: 'index.html');
}

// Main server function
Future<void> main() async {
  final userService = UserService();
  final router = setupRouter(userService);
  
  // Add CORS middleware
  final handler = const Pipeline()
      .addMiddleware(corsHeaders())
      .addMiddleware(logRequests())
      .addHandler(router);
  
  // Start server
  final server = await shelf_io.serve(
    handler,
    InternetAddress.anyIPv4,
    8080,
  );
  
  print('Server listening on port ${server.port}');
}

// Logging middleware
Middleware logRequests() {
  return (Handler innerHandler) {
    return (Request request) async {
      final startTime = DateTime.now();
      
      try {
        final response = await innerHandler(request);
        final duration = DateTime.now().difference(startTime);
        
        print(
          '${request.method} ${request.requestedUri} -> '
          '${response.statusCode} (${duration.inMilliseconds}ms)',
        );
        
        return response;
      } catch (error, stackTrace) {
        final duration = DateTime.now().difference(startTime);
        
        print(
          '${request.method} ${request.requestedUri} -> ERROR ($duration): $error',
        );
        
        rethrow;
      }
    };
  };
}
```

### WebSockets for Real-time Communication

```dart
// WebSocket server implementation
import 'dart:io';
import 'dart:convert';

class WebSocketServer {
  late HttpServer _server;
  final Set<WebSocket> _connections = {};
  
  Future<void> start(int port) async {
    _server = await HttpServer.bind(InternetAddress.anyIPv4, port);
    print('WebSocket server listening on port $port');
    
    await for (HttpRequest request in _server) {
      if (request.uri.path == '/ws') {
        await _handleWebSocket(request);
      } else {
        request.response.statusCode = HttpStatus.notFound;
        await request.response.close();
      }
    }
  }
  
  Future<void> _handleWebSocket(HttpRequest request) async {
    try {
      final webSocket = await WebSocketTransformer.upgrade(request);
      _connections.add(webSocket);
      
      print('New WebSocket connection: ${webSocket.hashCode}');
      
      // Send welcome message
      _sendMessage(webSocket, {
        'type': 'welcome',
        'message': 'Connected to chat server',
        'timestamp': DateTime.now().toIso8601String(),
      });
      
      // Listen for messages
      webSocket.listen(
        (data) => _handleMessage(webSocket, data),
        onDone: () => _handleDisconnection(webSocket),
        onError: (error) => print('WebSocket error: $error'),
      );
    } catch (e) {
      print('Failed to upgrade to WebSocket: $e');
      request.response.statusCode = HttpStatus.internalServerError;
      await request.response.close();
    }
  }
  
  void _handleMessage(WebSocket webSocket, dynamic data) {
    try {
      final message = json.decode(data as String) as Map<String, dynamic>;
      final messageWithTimestamp = {
        ...message,
        'timestamp': DateTime.now().toIso8601String(),
      };
      
      print('Received message: $messageWithTimestamp');
      
      // Broadcast message to all connected clients
      _broadcastMessage(messageWithTimestamp);
    } catch (e) {
      print('Error handling message: $e');
      _sendMessage(webSocket, {
        'type': 'error',
        'message': 'Invalid message format',
        'timestamp': DateTime.now().toIso8601String(),
      });
    }
  }
  
  void _handleDisconnection(WebSocket webSocket) {
    _connections.remove(webSocket);
    print('WebSocket disconnected: ${webSocket.hashCode}');
    
    _broadcastMessage({
      'type': 'user_disconnected',
      'message': 'A user left the chat',
      'timestamp': DateTime.now().toIso8601String(),
    });
  }
  
  void _sendMessage(WebSocket webSocket, Map<String, dynamic> message) {
    try {
      webSocket.add(json.encode(message));
    } catch (e) {
      print('Error sending message: $e');
    }
  }
  
  void _broadcastMessage(Map<String, dynamic> message) {
    final messageString = json.encode(message);
    
    for (final connection in _connections) {
      try {
        connection.add(messageString);
      } catch (e) {
        print('Error broadcasting message: $e');
      }
    }
  }
  
  Future<void> stop() async {
    await _server.close();
    for (final connection in _connections) {
      await connection.close();
    }
  }
}

// Chat message models
class ChatMessage {
  final String id;
  final String username;
  final String content;
  final DateTime timestamp;
  
  ChatMessage({
    required this.id,
    required this.username,
    required this.content,
    required this.timestamp,
  });
  
  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'username': username,
      'content': content,
      'timestamp': timestamp.toIso8601String(),
    };
  }
  
  factory ChatMessage.fromJson(Map<String, dynamic> json) {
    return ChatMessage(
      id: json['id'] as String,
      username: json['username'] as String,
      content: json['content'] as String,
      timestamp: DateTime.parse(json['timestamp'] as String),
    );
  }
}

// WebSocket client for Flutter
class WebSocketService {
  WebSocketChannel? _channel;
  final StreamController<ChatMessage> _messageController = StreamController<ChatMessage>.broadcast();
  final StreamController<String> _statusController = StreamController<String>.broadcast();
  
  Stream<ChatMessage> get messageStream => _messageController.stream;
  Stream<String> get statusStream => _statusController.stream;
  
  Future<void> connect(String url) async {
    try {
      _channel = WebSocketChannel.connect(Uri.parse(url));
      _statusController.add('Connected');
      
      _channel!.stream.listen(
        (data) {
          try {
            final message = ChatMessage.fromJson(json.decode(data));
            _messageController.add(message);
          } catch (e) {
            print('Error parsing message: $e');
          }
        },
        onDone: () {
          _statusController.add('Disconnected');
        },
        onError: (error) {
          _statusController.add('Error: $error');
        },
      );
    } catch (e) {
      _statusController.add('Connection failed: $e');
    }
  }
  
  void sendMessage(ChatMessage message) {
    if (_channel != null) {
      _channel!.sink.add(json.encode(message.toJson()));
    }
  }
  
  void disconnect() {
    _channel?.sink.close();
    _channel = null;
  }
  
  void dispose() {
    disconnect();
    _messageController.close();
    _statusController.close();
  }
}
```

## Modern Development Workflow

### Project Configuration

```yaml
# pubspec.yaml
name: my_flutter_app
description: A comprehensive Flutter application
publish_to: 'none'

version: 1.0.0+1

environment:
  sdk: '>=3.5.0 <4.0.0'
  flutter: ">=3.24.0"

dependencies:
  flutter:
    sdk: flutter
  
  # State management
  flutter_riverpod: ^2.5.0
  provider: ^6.1.2
  
  # Navigation
  go_router: ^13.2.0
  
  # HTTP client
  dio: ^5.4.3+1
  retrofit: ^4.0.3
  json_annotation: ^4.8.1
  
  # Local storage
  shared_preferences: ^2.2.3
  flutter_secure_storage: ^9.0.0
  
  # Database
  sqflite: ^2.3.3
  drift: ^2.17.0
  
  # Authentication
  local_auth: ^2.2.0
  
  # UI components
  material_color_utilities: ^0.8.0
  flutter_svg: ^2.0.10+1
  
  # Utilities
  uuid: ^4.4.0
  intl: ^0.19.0
  equatable: ^2.0.5
  json_serializable: ^6.7.1
  
  # Testing
  flutter_test:
    sdk: flutter
  mockito: ^5.4.4
  build_runner: ^2.4.9
  retrofit_generator: ^8.0.6
  json_serializable: ^6.7.1
  drift_dev: ^2.17.0

dev_dependencies:
  flutter_test:
    sdk: flutter
  
  # Code generation
  build_runner: ^2.4.9
  retrofit_generator: ^8.0.6
  json_serializable: ^6.7.1
  drift_dev: ^2.17.0
  
  # Linting and formatting
  flutter_lints: ^4.0.0
  very_good_analysis: ^5.1.0
  
  # Testing
  integration_test:
    sdk: flutter
  golden_toolkit: ^0.15.0
  network_image_mock: ^2.1.1

flutter:
  uses-material-design: true
  
  assets:
    - assets/images/
    - assets/icons/
    - assets/config/
  
  fonts:
    - family: Roboto
      fonts:
        - asset: fonts/Roboto-Regular.ttf
        - asset: fonts/Roboto-Bold.ttf
          weight: 700
```

### Analysis Configuration

```yaml
# analysis_options.yaml
include: package:very_good_analysis/analysis_options.yaml

analyzer:
  exclude:
    - "**/*.g.dart"
    - "**/*.freezed.dart"
  
  language:
    strict-casts: true
    strict-inference: true
    strict-raw-types: true
  
  errors:
    invalid_annotation_target: ignore
    missing_required_param: error
    missing_return: error
    todo: ignore

linter:
  rules:
    # Additional rules beyond very_good_analysis
    prefer_single_quotes: true
    sort_constructors_first: true
    sort_unnamed_constructors_first: true
    always_declare_return_types: true
    avoid_print: true
    avoid_unnecessary_containers: true
    sized_box_for_whitespace: true
    use_key_in_widget_constructors: true
    prefer_const_constructors: true
    prefer_const_declarations: true
    prefer_const_literals_to_create_immutables: true
    avoid_web_libraries_in_flutter: true
    prefer_const_constructors_in_immutables: true
    prefer_final_fields: true
    use_full_hex_values_for_flutter_colors: true
```

### CI/CD Configuration

```yaml
# .github/workflows/flutter.yml
name: Flutter CI/CD

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        channel: stable
        flutter-version: '3.24.x'
    
    - name: Install dependencies
      run: flutter pub get
    
    - name: Generate code
      run: flutter packages pub run build_runner build --delete-conflicting-outputs
    
    - name: Analyze code
      run: flutter analyze
    
    - name: Run tests
      run: flutter test --coverage --test-randomize-ordering-seed random
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: coverage/lcov.info
    
    - name: Run widget tests
      run: flutter test integration_test/
  
  build_android:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        channel: stable
        flutter-version: '3.24.x'
    
    - name: Install dependencies
      run: flutter pub get
    
    - name: Generate code
      run: flutter packages pub run build_runner build --delete-conflicting-outputs
    
    - name: Build APK
      run: flutter build apk --release
    
    - name: Build App Bundle
      run: flutter build appbundle --release
    
    - name: Upload APK
      uses: actions/upload-artifact@v3
      with:
        name: android-apk
        path: build/app/outputs/flutter-apk/app-release.apk
    
    - name: Upload App Bundle
      uses: actions/upload-artifact@v3
      with:
        name: android-aab
        path: build/app/outputs/bundle/release/app-release.aab
  
  build_ios:
    needs: test
    runs-on: macos-latest
    if: github.ref == 'refs/heads/main'
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        channel: stable
        flutter-version: '3.24.x'
    
    - name: Install dependencies
      run: flutter pub get
    
    - name: Generate code
      run: flutter packages pub run build_runner build --delete-conflicting-outputs
    
    - name: Build iOS
      run: flutter build ios --release --no-codesign
    
    - name: Upload iOS build
      uses: actions/upload-artifact@v3
      with:
        name: ios-build
        path: build/ios/iphoneos/Runner.app
  
  build_web:
    needs: test
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v4
    
    - name: Setup Flutter
      uses: subosito/flutter-action@v2
      with:
        channel: stable
        flutter-version: '3.24.x'
    
    - name: Install dependencies
      run: flutter pub get
    
    - name: Generate code
      run: flutter packages pub run build_runner build --delete-conflicting-outputs
    
    - name: Build web
      run: flutter build web --release
    
    - name: Deploy to GitHub Pages
      uses: peaceiris/actions-gh-pages@v3
      with:
        github_token: ${{ secrets.GITHUB_TOKEN }}
        publish_dir: build/web
```

---

**Created by**: MoAI Language Skill Factory  
**Last Updated**: 2025-11-06  
**Version**: 2.0.0  
**Dart Target**: 3.5.x with Flutter 3.24.x and modern async patterns  

This skill provides comprehensive Dart development guidance with 2025 best practices, covering everything from Flutter mobile applications to server-side development and real-time communication with WebSockets.
