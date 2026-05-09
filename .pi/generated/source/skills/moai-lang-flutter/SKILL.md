---
name: moai-lang-flutter
description: >
  Flutter 3.24+ / Dart 3.5+ development specialist covering Riverpod,
  go_router, and cross-platform patterns. Use when building cross-platform
  mobile apps, desktop apps, or web applications with Flutter.
license: Apache-2.0
compatibility: Designed for Claude Code
user-invocable: false
metadata:
  version: "1.1.0"
  category: "language"
  status: "active"
  updated: "2026-01-11"
  modularized: "false"
  tags: "flutter, dart, riverpod, cross-platform, mobile, desktop"
  context7-libraries: "/flutter/flutter, /rrousselgit/riverpod, /flutter/packages"
  related-skills: "moai-lang-swift, moai-lang-kotlin"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["Flutter", "Dart", "Riverpod", "go_router", "widget", ".dart", "pubspec.yaml", "cross-platform", "mobile", "adaptive", "responsive", "animation", "hero", "staggered", "physics"]
  languages: ["dart", "flutter"]
---

## Quick Reference (30 seconds)

Flutter/Dart Development Expert - Dart 3.5+, Flutter 3.24+ with modern patterns.

Auto-Triggers: Flutter projects (`.dart` files, `pubspec.yaml`), cross-platform apps, widget development

Core Capabilities:

- Dart 3.5: Patterns, records, sealed classes, extension types
- Flutter 3.24: Widget tree, Material 3, adaptive layouts
- Riverpod: State management with code generation
- go_router: Declarative navigation and deep linking
- Platform Channels: Native iOS/Android integration
- Testing: flutter_test, widget_test, integration_test

## Implementation Guide (5 minutes)

### Dart 3.5 Language Features

Pattern Matching with Sealed Classes: Define a sealed class Result with generic type parameter T. Create subclasses Success containing data field and Failure containing error string. Use switch expressions with pattern matching on the sealed type. Patterns use colon syntax to extract named fields. Add guard clauses with when keyword for conditional matching based on field values.

Records and Destructuring: Define type aliases for records using parentheses with named fields. Create functions returning multiple values as record tuples with positional elements. Use destructuring syntax with parentheses on the left side of assignment. In for loops, destructure named record fields directly using colon syntax.

Extension Types: Define extension types with the type keyword, wrapping a base type in parentheses. Add factory constructors for validation logic. Define getters for computed properties on the underlying value.

### Riverpod State Management

Provider Definitions: Import riverpod_annotation and add part directive for generated file. Use @riverpod annotation on functions to create simple providers. Return repository instances reading other providers with ref.read. Create async providers by returning Future types. For stateful providers, create classes extending the generated underscore class, override build method for initial state, and add methods that modify state using AsyncValue.guard.

Widget Integration: Create ConsumerWidget subclasses. In build method, receive WidgetRef as second parameter. Use ref.watch to observe provider values. Handle AsyncValue with when method providing data, loading, and error callbacks. Use ref.invalidate to refresh data. Use ref.listen in StatefulWidget for side effects like showing snackbars.

StatefulWidget with Riverpod: Extend ConsumerStatefulWidget and ConsumerState. Initialize TextEditingController in initState and dispose in dispose. Use ref.listen in build for side effects. Check isLoading state to disable buttons during async operations. Access notifier methods with ref.read and the notifier getter.

### go_router Navigation

Router Configuration: Create GoRouter instance with initialLocation. Define routes array with GoRoute objects containing path, name, and builder. Nest child routes in routes array. Use ShellRoute for persistent navigation shells. Create pageBuilder using NoTransitionPage for tab navigation. Implement redirect callback checking authentication state and returning redirect path or null to allow navigation. Define errorBuilder for error handling.

Navigation methods: Use context.go for declarative navigation with path. Use context.canPop to check before context.pop.

### Platform Channels

Dart Implementation: Define class with MethodChannel and EventChannel constants using channel name strings. Create async methods that invoke methods on the channel with invokeMethod, catching PlatformException. For streaming data, call receiveBroadcastStream on EventChannel and map events to typed objects. Set up method call handler with setMethodCallHandler to receive calls from native code.

### Widget Patterns

Adaptive Layouts: Create StatelessWidget with required parameters for child, destinations, selectedIndex, and onDestinationSelected callback. In build, get width using MediaQuery.sizeOf. Use conditional returns based on width breakpoints. Under 600 pixels, return Scaffold with NavigationBar in bottomNavigationBar. Under 840 pixels, return Scaffold with Row containing NavigationRail and expanded child. Above 840 pixels, return Scaffold with Row containing NavigationDrawer and expanded child.

### Testing

Widget Test Example: In test main function, create ProviderContainer with overrides for mock providers. Use tester.pumpWidget with UncontrolledProviderScope wrapping MaterialApp with the widget under test. Assert initial loading state with find.byType. Call tester.pumpAndSettle to wait for async operations. Assert final state with find.text.

For comprehensive testing patterns, see [examples.md](examples.md).

## Advanced Patterns

For comprehensive coverage including:

- Adaptive and responsive UIs across all platforms
- Animation patterns (implicit, explicit, hero, staggered, physics)
- Expert-level widget development and optimization
- Clean Architecture with Riverpod
- Isolates for compute-heavy operations
- Custom render objects and painting
- FFI and platform-specific plugins
- Performance optimization and profiling

See: [reference/adaptive.md](reference/adaptive.md) for responsive layouts, [reference/animations.md](reference/animations.md) for animation patterns, [reference/expert.md](reference/expert.md) for expert-level development

## Context7 Library Mappings

Flutter/Dart Core:

- `/flutter/flutter` - Flutter framework
- `/dart-lang/sdk` - Dart SDK

State Management:

- `/rrousselGit/riverpod` - Riverpod state management
- `/felangel/bloc` - BLoC pattern

Navigation and Storage:

- `/flutter/packages` - go_router and official packages
- `/cfug/dio` - HTTP client
- `/isar/isar` - NoSQL database

## Works Well With

- `moai-lang-swift` - iOS native integration for platform channels
- `moai-lang-kotlin` - Android native integration for platform channels
- `moai-domain-backend` - API integration and backend communication
- `moai-quality-security` - Mobile security best practices
- `moai-essentials-debug` - Flutter debugging and DevTools
