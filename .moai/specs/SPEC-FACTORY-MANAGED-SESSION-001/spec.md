---
id: SPEC-FACTORY-MANAGED-SESSION-001
title: "Factory cross-host managed-session layer — PR #1722 재작성 (현 develop 어휘·브로커 API 기준)"
version: "0.1.0"
status: completed
created: 2026-10-01
updated: 2026-10-02
author: GOOS (manager-spec)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tier: L
related_specs: [SPEC-FACTORY-RECORD-001, SPEC-FACTORY-LANE-AUTONOMY-001, SPEC-FACTORY-LANE-JOIN-SOCKET-001]
tags: "factory, managed-session, cross-host, broker, codex, loopback"
---

# SPEC-FACTORY-MANAGED-SESSION-001 — Factory 관리 세션(managed-session) 계층 재작성

## §A. 개요

PR #1722("feat(factory): managed cross-host lead-agent messaging")의 **관리 세션 계층**은 leader↔lane 메시지를 브로커(`internal/factorymsg`)만으로 끝내지 않고, 각 세션을 **소유한 런처**(`moai cc` / `moai glm` / `moai codex`)가 끝까지 배달하는 계층이다. 브로커의 pending 메시지를 폴링해 **메타데이터만** 모델의 다음 턴에 주입하고, 본문은 claim token으로 별도 조회하며, 처리 후 receipt로 확인한다. Codex가 포함된 조합에서는 Codex App Server를 루프백 WebSocket으로 런처가 직접 소유해 무인 수신을 가능하게 한다.

t1365 판정(`.moai/reports/t1365/verdict.md`, 2026-09-30)에 따라 PR 병합은 부적격이다 — 1,713커밋 뒤처짐, 기계 병합 시 30파일 충돌, 어휘가 develop 표준(`lane-<n>`·`-l/--leader`)과 **정반대**(PR은 `agent-1..N` 표준화), 브로커 API가 이미 진행(claim token·launch-pending). 본 SPEC은 PR의 **순수 신규 분(~950행: 관리 세션 코어 2파일 + 테스트 4종)**만 설계 참조로 삼아, **현 develop 어휘와 현 브로커 API 위에서 재작성**한다.

용어 정리 — "cross-host"는 머신 경계가 아니라 **세션을 소유하는 호스트 프로세스(런처·백엔드)의 조합**을 뜻한다(예: GLM lead ↔ Codex lane, Codex lead ↔ Claude lane). 머신 간 네트워크 전송은 범위 밖이다(§F).

## §B. 범위

세 하위 영역을 재구현한다.

1. **관리 Claude/GLM 세션** — 런처가 `--print --input-format stream-json --output-format stream-json`으로 Claude/GLM 프로세스를 소유하고, 연산자 stdin과 브로커 inbox를 하나의 직렬 턴 큐로 합친다.
2. **관리 Codex 세션** — 런처가 로컬 Codex App Server(루프백 WS + capability token)와 TUI를 함께 소유하고, 스레드 ID로 브로커 엔드포인트를 바인딩해 무인 턴 주입을 수행한다.
3. **런처 배선 + MCP 승인 스코핑** — cc/glm/codex_launcher의 factory 진입에 관리 경로를 연결하고, 자동 MoAI MCP 승인은 **런처가 소유한 프로세스에만** `-c` 인수로 전달한다.

## §C. 요구사항 (GEARS)

### C.1 관리 세션 소유와 브로커 바인딩

- **REQ-MS-001** — The managed Codex launcher shall own a local Codex App Server reachable on a loopback WebSocket endpoint guarded by a per-launch random capability token, and shall register the broker endpoint bound to the App Server thread ID and the owner process fingerprint before admitting any message.
- **REQ-MS-002** — **While** a launch is in flight, the launcher shall register a launch-pending broker endpoint and shall either bind it to the started session or roll it back, leaving no permanent launch-pending row when the session fails to start.
- **REQ-MS-003** — **When** a managed session becomes idle and the broker holds unclaimed inbox messages addressed to it, the launcher shall claim a bounded batch under the store's lease discipline and inject only message metadata (message id, claim token, kind, sender slot, task ref) into the session's next turn.
- **REQ-MS-004** — The managed session shall read each claimed message body through the broker's claim-token body-read path, and shall record a broker receipt after processing each claimed message.

