---
id: SPEC-CODEX-SESSION-MSG-001
title: "Codex-Claude 세션 간 양방향 메시징 — moai MCP 브로커 + A2A 정합 엔벨로프"
version: "0.1.0"
status: draft
created: 2026-08-23
updated: 2026-08-23
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: internal/sessionmsg
lifecycle: spec-anchored
tags: "codex, mcp, session-messaging, a2a, file-store, cross-session"
era: V3R6
tier: L
---

# SPEC-CODEX-SESSION-MSG-001 — Codex-Claude 세션 간 양방향 메시징

## HISTORY

- 2026-08-23 (plan-phase, v0.1.0) 최초 작성. 카드 t187 (운영자 지시 2026-08-23 "바로 진행"). 설계 비교 카드(Class C, Tier L)로서 3개 설계 축 — (i) A2A HTTP 전송 채택, (ii) A2A 정합 의미론만 채택(전송은 moai MCP 브로커 + 파일 스토어 하이브리드), (iii) 자체 스키마 — 을 research.md에서 실측 근거와 함께 비교하고 design.md에서 (ii)를 채택했다(운영자 질의 '구글 A2A' 반영). 측정 전제는 2026-08-23 rc.0 테스트(`.moai/reports/t187/codex-support-audit.md` §2.6·§3·§4)에서 왔다. 목표 산출: SPEC(설계) → 구현 → Claude 세션 × Codex 세션 실주고받기 e2e.

## §A. 검증된 기반선 (Measured Baseline)

이 절의 모든 주장은 2026-08-23 실측이며 출처를 달았다. 사전 지식이 아니다.

### §A.1 측정 전제 (rc.0, codex-cli 0.147.0)

