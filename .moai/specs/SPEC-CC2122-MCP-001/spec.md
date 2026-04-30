---
id: SPEC-CC2122-MCP-001
version: "0.1.0"
status: draft
created_at: 2026-04-30
updated_at: 2026-04-30
author: manager-spec
priority: Medium
labels: [mcp, claude-code-integration, templates, backward-compat]
issue_number: null
related_specs: [SPEC-CC2122-HOOK-001, SPEC-CC2122-STATUSLINE-001]
---

# SPEC-CC2122-MCP-001: Claude Code v2.1.119 MCP `alwaysLoad` 통합

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-04-30 | manager-spec | 초기 작성 — Claude Code v2.1.119 `.mcp.json` `alwaysLoad: true` 필드를 `context7` 와 `sequential-thinking` 서버 엔트리에 적용. `moai-lsp` 는 lazy-load 유지. |

## Status: Draft

## Overview

Claude Code v2.1.119 릴리스 노트에 따라 `.mcp.json` 의 `mcpServers` 엔트리에 선택적 `alwaysLoad: true` 필드가 추가되었다. 이 필드가 `true` 이면 해당 MCP 서버의 도구 스키마(tool schemas)가 세션 시작 시점에 컨텍스트로 미리 로드(pre-load)되며, `false` 이거나 부재하면 기존처럼 첫 도구 호출 시점에 지연 로드(lazy-load)된다.

본 SPEC 은 moai-adk-go 의 `internal/template/templates/.mcp.json.tmpl` 템플릿을 v2.1.119 호환으로 업데이트하여, **소형 스키마 + 빈번한 사용** 특성을 가진 `context7` (라이브러리 문서 조회) 와 `sequential-thinking` (단계별 추론) 두 서버에 한해 `alwaysLoad: true` 를 부여한다. **대형 스키마 + 비상시 사용** 특성을 가진 `moai-lsp` (LSP code intelligence) 는 lazy-load 를 유지하여 초기 컨텍스트 비용을 절감한다.

이 SPEC 은 **WHAT 과 WHY** 에 집중한다. 정확한 JSON 키 위치, Go 테스트 함수 시그니처, 문서 주석의 정확한 단어 선택 등 **HOW** 는 manager-cycle (또는 manager-tdd) 위임 시 결정된다.

## Background

`.mcp.json.tmpl` 의 현재(템플릿 기준) 상태는 3개의 mcpServers 엔트리를 포함한다:

- `context7` — `npx -y @upstash/context7-mcp@latest` 실행. 라이브러리 문서 조회용. 스키마 소형, 빈번 사용.
- `sequential-thinking` — `npx -y @modelcontextprotocol/server-sequential-thinking` 실행. 구조화된 단계별 추론용. 스키마 소형, 빈번 사용.
- `moai-lsp` — `moai mcp lsp` 실행. LSP 기반 코드 인텔리전스(goto definition, find references, hover, diagnostics, rename). 스키마 대형, 일부 세션에서만 사용.

세 엔트리 중 어느 것도 현재 `alwaysLoad` 필드를 갖지 않는다. v2.1.119 부터는 선택적 필드이므로 부재 시 기본 동작(lazy-load)으로 처리되어 하위 호환은 유지되지만, 자주 사용되는 두 서버에 한해 명시적으로 `alwaysLoad: true` 를 부여하면 첫 도구 호출 지연(latency)이 제거된다는 이점이 있다.

`moai-lsp` 에 대해 동일한 처리를 하지 않는 이유: LSP 도구 스키마는 다른 두 서버 대비 상대적으로 크고, 모든 세션에서 사용되지 않기 때문에 항상 로드 시 초기 컨텍스트가 불필요하게 증가한다. 따라서 lazy-load 가 적합하다.

문서 동기화 측면에서는 `.claude/rules/moai/core/settings-management.md` 와 그 템플릿 미러본인 `internal/template/templates/.claude/rules/moai/core/settings-management.md` 두 파일에 v2.1.119 `alwaysLoad` 의미와 사용 가이드를 짧은 노트로 추가해야 한다(REQ-004).

## Requirements (EARS)

### REQ-001 (Event-Driven)

[WHEN] `moai init` 또는 `moai update` 명령이 `internal/template/templates/.mcp.json.tmpl` 을 렌더링하여 사용자 프로젝트 루트의 `.mcp.json` 파일을 생성/갱신할 때
[THEN] 결과 `.mcp.json` 의 `mcpServers.context7` 엔트리와 `mcpServers["sequential-thinking"]` 엔트리는 각각 `"alwaysLoad": true` 키-값 쌍을 포함해야 한다.

### REQ-002 (Unwanted Behavior)

[IF] 동일한 템플릿 렌더링 결과의 `mcpServers["moai-lsp"]` 엔트리에 `alwaysLoad` 필드를 추가하려는 시도가 있는 경우
[THEN] 시스템은 해당 추가를 거부해야 한다. 즉, `moai-lsp` 엔트리는 `alwaysLoad` 키를 포함하지 않아야 하며, 만약 어떤 이유로 키가 존재한다면 그 값은 반드시 `false` 여야 한다 (lazy-load 유지).

### REQ-003 (Ubiquitous, 회귀 방지)

