---
id: SPEC-V3R6-HOOK-ASYNC-EXPAND-001
title: "Hook async 확대 (FileChanged + ConfigChange + TaskCreated + Notification)"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/hook, internal/hook/testutil"
lifecycle: spec-anchored
tags: "hook, async, goroutine, performance, latency, v3.0, sprint-2, wave-4-deferral"
tier: M
issue_number: null
depends_on: [SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001]
related_specs: [SPEC-V3R6-HOOK-CONTRACT-FIX-001, SPEC-V3R6-AGENT-MODEL-ROUTING-001, SPEC-V3R6-PROMPT-CACHE-001]
---

# SPEC-V3R6-HOOK-ASYNC-EXPAND-001: Hook Async Expansion

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-23 | manager-spec | Initial draft — Sprint 2 SPEC. Transition 4 hook handlers (`FileChanged`, `ConfigChange`, `TaskCreated`, `Notification`) from synchronous to asynchronous execution per `.moai/research/v3.0-design-2026-05-22.md` § 4 Layer 4 둘째 항목. Tier M (~600 LOC, 8-10 files in-scope including handlers, tests, testutil helper, and goleak verification). Cross-Sprint dependency: `SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001` gates REQ-HAE-003/004 (TaskCreated + Notification require `observability.enabled` flag). Out-of-scope: single-mega-dispatcher consolidation (deferred to a future Wave 4+ SPEC), other hook events (Stop/SubagentStop/UserPromptSubmit/PreToolUse/SessionStart retain current profile), external queue / message broker. |

## 1. Goal

Reduce per-hook **blocking time** for four high-frequency observe/notify hook events by transitioning their Go handlers from synchronous to asynchronous execution. The handler's main return path completes within ≤ 100 ms (p95) regardless of side-effect work, while side-effects (logging, JSONL append, metrics emission, diff-aware reload) execute in a background goroutine with a strict 5-second deadline.

This is **Layer 4 (Hook 효율화)** of the v3.0 token-economy design. Quoted verbatim from `.moai/research/v3.0-design-2026-05-22.md` § 4 Layer 4 둘째 항목 (lines 249-252):

> **Async 확대**:
> - 현재 `PostToolUse`만 async
> - v3.0: `FileChanged`, `ConfigChange`, `TaskCreated`, `Notification`도 async 전환
> - 효과: blocking 시간 감소 (사용자 인지 응답 속도 ↑)

## 2. Why

### 2.1 Current Synchronous Profile (Baseline)

Per `.moai/research/moai-adk-current-state-2026-05-22.md` § 6.2 hook event table (line 285-302), **only `PostToolUse` carries `Async? Yes`** today. The remaining 25+ events — including the four targets of this SPEC — block the Claude Code main loop while their Go handler in `internal/hook/<event>.go` completes side-effects. Per § 6.4 (lines 313-316), each shell wrapper invocation also carries ~50-100 tokens of stdin/stdout JSON overhead, so blocking time directly translates into perceived response latency in interactive sessions.

For the four target events, side-effects are observably non-critical to the response payload:

| Event | Side-effect work (current sync) | Critical-path payload |
|-------|----------------------------------|------------------------|
| `FileChanged` | MX tag delta scan across 16 language extensions (`internal/hook/file_changed.go:14-37`) | `{"continue": true}` |
| `ConfigChange` | Diff-aware reload + validation on `project_settings`/`local_settings`/`skills` matchers | `{"continue": true, "exitCode": 0}` |
| `TaskCreated` | JSONL append to `.moai/state/task-events.jsonl` (observability tap, opt-in) | empty `{}` payload |
| `Notification` | Payload processing + transcript collection (observability tap, opt-in) | empty `{}` payload |

The critical-path payload is fixed and small. The side-effect work is the latency dominant. Decoupling them is the design intent.

### 2.2 Cross-Sprint Dependency on HOOK-OBSERVE-OPT-IN-001

Per `.moai/research/v3.0-design-2026-05-22.md` § 4 Layer 4 첫째 항목 (lines 245-247), `TaskCreated`, `Notification`, and `handle-harness-observe-*` are scheduled for **opt-in default-OFF** behavior via `.moai/config/sections/observability.yaml`. SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 owns that flag.

The interaction:

