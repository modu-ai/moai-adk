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
   │                     logical lane ID ──> launch-pending(PID/start/gen)
   │                                               │
startup SessionStart after pending ──optional early bind─┐
first non-empty UserPromptSubmit ──required atomic bind──┴─> bound(UUID/PID/start/gen)
   │
MCP send/body/receipt ──> factory broker <── hook claim metadata
                              │
                     pending → claimed → acknowledged
                              └──────> dead-letter
```

The broker path is derived from `homestate.ProjectDir(CanonicalProjectRoot(cwd))`, then `factory/messages/<run-id>`. Worktree cwd is provenance only. Legacy `.moai/state/session-msg` remains separate.

## Identity

`FactoryPeer = {project_key, run_id, backend, role, slot, endpoint_phase, session_uuid?, generation, pid, process_start}`. The role/slot pair forms the stable logical lane address. The launcher creates a `launch-pending` endpoint from the real child PID/process-start and leaves `session_uuid` absent. Startup `SessionStart` may atomically bind that lane only if provisional registration already exists; because official lifecycle ordering does not guarantee that sequence, the first legitimate non-empty `UserPromptSubmit` is the required fallback. It binds the observed hook session UUID and resolved owner PID/process-start before the same hook checks the inbox. Empty or whitespace-only prompts cannot bind. After the identical endpoint is bound, later prompts skip the registry write entirely, preserving generation and `updated_at`, but still check the inbox. Every resolve/claim/read/receipt verifies the current full recipient tuple, and operations requiring a hook session reject `launch-pending`. Restart increments generation, and a stale endpoint cannot receive or acknowledge traffic.

## Worktree handoff seam

This SPEC does not create or enter worktrees. It establishes the initial-launch `launch-pending → bound` capability-truth contract. `t1082` owns the later state machine `reserve → create → SWITCH_PENDING → /cd or headless cwd fork → endpoint rebind → BOUND` and reuses the same atomic binding/stale-generation seam. `t1075` must address the stable lane and wake only the endpoint that is current after `BOUND`.

## Envelope and states

`FactoryEnvelope` contains all REQ-FMH-004 fields and a payload stored outside hook context. `Send` is idempotent per sender/run/idempotency key. `Claim` returns metadata plus opaque claim token. `ReadBody` requires the intended recipient generation and active claim. `RecordDispositionAndReceipt` atomically persists disposition and verifies claim token before acknowledgement. Transport receipt is a new envelope; it is not work completion.

## Hook merge rules

- Launcher: register process-backed `launch-pending` identity without a session UUID.
- SessionStart: handle only the documented startup/resume/clear/compact lifecycle and attempt an idempotent best-effort early bind when a matching provisional row already exists. Correctness never depends on it running after launcher registration.
- UserPromptSubmit: reject binding for empty/whitespace input; otherwise perform the required provisional-to-actual bind before inbox batch processing, then append metadata-only retrieval instructions to existing additionalContext. If the identical endpoint is already bound, skip the bind write but still run inbox processing.
- Stop: if existing logic denies/stops, preserve it. Otherwise claim a new actionable batch only when chain ledger has budget; return one continuation. `stop_hook_active` and ledger prevent a second continuation.
- No hook performs ACK. No receipt-only batch wakes a model.

## Failure posture

- Corrupt/unknown messages go to dead-letter under lock, preserving reason.
- Lock/deadline failure leaves pending state and reports degraded capability.
- Stale generation/token operations fail closed.
- Crash after claim relies on lease expiry and idempotency/revision guards.
- Active-run ambiguity fails before child launch.
- No empty/fake model turn, direct DB seed, manual registration, or hook-trust bypass may manufacture a `bound` endpoint.
- A repeated prompt on an already-bound identical endpoint is a no-write idempotent no-op for peer state; it cannot refresh `updated_at` merely because a turn occurred.

The lifecycle distinction above follows the official Codex hooks contract: `SessionStart` is emitted for startup, resume, clear, and compact sources, whereas `UserPromptSubmit` runs before a submitted user prompt is sent. Source: https://developers.openai.com/codex/hooks.

## Observability

Status exposes counts and IDs, not payloads: run, peers/generation, pending/claimed/acknowledged/dead-letter, oldest pending, last hook, next-delivery condition, and capability. Secrets and message bodies are excluded from logs.
