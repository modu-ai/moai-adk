# SPEC-CC2122-HOOK-001: Claude Code v2.1.119-121 Hook Feature Integration

## Status: COMPLETED

## Overview

Claude Code v2.1.119-121에서 추가된 3가지 기능을 moai-adk-go 템플릿에 통합한다:

1. **v2.1.119**: PostToolUse / PostToolUseFailure 훅 stdin JSON에 `duration_ms` 필드 추가
2. **v2.1.121**: PostToolUse `hookSpecificOutput.updatedToolOutput`이 MCP 전용에서 모든 도구로 확장
3. **v2.1.119**: `claude --print` 모드가 에이전트 `tools:` / `disallowedTools:` frontmatter를 준수 (CG Mode 회귀 리스크)

## Requirements

### REQ-CC2122-HOOK-001-001
[WHEN] PostToolUse 또는 PostToolUseFailure 훅이 실행되고 stdin JSON에 `duration_ms` 필드가 존재할 때
[THEN] 시스템은 `duration_ms` 값을 추출하여 slow_hook_threshold_ms와 비교해야 한다

### REQ-CC2122-HOOK-001-002
[WHEN] `duration_ms > slow_hook_threshold_ms` 이고 `.moai/observability/` 디렉토리가 존재할 때
[THEN] 시스템은 `.moai/observability/hook-metrics.jsonl`에 원자적으로 1줄의 JSON을 추가해야 한다
- 필드: ts (ISO 8601 UTC), hook, tool, duration_ms, session_id

### REQ-CC2122-HOOK-001-003
[WHEN] `.moai/observability/` 디렉토리가 존재하지 않을 때
[THEN] 훅은 메트릭 쓰기를 조용히 건너뛰고 정상 종료해야 한다 (exit 0)

### REQ-CC2122-HOOK-001-004
[WHEN] `duration_ms <= slow_hook_threshold_ms` 일 때
[THEN] 훅은 메트릭을 쓰지 않고 정상 종료해야 한다

### REQ-CC2122-HOOK-001-005
[WHEN] `MOAI_HOOK_OUTPUT_TRANSFORM=1` 환경변수가 설정되어 있을 때
[THEN] PostToolUse 훅은 `hookSpecificOutput.updatedToolOutput` JSON을 stdout에 출력할 수 있는 scaffold를 활성화해야 한다
- 기본값 (env 미설정): 아무것도 출력하지 않음 (현재 동작 보존)

### REQ-CC2122-HOOK-001-006
[WHEN] PostToolUseFailure 훅이 실행되고 `duration_ms > threshold` 일 때
[THEN] 메트릭 로그 항목에 `"outcome": "failure"` 필드가 포함되어야 한다

### REQ-CC2122-HOOK-001-007
[WHEN] `.moai/config/sections/observability.yaml`에 `hook_metrics.slow_hook_threshold_ms`가 정의되어 있을 때
[THEN] 해당 값이 hardcoded 기본값(5000ms)보다 우선적용되어야 한다

### REQ-CC2122-HOOK-001-008
[WHEN] `claude --print` 모드로 에이전트를 실행할 때
[THEN] 에이전트의 `disallowedTools:` frontmatter에 지정된 도구는 실행이 거부되어야 한다

## Files Affected

**수정:**
- `internal/template/templates/.claude/hooks/moai/handle-post-tool.sh.tmpl`
- `internal/template/templates/.claude/hooks/moai/handle-post-tool-failure.sh.tmpl`
- `internal/template/templates/.moai/config/sections/observability.yaml`
- `.claude/rules/moai/core/settings-management.md`
- `internal/template/templates/.claude/rules/moai/core/settings-management.md`

**생성:**
- `internal/hook/post_tool_metrics_test.go` (새 테스트)
- `internal/cli/cg_test.go` (새 테스트)

**간접 영향:**
- `internal/template/embedded.go` (make build로 재생성)