- `FileChanged` + `ConfigChange` async transition: **unconditional** — these are core MX/config flows, always-on in v3.0.
- `TaskCreated` + `Notification` async transition: **conditional on `observability.enabled == true`**. When the flag is false (v3.0 default), the handler emits the empty payload immediately and does NOT launch a goroutine (no work to defer).

REQ-HAE-003/004 explicitly gate on the observability flag to keep the two SPECs cleanly composable. Order: HOOK-OBSERVE-OPT-IN-001 merges first; this SPEC's REQ-HAE-003/004 acceptance verification then runs the matrix `{enabled=false, enabled=true} × {TaskCreated, Notification}`.

### 2.3 Out-of-scope deferral rationale

The design doc § 4 Layer 4 셋째 항목 (lines 258-260) lists a **single mega-dispatcher** consolidation (31 shell wrappers → one `moai hook dispatch <event>` entry-point). That is a long-term architectural change with cross-cutting blast radius (settings.json template, all shell wrappers, CLI subcommand registry) — it deserves its own Wave 4 SPEC. This SPEC limits scope to the four async transitions and the supporting test infrastructure.

## 3. EARS Requirements

### REQ-HAE-001 (Ubiquitous, FileChanged async return)

The `FileChanged` hook handler in `internal/hook/file_changed.go` shall return its `HookOutput` (with `continue: true`, `exitCode: 0`) within ≤ 100 ms (p95 under 10-concurrent benchmark, AC-HAE-002), regardless of side-effect completion time. MX tag delta scan across the 16 supported language extensions shall execute in a background goroutine launched via `go func()` from within the handler.

### REQ-HAE-002 (Ubiquitous, ConfigChange async return)

The `ConfigChange` hook handler in `internal/hook/config_change.go` shall return its `HookOutput` non-blocking within ≤ 100 ms (p95 under 10-concurrent benchmark, AC-HAE-003). Side-effects — diff-aware reload of `project_settings`/`local_settings`/`skills` matchers and validation — shall execute in a background goroutine with a 5-second deadline (REQ-HAE-005).

### REQ-HAE-003 (Where + When, TaskCreated conditional async)

Where the `observability.enabled` flag (owned by SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 in `.moai/config/sections/observability.yaml`) is `true`, when the `TaskCreated` hook handler in `internal/hook/task_created.go` is invoked, the handler shall complete its main return path within ≤ 100 ms (p95 under 10-concurrent benchmark, AC-HAE-004). JSONL append to `.moai/state/task-events.jsonl` shall execute asynchronously. Where `observability.enabled == false`, the handler shall return the empty payload immediately without launching a goroutine.

### REQ-HAE-004 (Where + When, Notification conditional async)

Where the `observability.enabled` flag is `true`, when the `Notification` hook handler in `internal/hook/notification.go` is invoked, the handler shall complete its main return path within ≤ 100 ms (p95 under 10-concurrent benchmark, AC-HAE-005). Payload processing — including transcript collection if applicable — shall execute asynchronously. Where `observability.enabled == false`, the handler shall return the empty payload immediately without launching a goroutine.

### REQ-HAE-005 (When, deadline enforcement)

When any background goroutine launched by REQ-HAE-001..004 exceeds its 5-second deadline (measured from goroutine start), the system shall log a structured warning to `.moai/state/hook-async-warnings.jsonl` (one JSON-object-per-line) containing `{timestamp, event, deadline_ms: 5000, elapsed_ms, message}`, and the goroutine shall self-cancel via `context.WithTimeout`-derived context. The warning log shall not block the main return path.

### REQ-HAE-006 (Where, test harness determinism)

Where unit or integration tests assert hook side-effects (JSONL writes, MX delta files, validation triggers), the test harness shall use a `WaitForAsync(t *testing.T, deadline time.Duration)` helper located at `internal/hook/testutil/wait_async.go` to deterministically await goroutine completion. The helper shall use `sync.WaitGroup`-or-equivalent (registered by the production goroutine), NOT `time.Sleep`. AC-HAE-008 verifies the helper is used by ≥ 4 test files.

## 4. Acceptance Criteria (binary)

See [acceptance.md](./acceptance.md) for the 8 binary AC matrix (AC-HAE-001..008), each with verification command, expected output, and traceability mapping.

