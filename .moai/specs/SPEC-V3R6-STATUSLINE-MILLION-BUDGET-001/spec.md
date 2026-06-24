---
id: SPEC-V3R6-STATUSLINE-MILLION-BUDGET-001
title: "Statusline memory_test AutoCompactScaling model-env isolation"
version: "0.2.0"
status: completed
created: 2026-06-18
updated: 2026-06-18
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/statusline"
lifecycle: spec-anchored
tags: "statusline, test-fixture, model-env-isolation, debt-cleanup"
---

# SPEC-V3R6-STATUSLINE-MILLION-BUDGET-001

## §A Problem Statement

The `TestCollectMemory_AutoCompactScaling` table-driven test in `internal/statusline/memory_test.go` fails when executed in a developer environment where ambient `ANTHROPIC_DEFAULT_OPUS_MODEL` (or the sonnet/haiku sibling env vars) contains a `glm-5.2`-family model identifier. Five subtests pass `ContextWindowSize: 200000` via `StdinData`, expecting the 200K-derived `TokenBudget` and `TokensUsed` values (e.g. `wantBudget: 170000` for the 85%-threshold cases). The production resolver `resolveContextWindowOverride()` at `internal/statusline/memory.go:139` consults the ambient model env vars and, on a `glm-5.2` match, returns `1_000_000` (per the built-in `glmContextWindows` table at `memory.go:28`). This 1M override silently replaces the test's 200K input at `memory.go:206-207`, producing `TokensUsed = 850000, want 170000` failures at `memory_test.go:193` / `:196`.

