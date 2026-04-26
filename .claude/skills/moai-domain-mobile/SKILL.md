---
name: moai-domain-mobile
description: >
  Mobile native and cross-platform development domain skill covering iOS native
  (Swift/SwiftUI), Android native (Kotlin/Jetpack Compose), React Native, and
  Flutter. Use when implementing mobile features or selecting between mobile
  paradigms. Harness seed: invoke builder-skill for deep sub-skills.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Bash, Grep, Glob, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-04-26"
  modularized: "true"
  tags: "mobile, ios, android, swift, kotlin, react-native, flutter, dart, cross-platform"
  author: "MoAI-ADK Team"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["mobile", "ios", "android", "swift", "swiftui", "kotlin", "jetpack", "react-native", "flutter", "dart", "expo", "cross-platform", "native app", "모바일", "아이폰", "안드로이드", "플러터"]
  agents: ["expert-mobile"]
---

# Mobile Domain

## Quick Reference

Mobile development across four paradigms — iOS native, Android native, React Native, and Flutter — unified under a single expert routing strategy.

Core Paradigms:

- **iOS native**: Swift, SwiftUI, UIKit, Xcode, Core Data — full Apple platform access
- **Android native**: Kotlin, Jetpack Compose, Coroutines, Room, Gradle — full Google platform access
- **React Native**: JavaScript/TypeScript, Expo, React Navigation — 80-90% code reuse
- **Flutter**: Dart, Riverpod/BLoC, go_router — 90-95% code reuse, pixel-perfect UI

When to Use This Skill:

- Paradigm selection for a new mobile project
- iOS-specific patterns (SwiftUI layouts, navigation, Core Data)
- Android-specific patterns (Compose UI, Coroutines, Hilt DI)
- React Native setup (Expo, navigation, state management)
- Flutter architecture (state management, platform channels)
- Cross-platform strategy consultation

Module References:

- `modules/ios-native.md` — Swift, SwiftUI, UIKit, SPM patterns
- `modules/android-native.md` — Kotlin, Jetpack Compose, Coroutines patterns
- `modules/react-native.md` — Expo, React Navigation, RN architecture
- `modules/flutter.md` — Widget tree, Riverpod, go_router, Dart patterns
- `modules/strategy-comparison.md` — 4-paradigm selection guide

---

## Implementation Guide

### Paradigm Selection Decision Tree

```
New mobile project?
  ├── Single platform (iOS only or Android only)?
  │     ├── iOS → iOS Native (Swift + SwiftUI)
  │     └── Android → Android Native (Kotlin + Compose)
  └── Cross-platform?
        ├── Existing JS/TS team? → React Native (Expo)
        ├── Pixel-perfect UI + performance priority? → Flutter
        └── Team new to mobile? → Flutter (fastest ramp-up)
```

For detailed selection criteria including performance, cost, team skill, platform features, and code reuse percentages, see `modules/strategy-comparison.md`.

### iOS Native Patterns

Key entry points:
- App lifecycle: `@main` struct conforming to `App` protocol
- Navigation: `NavigationStack` with `NavigationPath` (SwiftUI)
- State: `@State`, `@StateObject`, `@EnvironmentObject`, `@Observable` (iOS 17+)
- Data: Core Data or SwiftData (iOS 17+)
- Async: Swift Concurrency (`async/await`, `Task`, `actor`)

See `modules/ios-native.md` for patterns and examples.

### Android Native Patterns

Key entry points:
- App entry: `ComponentActivity` + `setContent { }` with Jetpack Compose
- Navigation: `NavHost` + `NavController` (Compose Navigation)
- State: `ViewModel` + `StateFlow` / `collectAsStateWithLifecycle()`
- DI: Hilt (`@HiltAndroidApp`, `@AndroidEntryPoint`, `@HiltViewModel`)
- Async: Kotlin Coroutines + Flow

See `modules/android-native.md` for patterns and examples.

### React Native Patterns

Key entry points:
- Project setup: `npx create-expo-app` (Expo managed workflow)
- Navigation: React Navigation v7 (`@react-navigation/native`)
- State: Zustand (simple) or Redux Toolkit (complex)
- Native modules: Expo SDK first, then Turbo Modules for custom bridges

See `modules/react-native.md` for patterns and examples.

### Flutter Patterns

Key entry points:
- App entry: `MaterialApp.router` with `go_router`
- State: Riverpod (`ConsumerWidget`, `StateNotifierProvider`, `AsyncNotifierProvider`)
- Networking: Dio with Retrofit-style annotations
- Platform bridge: `MethodChannel` for iOS/Android native calls

See `modules/flutter.md` for patterns and examples.

---

## Harness Seed Pattern: 4-Strategy Router

This skill is a **harness seed** — it provides routing and overview patterns. Deep framework-specific sub-skills are generated on demand via `builder-skill`.

### When to Invoke Harness for Sub-Skill Generation

Generate a deep sub-skill when:
- Project commits to a single paradigm (e.g., Flutter-only)
- Advanced patterns are needed beyond Quick Reference level
- Project uses a specific SDK version requiring tailored examples

```
# Example: Generate deep Flutter skill for this project
Use the builder-skill subagent to create moai-framework-flutter-deep:
  - Full-stack Flutter (state mgmt, navigation, networking, platform channels)
  - Exclude Firebase (covered by moai-platform-auth)
  - Target Flutter 3.x + Dart 3.x

# Example: Generate deep React Native skill
Use the builder-skill subagent to create moai-framework-react-native:
  - React Navigation v7, Redux Toolkit, Turbo Modules, JSI
  - Target React Native 0.74+ New Architecture
```

Generated skills land at `.claude/skills/moai-framework-{paradigm}/SKILL.md` and are automatically registered for the current project session.

---

## Advanced Patterns

### Multi-Paradigm Projects

When a project spans multiple platforms (e.g., iOS native + React Native shared services):

1. Read `modules/strategy-comparison.md` for trade-offs
2. Define shared API contract with expert-backend
3. Implement platform-specific UI in each paradigm
4. Use feature flags for platform-specific behaviors

### Performance Baselines

| Paradigm | UI Render | Startup | Memory |
|----------|-----------|---------|--------|
| iOS Native | 60-120 fps | Fast | Low |
| Android Native | 60-120 fps | Fast | Low |
| React Native | 60 fps (New Arch) | Medium | Medium |
| Flutter | 60-120 fps | Fast | Medium |

### Testing Strategy by Paradigm

- iOS: XCTest (unit), XCUITest (UI) — target 80%+ unit coverage
- Android: JUnit4 + Espresso + Hilt testing — target 80%+ unit coverage
- React Native: Jest + React Native Testing Library — target 80%+ unit coverage
- Flutter: flutter_test + integration_test — target 80%+ unit coverage

---

## Works Well With

- expert-mobile: Primary agent that loads this skill
- expert-backend: API contract definition, REST/GraphQL integration
- expert-security: Mobile-specific security (certificate pinning, biometric auth, keychain/keystore)
- moai-workflow-testing: Test strategy for mobile paradigms
- builder-skill: Generate deep framework-specific sub-skills on demand
