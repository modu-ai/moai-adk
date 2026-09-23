---
id: SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001
title: "Factory lane card worktree handoff"
version: "0.4.0"
status: draft
created: 2026-09-22
updated: 2026-09-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/factorymsg"
lifecycle: spec-anchored
tags: "factory,lane,worktree,codex,app-server,rebind"
tier: L
depends_on:
  - SPEC-FACTORY-MIXED-HOOK-001
card: t1082
---

# SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.4.0 | 2026-09-23 | t1074 착지분의 세 번째 slot 쓰기 경로인 UserPromptSubmit `RegisterPeer`를 모델링해, 미완료 handoff 동안 그 경로가 같은 broker transaction 안에서 거부되도록 규정했다(REQ-FLH-018). |
| 0.3.0 | 2026-09-23 | t1074 착지(develop `861510fb6`) 후 전제 재검증 결과를 반영해 launch-pending source endpoint의 handoff 거부(REQ-FLH-016)와 launcher provisional bind 대 handoff rebind의 직렬화·순서(REQ-FLH-017)를 추가했다. |
| 0.2.0 | 2026-09-22 | 관측된 Codex 0.155.1 경계를 반영해 interactive는 다음 정상 turn의 SessionStart, headless는 공식 app-server 반환 thread ID를 mode별 결합 증거로 분리했다. |
| 0.1.0 | 2026-09-22 | 카드 t1082의 안정 lane → 전용 card worktree 전환 및 generation-safe endpoint rebind 계약을 최초 정의했다. |

## WHY

혼합 factory의 안정 논리 lane은 카드를 배차받은 뒤에도 기존 checkout과 물리 session/thread endpoint에 머물 수 있다. 그 상태에서 곧바로 dispatch body를 처리하면 잘못된 cwd 수정, primary checkout의 branch 전환, stale generation ACK, worktree base drift가 섞인다. t1074가 만든 안정 lane 주소와 current endpoint seam을 유지하면서, 카드 전용 L1 worktree와 새 물리 endpoint를 검증된 순서로 결합해야 한다.

## WHAT

이 SPEC은 카드 배차 뒤 다음 순서를 하나의 내구성 있는 handoff 계약으로 정의한다.

`reserve → local develop HEAD pin → MoAI launcher L1 worktree create → WT-<slug> + card traceability → mode별 SWITCH_PENDING → interactive의 다음 정상 turn SessionStart 또는 headless의 공식 RPC 결과 검증 → atomic rebind → BOUND receipt → dispatch body`

논리 lane ID는 재배차와 재시작에도 유지되는 주소다. Codex thread/session UUID와 generation은 handoff마다 바뀔 수 있는 물리 endpoint다. interactive 경로와 headless 경로는 trigger와 실패 의미를 분리하지만, 같은 reservation nonce, provenance 검증, atomic rebind, stale-endpoint 차단 규칙으로 합류한다.

## HOW

- t1074의 `internal/factorymsg` broker/peer generation/receipt와 `internal/homestate` canonical project/run identity를 확장한다.
- worktree는 기존 MoAI launcher의 L1 materializer만 사용한다. raw `git worktree add` 경로를 새로 만들지 않는다.
- headless Codex는 기존 app-server JSON-RPC client를 확장하여 공식 `thread/fork` 또는 `thread/start`의 target `cwd` 요청과 반환 thread ID를 사용한다. 결합을 위해 SessionStart나 빈 `turn/start`를 만들지 않는다.
- interactive Codex는 idle 상태에서 사용자가 직접 `/cd`를 실행하도록 안내한다. hook, MCP, tmux `send-keys`, private socket, 모델에게 보내는 “cd 해라” 메시지는 relocation 수단이 아니다.
- rebind, old endpoint tombstone, durable `BOUND` receipt, pending dispatch release를 기존 factory broker의 한 transaction에 둔다. 새 broker나 daemon은 추가하지 않는다.

## User story

factory lead 운영자로서 안정 lane에 카드를 배차한 뒤 그 lane의 대화 연속성을 잃지 않고 카드 전용 worktree로 옮기고 싶다. 새 endpoint가 올바른 cwd와 branch에 실제로 결합되기 전에는 어떤 구현도 시작되지 않아야 하고, 이전 endpoint나 중복 dispatch가 뒤늦게 나타나도 현재 endpoint를 덮어쓰거나 ACK하지 못해야 한다.

## Scope and measured baseline

