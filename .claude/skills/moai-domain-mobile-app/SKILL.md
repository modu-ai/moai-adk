---

name: moai-domain-mobile-app
description: Mobile app development with Flutter and React Native, state management, and native integration. Use when working on mobile application flows scenarios.
allowed-tools:
  - Read
  - Bash
---

# Mobile App Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand for mobile flows |
| Trigger cues | iOS/Android releases, cross-platform tooling, app store compliance, mobile UX. |
| Tier | 4 |

## What it does

Provides expertise in cross-platform mobile app development using Flutter (Dart) and React Native (TypeScript), including state management patterns and native module integration.

## When to use

- Engages when mobile application development or release pipelines are in scope.
- “Mobile app development”, “Flutter widgets”, “React Native components”, “state management”
- Automatically invoked when working with mobile app projects
- Mobile app SPEC implementation (`/alfred:2-run`)

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
```markdown
- Generate platform-specific builds (`flutter build`, `xcodebuild`).
- Capture store submission checklist as Todo items.
```

## Inputs
- Domain-specific design documents and user requirements.
- Project technology stack and operational constraints.

## Outputs
- Domain-specific architecture or implementation guidelines.
- Recommended list of associated sub-agents/skills.

## Failure Modes
- When the domain document does not exist or is ambiguous.
- When the project strategy is unconfirmed and cannot be specified.

## Dependencies
- `.moai/project/` document and latest technical briefing are required.

## References
- Apple. "Human Interface Guidelines." https://developer.apple.com/design/human-interface-guidelines/ (accessed 2025-03-29).
- Google. "Material Design." https://m3.material.io/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Codified input/output and failure responses for domain skills.

## Works well with

- alfred-trust-validation (mobile testing)
- dart-expert (Flutter development)
- typescript-expert (React Native development)

## Best Practices
- Record supporting documentation (version/link) for each domain decision.
- Review performance, security, and operational requirements simultaneously at an early stage.
