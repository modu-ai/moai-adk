---
id: SPEC-V3R3-COV-001
title: Mobile Native Coverage — expert-mobile + iOS/Android/RN/Flutter strategy
version: "1.0.0"
status: implemented
created: 2026-04-25
updated: 2026-04-26
author: manager-spec
priority: P1 High
phase: "v3.0.0 R3 — Phase A — Coverage Gap Closure"
module: ".claude/agents/moai/expert-mobile.md, .claude/skills/moai-domain-mobile/, .claude/skills/moai-framework-react-native/, .claude/skills/moai-framework-flutter-deep/, .claude/skills/moai/SKILL.md (routing keywords)"
dependencies:
  - SPEC-V3R3-ARCH-003
breaking: false
bc_id: []
lifecycle: spec-anchored
tags: "expert-mobile, mobile-coverage, ios, android, react-native, flutter, skill-creation, agent-creation, v3r3, phase-a"
related_theme: "Phase A — Mobile Domain Coverage"
released_in: v2.15.0
---

# SPEC-V3R3-COV-001: Mobile Native Coverage

## HISTORY

| Version | Date       | Author       | Description |
|---------|------------|--------------|-------------|
| 1.0.0   | 2026-04-25 | manager-spec | Initial draft. Phase A P1 — expert-mobile agent + 3 mobile skills + 4-mobile-strategy guide. Zero coverage 해소. |

---

## 1. Goal (목적)

MoAI-ADK v3.0.0 R3에서 **mobile native development domain의 zero coverage 결함을 해소**한다. iOS native (Swift), Android native (Kotlin), React Native, Flutter — 4가지 모바일 개발 패러다임을 single expert agent (`expert-mobile`)로 통합 라우팅하고, 3개 mobile skill (`moai-domain-mobile`, `moai-framework-react-native`, `moai-framework-flutter-deep`)을 신설하여 progressive disclosure 시스템에 편입한다. 4-mobile-strategy comparison guide를 포함하여 어떤 paradigm을 선택할지 의사결정 지원.

본 SPEC은 SPEC-V3R3-ARCH-003 (Expert Tool Capability Uplift)의 Escalation Protocol contract를 활용하여 expert-mobile이 다른 expert와 cross-call 가능한 인프라를 전제로 한다.

### 1.1 배경

- 현재 MoAI-ADK는 expert-backend, expert-frontend, expert-testing 등 8개 expert agent를 보유하나 **mobile domain expert 부재**.
- moai-framework-flutter는 SPEC-FW-FLUTTER-001로 partial 존재하나 (기본 Flutter UI), iOS/Android native + React Native + Flutter 풀스택 (state mgmt, navigation, native bridges) 지원 부재.
- 16-language neutrality 원칙상 swift, kotlin은 supported language로 등록되어 있으나 expert-level 가이드 없음 (CLAUDE.local.md §15).
- 사용자가 mobile feature 요청 시 expert-frontend 또는 expert-backend로 잘못 라우팅되는 risk.

### 1.2 비목표 (Non-Goals)

- 기존 SPEC-FW-FLUTTER-001 본문 변경 금지 (별도 SPEC; 본 SPEC은 `moai-framework-flutter-deep`라는 별도 skill 신설)
- iOS/Android native build 자동화 (Xcode, Gradle 등) 구현 금지
- Mobile CI/CD pipeline 자동화 (Fastlane 등) 본 SPEC 범위 밖
- Mobile testing framework 통합 (XCTest, Espresso 등) — 별도 SPEC
- expert-mobile에 Mobile-specific MCP tool 추가 금지 (본 SPEC은 standard tools만)
- App Store / Play Store deployment 가이드 — 별도 SPEC

---

## 2. Scope (범위)

### 2.1 In Scope

- **Owns**: `.claude/agents/moai/expert-mobile.md` 신설 (description, triggers, tools 표준 패턴)
- **Owns**: `.claude/skills/moai-domain-mobile/SKILL.md` + `modules/{ios-native,android-native,react-native,flutter}.md` 신설
- **Owns**: `.claude/skills/moai-framework-react-native/SKILL.md` + `modules/` 신설
- **Owns**: `.claude/skills/moai-framework-flutter-deep/SKILL.md` + `modules/` 신설 (Firebase 외 풀스택)
- **Owns**: 4-mobile-strategy comparison guide (`moai-domain-mobile/modules/strategy-comparison.md`) — iOS native vs Android native vs RN vs Flutter selection criteria
- **Owns**: Template 동기화 (모든 신규 파일 `internal/template/templates/.claude/` 미러)
- **Owns**: Routing keyword 추가 (`.claude/skills/moai/SKILL.md`에 mobile, ios, android, swift, kotlin, react-native, flutter 매칭)

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- 기존 SPEC-FW-FLUTTER-001과 직접 충돌 (모두 보존, deep skill은 별도 path)
- iOS/Android Xcode/Gradle build automation
- Mobile CI/CD (Fastlane, Codemagic 등)
- Mobile-specific MCP tools (Pencil 외)
- 모바일 시각 디자인 도구 (디자인은 designer skill 별도)
- App store deployment scripts
- 16개 언어 중 모바일 외 언어 변경