This is a **test-fixture isolation defect**, not a production resolver defect. The production override behavior is correct and was introduced deliberately (Issue #653 — GLM "opus" slot reported as 1M but real limit is 128-230K). The test was authored under the implicit assumption that the ambient process environment would not contaminate the `ContextWindowSize` input; that assumption is violated on maintainer machines where `moai glm` / `moai cg` have populated the model env vars into the shell or `settings.local.json`.

**Root cause (verified, not re-derived this session)**: `resolveContextWindowOverride()` priority chain (env explicit override → llm.yaml → built-in table → stdin fallback) reads `os.Getenv` on `ANTHROPIC_DEFAULT_*_MODEL` at test execution time. The table-driven test does not neutralize these vars, so the built-in-table branch (priority 3) fires under GLM-ambient conditions and overrides the `StdinData.ContextWindowSize` input.

**Pre-existing scope**: The failure is reproducible on commit `bef24877d` (pre-NAMESPACE-V2). It is tracked as AC-HNS-011 PASS-WITH-DEBT in the closed SPEC-V3R6-HARNESS-NAMESPACE-V2-001. This SPEC retires that debt.

## §B Why Direction B (test isolation) over Direction A (update expectations to 1M)

Two candidate fix directions were considered:

- **Direction A (update expectations to 1M)**: Rewrite the 5 subtests' `wantUsed` / `wantBudget` to the 1M-derived values (e.g. `wantBudget: 850000`). This codifies the ambient override into the test's expected output.
- **Direction B (test isolation)**: Add `t.Setenv("ANTHROPIC_DEFAULT_OPUS_MODEL", "")` (plus sonnet + haiku siblings) at the top of each affected subtest (or via a small helper) so `resolveContextWindowOverride()` returns 0 on priority-3 lookup and the existing 200K expectations hold as the test author intended.

**Direction B is adopted.** Rationale:

1. **Durability across model renames.** Direction A hardcodes the 1M figure; the next time the GLM model identifier changes (e.g. `glm-5.2` → `glm-5.3[1m]` or a context-size table revision), Direction A's expectations re-stale. Direction B's isolation makes the test invariant to the model-env state of the machine it runs on.
2. **Test author intent preservation.** The test's prose doc-comment (`"TokenBudget is scaled to the auto-compact threshold so the CW bar shows 100% at compact point"`) and the deliberately varied `ContextWindowSize: 200000` input make clear that the author wanted to verify the threshold-scaling math at a fixed 200K baseline. Direction A would rewrite that intent; Direction B honors it.
3. **Minimal blast radius.** Direction B touches only the test file. Production `memory.go` is untouched, preserving the Issue #653 override contract that real GLM users depend on for accurate statusline context-window reporting.

The trade-off — Direction B slightly obscures the fact that the production resolver consults ambient env — is acceptable because that behavior is already covered by the dedicated `TestCollectMemory_*` override-priority tests at `memory_test.go:276+` / `:350+` / `:415+` / `:450+` which exercise the env/winover paths explicitly and intentionally.

## §C Requirements (GEARS)

**REQ-SMB-001** (Ubiquitous): The `TestCollectMemory_AutoCompactScaling` test shall neutralize the ambient `ANTHROPIC_DEFAULT_OPUS_MODEL`, `ANTHROPIC_DEFAULT_SONNET_MODEL`, and `ANTHROPIC_DEFAULT_HAIK_MODEL` environment variables within each subtest such that `resolveContextWindowOverride()` returns 0 and the test's `StdinData.ContextWindowSize` input is the sole source of the context-size baseline.

**REQ-SMB-002** (State-driven): **While** the developer's shell or `settings.local.json` populates any `ANTHROPIC_DEFAULT_*_MODEL` env var with a `glm-5.2`-family identifier, the `TestCollectMemory_AutoCompactScaling` subtests shall produce the `wantUsed` / `wantBudget` / `wantPctApprox` values derived from the `ContextWindowSize: 200000` input (unchanged from the pre-GLM-ambient expectations).

**REQ-SMB-003** (Capability gate): **Where** a future contributor adds a new subtest to `TestCollectMemory_AutoCompactScaling`, that subtest shall inherit the same model-env isolation discipline (either by reusing a shared helper or by invoking the isolation inline), so the 200K-derived expectation contract remains invariant to the contributor's ambient environment.

**REQ-SMB-004** (Unwanted behavior, GEARS event-detected form): **When** the model-env isolation is absent or removed from any `TestCollectMemory_AutoCompactScaling` subtest, the Go test runner shall fail the assertion at `memory_test.go:193` (`TokensUsed`) or `:196` (`TokenBudget`) under any GLM-ambient execution environment, surfacing the regression rather than silently passing.

**REQ-SMB-005** (Ubiquitous): The production resolver `resolveContextWindowOverride()` and its callers in `internal/statusline/memory.go` shall remain unmodified by this SPEC — the Issue #653 GLM context-window override contract is out of scope and must continue to behave identically before and after the test-only fix.

## §D Acceptance Criteria (Given-When-Then, inline per Tier S LEAN)

**AC-SMB-001** — model-env isolation applied to all 5 AutoCompactScaling subtests
> **Given** the `TestCollectMemory_AutoCompactScaling` table-driven test in `internal/statusline/memory_test.go` with subtests: (1) `"default threshold 85%: 83% used → ~97% display"`, (2) `"default threshold 85%: 85% used → 100% display"`, (3) `"threshold 90%: 83% used → 92% display"`, (4) `"threshold 100% (no scaling)"`, (5) `"exceeded threshold capped at 100%"`;
> **When** a developer runs `go test ./internal/statusline/ -run TestCollectMemory_AutoCompactScaling` on a machine where `ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5.2` (or sonnet/haiku siblings) is exported in the shell;
> **Then** all 5 subtests PASS with the existing 200K-derived expectations (`wantBudget` ∈ {170000, 170000, 180000, 200000, 160000}; `wantUsed` ∈ {166000, 170000, 166000, 166000, 170000}; `wantPctApprox` ∈ {97, 100, 92, 83, 100}) — no subtest emits `TokensUsed = 850000, want 170000` or `TokenBudget = 850000, want 170000`.

**AC-SMB-002** — isolation mechanism is env-neutralization, not expectation rewrite
> **Given** the fix applied per REQ-SMB-001;
> **When** a reviewer inspects the diff of `internal/statusline/memory_test.go`;
> **Then** the diff shows `t.Setenv(..., "")` calls (or a shared helper that wraps them) neutralizing `ANTHROPIC_DEFAULT_OPUS_MODEL` / `SONNET_MODEL` / `HAIKU_MODEL` — and does **not** show any change to the `wantUsed` / `wantBudget` / `wantPctApprox` table values (they remain the 200K-derived originals).

**AC-SMB-003** — production code unchanged
> **Given** the fix merged;
> **When** a reviewer runs `git diff bef24877d..HEAD -- internal/statusline/memory.go`;
> **Then** the diff is empty — `resolveContextWindowOverride()`, the `glmContextWindows` table, the default `200000` constant at `memory.go:204`, and the override-application block at `:206-207` are byte-identical to the pre-fix baseline.

**AC-SMB-004** — no regression in sibling override-priority tests
> **Given** the fix merged;
> **When** a developer runs `go test ./internal/statusline/...`;
> **Then** the override-priority tests (`TestCollectMemory_*` at `memory_test.go:276+`, `:350+`, `:415+`, `:450+` that intentionally exercise env/llm.yaml/explicit-override winover) continue to PASS — the model-env isolation added to `TestCollectMemory_AutoCompactScaling` does NOT contaminate the sibling tests' intentional env-population paths (isolation is scoped to the AutoCompactScaling test via `t.Setenv`'s test-lifetime semantics, not process-global).

