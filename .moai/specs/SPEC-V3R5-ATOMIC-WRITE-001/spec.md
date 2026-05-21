---
id: SPEC-V3R5-ATOMIC-WRITE-001
title: "P0-4 atomicWrite Safety Violation Fix — internal/migrate/hook_cleanup.go"
version: "0.1.0"
status: draft
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P0
phase: "v2.20.0-rc1"
module: "internal/migrate"
lifecycle: spec-anchored
tags: "security, atomic-write, p0, release-blocker, migrate, settings"
tier: S
---

# SPEC-V3R5-ATOMIC-WRITE-001 — P0-4 atomicWrite Safety Violation Fix

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-20 | GOOS Kim | Initial draft — Tier S release-blocker SPEC for P0-4 (review-v214-to-HEAD.md L53-58) |

## 1. Background

The function `atomicWrite` in `internal/migrate/hook_cleanup.go:124-132` (comment header L124-126 + function body L127-132) is named as if it provides atomic write semantics but its implementation is a plain wrapper around `os.WriteFile`:

```go
// Current implementation (defective)
func atomicWrite(path string, data []byte, perm os.FileMode) error {
    if err := os.WriteFile(path, data, perm); err != nil {
        return fmt.Errorf("write file: %w", err)
    }
    return nil
}
```

Two callers depend on this function:

- `internal/migrate/hook_cleanup.go:109` — writes the user's `.claude/settings.json` during the v3.0 migration cleanup of retired hook event names.
- `internal/migrate/hook_cleanup.go:121` — writes the per-day archive file `migration-YYYY-MM-DD.json`.

A crash, power loss, or signal interruption during `os.WriteFile` leaves the destination file in a partial (possibly empty or truncated) state. For `.claude/settings.json` this corrupts the user's Claude Code configuration — the exact class of file that demands atomic-write semantics.

The review report `.moai/reports/review-v214-to-HEAD.md` L53-58 classifies this as P0-4 release blocker (CWE-732 + CWE-552 class — incorrect permission preservation + insufficient information protection during write). It is one of the eight P0 release blockers that must be cleared before the v2.20.0-rc1 tag can be issued.

Six other writers in the codebase already use the correct tmp+rename pattern:

- `internal/runtime/persist.go:62-85` — canonical reference implementation
- `internal/config/manager.go:380` (atomicWrite)
- `internal/manifest/manifest.go:215` (atomicWriteFile)
- `internal/template/deployer.go:14` (atomicWriteFile)
- `internal/harness/applier.go:74`
- `internal/harness/tier/tier.go:280-285`
- `internal/harness/safety/canary_veto.go:267-271`

This SPEC adopts the same pattern, restricted to a Tier S (≤300 LOC, ≤5 files) surgical change.

## 2. Stakeholders

| Role | Concern |
|------|---------|
| End user (Claude Code operator) | `.claude/settings.json` must not be corrupted during migration; loss of settings disrupts daily work. |
| MoAI maintainer | P0-4 must clear before v2.20.0-rc1 release tag. |
| Future SPEC authors | The function name `atomicWrite` must match its semantics; future readers should not be deceived by the name. |

## 3. Goal

Replace the non-atomic `atomicWrite` body in `internal/migrate/hook_cleanup.go` with a tmp+write+sync+close+rename pattern that matches the function name and the codebase's existing convention, without changing the function signature or breaking the two existing callers.

## 4. EARS Requirements

### 4.1 Functional Requirements

**REQ-AWR-001** (Ubiquitous):
The function `atomicWrite(path string, data []byte, perm os.FileMode) error` in `internal/migrate/hook_cleanup.go` **shall** write data to the destination path using the sequence: `os.CreateTemp` in the destination directory → write all bytes → call `(*os.File).Sync` → close the file → `os.Rename` the temp file to the destination path.

**REQ-AWR-002** (Ubiquitous):
The function `atomicWrite` **shall** preserve its existing signature `(path string, data []byte, perm os.FileMode) error` so that both existing callers (`hook_cleanup.go:109` settings.json write site and `hook_cleanup.go:121` archive write site) continue to compile and behave without source modification at the call sites.

**REQ-AWR-003** (Event-Driven):
WHEN an error occurs at any step of the atomic write sequence (CreateTemp / Write / Sync / Close / Rename), the function **shall** remove the temp file via `os.Remove` (best-effort, errors ignored) and return a wrapped error containing the failing step's name.

**REQ-AWR-004** (Ubiquitous):
The function `atomicWrite` **shall** apply the supplied `perm` parameter to the destination file after the rename succeeds, via an explicit `os.Chmod` call, so that the permission contract observed by the caller (0o644) is preserved regardless of `os.CreateTemp`'s default (0o600).

### 4.2 Quality Requirements

**REQ-AWR-005** (Ubiquitous):
The function naming **shall** match its semantics; the implementation **shall not** rely on plain `os.WriteFile` as the write primitive.

**REQ-AWR-006** (Ubiquitous):
The function **shall** carry an `@MX:WARN` annotation with an `@MX:REASON` sub-line documenting the atomicity contract and the historical defect (P0-4 reference), per `.claude/rules/moai/workflow/mx-tag-protocol.md` mandatory-fields clause.

## 5. Boundary

### 5.1 In Scope

- Modification of the function body `atomicWrite` at `internal/migrate/hook_cleanup.go:124-132` (comment header + body — see §1 for the exact range).
- Addition of `@MX:WARN` + `@MX:REASON` annotation adjacent to the function declaration.
- Addition of a new test file `internal/migrate/hook_cleanup_test.go` (or extension of an existing one) with regression coverage for the atomic semantics.
- Verification that the two existing callers behave identically post-fix.

