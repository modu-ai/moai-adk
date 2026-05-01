---
id: SPEC-GLM-MCP-001
version: "0.1.0"
status: draft
created_at: 2026-05-01
updated_at: 2026-05-01
author: manager-spec (Claude via @claude trigger)
priority: Medium
labels: [mcp, glm, zai, vision, websearch, webreader, claude-code-integration, opt-in]
issue_number: 756
related_specs: [SPEC-GLM-001, SPEC-LSPMCP-001, SPEC-CC2122-MCP-001]
---

# SPEC-GLM-MCP-001: Z.AI 공식 MCP 서버 통합 (Vision + Web Search + Web Reader)

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-01 | manager-spec | 초기 작성 — `moai glm` 모드에 Z.AI 공식 MCP 서버 3종(@z_ai/mcp-server: Vision, Web Search, Web Reader) 옵트인 통합. Option 2 (Opt-in 서브커맨드 `moai glm tools enable\|disable`) 채택. 10 REQ (EARS: Ubiquitous=2, Event-Driven=4, State-Driven=1, Optional=1, Unwanted=2). |

## Status: Draft

Plan-audit verdict: **PASS** (score 0.91). Minor refinements 3건 (D1, D2, D3)은 Run phase 또는 v0.2에서 정리한다 (`plan.md` § Refinements 참조).

## Overview

`moai glm` 모드는 Claude Code 환경에서 GLM-4.6V (Z.AI) 백엔드를 사용하도록 settings 를 전환하는 모드이다. 본 SPEC 은 이 모드에 한정하여 Z.AI 가 공식 제공하는 3종 MCP 서버(`@z_ai/mcp-server` npm 패키지)를 **옵트인** 방식으로 통합한다:

- **Vision** — GLM-4.6V 이미지 인식 도구 (Pro 플랜 이상)
- **Web Search** — Z.AI 웹 검색 도구
- **Web Reader** — 웹페이지 본문 추출 도구

통합 방식은 `moai github auth glm` 패턴과 일관성을 유지하기 위해 **Option 2 (Opt-in 서브커맨드)** 를 채택한다. 사용자는 다음 명령으로 서버를 명시적으로 활성/비활성화한다:

```bash
moai glm tools enable [vision|websearch|webreader|all]
moai glm tools disable [vision|websearch|webreader|all]
moai glm tools status
```

설정 파일 스코프(R4)는 **user scope (`~/.claude.json`)** 를 권장한다. 이유: 모드 전환(`moai glm` ↔ `moai cc` ↔ `moai cg`) 시 프로젝트별 `.mcp.json` 을 매번 재기록하지 않아도 되며, GLM_AUTH_TOKEN 등 사용자 단위 자격증명과 자연스럽게 정렬된다.

본 SPEC 은 **WHAT 과 WHY** 에 집중한다. 정확한 Go 패키지 경로, 명령어 파서 구조, 테스트 함수 시그니처 등 **HOW** 는 manager-tdd (또는 manager-ddd) 위임 시 결정된다.

## Background

### Z.AI MCP 서버 패키지 (`@z_ai/mcp-server`)

Z.AI 는 자사 GLM-4.6V API 와 동반 사용 가능한 공식 MCP 서버 3종을 단일 npm 패키지(`@z_ai/mcp-server`)로 배포한다. 각 서버는 stdio transport 위에서 동작하며, 환경변수 `Z_AI_API_KEY` (또는 호환 변수) 로 인증한다.

| 서버 | 도구 (대표 예시) | 모델/티어 요구 | 사용 빈도 추정 |
|------|------------------|----------------|----------------|
| Vision | `glm_vision_describe`, `glm_vision_ocr` | GLM-4.6V Pro 플랜 이상 | 이미지 작업 시에만 |
| Web Search | `web_search`, `web_search_news` | GLM Coding Lite 이상 | 빈번 |
| Web Reader | `web_reader_fetch`, `web_reader_extract` | GLM Coding Lite 이상 | 중간 |

