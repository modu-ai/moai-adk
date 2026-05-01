# SPEC-CC2122-HOOK-002: HOOK-001 Follow-up (REQ-007 + REQ-008)

## Status: DRAFT

## Overview

SPEC-CC2122-HOOK-001 의 follow-up. 다음 두 요구사항을 완성한다:

1. **REQ-007** (yaml config read): `slow_hook_threshold_ms` 임계값을 hardcoded 5000ms 가 아닌 `.moai/config/sections/observability.yaml` 의 `observability.hook_metrics.slow_hook_threshold_ms` 에서 읽도록 한다.
2. **REQ-008** (CG Mode disallowedTools regression guard): 에이전트 frontmatter `disallowedTools:` 가 현실적으로 유효한 형식을 갖추고 있는지 정적으로 검증한다. 실제 `claude --print` 실행은 dev 프로젝트의 GLM 통합 테스트 금지 정책(CLAUDE.local.md §13)에 따라 별도 manual 검증으로 분리한다.

## Requirements

### REQ-CC2122-HOOK-002-001 (REQ-007 implementation)

[WHEN] `writeHookMetric` 가 호출되고 `.moai/config/sections/observability.yaml` 가 존재하며 `observability.hook_metrics.slow_hook_threshold_ms` 가 양의 정수로 정의되어 있을 때
[THEN] 시스템은 해당 값을 임계값으로 사용해야 한다 (hardcoded 5000ms 보다 우선)

### REQ-CC2122-HOOK-002-002

[WHEN] `.moai/config/sections/observability.yaml` 가 부재하거나 yaml 파싱이 실패하거나 키가 누락된 경우
[THEN] 시스템은 hardcoded 기본값 5000ms 로 fallback 해야 하며 panic 또는 stderr 출력 없이 정상 동작해야 한다

### REQ-CC2122-HOOK-002-003

[WHEN] yaml 의 `slow_hook_threshold_ms` 값이 0 이하 또는 비-숫자형인 경우
[THEN] 시스템은 hardcoded 기본값 5000ms 로 fallback 해야 한다 (의도적 invalid value 보호)

### REQ-CC2122-HOOK-002-004 (REQ-008 static linter)

[WHEN] `internal/template/templates/.claude/agents/**/*.md` 의 frontmatter 가 파싱될 때
[THEN] 모든 에이전트 정의는 다음 두 조건 중 하나를 충족해야 한다:
- `tools:` 만 정의 (allowlist 모드)
- `disallowedTools:` 만 정의 (denylist 모드)
- 동시 정의 금지 (mutual exclusion, claude-code v2.1.119+ 사양)

### REQ-CC2122-HOOK-002-005

[WHEN] `disallowedTools:` 가 정의되어 있을 때
[THEN] 값은 CSV 문자열 형식(공백 구분 금지) 이어야 한다 (CLAUDE.local.md §12 규칙)

### REQ-CC2122-HOOK-002-006

[WHEN] CI 또는 로컬에서 `go test ./internal/...` 가 실행될 때
[THEN] 위 정적 검증은 단위 테스트로 자동 실행되어야 한다 (수동 step 불요)

## Out of Scope

- **실제 `claude --print` 동시 실행 회귀 테스트**: dev 프로젝트의 GLM 통합 테스트 금지 정책(§13). 별도 manual smoke test 스크립트(`scripts/manual-cg-disallowed-test.sh`)로 분리하거나 CI 별도 job 으로 향후 처리.
- **observability.yaml 의 다른 필드(retention_days, max_file_size_mb 등) yaml read**: 현재 hook 메트릭 코드 외 사용처가 없으므로 불필요. SPEC-OBS 별 SPEC 으로 분리 가능.
- **runtime config hot-reload**: 매 hook 호출마다 yaml 재읽기 — 단순화를 위해 채택 (sync.Once 캐싱 없음). 성능 영향 없음(파일 크기 < 1KB).

## Acceptance Criteria

- AC-001: `internal/hook/post_tool_duration.go` 가 `loadSlowHookThreshold(projectRoot string) int64` helper 를 호출하여 yaml 우선, fallback default 동작
- AC-002: `loadSlowHookThreshold` 단위 테스트 4건 (file present + valid / file absent / file malformed / invalid value)
- AC-003: 기존 `post_tool_duration_test.go` 의 4건 회귀 테스트 모두 PASS (default 5000ms 동작 보존)
- AC-004: `internal/template/templates/.claude/agents/` validator 단위 테스트 추가, 모든 agent .md 파일 frontmatter 검증
- AC-005: `go test ./internal/hook/... ./internal/template/...` 단일 명령으로 새 테스트 모두 실행

## Files Affected

**수정:**
- `internal/hook/post_tool_duration.go` (loadSlowHookThreshold helper 추가, writeHookMetric 통합)

**신규:**
- `internal/hook/post_tool_duration_threshold_test.go` (yaml read 단위 테스트)
- `internal/template/agents_frontmatter_test.go` (REQ-008 정적 linter, 새 테스트 파일)
- `scripts/manual-cg-disallowed-test.sh` (선택 — manual smoke test 스크립트)

**무수정 (검증):**
- `internal/template/templates/.moai/config/sections/observability.yaml` (이미 정의됨)
- `internal/hook/types.go` (HookInput 변경 없음)

## Methodology

TDD (RED-GREEN-REFACTOR) — REQ-007 yaml read 는 테스트 우선 작성. REQ-008 static linter 는 단순한 표 기반 검증이라 테스트가 곧 명세.

## Risks & Mitigations

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| yaml parse 실패가 panic 으로 전파 | Low | recover() 또는 명시적 error swallow + default fallback |
| agent frontmatter 검증이 기존 agent 파일에서 fail | Medium | 검증 추가 전 1회 audit 후 위반 케이스 사전 수정 |
| Manual smoke test 가 실제로 수행되지 않음 | Medium | 스크립트 + README 가이드만 제공, CI 강제 안 함 (follow-up) |