시스템은 본 SPEC 의 변경이 적용된 후에도 `.mcp.json.tmpl` 의 모든 3개 mcpServers 엔트리(`context7`, `sequential-thinking`, `moai-lsp`)에서 기존에 존재하던 필드 — `command`, `args`, `timeout`, `$comment`, 그리고 플랫폼 분기(`{{- if eq .Platform "windows"}}` 블록의 윈도우/유닉스 명령 분기) — 를 변경 없이 그대로 보존해야 한다. `alwaysLoad` 추가 외의 어떤 필드 변경도 발생해서는 안 된다.

### REQ-004 (Event-Driven, 문서 동기화)

[WHEN] 프로젝트 문서가 Claude Code v2.1.119 호환성을 설명할 때
[THEN] `.claude/rules/moai/core/settings-management.md` (루트 사본) 와 `internal/template/templates/.claude/rules/moai/core/settings-management.md` (템플릿 미러본) 두 파일은 모두 `alwaysLoad: true` 의 의미를 설명하는 짧은 노트를 포함해야 한다. 노트는 최소한 다음 세 요점을 다뤄야 한다: (a) `alwaysLoad: true` 가 세션 시작 시 도구 스키마를 미리 로드하는 동작임, (b) 부재/`false` 시 기존 lazy-load 동작 유지로 하위 호환됨, (c) 사용 권장 기준 — 소형 스키마 + 빈번한 사용 시 `true` 를, 대형 스키마 또는 비상시 사용 시 `false` (또는 부재) 를 권장.

### REQ-005 (Event-Driven, 테스트 커버리지)

[WHEN] `internal/template/settings_test.go` 에 정의된 테스트 스위트를 `go test ./internal/template/...` 로 실행할 때
[THEN] 다음 3가지 동작을 각각 명시적으로 검증하는 신규 테스트 케이스가 최소 3개 이상 존재해야 한다: (a) REQ-001 검증 — 렌더링된 `.mcp.json` 의 `context7` 와 `sequential-thinking` 엔트리에 `alwaysLoad: true` 가 존재함, (b) REQ-002 검증 — `moai-lsp` 엔트리에 `alwaysLoad` 가 부재하거나 `false` 임, (c) REQ-003 검증 — 기존 필드(`command`, `args`, `timeout`, `$comment`)가 모든 3개 엔트리에서 회귀 없이 보존됨. 신규 테스트 함수명은 `TestMCPTemplateAlwaysLoad*` (또는 동등한 검색 가능한) prefix 를 사용한다.

## Files Affected

**수정 대상:**

- `internal/template/templates/.mcp.json.tmpl` — `context7` 와 `sequential-thinking` 두 엔트리에 `"alwaysLoad": true` 추가. `moai-lsp` 엔트리는 변경하지 않음.
- `internal/template/settings_test.go` — `TestMCPTemplateAlwaysLoad*` prefix 의 신규 테스트 케이스 최소 3개 추가 (REQ-001/002/003 매핑).
- `.claude/rules/moai/core/settings-management.md` — v2.1.119 `alwaysLoad` 의미 + 사용 권장 기준을 짧은 섹션 또는 노트 블록으로 추가.
- `internal/template/templates/.claude/rules/moai/core/settings-management.md` — 위 루트 사본과 동일한 노트를 미러링 (Template-First Rule, CLAUDE.local.md §2 준수).

**간접 영향 (수정 없음):**

- `internal/template/embedded.go` — 자동 생성 파일. `make build` 로 재생성됨. 직접 편집 금지.
- 사용자 프로젝트의 기존 `.mcp.json` — `moai update` 시점에 갱신될 수 있음. 본 SPEC 의 행동은 템플릿 변경에 한정되며, 기존 사용자 파일에 대한 마이그레이션 정책은 본 SPEC 범위 밖이다.

## Exclusions (What NOT to Build)

- `moai-lsp` 엔트리에 `alwaysLoad: true` 부여는 본 SPEC 에서 명시적으로 제외한다. 향후 LSP 사용 패턴이 변하거나 스키마가 축소되면 별도 SPEC 으로 검토할 수 있다.
- `staggeredStartup` 블록 (`enabled`, `delayMs`, `connectionTimeout`) 의 변경은 본 SPEC 범위 밖이다. 본 SPEC 은 `mcpServers.*` 내부의 `alwaysLoad` 필드만 다룬다.
- 사용자 정의 MCP 서버 (사용자가 자신의 `.mcp.json` 에 추가한 추가 엔트리) 에 대한 `alwaysLoad` 자동 부여 로직은 구현하지 않는다. 본 SPEC 은 템플릿이 제공하는 3개 표준 엔트리에만 적용된다.
- `.mcp.json` 스키마 검증 강화 (예: `alwaysLoad` 가 boolean 인지 type-check) 는 본 SPEC 범위 밖이다. Claude Code 런타임이 자체 검증을 수행한다고 가정한다.
- 기존 사용자의 `.mcp.json` 에 자동으로 `alwaysLoad: true` 를 주입하는 마이그레이션 도구는 구현하지 않는다. `moai update` 의 기존 머지 정책을 따른다.
- `.claude/settings.json` 의 MCP 관련 키 (예: `enabledMcpjsonServers`) 변경은 본 SPEC 의 대상이 아니다. 본 SPEC 은 `.mcp.json.tmpl` 템플릿과 그 동작을 설명하는 문서에만 영향을 미친다.

## Acceptance Reference

상세한 Given-When-Then 시나리오와 검증 항목은 `acceptance.md` 를 참조한다.

## Implementation Reference

마일스톤, 우선순위, 기술적 접근 방식은 `plan.md` 를 참조한다.
