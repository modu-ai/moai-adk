---
id: SPEC-OBS-ENABLED-GATE-001
title: "Observability enabled-key gate in deps.go"
version: "0.1.0"
status: in-progress
created: 2026-08-02
updated: 2026-08-02
author: manager-spec
priority: Medium
phase: v3.0.2
module: internal/cli
lifecycle: spec-anchored
tags: "observability, config, contract-violation, deps-go, cohabitation"
related_specs:
  - SPEC-HOOK-TRACE-FLUSH-001
  - SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001
  - SPEC-V3R6-HOOK-ASYNC-EXPAND-001
tier: S
---

# SPEC-OBS-ENABLED-GATE-001 — Observability `enabled`-key gate in deps.go

## HISTORY

- **2026-08-02** v0.1.0 — initial plan-phase draft (manager-spec). Identified as a follow-on contract gap exposed by SPEC-HOOK-TRACE-FLUSH-001: once the TraceWriter actually flushes, the deps.go `os.Stat`-only gate produces observable contract violations (`enabled: false` still writes traces).

## §A. Problem

`internal/cli/deps.go:245-259` `enableObservabilityIfConfigured(reg, cwd)` gates TraceWriter creation on **file existence only** (`os.Stat`). It never reads the `enabled` key documented in `.moai/config/sections/observability.yaml`. The shipped template sets `enabled: true`, so the contract is silently satisfied for the common path — but any user who sets `enabled: false` (the documented opt-out) still gets a TraceWriter created and trace files written. This is a contract violation that became observable the moment SPEC-HOOK-TRACE-FLUSH-001 made the TraceWriter actually flush.

The cohabitation note in `internal/hook/observability.go:16` and `internal/hook/observability_master.go:21` (both carry the literal line "Do NOT unify without a fresh SPEC.") protects THREE independent yaml-key read paths (the 3 `system.yaml` / `observability.yaml` keys enumerated in §D below). `deps.go`'s `os.Stat` gate is a 4th, cruder path — NOT one of the 3 protected reads. This SPEC is the "fresh SPEC" the cohabitation note requires before touching it.

## §B. Scope

**In scope:**

- `internal/cli/deps.go` `enableObservabilityIfConfigured` — add an `enabled`-key read so file-present + `enabled: false` does NOT enable observability.
- `internal/cli/deps_coverage_test.go` — add the missing regression test (`enabled: false` → not enabled).

**Out of scope:**

- `internal/hook/observability_master.go` `IsObservabilityEnabled()` implementation (the canonical `observability.yaml` reader) — reused as-is.
- The 3 cohabitation-protected yaml-key reads (see §D) — UNTOUCHED.
- TraceWriter internals, flush behavior, log rotation — owned by SPEC-HOOK-TRACE-FLUSH-001 and related.

### Out of Scope — Cohabitation-protected reads

The three yaml-key read paths documented in the cohabination note on `internal/hook/observability.go` / `observability_master.go` are NOT touched by this SPEC:

- `observabilityOptIn()` — reads `system.yaml` `hook.observability_events` (SPEC-V3R2-RT-006 REQ-040, per-event whitelist)
- `hookOptInEnabled()` — reads `system.yaml` `hook.opt_in.enabled` (SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 REQ-HOI-001 master)
- `IsObservabilityEnabled()` — reads `observability.yaml` `observability.enabled` (REQ-OBS-005 trace master)

`deps.go`'s gate is a 4th path. This SPEC does NOT unify it with the 3 protected reads; it reuses `IsObservabilityEnabled()` as the single source of truth for the `enabled` key while leaving the 3-way cohabitation invariant intact.

### Out of Scope — Structural unification

Option (b) from the design fork (remove the deps.go gate entirely and move TraceWriter creation under `IsObservabilityEnabled()` in package `hook`) is explicitly deferred. Scope is larger and risks the cohabitation boundary; this SPEC takes the surgical Tier-S path (option a). Escalation to Tier M would require a separate SPEC.

### Out of Scope — Absent-key behavior change at the template level

The shipped template sets `enabled: true`. This SPEC aligns deps.go's absent-key default with `IsObservabilityEnabled()` (absent → disabled). The template's explicit `enabled: true` is unaffected; users who removed the key but relied on file-presence = enabled are documented as the low-risk behavior-change surface in plan.md §D.

## §C. Requirements (GEARS)

**REQ-OEG-001** (Ubiquitous) The `enableObservabilityIfConfigured` function shall treat the `observability.enabled` key in `.moai/config/sections/observability.yaml` as the authoritative gate for `EnableObservability` invocation, such that the function enables observability only when the key is present and truthy.

**REQ-OEG-002** (Capability gate) **Where** `.moai/config/sections/observability.yaml` exists with `observability.enabled: false`, the `enableObservabilityIfConfigured` function shall NOT call `EnableObservability` and shall NOT create the trace-log directory.

**REQ-OEG-003** (Capability gate) **Where** `.moai/config/sections/observability.yaml` exists with `observability.enabled: true`, the `enableObservabilityIfConfigured` function shall preserve the current enabling behavior — call `EnableObservability(logDir)` after ensuring the log directory exists.