- 카드: `t1082`, Tier L.
- 대상 worktree/branch: `.claude/worktrees/t1082`, `WT-factory-lane-worktree-handoff`.
- 현재 plan 기준 HEAD: `bf39a539d97f49edf3b11517ee7c982239c60df3`.
- 이 worktree는 `moai worktree new t1082` 실행 시 primary `main@2213871af`에서 잘못 시작했고, 이후 `git merge --ff-only develop`로 `develop@3f3ffbb57`에 맞춰졌다. 이 관측된 creation-base drift는 REQ-FLH-004와 AC-FLH-016의 필수 실패 사례다.
- t1074 의존 merge는 `bf39a539d`, 의존 commit `8c5d9be99`는 현재 HEAD의 ancestor다.
- 현재 MCP catalog는 단위 검증에서 36개(쓰기 14, 읽기 22)로 관측됐다. 이 SPEC은 새 MCP tool을 기본 해법으로 추가하지 않는다.

## Requirements (GEARS)

### REQ-FLH-001 — Stable lane, replaceable endpoint

While a factory handoff is active, the handoff SHALL keep the stable logical lane ID as the address, treat the thread/session UUID and generation as a replaceable physical endpoint, and scope every handoff state, receipt, and dispatch to that lane within one canonical project key and run ID.

### REQ-FLH-002 — Separate interactive and headless state machines

Where the lane is interactive Codex, the handoff SHALL allow only a user-driven `/cd` relocation followed by the user's next normal turn. Where the lane is headless Codex, the handoff SHALL allow only the official app-server `thread/fork` or `thread/start` with the reserved target `cwd`. The two paths SHALL use distinct transitions, evidence, and error codes and converge only at the validated atomic rebind after their mode-specific `SWITCH_PENDING` state.

### REQ-FLH-003 — Fail-closed admission

When the source lane has an active turn, a permission wait, an interrupt in progress, a dirty source checkout, a stale reservation, an untrusted cwd, a branch owned by another worktree, a conflicting target path, or another unfinished handoff for the same lane, the handoff SHALL return a reason-specific NACK without creating a worktree, relocating the endpoint, or releasing a dispatch.

### REQ-FLH-004 — Develop pin and creation-base integrity

When a lane reserves a handoff, the handoff SHALL record the canonical primary checkout's local `develop` HEAD as an immutable base pin before using the existing MoAI launcher L1 materializer to create `<primary>/.claude/worktrees/<card-id>`. When the created worktree HEAD differs from the pin or the base moves during creation, the handoff SHALL NACK with `BASE_DRIFT` and emit neither `BOUND` nor a dispatch body.

### REQ-FLH-005 — Card, branch, commit, and report traceability

While a handoff is in progress, the handoff SHALL require the card ID in the worktree directory, a collision-free `WT-<descriptive-slug>` branch, the card and SPEC IDs in the dispatch envelope, the card ID in future commit subjects, and `.moai/reports/<card-id>/verdict.md` as the evidence path. It SHALL NOT enter `SWITCH_PENDING` before branch rename and collision validation complete.

### REQ-FLH-006 — Interactive relocation boundary

While an interactive lane is idle and has no permission request, the handoff SHALL emit only a user-executed `/cd <target>` instruction containing the absolute target and reservation nonce and remain `SWITCH_PENDING_INTERACTIVE` until SessionStart evidence arrives from the user's next normal turn. It SHALL NOT assume a hook or MCP tool can execute a slash command, use tmux `send-keys`, a private TUI socket, or a model prompt to change cwd, or create an empty model turn to obtain evidence.

### REQ-FLH-007 — Headless official relocation

While a headless source thread is idle, the handoff SHALL prefer a history-preserving `thread/fork` with the reserved target `cwd` through the existing app-server client and verify the returned new `thread.id`, `forkedFromId`, and `thread/started`. It SHALL use `thread/start(cwd)` only when no source thread exists, SHALL NACK relocation during an active turn, and SHALL NOT wait for SessionStart, issue an empty `turn/start`, or use `turn/steer` to manufacture binding evidence.

### REQ-FLH-008 — Atomic validated rebind and BOUND receipt

When mode-specific endpoint evidence arrives, the handoff SHALL validate the reservation nonce, card and SPEC, canonical target cwd, exact worktree root, pinned HEAD, current branch, new session/thread UUID, new generation, PID, and process-start identity. Interactive evidence SHALL be SessionStart from the user's next normal turn after `/cd`; headless evidence SHALL be the official `thread/fork` or `thread/start` result returned to the controller together with controller readback of target cwd/branch/HEAD. Only after every validation succeeds SHALL one broker transaction replace the current endpoint, tombstone the old endpoint, mark the handoff `BOUND`, and record a durable `BOUND` receipt.

### REQ-FLH-009 — Dispatch ordering and idempotency