### 5.2 Out of Scope

- Refactoring `atomicWrite` into a shared `pkg/atomicio` package consolidating all six existing implementations. (Tracked separately as a Tier L initiative — see EXCL-AWR-003.)
- Modification of other writers in the codebase (`runtime/persist.go`, `harness/applier.go`, `harness/tier/tier.go`, `harness/safety/canary_veto.go`, `config/manager.go`, `manifest/manifest.go`, `template/deployer.go`). They already implement the correct pattern.
- Concurrency protection for the two callers. The migration step is single-writer by design (REQ-MIG002-019 invariant), so no `sync.Mutex` is added by this SPEC.
- Cross-platform (Windows) build-tag split. The `os.CreateTemp` / `(*os.File).Sync` / `os.Rename` / `os.Chmod` primitives are POSIX-portable and work identically on Windows; no GOOS-specific code is introduced.

## 6. Risks

**R-AWR-001 — Caller regression**

Risk: One of the two callers (`hook_cleanup.go:109` or `:121`) might depend on side effects of the old implementation (for example, the file descriptor being held open in the same scope, or the file existing with `0o600` interim permission for some observation window).

Mitigation: The new implementation preserves the function signature, returns the same error wrapping shape, and applies an explicit `os.Chmod(path, perm)` post-rename to restore the requested mode. The acceptance criteria include a regression test verifying both call sites still pass.

**R-AWR-002 — Windows rename antivirus interference**

Risk: On some Windows environments antivirus or backup software may briefly hold a handle on the destination file, causing `os.Rename` to fail with `ERROR_SHARING_VIOLATION`. The other six writers in the codebase already use this exact pattern in production, so the risk is bounded to the same exposure they already carry.

Mitigation: The error path returns a wrapped error containing the rename step name; the caller can decide retry policy. The codebase-wide pattern remains consistent. Behavior matches `runtime/persist.go:80-83` (the reference implementation).

**R-AWR-003 — Permission preservation gap on temp file**

Risk: `os.CreateTemp` returns a file with mode `0o600` by default on POSIX. If `os.Rename` succeeds but the subsequent `os.Chmod` step is omitted or fails silently, the destination file may end up with `0o600` instead of the caller's requested `0o644`, which could change the visibility of `.claude/settings.json` for the user.

Mitigation: REQ-AWR-004 mandates explicit `os.Chmod` post-rename with the `perm` parameter. AC-AWR-005 verifies the resulting file's mode bits match the requested `perm`. The chmod error path returns a wrapped error.

## 7. Exclusions

| ID | Excluded scope | Rationale |
|----|---------------|-----------|
| EXCL-AWR-001 | P0-1 (GLM token file 0o644 permission) fix | Already resolved by SPEC-V3R5-SECURITY-CRIT-001 PR #1032 (merged 2026-05-20T13:05:12Z, commit `03a2552a2`). Out of P0-4 scope. |
| EXCL-AWR-002 | P1-S4 `RateLimiter.saveState` atomicity + concurrency-safe mutex | Separate `internal/harness/safety/` concern. Will be a follow-up SPEC. |
| EXCL-AWR-003 | Consolidation of all six `atomicWrite` / `atomicWriteFile` implementations into a single `pkg/atomicio` package | Tier L refactor with broader impact; deferred to post-v2.20.0-rc1. |
| EXCL-AWR-004 | The 24 dirty files currently in the working tree (`internal/cli/wizard`, `internal/cli/banner.go`, `internal/cli/doctor*.go`, `internal/cli/help.go`, `internal/cli/init*.go`, `internal/cli/testdata/*.golden`, `internal/cli/root_test.go`, `internal/cli/coverage*_test.go`, `internal/cli/clean.go`, `internal/tui/theme*.go`, `internal/hook/.moai/`, `.moai/harness/usage-log.jsonl`) | Pre-existing in-flight work unrelated to P0-4. Must not be touched by this SPEC. |
| EXCL-AWR-005 | P0-5 (`rate-limit-state.json` schema collision), P0-6 (`pipeline.go` unimplemented stubs), P0-7 (`defaultProjectedScorer` no-op), P0-8 (`internal/github/runner` stub success) | Each is a separate v2.20.0-rc1 blocker and will be its own SPEC. |

## 8. Implementation Hint (Tier S non-normative)

Reference pattern from `internal/runtime/persist.go:62-85`:

```go
// @MX:WARN: writes to .claude/settings.json — partial writes corrupt user configuration.
// @MX:REASON: P0-4 (review-v214-to-HEAD.md L53-58) — prior os.WriteFile body broke atomicity contract.
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

This is provided as guidance only. Implementation details belong to the run phase.

## 9. Cross-references

- Defect source: `.moai/reports/review-v214-to-HEAD.md` L53-58 (P0-4 section)
- Reference pattern: `internal/runtime/persist.go:62-85`
- Other correct writers: `internal/harness/applier.go:74`, `internal/harness/tier/tier.go:280-285`, `internal/harness/safety/canary_veto.go:267-271`, `internal/config/manager.go:380`, `internal/manifest/manifest.go:215`, `internal/template/deployer.go:14`
- Related SPEC: SPEC-V3R5-SECURITY-CRIT-001 (PR #1032 — P0-1/P0-2/P0-3 fixes, merged 2026-05-20)
- Pattern doctrine: `.claude/rules/moai/workflow/mx-tag-protocol.md` § Mandatory Fields
- Tier S policy: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier
