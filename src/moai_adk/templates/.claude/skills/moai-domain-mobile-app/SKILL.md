---
name: moai-domain-mobile-app
description: Mobile app development with Flutter and React Native, state management,
  and native integration
allowed-tools:
- Read
- Bash
---

# Mobile App Expert

## What it does

Provides expertise in cross-platform mobile app development using Flutter (Dart) and React Native (TypeScript), including state management patterns and native module integration.

## When to use

- "모바일 앱 개발", "Flutter 위젯", "React Native 컴포넌트", "상태 관리"
- Automatically invoked when working with mobile app projects
- Mobile app SPEC implementation (`/alfred:2-build`)

## How it works

**Flutter Development**:
- **Widget tree**: StatelessWidget, StatefulWidget
- **State management**: Provider, Riverpod, BLoC
- **Navigation**: Navigator 2.0, go_router
- **Platform-specific code**: MethodChannel

**React Native Development**:
- **Components**: Functional components with hooks
- **State management**: Redux, MobX, Zustand
- **Navigation**: React Navigation
- **Native modules**: Turbo modules, JSI

**Cross-Platform Patterns**:
- **Responsive design**: Adaptive layouts for phone/tablet
- **Performance optimization**: Lazy loading, memoization
- **Offline support**: Local storage, sync strategies
- **Testing**: Widget tests (Flutter), component tests (RN)

**Native Integration**:
- **Plugins**: Platform channels, native modules
- **Permissions**: Camera, location, notifications
- **Deep linking**: Universal links, app links
- **Push notifications**: FCM, APNs

## Examples

### Example 1: Flutter app with BLoC
User: "/alfred:2-build MOBILE-001"
Claude: (creates RED widget test, GREEN implementation with BLoC, REFACTOR)

### Example 2: React Native state management
User: "React Native Redux 설정"
Claude: (sets up Redux with TypeScript and async actions)

## Works well with

- alfred-trust-validation (mobile testing)
- dart-expert (Flutter development)
- typescript-expert (React Native development)