### C.2 관리 Claude/GLM 세션

- **REQ-MS-005** — **When** a managed Claude or GLM session launches, the launcher shall force the print and stream-JSON input/output flags onto the session process and shall refuse operator arguments that would take over those flags.
- **REQ-MS-006** — The managed Claude/GLM launcher shall serialize operator stdin lines and claimed broker inbox messages into one ordered turn queue gated on the previous turn's completion, so that operator input and broker delivery never race.

### C.3 어휘와 브로커 적합성

- **REQ-MS-007** — The managed layer shall use only the canonical factory vocabulary — lane joins with `-f lane` / `-f lane-<n>`, leader discovery with `-l/--leader`, leader labels `leader`, `leader-<n>`, `leader-<run-id>` — and shall not reintroduce PR #1722's `agent-<n>` / `worker-<n>` naming surface.
- **REQ-MS-008** — The managed layer shall consume the current `internal/factorymsg` store API (RegisterPeer / LaunchPending bind-rollback / Claim with lease / ReadBody with claim token / Receipt) without modifying `store.go`; where the API proves insufficient, the run phase shall return a blocker report instead of patching the store.

### C.4 MCP 승인 스코핑

- **REQ-MS-009** — The managed Codex launcher shall pass automatic MoAI MCP approval arguments only to the App Server and TUI processes it owns, and shall not change the project-level capability-based approval mode (`mcpApprovalMode = "writes"`).
- **REQ-MS-010** — **When** doctor inspects a project whose Codex config carries a project-global MoAI MCP approval override left by an older Factory generation, it shall warn about the stale global approval setting and shall not rewrite the user-owned config file.
- **REQ-MS-011** — **Where** a broker run is Claude-only, the broker send tool shall continue to refuse `factory_msg_send` so that Claude sessions keep the native SendMessage policy.

### C.5 플랫폼·검증 자세

- **REQ-MS-012** — **Where** the host is Windows, the managed layer shall operate through the child-process ownership model with zero `syscall` usage in newly added files, and the `GOOS=windows GOARCH=amd64 go build` cross build shall pass.
- **REQ-MS-013** — The managed layer shall be verifiable on a single host: the acceptance suite shall exercise broker claim, metadata injection, body read, and receipt through loopback tests that require no real second host and no real network.
- **REQ-MS-014** — The managed layer shall not perform card completion judgment, merge automation, controller re-alerting, or any merge-window machinery — those surfaces remain SPEC-FACTORY-CONTROLLER-001's (card t1241 F3).
- **REQ-MS-015** — The managed layer shall treat a broker receipt as delivery evidence only; card disposition remains message-body work judged by the receiving session, never inferred from the receipt itself.

### C.6 생략된 것 (비-목표 요약)

PR의 gorilla/websocket 기반 전송은 채택하되 App Server 클라이언트 한정이다(§ design.md D-2). 머신 간 전송, 사용자 설정 자동 이행, Windows 런타임 실측 인증은 §F에서 제외한다.

## §D. 제약

- **어휘**: `lane-<n>` · `-l/--leader` 만 사용. `internal/cli/factory.go`의 레거시 토큰 거부 경로는 회귀시키지 않는다.
- **브로커**: `internal/factorymsg/store.go` 무수정이 원칙 — API 부족 시 blocker.
- **의존성**: 신규 의존은 `github.com/gorilla/websocket` v1.5.3 하나(사유와 기각 대안은 design.md D-2).
- **F3 비침범**: 병합 자동화·Decider·lease reaping·재알림·T29b/T29c 핸드오버 발행은 전부 금지.
- **문서 언어**: 문서·안내는 한국어(`language.yaml` documentation: ko), 코드 주석·커밋은 영어.
- **진행 조건**: plan-audit `--deep` 통과가 run 착수 조건이며, 카드 절차는 레인 내 연속(plan → run → sync)으로 수행한다.

