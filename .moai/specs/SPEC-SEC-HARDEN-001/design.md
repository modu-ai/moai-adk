---
id: SPEC-SEC-HARDEN-001
title: "Security & Concurrency Hardening — Design"
version: "0.1.0"
status: draft
created: 2026-06-13
updated: 2026-06-13
author: GOOS행님
priority: P1
phase: "v3.0.0"
module: "internal/permission, internal/tmux, internal/lsp, internal/resilience, internal/cli/worktree"
lifecycle: spec-anchored
tags: "security, concurrency, design, rationale, trade-offs"
era: V3R6
tier: L
---

# SPEC-SEC-HARDEN-001 — Design

This document records the per-milestone design direction and the trade-offs considered. Concrete data-structure / parser-depth / seam decisions flagged here are deliberately left to run-phase (see spec.md §F.4); design.md proposes a direction, not a binding implementation.

## §M1 — Prefix-match command-chain bypass

### Root cause
`stack.go:127-128` does `return strings.HasPrefix(input, prefix)` for a `:*` rule. The prefix is the only thing checked; everything after the prefix is unvalidated, so any shell continuation rides in on an allowed prefix.

### Design direction: deny-on-unquoted-separator
When the `:*` prefix matches, scan the remainder `input[len(prefix):]` for an unquoted shell command separator. If present, return `false`.

Separators in scope: `;`, `&&`, `||`, `|`, `$(`, backtick, newline.

"Unquoted" detection: a single-pass scanner over the remainder tracking single-quote and double-quote state; a separator counts only when quote-depth is zero. This is intentionally a lightweight lexical scan, NOT a full shell parser — full POSIX shell parsing (here-docs, escaped quotes inside double quotes, `$'...'`, process substitution `<()`) is out of scope (spec.md §F.4). The scanner errs toward **denying** when ambiguous, which is the security-safe direction for an allow-rule matcher: a borderline input simply fails to match the prefix rule and falls through to the normal ask/deny path rather than being silently allowed.

### Alternatives considered
- **Require full-command match** (drop `:*` prefix semantics): rejected — breaks every existing `:*` allow rule and changes broad observable behavior; violates behavior-preservation.
- **Blanket-reject any separator char regardless of quoting**: rejected — falsely rejects legitimate single commands containing a quoted separator (REQ-SEC-M1-004 / AC-SEC-M1-004).
- **Regex blacklist**: rejected — quote-awareness is hard to express robustly in a single regex; a small explicit scanner is clearer and testable.

### Behavior-preservation argument
The change is confined to the `:*` branch and only adds a rejection condition (separator present AND unquoted). Inputs with no unquoted separator behave exactly as before. The `/*`, `*.`, and exact-match branches are untouched.

## §M2 — Conflict resolution: deny-precedence + audit log

### Root cause
- `resolveConflict` (26-55) ranks by `specificityScore` then `Origin` lexicographic; `Action` is never consulted. An equal-specificity allow can beat a deny by filesystem-path ordering, violating "deny wins on tie".
- `logConflict` (57-68) builds `origins` then `_ = origins // Reserved for future log-file writes (currently silent)` — the audit trail promised by REQ-V3R2-RT-002-042 AC-12 ("recorded in .moai/logs/permission.log") is never written.

### Design direction: scope the tiebreak to the deny set
Keep the existing specificity/Origin loop verbatim but choose its input set:
- If any candidate rule has `Action == DecisionDeny`, run the existing loop over **only the deny rules**.
- Otherwise, run it over all rules (unchanged).

This is minimal and behavior-preserving: the all-allow case is byte-identical to today; the mixed case now guarantees a deny winner; and across specificity tiers, a higher-specificity rule still wins because deny-precedence only kicks in to choose among the deny rules when a tie exists — the design must preserve "higher specificity wins regardless of action" (AC-SEC-M2-004). The simplest correct ordering: first compute the max specificity among ALL matched rules; if the max-specificity set contains a deny, restrict to denies; otherwise pick by the existing rule. Run-phase decides whether to implement via a pre-filter or an in-loop guard, provided the AC matrix passes.

### Conflict log write
Replace the `_ = origins` placeholder with an append to `.moai/logs/permission.log`. Requirements:
- Resolve the log path project-root-relative. The current `logConflict` signature has no root — a run-phase seam decision (inject the dir, derive from a package-level configurable root, or pass through). This is flagged in plan.md §F (M2) as a run-phase decision.
- Create the parent directory if absent; open append-only.
- **Best-effort**: any I/O error is swallowed; the permission decision MUST NOT change and MUST NOT surface an error (REQ-SEC-M2-004 / AC-SEC-M2-007).
- Update the `@MX:NOTE` at conflict.go:15-16 to state the write is implemented.

### Alternatives considered
- **Sort by Action string** (`"allow" < "deny"`): rejected — incidental/fragile (relies on lexical accident); explicit deny-precedence is clearer.
- **Synchronous error on unwritable log**: rejected — would change the decision path and could deny-by-side-effect; the audit trail must be observational only.

