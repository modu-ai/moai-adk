# SPEC-OBS-ENABLED-GATE-001 — plan.md

> Implementation plan. Plan-phase only — no code is implemented in this phase.

## §A. Context

`internal/cli/deps.go:245-259` gates TraceWriter creation on `os.Stat` of `observability.yaml`. It never reads the `enabled` key documented in the same file. SPEC-HOOK-TRACE-FLUSH-001 made the TraceWriter flush, which made the contract violation observable: `enabled: false` now still produces trace files. This SPEC closes the contract gap surgically by reusing `hook.IsObservabilityEnabled()` (the canonical `enabled`-key reader) as the single source of truth inside deps.go.

Two independent mechanisms currently read `observability.yaml`:

1. `internal/hook/observability_master.go` `IsObservabilityEnabled()` — reads `observability.enabled` correctly (`yaml.Unmarshal`, safe-default false, `sync.Once`-cached). Gates the TaskCreated/Notification async-expand dual-gate fast-path (SPEC-V3R6-HOOK-ASYNC-EXPAND-001 REQ-HAE-003/004).
2. `internal/cli/deps.go` `enableObservabilityIfConfigured()` — `os.Stat` only. Gates TraceWriter creation. **THIS BUG.**

The cohabitation note on `internal/hook/observability.go` / `observability_master.go` documents THREE independent yaml-key reads and binds any change to a fresh SPEC. deps.go's `os.Stat` gate is a 4th, cruder path — this SPEC IS the fresh SPEC for it.

## §B. Design Decision

**Decision: Option (a) — Tier S surgical fix.**

`enableObservabilityIfConfigured` delegates the `enabled`-key read to `hook.IsObservabilityEnabled()`. The `os.Stat` existence check stays (it gates the skip-on-missing-file behavior preserved by REQ-OEG-005), but when the file exists, the function additionally consults `IsObservabilityEnabled()` before calling `EnableObservability`.

**Why (a) over (b):**

- **(a) [Tier S, RECOMMENDED]** — one function reads one more yaml key (via reuse of the canonical reader) + one regression test. The cohabitation 3-read invariant is untouched; deps.go becomes a consumer of the existing `IsObservabilityEnabled()` rather than a new independent reader.
- **(b) [Tier M]** — remove the deps.go gate entirely and move TraceWriter creation under `IsObservabilityEnabled()` in package `hook`. Structural unification; larger blast radius; risks the cohabitation boundary; would force TraceWriter creation logic to cross packages.

**(a) wins** because the bug is a missing gate, not a structural misplacement. The deps.go path legitimately needs the project-root cwd that the caller passes in, and it legitimately owns "create the log dir + call EnableObservability". What it lacks is the `enabled` read — and the canonical reader already exists. Reusing it collapses two readers into one SSOT without touching the 3 protected reads.

## §C. Known Issues

- **Coverage gap (the bug).** `internal/cli/deps_coverage_test.go` carries `TestEnableObservabilityIfConfigured_NoConfigFile` (file missing → skip) and `TestEnableObservabilityIfConfigured_WithConfigFile` (file present, `enabled: true`). The case `file present + enabled: false` is untested. This is exactly the bug case.
- **sync.Once cache.** `IsObservabilityEnabled()` caches its result per process via `sync.Once`. `InitDependencies` runs once at process startup, so the cache is populated exactly once and reflects the startup-time cwd. The `cwd` parameter passed to `enableObservabilityIfConfigured` at the same call site is the project root, matching `IsObservabilityEnabled()`'s cwd resolution (`CLAUDE_PROJECT_DIR` then `os.Getwd()`). See §D reconciliation.
- **Test isolation (cwd/env-resolution hazard).** The new regression test in M2 MUST drive the enabled-decision via the `hook.SetObservabilityMasterForTesting(bool)` test-seam (internal/hook/observability_master.go:112) and MUST be non-parallel (no `t.Parallel()` call) with `t.Cleanup(hook.ResetObservabilityMasterForTesting)` registered. This matches the established precedent at `internal/hook/notification_test.go:87` — "Sequential (not t.Parallel) due to SetObservabilityMasterForTesting global state." — and `internal/hook/task_created_test.go:80`. The seam is required because `IsObservabilityEnabled()` (internal/hook/observability_master.go:70) takes no `cwd` parameter and resolves the project root via `CLAUDE_PROJECT_DIR` → `os.Getwd()`; writing `enabled: false` to a tempdir does NOT influence its return, so a naive tempdir-only test cannot exercise the gate. The tempdir is still used to satisfy the `os.Stat` existence gate and to provide the `logDir` join target; the master-toggle value is what drives the enabled/disabled decision.
- **Absent-key behavior change.** `IsObservabilityEnabled()` treats absent key as false (disabled). The shipped template sets `enabled: true`, so the common path is unaffected. Users who removed the key but relied on file-presence = enabled will see observability turn off. Risk is low (template ships the key explicitly) and is the documented safe-default.

