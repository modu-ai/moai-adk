# SPEC-OBS-ENABLED-GATE-001 — acceptance.md

> Verification layer. Each entry is `AC-XXX` labeled, binary-testable, written as `Given / When / Then`. The GEARS obligation lives in spec.md §C (the `REQ-XXX` entries); this file does not restate GEARS requirements.

## §A. Traceability

| AC-ID      | REQ(s)            | Falsifiable test                                                                  |
|------------|-------------------|-----------------------------------------------------------------------------------|
| AC-OEG-001 | REQ-OEG-002       | `TestEnableObservabilityIfConfigured_DisabledByEnabledKey`                       |
| AC-OEG-002 | REQ-OEG-003       | `TestEnableObservabilityIfConfigured_WithConfigFile` (existing, preserved)       |
| AC-OEG-003 | REQ-OEG-004       | `TestEnableObservabilityIfConfigured_AbsentKeyDefaultsDisabled` (new)            |
| AC-OEG-004 | REQ-OEG-005       | `TestEnableObservabilityIfConfigured_NoConfigFile` (existing, preserved)         |
| AC-OEG-005 | REQ-OEG-006/008   | `AC-HOI-007` 4-quadrant cohabitation test (existing, must remain green)          |
| AC-OEG-006 | REQ-OEG-007       | Static: deps.go calls `hook.IsObservabilityEnabled()`; grep-verifiable            |
| AC-OEG-007 | Regression        | Sabotage round-trip on AC-OEG-001 (see §C)                                       |

## §B. Acceptance Criteria (Given / When / Then)

### AC-OEG-001 — file present + `enabled: false` does NOT enable observability

**Given** the test process has `hook.SetObservabilityMasterForTesting(false)` active (simulating the `enabled: false` resolved state of the `observability.yaml` master toggle), the test is marked non-parallel (no `t.Parallel()` call, matching the `internal/hook/notification_test.go:87` precedent: "Sequential (not t.Parallel) due to SetObservabilityMasterForTesting global state."), and `t.Cleanup(hook.ResetObservabilityMasterForTesting)` is registered; and a temporary config tree under `tmpDir` carries `.moai/config/sections/observability.yaml` (whose literal contents are `observability:\n  enabled: false\n`) to satisfy the `os.Stat` existence gate
**When** `enableObservabilityIfConfigured(reg, tmpDir)` is invoked with a stub registry that counts `EnableObservability` calls
**Then** `EnableObservability` is NOT called (call count == 0) AND the `.moai/logs` directory is NOT created.

> Test-seam rationale: `hook.IsObservabilityEnabled()` (internal/hook/observability_master.go:70) takes no `cwd` parameter and resolves the project root via `CLAUDE_PROJECT_DIR` → `os.Getwd()`, so writing `enabled: false` to a tempdir does NOT influence its return. The master-toggle seam (`SetObservabilityMasterForTesting`) is the only way to control the enabled-decision from a tempdir-based test; the tempdir is still used because `enableObservabilityIfConfigured` consults `os.Stat` on it for the existence gate and joins `logDir` from it.

### AC-OEG-002 — file present + `enabled: true` preserves the current enabling behavior

**Given** the test process has `hook.SetObservabilityMasterForTesting(true)` active (simulating the `enabled: true` resolved state), the test is marked non-parallel, and `t.Cleanup(hook.ResetObservabilityMasterForTesting)` is registered; and a temporary config tree under `tmpDir` carries `.moai/config/sections/observability.yaml` (literal contents `observability:\n  enabled: true\n`)
**When** `enableObservabilityIfConfigured(reg, tmpDir)` is invoked
**Then** `EnableObservability(logDir)` IS called exactly once with `logDir == filepath.Join(tmpDir, ".moai", "logs")` AND the log directory exists.

### AC-OEG-003 — file present + key absent defaults to disabled

