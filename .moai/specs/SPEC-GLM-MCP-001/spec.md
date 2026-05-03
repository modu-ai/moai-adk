---
id: SPEC-GLM-MCP-001
version: "0.1.0"
status: draft
created_at: 2026-05-01
updated_at: 2026-05-01
author: manager-spec
priority: Medium
labels: [glm, mcp, vision, websearch, integration, enhancement]
issue_number: 756
related_specs: [SPEC-GLM-001, SPEC-CC2122-MCP-001, SPEC-LSPMCP-001]
---

# SPEC-GLM-MCP-001: Z.AI 공식 MCP 서버 통합 (Vision + Web Search + Web Reader)

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-01 | manager-spec | 초기 작성 — Z.AI `@z_ai/mcp-server` 통합. Opt-in subcommand (`moai glm tools enable/disable`) 권장. Mainland China (Z_AI_MODE=ZHIPU) 는 v0.2 로 연기. |

## Status: Draft

## Metadata

| Field | Value |
|-------|-------|
| SPEC ID | SPEC-GLM-MCP-001 |
| Domain | GLM-MCP |
| Number | 001 |
| Lifecycle Level | spec-anchored |
| Recommended Integration Option | Option 2 (Opt-in subcommand) |
| Token Reuse Source | `~/.moai/.env.glm` `GLM_AUTH_TOKEN` |
| Node.js Prerequisite | `>= v22.0.0` |
| Z.AI Package Min Version | `@z_ai/mcp-server >= v0.1.2` |
| Z.AI Plan Tier (Vision/WebSearch) | Pro ($9/월) 이상 |
| Default MCP Scope | user (`~/.claude.json`) |
| Z_AI_MODE Default | `ZAI` (국제) |

## Overview

Z.AI 의 공식 MCP 서버 패키지(`@z_ai/mcp-server`) 는 GLM-4.6V Vision (이미지 인식, OCR), Web Search (최신 기술 문서 검색), Web Reader (웹페이지 본문/메타데이터 추출) 세 도구를 Claude Code 내부에서 노출한다. 본 SPEC 은 `moai glm` 모드를 사용하는 사용자가 위 도구들을 *명시적이고 가역적으로* 활성화할 수 있도록 신규 subcommand `moai glm tools enable|disable [vision|websearch|webreader|all]` 를 도입한다.

본 SPEC 은 SPEC-GLM-001 (호환성 자동화: `DISABLE_BETAS=1` + `DISABLE_PROMPT_CACHING=1`) 위에 *MCP 서버 등록* 레이어를 추가한다. 캐싱/베타 정책은 변경하지 않으며, 모드 전환 (`moai cc` ↔ `moai glm` ↔ `moai cg`) 의 기존 의미를 보존한다.

본 SPEC 은 **WHAT 과 WHY** 에 집중한다. 정확한 Go 함수 시그니처, 패키지 구조, 명령 라우팅 코드는 manager-cycle 위임 시 결정한다.

## Background

연구 산출물 `research.md` 에 정리된 사실 카탈로그를 요약하면:
- Z.AI 가 공식 npm 패키지 `@z_ai/mcp-server` (`>= v0.1.2`) 를 통해 3개 도구를 MCP 인터페이스로 노출
- 공식 등록 명령: `claude mcp add -s user zai-mcp-server --env Z_AI_API_KEY=<key> Z_AI_MODE=ZAI -- npx -y "@z_ai/mcp-server@latest"`
- moai-adk-go 의 기존 패턴 `moai github auth glm <token>` 와 동일한 명시적 사용자 의도 모델이 안전성 + 가역성 + 사용자 신뢰 측면에서 우월함

3가지 통합 옵션 (Auto-inject / Opt-in subcommand / Documentation guide) 중 **Opt-in subcommand** 가 권장됨 (research.md §3.4 참조). Auto-inject 는 사용자 동의 없이 침습적이고, Documentation guide 는 마찰이 크고 token 재사용을 활용하지 못한다.

## 3.0 EARS 분배표

| 패턴 | 갯수 | REQ ID |
|------|------|--------|
| Ubiquitous | 2 | REQ-GMC-001, REQ-GMC-002 |
| Event-Driven | 4 | REQ-GMC-003, REQ-GMC-004, REQ-GMC-005, REQ-GMC-006 |
| State-Driven | 1 | REQ-GMC-007 |
| Optional | 1 | REQ-GMC-008 |
| Unwanted | 2 | REQ-GMC-009, REQ-GMC-010 |
| **합계** | **10** | — |