**AC-SMB-005** — durability: isolation survives a simulated model rename
> **Given** the fix applied;
> **When** a developer temporarily edits the `glmContextWindows` table in `memory.go` to map a hypothetical `"glm-99.9"` to `2_000_000`, exports `ANTHROPIC_DEFAULT_OPUS_MODEL=glm-99.9`, and re-runs `go test ./internal/statusline/ -run TestCollectMemory_AutoCompactScaling`;
> **Then** all 5 subtests still PASS with the 200K-derived expectations — confirming the isolation is keyed on "any model-env value present" rather than on a specific model identifier, and would not re-stale if the GLM model name changes in the future. (This AC is verified manually during run-phase sanity check, then the table edit is reverted before commit.)

## §E Constraints

- **Tier S LEAN**: ACs are inline in this spec.md §D; no separate `acceptance.md` artifact is emitted for this SPEC.
- **Test-only change**: production `internal/statusline/memory.go` MUST NOT be modified (REQ-SMB-005, AC-SMB-003).
- **`t.Setenv` discipline**: the isolation MUST use `t.Setenv` (not `os.Unsetenv` + deferred restore) so the test-runtime guarantees automatic restoration and parallel-test safety. `t.Setenv` with an empty string value is the canonical Go idiom for "neutralize this env var for the lifetime of this test."
- **No `t.Parallel()` introduction**: the existing test does not call `t.Parallel()`; this SPEC does not introduce it (out of scope). `t.Setenv` forces the test to be non-parallel at the runtime level anyway, so this constraint is naturally satisfied.
- **No new dependencies**: the fix uses only stdlib `testing` + `os` facilities already imported. No new imports of third-party env-test libraries.

## §F Exclusions (What NOT to Build)

### Out of Scope — production code, Direction A, sibling tests, and structural refactor

- **NOT in scope**: modifying `resolveContextWindowOverride()` in production code. The GLM override contract (Issue #653) is correct and ships to real users.
- **NOT in scope**: rewriting the `wantUsed` / `wantBudget` expectations of the 5 subtests to 1M-derived values (Direction A — explicitly rejected in §B).
- **NOT in scope**: adding model-env isolation to the sibling override-priority tests (`TestCollectMemory_*` at `:276+` / `:350+` / `:415+` / `:450+`). Those tests **intentionally** populate model env to exercise the override chain; isolating them would defeat their purpose.
- **NOT in scope**: adding `t.Parallel()` to the test, or refactoring the table-driven structure into separate `Test*` functions.
- **NOT in scope**: changing the `glmContextWindows` built-in table or the `1_000_000` constant.
- **NOT in scope**: any change to `.moai/config/sections/llm.yaml` or `settings.local.json` model configuration.

## §G Cross-References

- **Debt origin**: AC-HNS-011 (PASS-WITH-DEBT) of closed SPEC-V3R6-HARNESS-NAMESPACE-V2-001.
- **Production resolver**: `internal/statusline/memory.go:139` (`resolveContextWindowOverride`), `:28` (`glmContextWindows` table), `:204` (default `200000`), `:206-207` (override application).
- **Failing assertions**: `internal/statusline/memory_test.go:193` (`TokensUsed`), `:196` (`TokenBudget`).
- **5 affected subtests**: `internal/statusline/memory_test.go:119-183` (table entries), `:186-204` (runner).
- **Pre-existing baseline commit**: `bef24877d` (verified — NOT a NAMESPACE-V2 regression).
- **Original issue**: Claude Code Issue #653 (GLM "opus" slot context-window reporting).

## §H Out-of-Scope Risk Notes

- **Risk**: a future contributor adds a 6th subtest to `TestCollectMemory_AutoCompactScaling` without inheriting the isolation helper, reintroducing the GLM-ambient failure. **Mitigation**: REQ-SMB-003 + AC-SMB-002 make the isolation mechanism explicit in the diff; a code review pass is the primary defense. A grep-based lint rule ("any `t.Run` inside `TestCollectMemory_AutoCompactScaling` MUST call the isolation helper") is conceivable but out of scope for this Tier S SPEC.
- **Risk**: `t.Setenv` semantics change in a future Go version. **Mitigation**: `t.Setenv` is a stable stdlib API since Go 1.17; this risk is theoretical and not actionable.

## HISTORY

- **2026-06-18** (v0.1.0, manager-spec): plan-phase draft authored. Direction B adopted per orchestrator decision (test isolation over expectation rewrite). Retires AC-HNS-011 PASS-WITH-DEBT of SPEC-V3R6-HARNESS-NAMESPACE-V2-001. Run-phase + plan-auditor + Implementation Kickoff deferred to follow-up session.
