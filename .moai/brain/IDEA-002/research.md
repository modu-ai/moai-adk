# Research: MoAI-ADK Console — 로컬 웹 GUI for MoAI-ADK 개발자
*Phase 3 — Brain Workflow | Date: 2026-05-04 | Idea: IDEA-002*

## Executive Summary

MoAI-ADK Console은 단일 사용자(GOOS) 로컬 환경에서 MoAI-ADK의 4종 설정(user/project/design/harness), SPEC 문서, 규칙 문서를 GUI로 CRUD하고 Claude Agent SDK 지시를 채팅으로 보내는 도구이다. 시장 조사 결과 (1) BHVR 스택(Bun+Hono+Vite+React)이 2026년 풀스택 TypeScript 모노레포의 사실상 표준으로 자리잡았고, (2) Claude Agent SDK의 TypeScript 패키지(`@anthropic-ai/claude-agent-sdk` v0.2.71)가 `query()` async generator로 스트리밍 응답·툴 실행·세션 지속성을 제공하며, (3) AG-UI 프로토콜과 SSE가 에이전트 스트리밍 UI의 사실상 표준이다. EARS 문법 인식 마크다운 에디터는 시장에 직접 제품이 없으나 Astro Editor·Front Matter CMS 같은 schema-aware editor 패턴을 차용하면 frontmatter 보존 + 본문 prose 편집이 가능함을 확인했다. localhost-only 단일 사용자 GUI는 Electron 대안으로 "로컬 HTTP 서버 + 시스템 브라우저" 패턴이 자원·복잡도 양면에서 우수하다는 결론이 다수 의견이다.

## Market Landscape

### 풀스택 TypeScript 모노레포 (Bun + Hono + React)

2026년 시점에서 BHVR(Bun + Hono + Vite + React) 스택이 모노레포 풀스택의 **사실상 기본 시작점**으로 자리잡았다. 핵심 이점:

- **공유 타입 패키지** (`packages/shared` 또는 `shared/src/types`)로 frontend·backend 간 end-to-end 타입 안정성 확보
- **Bun 워크스페이스 + Turborepo** 조합이 의존성 호이스팅·atomic commit·통합 CI 파이프라인 제공
- **Hono RPC**로 백엔드 핸들러 시그니처를 frontend에서 직접 참조 가능 (Hono 버전 핀 필수)
- **타입 검사 분리**: Bun은 타입 검사를 하지 않으므로 CI에서 `tsc --noEmit` 별도 필요

대안적 1-process 패턴(Hono가 React SPA + API를 함께 서빙)은 SEO 불필요·내부용·webview 시나리오에 적합. Console은 localhost-only 단일 사용자이므로 이 패턴이 더 단순함.

### Claude Agent SDK 생태계 (2026)

- 2025년 말 Claude Code SDK → Claude Agent SDK로 리네임: 코딩 외 deep research·video·notes 등 범용 agent runtime으로 확장
- 2026년 3월 시점: Python `v0.1.48` (PyPI), TypeScript `v0.2.71` (npm)
- 핵심 차이: **Anthropic Client SDK는 직접 API 호출 + tool loop 직접 구현**, **Agent SDK는 tool execution 내장 + autonomous loop**
- 기본 패턴: `query({ prompt, options })` async generator → assistant·tool_use·result 메시지 스트림
- 2026년 4월 8일 launch: Managed Agents (호스팅 REST API, `managed-agents-2026-04-01` 헤더)
- 프로덕션 안전장치: `maxTurns` 가드(권장 ≤20), exponential backoff, 비재시도 오류 분리(400/401/403/413)

### 로컬 단일 사용자 GUI 패턴

Electron 대안 비교 결과:

| 옵션 | 자원 | 복잡도 | Console 적합성 |
|------|------|--------|----------------|
| Electron (Chromium 번들) | 100-200MB+ | 중 | 과한 자원 사용 |
| Tauri (시스템 webview) | ~5-15MB | 중 | Rust 학습 비용 |
| **로컬 HTTP 서버 + 시스템 브라우저** | ~20MB | 저 | 최적 (자원·디버깅·deploy) |
| PWA | n/a | 저 | 파일시스템 접근 제약 |
| NW.js | 100MB+ | 중 | 과함 |

다수 의견: 단일 사용자 localhost 시나리오는 시스템 브라우저로 `http://localhost:PORT` 접근이 가장 단순하고 자원 효율적. Console 스택이 이 패턴과 정확히 일치(Bun이 Hono 서버 호스팅, 브라우저가 React SPA 렌더).

## User Needs

### 핵심 페인 포인트 (Phase 1 Discovery 검증)

- **YAML 설정 파편화**: `.moai/config/sections/{user,project,design,harness}.yaml` 4개 파일을 터미널·IDE에서 직접 편집해야 함. 스키마 검증·미리보기·실수 방지 안전망 부재
- **SPEC 문서 다중 파일 관리**: `.moai/specs/SPEC-XXX/` 하나당 spec/plan/acceptance/research/progress 5개 마크다운 + frontmatter. 터미널·IDE는 frontmatter 무결성 보장 못 함
- **Rule 트리 탐색 비용**: `.claude/rules/**/*.md` 트리가 깊고 path-scoped 로딩 규칙 때문에 어떤 규칙이 어떤 시점에 적용되는지 직관적이지 않음
- **Claude Agent SDK 직접 호출 부재**: 현재는 Claude Code 인터페이스를 통해서만 사용. Console이 지시 → 스트리밍 응답 받는 별도 채팅 패널 제공 시 일상 워크플로우에서 즉시 호출 가능

### 검증된 패턴 (외부 사례)

- **schema-aware editor**: Astro Editor가 Zod 스키마를 읽어 frontmatter를 폼으로 변환 (date → picker, enum → dropdown). 같은 패턴을 4개 settings YAML 파일에 적용 가능
- **frontmatter CRUD**: 2026년 emerging pattern — `frontmatter get/set/merge/validate` 명령어가 grep 대비 5K-10K 토큰 절감. Console SPEC editor는 frontmatter API 분리 + 본문 prose 분리 패턴이 최적
- **agent step observability**: 2026년 UI 프레임워크 평가의 핵심 항목 — 사용자가 agent의 reasoning·tool call·state delta를 실시간으로 보고 싶어함. AG-UI 프로토콜이 `TEXT_MESSAGE_CONTENT`·`TOOL_CALL_START/END`·`STATE_DELTA` 이벤트로 표준화

## Technical Ecosystem

### Hono — Bun 호환 SSE 스트리밍 (Context7 검증)

Hono `streamSSE()` API는 Bun 런타임에서 native 동작:

```typescript
// Hono /llmstxt/hono_dev_llms_txt 공식 예시 (검증됨)
import { streamSSE } from 'hono/streaming'

app.get('/sse', async (c) => {
  return streamSSE(c, async (stream) => {
    // Claude Agent SDK 메시지 stream 연결 가능
    await stream.writeSSE({ data: messageJson, event: 'agent-message', id: String(id++) })
  })
})
```

- `stream.onAbort()`로 정리 콜백 등록 (브라우저 탭 닫힘 시 SDK 호출도 중단)
- `streamText()`·`stream()` 변형: 텍스트·바이너리 스트리밍 지원
- Hono `4.x` 시리즈가 stable. RPC 사용 시 backend·frontend Hono 버전 핀 필수

### Claude Agent SDK TypeScript (Context7 검증)

`@anthropic-ai/claude-agent-sdk` 표준 사용 패턴 (검증됨):

