---
id: SPEC-FACTORY-MIXED-HOOK-001
title: "Mixed Claude/Codex factory hook-boundary messaging"
version: "0.1.1"
status: in-progress
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/factorymsg"
lifecycle: spec-anchored
card: t1074
tags: "factory,codex,claude,hooks,messaging,receipt"
---

# SPEC-FACTORY-MIXED-HOOK-001

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.1 | 2026-09-22 | Clarify the stable logical lane versus replaceable Codex session endpoint seam and defer worktree handoff to t1082. |
| 0.1.0 | 2026-09-22 | Card t1074 plan baseline for mixed Claude/Codex hook-boundary messaging. |

## WHY

The existing worker registration and legacy session message store do not bind a worker to a canonical factory run/session generation across worktrees, and hook output cannot prove model processing. A new factory-scoped protocol is required to prevent misdelivery, stale acknowledgement, raw-payload privilege elevation, and false claims of idle autonomy.

## WHAT

This SPEC defines run membership, a factory-only canonical broker, generation/token-bound explicit receipts, metadata-only hook delivery, bounded Stop continuation, MCP surfaces, compatibility, live cross-harness verification, and performance evidence.

## HOW

Implementation proceeds through identity/namespace, broker/receipt, hook integration, then real mixed-session verification. Existing homestate resolution, factory registry, MCP registration, hook merge seams, and Codex adapter are reused where their measured contracts fit; legacy sessionmsg data remains separate.

## User story

혼합 factory 운영자로서 Codex/Claude lead와 worker가 서로 다른 CLI·세션·worktree에서도 같은 factory run에 결합되어, 턴 경계에서 검증 가능한 메시지와 명시 receipt를 양방향 교환하길 원한다. 그래야 transport 전달과 실제 업무 완료를 혼동하지 않고 기존 `moai codex -f`, `moai cc -f`, worker 별칭을 유지할 수 있다.

## Scope and baseline

- 카드: `t1074`, Tier L.
- 기준선: `WT-factory-mixed-hook@758314007`, 작성 시 로컬 `develop`과 동일.
- 필수 조합: Codex lead↔Codex worker, Codex lead↔Claude worker, Claude lead↔Codex worker. Claude↔Claude는 회귀 대상이다.
- 이 카드는 hook-boundary 전달만 지원한다. Worktree 생성, interactive `/cd`, headless `cwd` handoff, 그에 따른 endpoint rebind는 `t1082`가 소유한다. idle wake는 rebind가 `BOUND`가 된 뒤 `t1075`가 담당하며, `CLAUDE.local.md` Codex 로딩은 `t1078` 범위다.
- 기존 legacy `.moai/state/session-msg` 저장소는 자동 이동·삭제·의미 변경하지 않는다.

## Requirements (GEARS)

### REQ-FMH-001 — Canonical factory identity

The factory launcher SHALL bind every lead/worker to canonical project key, run ID, stable logical lane ID, runtime backend, session UUID, generation, PID, and process-start identity. The logical lane ID SHALL remain the broker address while the session UUID/generation is a replaceable physical endpoint, and all delivery operations SHALL resolve the currently bound endpoint. When no active run exists, the launcher SHALL fail with `NO_ACTIVE_FACTORY`; when multiple active runs exist, it SHALL fail with `AMBIGUOUS_FACTORY`; when `--factory-run <id>` is supplied before `--`, it SHALL select one run and SHALL NOT forward that option to the child.

### REQ-FMH-002 — Atomic membership

The factory registry SHALL claim worker slots atomically inside the selected run and SHALL reject stale generation/session ownership. While an owner's PID and process-start identity remain live, the registry SHALL NOT displace that owner because of heartbeat expiry alone. The launcher SHALL preserve `agent-N` and `lane-N` compatibility.

### REQ-FMH-003 — Isolated durable broker

