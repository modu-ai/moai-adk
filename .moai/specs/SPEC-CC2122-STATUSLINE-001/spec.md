---
id: SPEC-CC2122-STATUSLINE-001
version: "0.1.0"
status: draft
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
priority: Medium
labels: [statusline, claude-code-integration, ui, backward-compat]
issue_number: null
related_specs: [SPEC-CC2122-HOOK-001]
---

# SPEC-CC2122-STATUSLINE-001: Claude Code v2.1.122 Statusline Effort + Thinking 통합

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-30 | manager-spec | 초기 작성 — Claude Code v2.1.122 stdin JSON의 `effort` 및 `thinking` 필드를 statusline 세그먼트로 노출 |

## Status: Draft

## Overview

Claude Code v2.1.122 릴리스 노트에 따라 statusline stdin JSON에 두 가지 필드가 추가되었다:

- `effort.level` (string): `CLAUDE_CODE_EFFORT_LEVEL` 값 — `"low"`, `"medium"`, `"high"`, `"xhigh"`, `"max"` 중 하나
- `thinking.enabled` (boolean): 확장 추론(extended reasoning) 활성화 여부

본 SPEC은 두 필드를 moai-adk-go의 기존 Go statusline 패키지(`internal/statusline/`)에 통합하여 컴팩트한 인디케이터 (예: `e:high·t`)로 렌더링하는 것을 정의한다. 이전 버전의 Claude Code와의 하위 호환을 위해 두 필드가 부재하거나 `null`일 때는 세그먼트 자체를 침묵 생략(silent omit)해야 한다.

이 SPEC은 **WHAT 과 WHY**에 집중한다. 구현 함수 시그니처, 내부 호출 그래프, 정확한 segment 위치 등 **HOW**는 manager-cycle(또는 manager-tdd) 위임 시 결정된다.

## Background

기존 `internal/statusline/types.go`는 Claude Code 버전별 추가 필드를 다루는 명확한 패턴을 갖고 있다:

- `RateLimitInfo` (lines 72-83) — v2.1.80+ rate limit 필드를 nested pointer struct로 표현. 부재 시 `nil`로 처리되어 하위 호환 유지.
- `WorkspaceInfo.GitWorktree` (line 126) — v2.1.97+ 추가 필드를 빈 문자열로 그레이스 처리.

본 SPEC은 동일한 nested pointer struct 컨벤션을 따라 `*EffortInfo` 와 `*ThinkingInfo` 를 추가하고, 렌더러는 `nil` 검사로 silent omit을 보장한다.

## Requirements (EARS)

### REQ-001 (Event-Driven)

[WHEN] statusline stdin JSON 페이로드가 `effort.level` 필드를 포함하고 그 값이 비어있지 않은 문자열일 때
[THEN] 시스템은 해당 값을 추출하여 statusline 세그먼트에 `e:<level>` 형식으로 표시해야 한다 (예: `e:high`).

### REQ-002 (Event-Driven)

[WHEN] statusline stdin JSON 페이로드가 `thinking.enabled: true` 를 포함할 때
[THEN] 시스템은 thinking 인디케이터(예: 점 문자 `·t`)를 effort 표시 뒤에 부착하여 표시해야 한다 (예: `e:high·t`). `thinking.enabled` 가 `false` 이거나 부재하면 인디케이터를 표시하지 않아야 한다.

### REQ-003 (Unwanted Behavior, 하위 호환)

[IF] stdin JSON 페이로드에 `effort` 와 `thinking` 두 필드가 모두 부재하거나 `null` 인 경우
[THEN] 시스템은 효과/추론 세그먼트 전체를 침묵 생략(silent omit)해야 하며, 다른 세그먼트(model, context, git 등)에 영향을 주지 않아야 한다. 이전 버전의 Claude Code가 보낸 stdin과의 하위 호환이 보장되어야 한다.

### REQ-004 (Unwanted Behavior, 그레이스풀 폴백)

