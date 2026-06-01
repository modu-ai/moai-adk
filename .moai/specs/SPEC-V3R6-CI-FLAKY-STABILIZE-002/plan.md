# Implementation Plan — SPEC-V3R6-CI-FLAKY-STABILIZE-002

> Tier S (LEAN). Test-only stabilization of 2 contention-driven flaky tests. ZERO production-code change.
> cycle_type recommendation: `tdd` is not the right framing here (no new behavior); treat as a **targeted test-stabilization patch** — apply the recommended edit, then run the AC verification harness. Manager-develop run-phase note: this is a 2-file, < 20-LOC change; no Section A-E full template required (Tier S minimal form).

## §A — Context

- **Predecessor**: `SPEC-V3R6-CI-FLAKY-STABILIZE-001` (status: completed). Its §C deferred "the 3 pre-existing flaky tests' broader refactor". This SPEC covers 2 of those 3; the 3rd stays out of scope.
- **Target files (the ONLY two files this SPEC modifies)**:
  - `internal/lsp/subprocess/supervisor_test.go` (FLAKY-1)
  - `internal/hook/quality/gate_test.go` (FLAKY-2)
- **PRESERVE (byte-identical to HEAD — DO NOT TOUCH)**:
  - `internal/lsp/subprocess/supervisor.go`
  - `internal/lsp/.../client.go` (the sole production caller of `Watch`)
  - `internal/hook/quality/gate.go`
  - `internal/hook/quality/gate_test.go` :: `TestQualityGate_SkipsDotnetFormatWhenNoCSharpStaged` (immune sibling — do not edit this function)
- **Root-cause verdict**: both are TEST-ARTIFACTS. Each test binds its assertion to an arbitrary 5s wall-clock unrelated to what it verifies; under whole-suite CPU/scheduler contention the per-test subprocess cannot fork/exec-and-run within 5s, so the deadline fires and the test fails (while passing in isolation).

## §B — Known Issues / Risk Register (filtered to Tier S relevance)

| Risk | Category | Mitigation (binding AC) |
|------|----------|-------------------------|
| Converting a timeout into Skip (FLAKY-1) could conceal a genuine production hang where the subprocess never exits / `Watch` never delivers | MODERATE | AC-CFS2-002 — an exit-0 stub still produces a deterministic FAIL ("ExitCode = 0, want 1"). The comma-ok guard skips ONLY the closed-empty (`ok == false`) case; a real `ExitEvent` (`ok == true`) still hits the assertion. |
| Raising a timeout (FLAKY-2) is the classic move that can hide a real hang | LOW | The timed subprocess is a deterministic fake exit-0 shell script (not real tooling), and production already budgets 60s for this step (`gate.go:46`). 60s is pure scheduling headroom, not masking of real tooling slowness. Routing assertion `passed == true` is preserved (AC-CFS2-005). |
| Accidentally editing production code | HARD | AC-CFS2-006 — `git diff --stat` shows ONLY the two `*_test.go` files; `supervisor.go` / `client.go` / `gate.go` byte-identical to HEAD. |
| Accidentally editing the immune sibling | Scope | AC-CFS2-005 / REQ-CFS2-008 — `TestQualityGate_SkipsDotnetFormatWhenNoCSharpStaged` must remain unchanged. |
| Over-engineering (broad test refactor) | Anti-pattern | Smallest robust change only — comma-ok + skip (FLAKY-1), single-literal timeout raise (FLAKY-2). No `t.Parallel()` topology change, no fixture restructure. |

## §C — Pre-flight Check (run-phase, before editing)

```bash
# 1. baseline build + the two affected packages green in isolation
go build ./...
go test ./internal/lsp/subprocess/ -count=1
go test ./internal/hook/quality/ -run Dotnet -count=1

# 2. capture HEAD hashes of the 3 PRESERVE production files (compare after edit)
git rev-parse HEAD:internal/lsp/subprocess/supervisor.go
git rev-parse HEAD:internal/hook/quality/gate.go
# client.go path: confirm exact path before capture
ls internal/lsp/core/ 2>/dev/null || ls internal/lsp/ | grep -i client
```

## §D — Constraints (DO NOT VIOLATE)

- **Test-only**: modify ONLY the two `*_test.go` files. `supervisor.go`, `client.go`, `gate.go` byte-identical to HEAD.
- **Preserve deterministic production-behavior assertions**: FLAKY-1 keeps "ExitCode = N, want M" on a delivered event; FLAKY-2 keeps `passed == true` routing assertion.
- **Do not modify the immune sibling** `TestQualityGate_SkipsDotnetFormatWhenNoCSharpStaged`.
- **Anti-over-engineering**: smallest robust change; no broad refactor; no `t.Parallel()` topology change.
- Conventional Commits, `🗿 MoAI` trailer. No `--no-verify`.

## §E — Self-Verification (manager-develop completion checklist)

