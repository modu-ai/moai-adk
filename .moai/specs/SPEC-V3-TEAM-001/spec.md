---
id: SPEC-V3-TEAM-001
title: "Team Mailbox Protocol v2 (10 structured message types)"
version: "0.1.0"
status: draft
created: 2026-04-22
updated: 2026-04-22
author: GOOS
priority: P1 High
phase: "v3.0.0 — Phase 6a Tier 2 Strategic"
module: "internal/team/mailbox/"
dependencies:
  - SPEC-V3-SCH-001
related_gap:
  - gm#71
  - gm#72
  - gm#75
  - gm#76
related_theme: "Theme 8 — Team Protocol v2 (Mailbox + In-process)"
breaking: true
bc_id: BC-008
lifecycle: spec-anchored
tags: "team, mailbox, message-schema, shutdown, plan-approval, permission, v3"
---

# SPEC-V3-TEAM-001: Team Mailbox Protocol v2 (10 structured message types)

## HISTORY

| Version | Date       | Author | Description                                              |
|---------|------------|--------|----------------------------------------------------------|
| 0.1.0   | 2026-04-22 | GOOS   | Initial draft (Wave 4 SPEC writer)                       |

---

## 1. Goal (목적)

moai team mode의 mailbox(SendMessage / recv) 페이로드를 ad-hoc JSON에서 **Zod-equivalent 구조화 스키마**로 전환하여 silent payload-shape drift를 방지하고, CC의 10개 typed message(findings-wave1-agent-team.md §6.8, utils/teammateMailbox.ts:682-860)와 동등 수준의 계약을 제공한다. Go 구현에서는 `encoding/json` 태그 + `go-playground/validator/v10` 태그로 Zod와 동등한 검증 효과를 얻으며, 새로운 의존성은 추가하지 않는다.

### 1.1 배경

현재 moai team mode는 `TeamCreate` / `SendMessage` 페이로드가 untyped(findings-wave1-moai-current.md 관련, 본 문서 내 memory 인용). CC는 `utils/teammateMailbox.ts`에서 10+ 타입의 Zod 스키마(findings-wave1-agent-team.md §6.8)를 정의하여:

- shutdown_request / shutdown_approved / shutdown_rejected — 팀 종료 합의 프로토콜
- plan_approval_request / plan_approval_response — teammate가 plan을 leader에 제출 → leader 승인/거절
- permission_request / permission_response — worker tool permission flow
- sandbox_permission_request / sandbox_permission_response — network host allowlisting
- task_assignment — Task tool에 의해 자동 전송되는 owner change 알림

moai가 이 계약을 채택하지 않으면 team mode에서 다음 클래스의 버그가 반복 발생:
- teammate가 shutdown_request에 응답하지 않아 leader 무한 대기
- plan approval feedback 누락으로 teammate가 stale plan 실행
- permission_request 브로드캐스트로 인한 race condition

본 SPEC은 10개 message type의 Go 구조체 + validator/v10 검증 + discriminated union decoder를 도입한다.

### 1.2 Non-Goals