[IF] `effort.level` 값이 알려진 enum 집합(`low`/`medium`/`high`/`xhigh`/`max`)에 속하지 않는 임의 문자열인 경우
[THEN] 시스템은 panic 또는 에러 없이 raw 값을 그대로 `e:<raw>` 형식으로 표시해야 한다 (예: `effort.level="bogus"` → `e:bogus`). 이는 Claude Code 미래 버전에서 새 enum 값이 추가되어도 statusline이 깨지지 않게 한다.

### REQ-005 (State-Driven)

[WHILE] statusline이 default StatuslineMode 또는 full StatuslineMode로 렌더링되는 동안
[THEN] 신규 effort/thinking 세그먼트는 기존 3-line layout 의 다른 세그먼트(model, workspace, context window, cost, rate limits, output style 등)를 깨뜨리거나 길이를 비정상적으로 늘리지 않아야 한다. 시스템은 기존 빌더/렌더러 패턴(per-segment helper function)을 따라야 한다.

### REQ-006 (Optional / 환경 의존)

[WHERE] `.moai/config/sections/language.yaml` 의 `code_comments` 설정이 `"ko"` 인 경우
[THEN] 신규 추가되는 exported Go 타입(`EffortInfo`, `ThinkingInfo`)에 부착될 `@MX:NOTE` 태그의 설명 문자열은 한국어로 작성되어야 한다. 단, 코드 식별자 자체(타입명, 필드명, JSON 태그)는 영어로 유지한다.

> 주의: `@MX:NOTE` 태그 부착은 구현 단계(Run phase)의 책임이며, 본 SPEC 단계에서 미리 코드를 작성하거나 태그를 삽입하지 않는다. REQ-006은 manager-cycle/manager-tdd 위임 시 적용될 가이드라인이다.

## Files Affected

**수정 대상:**

- `internal/statusline/types.go` — `EffortInfo`, `ThinkingInfo` struct 추가 및 `StdinData` / `StatusData` 확장
- `internal/statusline/builder.go` — stdin → StatusData 변환 시 effort/thinking 필드 매핑
- `internal/statusline/renderer.go` — effort/thinking 세그먼트 렌더링 helper 추가, silent omit 분기 처리

**신규/확장 테스트:**

- `internal/statusline/types_test.go` (확장 또는 추가) — JSON 언마샬 하위 호환 케이스
- `internal/statusline/builder_test.go` — stdin → StatusData 매핑 테이블 드리븐 테스트
- `internal/statusline/renderer_test.go` — 렌더 출력 케이스 (6개 시나리오: present/absent/null/invalid/effort-only/thinking-only)

**간접 영향 (수정 없음, 빌드 산출물):**

- 별도 임베디드 템플릿 변경 없음 (statusline은 Go 코드 패키지)

## Exclusions (What NOT to Build)

- Effort level 값에 대한 사용자 색상 커스터마이즈 옵션은 본 SPEC에서 제외한다 (별도 SPEC으로 분리 가능).
- Thinking 인디케이터의 추가 메타데이터(예: 토큰 사용량, 추론 모드 이름)는 다루지 않는다. 본 SPEC은 boolean enabled 필드만 처리한다.
- Claude Code v2.1.122 이전 버전을 대상으로 한 polyfill 또는 fallback 데이터 소스(예: 환경변수 직접 조회)는 구현하지 않는다. stdin JSON에 필드가 없으면 그대로 silent omit 한다.
- statusline 외부 컴포넌트(예: hook 스크립트, CLI 명령어)에서 effort/thinking 정보를 노출하는 작업은 본 SPEC 범위 밖이다.
- @MX:NOTE 태그의 실제 삽입은 구현 단계의 책임이며 본 SPEC은 가이드라인만 명시한다(REQ-006 참조).

## Acceptance Reference

상세한 Given-When-Then 시나리오와 검증 항목은 `acceptance.md` 를 참조한다.

## Implementation Reference

마일스톤, 우선순위, 기술적 접근 방식은 `plan.md` 를 참조한다.