## Requirements (EARS Format)

### REQ-GMC-001 (Ubiquitous)

시스템은 신규 subcommand `moai glm tools` 를 제공해야 한다. 이 subcommand 는 최소 두 가지 동작 모드를 가진다: `enable [vision|websearch|webreader|all]` 와 `disable [vision|websearch|webreader|all]`. 두 동작 모드 모두 idempotent 해야 한다 — 동일 인자로 반복 호출해도 결과가 동일하고 부수 효과가 누적되지 않아야 한다.

### REQ-GMC-002 (Ubiquitous)

시스템은 본 SPEC 에 의해 도입된 어떤 동작도 SPEC-GLM-001 의 호환성 자동화 (`CLAUDE_CODE_DISABLE_EXPERIMENTAL_BETAS=1` + `DISABLE_PROMPT_CACHING=1` 의 mode 별 inject/strip) 의미와 충돌하지 않도록 보장해야 한다. 즉, MCP 등록은 환경변수 정책과 직교적이며, `moai cc/glm/cg` 모드 전환은 기존 정책을 그대로 유지한다.

### REQ-GMC-003 (Event-Driven)

[WHEN] 사용자가 `moai glm tools enable vision` 또는 `moai glm tools enable all` 을 실행하고
[AND] 시스템 PATH 에 Node.js `>= v22.0.0` 가 존재하고
[AND] `~/.moai/.env.glm` 에 유효한 `GLM_AUTH_TOKEN` 이 존재할 때
[THEN] 시스템은 `~/.claude.json` (user scope) 의 `mcpServers` 객체에 `zai-mcp-server` 엔트리를 추가해야 하며, 그 엔트리는 다음을 포함해야 한다:
- `command: "npx"`
- `args: ["-y", "@z_ai/mcp-server@latest"]`
- `env.Z_AI_API_KEY`: `~/.moai/.env.glm` 의 `GLM_AUTH_TOKEN` 값
- `env.Z_AI_MODE`: `ZAI` (기본값)

[AND] 시스템은 사용자에게 활성화된 도구 목록(Vision / Web Search / Web Reader) + 사용 안내(Pro 플랜 권장 + Claude Code 재시작 필요) 를 출력해야 한다.

### REQ-GMC-004 (Event-Driven)

[WHEN] 사용자가 `moai glm tools disable [도구명|all]` 을 실행할 때
[THEN] 시스템은 `~/.claude.json` 의 `mcpServers.zai-mcp-server` 엔트리를 제거해야 한다 (또는 도구별 부분 disable 모델을 도입하는 경우 해당 도구만 비활성화). 제거 후 `mcpServers` 객체가 비면 빈 객체로 유지하고 다른 사용자 정의 MCP 엔트리는 절대 손대지 않는다.

[AND] 시스템은 제거된 도구 목록을 사용자에게 출력해야 한다.

### REQ-GMC-005 (Event-Driven)

[WHEN] 사용자가 `moai glm tools enable` 또는 `moai glm tools disable` 을 실행할 때
[THEN] 시스템은 `~/.claude.json` 파일을 수정하기 전에 안전 백업을 동일 디렉토리에 `~/.claude.json.bak-<ISO timestamp>` 로 저장해야 한다. 단, 동일 명령의 idempotent skip (변경 없음) 의 경우에는 백업을 생성하지 않아도 된다.

### REQ-GMC-006 (Event-Driven)

[WHEN] 사용자가 `moai glm tools enable` 을 실행하고
[AND] `~/.claude.json` 의 `mcpServers.zai-mcp-server` 엔트리가 이미 존재할 때
[THEN] 시스템은 다음을 검사해야 한다:
- (a) 기존 엔트리의 `env.Z_AI_API_KEY` 가 현재 `~/.moai/.env.glm` 의 `GLM_AUTH_TOKEN` 과 동일하면 idempotent skip 처리하고 "이미 활성화됨" 메시지를 출력한다.
- (b) 토큰이 다르거나 다른 필드가 사용자 정의 변경된 경우, 시스템은 *덮어쓰지 말고* `--force` 플래그 안내 + 차이 요약을 출력한 후 종료해야 한다 (R1 위험 완화).

### REQ-GMC-007 (State-Driven)

