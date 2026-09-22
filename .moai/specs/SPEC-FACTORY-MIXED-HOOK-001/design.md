---
id: SPEC-FACTORY-MIXED-HOOK-001
document: design
created: 2026-09-22
updated: 2026-09-22
author: manager-spec
card: t1074
module: "internal/factorymsg"
---

# Design — SPEC-FACTORY-MIXED-HOOK-001

## Components

```text
launcher ──select/join──> canonical factory run registry
   │                     logical lane ID ──> current endpoint(UUID/gen)
SessionStart ──bind/rebind UUID/gen─────────┘
   │
MCP send/body/receipt ──> factory broker <── hook claim metadata
                              │
                     pending → claimed → acknowledged
                              └──────> dead-letter
```

The broker path is derived from `homestate.ProjectDir(CanonicalProjectRoot(cwd))`, then `factory/messages/<run-id>`. Worktree cwd is provenance only. Legacy `.moai/state/session-msg` remains separate.

## Identity

`FactoryPeer = {project_key, run_id, backend, role, slot, session_uuid, generation, pid, process_start}`. The role/slot pair forms the stable logical lane address; `session_uuid` and `generation` identify its replaceable physical endpoint. Launcher creates/joins; SessionStart binds the real hook session UUID. Every resolve/claim/read/receipt verifies the current full recipient tuple. Restart increments generation, and a stale endpoint cannot receive or acknowledge traffic.

## Worktree handoff seam

This SPEC does not create or enter worktrees. `t1082` owns the state machine `reserve → create → SWITCH_PENDING → /cd or headless cwd fork → SessionStart rebind → BOUND`. t1074 supplies only the atomic current-endpoint binding and stale-generation rejection required by that state machine. `t1075` must address the stable lane and wake only the endpoint that is current after `BOUND`.

## Envelope and states

`FactoryEnvelope` contains all REQ-FMH-004 fields and a payload stored outside hook context. `Send` is idempotent per sender/run/idempotency key. `Claim` returns metadata plus opaque claim token. `ReadBody` requires the intended recipient generation and active claim. `RecordDispositionAndReceipt` atomically persists disposition and verifies claim token before acknowledgement. Transport receipt is a new envelope; it is not work completion.

## Hook merge rules

- SessionStart: capability and bound identity only.
- UserPromptSubmit: append metadata-only retrieval instructions to existing additionalContext.
- Stop: if existing logic denies/stops, preserve it. Otherwise claim a new actionable batch only when chain ledger has budget; return one continuation. `stop_hook_active` and ledger prevent a second continuation.
- No hook performs ACK. No receipt-only batch wakes a model.

## Failure posture

- Corrupt/unknown messages go to dead-letter under lock, preserving reason.
- Lock/deadline failure leaves pending state and reports degraded capability.
- Stale generation/token operations fail closed.
- Crash after claim relies on lease expiry and idempotency/revision guards.
- Active-run ambiguity fails before child launch.

## Observability

Status exposes counts and IDs, not payloads: run, peers/generation, pending/claimed/acknowledged/dead-letter, oldest pending, last hook, next-delivery condition, and capability. Secrets and message bodies are excluded from logs.