**Given** the test process has `hook.SetObservabilityMasterForTesting(false)` active (simulating the absent-key resolved state — `IsObservabilityEnabled()` safe-defaults to false when the YAML key is absent, so the injected master-toggle value encodes the same outcome), the test is marked non-parallel, and `t.Cleanup(hook.ResetObservabilityMasterForTesting)` is registered; and a temporary config tree under `tmpDir` carries `.moai/config/sections/observability.yaml` (literal contents `observability:\n` or `{}\n` — no `enabled` key)
**When** `enableObservabilityIfConfigured(reg, tmpDir)` is invoked
**Then** `EnableObservability` is NOT called (aligns with `IsObservabilityEnabled()`'s safe-default false on absent key).

### AC-OEG-004 — file missing preserves the current skip behavior

**Given** a temporary config tree with NO `.moai/config/sections/observability.yaml`
**When** `enableObservabilityIfConfigured(reg, tmpDir)` is invoked
**Then** `EnableObservability` is NOT called AND no log directory is created.

### AC-OEG-005 — cohabitation 4-quadrant test remains green

**Given** the `internal/hook` test suite carrying `AC-HOI-007` (4-quadrant cohabitation test)
**When** `go test ./internal/hook/...` runs after the deps.go change lands
**Then** the 4-quadrant test passes unchanged AND no assertion in that test file was modified by this SPEC.

### AC-OEG-006 — deps.go delegates the read to `hook.IsObservabilityEnabled()`

**Given** the patched `internal/cli/deps.go`
**When** a static analysis pass (`grep -n 'IsObservabilityEnabled' internal/cli/deps.go`) runs
**Then** at least one match exists inside `enableObservabilityIfConfigured` AND a companion grep for `hook.observability_events` / `hook.opt_in.enabled` inside the same file returns zero matches (REQ-OEG-006).

### AC-OEG-007 — regression test is falsifiable (sabotage round-trip)

**Given** the new test `TestEnableObservabilityIfConfigured_DisabledByEnabledKey`
**When** the implementation is sabotaged (revert to `os.Stat`-only gate, ignoring `enabled`) and the test runs
**Then** the test FAILS (proves it actually exercises REQ-OEG-002). **When** the correct implementation is restored and the test runs again
**Then** the test PASSES. Both phases of the round-trip are required evidence; a test that passes under both the correct and the sabotaged implementation is vacuous and fails this AC.

## §C. Sabotage Round-Trip Protocol (AC-OEG-007 operationalization)

**Canonical test pattern.** The regression test `TestEnableObservabilityIfConfigured_DisabledByEnabledKey` MUST use the `hook.SetObservabilityMasterForTesting(false)` + non-parallel + `t.Cleanup(hook.ResetObservabilityMasterForTesting)` form — the same pattern established at `internal/hook/notification_test.go:87` ("Sequential (not t.Parallel) due to SetObservabilityMasterForTesting global state.") and reused at `internal/hook/task_created_test.go:80`. This is the ONLY honest way to control the `IsObservabilityEnabled()` enabled-decision from a tempdir-based test, because `IsObservabilityEnabled()` resolves cwd via `CLAUDE_PROJECT_DIR` → `os.Getwd()` and cannot read a tempdir's `enabled` key. The tempdir remains in use only to satisfy the `os.Stat` existence gate and to provide the `logDir` join target.

1. **GREEN run.** Apply the correct implementation (deps.go consults `IsObservabilityEnabled()`). Run `go test -run TestEnableObservabilityIfConfigured_DisabledByEnabledKey ./internal/cli/`. Capture the PASS output.
2. **RED run (sabotage).** Temporarily revert deps.go to the `os.Stat`-only gate (the bug — ignoring the master toggle). Run the same test. It MUST FAIL: under the sabotaged implementation the gate ignores `SetObservabilityMasterForTesting(false)` and calls `EnableObservability` anyway, which the stub registry's call-count assertion rejects. Capture the FAIL output.
3. **Restore + re-GREEN.** Re-apply the correct implementation. Re-run the test. It MUST PASS.
4. **Evidence.** Both outputs (RED sabotage + restored GREEN) are recorded in `progress.md §E.2` by the run-phase implementer. A test that cannot be made to FAIL under sabotage is vacuous — the implementer must rewrite it before claiming AC-OEG-007. The RED leg here is reliable precisely because the master-toggle seam bypasses the cwd/env resolution that would otherwise mask the sabotage (D4 closure: AC-OEG-007's falsifiability gap is resolved by the seam-based pattern).

## §D. Severity Matrix

| AC-ID      | Severity | Rationale                                                                  |
|------------|----------|----------------------------------------------------------------------------|
| AC-OEG-001 | MUST-PASS | Direct evidence for the contract violation this SPEC exists to fix.       |
| AC-OEG-002 | MUST-PASS | Preserve the documented enabling path; regression guard.                  |
| AC-OEG-003 | MUST-PASS | Absent-key safe-default alignment with `IsObservabilityEnabled()`.        |
| AC-OEG-004 | MUST-PASS | Preserve existing skip-on-missing behavior.                               |
| AC-OEG-005 | MUST-PASS | Co-existence invariant; breaking AC-HOI-007 escalates to Tier M.          |
| AC-OEG-006 | SHOULD-PASS | Static verification of SSOT reuse; falsifiable by grep.                  |
| AC-OEG-007 | MUST-PASS | Falsifiability gate for the entire AC suite — no vacuous greens.          |

## §E. Closure Gates (Definition of Done)

- All MUST-PASS ACs are GREEN with attributed command output.
- AC-OEG-007 sabotage round-trip evidence (RED + restored GREEN) is recorded.
- `go test ./internal/cli/ ./internal/hook/` exits 0.
- `AC-HOI-007` remains on its pre-existing behavior (no edit to the 4-quadrant test file).
- No new yaml-key reader was introduced in deps.go (AC-OEG-006 grep-verifies `IsObservabilityEnabled()` reuse and the absence of `system.yaml` key reads).

## §F. Forward-Looking Checks (post-merge drift guards)

- **Drift guard.** If a future PR touches `enableObservabilityIfConfigured`, CI should re-run AC-OEG-001 and AC-OEG-007. A regression here is the canonical "we re-introduced the os.Stat-only gate" signal.
- **Cohabitation sentinel.** If `AC-HOI-007` is ever modified, the modifying PR must cite this SPEC (or a successor) as the fresh-SPEC authority — per the cohabitation note's "Do NOT unify without a fresh SPEC" rule.
- **Reader-count invariant.** The count of independent `observability.yaml` `enabled`-key readers in the codebase must NOT increase beyond the current set. After this SPEC, deps.go is a consumer of the existing reader (`IsObservabilityEnabled()`), not a new reader; a future grep that finds a NEW independent reader inside `cli/` is a regression signal.

## §G. Edge Cases

- **YAML parse error in `observability.yaml`.** `IsObservabilityEnabled()` already safe-defaults to false on parse errors. deps.go inherits that behavior by delegation — no special handling required.
- **`enabled` key present but non-boolean.** Same as parse-error: safe-default false via `IsObservabilityEnabled()`.
- **Temp-dir cleanup.** Tests MUST use `t.TempDir()` per CLAUDE.local.md §6; no test writes under the project root.
- **Concurrent `InitDependencies` calls.** `sync.Once` inside `IsObservabilityEnabled()` is goroutine-safe; the single-startup-call invariant (plan.md §D) means this is not exercised in practice, but the guard is correct if it ever is.