---

## 3. Environment

- File system writable: `.claude/agents/moai/`, `.claude/skills/`, `internal/template/templates/.claude/`
- `make build` 가능 환경

## 4. Assumptions

- SPEC-V3R3-ARCH-003 완료 후 진행 (expert agent에 Agent tool + Escalation Protocol baseline 갖춤)
- Existing skill patterns (moai-domain-backend, moai-domain-frontend) follow consistent structure
- routing keyword 시스템 (`.claude/skills/moai/SKILL.md`)이 키워드 매칭으로 expert agent dispatch
- 4 mobile paradigms (iOS native / Android native / RN / Flutter) 모두 활성 user demand 존재

## 5. Requirements (EARS)

### REQ-COV001-001 (Ubiquitous)

The MoAI-ADK agent catalog **shall** include `expert-mobile` agent at `.claude/agents/moai/expert-mobile.md` and template pair, with frontmatter description, triggers (en/ko/ja/zh), tools (Read, Write, Edit, Bash, Agent, Grep, Glob, WebFetch, WebSearch, Skill, mcp__sequential-thinking, mcp__context7), model, permissionMode.

### REQ-COV001-002 (Ubiquitous)

The expert-mobile agent **shall** include the following trigger keywords: `mobile, ios, android, swift, kotlin, react-native, flutter, swiftui, jetpack compose, expo, dart` (English), plus equivalents in Korean (`모바일, 아이폰, 안드로이드`), Japanese (`モバイル, アイフォン, アンドロイド`), Chinese (`移动, 移动应用`).

### REQ-COV001-003 (Ubiquitous)

The MoAI-ADK skill collection **shall** include `moai-domain-mobile` skill at `.claude/skills/moai-domain-mobile/SKILL.md` plus 4 modules under `modules/`: `ios-native.md`, `android-native.md`, `react-native.md`, `flutter.md`.

### REQ-COV001-004 (Ubiquitous)

The MoAI-ADK skill collection **shall** include `moai-framework-react-native` skill at `.claude/skills/moai-framework-react-native/SKILL.md` with at least one module file covering navigation + state management.

### REQ-COV001-005 (Ubiquitous)

The MoAI-ADK skill collection **shall** include `moai-framework-flutter-deep` skill at `.claude/skills/moai-framework-flutter-deep/SKILL.md` with full-stack Flutter coverage (state management, navigation, networking) excluding Firebase (covered separately).

### REQ-COV001-006 (Ubiquitous)

The 4-mobile-strategy comparison guide **shall** be located at `.claude/skills/moai-domain-mobile/modules/strategy-comparison.md` and include selection criteria across 4 paradigms (iOS native, Android native, React Native, Flutter) covering: performance, development cost, team skill, platform features, code reuse %.

### REQ-COV001-007 (Ubiquitous)

The `.claude/skills/moai/SKILL.md` routing keywords **shall** include mobile-related keywords (mobile, ios, android, swift, kotlin, react-native, flutter) such that user requests with these keywords route to expert-mobile.

### REQ-COV001-008 (Event-Driven)

**When** any of the new mobile skill files or the expert-mobile agent file is created, the change **shall** be applied to both `.claude/` and `internal/template/templates/.claude/` paths in the same commit.

### REQ-COV001-009 (Event-Driven)

**When** the expert-mobile agent encounters work outside its mobile domain (e.g., backend API design), it **shall** invoke `Agent()` to delegate to the appropriate T2 expert (per SPEC-V3R3-ARCH-003 Escalation Protocol).

### REQ-COV001-010 (Optional)

**Where** a project's mobile target spans multiple paradigms (e.g., iOS native + React Native shared codebase), the expert-mobile agent **shall** consult the 4-mobile-strategy comparison guide before recommending an approach.

### REQ-COV001-011 (Unwanted)

The new mobile skills **shall not** modify or supersede the existing SPEC-FW-FLUTTER-001 (`moai-framework-flutter`) skill. The new `moai-framework-flutter-deep` is a separate skill targeting full-stack Flutter (different scope).