```typescript
// Context7 /nothflare/claude-agent-sdk-docs 검증
import { query } from "@anthropic-ai/claude-agent-sdk";

for await (const message of query({
  prompt: userPrompt,
  options: {
    allowedTools: ["Read", "Edit", "Glob"],
    permissionMode: "acceptEdits",
    maxTurns: 20,    // 무한 루프 방지
    continue: true   // 세션 유지
  }
})) {
  // message.type: "assistant" | "tool_use" | "result"
  if (message.type === "assistant" && message.message?.content) {
    for (const block of message.message.content) {
      if ("text" in block) /* 텍스트 토큰 → SSE 전송 */;
      else if ("name" in block) /* tool 호출 → SSE event 전송 */;
    }
  }
}
```

- `query()`는 async generator → Hono `streamSSE()` 안에서 `for await` 직접 소비 가능
- `options.continue: true`로 동일 세션 유지 (다중 턴 채팅에 필수)
- `mcpServers` 옵션으로 사용자 정의 MCP 서버 주입 가능 (Console이 직접 MoAI 도구 노출 가능)

### Zod / 폼·YAML 스키마 검증

- `colinhacks/zod` v3/v4 안정. `@hono/zod-validator`로 Hono 라우트 입력 검증
- `paolostyle/hono-zod-openapi` (benchmark 94.8): Zod 스키마에서 OpenAPI 문서 자동 생성 → Console settings CRUD API에 적용 가능
- `alexmarqs/zod-config`: 파일 소스(YAML 포함)에서 Zod 검증된 설정 로드. Console 4 settings YAML 파서·검증 레이어로 직접 채택 가능

### EARS 문법 인식 에디터

직접 제품 부재. 가장 가까운 패턴:

- **Astro Editor** (`astroeditor.danny.is`): Zod schema → form-aware frontmatter editor. EARS 절을 frontmatter `acceptance.criteria[]` 배열로 모델링하면 form 입력 가능
- **Front Matter CMS** (VS Code extension): frontmatter UI + markdown 본문 분리 편집 패턴
- **OpenMark** (macOS): YAML frontmatter + markdown 동시 렌더
- 결론: EARS 검증은 Zod 스키마(`When [trigger], the [system] shall [response]` 정규식 패턴) + 사용자 정의 lint rule로 직접 구현해야 함. 시장 솔루션 차용 불가, 자체 구현 필수

### 스트리밍 UI 표준 (AG-UI / Vercel AI SDK)

- **AG-UI 프로토콜** (2026 emerging): Microsoft·CopilotKit이 추진하는 agent ↔ UI 표준. POST 요청 + SSE 응답, `TEXT_MESSAGE_CONTENT`·`TOOL_CALL_START/END`·`STATE_DELTA`·`RUN_ERROR`·`RUN_FINISHED` 이벤트 타입
- **Vercel AI SDK 6**: `streamUI()` + `Agent` abstraction, SSE format with start/delta/end IDs
- Console 채팅 패널은 AG-UI 이벤트 타입 차용 권장 (비록 표준 100% 채택 안 해도 mental model로 가치). 단, AG-UI dependency 추가는 과함 — Hono SSE + 자체 이벤트 enum으로 충분

## Risk Signals

### 기술 리스크

- **Hono RPC 버전 mismatch**: backend·frontend Hono 버전 다르면 type 추론 깨짐. 모노레포에서 Hono 정확한 버전 핀 (peerDependencies 활용)
- **Bun 타입 검사 누락**: Bun은 native typecheck 안 함 → CI에서 `tsc --noEmit` 별도 단계 필수
- **Claude Agent SDK rate limit**: API 호출 폭주 시 429 발생. `maxTurns ≤ 20` + exponential backoff 필수
- **EARS 검증 복잡도**: spec.md frontmatter는 EARS clause 배열. 무결성 깨질 시 `/moai run` 후속 워크플로우가 silently 실패. Zod 스키마·서버 측 검증·UI 차원 inline lint 3-tier 방어 필요

### 워크플로우 리스크