## §D. Cross-package Reuse Analysis (cwd + cache reconciliation)

`hook.IsObservabilityEnabled()` resolves cwd via `CLAUDE_PROJECT_DIR` → `os.Getwd()` fallback, then `sync.Once`-caches the parse result. `cli.enableObservabilityIfConfigured(reg, cwd)` receives `cwd` as a parameter from `InitDependencies`. The two cwd sources can in principle diverge (env-var override vs. the caller-supplied project root).

Reconciliation: `InitDependencies` is invoked at process startup with the project root as `cwd`. In a `moai`/Claude Code session, `CLAUDE_PROJECT_DIR` is set to the same project root by the runtime before the moai binary starts. The `sync.Once` cache is therefore populated exactly once, at startup, against the same project root the caller passed in. There is no second invocation that could observe a stale cache against a different cwd.

**Decision:** call `hook.IsObservabilityEnabled()` from deps.go (preferred — single source of truth for the enabled read). Do NOT duplicate a minimal local parse. The cwd-mismatch theoretical risk is mitigated by the single-startup-call invariant; if a future caller invokes deps.go from a non-project-root cwd, `IsObservabilityEnabled()`'s own resolution is the authoritative one and matches the rest of the hook subsystem.

**Test isolation addendum (test-side, not production).** The §D analysis above covers production (`InitDependencies` startup path). The test side has a separate constraint: `IsObservabilityEnabled()` resolves cwd via env/getwd and therefore cannot be steered by writing `enabled` to a tempdir. The codebase already ships the test-seam for exactly this case — `hook.SetObservabilityMasterForTesting(bool)` (internal/hook/observability_master.go:112) directly mutates the cached toggle and marks the `sync.Once` done, so the file-read path is bypassed entirely; `hook.ResetObservabilityMasterForTesting()` (internal/hook/observability_master.go:126) clears both the cached value and the `sync.Once` gate for the next test. Tests MUST call `SetObservabilityMasterForTesting` before invoking the code under test and MUST register `t.Cleanup(ResetObservabilityMasterForTesting)`; tests MUST NOT call `t.Parallel()` because the seam mutates process-global state (the cached toggle + the `sync.Once`). The canonical pattern is established at `internal/hook/notification_test.go:87,128-130,154-156,184-186` and mirrored at `internal/hook/task_created_test.go:80,127-130,154-156,185-187`. The test-side seam does NOT weaken REQ-OEG-007: deps.go remains a consumer of `IsObservabilityEnabled()` (not a 5th independent yaml-key reader) — the seam merely injects the value that `IsObservabilityEnabled()` would return, so the SSOT argument is preserved.

## §E. Milestones (priority-ordered; no time estimates)

Order is by decision-reversibility — highest-change-likelihood decisions first.

### M1. Add the `enabled`-read gate to deps.go [Priority High]