## §M3 — tmux credential argv leak

### Root cause
`tmux_integration.go:77-84` bulk-injects `cfg.GLMEnvVars` — which includes `ANTHROPIC_AUTH_TOKEN` (added by the `ANTHROPIC_` prefix filter at line 201) — through `InjectEnv`. `session.go:189-203` passes each value as positional argv to `tmux set-environment`, and its own doc (191-194) states this leaks via `/proc/<pid>/cmdline` and `ps -ef` (CWE-214) and says "For credentials use InjectSensitiveEnv instead." The worktree `--team` path simply never adopted the glm.go P0-2 fix.

### Design direction: mirror glm.go:389-408 exactly
The canonical correct pattern already exists and is verified:
```
const sensitiveKey = "ANTHROPIC_AUTH_TOKEN"
if token := vars[sensitiveKey]; token != "" {
    if err := mgr.InjectSensitiveEnv(ctx, sensitiveKey, token); err != nil {
        return fmt.Errorf("inject sensitive tmux env: %w", err)   // NO argv fallback
    }
    delete(vars, sensitiveKey)
}
if len(vars) == 0 { return nil }
return mgr.InjectEnv(ctx, vars)
```
Apply the identical shape inside the `if cfg.ActiveMode == "glm" || cfg.ActiveMode == "cg"` block of `tmux_integration.go`, operating on `cfg.GLMEnvVars` and the function's `tmuxMgr`.

### Interface-seam constraint (D1 — compile blocker)
A subtlety distinguishes this path from glm.go: `glm.go` calls `mgr := tmux.NewSessionManager()` which returns the **concrete** `*DefaultSessionManager`, so `mgr.InjectSensitiveEnv(...)` compiles. `CreateTmuxSession`, by contrast, receives `tmuxMgr tmux.SessionManager` — the **interface** (verified at `tmux_integration.go:48`). The `SessionManager` interface (`session.go:50-59`) declares ONLY `Create` / `InjectEnv` / `ClearEnv`; `InjectSensitiveEnv` is defined only on `*DefaultSessionManager` (`session.go:271`). Calling `tmuxMgr.InjectSensitiveEnv(...)` on the interface value therefore **will not compile**.

RECOMMENDED resolution (run-phase REQUIRED): extend the `SessionManager` interface with the one additive method `InjectSensitiveEnv(ctx context.Context, key, value string) error`. `*DefaultSessionManager` already implements it (the `var _ SessionManager = (*DefaultSessionManager)(nil)` compile-time assertion at session.go:68 still holds), so this is a pure type-surface widening with zero behavior change. This single addition is carved into scope by spec.md §F.5; the broad "no tmux API refactor" exclusion is narrowed to permit exactly this method and nothing further. The earlier draft claim that the path could call `InjectSensitiveEnv` without an interface change was incorrect at the static-type level and is retracted (see plan.md §M3).

### Testability seam
The reproduction test must observe the argv handed to `tmux set-environment`. Because `CreateTmuxSession` accepts the `tmux.SessionManager` **interface**, the test can substitute a recorder fake — but only once the interface declares the methods the fix calls. After the §F.5 interface extension above, the fake must implement all four interface methods (`Create` / `InjectEnv` / `ClearEnv` / `InjectSensitiveEnv`), recording the argv each receives. The recorder must allow asserting both "token never appears in any `InjectEnv` argv" and "no `InjectEnv`-with-token fallback when the fake's `InjectSensitiveEnv` returns an error". The seam shape (interface-fake vs. injected `run` func at the concrete-manager level) is a run-phase decision (plan.md §F M3), but the interface-fake route is the natural fit given `CreateTmuxSession` already takes the interface.

### Behavior-preservation argument
Non-sensitive vars take the identical `InjectEnv` path as before. The only behavior change is for the token: it moves from argv to the sensitive channel, and on failure the path returns an error instead of silently leaking. Non-GLM/CG mode is untouched.

## §M4 — LSP tracker data race

### Root cause
`GetBaseline` (69-84) holds `RLock` (line 71) and calls `loadBaselineLocked`, which assigns `t.baseline = &baseline` (line 128). RWMutex permits multiple concurrent readers, so two callers entering with `t.baseline == nil` both write `t.baseline` and both read `t.baseline.Files` (line 78) — an unsynchronized write-write + read. `CompareWithBaseline` (86-95) routes through `GetBaseline` and inherits it. `go vet`/lint do not catch it; only `-race` does.

### Design direction: write lock in GetBaseline
Simplest behavior-preserving fix: change `t.mu.RLock()/RUnlock()` to `t.mu.Lock()/Unlock()` in `GetBaseline`. The read path is not hot (called on hook/compare, not in a tight loop), so serializing readers is acceptable and correctness dominates. `loadBaselineLocked` already early-returns when `t.baseline != nil`, so the second caller's load is a cheap no-op under the write lock.

