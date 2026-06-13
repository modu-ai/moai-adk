---
id: SPEC-SEC-HARDEN-001
title: "Security & Concurrency Hardening (full-codebase review HIGH findings)"
version: "0.1.0"
status: draft
created: 2026-06-13
updated: 2026-06-13
author: GOOS행님
priority: P1
phase: "v3.0.0"
module: "internal/permission, internal/tmux, internal/lsp, internal/resilience, internal/cli/worktree"
lifecycle: spec-anchored
tags: "security, concurrency, permission, tmux, lsp, circuit-breaker, behavior-preservation, cwe-214, data-race"
era: V3R6
tier: L
---

# SPEC-SEC-HARDEN-001 — Security & Concurrency Hardening

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-13 | manager-spec | Initial draft. 5 HIGH-severity defects found and independently verified during a full-codebase review at HEAD `0ef553617`. Five milestones M1-M5, one per finding, across 4 packages. All fixes behavior-preserving — characterization/reproduction test FIRST (RED), then fix, then no-regression verification. |

---

## §A. Context & Motivation

A full-codebase review at HEAD `0ef553617` surfaced 5 HIGH-severity defects spanning security allow/deny logic, credential injection, and concurrency. Each was confirmed by reading the actual source. They share a common risk profile: the code paths are security-sensitive or concurrency-sensitive, and a naive fix risks silently changing observable allow/deny/inject behavior.

The motivation for grouping them into one Tier-L SPEC is that all five require the **same disciplined methodology**: a characterization or reproduction test that demonstrates the defect FIRST (RED), then the minimal behavior-preserving fix, then verification that the existing allow/deny/inject behavior is unchanged for the non-defect paths. Bundling them keeps that methodology consistent and lets a single quality gate verify behavior preservation across all five.