[WHILE] `~/.moai/.env.glm` 의 `GLM_AUTH_TOKEN` 이 존재하지 않거나 빈 문자열인 동안
[THEN] 시스템은 `moai glm tools enable` 명령을 거부해야 하며, 사용자에게 `moai github auth glm <token>` 또는 동등한 토큰 등록 명령으로 먼저 토큰을 설정하도록 안내 메시지를 출력해야 한다 (한 번에 한 가지 액션만 요구).

### REQ-GMC-008 (Optional)

[WHERE] 사용자가 `moai glm tools enable --scope project` 옵션을 명시한 경우
[THEN] 시스템은 user scope 인 `~/.claude.json` 대신 프로젝트 루트의 `.mcp.json` 에 `zai-mcp-server` 엔트리를 추가해야 한다. 이 옵션이 부재하면 기본은 user scope 이다 (Z.AI 공식 가이드와 일관).

### REQ-GMC-009 (Unwanted Behavior)

[IF] `node --version` 실행이 실패하거나 결과가 `v22.0.0` 미만인 경우
[THEN] 시스템은 `moai glm tools enable` 을 즉시 중단하고 다음을 출력해야 한다:
- 감지된 Node.js 버전 (또는 "not found")
- 최소 요구 버전 (`>= v22.0.0`)
- 설치 가이드 링크 또는 권장 설치 명령

[AND] 종료 코드는 비-0 이어야 하며, `~/.claude.json` 은 *어떤 변경도 없이* 유지되어야 한다.

### REQ-GMC-010 (Unwanted Behavior)

[IF] `~/.claude.json` 에 본 SPEC 이 도입한 `zai-mcp-server` 엔트리 외의 사용자 정의 `mcpServers` 엔트리가 존재하는 경우
[THEN] 시스템은 `enable` / `disable` 동작 중 어떤 경우에도 사용자 정의 엔트리를 변경/이동/삭제해서는 안 된다. 본 SPEC 의 변경 범위는 `mcpServers.zai-mcp-server` 키 하나에 한정된다.

## Specifications (HOW — 간략)

본 섹션은 *구현 단계로 위임될 결정* 의 윤곽만 제공한다. 정확한 함수명, 패키지 경로, JSON 직렬화 방식은 plan.md 와 manager-cycle 위임 시점에 확정된다.

- 신규 명령 라우팅: `internal/cli/glm.go` (또는 분리된 `glm_tools.go`) 에 `glmToolsCommand` 추가
- JSON 머지 로직: 기존 `~/.claude.json` 을 안전하게 읽고, 사용자 정의 엔트리 보존하면서 `zai-mcp-server` 만 갱신
- 토큰 조회: 기존 `loadGLMKey()` 헬퍼 재사용 (research.md §2.2)
- Node.js 버전 검증: `exec.LookPath("node")` + `node --version` 출력 파싱
- 출력 메시지: conversation_language 정책 (영어 에러, 사용자 메시지 ko)
- 패키지 위치 후보: `internal/cli/glm_tools.go` 또는 `internal/integration/zaimcp/`

## Files Affected (Anticipated, Plan-Stage Only)

본 SPEC 은 코드를 수정하지 않으나, manager-cycle 위임 시 다음 파일이 영향을 받을 것으로 예측된다 (plan.md 에서 정확한 라인 범위 결정):

**예상 신규 / 수정:**

- `internal/cli/glm.go` 또는 `internal/cli/glm_tools.go` — `tools enable|disable` 명령 추가
- `internal/cli/glm_test.go` 또는 `internal/cli/glm_tools_test.go` — 신규 테스트 케이스
- `internal/template/templates/.claude/rules/moai/core/settings-management.md` — Z.AI MCP 통합 노트 추가 (Template-First Rule, CLAUDE.local.md §2)
- `.claude/rules/moai/core/settings-management.md` — 위 미러본
- `CHANGELOG.md` — v0.1 항목 추가

**간접 영향 (수정 없음, 정책 검증):**

- `internal/cli/launcher.go` — mode 전환 로직, 본 SPEC 변경 없음
- `internal/hook/session_start.go` — SessionStart 자동 감지, 본 SPEC 변경 없음
- `internal/template/templates/.claude/settings.json.tmpl` — settings 템플릿, 본 SPEC 변경 없음 (R5 검증 대상)

## Exclusions (What NOT to Build)

