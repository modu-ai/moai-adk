---
id: SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001
title: "Hook cwd leak audit + resolveProjectRoot consistency — Implementation Plan"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
tier: S
---

# SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001 — Implementation Plan

## Section A — Pre-flight Verification

Before executing M1, the implementing agent (manager-develop, cycle_type=ddd) shall:

1. **Verify branch and baseline**:
   ```bash
   git branch --show-current   # Expected: main (or feat/SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001)
   git log --format='%h %s' -1 # Expected: eaff5f272 (or downstream commit)
   ```

2. **Capture M0 leak baseline**:
   ```bash
   grep -rn "os\.Getwd" internal/hook/ | grep -v "_test.go" | wc -l
   # Expected: 9 (the 9 sites enumerated in spec.md §A.1)
   ```

3. **Capture M0 lint baseline** (for AC-HCWA-005 "0 NEW" comparison):
   ```bash
   golangci-lint run --timeout=2m ./internal/hook/... 2>&1 | tee /tmp/m0-lint-baseline.txt
   # Save count of WARNING/ERROR lines for diff at M3
   ```

4. **Capture M0 race baseline**:
   ```bash
   go test -race -count=1 ./internal/hook/... 2>&1 | tail -20
   # Expected: PASS or ok all packages
   ```

5. **Verify PRESERVE list integrity** (these files must remain byte-identical at end of M3):
   ```bash
   sha256sum internal/hook/subagent_stop.go internal/hook/post_tool_metrics.go internal/hook/cohabitation_guard_test.go > /tmp/preserve-baseline.sha256
   ```

---

## Section B — Implementation Approach

### B1. Scope and Files

**5 source files modified** (3 in `package hook`, 1 in `package hook/quality`, 1 helper-extraction location):

1. `internal/hook/subagent_start.go` — 2 sites (lines 37, 211)
2. `internal/hook/pre_tool.go` — 2 sites (lines 326, 336)
3. `internal/hook/observability_master.go` — 1 site (line 82)
4. `internal/hook/quality/gate.go` — 4 sites (lines 276, 297, 399, 480)
5. `internal/hook/quality/path_resolve.go` — NEW helper file (or inline at `gate.go` top — implementer choice; see B4)

**Total LOC delta estimate**: ~80-120 lines (+30 helper code, ~50 refactored call sites, ~30 warning log statements).

### B2. Two-Strategy Refactor (per spec.md §A.1.5)

**Strategy A — `package hook`** (3 files, 5 sites):

| Site | Current code | Refactor target |
|------|-------------|-----------------|
| subagent_start.go:37 | `os.Getwd()` (after env var check) | Call `resolveProjectRootFromInputOrEnv(nil)` thin wrapper, OR inline env→Getwd with `slog.Warn("cwd_fallback", ...)` |
| subagent_start.go:211 | `os.Getwd()` (after `input.CWD` + env var checks) | Already prefers `input.CWD` → env var → Getwd; refactor to centralize via shared helper that emits the warning log. Preserve current order: input.CWD first |
| pre_tool.go:326 | `os.Getwd()` (after env var check) | Same as subagent_start.go:37 |
| pre_tool.go:336 | `os.Getwd()` (after env var check) | Same as subagent_start.go:37 |
| observability_master.go:82 | `os.Getwd()` (after `CLAUDE_PROJECT_DIR` check) | Replace with shared helper that does env→Getwd with warning log (no `.moai/` guard) |

**Refactor strategy**: Extract a thin helper inside `package hook` (place in `internal/hook/post_tool_metrics.go` adjacent to `resolveProjectRoot`, OR new file `internal/hook/path_resolve.go`):

```go
// resolveProjectRootFromEnv returns CLAUDE_PROJECT_DIR or os.Getwd() fallback
// without the .moai/ existence guard. Used by hook handlers that need a cwd
// for config-file reads where the .moai/ directory may not exist on the
// resolved path (e.g., observability_master loading observability.yaml itself).
// Emits slog.Warn with key "cwd_fallback":true when os.Getwd() is used.
func resolveProjectRootFromEnv(caller string) string {
    if root := os.Getenv(config.EnvClaudeProjectDir); root != "" {
        return root
    }
    cwd, err := os.Getwd()
    if err != nil {
        slog.Debug("cwd fallback failed", "caller", caller, "error", err)
        return ""
    }
    slog.Warn("cwd fallback used (CLAUDE_PROJECT_DIR not set)",
        "cwd_fallback", true,
        "caller", caller,
        "resolved_cwd", cwd,
    )
    return cwd
}
```

For `subagent_start.go:211` (which prefers `input.CWD` first), introduce a parallel helper:

```go
// resolveProjectRootFromInputOrEnv returns input.CWD, then CLAUDE_PROJECT_DIR,
// then os.Getwd() fallback. No .moai/ existence guard. Used by handlers that
// have HookInput available and prefer input.CWD over env var.
func resolveProjectRootFromInputOrEnv(input *HookInput, caller string) string {
    if input != nil && input.CWD != "" {
        return input.CWD
    }
    return resolveProjectRootFromEnv(caller)
}
```

These helpers are NOT the same as `resolveProjectRoot(input *HookInput) string` (which has the `.moai/` guard for writes). They are siblings used for read-side and registration-time resolution.

**Strategy B — `package hook/quality`** (1 file, 4 sites):

`package quality` does not import `package hook` and has no `HookInput`. Introduce a package-local helper inside `internal/hook/quality/gate.go` (or new file `internal/hook/quality/path_resolve.go`):

```go
// resolveQualityProjectDir returns the project directory for quality-gate
// operations. Preference order: cfg.ProjectDir → CLAUDE_PROJECT_DIR env var →
// os.Getwd() fallback. Emits slog.Warn when os.Getwd() fallback is used.
// No .moai/ existence guard — quality-gate operations are contractually
// invoked from project roots and may need to operate in subdirectories during
// pre-commit hook execution.
func resolveQualityProjectDir(cfg GateConfig, caller string) string {
    if cfg.ProjectDir != "" {
        return cfg.ProjectDir
    }
    if root := os.Getenv("CLAUDE_PROJECT_DIR"); root != "" {
        return root
    }
    cwd, err := os.Getwd()
    if err != nil {
        slog.Debug("cwd fallback failed", "caller", caller, "error", err)
        return ""
    }
    slog.Warn("cwd fallback used in quality gate",
        "cwd_fallback", true,
        "caller", caller,
        "resolved_cwd", cwd,
    )
    return cwd
}
```

Note: `package quality` does NOT have access to `config.EnvClaudeProjectDir` constant (from `internal/config`). Either (a) import the constant via a stable path (`internal/hook/internal/envkeys` if exposed), or (b) use the literal string `"CLAUDE_PROJECT_DIR"`. The implementer chooses (b) for the minimal patch — the env var name is itself a stable contract with Claude Code. Document this decision in M3 Section E.

### B3. Refactor Order (Risk-Minimizing)

M1 → M2 → M3 sequence preferred because each milestone is independently shippable and tests pass after each:

1. **M1**: `subagent_start.go` (2 sites) — narrow scope, has direct test coverage in `subagent_start_test.go`. Start here to validate helper-extraction pattern.
2. **M2**: `pre_tool.go` (2 sites) + `observability_master.go` (1 site) — slightly broader; tests in `pre_tool_test.go`, `observability_master_test.go`. Helper from M1 reused.
3. **M3**: `quality/gate.go` (4 sites) — different package, different helper. Independent of M1/M2.

### B4. Helper File Placement

Choice between:
- **(A) Inline in existing files**: Add `resolveProjectRootFromEnv` and `resolveProjectRootFromInputOrEnv` to `post_tool_metrics.go` (adjacent to `resolveProjectRoot`); add `resolveQualityProjectDir` to `gate.go`. Minimal file churn.
- **(B) New helper files**: `internal/hook/path_resolve.go` and `internal/hook/quality/path_resolve.go`. Better discoverability.

**Decision**: Choice (A) preferred for Tier S minimal scope. Implementer may choose (B) if it improves clarity — document in M3 Section E "Notable Implementation Decisions".

### B5. Logging Strategy (REQ-HCWA-008)

Every `os.Getwd()` fallback shall emit `slog.Warn` with structured fields:

```go
slog.Warn("cwd fallback used",
    "cwd_fallback", true,         // REQ-HCWA-008 marker
    "caller", caller,              // function name passed by caller
    "resolved_cwd", cwd,           // observability
)
```

This satisfies REQ-HCWA-008 and enables future telemetry on cwd-fallback frequency.

### B6. Excluded Refactoring

Per spec.md § Exclusions:
- DO NOT touch `subagent_stop.go` (lines 205-213 are the prior fix from `a9b3e8cd8`)
- DO NOT touch `post_tool_metrics.go` `resolveProjectRoot` function (lines 93-113 are the reference pattern)
- DO NOT touch `cohabitation_guard_test.go` (HOI safety net)
- DO NOT fix `*_test.go` `os.Getwd()` patterns (REQ-HCWA-011 — documented only)
- DO NOT introduce a `internal/hook/internal/pathutil/` consolidation package (out of scope; defer to follow-up SPEC)