## §E. 성공 기준 요약

`acceptance.md`의 AC-MS-001..017 전부 통과. 특히: (1) F3 비침범 grep이 0적중, (2) 어휘 grep이 `lane-<n>` 이외 라벨 생성 0적확인, (3) `GOOS=windows` 크로스빌드 + 신규 파일 zero-syscall, (4) 실제 제2호스트 없는 loopback 통합 테스트가 CI에서 녹색. 관리 계층 착지 후에도 기존 factory 어휘·조인·디스패치 테스트(`internal/cli`·`internal/kanban`·`internal/hook`의 Factory 스위트)가 회귀 없이 통과한다.

## §F. Exclusions

### Out of Scope — F3 controller machinery (card t1241, SPEC-FACTORY-CONTROLLER-001)

- No merge-window automation (autonomous merge on sync-audit PASS + no-conflict + tree identity), no `Decider` interface (`human|llm|jev`), no lease-expiry reaping, no stalled-card re-alert, and no `T29b`/`T29c` hand-over event emission from this layer — the managed layer delivers messages and records receipts; the controller owns card disposition.

### Out of Scope — PR #1722 vocabulary rewrite

- No `agent-<n>` standardization, no `worker`-token revival, and none of the 30-file conflict surface the PR's vocabulary changes would reopen — `factory.go`, `internal/kanban`, `internal/hook` i18n, and the rule documents already carry the canonical `lane-<n>` direction on develop.

### Out of Scope — machine-boundary transport

- No network transport between distinct machines and no remote broker synchronization — the broker stays a SQLite store under the shared `projectRoot` (`factorymsg.BrokerPath`); "cross-host" in this SPEC names distinct launcher/backend process ownership on one machine.

### Out of Scope — Windows native runtime certification

- The `GOOS=windows` cross build and unit/loopback tests are in scope; running live managed sessions on Windows native or WSL 2 is deferred — the PR measured macOS only, and this SPEC keeps that gap documented rather than claimed closed.

### Out of Scope — user-owned config migration

- No automatic migration of an existing project's `.codex/config.toml` MoAI MCP approval settings; doctor warns, the user-owned file stays untouched (carries PR report's own limitation forward).

### Out of Scope — completion judgment from receipts

- A receipt proves delivery, not completion; no card state transition may be driven by receipt counts (the PR report's own judgment-scope rule, adopted verbatim).

### Out of Scope — Anthropic-backend session re-verification

- The PR's live evidence covers Codex↔Codex and GLM↔Codex on macOS; re-verifying the Anthropic-account backend path is a follow-up, not this SPEC's gate.

## §G. 교차 참조

- `plan.md` — 마일스톤(M1..M6), 사전 점검, 운영자 게이트 자세.
- `acceptance.md` — AC-MS-001..017, 에지 케이스, 품질 게이트, DoD.
- `design.md` — D-1..D-7 설계 결정(F3 경계 인터페이스 테이블 포함)과 기각 대안.
- `research.md` — 정찰 R1..R14(PR 브랜치 파일, develop 어휘·브로커 API, 충돌 맵) 인용.
- `.moai/reports/t1365/verdict.md` — 재작성 경로 판정(병합 부적격 + 관리 세션 가치 인정).
- SPEC-FACTORY-RECORD-001 (F1 카드 기록층, T29b/T29c amendment c6fab1eab) — F3 경계의 상대면.
- SPEC-FACTORY-LANE-AUTONOMY-001 — umbrella exclusion 선행 사례(§F 표기 준거).
