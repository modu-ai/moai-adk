# Proposal: MoAI-ADK Console
*Generated: 2026-05-04 | Idea: IDEA-002*

> ⚠️ **DEFERRED to v3.0 release window** (2026-05-04 사용자 결정)
>
> 본 proposal의 stack 결정(Bun + Hono + **React** + **Claude Agent SDK**)은 paradigm shift로 v3.0까지 보류.
> v3.0 재개 시 stack 재설계: **HTMX 기반** (server-rendered + 부분 업데이트), **Claude Agent SDK 미사용** (메커니즘 v3.0 plan 시 재결정).
>
> 현재 v2.x 사이클(~v2.20)은 Console 작업 정지. 본 문서는 v3.0 시 paradigm 재설계 input으로 보존.
>
> 결정 메모리: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/project_idea002_console_deferred_v3.md`

## Executive Summary

MoAI-ADK Console은 단일 사용자(GOOS) 로컬호스트 환경에서 동작하는 document-centric CRUD 워크벤치이다. 4종 settings 파일(`user/project/design/harness`), SPEC 트리(`.moai/specs/SPEC-*/`), rule 트리(`.claude/rules/**/*.md`)에 대해 schema-aware 폼과 frontmatter-aware 마크다운 에디터를 제공하고, 우측 사이드패널 채팅을 통해 Claude Agent SDK에 지시를 보내 SSE로 응답을 스트리밍한다. 기존 Claude Code의 generic 인터페이스와 달리 MoAI-ADK 도메인(EARS 검증, frontmatter 무결성, rule `paths:` 적용 범위)을 first-class 시민으로 다루며, 단일 사용자·인증 없음·로컬호스트만 가정하여 코드 단순성을 유지한다.

## Capabilities

사용자가 Console에서 수행할 수 있는 행위(구현 방식이 아닌 능력 기준):

### 4 Settings CRUD
- 4개 카테고리(`user/project/design/harness`)의 YAML 설정을 schema-driven 폼으로 편집
- 폼·raw YAML 에디터 토글 지원 (power user의 복잡한 YAML 패턴 보존)
- inline 검증 + save 시 무결성 lint
- 외부 변경(터미널 직접 편집) 자동 감지 후 reload

### SPEC 뷰어 + EARS-aware 에디터
- `.moai/specs/SPEC-XXX/` 트리 탐색 + 5개 artifact(`spec/plan/acceptance/research/progress`) 통합 뷰
- frontmatter 폼(SPEC 메타데이터) + 본문 markdown 에디터 분리 편집
- EARS 절(`When [trigger], the [system] shall [response]`) 패턴 lint (v0.1: warning-only)
- frontmatter 무결성 hard validation (구조 깨짐 방지)
- 활성 SPEC 대시보드(phase·진행 상태)를 홈 페이지로 노출

### Rule 뷰어 + 에디터
- `.claude/rules/**/*.md` 트리 탐색 + `paths:` 필드 기반 적용 범위 시각화
- frontmatter 폼 + 본문 markdown 에디터 분리
- frontmatter 무결성 보존

### Chat-based Claude Agent SDK Directive Panel
- 우측 사이드패널 고정, 모든 탭에서 항상 가시
- 현재 열린 SPEC/rule 컨텍스트를 시스템 프롬프트에 자동 prefill
- SSE 토큰 스트리밍, assistant·tool_use·result 메시지 실시간 표시
- 다중 턴 세션 유지 (`continue: true`)
- per-session token usage 및 turn count 표시

### 공통 UX
- 라이트/다크 테마
- 기본 반응형 (단일 사용자 데스크톱 우선)
- random localhost 토큰 기반 API 인증 (외부 프로세스 격리)

## Stack & Architecture Decisions

Phase 1에서 사용자가 락인한 기술 스택을 재확인:

- **런타임**: Bun (단일 JavaScript 런타임 + 패키지 매니저 + 워크스페이스)
- **백엔드**: Hono (TypeScript 웹 프레임워크, `streamSSE()` native 지원, Bun 호환)
- **프론트엔드**: React (TypeScript)
- **타입 안정성**: TypeScript end-to-end (모노레포 `packages/shared` 패턴, Hono RPC)
- **Agent 통합**: `@anthropic-ai/claude-agent-sdk` (TypeScript v0.2.x), `query()` async generator
- **검증**: Zod (settings 스키마 + EARS 패턴 + Hono 입력 검증)
- **모노레포 구조**: BHVR 스타일 (`client/`, `server/`, `shared/`)

### Architecture Constraints (Phase 5 Risk Register 반영)

- **Localhost-only**: 127.0.0.1 binding 강제, 외부 IP binding 코드 경로 차단
- **단일 사용자 인증**: 시작 시 random token 발급 → 모든 API 검증 (future multi-user 확장 시 토큰 발급기만 교체)
- **파일 무결성**: 모든 save 전 mtime 재확인, 외부 변경 감지 시 conflict resolution UI
- **Claude Agent SDK 가드**: `maxTurns ≤ 20`, per-session token usage 표시, 비용 임계값 confirmation
- **Hono 버전 핀**: backend·frontend 정확히 동일 버전 (RPC type 추론 깨짐 방지)
- **CI 타입 검사**: Bun은 typecheck 안 함 → `tsc --noEmit` 별도 단계 필수

## Out of Scope (v0.1)

다음 항목은 명시적으로 제외됨. 추후 SPEC plan 진입 시 게이트에서 재확인:

- ❌ **Multi-user / authentication / RBAC**: 단일 사용자(GOOS) 가정 락인
- ❌ **Cloud / SaaS / Vercel / Railway 배포**: localhost-only
- ❌ **IDEA-001 Cockpit의 7-screen 모니터링** (pipeline DAG, agent roster, hook monitor, harness state, TRUST 5 quality gates, sandbox/permission panel, skill/rule library search): read-only monitoring은 archived (commit 953ed336d)
- ❌ **실시간 협업** (CRDT, Y.js 등)
- ❌ **모바일 native 앱**
- ❌ **Analytics / telemetry**
- ❌ **MCP server config UI**: 사용자 명시적 요구 부재. `.mcp.json` 직접 편집 유지
- ❌ **워크플로우 자동화** (`/moai plan/run/sync` GUI 트리거): Console은 문서·설정 surface, 자동화는 기존 CLI/Claude Code 유지

## SPEC Decomposition Candidates

- SPEC-V3R3-CONSOLE-001: Bun+Hono+React 모노레포 스캐폴드 + Claude Agent SDK 와이어링 + dev server + localhost 토큰 인증
- SPEC-V3R3-CONSOLE-002: 4 settings(`user/project/design/harness`) schema-driven CRUD 폼 + raw YAML 토글 + 외부 변경 감지
- SPEC-V3R3-CONSOLE-003: SPEC 뷰어 + EARS-aware 에디터 (5 artifact 통합 뷰, frontmatter 폼, EARS warning-only lint, 무결성 hard validation)
- SPEC-V3R3-CONSOLE-004: Rule 뷰어 + 에디터 (`.claude/rules/**` 트리, `paths:` 시각화, frontmatter 보존)
- SPEC-V3R3-CONSOLE-005: Chat-based Claude Agent SDK directive 사이드패널 (SSE 스트리밍, 컨텍스트 자동 prefill, maxTurns 가드, token usage UI)

### Recommended Wave Grouping

**Wave 1** (FOUNDATION + 2 priority CRUDs, 직렬):
1. SPEC-V3R3-CONSOLE-001 (FOUNDATION) — 모노레포·런타임·인증·dev server, 후속 SPEC의 전제조건
2. SPEC-V3R3-CONSOLE-002 (SETTINGS) — schema-driven CRUD 패턴 확립, SPEC/rule editor의 frontmatter 폼 패턴 재사용 기반
3. SPEC-V3R3-CONSOLE-003 (SPEC editor) — daily 사용 빈도 최대 surface

**Wave 2** (Wave 1 머지 후, 병렬 가능):
4. SPEC-V3R3-CONSOLE-004 (RULES) — SPEC editor의 frontmatter 패턴 재사용, 독립 진행 가능
5. SPEC-V3R3-CONSOLE-005 (CHAT) — FOUNDATION의 Hono SSE + Claude Agent SDK 와이어링 위에 채팅 UI 추가, 독립 진행 가능

Wave 2의 두 SPEC은 서로 다른 파일 범위(`.claude/rules/**` vs 채팅 컴포넌트)이므로 병렬 안전.

## Acceptance Heuristics

각 SPEC의 high-level 성공 기준 (EARS 형식 변환은 manager-spec이 `/moai plan` 단계에서 수행):

### SPEC-V3R3-CONSOLE-001 (FOUNDATION)
- `bun install && bun dev` 실행 시 `http://localhost:PORT` 접근 가능
- 외부 IP에서 접근 시도는 차단 (127.0.0.1 binding 검증)
- random token 없이 API 호출 시 401 반환
- backend·frontend 모두 TypeScript end-to-end 타입 추론 동작
- CI에서 `tsc --noEmit` 통과

### SPEC-V3R3-CONSOLE-002 (SETTINGS)
- 4개 카테고리(`user/project/design/harness`) 모두 read+create+update+delete 가능
- 각 폼은 Zod 스키마 기반 자동 생성 (필드 타입 자동 매핑)
- raw YAML 토글 시 주석·앵커·alias 보존
- 외부 변경(터미널 직접 편집) 감지 후 reload 동작
- save 시 mtime 재확인 + conflict UI 동작

### SPEC-V3R3-CONSOLE-003 (SPEC editor)
- `.moai/specs/SPEC-XXX/` 5 artifact 모두 렌더 + 편집 가능
- frontmatter 구조 깨짐 시 save 차단 (hard validation)
- EARS 패턴 미준수 시 inline warning 표시 (block 아님)
- 활성 SPEC 대시보드가 홈 페이지에서 phase·진행 상태 표시

### SPEC-V3R3-CONSOLE-004 (RULES)
- `.claude/rules/**/*.md` 모두 트리 탐색 + 편집 가능
- `paths:` frontmatter 필드가 적용 범위 시각화 (예: 어떤 파일 패턴에 매칭되는지 표시)
- frontmatter 무결성 보존

### SPEC-V3R3-CONSOLE-005 (CHAT)
- 사이드패널이 모든 탭에서 항상 가시
- 사용자가 prompt 입력 → SSE 스트리밍 → first token latency < 1초
- assistant 텍스트·tool_use 호출·result 모두 구분 표시
- 현재 열린 SPEC/rule 컨텍스트가 시스템 프롬프트에 자동 prefill
- `maxTurns ≤ 20` 강제, per-session token usage UI 표시
- 다중 턴 세션 유지 (continue 동작)

## Open Questions (deferred to /moai plan)

다음은 brain 단계가 아닌 SPEC plan 단계에서 manager-spec이 해결할 항목이다:

1. **EARS lint 패턴의 정확한 정규식**: `When/While/Where/If` 변형의 어떤 부분집합을 v0.1에서 지원? 각 변형별 false positive 임계 어떻게 측정?
2. **Settings YAML 주석 보존 방식**: js-yaml round-trip vs yaml.js (eemeli/yaml) 라이브러리 선택, 어느 쪽이 anchor/alias 보존 정확?
3. **Claude Agent SDK 컨텍스트 prefill 포맷**: 시스템 프롬프트에 SPEC 전체를 inject vs ID + path만, 토큰 비용·정확도 트레이드오프 측정 방식?
4. **외부 변경 감지 메커니즘**: chokidar vs Bun native FSWatcher, 단일 사용자 가정에서 어느 쪽이 reliability 우월?
5. **localhost 토큰 저장 위치**: localStorage vs sessionStorage vs HttpOnly cookie, XSS 위험 vs 사용성 트레이드오프?
6. **모노레포 구조 세부**: `packages/shared` vs `shared/`, Turborepo 도입 vs Bun workspaces 단독, dev 단계에서 hot reload 행동 차이?
7. **Frontend state management**: 채팅 세션·열린 SPEC·외부 변경 알림을 어떤 store(Zustand/Jotai/native React)로? 단일 사용자 가정상 단순한 선택 우선?
8. **EARS·SPEC frontmatter 스키마의 정식 위치**: Console이 자체 정의 vs `.moai/config/`에서 읽기? 후자가 MoAI 진영과 일관성 유지에 유리하지만 v0.1 시점 schema 위치 결정 필요?

이 질문들은 acceptance heuristics를 EARS로 변환하는 plan 단계에서 답변되며, 본 proposal은 capability 수준에서만 정의한다.