While a handoff is not `BOUND`, the factory broker SHALL preserve dispatch metadata for the lane but SHALL deny body claim/read and implementation side effects. When the matching durable `BOUND` receipt is observed, the broker SHALL release the dispatch body exactly once to the current generation and SHALL handle duplicate dispatch and same-lane redispatch without duplicate execution by binding idempotency to the handoff generation.

### REQ-FLH-010 — Tombstone and stale traffic rejection

When an old endpoint, stale generation, stale reservation token, or pre-handoff claim token attempts send, read, receipt, or ACK, the broker SHALL reject it with `STALE_ENDPOINT` or `STALE_GENERATION` and return non-secret redirect metadata naming the current lane endpoint and generation. The tombstone SHALL survive restart.

### REQ-FLH-011 — Crash, restart, and abandoned worktree recovery

When a crash occurs before or after creation, branch rename, `SWITCH_PENDING`, the rebind transaction, or receipt delivery, recovery SHALL reread the stored nonce and filesystem, Git, and broker facts and select exactly one of resume, idempotent finalize, or `ABANDONED`. It SHALL preserve dirty or unmerged worktrees without automatic deletion and SHALL NOT guess past a live foreign owner or uncertain endpoint.

### REQ-FLH-012 — Authority and product-boundary truth

While implementation handles a handoff, it SHALL NOT assume hook or MCP authority to execute slash commands, and SHALL NOT describe ChatGPT Desktop Handoff or a Codex-managed detached worktree as a headless MoAI CLI capability. The card worktree SHALL be a MoAI-launcher-managed worktree on a named `WT-*` branch, distinct from a Desktop-managed worktree.

### REQ-FLH-013 — Zero-write safety before BOUND

While a handoff is not `BOUND`, the lane SHALL perform no code edit, commit, card completion, or task ACK. Acceptance evidence SHALL separately count and prove zero pre-BOUND code edits, zero wrong-cwd edits, zero primary-checkout branch switches, and zero messages lost during endpoint replacement.

### REQ-FLH-014 — Live cross-harness proof

When LIVE verification runs, it SHALL exercise an actual interactive `/cd` path and an actual stored-history headless app-server `thread/fork(cwd)` path across Codex lead↔Codex lane and Claude lead↔Codex lane in real separate model/CLI contexts and observe post-rebind bidirectional unique nonces, generation-bound explicit receipts, statusline, cwd, branch, and old-endpoint rejection. The interactive proof SHALL use the user's next normal turn and record zero empty model turns; the headless proof SHALL require method `thread/fork`, a nonempty returned thread ID, a nonempty `forkedFromId`, and `thread/started=true`, then bind from that official result without waiting for SessionStart. Each proof SHALL write versioned card-scoped structured evidence whose process identity, mode transition, target provenance, receipts, message counters, cleanup, and bypass-denial fields are validated by exact typed predicates rather than self-reported stdout strings. A gate-quality test SHALL prove that missing-field, mock, fixture, direct-registration, child-failure, child-skip, and stored-history `wrong_method_thread_start` mutants are rejected. No-history `thread/start(cwd)` remains valid only for the unit contract in REQ-FLH-007/AC-FLH-004 and SHALL NOT satisfy the stored-history LIVE proof. A child/subtest/package failure, skip, mock, fixture-only path, direct registration, manual DB seed, private control, malformed or missing evidence, or `NOT_RUN` SHALL fail the affected LIVE criterion.

### REQ-FLH-015 — Reuse and compatibility

While implementation handles a handoff, it SHALL reuse t1074 canonical run selection, stable lane/current endpoint generation, launcher provisional endpoint followed by first-normal-turn SessionStart rebind, the factory message broker, peer roster, explicit receipts, existing homestate, the MoAI worktree launcher, and the app-server client. Headless controller binding SHALL coexist with that interactive first-turn contract without adding a broker, daemon, private transport, or parallel message store or changing legacy `sessionmsg` semantics.

### REQ-FLH-016 — Launch-pending source endpoint admission

When a handoff is requested for a lane whose current endpoint is still the t1074 launcher provisional (launch-pending) endpoint, the handoff SHALL return NACK with reason `ENDPOINT_LAUNCH_PENDING`, SHALL write no reservation, tombstone, worktree, app-server request, endpoint mutation, or dispatch release, and SHALL leave the provisional endpoint neither bound nor rolled back. The handoff SHALL NOT wait for, poll, retry against, or manufacture the lane's first normal turn; a later handoff SHALL require a fresh reservation after the lane's endpoint is bound.