**REQ-OEG-004** (State-driven) **While** `.moai/config/sections/observability.yaml` exists but the `observability.enabled` key is absent, the `enableObservabilityIfConfigured` function shall treat the configuration as disabled (safe-default false), aligning with the `IsObservabilityEnabled()` absent-key semantics.

**REQ-OEG-005** (Event-driven) **When** `.moai/config/sections/observability.yaml` does not exist on disk, the `enableObservabilityIfConfigured` function shall preserve the current skip behavior and return without enabling observability.

**REQ-OEG-006** (Unwanted) The `enableObservabilityIfConfigured` function shall NOT read `system.yaml`'s `hook.observability_events` or `hook.opt_in.enabled` keys — those are the cohabitation-protected reads owned by `observabilityOptIn()` and `hookOptInEnabled()` respectively.

**REQ-OEG-007** (Ubiquitous) The `enableObservabilityIfConfigured` function shall delegate the `enabled`-key read to `hook.IsObservabilityEnabled()` as the single source of truth, rather than introducing a fifth independent yaml-key reader.

**REQ-OEG-008** (Unwanted) This SPEC shall NOT alter the `AC-HOI-007` 4-quadrant cohabitation test — `system.yaml` reads, the async-expand dual-gate, and the 3 protected yaml-key readers remain on their current behavior.

## §D. Cohabitation Context (informational — restates the permanent regression guard)

The cohabitation note in `internal/hook/observability.go:16` and `internal/hook/observability_master.go:21` (literal token: "ALL 3 KEYS ARE INDEPENDENT. Do NOT unify without a fresh SPEC." / "ALL 3 READ PATHS ARE INDEPENDENT. Do NOT unify without a fresh SPEC.") documents THREE independent yaml-key read paths and marks them "Do NOT unify without a fresh SPEC." `AC-HOI-007` is the 4-quadrant cohabitation test that makes the invariant falsifiable. This SPEC IS the fresh SPEC for the 4th path (deps.go's `os.Stat` gate), but REQ-OEG-006 and REQ-OEG-008 bind this SPEC to leave the 3 protected reads and the 4-quadrant test intact. The single source of truth introduced in REQ-OEG-007 reuses the existing `IsObservabilityEnabled()` reader — no new reader, no unification of the 3 protected paths.

## §E. Acceptance Criteria Matrix (summary — full GWT in acceptance.md)

| AC-ID      | Requirement   | Summary                                                                  |
|------------|---------------|--------------------------------------------------------------------------|
| AC-OEG-001 | REQ-OEG-002   | file present + `enabled: false` → `EnableObservability` NOT called       |
| AC-OEG-002 | REQ-OEG-003   | file present + `enabled: true` → `EnableObservability` called (preserve) |
| AC-OEG-003 | REQ-OEG-004   | file present + key absent → NOT called (safe-default false)              |
| AC-OEG-004 | REQ-OEG-005   | file missing → NOT called (preserve skip)                                |
| AC-OEG-005 | REQ-OEG-006/008| `AC-HOI-007` 4-quadrant test remains green; `system.yaml` reads untouched |
| AC-OEG-006 | REQ-OEG-007   | deps.go delegates the read to `hook.IsObservabilityEnabled()`            |
| AC-OEG-007 | Regression    | falsifiable test `TestEnableObservabilityIfConfigured_DisabledByEnabledKey` with sabotage round-trip |

## §F. Constraints

- **No import cycle.** `cli` already imports `hook` — verified by the existing call sites in deps.go. No new import direction.
- **sync.Once cache reconciliation.** `IsObservabilityEnabled()` is `sync.Once`-cached per process and resolves cwd via `CLAUDE_PROJECT_DIR` then `os.Getwd()`. `InitDependencies` runs once at startup with `cwd` passed in (the project root), so the cache is consistent with the caller's cwd at startup time. See plan.md §D for the reconciliation analysis.
- **Tier S ceiling.** Surgical: one gate function reads one more yaml key (via reuse) + one regression test + cohabitation-preservation note. If cross-package reuse forces broader changes, escalate per plan.md §G.

## §G. Cross-References

- **SPEC-HOOK-TRACE-FLUSH-001** — made the TraceWriter actually flush, which is what made this contract violation observable. This SPEC is its follow-on contract-gap fix.
- **SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001** — owns `hookOptInEnabled()` (`system.yaml` `hook.opt_in.enabled`), one of the 3 protected reads. REQ-HOI-001 / AC-HOI-007 (4-quadrant cohabitation test).
- **SPEC-V3R6-HOOK-ASYNC-EXPAND-001** — owns `IsObservabilityEnabled()` (REQ-HAE-003/004) — the canonical reader this SPEC reuses.
- `.moai/config/sections/observability.yaml` — the documented `enabled` key this SPEC enforces.
- `internal/cli/deps.go:245-259` — the buggy gate.
- `internal/cli/deps_coverage_test.go` — the test file with the coverage gap.

## §H. Out-of-Scope Restatement

See §B's three `### Out of Scope —` sub-headings. This section cross-references them per the `OutOfScopeRule` lint convention; no scope is silently omitted.
