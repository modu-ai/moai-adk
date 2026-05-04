# Idea: MoAI-ADK Console — local web GUI for MoAI-ADK developers to manage settings, view/edit SPEC and rule documents, and direct the Claude Agent SDK via chat
*Session: 2026-05-04 | Phase 4 Converge → Phase 5 Critical Evaluation appended*

## Selected Concept

**MoAI-ADK Console**은 단일 사용자(GOOS) 로컬호스트 환경에서 동작하는 **document-centric CRUD 워크벤치**이다. 4개 settings 파일·SPEC 트리·rule 트리에 대해 schema-aware 폼·frontmatter-aware 에디터를 제공하고, 우측 고정 채팅 패널을 통해 Claude Agent SDK 지시를 내려 응답을 SSE로 스트리밍한다. 워크플로우 자동화나 모니터링이 아니라 "문서를 안전하게 편집하고 즉시 agent에 지시한다"는 두 축에 집중한다.

Phase 2에서 발산한 12개 각도 중 다음 클러스터들의 교집합으로 수렴:
- **Layout: 다중 화면 탭 기반** (단일 페이지 command-palette는 빠른 호출에 강하지만 4개 settings 동시 비교·SPEC 본문 편집 같은 long-form 작업에는 화면 분리가 우월)
- **Agent: copilot 사이드패널** (항상 보이지만 메인 작업영역은 문서 편집. agent를 도구로만 두는 page-based는 즉시성을 잃음)
- **Home priority: SPEC dashboard** (active SPEC 진행 현황 + 빠른 진입이 일상 워크플로우에 가장 자주 호출됨. settings는 second-class)
- **Edit philosophy: schema-driven form + raw editor 토글** (4 settings는 form 우선, SPEC/rule은 본문 markdown 편집 + frontmatter form 분리)

## Lean Canvas

### Problem
1. 4종 settings YAML(`user/project/design/harness`)을 터미널·IDE에서 직접 편집해야 함. 스키마 검증·실수 방지 안전망 부재
2. SPEC 5종 마크다운 문서(`spec/plan/acceptance/research/progress`)와 rule 트리(`.claude/rules/**/*.md`)가 frontmatter 무결성·탐색성 면에서 GUI 부재로 인지 비용 높음
3. Claude Agent SDK 지시가 별도 채널(Claude Code 인터페이스)에서만 가능. MoAI-ADK 도메인 컨텍스트가 자동 prefill되는 채팅 surface 부재

### Customer Segments
- **Primary**: GOOS — 프로젝트 저자, 일상적으로 SPEC·rule·settings를 모두 편집하는 단일 power user
- **Secondary**: 향후 MoAI-ADK 채택자 중 터미널·IDE 직접 편집보다 GUI를 선호하는 개발자 (현재는 가설)

### Unique Value Proposition
"MoAI-ADK 문서·설정을 frontmatter 깨질 걱정 없이 편집하고, 그 자리에서 Claude Agent SDK에 지시하는 단일 로컬 워크벤치." — 터미널의 power와 GUI의 안전성을 동시에, MoAI 도메인에 특화.

### Solution
1. **4 Settings CRUD 폼** — Zod 스키마 기반 자동 폼 생성, YAML round-trip(주석·앵커 보존), inline 검증
2. **SPEC viewer + EARS-aware editor** — `.moai/specs/SPEC-*/` 트리, 5개 파일 통합 뷰, frontmatter 폼 + markdown 본문 분리, EARS 절(`When [trigger], the [system] shall [response]`) 패턴 lint
3. **Rule viewer + editor** — `.claude/rules/**` 트리, `paths:` 필드 기반 적용 범위 시각화, frontmatter 무결성 보존
4. **Chat-based Claude Agent SDK directive panel** — 우측 사이드패널 고정, 현재 열린 SPEC/rule 컨텍스트 자동 prefill, SSE 토큰 스트리밍, tool call·assistant 메시지 실시간 표시

### Channels
- 단일 사용자 로컬 설치: `bun install && bun dev` → `http://localhost:PORT` 시스템 브라우저 접근
- 향후 secondary 사용자: `moai console` CLI 명령으로 sidecar 실행 (가설, OUT for v0.1)

### Revenue Streams
N/A: localhost-only personal tool. 수익 모델 부재.