- In `enableObservabilityIfConfigured`, after the `os.Stat` existence check succeeds, additionally call `hook.IsObservabilityEnabled()`; if it returns false, return without calling `EnableObservability` and without creating the log dir.
- The `os.Stat` check is preserved so that REQ-OEG-005 (file missing → skip) behavior is unchanged; `IsObservabilityEnabled()` is only consulted when the file exists (matching its own safe-default semantics for a missing file).
- No change to the `EnableObservability(logDir)` call site, the `os.MkdirAll` call, or the `observabilityEnabler` interface assertion.

### M2. Add the regression test [Priority High]

- New test `TestEnableObservabilityIfConfigured_DisabledByEnabledKey` in `internal/cli/deps_coverage_test.go` using the established test-seam form:
  - Call `hook.SetObservabilityMasterForTesting(false)` BEFORE invoking the code under test (simulating the `enabled: false` resolved state; required because `IsObservabilityEnabled()` at internal/hook/observability_master.go:70 resolves cwd via env/getwd and cannot read a tempdir).
  - Register `t.Cleanup(hook.ResetObservabilityMasterForTesting)` so the `sync.Once` cache is reset for subsequent tests.
  - Do NOT call `t.Parallel()` — the seam mutates process-global state (matching `internal/hook/notification_test.go:87` precedent: "Sequential (not t.Parallel) due to SetObservabilityMasterForTesting global state.").
  - Write `observability.yaml` (contents `observability:\n  enabled: false\n`) to a temp config tree under `tmpDir` to satisfy the `os.Stat` existence gate and provide the `logDir` join target; the tempdir does NOT drive the enabled-decision (the seam does).
  - Assert `EnableObservability` is NOT called (call count == 0) using the existing stub registry pattern already in that file.
- Companion tests use the same seam form with the injected value flipped: `TestEnableObservabilityIfConfigured_WithConfigFile` (existing, preserved) calls `SetObservabilityMasterForTesting(true)`; a new `TestEnableObservabilityIfConfigured_AbsentKeyDefaultsDisabled` calls `SetObservabilityMasterForTesting(false)` to lock REQ-OEG-004. All three are non-parallel and register `t.Cleanup(ResetObservabilityMasterForTesting)`.
- Sabotage round-trip (falsifiability): flip the implementation to ignore `enabled` (restore the `os.Stat`-only gate) and confirm the new test FAILS — under sabotage the gate ignores the injected `SetObservabilityMasterForTesting(false)` and calls `EnableObservability` anyway, so the stub-registry call-count assertion rejects it. Restore the correct implementation and confirm it PASSES. This round-trip is the falsifiability proof the test actually exercises REQ-OEG-002, not a vacuous green.

### M3. Co-existence verification [Priority Medium]

- Run the `AC-HOI-007` 4-quadrant cohabitation test in `internal/hook/` — must remain green.
- Run `go test ./internal/cli/ ./internal/hook/` — must exit 0.
- Grep-verify deps.go does NOT read `system.yaml`'s `hook.observability_events` or `hook.opt_in.enabled` (REQ-OEG-006).

### M4. Sync-phase (handed off to manager-docs) [Priority Low]

- CHANGELOG entry noting the contract-gap fix and the absent-key behavior alignment.
- Cross-reference SPEC-HOOK-TRACE-FLUSH-001 as the change that exposed the contract violation.

## §F. Anti-Patterns to Avoid

- **Do NOT introduce a 5th yaml-key reader in deps.go.** The whole point of reusing `IsObservabilityEnabled()` is to avoid proliferating readers (REQ-OEG-007). A local `yaml.Unmarshal` scoped to deps.go would be a 5th reader and would re-create the cohabitation hazard the note guards against.
- **Do NOT unify the 3 protected reads.** `observabilityOptIn()`, `hookOptInEnabled()`, `IsObservabilityEnabled()` stay independent (REQ-OEG-006, REQ-OEG-008).
- **Do NOT remove the `os.Stat` existence check.** It carries REQ-OEG-005 (file missing → skip) and it short-circuits before the `IsObservabilityEnabled()` call when the file is absent (avoiding a pointless parse). Removing it would change the missing-file behavior.
- **Do NOT make the absent-key default `true`.** Aligning with `IsObservabilityEnabled()` means absent → false (safe-default). Flipping the default would re-introduce a divergent read path.