### 기존 GLM 자격증명 흐름 (재사용)

`moai glm` 모드는 이미 `~/.moai/.env.glm` 에 `GLM_AUTH_TOKEN` 을 기록한다 (SPEC-GLM-001). 본 SPEC 은 신규 자격증명 입력 UX 를 만들지 않고, 동일 토큰을 Z.AI MCP 서버 환경변수(`Z_AI_API_KEY`) 로 매핑하여 재사용한다 (REQ-GMC-003).

### 기존 `mcpServers` 엔트리와의 충돌 (R1)

사용자가 이미 `~/.claude.json` 또는 `.mcp.json` 의 `mcpServers` 에 `zai-mcp-server` 라는 키로 자체 엔트리를 추가했을 가능성이 있다. 본 SPEC 은 키 이름을 `glm-vision`, `glm-websearch`, `glm-webreader` (모두 `glm-` prefix) 로 표준화하여 충돌 가능성을 낮추고, 기존 엔트리 보존 정책을 명시한다 (REQ-GMC-005, REQ-GMC-008).

### 모드 전환 무영향성 (R5)

`moai glm tools enable` 로 활성화된 MCP 엔트리는 `moai cc` (Claude Anthropic API 모드) 또는 `moai cg` (CG mode) 로 전환되어도 user scope 에 그대로 남는다. Claude Code 런타임이 활성 settings 의 `enabledMcpjsonServers` 또는 환경변수에 따라 활성 여부를 분기하므로, 본 SPEC 은 settings.json 의 모드 전환 키만 동기화하면 충분하다 (REQ-GMC-009).

## Requirements (EARS)

### REQ-GMC-001 (Ubiquitous, 패키지 식별 표준화)

시스템은 Z.AI 공식 MCP 서버 3종을 다음 정규화된 키로 식별해야 한다: `glm-vision`, `glm-websearch`, `glm-webreader`. 모든 등록/해제/상태 조회 명령은 이 키 집합을 단일 진실 공급원으로 사용해야 한다. 각 키의 패키지 진입점은 `npx -y @z_ai/mcp-server vision`, `npx -y @z_ai/mcp-server websearch`, `npx -y @z_ai/mcp-server webreader` (또는 동등한 cli 진입점) 이다.

### REQ-GMC-002 (Ubiquitous, Opt-in 원칙)

시스템은 Z.AI MCP 서버를 사용자의 명시적 `moai glm tools enable` 호출 없이는 자동 등록하지 않아야 한다. `moai init`, `moai update`, `moai glm` (모드 전환), `moai cc`, `moai cg` 명령 모두 본 SPEC 의 3종 MCP 엔트리를 **추가하지 않는다**. 사용자가 명시적으로 활성화한 후에만 등록된다.

### REQ-GMC-003 (Event-Driven, 자격증명 재사용)

[WHEN] `moai glm tools enable [vision|websearch|webreader|all]` 명령이 실행되고 `~/.moai/.env.glm` 에 유효한 `GLM_AUTH_TOKEN` 이 존재할 때
[THEN] 시스템은 신규 토큰 입력을 요구하지 않고, 해당 토큰 값을 등록되는 MCP 엔트리의 환경변수 `Z_AI_API_KEY` (그리고 패키지가 요구하는 호환 변수 — D1 에서 정확한 변수명 확정) 에 매핑하여 기록해야 한다.

### REQ-GMC-004 (Event-Driven, Node.js 부재 graceful 실패)

[WHEN] `moai glm tools enable` 명령이 실행되었으나 시스템 PATH 에서 `node` 또는 `npx` 실행 파일을 찾을 수 없을 때
[THEN] 시스템은 등록을 진행하지 않고, 다음 3가지를 모두 충족하는 명확한 에러 메시지를 표준 에러 출력으로 반환해야 한다: (a) Node.js 가 설치되지 않았거나 PATH 에 없다는 사실, (b) 최소 권장 버전 (예: Node.js 18+) 안내, (c) 설치 가이드 링크 또는 `nvm install --lts` 등 즉시 실행 가능한 권장 명령. 종료 코드는 비-0 이어야 한다. 사용자의 settings 파일은 변경되지 않아야 한다.

