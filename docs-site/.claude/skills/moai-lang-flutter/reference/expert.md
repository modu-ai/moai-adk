# Flutter Expert Patterns

## Role Definition

Senior Flutter developer with 6+ years of experience specializing in Flutter 3.19+, Riverpod 2.0, GoRouter, and building apps for iOS, Android, Web, and Desktop.

## Core Workflow

1. **Setup** - Project structure, dependencies, routing
2. **State** - Riverpod providers or Bloc setup
3. **Widgets** - Reusable, const-optimized components
4. **Test** - Widget tests, integration tests
5. **Optimize** - Profile, reduce rebuilds

## MUST DO

- Use const constructors wherever possible
- Implement proper keys for lists
- Use Consumer/ConsumerWidget for state (not StatefulWidget)
- Follow Material/Cupertino design guidelines
- Profile with DevTools, fix jank
- Test widgets with flutter_test

## MUST NOT DO

- Build widgets inside build() method
- Mutate state directly (always create new instances)
- Use setState for app-wide state
- Skip const on static widgets
- Ignore platform-specific behavior
- Block UI thread with heavy computation (use compute())

## Riverpod State Management

### Provider Definitions

Import riverpod_annotation and add part directive for generated file. Use @riverpod annotation on functions to create simple providers. Return repository instances reading other providers with ref.read. Create async providers by returning Future types. For stateful providers, create classes extending the generated underscore class, override build method for initial state, and add methods that modify state using AsyncValue.guard.

### Widget Integration

Create ConsumerWidget subclasses. In build method, receive WidgetRef as second parameter. Use ref.watch to observe provider values. Handle AsyncValue with when method providing data, loading, and error callbacks. Use ref.invalidate to refresh data. Use ref.listen in StatefulWidget for side effects like showing snackbars.

### StatefulWidget with Riverpod

Extend ConsumerStatefulWidget and ConsumerState. Initialize TextEditingController in initState and dispose in dispose. Use ref.listen in build for side effects. Check isLoading state to disable buttons during async operations. Access notifier methods with ref.read and the notifier getter.

## GoRouter Navigation

### Router Configuration

Create GoRouter instance with initialLocation. Define routes array with GoRoute objects containing path, name, and builder. Nest child routes in routes array. Use ShellRoute for persistent navigation shells. Create pageBuilder using NoTransitionPage for tab navigation. Implement redirect callback checking authentication state and returning redirect path or null to allow navigation. Define errorBuilder for error handling.

### Navigation methods

Use context.go for declarative navigation with path. Use context.canPop to check before context.pop.

## Widget Patterns

### Reusable Widgets

Create small, focused widgets instead of large complex ones:
- Improves performance with `const` widgets
- Makes testing and refactoring easier
- Share common components across different layouts

### Const Optimization

Use const constructors wherever possible:
- Reduces widget rebuilds
- Improves performance
- Allows widget tree optimizations

### Keys

Implement proper keys for lists:
- Use ValueKey for items with unique IDs
- Use ObjectKey for complex objects
- Use UniqueKey for forcing rebuilds

## Platform Channels

### Dart Implementation

Define class with MethodChannel and EventChannel constants using channel name strings. Create async methods that invoke methods on the channel with invokeMethod, catching PlatformException. For streaming data, call receiveBroadcastStream on EventChannel and map events to typed objects. Set up method call handler with setMethodCallHandler to receive calls from native code.

## Performance Optimization

### DevTools Profiling

- Profile with Flutter DevTools
- Fix jank (missed frames)
- Reduce unnecessary rebuilds
- Optimize widget tree

### Compute for Heavy Operations

Use compute() for CPU-intensive tasks:
- Runs code in separate isolate
- Prevents UI jank
- Returns result to main thread

### Image Optimization

- Use cached_network_image for network images
- Implement image placeholders
- Optimize image sizes
- Use appropriate image formats

## Testing

### Widget Tests

In test main function, create ProviderContainer with overrides for mock providers. Use tester.pumpWidget with UncontrolledProviderScope wrapping MaterialApp with the widget under test. Assert initial loading state with find.byType. Call tester.pumpAndSettle to wait for async operations. Assert final state with find.text.

### Integration Tests

Use integration_test for:
- End-to-end workflows
- Multi-screen interactions
- Performance testing
- Platform-specific behavior

## Knowledge Reference

Flutter 3.19+, Dart 3.3+, Riverpod 2.0, Bloc 8.x, GoRouter, freezed, json_serializable, Dio, flutter_hooks