### REQ-FLH-017 — Serialized ordering of launcher bind and handoff rebind

The handoff SHALL order the t1074 launcher provisional-endpoint bind before handoff admission, and SHALL serialize its atomic rebind against every launcher provisional registration or bind on the same lane as write transactions within the existing factory broker's SQLite transaction domain on that lane's single endpoint row. While a handoff is `SWITCH_PENDING`, the rebind SHALL commit only if the lane endpoint still equals the reserved source endpoint (session/thread UUID, generation, PID, process-start). When a launcher registration or bind has changed that endpoint first, the rebind SHALL NACK with `STALE_GENERATION`, write no tombstone, BOUND receipt, or dispatch release, and leave the launcher-owned endpoint current. When the rebind commits first, a later launcher bind SHALL leave the handoff-bound endpoint unchanged. In every interleaving the lane SHALL end with exactly one current bound endpoint, a generation that never decreases, and no orphan launch-pending endpoint, and the handoff SHALL NOT create its new endpoint through the launcher provisional registration path.

### REQ-FLH-018 — UserPromptSubmit peer registration during an unfinished handoff

While a lane has a handoff in a non-final state (`RESERVED`, `WT_READY`, `SWITCH_PENDING_INTERACTIVE`, or `SWITCH_PENDING_HEADLESS`), the t1074 UserPromptSubmit peer registration for that lane's slot SHALL be rejected with `ENDPOINT_HANDOFF_PENDING` inside the same broker write transaction that reads the handoff state, SHALL leave the endpoint row, generation, tombstones, receipts, and dispatch release markers unchanged, and SHALL NOT move the endpoint even when the caller's PID and process-start identity equal the current owner's and only the session UUID differs. When the handoff rebind has committed first, a later UserPromptSubmit registration carrying the handoff-bound endpoint identity SHALL leave that endpoint and its generation unchanged, and one carrying the tombstoned source session UUID SHALL be rejected with `STALE_ENDPOINT`. When the handoff reaches `NACK` or `ABANDONED`, UserPromptSubmit registration SHALL resume t1074 semantics without releasing any dispatch body. The rejection SHALL be enforced within the existing factory broker and hook path without adding a broker, daemon, or store.

## Requirement-to-acceptance traceability

| Requirement anchor | Acceptance criteria |
|---|---|
| § REQ-FLH-001 | AC-FLH-001, AC-FLH-007 |
| § REQ-FLH-002 | AC-FLH-003, AC-FLH-004 |
| § REQ-FLH-003 | AC-FLH-001 |
| § REQ-FLH-004 | AC-FLH-002, AC-FLH-016 |
| § REQ-FLH-005 | AC-FLH-002 |
| § REQ-FLH-006 | AC-FLH-003, AC-FLH-014 |
| § REQ-FLH-007 | AC-FLH-004, AC-FLH-014 |
| § REQ-FLH-008 | AC-FLH-005, AC-FLH-011 |
| § REQ-FLH-009 | AC-FLH-006, AC-FLH-008, AC-FLH-011 |
| § REQ-FLH-010 | AC-FLH-007 |
| § REQ-FLH-011 | AC-FLH-009, AC-FLH-010 |
| § REQ-FLH-012 | AC-FLH-014 |
| § REQ-FLH-013 | AC-FLH-011, AC-FLH-012, AC-FLH-013 |
| § REQ-FLH-014 | AC-FLH-012, AC-FLH-013 |
| § REQ-FLH-015 | AC-FLH-015 |
| § REQ-FLH-016 | AC-FLH-017 |
| § REQ-FLH-017 | AC-FLH-018 |
| § REQ-FLH-018 | AC-FLH-019 |

### Out of Scope — Idle wake and autonomous TUI control

- BOUND 이후 idle session을 깨우는 기능은 t1075 범위다.
- tmux 키 주입, private socket, TUI 내부 API를 통한 무인 조작은 구현하지 않는다.

### Out of Scope — Desktop Handoff emulation

- ChatGPT Desktop의 Handoff UI와 Git 이동 동작을 MoAI CLI에서 복제하지 않는다.
- Codex-managed detached worktree를 card worktree로 채택하지 않는다.

### Out of Scope — Queue ownership and completion

- 이 handoff는 card 선택, queue 완료, PR/merge, deploy 권한을 부여하지 않는다.
- `BOUND` receipt는 transport/work-root 결합 증거이며 구현 완료 증거가 아니다.

### Out of Scope — New transport infrastructure

- 새 message broker, daemon, polling service, private IPC를 만들지 않는다.
- t1074 broker와 legacy `sessionmsg` 사이를 자동 이관하지 않는다.