## 5. Risks

Five risks identified. Full mitigation strategy in [plan.md](./plan.md) § Risks.

| # | Risk | Likelihood / Impact | Mitigation |
|---|------|---------------------|------------|
| R1 | Goroutine leaks if 5-second deadline not strictly enforced | Medium / High | AC-HAE-007 `goleak.VerifyNone` + REQ-HAE-005 self-cancel via `context.WithTimeout` |
| R2 | Race conditions in test assertions reading async side-effects | Medium / Medium | REQ-HAE-006 `WaitForAsync` helper + AC-HAE-006 `go test -race` |
| R3 | Cross-Sprint conflict with SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 | Medium / Medium | `depends_on` declared in frontmatter; REQ-HAE-003/004 explicitly gate on `observability.enabled` |
| R4 | Claude Code SDK expectation of synchronous return for some events | Low / Medium | Contract test: re-verify `exitCode: 0, continue: true` accepted regardless of side-effect timing |
| R5 | Sync OK while async fails silently — error reporting asymmetry | Low / Low | REQ-HAE-005 warning log + `moai doctor` reports last 7 days async warnings |

## 6. Exclusions (What NOT to Build)

### 6.1 Out of Scope: Single mega-dispatcher consolidation

Unification of 31 shell wrappers under `.claude/hooks/moai/` into one `moai hook dispatch <event>` entry-point (per design doc § 4 Layer 4 셋째 항목, lines 258-260) is a separate Wave 4+ initiative. This SPEC keeps the per-event shell-wrapper-to-Go-handler topology unchanged. Only the body of the four target Go handlers is modified.

### 6.2 Out of Scope: External queue / message broker

A Redis-, Kafka-, or disk-queue-backed async dispatcher for hook side-effects is **explicitly excluded**. The 5-second deadline + in-process `go func()` + `context.WithTimeout` approach is sufficient for v3.0 latency goals (≤ 100 ms p95 handler return). External queues add operational surface (broker availability, retry semantics, dead-letter handling) that does not justify the cost at v3.0.

### 6.3 Out of Scope: Async transition for other hook events

The following hook events retain their current sync/async profile per `.moai/research/moai-adk-current-state-2026-05-22.md` § 6.2:

- `Stop`, `SubagentStop` — sync, blocking (`exit 2` for quality gate; tmux pane cleanup is intentionally synchronous to ensure resources are freed before the next prompt).
- `UserPromptSubmit` — sync, blocking (SPEC detection + session title must complete before user sees the model's first token).
- `PreToolUse` — sync, blocking (tool-blocking decision via `exit 2`).
- `SessionStart` — sync (GLM setup, skill discovery, memory load — must complete before user interaction).
- `PostToolUse` — **already async** (no change).
- `WorktreeCreate`, `WorktreeRemove` — sync (active-creator contract per SPEC-V3R6-HOOK-CONTRACT-FIX-001 REQ-HCF-001/002 — plain-text stdout return that Claude Code parses as worktree path; converting these to async would break the contract).

Only the four events enumerated in § 1 Goal and § 3 EARS REQs transition here.

## 7. Cross-References

- **Design doc**: `.moai/research/v3.0-design-2026-05-22.md` § 4 Layer 4 (lines 243-261, especially 249-252)
- **Baseline**: `.moai/research/moai-adk-current-state-2026-05-22.md` § 6.2 hook event table (line 285-302) + § 6.4 token overhead (lines 313-316)
- **Frontmatter SSOT**: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- **GEARS notation**: Adopted via SPEC-V3R6-GEARS-MIGRATION-001 (MERGED PR #1046)
- **Sprint/Wave SSOT**: `.claude/rules/moai/development/sprint-wave-naming.md` — this SPEC is one of four in Sprint 2
- **Dependency partner**: `SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001` (must merge first; gates REQ-HAE-003/004)
- **Independent siblings**: `SPEC-V3R6-AGENT-MODEL-ROUTING-001`, `SPEC-V3R6-PROMPT-CACHE-001` (Sprint 2 same wave, no overlap)
- **Related (prior)**: `SPEC-V3R6-HOOK-CONTRACT-FIX-001` (defensive consolidation; this SPEC respects its REQ-HCF-001/002 by excluding WorktreeCreate/Remove from § 6.3)