### REQ-GMC-005 (Event-Driven, 기존 엔트리 충돌 처리)

[WHEN] `moai glm tools enable` 이 실행되어 settings 의 `mcpServers` 에 본 SPEC 의 정규 키(`glm-vision`/`glm-websearch`/`glm-webreader`) 를 추가하려 할 때, 그리고 동일 키가 이미 존재하는 경우
[THEN] 시스템은 기존 값을 자동 덮어쓰지 않고 다음 옵션을 제시해야 한다: (i) 기존 엔트리 보존 + 등록 건너뜀, (ii) 백업 후 덮어쓰기 (`<key>.bak.<timestamp>` 형태로 인접 키에 보존), (iii) 작업 취소. 비대화형 모드에서는 기본값으로 (i) 보존을 선택하고 종료 코드 비-0 + warning 메시지로 사용자에게 알린다. 사용자가 자체 추가한 비표준 키 (예: `zai-mcp-server`) 는 본 SPEC 의 정규 키 집합에 포함되지 않으므로 충돌 검사 대상이 아니며, 그대로 보존된다.

### REQ-GMC-006 (State-Driven, Pro 플랜 미가입 시 명확한 에러)

[WHILE] 사용자의 GLM 구독 티어가 Vision MCP 호출에 필요한 최소 등급(Pro 이상) 미만 상태일 때
[THE SYSTEM SHALL] `moai glm tools enable vision` 또는 `moai glm tools enable all` 명령 실행 시 다음 정보를 모두 포함하는 명확한 에러 메시지를 반환해야 한다 — (a) Vision 도구가 Pro 플랜 이상을 요구한다는 사실, (b) 사용자의 현재 티어가 검출 가능한 경우 그 값을, 검출 불가능한 경우 "verification required", (c) 업그레이드 링크 또는 안내. (b) 의 자동 검출 메커니즘 자체는 본 SPEC 에서 강제하지 않으며, 검출 불가 시 안전하게 (c) 안내로 fallback 한다. `websearch` 와 `webreader` 의 활성화는 본 요구사항의 영향을 받지 않는다.

### REQ-GMC-007 (Event-Driven, 비활성화 + 상태 조회)

[WHEN] `moai glm tools disable [vision|websearch|webreader|all]` 명령이 실행될 때
[THEN] 시스템은 지정된 키(또는 `all` 인 경우 본 SPEC 의 3종 정규 키 모두) 를 settings 의 `mcpServers` 에서 제거해야 한다. 다른 사용자 정의 mcpServers 엔트리는 보존되어야 한다. `moai glm tools status` 호출 시 시스템은 본 SPEC 의 3종 정규 키 각각에 대해 `enabled | disabled` 상태를 반환해야 하며, 사용자 정의 비표준 키는 출력에 포함하지 않는다.

### REQ-GMC-008 (Optional, settings 파일 스코프 선택)

`moai glm tools enable --scope project` 옵션이 명시적으로 지정된 경우, 시스템은 user scope (`~/.claude.json`) 대신 프로젝트 루트의 `.mcp.json` 에 엔트리를 기록할 수 있다. 옵션 부재 시 기본값은 user scope (`~/.claude.json`) 이다 (R4 결정). 두 스코프에 동일 키가 동시에 존재해도 동작은 정의되며, Claude Code 런타임의 우선순위 규칙(project > user) 을 따른다. 시스템은 두 스코프 동시 등록을 권장하지 않으며 `moai glm tools status` 출력에 경고를 표시한다.

### REQ-GMC-009 (Unwanted Behavior, 모드 전환 무영향)