### Cost Structure
Development time. 외부 서비스·인프라 비용 없음 (Claude Agent SDK API 호출 비용은 사용자 본인 계정).

### Key Metrics
**필수 (Dimension 3 락인)**:
- 4 settings 모두 GUI에서 read+create+update+delete 가능
- SPEC viewer가 모든 `.moai/specs/SPEC-XXX/` artifact 렌더; 편집 시 frontmatter 무결성 보존 (lint 통과율 100%)
- Rule viewer가 `.claude/rules/**/*.md` 모두 렌더; 편집 시 frontmatter 보존
- 채팅 패널 SSE 응답 first token latency < 1초; 세션 다중 턴 유지

**보조 지표**:
- 일일 사용 빈도 (GOOS 개인 사용 기준 daily 1+ 회)
- 외부 변경(터미널 직접 편집) 감지 + 재로드 정상 동작
- 4 settings 폼 검증 오류 false positive 0건

### Unfair Advantage
1. **MoAI-ADK 도메인 first-class 통합**: SPEC/rule frontmatter 스키마, EARS 검증, `paths:` 적용 범위 — 외부 generic 도구가 따라잡으려면 MoAI 컨벤션 전체 구현 필요
2. **Claude Agent SDK + MoAI 컨텍스트 자동 prefill**: Console이 현재 작업 SPEC/rule 컨텍스트를 시스템 프롬프트에 자동 주입. Claude Code generic 인터페이스보다 즉시성·정확성 우월
3. **단일 사용자 가정**: 인증·멀티테넌시 부담 제거 → 코드 단순성·반응 속도에서 경쟁자 대비 우위
4. **저자 본인이 power user**: 사용 흐름 피드백 루프 즉시 반영 가능. 외부 도구로는 도달 불가능한 fit

## Why This Concept Won

### Phase 2 발산 12개 각도 중 제거된 옵션과 이유

| 제거된 각도 | 제거 이유 |
|------------|----------|
| 단일 페이지 command-palette | 4 settings·SPEC·rule을 동시에 비교·작성하는 long-form 워크플로우에 부적합 |
| Agent-as-tool (별도 페이지) | 일상 문서 편집 중 즉시 호출 불가 → MoAI 컨텍스트 prefill 가치 상실 |
| Settings-priority home | 일상 빈도가 SPEC 편집보다 낮음. settings는 secondary nav로 충분 |
| Workflow-phase-aware home | plan/run/sync 자동화는 OUT (Cockpit-archived 영역 재현 위험) |
| Form-only CRUD | SPEC/rule 본문은 markdown prose. 폼만으로 부족 |
| Editor-only CRUD (raw) | 4 settings YAML에서 frontmatter 실수 위험 — schema-driven form 필요 |
| Read-only preview default | 일상 작업이 편집이므로 preview default는 마찰 |

### 수렴 결과의 정당성

- **Layout**: 다중 화면 탭이 4 도메인(Settings/SPECs/Rules/Chat)을 명확히 분리. 단, Chat은 사이드패널로 모든 탭에서 항상 가시
- **Edit**: hybrid — 4 settings는 form-first(스키마 강함), SPEC/rule은 markdown editor + frontmatter form 분리(본문 prose 우선)
- **Home**: SPEC dashboard가 daily 사용 빈도 최대 + Phase 1 사용자 우선순위와 일치
- **Agent positioning**: copilot 사이드패널 — Claude Agent SDK가 현재 컨텍스트를 자동 받음, but 사용자 메인 작업은 문서 편집

이 수렴은 Lean Canvas의 모든 9 블록과 일관되며, Phase 1 사용자 락인된 IN/OUT 스코프를 정확히 충족한다.

---

## Evaluation Report

### Critical Evaluation Findings

#### Weakness 1 — EARS 검증 자체 구현의 복잡성
- **증거**: Phase 3 research에서 시장에 EARS-aware editor 부재 확인. Astro Editor·Front Matter CMS는 frontmatter form만, EARS clause 패턴(`When/the/shall`)은 직접 lint 구현 필요
- **심각도**: 중. EARS 검증이 부정확하면 SPEC 작성 UX는 raw markdown editor 대비 별 우위 없음
- **완화책**: v0.1에서는 EARS lint를 "warning only" 수준으로 시작 (block 아님). frontmatter 구조(`acceptance.criteria[]` 배열) 강제만 hard validation. EARS 자연어 패턴은 점진 강화

