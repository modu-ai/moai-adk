# Implementation Plan — SPEC-V3R5-ATOMIC-WRITE-001

Tier: S (≤300 LOC, ≤5 files affected, 2 required artifacts per LEAN workflow)
Brownfield strategy: surgical fix (PRESERVE existing function signature and call sites, REPLACE body only).

## Milestones

| Priority | Milestone | Deliverable |
|----------|-----------|-------------|
| P0 | M1 — Body rewrite | `atomicWrite` body in `internal/migrate/hook_cleanup.go` replaced with tmp+write+sync+close+rename+chmod pattern. `@MX:WARN` + `@MX:REASON` annotation added. |
| P0 | M2 — Test harness | New `internal/migrate/hook_cleanup_test.go` (or extension) covering the 6 Given-When-Then scenarios. |
| P0 | M3 — Verification | Self-verification batch: full test suite, cross-platform build, coverage gate, lint baseline, race detector. |
| P1 | M4 — Commit | Single Conventional Commit `fix(SPEC-V3R5-ATOMIC-WRITE-001): ...` with `🗿 MoAI` trailer. |

No time estimates (per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation).

## Technical Approach

### M1 — Body Rewrite

File: `internal/migrate/hook_cleanup.go:124-132` (comment header L124-126 + function body L127-132)

Replace the existing body:

```go
func atomicWrite(path string, data []byte, perm os.FileMode) error {
    if err := os.WriteFile(path, data, perm); err != nil {
        return fmt.Errorf("write file: %w", err)
    }
    return nil
}
```

With the canonical pattern (model: `internal/runtime/persist.go:62-85`):

```go
// @MX:WARN: writes to user .claude/settings.json — partial writes corrupt configuration.
// @MX:REASON: P0-4 (review-v214-to-HEAD.md L53-58) — prior os.WriteFile body broke the
//             atomicity contract implied by the function name. The two callers (line 109
//             settings.json, line 121 archive) depend on all-or-nothing semantics.
func atomicWrite(path string, data []byte, perm os.FileMode) error {
    dir := filepath.Dir(path)
    tmp, err := os.CreateTemp(dir, ".hook_cleanup_tmp_*")
    if err != nil {
        return fmt.Errorf("atomicWrite: create temp file: %w", err)
    }
    tmpName := tmp.Name()

    if _, err := tmp.Write(data); err != nil {
        _ = tmp.Close()
        _ = os.Remove(tmpName)
        return fmt.Errorf("atomicWrite: write temp: %w", err)
    }
    if err := tmp.Sync(); err != nil {
        _ = tmp.Close()
        _ = os.Remove(tmpName)
        return fmt.Errorf("atomicWrite: sync temp: %w", err)
    }
    if err := tmp.Close(); err != nil {
        _ = os.Remove(tmpName)
        return fmt.Errorf("atomicWrite: close temp: %w", err)
    }
    if err := os.Rename(tmpName, path); err != nil {
        _ = os.Remove(tmpName)
        return fmt.Errorf("atomicWrite: rename: %w", err)
    }
    if err := os.Chmod(path, perm); err != nil {
        return fmt.Errorf("atomicWrite: chmod: %w", err)
    }
    return nil
}
```

Signature is preserved verbatim. The two call sites (line 109 and line 121) require no source-level changes.

### M2 — Test Harness

Test file: `internal/migrate/hook_cleanup_test.go`

Required tests (mapped from acceptance.md scenarios):

| Test name | Scenario | Method |
|-----------|----------|--------|
| `TestAtomicWrite_HappyPath` | Scenario 1 | `t.TempDir()` → call → assert content + mode + no temp residue |
| `TestAtomicWrite_ReplaceExisting` | Scenario 2 | Pre-write OLD → call with NEW → assert exact replacement |
| `TestAtomicWrite_PreservesPerm` | Scenario 5 | Call with `0o644` → `os.Stat` Perm == 0o644 |
| `TestAtomicWrite_CleanupOnError` | Scenario 3 + 4 | Force error via unwritable dir or pre-existing file lock → assert OLD content preserved + no temp residue |
| `TestCleanupUserSettings_Regression` | Scenario 6 | Caller-level test verifying `CleanupUserSettings` behavior is unchanged |

All tests use `t.TempDir()` for isolation (per `.moai/config/sections/quality.yaml` Go test isolation rule).

### M3 — Verification (single-turn parallel batch)