### Alternative: double-checked locking
Fast-path RLock read of `t.baseline != nil`; if nil, drop RLock, take Lock, re-check, load. More code, marginal throughput benefit. Run-phase MAY choose this if a benchmark shows the write-lock serialization matters; default is the plain write lock. Either way the observable contract is identical (AC-SEC-M4-003/004).

### Behavior-preservation argument
No change to return values, error types, or lazy-load semantics — only the lock type changes. `ClearBaseline` and `RecordBaseline` (which already use `Lock`) are unaffected.

## §M5 — Circuit breaker half-open permit + recovered goroutine

### Root cause
- `Call` (61-97) reads state under the lock (73-75), releases it, then runs `fn()` lock-free (83). There is no half-open in-flight permit anywhere, so N concurrent half-open callers all execute `fn()` against a recovering backend — the documented invariant at line 63 ("only one request is allowed through") is structurally absent.
- `transitionTo` (181-198) dispatches `go cb.config.OnStateChange(oldState, newState)` (196) with no `recover()`; the code's own `@MX:WARN` (193-194) documents that a panic crashes the process.

### Design direction
**Half-open permit**: add a `halfOpenInFlight bool` guarded by `cb.mu`. In `Call`, under the lock where state is checked: if `state == StateHalfOpen`, inspect `halfOpenInFlight` — if true, count a rejection and return `ErrCircuitOpen`; if false, set `halfOpenInFlight = true` and proceed. After `fn()` resolves, under the post-call lock, set `halfOpenInFlight = false`. The guard must also be cleared on any transition out of half-open (success→closed, failure→open) so it cannot leak a stuck permit. Run-phase decides the exact field/clear-site placement (atomic bool vs. mutex-guarded bool) per spec.md §F.4; the mutex-guarded bool integrates cleanly with the existing lock discipline and is the proposed default.

**Recovered goroutine**: wrap the goroutine body:
```
go func() {
    defer func() { _ = recover() }()
    cb.config.OnStateChange(oldState, newState)
}()
```
Update the `@MX:WARN` (193-194) to reflect the recover wrapper (or demote to `@MX:NOTE` per mx-tag-protocol since the danger is mitigated).

### RED test soundness for the panicking callback (D3)
A panic in an unrecovered goroutine (`go cb.config.OnStateChange(...)` at circuit.go:196 pre-fix) propagates to the Go runtime and aborts the **entire `go test` process** — it cannot be caught by the spawning goroutine's `recover()`, by `defer`, or by any in-test assertion. Therefore a pre-fix RED test that "asserts the process crashes" is **not soundly automatable**: the crash takes the test binary down with it, so the assertion line never executes.

Resolution: split M5's panic concern across two AC of different kinds.
- The pre-fix panicking-callback crash (AC-SEC-M5-004) is an **OBSERVATIONAL RED**: run once manually against the pre-fix code, observe the test-binary crash, and record the observation in `progress.md` per the Quality Gate checklist ("RED reproduction confirmed to FAIL pre-fix"). It is NOT committed as an automated assertion (it would abort the suite).
- The post-fix recover-path behavior (AC-SEC-M5-005) is the **AUTOMATED AC**: with the recover wrapper in place, a panicking `OnStateChange` is caught inside the goroutine, the test completes normally, and the test asserts (a) the test process survived and (b) the breaker's post-transition state is correct. This is the committed, repeatable assertion.

This keeps the half-open permit ACs (M5-001/002/003) fully automated under `-race`, while the panic-recovery concern is verified by the post-fix automated AC (M5-005) plus a one-time observational RED note (M5-004).

### Trade-offs
- The half-open permit serializes only the half-open trial selection (a brief lock hold), not `fn()` execution. The single admitted call still runs `fn()` lock-free, preserving the existing non-blocking execution model for the trial.
- Keeping `OnStateChange` async (only adding recover) preserves the existing timing/non-blocking contract (REQ-SEC-M5-004) — making it synchronous would change observable blocking behavior and is rejected.
- `Reset()` (108-121) calls `OnStateChange` **synchronously** (line 119), not via goroutine. That path is out of M5's panic-recovery scope (different call site, synchronous by design); M5 touches only the `transitionTo` goroutine. If run-phase wishes to also guard the `Reset` synchronous call it MUST justify it as in-scope; default is to leave `Reset` unchanged (AC-SEC-M5-006 requires `Reset` path unchanged).

### Behavior-preservation argument
Closed/open threshold counting, timeout promotion, and metrics are untouched. The only new rejections are the N-1 concurrent half-open callers (which the invariant always intended to reject). The recover wrapper changes nothing observable except that a panicking callback no longer crashes the process.

## §Cross-milestone notes

- All five milestones are code-independent and can be implemented/committed separately.
- M3 is the only milestone touching a CLI package; the CLI subagent boundary (no `AskUserQuestion`) is unaffected because M3 changes only env-injection plumbing.
- M2 introduces a filesystem write (`permission.log`); tests must use `t.TempDir()`-rooted log dirs per CLAUDE.local.md §6 — no writes to the real project tree during tests.