Bind to acceptance.md AC matrix (AC-CFS2-001 .. AC-CFS2-007). Report the AC PASS/FAIL table with actual command output.

## §F — Milestones (priority-ordered, no time estimates)

### M1 — FLAKY-1: comma-ok receive + load-attribution skip (Priority High)

**File**: `internal/lsp/subprocess/supervisor_test.go`

**Recommended fix** (masks_or_fixes = **fixes-root-cause**): comma-ok receive + load-induced-timeout `t.Skip`, applied to BOTH `TestSupervisor_NonZeroExit` and `TestSupervisor_NormalExit`.

- `TestSupervisor_NonZeroExit` (lines 124-127): change `ev := <-ch` (line 124) to `ev, ok := <-ch`, and add immediately before the existing assertion:
  ```go
  if !ok {
      t.Skip("Watch channel closed without ExitEvent: ctx deadline fired before stub exit under load (no production impact: client.go ignores ExitEvent value)")
  }
  ```
- `TestSupervisor_NormalExit` (lines 105-108): identical change — `ev := <-ch` (line 105) → `ev, ok := <-ch` + the same `if !ok { t.Skip(...) }` guard before its `ExitCode != 0` assertion.

**Why it works**: the comma-ok signal is the exact channel signal `Watch()` already uses to mean "no event delivered (ctx-cancel close)". The closed-empty (timeout) channel becomes an explicit `Skip`; a genuine wrong `ExitCode` (stub really exits 0 when 1 expected) still delivers a real `ExitEvent` (`ok == true`) and fails the assertion deterministically.

**Scope**: ~2 lines added + 1 changed per test. `supervisor.go`: NO CHANGE.

**Verify**: AC-CFS2-001, AC-CFS2-002, AC-CFS2-003.

### M2 — FLAKY-2: timeout SLA alignment 5s → 60s (Priority High)

**File**: `internal/hook/quality/gate_test.go`, function `TestQualityGate_RunsDotnetFormatWhenCSharpStaged`

**Recommended fix** (masks_or_fixes = **fixes-root-cause**): raise the `executeStep` timeout argument and align the inert config field with the production SLA.

- Line 724: `g.executeStep(context.Background(), step, 5*time.Second)` → `g.executeStep(context.Background(), step, 60*time.Second)`.
- Line 720: `LintTimeout: 5 * time.Second` → `LintTimeout: 60 * time.Second` (mirrors `DefaultGateConfig.LintTimeout = 60s` at `gate.go:46`; the field is inert but updating it removes the misleading 5s literal).

**Why it works**: the test verifies ROUTING (`changedExts=[".cs"]` causes dotnet to execute), not timeout/perf. The 5s was arbitrary scaffolding. 60s mirrors the production SLA and gives the deterministic fake exit-0 shell-script scheduling headroom under saturation. Cannot mask a real bug because the timed subprocess is a fake exit-0 script and production already budgets 60s for this step.

**Scope**: single-literal change (plus the inert-field alignment). `gate.go`: NO CHANGE. Immune sibling `TestQualityGate_SkipsDotnetFormatWhenNoCSharpStaged`: NO CHANGE.

**Verify**: AC-CFS2-004, AC-CFS2-005.

### M3 — Green-gate + no-production-diff verification (Priority High)

- AC-CFS2-006: `git diff --stat` shows ONLY `supervisor_test.go` + `gate_test.go`; production files byte-identical.
- AC-CFS2-007: full-suite stability (`go test ./...`, or at minimum the two affected packages under `-count=5`).

## §G — Anti-Patterns to Avoid

- Editing `supervisor.go` / `client.go` / `gate.go` "to be safe" — the production code is correct; any prod edit is a scope violation.
- Skipping the FLAKY-1 deterministic-assertion guard (turning EVERY receive into a skip) — would mask a genuine wrong exit code.
- Editing the immune sibling `TestQualityGate_SkipsDotnetFormatWhenNoCSharpStaged`.
- Broad `t.Parallel()` topology rework or fixture restructuring (Tier S = smallest robust change).
- Raising the FLAKY-2 timeout to an arbitrary larger value than the production SLA — 60s is chosen specifically to mirror `gate.go:46`, not as a generic "make it bigger" hack.

## §H — Cross-References

- Predecessor: `.moai/specs/SPEC-V3R6-CI-FLAKY-STABILIZE-001/spec.md` §C (deferred-3 Out-of-Scope, verbatim cross-referenced in this SPEC §A and §C).
- `internal/lsp/subprocess/supervisor.go` lines 70-91 (Watch two-arm select + godoc contract), lines 107-109 (`buildExitEvent` zero-value).
- `internal/hook/quality/gate.go` line 46 (`DefaultGateConfig.LintTimeout = 60s`), lines 487-509 (`runStep` DeadlineExceeded branch).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability (Tier S minimal-form permission).