The orchestrator MUST batch the following read-only verifications in one assistant turn (per `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution):

```bash
go test ./...                                              # AC-AWR-004
go test -cover ./internal/migrate/...                      # AC-AWR-007
go test -run TestAtomicWrite ./internal/migrate/...        # AC-AWR-003
go test -race ./internal/migrate/...                       # race gate
GOOS=windows GOARCH=amd64 go build ./...                   # AC-AWR-008
golangci-lint run --timeout=2m                             # lint baseline
grep -A 30 '^func atomicWrite' internal/migrate/hook_cleanup.go  # AC-AWR-001
grep -B 4 '^func atomicWrite' internal/migrate/hook_cleanup.go   # AC-AWR-006
```

### M4 — Commit

Conventional Commits format:

```
fix(SPEC-V3R5-ATOMIC-WRITE-001): replace os.WriteFile with atomic tmp+rename pattern

internal/migrate/hook_cleanup.go atomicWrite() was a misnamed wrapper around
os.WriteFile, breaking the atomicity contract implied by its name. Two callers
(settings.json at line 109, archive at line 121) could corrupt .claude/settings.json
on crash or signal mid-write.

Replace body with the canonical pattern from internal/runtime/persist.go:62-85
(CreateTemp + Write + Sync + Close + Rename + Chmod). Signature unchanged.
Add @MX:WARN + @MX:REASON annotation. Add regression test coverage.

Closes P0-4 (review-v214-to-HEAD.md L53-58) — v2.20.0-rc1 release blocker 4/8.

🗿 MoAI <email@mo.ai.kr>
```

## Files Affected (≤5, per Tier S budget)

| Path | Change type | Notes |
|------|-------------|-------|
| `internal/migrate/hook_cleanup.go` | Modify | Body of `atomicWrite` rewritten. `filepath` import already present. `@MX:WARN` annotation added. |
| `internal/migrate/hook_cleanup_test.go` | New | Five tests covering Scenarios 1–6. |

Total: 2 files. Well within Tier S budget (≤5).

## PRESERVE List (must not be modified by this SPEC)

The following 24 in-flight files are unrelated to P0-4 and must remain untouched:

- `.moai/harness/usage-log.jsonl` (runtime-managed; never commit)
- `internal/cli/banner.go`
- `internal/cli/clean.go`
- `internal/cli/coverage_improvement_test.go`
- `internal/cli/coverage_test.go`
- `internal/cli/doctor.go`
- `internal/cli/doctor_golden_test.go`
- `internal/cli/doctor_test.go`
- `internal/cli/help.go`
- `internal/cli/init.go`
- `internal/cli/init_layout.go` (untracked)
- `internal/cli/root_test.go`
- `internal/cli/testdata/banner-current-dark.golden`
- `internal/cli/testdata/banner-current-light.golden`
- `internal/cli/testdata/banner-current-nocolor.golden`
- `internal/cli/testdata/doctor-dark.golden`
- `internal/cli/testdata/doctor-light.golden`
- `internal/cli/testdata/doctor-nocolor.golden`
- `internal/cli/wizard/wizard.go`
- `internal/cli/wizard/wizard_test.go`
- `internal/cli/wizard/review.go` (untracked)
- `internal/tui/theme.go`
- `internal/tui/theme_test.go`
- `internal/hook/.moai/` (untracked — runtime CWD leak; separate cleanup)

Stage rule: `git add .moai/specs/SPEC-V3R5-ATOMIC-WRITE-001/` only for the plan-phase commit. For the run-phase commit, `git add internal/migrate/hook_cleanup.go internal/migrate/hook_cleanup_test.go .moai/specs/SPEC-V3R5-ATOMIC-WRITE-001/` only. Never use `git add -A` or `git add .` or `git add internal/`.

## Risks and Mitigations (mapped to spec.md §6)

| Risk | Mitigation | Verification |
|------|------------|--------------|
| R-AWR-001 caller regression | Signature unchanged; explicit `os.Chmod` post-rename preserves observable mode | AC-AWR-004 (full suite) + Scenario 6 test |
| R-AWR-002 Windows rename antivirus | Pattern identical to 6 other writers already in production | AC-AWR-008 (cross-platform build); behavioral parity with `runtime/persist.go` |
| R-AWR-003 permission gap | `os.Chmod(path, perm)` after rename; explicit test | AC-AWR-005 (`TestAtomicWrite_PreservesPerm`) |

## Cross-references

- `spec.md` § 4 (EARS REQs), § 6 (Risks), § 7 (Exclusions)
- `acceptance.md` § Binary AC Matrix, § Test Scenarios, § Definition of Done
- Pattern source: `internal/runtime/persist.go:62-85`
- Defect source: `.moai/reports/review-v214-to-HEAD.md` L53-58
- Tier S policy: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier
- @MX rules: `.claude/rules/moai/workflow/mx-tag-protocol.md` § Mandatory Fields
