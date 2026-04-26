---
spec_id: SPEC-V3R3-COV-001
title: Implementation Plan — Mobile Native Coverage
version: "1.0.0"
status: draft
created: 2026-04-25
related_spec: .moai/specs/SPEC-V3R3-COV-001/spec.md
---

# Plan — SPEC-V3R3-COV-001

## 1. Objectives

- expert-mobile agent 신설 (4 mobile paradigm 통합 라우팅)
- 3 mobile skills 신설:
  - `moai-domain-mobile` (4 paradigm modules + strategy comparison)
  - `moai-framework-react-native`
  - `moai-framework-flutter-deep` (Flutter 풀스택, Firebase 제외)
- 4-mobile-strategy comparison guide 작성
- routing keyword 통합 (`.claude/skills/moai/SKILL.md`)
- Template + local 동시 적용

## 2. Technical Approach

### 2.1 expert-mobile.md frontmatter (참고: expert-backend 패턴)

```yaml
---
name: expert-mobile
description: |
  Mobile native and cross-platform application specialist. Use PROACTIVELY for iOS native (Swift, SwiftUI), Android native (Kotlin, Jetpack Compose), React Native, and Flutter development.
  MUST INVOKE when ANY of these keywords appear in user request:
  --deepthink flag: Activate Sequential Thinking MCP for mobile architecture decisions, paradigm selection, and cross-platform trade-offs.
  EN: mobile, ios, android, swift, swiftui, kotlin, jetpack compose, react-native, react native, expo, flutter, dart, mobile app, native app
  KO: 모바일, 아이폰, 안드로이드, 스위프트, 코틀린, 플러터, 리액트네이티브
  JA: モバイル, アイフォン, アンドロイド, スイフト, コトリン, フラッター
  ZH: 移动, 移动应用, 苹果, 安卓, 移动开发
  NOT for: web frontend, backend APIs, CLI tools, DevOps/deployment, security audits, mobile CI/CD (별도 SPEC)
tools: Read, Write, Edit, Grep, Glob, WebFetch, WebSearch, Bash, TodoWrite, Skill, Agent, mcp__sequential-thinking__sequentialthinking, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
model: sonnet
permissionMode: bypassPermissions
memory: project
skills:
  - moai-foundation-core
  - moai-domain-mobile
  - moai-framework-react-native
  - moai-framework-flutter-deep
  - moai-workflow-testing
---

# expert-mobile

Mobile native and cross-platform development expert.

## Primary Mission
... (생략, body는 다른 expert agent 패턴)

## Escalation Protocol
... (SPEC-V3R3-ARCH-003 패턴)

## Workflow Steps
... (mobile-specific)
```

### 2.2 moai-domain-mobile/SKILL.md

```yaml
---
name: moai-domain-mobile
description: >
  Mobile native and cross-platform development domain skill covering iOS native, Android native, React Native, and Flutter. Use when implementing mobile features or selecting between mobile paradigms.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Bash, Grep, Glob, mcp__context7__resolve-library-id, mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.0.0"
  category: "domain"
  status: "active"
  updated: "2026-04-25"
  modularized: "true"
  tags: "mobile, ios, android, swift, kotlin, react-native, flutter, dart, cross-platform"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
---

# Mobile Domain

(Quick Reference + Implementation Guide + Advanced + Works Well With)
... (다른 domain skill 패턴)
```

modules:
- `modules/ios-native.md`: SwiftUI, UIKit, Xcode, Swift Package Manager
- `modules/android-native.md`: Jetpack Compose, Gradle, Kotlin coroutines
- `modules/react-native.md`: Expo, React Navigation, RN architecture
- `modules/flutter.md`: Widget tree, Provider/Riverpod, Dart
- `modules/strategy-comparison.md`: 4-paradigm selection guide

### 2.3 moai-framework-react-native/SKILL.md

표준 framework skill 패턴. 최소 modules:
- `modules/navigation.md` (React Navigation)
- `modules/state-management.md` (Redux Toolkit, Zustand, Jotai)
- `modules/native-modules.md` (Bridge, Turbo Modules, JSI)

### 2.4 moai-framework-flutter-deep/SKILL.md

Full-stack Flutter (Firebase 제외):
- `modules/state-management.md` (Provider, Riverpod, BLoC)
- `modules/navigation.md` (go_router, Navigator 2.0)
- `modules/networking.md` (Dio, Retrofit-style)
- `modules/platform-channels.md` (iOS/Android native bridge)

### 2.5 strategy-comparison.md content sketch

| Criterion | iOS native | Android native | React Native | Flutter |
|-----------|------------|----------------|--------------|---------|
| Performance | Best | Best | Good | Very good |
| Development cost | High | High | Medium | Low |
| Team skill | Swift | Kotlin | JS/TS | Dart |
| Platform features | Full | Full | Partial (RN modules) | Full (plugins) |
| Code reuse % | 0% | 0% | 80-90% | 90-95% |
| Hot reload | None | None | Yes | Yes |
| Best for | iOS-first, Apple Watch, complex UX | Android-first, Wear OS | Existing JS team, MVP | Cross-platform from scratch |

### 2.6 routing keyword 추가

`.claude/skills/moai/SKILL.md` 또는 routing 정의 파일에 mobile keyword 매칭 추가:

```yaml
# (skill body 또는 dispatch table)
- pattern: "mobile|ios|android|swift|kotlin|react-native|flutter|swiftui|jetpack"
  expert: expert-mobile
```

(실제 routing 파일 위치는 implementation 시 확인)

## 3. Wave / Phase 설계

### Wave A.1 — expert-mobile agent 신설 (1 task)