### REQ-COV001-012 (Unwanted)

The new mobile coverage **shall not** introduce mobile-specific MCP tools into expert-mobile's frontmatter beyond the standard set used by other expert agents.

---

## 6. Acceptance Criteria (요약)

전체 acceptance.md 참조. 핵심:

- AC-COV001-01: expert-mobile agent 파일 존재 (local + template)
- AC-COV001-02: 3 신규 skill 모두 SKILL.md + 최소 module 파일 존재
- AC-COV001-03: 4 mobile paradigm module 모두 존재 (`moai-domain-mobile/modules/`)
- AC-COV001-04: 4-mobile-strategy comparison guide 존재 + 5 selection criteria 포함
- AC-COV001-05: routing keyword가 `.claude/skills/moai/SKILL.md`에 추가됨
- AC-COV001-06: Template + local 동기화 완료
- AC-COV001-07: make build + go test 통과
- AC-COV001-08: 기존 SPEC-FW-FLUTTER-001 무수정 (`moai-framework-flutter` skill body)
- AC-COV001-09: expert-mobile frontmatter에 Agent tool 포함 (ARCH-003 baseline 활용)

---

## 7. Constraints

- **C1**: SPEC-V3R3-ARCH-003 완료 후 진행 (Escalation Protocol baseline 활용)
- **C2**: 16-language neutrality — mobile skill은 swift, kotlin, dart, javascript/typescript 모두 동등 취급
- **C3**: 신규 skill의 SKILL.md는 progressive_disclosure 블록 포함 (DEF-007 baseline 준수, level1_tokens: 100, level2_tokens: 5000)
- **C4**: expert-mobile 본문에 Escalation Protocol 섹션 포함 (ARCH-003 패턴)
- **C5**: 4-mobile-strategy guide는 객관적 비교표 형식, 특정 paradigm 옹호 금지
- **C6**: 기존 `moai-framework-flutter` skill과 명확 구분: 본 SPEC의 `moai-framework-flutter-deep`은 별도 skill (이름 + 디렉터리 모두 분리)

---

## 8. Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| 기존 moai-framework-flutter와 confusing overlap | Medium | 명명 명확화 (`flutter` vs `flutter-deep`), 각 SKILL.md description에 scope 명시 |
| 4 mobile paradigm 동등 취급 어려움 (실제 user demand 차이) | Low | strategy-comparison.md에 객관적 비교, 추천 회피 |
| expert-mobile triggers가 너무 광범위 → false dispatch | Medium | 명시적 exclusion ("NOT for: web frontend, backend APIs, CLI tools") 명시 |
| skill 신규 생성 시 progressive_disclosure 블록 누락 | Low | DEF-007 baseline 준수, 작성 시 즉시 포함 |
| Template / local drift | Medium | 모든 신규 파일은 동시에 두 경로에 작성 |
| Routing keyword 충돌 (mobile vs frontend) | Medium | 우선순위 명시; mobile 키워드 매치 우선 |

---

## 9. Dependencies

- **선행**: SPEC-V3R3-ARCH-003 (expert agent의 Agent tool + Escalation Protocol baseline)

후속:
- 별도 SPEC: Mobile testing framework (XCTest, Espresso) 통합
- 별도 SPEC: App Store / Play Store deployment automation
- 별도 SPEC: Mobile CI/CD (Fastlane)

---

## 10. Traceability

| REQ ID | Acceptance Criteria | Source |
|--------|---------------------|--------|
| REQ-COV001-001 | AC-COV001-01 | Master plan §1.4 mobile coverage |
| REQ-COV001-002 | AC-COV001-01 (triggers) | 16-language neutrality + triggers pattern |
| REQ-COV001-003 | AC-COV001-02, 03 | Master plan, 4 paradigm modules |
| REQ-COV001-004 | AC-COV001-02 | RN as separate framework skill |
| REQ-COV001-005 | AC-COV001-02 | Flutter deep coverage |
| REQ-COV001-006 | AC-COV001-04 | Strategy comparison guide |
| REQ-COV001-007 | AC-COV001-05 | Routing keyword integration |
| REQ-COV001-008 | AC-COV001-06 | CLAUDE.local.md §2 Template-First |
| REQ-COV001-009 | AC-COV001-09 (Escalation) | SPEC-V3R3-ARCH-003 baseline |
| REQ-COV001-010 | (recommendation only) | Cross-paradigm guidance |
| REQ-COV001-011 | AC-COV001-08 | 기존 SPEC 무수정 |
| REQ-COV001-012 | AC-COV001-01 (tools 표준) | Standard tool set |
