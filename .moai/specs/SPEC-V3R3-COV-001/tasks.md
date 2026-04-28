---
spec_id: SPEC-V3R3-COV-001
title: Task Decomposition — Mobile Native Coverage
version: "1.0.0"
status: draft
created: 2026-04-25
related_plan: .moai/specs/SPEC-V3R3-COV-001/plan.md
related_spec: .moai/specs/SPEC-V3R3-COV-001/spec.md
---

# 작업 분해 — SPEC-V3R3-COV-001

> **범례**:
> - **File owner**: 단독 소유 파일 경로
> - **Depends on**: 선행 task ID
> - **Wave**: A.1 ~ A.5
> - **Parallel OK**: 동일 Wave 내 병렬 가능 여부

---

## 전체 Task 개요

| Wave | Task 수 | Parallel 가능 | Sequential 필수 |
|------|---------|---------------|-----------------|
| A.1 expert-mobile agent | 1 | — | T-A1-1 |
| A.2 moai-domain-mobile skill | 6 | T-A2-2 ~ 6 (다른 module 파일) parallel | T-A2-1 first |
| A.3 framework skills | 2 | T-A3-1, T-A3-2 (다른 디렉터리) parallel | — |
| A.4 routing 통합 | 1 | — | T-A4-1 |
| A.5 Verification | 3 | T-A5-1 단독, 나머지 sequential | T-A5-1 ~ 3 |

**총 task 수: 13**

---

## Wave A.1 — expert-mobile Agent (1 task)

### T-A1-1: expert-mobile.md 작성 (template + local)
- **File owner**: `.claude/agents/moai/expert-mobile.md`, `internal/template/templates/.claude/agents/moai/expert-mobile.md`
- **Depends on**: 없음 (단, SPEC-V3R3-ARCH-003 완료 전제)
- **Parallel OK**: —
- **Action**:
  1. plan.md §2.1의 frontmatter 작성 (name, description with triggers, tools 포함 Agent, model, permissionMode, memory, skills)
  2. Body 작성: Primary Mission, Core Capabilities, Scope Boundaries, Delegation Protocol, Escalation Protocol (ARCH-003 패턴), Workflow Steps
  3. Triggers 다국어 (en, ko, ja, zh) 모두 포함
  4. Template + local 동시 작성
- **Verification**: AC-COV001-01, AC-COV001-09 통과
- **Rollback**: 두 파일 삭제

### Wave A.1 Checkpoint
- AC-COV001-01, 09 부분 통과

---

## Wave A.2 — moai-domain-mobile Skill (6 tasks)

### T-A2-1: SKILL.md 작성 (template + local)
- **File owner**: `.claude/skills/moai-domain-mobile/SKILL.md`, template pair
- **Depends on**: 없음
- **Parallel OK**: First (modules 작성 전)
- **Action**:
  1. plan.md §2.2의 frontmatter (progressive_disclosure 블록 포함)
  2. Body: Quick Reference (30초), Implementation Guide (5분), Advanced Patterns (10+분), Works Well With
  3. Skill 본문에 4 paradigm 개요 + module reference
- **Verification**: SKILL.md 존재, progressive_disclosure 블록 매치
- **Rollback**: 디렉터리 또는 파일 삭제

### T-A2-2: modules/ios-native.md
- **File owner**: `.claude/skills/moai-domain-mobile/modules/ios-native.md`, template pair
- **Depends on**: T-A2-1
- **Parallel OK**: Yes
- **Action**: SwiftUI, UIKit, Xcode, Swift Package Manager 핵심 패턴 + 1 example
- **Verification**: AC-COV001-03 (ios-native.md 존재)

### T-A2-3: modules/android-native.md
- **File owner**: 동일 패턴
- **Depends on**: T-A2-1
- **Parallel OK**: Yes
- **Action**: Jetpack Compose, Gradle, Kotlin coroutines

### T-A2-4: modules/react-native.md
- **File owner**: 동일
- **Parallel OK**: Yes
- **Action**: Expo, React Navigation, RN architecture 개요

### T-A2-5: modules/flutter.md
- **File owner**: 동일
- **Parallel OK**: Yes
- **Action**: Widget tree, Provider/Riverpod, Dart 핵심

### T-A2-6: modules/strategy-comparison.md
- **File owner**: 동일
- **Depends on**: T-A2-2 ~ A2-5 (4 paradigm 개요 작성 완료 후)
- **Parallel OK**: No
- **Action**: plan.md §2.5 비교표 + selection criteria 5개 이상
- **Verification**: AC-COV001-04 통과 (5 criteria + 4 paradigms)

### Wave A.2 Checkpoint
- moai-domain-mobile 모든 파일 (SKILL + 5 modules) 작성 완료
- AC-COV001-02 (mobile skill 부분), 03, 04 통과

---

## Wave A.3 — Framework Skills (2 tasks, parallel)