[IF] 사용자가 `moai cc` (Anthropic 모드) 또는 `moai cg` (CG mode) 로 전환하는 명령을 실행할 때
[THEN] 시스템은 본 SPEC 의 3종 MCP 엔트리(`glm-vision`, `glm-websearch`, `glm-webreader`) 를 settings 의 `mcpServers` 에서 자동으로 삭제하지 않아야 한다. 모드 전환은 settings.json 의 `env`, `model`, `apiKeyHelper` 등 GLM 호환 키만 갱신하며, `mcpServers` 엔트리는 사용자 명시 호출(`moai glm tools disable`) 에 의해서만 수정된다.

### REQ-GMC-010 (Unwanted Behavior, 비대화형 환경에서의 안전성)

[IF] `moai glm tools enable` 또는 `moai glm tools disable` 명령이 비대화형 환경(CI, 스크립트, `--yes` 플래그) 에서 실행되고, 충돌 처리(REQ-GMC-005) 또는 자격증명 누락(REQ-GMC-003 의 `~/.moai/.env.glm` 부재) 이 발생한 경우
[THEN] 시스템은 사용자에게 대화형 프롬프트를 요구하지 않고, 종료 코드 비-0 + 명확한 에러 메시지로 즉시 실패해야 한다. settings 파일 부분 변경(partial write) 은 발생해서는 안 된다 — 모든 수정은 원자적으로 적용되거나 롤백된다.

## Files Affected

### 신규 작성

- `internal/cli/cmd/glm_tools.go` — `moai glm tools enable|disable|status` 서브커맨드 구현 (예시 경로; HOW 는 Run phase 결정).
- `internal/glm/mcp/registrar.go` — settings 의 `mcpServers` 키 추가/제거/조회 로직, R1 충돌 처리, R5 모드 전환 동기화.
- `internal/cli/cmd/glm_tools_test.go` — REQ-GMC-001 ~ REQ-GMC-010 매핑 단위 테스트 (≥85% 커버리지 목표).
- `internal/glm/mcp/registrar_test.go` — registrar 로직 + R4 (scope) + R1 (충돌) edge case.

### 수정 대상

- `internal/cli/cmd/glm.go` — `tools` 서브커맨드 라우팅 추가.
- `.claude/rules/moai/core/settings-management.md` — Z.AI MCP 서버 통합 노트 추가 (Pro 플랜 의존성, scope 결정, 모드 전환 동작).
- `internal/template/templates/.claude/rules/moai/core/settings-management.md` — 위 루트 사본의 미러 (Template-First Rule, CLAUDE.local.md §2 준수).
- `docs-site/content/{ko,en,ja,zh}/reference/cli/glm-tools.md` — 4개국어 reference 문서 (§17 4개국어 동기화 규칙 준수, post-merge 별도 PR 가능).

### 간접 영향 (수정 없음)

- 사용자 프로젝트의 기존 `~/.claude.json` 또는 `.mcp.json` — `moai glm tools enable` 시점에만 갱신. 자동 마이그레이션 없음.
- `~/.moai/.env.glm` — REQ-GMC-003 의 토큰 재사용 대상. 본 SPEC 은 이 파일을 읽기만 하며 수정하지 않는다.

## Exclusions (What NOT to Build)