### B7. Test-side Documentation (REQ-HCWA-011)

After M3, run:
```bash
grep -rn "os\.Getwd" internal/hook/ | grep "_test.go" > .moai/specs/SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001/test-getwd-inventory.txt
```

This inventory is documentation only — NOT a fix target. Place in the SPEC directory for future hardening SPEC reference.

### B8. Subagent Boundary (B11 HARD)

The implementing agent (manager-develop) shall NOT call `AskUserQuestion` or interact with the user directly. If a blocker is discovered (e.g., the cwd-leak pattern in `quality/gate.go` requires a non-trivial config refactor), the agent returns a structured blocker report to the orchestrator per `.claude/rules/moai/core/agent-common-protocol.md` § Blocker Report Format.

### B9. Git Remote Sync (B11 HARD)

The implementing agent (manager-develop) shall NOT push to git remote. Commit creation is the agent's responsibility; the orchestrator handles `git push` and PR creation per Hybrid Trunk policy (Tier S → main direct push allowed for sibling SPECs with passing tests).

---

## Section C — Milestones

### M1 — `subagent_start.go` refactor (2 sites)

**Deliverables**:
- Add helper `resolveProjectRootFromEnv(caller string) string` to `internal/hook/post_tool_metrics.go` (adjacent to existing `resolveProjectRoot`).
- Add helper `resolveProjectRootFromInputOrEnv(input *HookInput, caller string) string` to same file.
- Refactor `subagent_start.go:37` (in `NewSubagentStartHandlerWithConfig`) to call `resolveProjectRootFromEnv("NewSubagentStartHandlerWithConfig")`.
- Refactor `subagent_start.go:211` (in `Handle`) to call `resolveProjectRootFromInputOrEnv(input, "subagentStartHandler.Handle")`.

**Verification** (run before commit):
```bash
go test -race -count=1 ./internal/hook/...
grep -n "os\.Getwd" internal/hook/subagent_start.go     # Expected: 0 matches
grep -n "resolveProjectRootFromEnv\|resolveProjectRootFromInputOrEnv" internal/hook/subagent_start.go  # Expected: 2 matches
sha256sum -c /tmp/preserve-baseline.sha256              # Expected: all OK
```

**Commit message**:
```
feat(SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001): M1 subagent_start.go cwd leak fix via resolveProjectRootFromEnv

- Add resolveProjectRootFromEnv + resolveProjectRootFromInputOrEnv helpers (post_tool_metrics.go)
- Refactor subagent_start.go:37 (NewSubagentStartHandlerWithConfig) to use env-or-getwd helper
- Refactor subagent_start.go:211 (Handle) to use input.CWD-or-env-or-getwd helper
- Emit slog.Warn cwd_fallback:true on os.Getwd() fallback (REQ-HCWA-008)
- PRESERVE: subagent_stop.go, post_tool_metrics.go resolveProjectRoot, cohabitation_guard_test.go

🗿 MoAI <email@mo.ai.kr>
```

### M2 — `pre_tool.go` + `observability_master.go` refactor (3 sites)

**Deliverables**:
- Refactor `pre_tool.go:326` (in `NewPreToolHandler`) to call `resolveProjectRootFromEnv("NewPreToolHandler")`.
- Refactor `pre_tool.go:336` (in `NewPreToolHandlerWithScanner`) to call `resolveProjectRootFromEnv("NewPreToolHandlerWithScanner")`.
- Refactor `observability_master.go:82` (in `loadObservabilityMaster`) to call `resolveProjectRootFromEnv("loadObservabilityMaster")`. Replace the inline `cwd, err := os.Getwd()` + slog.Debug block with the helper call. Path computation (`filepath.Join(root, ".moai", "config", "sections", "observability.yaml")`) remains in-place.

**Verification**:
```bash
go test -race -count=1 ./internal/hook/...
grep -n "os\.Getwd" internal/hook/pre_tool.go internal/hook/observability_master.go  # Expected: 0 matches
grep -n "resolveProjectRootFromEnv" internal/hook/pre_tool.go internal/hook/observability_master.go  # Expected: 3 matches
sha256sum -c /tmp/preserve-baseline.sha256
```

**Commit message**:
```
feat(SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001): M2 pre_tool + observability_master cwd leak fix

- Refactor pre_tool.go:326 (NewPreToolHandler) + :336 (NewPreToolHandlerWithScanner) to resolveProjectRootFromEnv
- Refactor observability_master.go:82 (loadObservabilityMaster) to resolveProjectRootFromEnv
- 3 sites converted; helper from M1 reused (no new helpers added)
- REQ-HCWA-005, REQ-HCWA-006 satisfied

🗿 MoAI <email@mo.ai.kr>
```