### T-A3-1: moai-framework-react-native skill
- **File owner**: `.claude/skills/moai-framework-react-native/SKILL.md` + `modules/`, template pair
- **Depends on**: 없음
- **Parallel OK**: Yes (T-A3-2와 다른 디렉터리)
- **Action**:
  1. SKILL.md (frontmatter with progressive_disclosure)
  2. modules/navigation.md (React Navigation)
  3. modules/state-management.md (Redux Toolkit, Zustand, Jotai)
  4. modules/native-modules.md (Bridge, Turbo Modules, JSI)
  5. Template + local 동시
- **Verification**: AC-COV001-02 (rn skill 부분)

### T-A3-2: moai-framework-flutter-deep skill
- **File owner**: `.claude/skills/moai-framework-flutter-deep/SKILL.md` + `modules/`, template pair
- **Depends on**: 없음
- **Parallel OK**: Yes
- **Action**:
  1. SKILL.md (frontmatter; description에 "Excludes Firebase" 명시)
  2. modules/state-management.md (Provider, Riverpod, BLoC)
  3. modules/navigation.md (go_router, Navigator 2.0)
  4. modules/networking.md (Dio)
  5. modules/platform-channels.md (iOS/Android native bridge)
  6. Template + local 동시
- **Verification**: AC-COV001-02 (flutter-deep 부분), AC-COV001-08 (기존 flutter 무수정 사전 확인)

### Wave A.3 Checkpoint
- 2개 framework skill 모두 작성 완료
- 기존 moai-framework-flutter 무수정 확인

---

## Wave A.4 — Routing 통합 (1 task)

### T-A4-1: routing keyword 추가
- **File owner**: `.claude/skills/moai/SKILL.md` 또는 routing 정의 파일 (T-A4-1 시작 전 grep으로 확인)
- **Depends on**: T-A1-1 (expert-mobile 존재 후)
- **Parallel OK**: —
- **Action**:
  1. `grep -rn "expert-backend\\|expert-frontend" .claude/skills/moai/` 로 routing 정의 위치 식별
  2. 해당 파일에 mobile 매칭 추가:
     - keyword: mobile, ios, android, swift, kotlin, react-native, flutter, swiftui, jetpack
     - dispatch: expert-mobile
  3. 우선순위 명시 (mobile 키워드 우선 매치)
  4. Template + local 동시
- **Verification**: AC-COV001-05 통과 (`grep "expert-mobile" .claude/skills/moai/` 매치)
- **Rollback**: 추가 라인 제거

### Wave A.4 Checkpoint
- AC-COV001-05 통과

---

## Wave A.5 — Verification (3 tasks)

### T-A5-1: AC verification 일괄 실행
- **Depends on**: T-A1-1, T-A2-1 ~ A2-6, T-A3-1, T-A3-2, T-A4-1 모두 완료
- **Parallel OK**: First
- **Action**: acceptance.md AC-01 ~ 09 모든 verification 스크립트 실행
- **Verification**: 모든 AC 통과 (no MISSING, no DRIFT, no FAIL)

### T-A5-2: make build + go test
- **Depends on**: T-A5-1
- **Parallel OK**: No
- **Action**:
  - `make build` (exit 0; embedded.go 갱신)
  - `go test -count=1 ./internal/template/...` (PASS)
- **Verification**: AC-COV001-07 통과

### T-A5-3: 기존 moai-framework-flutter 무수정 확인
- **Depends on**: T-A5-2
- **Parallel OK**: No
- **Action**:
  - `git diff --name-only .claude/skills/moai-framework-flutter/ internal/template/templates/.claude/skills/moai-framework-flutter/`
  - 결과 0 lines 확인 (수정 없음)
- **Verification**: AC-COV001-08 통과

### Wave A.5 Checkpoint
- AC-COV001-01 ~ 09 모두 통과
- DoD 모두 충족

---

## Edge Case Tasks (조건부)

### T-EC-1: routing 정의 파일 미식별
- **Trigger**: T-A4-1 시작 전 grep 결과 후보 없음
- **Action**: STOP, 사용자에게 routing 정의 위치 문의 (또는 manager-strategy 위임)
- **Verification**: 식별 후 진행

### T-EC-2: moai-framework-flutter scope confusion
- **Trigger**: T-A3-2 작성 시 기존 flutter skill scope와 confusing 발견
- **Action**: SKILL.md description에 명확 구분 명시 (acceptance.md EC-2)
- **Verification**: 두 skill description 비교 후 구분 명확

### T-EC-3: 4 paradigm 깊이 불균형
- **Trigger**: T-A2-2 ~ A2-5 작성 시 일부 module이 다른 module 대비 현저히 짧음
- **Action**: 모든 module 최소 baseline (Quick Reference + 1 example + Works Well With) 충족
- **Verification**: 각 module 최소 50 lines

### T-EC-4: progressive_disclosure 블록 누락
- **Trigger**: T-A2-1, T-A3-1, T-A3-2 작성 시 frontmatter에 PD 블록 누락
- **Action**: 즉시 추가 (level1_tokens: 100, level2_tokens: 5000)
- **Verification**: AC-COV001-02 PD 블록 grep 통과