| # | 전제 | 출처 |
|---|------|------|
| P1 | Codex CLI 0.147.0에는 네이티브 세션 메시징 런타임이 없다 | 감사 보고서 §0·§3 |
| P2 | codex→moai MCP 경로는 종단 실증됨: 수동 `codex mcp add moai -- moai mcp-server` 등록 → 세션에서 21개 도구 전부 인식 → `session_list` 실호출 성공, CODEX_HOME 격리 | 감사 §2.6 (t91 §5 실측) |
| P3 | 비대화형 `codex exec`에서 승인 정책 없이 MCP 도구를 부르면 `"user cancelled MCP tool call"` 로 실패한다 — e2e 절차는 승인 정책을 처리하거나 대화형 세션을 써야 한다 | 감사 §2.6·§3 gap 6 |
| P4 | Codex 세션은 moai 훅을 발화하지 않는다(hooks.json 생성기 부재, gap #1) — 세션 레지스트리 자동 등록 경로가 Codex에는 없다 | 감사 §3 gap 1 |
| P5 | MCP 서버는 세션이 소유한 수명이 긴 서브프로세스다 — 새 도구 노출에는 서버(=세션) 재시작이 필요하다 (버전 스큐, 카드 t184) | 카드 t187 지시 + `internal/cli/mcp_server.go:99-107` 구조 |

### §A.2 이 SPEC이 확장하는 코드 표면 (실측 file:line)

| 표면 | 위치 | 이 SPEC과의 관계 |
|------|------|------------------|
| MCP 도구 등록 `add()` 헬퍼 + per-tool enablement | `internal/cli/mcp_server.go:113-135` | 4개 신규 도구를 같은 패턴으로 등록 |
| 단일 카탈로그 선언 (가드 테스트가 등록 집합과 동등성 강제) | `internal/mcp/catalog.go:38-60` | 4개 도구 추가 — 콘솔 스키마는 카탈로그에서 자동 파생 |
| 세션 레지스트리 (스키마 동결) | `internal/session/registry.go:39, 84-95` | **수정하지 않는다** — Entry 스키마는 REQ-COORD-024로 동결. 주소 체계는 이를 참조만 하고 별도 스토어를 둔다 |
| 현재 세션 id 사이드채널 | `internal/session/registry.go:52` (`.moai/state/current-session-id.txt`) | Claude 세션 자기 식별 파생 후보로 참조 (설계는 명시적 등록으로 결정, design.md §4.2) |
| 자문적 록 패턴 | `internal/session/registry_lock_unix.go`, `LockTimeout` 2s (registry.go:71) | 사서함 동시성 제어에 같은 패턴 재사용 |
| 원자적 파일 쓰기 | `internal/atomicfile` 패키지 | 엔벨로프/기록 쓰기에 재사용 |
| codex_task 패밀리 (위임) | `internal/cli/codex_task.go`, `mcp_codex.go`, `.moai/state/codex-jobs/` | **경계 대상** — 이 SPEC은 이것과 절대 별칭되지 않는다 (§F.1) |
| C-HRA-008 정적 가드 선례 | `internal/cli/mcp_boundary_test.go` (`TestMCP_NoAskUserQuestion` / `TestMCP_NoInlineGetenv`) | 신규 도구 소스에 동일 패턴 적용 |
| 임계값 단일 원천 | `internal/config/defaults.go` (`DefaultCodexReviewGateTimeout` 등) | 세션 메시징 임계값도 여기에 둔다 |

### §A.3 부재 확인 (이 SPEC이 만들 것)

`grep -rn "session_msg\|session-msg" internal/ .moai/specs/` (2026-08-23, 본 워크트리) — 구현 히트 0건, 기존 SPEC 충돌 0건. 신규 패키지 `internal/sessionmsg`와 신규 상태 루트 `.moai/state/session-msg/`는 어느 쪽과도 충돌하지 않는다.

### §A.4 A2A 프로토콜 실측 (research.md §3 요약)

운영자 질의('구글 A2A')에 따라 공식 규격을 페치해 판정했다(사전 지식 판정 금지 지시). A2A v1.0 (Linux Foundation, 원래 Google 개발, Apache 2.0) — 핵심 데이터 모델: **AgentCard**, **Message** (`message_id`, `context_id`, `task_id`, `role(ROLE_USER|ROLE_AGENT)`, `parts`, `metadata`), **Part** (oneof `text|raw|url|data`), **Task/TaskState** (`SUBMITTED|WORKING|COMPLETED|FAILED|CANCELED|INPUT_REQUIRED|REJECTED|AUTH_REQUIRED`), **Artifact**. 전송은 HTTP/SSE/JSON-RPC 기반, 발견은 well-known AgentCard URL. 페치 원문 기록: research.md §3.

## §B. 사용자 스토리

**로서** 같은 머신에서 병렬로 작동하는 Claude 세션과 Codex 세션을 오케스트레이션하는 운영자,

**원한다** 두 종류 세션이 서로에게 짧은 메시지를 보내고 회신을 받을 수 있는 공용 우편함 — Claude 쪽은 이미 네이티브 `SendMessage`/`ListAgents` 런타임이 있지만 Codex 쪽은 그 어떤 메시징 런타임도 없다(P1),

**그래서** Codex 세션이 사실을 보고하고 Claude 세션이 그것을 폴링으로 수신하는, 세션 종류에 무관한 대칭 표면을 얻는다 — 네이티브 런타임은 그대로 두고(claude↔claude 무관여), 전송은 검증된 MCP 경로(P2) 위에 올린다.

## §C. 범위 요약과 범위 외 (Scope Summary & Out of Scope)

이 SPEC은 4개의 신규 moai MCP 도구(`session_msg_register` / `session_msg_list` / `session_msg_send` / `session_msg_poll`)와 그 아래 파일 스토어(에이전트 기록 + 사서함 + 클레임/만료)를 만든다. 엔벨로프는 A2A Message 데이터 모델에 필드명 수준으로 정합시킨다(축 (ii) 채택 — design.md §2). 교리(`cross-session-messaging.md`)의 Codex 경로 확장과 템플릿 미러를 포함한다. 납품 끝은 e2e 실주고받기다.

### Out of Scope — t88/M4 배선 생성기

- `.codex/hooks.json`·`.codex/config.toml`(mcp_servers.moai) 생성기는 t88 카드 소관이며 이 SPEC이 대체하지 않는다. 이 SPEC의 e2e는 감사 §4-4에서 이미 실증된 수동 `codex mcp add` 경로를 쓴다.
- t88이 착지하면 이 SPEC의 브로커는 자동 등록의 이득을 받는다 — 미래 통합 지점으로 §H에 기록만 한다.

### Out of Scope — claude↔claude 네이티브 메시징

- Claude 세션 사이의 메시징은 Claude Code 네이티브 `SendMessage`/`ListAgents` 런타임이 소유한다. 이 SPEC은 그 표면을 수정·대체하지 않는다. 브로커 도구는 세션 종류 무관 대칭이지만, claude↔claude 사용은 네이티브 경로를 권장하는 문서화만 한다.

### Out of Scope — 크로스 머신/네트워크 전송

- 단일 머신만 범위로 한다. HTTP 서버, 포트 할당, 원격 발견은 만들지 않는다. 축 (ii) 채택의 전제인 "A2A HTTP로의 이식 경로 가시성"은 엔벨로프 정합(REQ-CSM-002)으로만 확보하고, 전송 이식 자체는 별도 미래 SPEC이다.

### Out of Scope — AGENTS.md 정본 수정

- `AGENTS.md`(SPEC-AGENTS-MD-CANON-001 소관)은 수정하지 않는다. Codex 독자 표면은 도구 설명 문자열(MCP 세션이 컨텍스트에 로드 — P2로 실증)과 교리 파일이 담당한다.

### Out of Scope — Task 수명주기 상태기계·보안 스킴

- A2A TaskState 상태기계 관리, AgentCard의 `security_schemes`/`signatures`/`protocol_version` 협상은 구현하지 않는다. 참조 형상(reference shape)으로만 반영한다(design.md §3.1 정합 범위 명세).

## §D. 요구사항 (GEARS)

> 도메인 접두사 `REQ-CSM-NNN`. 기존 코드에 기대는 요구사항은 §A.2의 실측 위치를 인용한다. Tier L 요구사항 상한 25 중 15.

### D.1 전송·데이터 모델 (M1)

**REQ-CSM-001** (Ubiquitous) 세션 메시징 브로커는 단일 머신에서 Claude 세션과 Codex 세션 사이의 메시지를 moai MCP 도구 호출과 `.moai/state/session-msg/` 파일 스토어만으로 전달해야 하고, 네트워크 포트를 열거나 대기 프로세스를 추가로 띄워서는 안 된다.

**REQ-CSM-002** (Ubiquitous) 모든 메시지 엔벨로프는 A2A Message 데이터 모델에 정합한 코어 필드 — `messageId`, `contextId`, `taskId`, `role`, `parts`, `metadata` — 를 camelCase JSON 키(proto3 JSON 명명 규약)로 운반해야 한다. `parts`의 각 항목은 `kind` 판별자를 가진 `text`|`data` 부분만 허용한다(A2A `raw`|`url` 부분은 이 SPEC에서 의도적으로 제외).

**REQ-CSM-003** (Event-driven) **When** 한 세션이 `session_msg_register`를 `kind`(claude|codex)와 `name`으로 호출하면, 브로커는 A2A AgentCard 참조 형상의 에이전트 기록을 `.moai/state/session-msg/agents/<agentId>.json`에 원자적으로 생성하거나(같은 kind+name이면 하트비트 갱신) 아니면 하고, 호출자가 재사용할 수 있는 안정적 `agentId`를 반환해야 한다.

**REQ-CSM-004** (State-driven) **While** 한 에이전트 기록의 마지막 하트비트가 오프라인 임계(`internal/config/defaults.go` 단일 원천)를 넘지 않았으면, 브로커는 그 에이전트를 온라인으로 보고해야 하고; **When** 임계를 넘으면 오프라인으로 표시해야 한다. `register`·`poll`·`send`(발신자 측) 호출은 하트비트를 갱신해야 한다.

### D.2 배달 의미론 (M1)

**REQ-CSM-005** (Event-driven) **When** 등록된 발신자가 `session_msg_send`로 등록된 수신자에게 메시지를 보내면, 브로커는 송수신자 등록·본문 크기 상한·부분 수 상한을 검증한 뒤 엔벨로프를 수신자 사서함 `mailbox/<agentId>/pending/<messageId>.json`에 원자적으로 적재하고 `messageId`를 반환해야 한다. 수신자가 등록되어 있지 않으면 브로커는 알려진 에이전트 목록을 포함한 구조화 오류를 반환해야 한다.

**REQ-CSM-006** (Event-driven) **When** 등록된 에이전트가 `session_msg_poll`을 호출하면, 브로커는 배치 상한(defaults.go)까지 pending 메시지를 claimed 상태로 원자적으로 클레임(pending→claimed 이동)하여 반환하고, 하트비트를 갱신하고, 선택 인자 `ack_ids`에 담긴 메시지를 claimed에서 삭제하며, 만료된 메시지 수와 잔여 pending 수를 함께 반환해야 한다.

**REQ-CSM-007** (State-driven) **While** claimed 메시지가 확인(ack) 없이 클레임 TTL(defaults.go)을 넘겼으면, 브로커는 다음 스윕에서 그 메시지를 pending으로 환원해야 한다 — 수신 세션이 도중에 죽어도(P4: Codex에는 정리 훅이 없다) 메시지가 소실되지 않는다.

**REQ-CSM-008** (State-driven) **While** pending 또는 claimed 메시지의 수명이 메시지 TTL(defaults.go)을 넘겼으면, 브로커는 그것을 삭제해야 한다(지연 스윕 — 임의의 브로커 호출 시점에 정리).

**REQ-CSM-009** (Ubiquitous) 브로커의 모든 상태 전이는 에이전트별 자문적 록(`internal/session`의 록 패턴 참조)과 `internal/atomicfile` 쓰기로 수행해야 하고, 서로 다른 세션에서 온 동시 `send`·`poll`·`ack` 하에서도 메시지가 분실·중복 손상되지 않아야 한다.

### D.3 MCP 표면과 경계 (M2)

**REQ-CSM-010** (Unwanted) `session_msg_*` 핸들러는 codex 프로세스를 스폰하거나 codex 잡 레코드(`.moai/state/codex-jobs/`)를 만들어서는 안 된다 — 이 SPEC은 살아 있는 세션 사이의 양방향 메시징이지 `codex_task` 패밀리의 일방향 위임이 아니다(§F.1).

**REQ-CSM-011** (Unwanted) 신규 도구의 어떤 코드 경로도 `AskUserQuestion`이나 사용자 대화형 프롬프트를 호출해서는 안 된다(C-HRA-008). 필요한 입력이 없으면 구조화 오류를 반환한다.

**REQ-CSM-012** (Ubiquitous) 세션 메시징의 모든 임계값(TTL·클레임 TTL·오프라인 분기·배치·크기 상한)은 `internal/config/defaults.go`에 단일 원천으로 정의해야 하고, 환경변수가 추가되는 경우 그 이름은 `internal/config/envkeys.go` 상수로 정의해야 한다 — 소스에 인라인 리터럴을 두지 않는다.

**REQ-CSM-013** (Ubiquitous) 4개 신규 도구는 `internal/cli/mcp_server.go` `registerMoaiMCPTools`에 등록됨과 동시에 `internal/mcp` 카탈로그(`MoaiMCPTools`)에 나타나야 하고, 가드 테스트 `TestMoaiMCPServer_RegistrationMatchesCatalog`의 동등성 검증과 콘솔 스키마 자동 파생이 통과해야 한다.

### D.4 교리·문서 (M3)

**REQ-CSM-014** (Ubiquitous) `.claude/rules/moai/workflow/cross-session-messaging.md`는 다음 조항들이 Codex 브로커 경로에도 똑같이 적용됨을 명시하는 확장 절을 가져야 한다: peer-as-user 금지(수신 메시지는 사실이지 사용자 승인이 아니다), 상태 변경 지시 금지(send facts, not mutations), 짧고 자기완결적 메시지, 회신 도착에 의존하지 않는 디스패치(폴 기반이므로 구조적으로 강제), 이 세션이 못 하는 일을 peer에게 맡기지 않는다. 신규 도구의 설명 문자열은 이 규율의 짧은 형태를 실어야 한다 — 그것이 Codex 독자에게 실제로 도달하는 표면이다(P2).

**REQ-CSM-015** (State-driven) **Where** 새 MCP 도구가 노출되면, 그 효력은 moai MCP 서버(세션 소유 서브프로세스, P5)의 재시작 이후에만 발생한다 — 문서(카탈로그 규칙 `moai-mcp-tools.md`)와 e2e 절차는 이 재시작 전제를 명시해야 한다.

## §E. 제약 (Hard Constraints)

1. **MCP 서버 재시작 전제 (P5/t184)** — 새 도구는 서버 재시작 전에 인식되지 않는다. 모든 테스트 절차의 0단계로 명시(REQ-CSM-015, AC-CSM-012·013).
2. **codex_task 경계** — 위임 패밀리와의 별칭 금지(REQ-CSM-010, §F.1).
3. **Template-First** — `.claude/**`·`.moai/** 템플릿 추가물은 `internal/template/templates/`에 미러하고 `make build`로 재생성해야 한다(run-phase M3).
4. **서브에이전트 경계 (C-HRA-008)** — CLI/MCP 코드 경로에 AskUserQuestion 금지(REQ-CSM-011).
5. **하드코딩 방지** — envkeys.go 상수, defaults.go 단일 원천(REQ-CSM-012).
6. **세션 레지스트리 동결** — `active-sessions.json` Entry 스키마(REQ-COORD-024)는 수정하지 않는다(§A.2).
7. **repo-local PR 정책** — 이 저장소는 전 Tier Route B(PR). 브랜치 `WT-codex-session-msg`에서 최종 PR은 main을 목표한다(release/v3.1.3 배치 밖).

## §F. 인접 경계

### §F.1 codex_task 패밀리 (위임) vs session_msg 패밀리 (메시징)

| | `codex_task`/`codex_job_*` | `session_msg_*` (이 SPEC) |
|---|---|---|
| 방향 | 일방향 위임 — 서버가 codex 서브프로세스를 스폰해 잡 실행 | 양방향 — 살아 있는 세션 사이의 메시지 교환 |
| 수명 주체 | 잡 레코드(`.moai/state/codex-jobs/<jobId>.json`, REQ-CX2-003) | 사서함 엔벨로프(`.moai/state/session-msg/`) |
| 프로세스 | codex app-server 서브프로세스를 띄운다 | 어떤 프로세스도 스폰하지 않는다 (REQ-CSM-010) |
| 소유 SPEC | SPEC-CODEX-PHASE2-001 | SPEC-CODEX-SESSION-MSG-001 |

### §F.2 Claude 네이티브 런타임

`SendMessage`/`ListAgents`(Claude Code 내장)는 무관여·무수정. 브로커는 세션 종류 무관 대칭이지만 claude↔claude 문서상 권장은 네이티브.

### §F.3 t88/M4 (미래 통합 지점)

t88이 `.codex/config.toml`에 `mcp_servers.moai`를 자동 등록하면 수동 등록 절차(e2e 1단계)는 사라진다. 이 SPEC은 t88에 의존하지 않는다(P2 실증 경로 사용).

## §G. 교차 참조

- 감사 보고서: `.moai/reports/t187/codex-support-audit.md` (§2.6 MCP, §3 갭 표, §4 테스트 위생)
- 선행 SPEC: SPEC-CODEX-PHASE2-001 (codex_task — 경계 대상), SPEC-V3R6-MULTI-SESSION-COORD-001 (레지스트리 — 동결 대상), SPEC-MCP-CONSOLE-001 (카탈로그/가드 — 확장 대상)
- 교리: `.claude/rules/moai/workflow/cross-session-messaging.md` (M3 확장 대상)
- 카탈로그 규칙: `.claude/rules/moai/core/moai-mcp-tools.md` (21→25 도구 갱신 대상)
- 설계 근거: `design.md` (축 (ii) 채택·거절 사유), `research.md` (A2A 실측 기록·3축 비교)