- T-A1-1: expert-mobile.md frontmatter + body (template + local pair)

### Wave A.2 — moai-domain-mobile skill 신설 (6 tasks, parallel modules)

- T-A2-1: SKILL.md
- T-A2-2 ~ A2-5: 4 paradigm modules (parallel)
- T-A2-6: strategy-comparison.md

### Wave A.3 — moai-framework-react-native + flutter-deep skills (2 tasks)

- T-A3-1: moai-framework-react-native (SKILL.md + 3 modules)
- T-A3-2: moai-framework-flutter-deep (SKILL.md + 4 modules)

### Wave A.4 — Routing 통합 (1 task)

- T-A4-1: `.claude/skills/moai/SKILL.md` 또는 routing 정의 파일에 mobile keyword 추가

### Wave A.5 — Verification (3 tasks)

- T-A5-1: AC verification 일괄 실행
- T-A5-2: make build + go test
- T-A5-3: 기존 moai-framework-flutter 무수정 확인

## 4. File 영향 요약

| File | Change Type | Path Pair |
|------|-------------|-----------|
| `expert-mobile.md` | NEW | `.claude/agents/moai/` + template |
| `moai-domain-mobile/SKILL.md` | NEW | `.claude/skills/` + template |
| `moai-domain-mobile/modules/ios-native.md` | NEW | both |
| `moai-domain-mobile/modules/android-native.md` | NEW | both |
| `moai-domain-mobile/modules/react-native.md` | NEW | both |
| `moai-domain-mobile/modules/flutter.md` | NEW | both |
| `moai-domain-mobile/modules/strategy-comparison.md` | NEW | both |
| `moai-framework-react-native/SKILL.md` | NEW | both |
| `moai-framework-react-native/modules/*.md` (3 files) | NEW | both |
| `moai-framework-flutter-deep/SKILL.md` | NEW | both |
| `moai-framework-flutter-deep/modules/*.md` (4 files) | NEW | both |
| `.claude/skills/moai/SKILL.md` (또는 routing 정의 파일) | EDIT | both |
| `internal/template/embedded.go` | AUTO-REGEN | `make build` |

총 신규 ~18 × 2 (template+local) = ~36 파일 + 1 EDIT × 2 + 1 AUTO.

## 5. Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| 기존 moai-framework-flutter 충돌 | Medium | 명명 분리 (flutter vs flutter-deep), AC-08 검증 |
| 4 paradigm module 깊이 불균형 | Medium | 모든 module 최소 quick reference + 1 example 패턴 |
| Routing keyword 충돌 (mobile vs frontend) | Medium | 우선순위 명시, false dispatch 시 expert-mobile에서 frontend로 escalate |
| Strategy comparison 편향 | Medium | 객관적 비교표만 작성, 추천 회피 |
| Skill SKILL.md 본문 + frontmatter 일관성 | Low | DEF-007 baseline 사용, progressive_disclosure 블록 자동 포함 |
| Template-local drift | Medium | 각 신규 파일 동시에 두 경로 작성 |
| Routing 정의 파일 위치 불명확 | Medium | T-A4-1 시작 전 grep으로 확인 |

## 6. Open Questions

- OQ1: routing keyword 정의 파일 위치 — `.claude/skills/moai/SKILL.md`? `.claude/skills/moai/workflows/`? → **Decision**: T-A4-1 실행 전 grep 확인. 후보: `.claude/skills/moai/SKILL.md` body 또는 dispatch table.
- OQ2: moai-framework-flutter-deep와 기존 moai-framework-flutter 차이? → **Decision**: 기존은 UI 위주 partial, deep은 풀스택 (state, navigation, networking, platform channels) — Firebase 제외 (별도 SPEC).
- OQ3: Mobile testing skill 본 SPEC 범위? → **Decision**: NO (Out of Scope). 별도 SPEC 권고.
- OQ4: expert-mobile color/model? → **Decision**: model: sonnet (다른 expert와 동일), color는 다른 agent 패턴 따라 미지정 또는 mobile-orange 등 (기존 컨벤션 확인 후 결정)
- OQ5: 4 paradigm 모두 동일 module depth 강제? → **Decision**: Yes, baseline (Quick Reference + 1 핵심 패턴). 깊이 확장은 별도 SPEC.

## 7. Milestones

- M1: expert-mobile.md 작성 + frontmatter validation
- M2: moai-domain-mobile/ 모든 파일 작성 (SKILL + 5 modules)
- M3: moai-framework-react-native, flutter-deep skill 작성
- M4: Routing keyword 통합
- M5: AC verification + make build + go test 통과
- M6: 기존 moai-framework-flutter 무수정 확인

## 8. Definition of Done

- [ ] expert-mobile.md (template + local) 존재 + 표준 frontmatter
- [ ] moai-domain-mobile/SKILL.md + 5 modules (ios, android, rn, flutter, strategy-comparison) 존재
- [ ] moai-framework-react-native/ skill 존재 + 최소 1 module
- [ ] moai-framework-flutter-deep/ skill 존재 + 최소 1 module
- [ ] routing keyword 추가 (mobile, ios, android, swift, kotlin, react-native, flutter)
- [ ] 모든 신규 SKILL.md에 progressive_disclosure 블록 포함 (DEF-007 baseline)
- [ ] expert-mobile frontmatter `tools:` CSV에 Agent 포함 (ARCH-003 baseline)
- [ ] expert-mobile body에 Escalation Protocol 섹션 (ARCH-003 패턴)
- [ ] 기존 moai-framework-flutter 무수정 확인 (`git diff` empty for that path)
- [ ] make build + go test 통과
- [ ] AC-COV001-01 ~ 09 모두 만족