#### Weakness 2 — Claude Code와의 기능 중복 가능성
- **증거**: Claude Code 자체가 코드·문서 편집 + agent 채팅 + tool execution을 모두 제공. Console이 동일 기능을 단순 재포장하면 "왜 Claude Code 대신 Console?"에 답하기 어려움
- **심각도**: 중-고. UVP가 흔들리면 daily 사용 동기 약화
- **완화책**: 명확한 차별점 2개에 집중 — (1) 4 settings 통합 schema-aware CRUD (Claude Code는 generic file edit), (2) MoAI-ADK 컨텍스트 자동 prefill (rule `paths:` 시각화·SPEC frontmatter 무결성 lint·EARS 검증)

#### Weakness 3 — 단일 사용자·인증 없음의 보안 표면
- **증거**: 127.0.0.1 binding이라도 동일 머신의 다른 프로세스(브라우저 확장, 악성 npm 패키지 등)가 localhost API 접근 가능. file write 권한이 모든 SPEC/rule/settings에 미침
- **심각도**: 중. 직접 공격 표면은 작지만 single-user assumption이 공격 벡터를 가림
- **완화책**: 시작 시 random localhost token 발급 → 브라우저 URL 쿼리·localStorage에 저장. 모든 API 요청에 토큰 검증. 외부 IP binding 코드 경로 자체 차단 (test-time도 127.0.0.1만 허용)

#### Weakness 4 — 동시 편집(Console + 터미널) race condition
- **증거**: GOOS는 일상적으로 터미널·IDE에서도 SPEC·rule 직접 편집. Console이 파일 캐시를 들고 있는 동안 외부 변경이 발생하면 last-write-wins → 사용자 작업 silently 손실
- **심각도**: 중-고. 데이터 손실은 신뢰 즉시 붕괴
- **완화책**: 파일 mtime watch + 모든 save 전 mtime 재확인. 외부 변경 감지 시 사용자에게 conflict resolution UI 표시 (reload / overwrite / merge). FOUNDATION SPEC에 file watcher 포함 필수

#### Weakness 5 — Claude Agent SDK rate limit·비용 노출
- **증거**: 채팅 패널이 항상 활성화되면 무심코 호출이 누적. SDK는 사용자 본인 API key 사용 → 직접 비용 발생
- **심각도**: 저-중. 단일 사용자 가정상 폭발적 누적은 어려우나 long-running tool loop 시 위험
- **완화책**: `maxTurns ≤ 20` 강제, per-session token usage 표시 UI, 비용 초과 임계값 시 사용자 confirmation prompt

### First Principles Decomposition

#### Assumption 1 — "GUI가 터미널보다 SPEC 편집에서 더 안전하다"
- **검증 상태**: 부분 검증. frontmatter 무결성·EARS 검증은 GUI에서 우월. 그러나 GOOS가 vim·neovim 숙련자라면 단순 raw 편집은 GUI보다 빠름
- **결론**: GUI의 우위는 검증 자동화에 한정. 단순 prose 편집 자체에는 큰 우위 없음. 따라서 EARS 검증·frontmatter 보존이 핵심 가치 — 이 기능 구현 우선순위가 최상

#### Assumption 2 — "단일 사용자·localhost-only가 영구 제약이다"
- **검증 상태**: 락인됨 (Phase 1). 단, secondary 사용자(future MoAI-ADK 채택자) 가설이 활성화되면 multi-user·remote 요구가 발생할 수 있음
- **결론**: 아키텍처를 future-proof — auth/RBAC을 v0.1 스코프 외로 두되, API 레이어는 인증 토큰 기반으로 처음부터 설계 (localhost-only 토큰도 형식상 같은 패턴). 향후 확장 시 토큰 발급기만 교체

#### Assumption 3 — "Claude Agent SDK가 SPEC/rule 컨텍스트를 자동 prefill받으면 가치가 발생한다"
- **검증 상태**: 가설. 실제 prefill된 컨텍스트가 generic Claude Code 호출보다 결과물이 더 정확한지 측정되지 않음
- **결론**: Phase 6 acceptance heuristics에 "동일 prompt를 generic vs Console-prefilled로 호출 시 후자가 더 정확/관련성 높음" 검증 항목 포함. 측정 방식은 manager-spec이 plan에서 정의