Two of the findings (M3, M5) are regressions or omissions relative to prior security/resilience work:
- M3 re-leaks `ANTHROPIC_AUTH_TOKEN` via the worktree `--team` path, regressing the SPEC-V3R5-SECURITY-CRIT-001 P0-2 fix that was applied to `glm.go` but not mirrored to `tmux_integration.go`.
- M5 leaves a documented circuit-breaker invariant ("only one request through in half-open") structurally absent and an `OnStateChange` goroutine unrecovered (the code's own `@MX:WARN` already flags it).

### Verified evidence index (all confirmed against HEAD `0ef553617`)

| Milestone | Package | File:Line (verified) | Class |
|-----------|---------|----------------------|-------|
| M1 | internal/permission | `stack.go:127-128` | SECURITY — prefix-match command-chain bypass |
| M2 | internal/permission | `conflict.go:26-55` (resolveConflict) + `conflict.go:57-68` (logConflict) | SECURITY — deny does not win on tie + audit log not written |
| M3 | internal/cli/worktree + internal/tmux | `tmux_integration.go:77-84` + `session.go:189-203` (leak) + `glm.go:389-408` (canonical fix) | SECURITY — tmux credential argv leak (CWE-214) |
| M4 | internal/lsp/hook | `tracker.go:69-84` (GetBaseline) + `tracker.go:112-130` (loadBaselineLocked, write at line 128) | CONCURRENCY — shared-state write under read lock (data race) |
| M5 | internal/resilience | `circuit.go:61-97` (Call) + `circuit.go:181-198` (transitionTo) | RESILIENCE/CONCURRENCY — half-open permit absent + unrecovered callback goroutine |

---

## §B. GEARS Requirements

### M1 — Permission `:*` prefix-match command-chain bypass

- **REQ-SEC-M1-001** (Unwanted behavior): The permission resolver shall not treat a `:*` prefix-match allow rule as matching an input that carries an unquoted shell command separator beyond the matched prefix.
- **REQ-SEC-M1-002** (Event-driven): When the resolver matches a `:*` prefix rule against an input whose remainder past the prefix contains an unquoted shell command separator (`;`, `&&`, `||`, `|`, `$(`, backtick, or newline), the resolver shall report no match for that rule (so the chained command is not allowed by the prefix rule).
- **REQ-SEC-M1-003** (Ubiquitous): The resolver shall continue to match a `:*` prefix rule against a legitimate single command that shares the prefix and contains no unquoted shell separator in its remainder.
- **REQ-SEC-M1-004** (State-driven): While a shell separator character appears only inside a quoted argument segment of the input, the resolver shall not treat that quoted separator as a command-chain boundary (no false rejection of a single command containing a quoted separator).

### M2 — Permission conflict resolution: deny wins on tie + conflict audit log written

- **REQ-SEC-M2-001** (Event-driven): When two or more matched rules of the same tier have equal specificity and at least one is a `deny`, the resolver shall return a `deny` rule as the winner before applying any specificity- or Origin-based ordering.
- **REQ-SEC-M2-002** (Ubiquitous): The resolver shall preserve the existing specificity-then-Origin ordering for ties where all candidate rules carry the same `Action` (no deny present).
- **REQ-SEC-M2-003** (Event-driven): When a conflict among two or more matched same-tier rules is resolved, the resolver shall write a conflict audit record to `.moai/logs/permission.log`.
- **REQ-SEC-M2-004** (State-driven): While the conflict-log destination cannot be written (directory missing, permission denied), the resolver shall not fail the permission decision (the conflict-log write is best-effort and must not change the allow/deny outcome).

### M3 — tmux credential argv leak (CWE-214) on the worktree `--team` path

- **REQ-SEC-M3-001** (Unwanted behavior): The worktree tmux integration path shall not pass `ANTHROPIC_AUTH_TOKEN` as a positional argv to `tmux set-environment`.
- **REQ-SEC-M3-002** (Event-driven): When the worktree tmux integration injects environment variables in GLM/CG mode and `ANTHROPIC_AUTH_TOKEN` is present, the integration shall route the token through `InjectSensitiveEnv`, remove it from the bulk map, and bulk-inject only the remaining non-sensitive variables via `InjectEnv`.
- **REQ-SEC-M3-003** (Event-driven): When sensitive-token injection fails on the worktree `--team` path, the integration shall return the error and shall not fall back to argv-based bulk injection for the token.
- **REQ-SEC-M3-004** (Ubiquitous): The worktree tmux integration shall continue to bulk-inject the non-sensitive GLM/CG environment variables (model slots, proxy compatibility flags) via the existing `InjectEnv` path.

### M4 — LSP regression tracker shared-state write under read lock (data race)

- **REQ-SEC-M4-001** (Unwanted behavior): The regression tracker shall not assign the shared `baseline` field while holding only a read lock.
- **REQ-SEC-M4-002** (Event-driven): When `GetBaseline` (or any path reaching it, including `CompareWithBaseline`) loads the baseline lazily, the tracker shall perform the load-and-assign of the shared `baseline` field under a write lock so no two callers concurrently assign it.
- **REQ-SEC-M4-003** (Ubiquitous): The tracker shall preserve its existing observable return contract for `GetBaseline`, `CompareWithBaseline`, and `ClearBaseline` (same return values and `ErrBaselineNotFound` semantics for present/absent baselines).

### M5 — Circuit breaker: half-open single-permit invariant + recovered callback goroutine

- **REQ-SEC-M5-001** (State-driven): While the circuit is in the half-open state, the breaker shall admit at most one in-flight trial request; all other concurrent callers shall be rejected with `ErrCircuitOpen` until the single trial resolves.
- **REQ-SEC-M5-002** (Event-driven): When the single half-open trial request resolves (success or failure), the breaker shall release the half-open in-flight permit so that the next state evaluation governs subsequent calls.
- **REQ-SEC-M5-003** (Event-driven): When the `OnStateChange` callback panics inside its goroutine, the breaker shall recover from the panic so that the panic does not crash the host process.
- **REQ-SEC-M5-004** (Ubiquitous): The breaker shall preserve its existing closed/open transition behavior, threshold counting, timeout-based half-open promotion, and metrics accounting.

---

## §C. Behavior-Preservation Mandate (cross-cutting, all milestones)

[HARD] Every milestone touches security allow/deny logic, credential injection, or concurrency. Each fix MUST be behavior-preserving for the non-defect paths.

- **RED first**: Each milestone MUST land a characterization/reproduction test FIRST that demonstrates the defect against the current code (the test must FAIL on the pre-fix code, asserting the defective behavior is present).
- **Minimal fix**: The implementation change MUST be the minimal change that makes the reproduction test pass. No drive-by refactors, no adjacent cleanup.
- **No regression**: After the fix, the existing allow/deny/inject/transition behavior for all non-defect inputs MUST remain identical. Each milestone's AC includes explicit no-regression assertions and the existing package test suite MUST stay green — **with one carve-out**: a pre-existing test that asserts the very pre-fix defect a milestone is correcting is an **intended-behavior-change casualty** and MUST be rewritten to assert the corrected behavior (NOT preserved green). The single known casualty is `internal/permission/conflict_test.go::TestResolveConflict_FsOrderTiebreak` under M2 (it asserts an equal-specificity allow+deny tie resolves to the ALLOW fs-order winner, which REQ-SEC-M2-001 deliberately inverts to deny-wins). See plan.md §M2 and acceptance.md §M2 for the rewrite contract. No other existing test is expected to change behavior; all others MUST stay green.
- **Race milestones** (M4, M5): the reproduction and verification MUST run under `go test -race`.

---

## §D. Acceptance Criteria Summary

Full Given-When-Then scenarios and per-milestone testable AC live in `acceptance.md`. Each milestone has its own AC group (AC-SEC-M1-*, AC-SEC-M2-*, AC-SEC-M3-*, AC-SEC-M4-*, AC-SEC-M5-*) covering: (a) the reproduction test that FAILS pre-fix, (b) the fix making it PASS, (c) explicit no-regression assertions for the legitimate paths.

---

## §E. Verification Approach

- M1: table-driven `Matches` tests — legitimate prefix command (allowed), chained command (not matched), quoted-separator single command (not falsely rejected), each separator variant.
- M2: `resolveConflict` tests — equal-specificity allow+deny → deny wins; all-allow tie preserves specificity/Origin ordering; `logConflict` writes to `.moai/logs/permission.log`; unwritable log dir does not change decision.
- M3: capture the argv passed to `tmux set-environment` (via an injectable run-func / fake session manager); assert the token value never appears in any positional argv; assert non-sensitive vars still bulk-injected; assert no argv fallback on sensitive-injection failure.
- M4: concurrent `GetBaseline` on a nil baseline under `-race`; assert clean (no data race); assert single-reader return contract unchanged.
- M5: N concurrent half-open `Call`s — exactly 1 executes `fn()`, the rest get `ErrCircuitOpen`; a panicking `OnStateChange` callback does not crash the process (test completes and asserts post-state).

---

## §F. Exclusions (What NOT to Build)

[HARD] The following are explicitly OUT of scope for this SPEC. They are tracked separately from this review and MUST NOT be pulled into SPEC-SEC-HARDEN-001. They are listed here as deferred follow-ups only.

### F.1 Out of scope — the 10 MEDIUM findings (deferred follow-up)

- `gate.go` `IsGitCommit` anchor weakness
- `cwd_changed.go` shell-escape handling
- `pre_tool.go` fail-open behavior + Write-only content scan
- config `Save` / `SetSection` persistence gap
- `async_recorder` send-on-closed-channel
- lsp `client.go` `Start`-error leak
- loop storage `specID` path traversal
- spec `closer.go` no-pathspec commit
- github `spec_linker.go` unlocked reads

(The above enumerates 9 named MEDIUM items; the review's MEDIUM bucket is 10 — the 10th is held in the review report and is likewise deferred. None are addressed here.)

### F.2 Out of scope — the 6 LOW findings (deferred follow-up)

The 6 LOW-severity findings from the same review are deferred. Not enumerated here; tracked in the review report.

### F.3 Out of scope — the 2 uncovered package groups (sweep stalled)

- `cli/cmd`
- `template` / `statusline` / `mx` / `merge`

The full-codebase sweep stalled before covering these package groups. They are NOT reviewed by this SPEC and any defects within them are out of scope.

### F.4 Out of scope — implementation-detail decisions deferred to run-phase

- Exact mechanism for half-open permit (atomic bool vs. dedicated state field vs. token) — design.md proposes a direction; the concrete data structure is a run-phase decision.
- Exact shell-separator detection strategy for M1 (deny-on-any-unquoted-separator vs. require full-command match) — design.md proposes deny-on-unquoted-separator; the concrete parser depth is a run-phase decision.
- Log line format / rotation policy for `.moai/logs/permission.log` beyond "a conflict record is written" — only the write itself is in scope.

### F.5 Out of scope — broad refactors

No refactor of the permission resolver architecture, the tmux session manager API, the LSP tracker design, or the circuit breaker public API beyond the minimal behavior-preserving changes required by M1-M5.

**Narrow carve-out (D1)**: M3 REQUIRES extending the `tmux.SessionManager` interface (`internal/tmux/session.go:50-59`) with exactly ONE additive method — `InjectSensitiveEnv(ctx context.Context, key, value string) error`. This is the minimal behavior-preserving enabler for routing `ANTHROPIC_AUTH_TOKEN` through the argv-safe channel (the interface, not the concrete type, is what `CreateTmuxSession` receives — see plan.md §M3 interface-seam decision). This ONE additive interface method is **in scope**; the "no tmux session manager API refactor" exclusion above is narrowed to permit it and nothing further. `*DefaultSessionManager` already implements the method, so the addition is purely a type-surface widening with no behavior change. No other interface method may be added, removed, or re-signatured.