- **EX-1: Mainland China endpoint variant (`Z_AI_MODE=ZHIPU`)** — v0.2 로 연기. 본 SPEC v0.1.0 은 글로벌 Z.AI 엔드포인트만 지원하며, 중국 본토 변형(open.bigmodel.cn 호환) 은 별도 SPEC (예: SPEC-GLM-MCP-002) 에서 다룬다.
- **EX-2: 자체 Vision MCP 서버 구현** — `@z_ai/mcp-server` 공식 패키지를 npx 로 실행하는 thin wrapper 만 제공한다. moai 가 자체 stdio 서버를 패키징하지 않는다.
- **EX-3: GLM 토큰 자동 갱신 / 만료 감지** — 토큰 재발급은 SPEC-GLM-001 의 책임 영역. 본 SPEC 은 `~/.moai/.env.glm` 의 토큰을 read-only 로 사용한다.
- **EX-4: `mcpServers` 의 비표준 사용자 키 자동 정리** — 사용자가 임의로 추가한 `zai-mcp-server` 등의 키는 본 SPEC 의 충돌 검사/제거 대상이 아니다 (REQ-GMC-005 후반부).
- **EX-5: GLM 모델 자동 다운그레이드 (Vision Pro 미가입 시 텍스트 전용 fallback)** — Pro 미가입 시 명확한 에러 (REQ-GMC-006) 만 제공한다. 자동 fallback 모델 선택은 본 SPEC 범위 밖.
- **EX-6: 샘플 통합 테스트 (실 API 키 필요)** — v0.2 로 연기. 본 SPEC v0.1.0 의 테스트는 모킹 기반(D3 에서 정의되는 mechanism) 단위 테스트로 한정한다. 실 API 호출 통합 테스트는 별도 환경(Z_AI_TEST_API_KEY 환경변수 보호) 에서 향후 추가한다.
- **EX-7: `moai glm tools` 외 진입점** — 본 SPEC 은 `moai glm tools` 서브커맨드만 정의한다. 슬래시 커맨드(`/moai glm-tools`), GUI, 자동 탐지(이미지 첨부 시 자동 vision enable) 등은 범위 밖.

## Acceptance Reference

상세한 Given-When-Then 시나리오 22건은 `acceptance.md` 를 참조한다. 모든 10개 REQ 를 커버하며, edge cases (Node.js 부재, GLM_AUTH_TOKEN 부재, 기존 충돌, Pro 플랜 미가입, 모드 전환, 비대화형) 를 포함한다.

## Implementation Reference

마일스톤(M1~M3), 우선순위, 기술적 접근 방식, 리스크 R1~R5 의 상세 mitigations, 그리고 plan-audit 가 지적한 minor refinements (D1, D2, D3) 의 처리 방안은 `plan.md` 를 참조한다.

## Performance and Quality Gates

- **Command latency**: `moai glm tools enable`, `moai glm tools disable`, `moai glm tools status` 각각 5초 이내 응답 (settings.json read + write + validate 포함; npx 첫 실행으로 인한 패키지 다운로드 시간은 제외; D2 에서 측정 mechanism 확정).
- **Test coverage**: 본 SPEC 이 신규 추가하는 패키지(`internal/glm/mcp`, `internal/cli/cmd/glm_tools.go`) 에 대해 line coverage ≥85% (CLAUDE.local.md §6 의 critical packages 기준 90%+ 미만이지만 본 SPEC 의 신규 패키지는 'critical' 미분류로 간주).
- **TRUST 5**: 모든 신규 코드는 .claude/rules/moai/core/moai-constitution.md 의 TRUST 5 게이트(Test, Readability, Understandability, Safety, Tracing) 를 통과해야 한다.
- **No partial write (REQ-GMC-010)**: settings.json 수정은 항상 원자적이어야 하며, 실패 시 롤백된다. 부분 적용된 settings 가 디스크에 남아서는 안 된다.

## Backward Compatibility

본 SPEC 은 신규 서브커맨드를 추가하는 것이며, 기존 `moai glm` 동작에는 영향이 없다. 기존 `moai glm` 사용자가 `tools` 서브커맨드를 호출하지 않는 한 settings 는 변경되지 않는다 (REQ-GMC-002 의 Opt-in 원칙).

기존 사용자가 임의로 추가한 `zai-mcp-server` 등 비표준 키는 본 SPEC 의 정규 키 집합과 분리되어 보존된다 (REQ-GMC-005 후반부, EX-4).

---

**Plan-audit reference**: `.moai/reports/plan-audit/SPEC-GLM-MCP-001-review-1.md` (verdict: PASS, score 0.91).