- **단일 사용자 가정의 함정**: 인증 없음 → 로컬 PORT 노출 시 동일 머신 다른 프로세스가 접근 가능. 127.0.0.1 binding + localhost-only로 제한 필수 (외부 노출 방지)
- **파일시스템 race condition**: Console과 터미널이 같은 SPEC 파일을 동시 편집 시 last-write-wins. 파일 수정 시각 비교 + 외부 변경 감지 후 재로드 필요
- **Claude Code 진영과의 중복 가능성**: Console이 Claude Code의 일부 기능을 재현할 위험. 명확한 차별점은 (1) MoAI-ADK 도메인 (SPEC/EARS/rules), (2) 4 settings 통합 CRUD — 이 2개 축에 집중

### 스코프 크리프 리스크

- **Cockpit 7-screen 모니터링 부활 유혹**: pipeline DAG·hook monitor·harness state 등은 archived. Console 출시 후 "monitoring 추가하면 어떨까" 압력에 대해 명시적으로 OUT 표시 (proposal.md에 anti-scope 섹션 필수)
- **MCP server config UI**: 사용자 명시적 요구 없으면 OUT 유지. MCP 설정은 .mcp.json 직접 편집이 현재 표준이고 잘 작동 중

## Opportunities

### 차별화 레버

- **EARS 문법 first-class 인식**: 시장 부재. Zod + lint + form helper 통합 시 SPEC 작성 UX 압도적 우위
- **MoAI-ADK 도메인 규칙 내재화**: rule frontmatter `paths:` 필드로 어떤 파일에 어떤 rule이 적용되는지 시각화 (rule visualizer)
- **Claude Agent SDK 직접 통합**: Console 채팅이 실제 MoAI 워크플로우(plan/run/sync) 컨텍스트를 prefill 가능. Claude Code는 generic agent 인터페이스, Console은 MoAI-specific surface
- **schema-aware form generation**: 4 settings YAML 스키마를 Zod로 정의하면 frontend form 자동 생성 + 검증 + OpenAPI 문서까지 일관

### 타이밍 요인

- 2026년 Bun/Hono 생태계 maturation 완료 → BHVR 스택 안정적 사용 가능
- Claude Agent SDK TypeScript v0.2.x 안정화 → 프로덕션 사용 risk 낮음
- AG-UI 프로토콜 emerging — Console 자체 이벤트 모델을 AG-UI 호환 가능하게 설계 시 향후 표준 채택 비용 최소화

## Sources Summary

| Source | Type | Relevance |
|--------|------|-----------|
| Anthropic Claude Agent SDK overview | technical_ecosystem | SDK 핵심 패턴·`query()` API·세션 관리 |
| stevedylandev/bhvr | technical_ecosystem | Bun+Hono+Vite+React 모노레포 표준 starter |
| Building Full-Stack TypeScript Monorepo with React and Hono | case_study | 모노레포 구조 검증 |
| Hono RPC TypeScript Project References | technical_ecosystem | 모노레포 Hono RPC 버전 핀 패턴 |
| 27 Open-source Electron Alternatives | competitor | 로컬 GUI 옵션 비교 (HTTP 서버 + 브라우저 우위 확인) |
| ClojureVerse — desktop app vs local webserver | user_research | localhost-only 단일 사용자 패턴 검증 |
| Astro Editor (schema-aware markdown editor) | competitor | frontmatter form 생성 패턴 |
| Front Matter CMS (VS Code) | competitor | frontmatter + body 분리 편집 |
| Mavin EARS official guide | technical_ecosystem | EARS 문법 정의 (When/the/shall) |
| AG-UI Protocol (Microsoft + CopilotKit) | technical_ecosystem | agent 스트리밍 이벤트 타입 표준 |
| Vercel AI SDK Stream Protocol | technical_ecosystem | SSE start/delta/end 이벤트 패턴 |
| Anthropic Managed Agents launch (2026-04-08) | market_data | hosted agent 옵션 (Console 자체는 self-hosted 유지) |
| Context7 `/llmstxt/hono_dev_llms_txt` Hono SSE | technical_ecosystem | Hono `streamSSE()` API 검증 |
| Context7 `/nothflare/claude-agent-sdk-docs` | technical_ecosystem | Claude Agent SDK TypeScript `query()` API 검증 |
| Context7 `/oven-sh/bun` | technical_ecosystem | Bun 런타임 + 워크스페이스 |
| Context7 `/colinhacks/zod` | technical_ecosystem | Zod schema validation (4 settings + EARS) |
| Context7 `/paolostyle/hono-zod-openapi` | technical_ecosystem | Hono + Zod + OpenAPI 자동화 |
| YAML Editor Tools 2026 — Red Hat YAML extension | competitor | YAML schema-aware editing 패턴 |
| Refine — CRUD Apps 2026 guide | technical_ecosystem | 단일 struct 안티패턴 → 분리 모델 권장 |
| 7 Best UI Frameworks for AI Agents 2026 | market_data | streaming + observability가 표준이 됨 |