### M3 — `quality/gate.go` refactor (4 sites) + status:implemented

**Deliverables**:
- Add helper `resolveQualityProjectDir(cfg GateConfig, caller string) string` to `internal/hook/quality/gate.go` (top of file, after imports, before `GateConfig` struct). Helper uses literal `"CLAUDE_PROJECT_DIR"` string (no `internal/config` import).
- Refactor `gate.go:276` (in `executeStep` ast-grep gate path) to use `resolveQualityProjectDir(g.config, "QualityGate.executeStep.astgrep")`.
- Refactor `gate.go:297` (in `detectToolchain`) to use `resolveQualityProjectDir(g.config, "QualityGate.detectToolchain")`.
- Refactor `gate.go:399` (in `executeStep` ext-filter path) to use `resolveQualityProjectDir(g.config, "QualityGate.executeStep.extfilter")`.
- Refactor `gate.go:480` (in `anyConfigFileExists`) to use `resolveQualityProjectDir(g.config, "QualityGate.anyConfigFileExists")`.
- Update spec.md frontmatter `status: draft` → `status: implemented` and bump `version: "0.1.0"` → `version: "0.2.0"`.
- Update `progress.md` to mark M1, M2, M3 complete.
- Run REQ-HCWA-011 test-side inventory: `grep -rn "os\.Getwd" internal/hook/ | grep "_test.go" > .moai/specs/SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001/test-getwd-inventory.txt`

**Verification** (full AC suite):
```bash
# AC-HCWA-001
test "$(grep -rn 'os\.Getwd' internal/hook/ | grep -v '_test.go' | wc -l)" -eq 0

# AC-HCWA-002 (helper call site counts)
grep -rn "resolveProjectRootFromEnv\|resolveProjectRootFromInputOrEnv" internal/hook/ | wc -l  # Expected: >= 5
grep -rn "resolveQualityProjectDir" internal/hook/quality/ | wc -l                              # Expected: >= 4

# AC-HCWA-003 (cohabitation guard untouched)
go test -run TestCohabitationGuard ./internal/hook/...

# AC-HCWA-004 (race)
go test -race -count=1 ./internal/hook/...

# AC-HCWA-005 (lint baseline 0 NEW)
golangci-lint run --timeout=2m ./internal/hook/... > /tmp/m3-lint-final.txt
diff /tmp/m0-lint-baseline.txt /tmp/m3-lint-final.txt || echo "DIFF — investigate"

# AC-HCWA-006 / AC-HCWA-007 (PRESERVE checks)
sha256sum -c /tmp/preserve-baseline.sha256
```

**Commit message** (M3 final):
```
feat(SPEC-V3R6-HOOK-CWD-LEAK-AUDIT-001): M3 quality/gate.go cwd leak fix + status:implemented v0.2.0

- Add resolveQualityProjectDir helper to package quality (cfg.ProjectDir → env → getwd)
- Refactor 4 sites in quality/gate.go (executeStep×2 + detectToolchain + anyConfigFileExists)
- Total 9 cwd leaks closed across 5 files (subagent_start×2 + pre_tool×2 + observability_master + quality/gate×4)
- REQ-HCWA-011 test-side inventory written to test-getwd-inventory.txt
- All 7 ACs verified: AC-HCWA-001 (0 leaks), AC-HCWA-002 (helper coverage), AC-HCWA-003 (cohabitation guard PASS), AC-HCWA-004 (race clean), AC-HCWA-005 (lint 0 NEW), AC-HCWA-006/007 (PRESERVE files byte-identical)
- spec.md status: draft → implemented, version: 0.1.0 → 0.2.0

🗿 MoAI <email@mo.ai.kr>
```

---

## Section D — Risk Mitigation

| Risk | Mitigation | Trigger |
|------|------------|---------|
| Refactor breaks `subagent_start_test.go` test that depends on cwd-relative paths | Run `go test -count=1 ./internal/hook/subagent_start_test.go` after M1 before commit | If fails: revert M1, escalate to orchestrator |
| `pre_tool_test.go` mocks `os.Getwd()` and breaks when helper is introduced | Inspect test for `os.Getwd` mocking; if present, refactor test to mock `resolveProjectRootFromEnv` instead | M2 verification step |
| `observability_master_test.go` `SetObservabilityMasterForTesting` test helper bypasses the refactor | Verified: `SetObservabilityMasterForTesting` overrides the entire `loadObservabilityMaster` function; refactor is safe | None — already handled |
| `quality/gate.go` import cycle if `internal/config` is imported in `package quality` | Use literal `"CLAUDE_PROJECT_DIR"` string in `resolveQualityProjectDir` (no `internal/config` import); document in M3 Section E | M3 design choice |
| Race detector finds new race due to slog.Warn from multiple goroutines | slog is goroutine-safe by design; verified by checking sibling fixes (`a9b3e8cd8`) | AC-HCWA-004 verification |
| Lint complains about unused parameter `caller` if a future refactor passes "" | Helper always logs `caller`; not optional; ignore false positive | AC-HCWA-005 baseline 0 NEW |

