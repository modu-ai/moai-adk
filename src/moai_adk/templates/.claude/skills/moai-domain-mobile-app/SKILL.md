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
- 도메인 관련 설계 문서 및 사용자 요구사항.
- 프로젝트 기술 스택 및 운영 제약.

## Outputs
- 도메인 특화 아키텍처 또는 구현 가이드라인.
- 연관 서브 에이전트/스킬 권장 목록.

## Failure Modes
- 도메인 근거 문서가 없거나 모호할 때.
- 프로젝트 전략이 미확정이라 구체화할 수 없을 때.

## Dependencies
- `.moai/project/` 문서와 최신 기술 브리핑이 필요합니다.

## References
- Apple. "Human Interface Guidelines." https://developer.apple.com/design/human-interface-guidelines/ (accessed 2025-03-29).
- Google. "Material Design." https://m3.material.io/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 도메인 스킬에 대한 입력/출력 및 실패 대응을 명문화했습니다.

## Works well with

- alfred-trust-validation (mobile testing)
- dart-expert (Flutter development)
- typescript-expert (React Native development)

## Best Practices
- 도메인 결정 사항마다 근거 문서(버전/링크)를 기록합니다.
- 성능·보안·운영 요구사항을 초기 단계에서 동시에 검토하세요.