#### Assumption 4 — "4 settings의 schema-driven form이 raw YAML 편집보다 일관되게 더 빠르다"
- **검증 상태**: 일반적 통념. 그러나 power user GOOS가 anchor·alias·complex YAML pattern을 사용한다면 form은 표현력 부족 → raw editor toggle 필수
- **결론**: form-first + raw toggle 양면 지원. settings YAML이 복잡할수록 raw editor 사용 빈도 ↑ 가능성 인정

#### Assumption 5 — "Bun + Hono + React + Claude Agent SDK 조합이 안정적이다"
- **검증 상태**: Phase 3 research에서 BHVR 스택·Hono SSE·Claude Agent SDK TypeScript 모두 검증됨. 단, Hono RPC 버전 mismatch·Bun typecheck 누락은 알려진 함정
- **결론**: 안정적이지만 함정 회피 의무 — Hono 버전 정확히 핀, CI에서 `tsc --noEmit` 별도, Claude Agent SDK `maxTurns` guard 필수

### Pivot Rationale Validation (Cockpit → Console)

Phase 1에서 IDEA-001 Cockpit(7-screen read-only monitoring) → IDEA-002 Console(CRUD + 채팅) 피벗이 정당했는지 research 근거로 검증:

- **읽기 전용 모니터링의 한계**: Phase 3 research의 "7 Best UI Frameworks for AI Agents 2026"에 따르면 streaming + observability가 표준이 되었으나, 사용자가 실제로 원하는 것은 monitoring을 넘어 "agent 실행 중 개입·지시"이다. read-only는 인지 부담 추가만 발생
- **CRUD가 daily workflow와 일치**: 단일 사용자(GOOS) 일상 행위는 SPEC 편집 + settings 조정 + agent 호출. monitoring은 부수적 (필요 시 터미널 로그로 충분)
- **Cockpit 스코프(7-screen)가 대규모**: 7개 화면을 실제 사용 빈도순으로 나열하면 daily 호출되는 것은 1-2개. 나머지는 dead UI 위험. Console의 4 surface(Settings/SPECs/Rules/Chat)는 모두 일상 대상
- **결론**: 피벗은 정당. 단, "monitoring 추가하면 어떨까" 압력 재발 방지 위해 proposal.md에 명시적 OUT 표시 필수

### Risk Register

| # | Risk | Likelihood | Impact | Mitigation |
|---|------|-----------|--------|-----------|
| 1 | **스코프 크리프** — Cockpit 7-screen 모니터링 부활, MCP config UI 추가, 멀티유저 요구 | High | High | proposal.md에 OUT 명시 + 모든 SPEC plan 진입 시 OUT 재확인 게이트 |
| 2 | **EARS lint 자체 구현 복잡도** — 정규식 패턴 다양성, false positive | Medium | Medium-High | v0.1 warning-only, frontmatter 구조 강제만 hard validation, 점진 강화 |
| 3 | **동시 편집 race condition** — Console·터미널 동시 수정으로 데이터 손실 | High | High | 모든 save 전 mtime 재확인, conflict UI, file watcher FOUNDATION 필수 |
| 4 | **Claude Code 기능 중복** — UVP 약화 | Medium | Medium-High | 차별점 2축(4 settings + MoAI 컨텍스트)에 집중, generic editor 기능 확장 자제 |
| 5 | **Claude Agent SDK rate/cost 폭주** — long tool loop, 무한 retry | Low-Medium | Medium | `maxTurns ≤ 20`, per-session token UI, 임계값 confirmation |

### Verdict

**Proceed with caveats.**

- 핵심 가치 (4 settings schema-aware CRUD + MoAI 컨텍스트 자동 prefill)에 집중하면 Claude Code와 차별화 가능
- Risk 1 (스코프 크리프)이 최대 위협 — proposal.md OUT 섹션 + plan 진입 게이트에서 반복 차단 필수
- Risk 2 (EARS), Risk 3 (race condition), Risk 5 (rate limit)는 v0.1 SPEC 분해 단계에서 명시적 mitigation 포함 필수
- Phase 1 사용자 락인된 스택(Bun+Hono+React+TS+Claude Agent SDK)은 Phase 3 research로 안정성 확인 — 변경 사유 없음