Total sources: 20

## Sources (URL List)

- [Anthropic Claude Agent SDK overview](https://code.claude.com/docs/en/agent-sdk/overview)
- [BHVR — Bun + Hono + Vite + React monorepo](https://github.com/stevedylandev/bhvr)
- [Building a Full-Stack TypeScript Monorepo with React and Hono](https://blog.raulnq.com/building-a-full-stack-typescript-monorepo-with-react-and-hono)
- [Hono RPC And TypeScript Project References](https://catalins.tech/hono-rpc-in-monorepos/)
- [27 Open-source Electron Alternatives](https://medevel.com/27-os-electron-alternatives/)
- [ClojureVerse — desktop app vs local webserver vs electron](https://clojureverse.org/t/making-a-desktop-app-vs-local-webserver-vs-electron/2261)
- [Astro Editor — Schema-Aware Markdown Editor](https://astroeditor.danny.is/)
- [Front Matter CMS](https://frontmatter.codes/)
- [Alistair Mavin EARS official guide](https://alistairmavin.com/ears/)
- [AG-UI Protocol (Microsoft Learn)](https://learn.microsoft.com/en-us/agent-framework/integrations/ag-ui/)
- [AG-UI Protocol Bridging Agents (CopilotKit)](https://www.copilotkit.ai/blog/ag-ui-protocol-bridging-agents-to-any-front-end)
- [Vercel AI SDK Stream Protocols](https://ai-sdk.dev/docs/ai-sdk-ui/stream-protocol)
- [Claude Managed Agents launch (Anthem Creation)](https://anthemcreation.com/en/artificial-intelligence/claude-managed-agents-anthropic-ai/)
- [Hono streaming helpers (Context7 hono.dev)](https://hono.dev/docs/helpers/streaming)
- [Bun docs (Context7 oven-sh/bun)](https://bun.sh/docs)
- [Zod TypeScript-first schema validation](https://zod.dev/)
- [Hono Zod OpenAPI middleware](https://github.com/paolostyle/hono-zod-openapi)
- [Best YAML Editor Tools for Developers in 2026](https://www.digitaltoolpad.com/blog/yaml-editor)
- [Refine — What is a CRUD App? Complete Guide for 2026](https://refine.dev/blog/crud-apps/)
- [7 Best UI Frameworks for AI Agents (Fastio)](https://fast.io/resources/best-ui-frameworks-ai-agents/)

## Research Limitations

- WebSearch와 Context7 모두 정상 응답. 부분 실패 없음.
- EARS 문법을 직접 인식하는 시장 제품은 발견되지 않음 — Zod 기반 자체 검증 필요로 결론 (이는 발견사항이지 도구 실패 아님).
- Context7 `@hono/zod-validator` 직접 항목은 자동 매칭되지 않았으나 `colinhacks/zod` + `hono-zod-openapi` 조합으로 동일 유스케이스 커버 가능 확인.