---

## Section E — Notable Implementation Decisions (recorded post-M3)

Decisions recorded by manager-develop during M3 (cycle_type=tdd run-phase 2026-05-23):

- **Helper file placement** (B4 option B chosen over option A): Separate helper files (`internal/hook/path_resolve.go` and `internal/hook/quality/path_resolve.go`) rather than inline in `post_tool_metrics.go` and `gate.go`. Reason: AC-HCWA-007 verification uses `awk '/^func resolveProjectRoot/'` to extract `resolveProjectRoot`'s function body. If the new helpers `resolveProjectRootFromEnv` and `resolveProjectRootFromInputOrEnv` were placed in `post_tool_metrics.go`, the awk regex would glob-match all three function names and produce a 38-line extraction outside the 15-17 range. Separating the helpers into a new file isolates the AC-HCWA-007 awk scan to the reference pattern only.

- **`package quality` env var literal vs import** (CLAUDE_PROJECT_DIR literal chosen): `internal/hook/quality/path_resolve.go` uses the literal string `"CLAUDE_PROJECT_DIR"` rather than importing `internal/config.EnvClaudeProjectDir`. Reason: the env var name is itself a stable contract with Claude Code (defined in upstream Claude Code hook protocol), so the literal is no less stable than the constant. Avoiding the `internal/config` import keeps `package hook/quality` cycle-free with respect to `package hook`. Documented in `path_resolve.go` file header comment.

- **slog.Warn for cwd_fallback**: Both helpers emit `slog.Warn` (not `slog.Debug` or `slog.Info`) when `os.Getwd()` fallback is taken. REQ-HCWA-008 explicitly mandates `slog.Warn` level so that production deployments (where `CLAUDE_PROJECT_DIR` SHOULD always be set by the Claude Code hook system) can filter on `cwd_fallback:true` to detect environment misconfiguration. Local development is the intended log target.

- **REQ-HCWA-011 inventory file**: 1 line per match with file path + line number, generated by `grep -rn` raw output. Total 10 entries: 6 are comments in *_test.go files (documenting the fallback behavior), 4 are actual `os.Getwd()` calls in test files (`subagent_stop_capture_path_test.go:25,27` and `session_end_extra_coverage_test.go:46`). Test scope fixes are NOT in this SPEC's scope (REQ-HCWA-011 documentation-only); future hardening SPEC may address.

- **AC-HCWA-001 refinement post-M3**: Original verification command `grep -rn 'os\.Getwd' internal/hook/ | grep -v '_test.go' | wc -l == 0` was strict-zero but did not account for (1) the two helper definition files which legitimately call `os.Getwd()` as the documented last-resort fallback, and (2) call-site comments referencing the fallback behavior. Refined command adds `grep -v 'path_resolve\.go'` (helper exclusion) and `grep -v '^[^:]*:[0-9]*:[ \t]*//'` (comment exclusion). Refinement documented in acceptance.md AC-HCWA-001 § Exclusion rationale.

- **Other deviations**: None. Plan M1 → M2 → M3 sequence executed as designed. Race condition with parallel SPEC `CI-BASELINE-DRIFT-001` manager-develop caused one M2 commit retry (working tree was reset between commit attempts); recovery was straightforward by re-applying the 3-site M2 changes.

---

## Section F — Cross-References

- Reference pattern: `internal/hook/post_tool_metrics.go:98-113` (`resolveProjectRoot`)
- Prior fix: commit `a9b3e8cd8` (subagent_stop dispatchCapture)
- HOI safety net: `internal/hook/cohabitation_guard_test.go`
- Sibling SPECs: SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001, SPEC-V3R6-HOOK-ASYNC-EXPAND-001, SPEC-V3R6-HOOK-CONTRACT-FIX-001
- Verification batch pattern: `.claude/rules/moai/workflow/verification-batch-pattern.md` (use single-turn multi-Bash for run-phase completion)

---

Version: 0.2.0
Status: implemented
Last Updated: 2026-05-23