## §G. Risks / Escalation Triggers

- **Tier escalation (S → M).** If reusing `IsObservabilityEnabled()` from deps.go surfaces an import-layer issue (e.g. the `sync.Once` cache misbehaves when `InitDependencies` is invoked from a test harness that is NOT the project root), escalate to Tier M and consider option (b). Trigger: any green test in M2/M3 that cannot be made green under option (a) without restructuring.
- **Cwd-mismatch regression.** If a future caller invokes `InitDependencies` from a cwd that is not the project root and `IsObservabilityEnabled()`'s env-based resolution disagrees, the cache could lock in a stale result. Mitigation: document in the deps.go comment that `InitDependencies` assumes the caller-supplied `cwd` IS the project root and that `CLAUDE_PROJECT_DIR` is expected to point at the same root.

## §H. Cross-References

- `internal/cli/deps.go:245-259` — the gate under fix.
- `internal/cli/deps_coverage_test.go` — the test file carrying the coverage gap.
- `internal/hook/observability_master.go` `IsObservabilityEnabled()` — the canonical reader to be reused.
- `.moai/specs/SPEC-HOOK-TRACE-FLUSH-001/` — the SPEC whose flush fix exposed this contract violation.
- `.moai/specs/SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001/` — owner of `AC-HOI-007` (4-quadrant cohabitation test).
- `.moai/config/sections/observability.yaml` — the documented `enabled` key.

---

## §P.1 Plan-phase Audit-Ready Signal

> Lettered `§P` (not `§E`) deliberately: the `§E.*` namespace is parser-load-bearing for `progress.md` (`internal/spec/era.go` `hasAnyProgressMarker` greps for literal `§E.2`/`§E.3`/`§E.4`). This `plan.md` audit-ready subsection uses `§P.*` to avoid any collision; `progress.md` will own the canonical `§E.*` sections when run-phase begins.

- **GEARS compliance:** every REQ in spec.md §C uses a GEARS pattern (Ubiquitous / When / While / Where / shall not). No residual `IF/THEN`.
- **Frontmatter:** all 12 canonical fields present; `phase: v3.0.2` is a release-target string (matches SPEC-HOOK-TRACE-FLUSH-001 / SPEC-PHASE-FIELD-VALIDATION-001 convention), NOT a workflow-stage token; `lifecycle: spec-anchored`, `tags:` is a comma-separated string, `version: "0.1.0"` is quoted (iter-2 fixes D1/D2/D6).
- **SPEC-ID:** `SPEC-OBS-ENABLED-GATE-001` — regex PASS confirmed (`^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`), no collision against existing `.moai/specs/SPEC-OBS*` / `SPEC-OBSERVE-*` directories.
- **Out of Scope:** three `### Out of Scope — <topic>` H3 sub-headings in spec.md §B (Cohabitation-protected reads / Structural unification / Absent-key behavior change at the template level), each with `-` bullets — satisfies `OutOfScopeRule`.
- **Tier:** S (surgical; escalation trigger documented in §G).
- **Acceptance:** 7 ACs enumerated in spec.md §E; full GWT scenarios + sabotage round-trip in acceptance.md. AC-OEG-001/002/003 GWT rewritten in iter-2 to reference the `SetObservabilityMasterForTesting` test-seam + non-parallel form (D3 closure: the original tempdir-only GWT was structurally non-exercisable because `IsObservabilityEnabled()` resolves cwd via env/getwd).
- **Cohabitation invariant:** REQ-OEG-006 + REQ-OEG-008 bind this SPEC to leave the 3 protected reads and `AC-HOI-007` intact.

## §P.2 Run-phase Evidence

_<pending run-phase>_

## §P.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §P.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