- In-process teammate backend 구현 (master §9 Open Question #7: v3.1로 defer)
- iTerm2 native-split backend (master Non-Goals)
- MEMORY.md 또는 agent memory와의 통합 (SPEC-V3-MEM-001 범위)
- UDS(unix domain socket) cross-session messaging (CC는 지원하지만 moai Non-Goal)
- Bridge transport (Anthropic-cloud specific; Non-Goal)
- 메시지 encryption (scope out; team config 파일은 local trust)
- Zod 포팅(Go-native validator/v10로 기능적 등가)

---

## 2. Scope (범위)

### 2.1 In Scope

- `internal/team/mailbox/types.go`: 10개 message struct + `Message` interface (`Type()`, `RequestID()`)
- `internal/team/mailbox/registry.go`: message type string → factory function 매핑
- `internal/team/mailbox/decoder.go`: discriminated union decoder (raw JSON → concrete Message type)
- `internal/team/mailbox/validate.go`: validator/v10 검증 wrapper
- `internal/team/mailbox/legacy.go`: v2 ad-hoc JSON을 `LegacyMessage{RawBytes}`로 wrap하는 backward-compat shim (v3.0 warn; v3.2 remove per BC-008)
- `internal/team/mailbox/capped_history.go`: `TEAMMATE_MESSAGES_UI_CAP = 50` 적용 (findings-wave1-agent-team.md §7.2)
- `internal/team/mailbox/approval/plan.go`: plan-approval flow 통합(teammate → leader)
- `SendMessage` 엔트리포인트 strict mode 추가: invalid message → reject
- 10개 message type schema 문서: `.claude/rules/moai/workflow/team-mailbox-protocol.md`

### 2.2 Out of Scope (Exclusions — What NOT to Build)

- In-process teammate backend (v3.1+; master §9 Open Q #7)
- 5개 backend 중 tmux / iterm2 / in-process detection priority (v3.1)
- Word-slug team-name collision via `generateWordSlug()` (v3.1+)
- Session cleanup registry의 SIGINT 자동 정리 강화(현재 moai의 tmux pane cleanup은 유지)
- MCP sandbox permission의 network host 허용 로직 자체(request/response 스키마만 정의; 실제 host allowlist 집행은 SPEC-V3-SCH-002 범위)
- UDS socket transport
- Bridge cross-machine transport
- Message encryption
- Teammate mailbox GC (old messages auto-delete)

---

## 3. Environment (환경)

- 런타임: moai-adk-go v3.0.0+ (Go 1.23+)
- Claude Code v2.1.111+
- 의존: SPEC-V3-SCH-001 (`go-playground/validator/v10`)
- 영향 디렉터리: `internal/team/mailbox/`, `internal/team/approval/`
- team mode 전제: `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` + `workflow.team.enabled: true`
- 영향 받는 호출 경로: `moai glm`, `moai cg`, 기존 `/moai` workflow의 team mode 옵션
- OS 동등성: macOS / Linux / Windows
- 참조 CC 파일: `utils/teammateMailbox.ts:682-860`, `tools/SendMessageTool/SendMessageTool.ts:67-917`

---

## 4. Assumptions (가정)

- SPEC-V3-SCH-001이 먼저 머지되어 validator/v10이 직접 의존성으로 추가되어 있다.
- 기존 team mode 사용자는 minority이며 v3.0 dual-parse window(1 minor version) 동안 wrapper를 갱신할 수 있다.
- Request-response correlation은 UUID v4 `RequestID`로 충분하다(CC 동일).
- `TEAMMATE_MESSAGES_UI_CAP = 50`은 BigQuery 분석 기반 검증값(findings-wave1-agent-team.md §7.2: 292-agent whale session reached 36.8GB).
- Legacy JSON 메시지는 식별 가능한 `type` 필드가 없거나 v3에서 미지원인 type을 가진다.
- Team mailbox 저장소(`~/.claude/teams/{team_name}/inboxes/`)는 read/write 가능 상태다.

---

## 5. Requirements (EARS 요구사항)

### 5.1 Ubiquitous (항시)

**REQ-TEAM-001-001 (Ubiquitous) — Message 인터페이스**
The `internal/team/mailbox/types.go` **shall** define a `Message` interface with methods `Type() string` (returns discriminator) and `RequestID() string` (returns UUID for request/response pairing).

**REQ-TEAM-001-002 (Ubiquitous) — 10 concrete message types**
The package **shall** define the following 10 concrete struct types, each implementing `Message`:
1. `ShutdownRequest` — `{type, request_id, team_name, initiator, reason?}`
2. `ShutdownApproved` — `{type, request_id, approver, pane_id?, backend_type?}`
3. `ShutdownRejected` — `{type, request_id, from, reason}`
4. `PlanApprovalRequest` — `{type, request_id, teammate_name, plan_path, summary}`
5. `PlanApprovalResponse` — `{type, request_id, approved, feedback, permission_mode?}`
6. `PermissionRequest` — `{type, request_id, teammate_name, tool_name, tool_input}`
7. `PermissionResponse` — `{type, request_id, behavior, updated_input?, message?}`
8. `SandboxPermissionRequest` — `{type, request_id, teammate_name, network_host}`
9. `SandboxPermissionResponse` — `{type, request_id, allow}`
10. `TaskAssignment` — `{type, request_id, task_id, owner, subject}`

**REQ-TEAM-001-003 (Ubiquitous) — validator tag**
Every concrete message struct **shall** declare validator/v10 tags enforcing at minimum: `RequestID` required + `uuid4`, `type` literal match, non-empty for all required string fields.

**REQ-TEAM-001-004 (Ubiquitous) — registry factory**
The `internal/team/mailbox/registry.go` **shall** define a map `registry map[string]func() Message` enumerating all 10 type strings to concrete type constructors.

**REQ-TEAM-001-005 (Ubiquitous) — message size cap**
The mailbox storage layer **shall** retain at most `TEAMMATE_MESSAGES_UI_CAP = 50` recent messages per teammate inbox, evicting oldest on overflow.

### 5.2 Event-Driven (이벤트 기반)

**REQ-TEAM-001-006 (Event-Driven) — decode entry point**
**When** `SendMessage` receives a raw JSON payload, the decoder **shall** execute the following sequence:
1. Unmarshal into `{"type": string}` envelope
2. Look up factory in registry
3. If not found → return `LegacyMessage{RawBytes}` (v3.0/v3.1 only)
4. If found → unmarshal full payload → run validator → return typed Message or validation error

**REQ-TEAM-001-007 (Event-Driven) — plan approval flow**
**When** a teammate sends a `PlanApprovalRequest` to the team leader, the leader's mailbox handler **shall**:
1. Persist the request to `~/.claude/teams/{team}/inboxes/team-lead.json`
2. Surface the request to the orchestrator via existing `AskUserQuestion` path
3. Emit `PlanApprovalResponse` (approved OR rejected with feedback) back to the teammate's mailbox

**REQ-TEAM-001-008 (Event-Driven) — shutdown request propagation**
**When** `ShutdownRequest` is received by a teammate, the teammate runtime **shall** reply with either `ShutdownApproved` (if work is complete) or `ShutdownRejected` (with non-empty `reason`) within 30 seconds.

**REQ-TEAM-001-009 (Event-Driven) — task assignment auto-emit**
**When** the Task tool updates an `owner` field on any task, the Task layer **shall** auto-emit a `TaskAssignment` message to the new owner's mailbox (parity with CC TaskUpdateTool.ts — findings §6.5).

### 5.3 State-Driven (상태 기반)

**REQ-TEAM-001-010 (State-Driven) — strict mode validation**
**While** `workflow.team.mailbox_strict: true` in `.moai/config/sections/workflow.yaml`, the decoder **shall** reject any message that fails validator/v10 validation and log the rejection to `.moai/logs/team-mailbox-rejects.log`.

**REQ-TEAM-001-011 (State-Driven) — legacy warning window**
**While** v3.0 or v3.1 is running, decoding a legacy ad-hoc JSON message (unrecognized `type` or missing `type`) **shall** emit a one-time-per-session warning logged with the raw payload's first 200 characters and a link to migration guide.

**REQ-TEAM-001-012 (State-Driven) — message cap eviction order**
**While** a teammate inbox holds `TEAMMATE_MESSAGES_UI_CAP` messages, receiving a new message **shall** evict the oldest unread message first; if all messages are unread, the oldest message by timestamp is evicted.

### 5.4 Optional (선택)

**REQ-TEAM-001-013 (Optional) — custom validators**
**Where** a project defines custom field validators via `.moai/config/sections/workflow.yaml` under `workflow.team.mailbox_custom_validators`, the decoder **shall** merge the custom validators with the default set before validation.

**REQ-TEAM-001-014 (Optional) — telemetry emission**
**Where** `workflow.team.mailbox_telemetry: true`, each successful decode **shall** emit a `tengu_team_message_decoded` event (local log file only — no network) with `type`, `duration_ms`, and `valid`.

### 5.5 Unwanted Behavior

**REQ-TEAM-001-015 (Unwanted Behavior) — shutdown_response without reason on reject**
**If** a `ShutdownRejected` message is received with an empty `reason` field, **then** the decoder **shall** reject the message with error `TEAM_INVALID_SHUTDOWN_REJECT_REASON` (parity with CC SendMessageTool.ts:604-718 validation rule).

**REQ-TEAM-001-016 (Unwanted Behavior) — broadcast of structured message**
**If** a structured message (non-string, any of the 10 types) is sent to `"*"` broadcast target, **then** the SendMessage entry point **shall** reject the send with error `TEAM_STRUCTURED_BROADCAST_FORBIDDEN` (parity with CC SendMessageTool rule).

**REQ-TEAM-001-017 (Unwanted Behavior) — cross-team shutdown_response**
**If** a `ShutdownApproved` or `ShutdownRejected` is sent to any target other than `"team-lead"`, **then** the SendMessage entry point **shall** reject the send with error `TEAM_SHUTDOWN_WRONG_TARGET` (parity with CC SendMessageTool.ts:604-718).

**REQ-TEAM-001-018 (Unwanted Behavior) — v3.2 strict mode preview**
**If** the environment variable `MOAI_TEAM_MAILBOX_STRICT_PREVIEW=1` is set, **then** the decoder **shall** treat legacy messages as hard errors (simulating v3.2 behavior). This is for testing only; default behavior remains warn-only in v3.0.

### 5.6 Complex (복합)

**REQ-TEAM-001-019 (Complex) — plan approval feedback permission inheritance**
**While** a `PlanApprovalResponse` carries a `permission_mode` field AND `approved: true`, **when** the teammate receives the response, the teammate **shall** adopt the specified `permission_mode` for subsequent tool calls until the next plan submission (parity with CC SendMessageTool.ts:448-457).

---

## 6. Acceptance Criteria (수용 기준 요약)

상세 Given-When-Then 시나리오는 `acceptance.md`에서 정의한다.

핵심 기준:

- **AC-TEAM-001-01**: 10개 message struct 모두 `Type()`과 `RequestID()` 메서드 구현 확인
- **AC-TEAM-001-02**: Invalid JSON (empty `request_id`) → validator/v10 reject, 정확한 field name 포함 에러 메시지
- **AC-TEAM-001-03**: Unknown `type` string → `LegacyMessage{RawBytes}` 반환 + warning 로그 1회
- **AC-TEAM-001-04**: `PlanApprovalRequest` → leader mailbox 기록 → `AskUserQuestion` 호출 → `PlanApprovalResponse` 송신 end-to-end 통합 테스트
- **AC-TEAM-001-05**: `ShutdownRejected` with empty `reason` → `TEAM_INVALID_SHUTDOWN_REJECT_REASON`
- **AC-TEAM-001-06**: Structured message broadcast to `"*"` → `TEAM_STRUCTURED_BROADCAST_FORBIDDEN`
- **AC-TEAM-001-07**: 50개 메시지 보유 inbox에 51번째 수신 → oldest unread 제거 확인
- **AC-TEAM-001-08**: `MOAI_TEAM_MAILBOX_STRICT_PREVIEW=1` + legacy message → hard error 반환
- **AC-TEAM-001-09**: `PlanApprovalResponse(approved=true, permission_mode=acceptEdits)` → teammate가 `acceptEdits`로 전환 확인
- **AC-TEAM-001-10**: `go test ./internal/team/mailbox/... -race` 전체 통과, coverage ≥ 85%

---

## 7. Constraints (제약)

- [HARD] Zero net new dependencies: `encoding/json` + validator/v10만 사용.
- [HARD] 10 message types는 본 SPEC에서 고정. 추가 타입은 별도 SPEC (CC parity 유지).
- [HARD] BC-008: legacy JSON은 v3.0에서 warn, v3.2에서 제거.
- [HARD] Strict mode default false in v3.0(opt-in via workflow.yaml). v3.2에서 default true.
- [HARD] `TEAMMATE_MESSAGES_UI_CAP = 50`은 HARD constant; 완화 시 36GB whale session risk.
- [HARD] 하드코딩 금지: cap, timeout 등은 `internal/team/mailbox/constants.go`에 const로 정의.
- [HARD] Message size limit: 각 message JSON ≤ 64 KB (prevent pathological payload; Master §3.8).
- [HARD] broadcast(`"*"`) + structured message 조합 금지 (CC parity).
- [HARD] shutdown_response는 항상 `team-lead` 대상.
- [HARD] 16개 언어 중립성: team mode는 moai tool 자체 기능이며 사용자 프로젝트 언어와 무관.

---

## 8. Risks & Mitigations (리스크 및 완화)

| 리스크                                                            | 영향     | 완화                                                                                              |
|-------------------------------------------------------------------|----------|---------------------------------------------------------------------------------------------------|
| Legacy JSON 사용 기존 사용자 breakage                              | High     | BC-008 1-minor-version dual-parse; warning with migration link; v3.2에서만 strict (Master §4)     |
| Message validation overhead로 team latency 증가                    | Medium   | validator/v10은 reflection-based이나 typical message ≤ 64 KB → ~100 µs overhead; 측정 후 benchmark  |
| 10개 type 부족으로 사용자 커스텀 type 요청                         | Medium   | REQ-TEAM-001-013 custom validator hook; 추가 type은 별도 SPEC                                     |
| `PlanApprovalResponse`의 `permission_mode` 필드 오용                | Medium   | REQ-TEAM-001-019 Complex 요구사항; validator로 enum 강제; doc 예제 포함                            |
| Mailbox GC 부재로 inbox 파일 비대화                                | Low      | REQ-TEAM-001-005 50-cap이 implicit GC 역할; 하드 GC는 scope out                                    |
| Race condition (concurrent write to same inbox)                    | Low      | CC 방식 file-lock 적용; `proper-lockfile` 대응 Go 구현은 `github.com/gofrs/flock`(신규 의존 금지)  |
| 실수로 broadcast + structured 조합 시도                            | Low      | REQ-TEAM-001-016 reject; clear error message                                                     |
| shutdown 30-second timeout이 실 환경에서 부족                      | Low      | env var `MOAI_TEAM_SHUTDOWN_TIMEOUT_SECONDS`로 override                                          |
| 메시지 deserialize 시 타임존 drift                                 | Very Low | 모든 timestamp는 ISO-8601 UTC; validator `datetime` 태그                                          |

Race condition 완화 상세: Go 구현은 `os.O_EXCL` + atomic rename 패턴을 사용하여 새로운 의존성 없이 mutual exclusion 달성. file-lock 라이브러리 추가 금지(9-deps 원칙 유지).

---

## 9. Dependencies (의존성)

### 9.1 Blocked by

- **SPEC-V3-SCH-001** (Formal config schemas): validator/v10 도입이 선행 조건.

### 9.2 Blocks

- (없음) — v3.0 범위에서 team mailbox는 독립적.

### 9.3 Related

- **SPEC-V3-AGT-002** (permissionMode extensions): `PermissionResponse`의 `behavior` field 값은 `permissionMode`와 일관된 enum(`allow`/`deny`).
- **SPEC-HOOKS-006** (sibling writer; Hook permission decision): hook이 발생시키는 `PermissionResponse`와 본 SPEC의 mailbox `PermissionResponse`는 별도 경로이나 동일 enum 공유.
- **SPEC-V3-MEM-001** (Memory 2.0): team memory `~/.claude/teams/{team}/team/MEMORY.md` 경로 validation은 MEM-001 rules 재사용.
- **SPEC-V3-SCH-002** (sibling writer; Settings source layering): `workflow.yaml.team.*` 필드가 3-tier layering을 준수.

---

## 10. Traceability (추적성)

- 총 REQ 개수: 19 (Ubiquitous 5, Event-Driven 4, State-Driven 3, Optional 2, Unwanted Behavior 4, Complex 1)
- 예상 AC 개수: 10
- 관련 Wave 1 근거:
  - findings-wave1-agent-team.md §6.4 (SendMessage routing; structured message types; SendMessageTool.ts:604-718 validation rules)
  - findings-wave1-agent-team.md §6.8 (Teammate Mailbox System; utils/teammateMailbox.ts:682-860; 10 Zod schemas)
  - findings-wave1-agent-team.md §7.2 (TEAMMATE_MESSAGES_UI_CAP=50; 36GB whale session rationale)
  - findings-wave1-agent-team.md §8.1 (Gap: teammateMailbox schemas; plan-approval flow; TEAMMATE_MESSAGES_UI_CAP; word-slug; session cleanup)
  - findings-wave1-agent-team.md §8.2 (StructuredMessage discriminated union)
  - master-v3 §3.8 (Theme 8; design approach; API sketch)
  - master-v3 §4 BC-008 (breaking change catalog)
- 구현 시 각 소스 파일에 `@SPEC:SPEC-V3-TEAM-001:REQ-TEAM-001-NNN` 주석 부착
- 코드 구현 예상 경로:
  - `internal/team/mailbox/types.go` (REQ-TEAM-001-001, 002, 003)
  - `internal/team/mailbox/registry.go` (REQ-TEAM-001-004)
  - `internal/team/mailbox/decoder.go` (REQ-TEAM-001-006, 010, 011, 015, 016, 017, 018)
  - `internal/team/mailbox/validate.go` (REQ-TEAM-001-003, 010)
  - `internal/team/mailbox/legacy.go` (REQ-TEAM-001-011, 018)
  - `internal/team/mailbox/capped_history.go` (REQ-TEAM-001-005, 012)
  - `internal/team/approval/plan.go` (REQ-TEAM-001-007, 019)
  - `internal/team/mailbox/shutdown.go` (REQ-TEAM-001-008, 015, 017)
  - `internal/team/mailbox/task_assignment.go` (REQ-TEAM-001-009)
  - `internal/team/mailbox/constants.go` (TEAMMATE_MESSAGES_UI_CAP, timeouts)
  - `internal/team/mailbox/*_test.go` (AC-TEAM-001-01..10)
  - `.claude/rules/moai/workflow/team-mailbox-protocol.md` (문서)

---

End of SPEC.