The factory broker SHALL use a factory-only namespace under the canonical homestate project directory, isolated by run ID. Where linked worktrees belong to the same repository and run, the broker SHALL resolve one namespace; where project or run differs, it SHALL resolve a distinct namespace. The implementation SHALL leave the legacy sessionmsg namespace untouched.

### REQ-FMH-004 — Closed envelope

The factory broker SHALL validate every message's schema version, project/run, sender and recipient session/generation, message ID, idempotency key, closed kind, task reference, correlation ID, timestamps, TTL, and payload. The broker SHALL accept only `dispatch_notice`, `status_request`, `status_report`, `blocker`, and `receipt` kinds.

### REQ-FMH-005 — Delivery and explicit receipt

The factory broker SHALL transition delivery `pending → claimed → acknowledged`. The broker SHALL NOT acknowledge from hook output, a later hook, or heartbeat. When the model has read the body and persisted an `accepted`, `rejected`, `duplicate`, or `deferred` disposition, the broker SHALL acknowledge only an explicit receipt whose recipient generation and claim token match. When a claim lease expires, the broker SHALL redeliver; when a message is poison, expired, or over policy bounds, it SHALL write a bounded dead-letter record with a reason.

### REQ-FMH-006 — Idempotency and revision safety

When a sender retries, the factory broker SHALL deduplicate on idempotency key. The broker SHALL preserve at-least-once delivery and SHALL NOT claim exactly-once execution. When a task-bearing message is about to cause a side effect, the consumer SHALL validate its expected task revision.

### REQ-FMH-007 — Safe hook context

The UserPromptSubmit and Stop integrations SHALL inject only validated message IDs, kind, sender identity, and the body lookup procedure. The integrations SHALL NOT inject raw peer payload as developer context or continuation text. The injection SHALL contain at most 16 IDs and 2 KiB. The model SHALL fetch the body as untrusted MCP data.

### REQ-FMH-008 — Bounded continuation

The Claude Stop hook SHALL execute synchronously. While a continuation chain is active, the hook SHALL auto-continue at most once and only for a newly actionable batch. Where the inbox is empty or receipt-only, or the session is lock-blocked, permission-waiting, interrupted, or stopped, the hook SHALL create zero additional model turns. The existing security, quality, and goal stop decisions SHALL retain precedence.

### REQ-FMH-009 — Common MCP and hook core

The MCP send/list/body-read/receipt tools and synchronous hook receive logic SHALL use one factory broker core. The dispatch message SHALL reference an already-authorized card/SPEC and SHALL NOT grant queue mutation, command execution, or completion authority.

### REQ-FMH-010 — Capability truth

Where hook wiring is disabled, untrusted, or incompatible, the integration SHALL produce an explicit capability error. The hook-boundary delivery status SHALL report messages arriving after idle as pending until the next turn and SHALL NOT describe that state as idle-wake or fully autonomous factory.

### REQ-FMH-011 — Compatibility

The launcher SHALL keep bare `-f` as lead mode and `-f agent`, `agent-N`, and `lane-N` as worker forms. The launcher SHALL pass every argument after `--` to the child byte-for-byte and SHALL parse MoAI-owned factory options only before `--`.

### REQ-FMH-012 — Evidence and performance

The live verification SHALL make all three mixed combinations and the Claude↔Claude regression exchange unique nonces in both directions across real separate CLI/model contexts and record explicit receipts. The benchmark SHALL test the acceptance targets of empty-inbox added p95≤50 ms and factory inspection hard deadline≤200 ms without treating them as pre-existing facts. The verdict SHALL NOT pass a live criterion whose state is `NOT_RUN`.

### Out of Scope — Deferred and unrelated work

- Idle-session wake or private TUI control (`t1075`).
- Worktree creation, interactive `/cd`, App Server/SDK `cwd` handoff, endpoint transition state, and atomic old→new session rebind (`t1082`).
- `CLAUDE.local.md` instruction loading (`t1078`).
- A general remote message bus or daemon rewrite.
- Treating transport receipt as card/SPEC completion evidence.