- **EX-1**: Mainland China 엔드포인트 (`Z_AI_MODE=ZHIPU`, `https://open.bigmodel.cn/api/anthropic`) 지원은 본 SPEC v0.1 에서 명시적으로 제외하며 v0.2 로 연기한다. 본 SPEC 의 enable 명령은 `Z_AI_MODE=ZAI` (국제) 만 설정한다. 향후 `--region zhipu` 플래그 추가는 별도 SPEC 으로 다룬다.
- **EX-2**: 사용자 GLM Coding Plan 등급 (Lite vs Pro) 의 사전 API 검증은 구현하지 않는다. enable 시 단순 경고("Vision 과 Web Search 는 Pro 플랜 이상 필요") 를 출력하고 활성화는 진행한다. 런타임 401/403 응답은 Z.AI 가 반환하므로 본 SPEC 은 *명확한 가이드* 만 제공한다.
- **EX-3**: 기존 사용자 `.mcp.json` (project scope) 또는 `~/.claude.json` (user scope) 의 구조 변경 또는 다른 MCP 엔트리(`context7`, `sequential-thinking`, `moai-lsp`) 에 대한 정책 변경은 본 SPEC 범위 밖이다.
- **EX-4**: `moai cc` ↔ `moai glm` ↔ `moai cg` 모드 전환 시 자동으로 zai-mcp-server 엔트리를 enable/disable 하는 로직은 *구현하지 않는다*. 본 SPEC 의 enable/disable 은 사용자 명시적 명령에 의해서만 동작한다 (R5 결정, 단순성 원칙).
- **EX-5**: `~/.claude.json` 스키마 검증 강화 (예: `alwaysLoad` 또는 다른 필드의 type-check) 는 본 SPEC 범위 밖이다. SPEC-CC2122-MCP-001 의 결정과 동일하게 Claude Code 런타임이 자체 검증한다고 가정.
- **EX-6**: 사전 sample test (Vision OCR ping 등) 자동 실행은 v0.1 에서 제외한다 (research.md OQ-6). 향후 `--test` 옵트인 플래그로 검토 가능.

## Risks Reference

상세 위험 분석 (R1~R5) 과 완화 전략은 `plan.md` §Risks 와 `research.md` §4 를 참조한다.

## Acceptance Reference

상세 Given-When-Then 시나리오, edge cases, 품질 게이트 기준, Definition of Done 은 `acceptance.md` 를 참조한다. 본 spec.md 는 acceptance criteria 를 *간단 요약* 만 제공하며, 구체적 시나리오는 acceptance.md 가 single source of truth 이다.

### Acceptance Summary (간단 요약)

- 사용자가 `moai glm tools enable vision` 실행 시 Vision MCP 가 Claude Code 에 노출되어야 함
- `moai glm tools disable all` 실행 시 zai-mcp-server 엔트리가 제거되어야 하며 다른 MCP 엔트리는 보존됨
- Node.js < 22 인 환경에서 enable 명령은 graceful fail 해야 함
- `GLM_AUTH_TOKEN` 부재 시 enable 명령은 토큰 등록 가이드를 출력 후 종료해야 함
- 모드 전환 (`moai cc` ↔ `moai glm`) 후에도 zai-mcp-server 엔트리는 사용자 명시 disable 까지 유지되어야 함

## Implementation Reference

마일스톤, 우선순위, 기술적 접근 방식, 위험 완화는 `plan.md` 를 참조한다.

## References

- Z.AI 공식 MCP 문서: https://docs.z.ai/devpack/mcp/vision-mcp-server
- `.moai/specs/SPEC-GLM-MCP-001/research.md` — 본 SPEC 의 deep research 산출물
- `.moai/specs/SPEC-GLM-001/spec.md` — 베이스 GLM 호환성 자동화
- `.moai/specs/SPEC-CC2122-MCP-001/spec.md` — `.mcp.json.tmpl` 변경 패턴 참조
- `.moai/specs/SPEC-LSPMCP-001/spec.md` — moai-lsp MCP 통합 패턴 참조
- CLAUDE.local.md §13 (GLM Integration Testing 정책)
- CLAUDE.local.md §15 (Template Language Neutrality)
- `internal/cli/glm.go` — GLM 환경 주입 진입점 (참조만, 본 SPEC 단계에서 수정 없음)
- `internal/cli/github_auth.go` — `moai github auth glm` 패턴 참조
