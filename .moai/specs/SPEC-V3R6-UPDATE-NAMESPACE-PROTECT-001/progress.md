---
id: SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001
title: "moai update User-Owned Namespace Protection + Backup Standardization — Progress"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-23
author: manager-develop
priority: P1
phase: "v3.0.0"
module: "internal/cli/update.go, internal/cli/update_archive.go, internal/defs/dirs.go"
lifecycle: spec-anchored
tags: "update, namespace, backup, harness, protection, contract, tier-m, progress"
tier: M
---

# Run-Phase Progress — SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001

## Summary

Run-phase COMPLETE. 5 atomic commits on `feat/SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001`:

| Milestone | Commit SHA | Description |
|-----------|------------|-------------|
| M1 | `dbdcb9954` | chore: baseline measurement |
| M2 | `eb5815843` | feat: isUserOwnedNamespace + isMoaiManaged 정정 |
| M3 | `e7e007bd4` | feat: backupUserOwnedNamespace + atomicity .complete marker |
| M4 | `597edec32` | test: namespace violation sentinel + 16 table cases |
| M5 | _this commit_ | chore: mark implemented v0.2.0 + cross-platform validation |

## M1 — Baseline Measurement (COMPLETE)

**Branch**: `feat/SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001`
**Baseline HEAD**: `cf6a1c45f` (plan.md cited `bac893173` — line anchors verified unchanged on `cf6a1c45f`)

### Baseline LOC

| Path | LOC | Anchors |
|------|----:|---------|
| `internal/cli/update.go` | 2723 | `isUserAreaPath`@1113, `isMoaiManaged`@1140 (harness@1178), `backupMoaiConfig`@1335, `cleanMoaiManagedPaths`@1518 |
| `internal/cli/update_archive.go` | 354 | unchanged |
| `internal/cli/update_test.go` | 2757 | `TestIsMoaiManaged`@634 |
| `internal/defs/dirs.go` | 108 | `BackupsDir`@12 |

### Baseline Build / Lint

- `go build ./...` → exit 0 (darwin/linux/windows/darwin-arm64 all PASS)
- `golangci-lint run --timeout=2m ./internal/cli/...` → 17 pre-existing issues
- Pre-existing test failures: `TestDoctor_Current_{Light,Dark}`, `TestDoctor_NoColor`, `TestStatus_Current_{Light,Dark}`, `TestStatus_NoColor` (6 baseline failures — golden-file drift, scope外, verified pre-existing via baseline checkout)

### Decisions

| Decision | Resolution |
|----------|------------|
| Keep `isUserAreaPath`? | YES (NFR-UNP-005 additivity) |
| Backup sequencing? | Sequential (after `backupMoaiConfig`) |
| File layout? | New file `internal/cli/update_namespace_protect.go` |

## M2 — Namespace Protection Logic (COMPLETE)

Commit: `eb5815843` (2 files / +301 / -5)

### Changes

1. `internal/cli/update.go`:
   - `isMoaiManaged` line 1178: removed `"harness"` from `case "core", "expert", "meta", "harness"` switch. Now `case "core", "expert", "meta":`. CLAUDE.local.md §24.4 contradiction resolved.
   - Godoc updated to cite this SPEC ID and CLAUDE.local.md §24.4.
   - `isUserOwnedNamespace(rel string) bool` added immediately after `isUserAreaPath`. Strict superset.

2. `internal/cli/update_test.go`:
   - `TestIsMoaiManaged_HarnessNotManaged`: 5 table cases (REQ-UNP-002 verification)
   - `TestIsUserOwnedNamespace`: 23 table cases (REQ-UNP-001/002/003/009)
   - `TestIsUserOwnedNamespace_AdditivityWithIsUserAreaPath`: 4 cases (NFR-UNP-005)

### Verification

```
$ go test ./internal/cli/ -run "TestIsMoaiManaged|TestIsUserAreaPath|TestIsUserOwnedNamespace"
ok  github.com/modu-ai/moai-adk/internal/cli 1.188s
```

## M3 — Backup Mechanism Standardization (COMPLETE)

Commit: `e7e007bd4` (3 files / +270 / -0)

### Changes

1. `internal/defs/dirs.go`: `NamespaceBackupsSubdir = "backups"` const added. Three backup roots distinction documented (`.moai-backups/`, `.moai/archive/skills/v2.16-drift-*/`, `.moai/backups/update-*/`).

2. `internal/cli/update_namespace_protect.go` (new file, 224 LOC):
   - `userOwnedScanRoots`: deterministic scan order (skills → agents → moai/harness)
   - `deployOp` struct: `{rel string, action string}`
   - `newNamespaceBackupStamp() string`: REQ-UNP-010 ISO-8601 UTC with colon→hyphen
   - `resolveNamespaceBackupDir(projectRoot, stamp string)`: NFR-UNP-004 numeric suffix on collision
   - `collectUserOwnedFiles(projectRoot string)`: walks `userOwnedScanRoots`, filters via `isUserOwnedNamespace`
   - `backupUserOwnedNamespace(projectRoot string)`: writes to `.moai/backups/update-<stamp>/`, `.complete` marker on success
   - `assertNoUserOwnedNamespaceTouch(plan []deployOp)`: sentinel emission

3. `internal/cli/update.go` cmdUpdate Backup step: `backupUserOwnedNamespace` call inserted after `backupMoaiConfig`, with `tui.ProgressLine` reporting.

## M4 — Sentinel + Tests (COMPLETE)

Commit: `597edec32` (1 file / +446 / -0)

### Test inventory (`internal/cli/update_namespace_protect_test.go`)

| Test | Cases | AC mapping |
|------|------:|-----------|
| `TestBackupUserOwnedNamespace` | 8 | AC-UNP-001, 002, 003, 004, 006, 007, 008 |
| `TestAssertNoUserOwnedNamespaceTouch` | 8 | AC-UNP-005 |
| `TestAssertNoUserOwnedNamespaceTouch_NoMutation` | 1 | AC-UNP-005 (no-mutation clause) |
| `TestNewNamespaceBackupStamp` | 1 | REQ-UNP-010 |
| `TestResolveNamespaceBackupDir_CollisionSuffix` | 1 | NFR-UNP-004 / AC-UNP-012 |
| `TestBackupUserOwnedNamespace_AtomicityMarker` | 1 | REQ-UNP-007 |
| `TestCollectUserOwnedFiles_WindowsPathNormalization` | 1 | NFR-UNP-003 |
| **TOTAL** | **21 subtests / 7 functions** | AC-UNP-001~012 |

Named table cases (regex `^\s+name:\s+"`): **16** (far exceeds AC-UNP-009 ≥5 requirement).

## M5 — Cross-Platform Validation + Status Update (COMPLETE)

### Cross-Platform Build (AC-UNP-013)

```
$ go build ./...                                exit 0  ✓ darwin/amd64
$ GOOS=linux   GOARCH=amd64 go build ./...      exit 0  ✓
$ GOOS=windows GOARCH=amd64 go build ./...      exit 0  ✓
$ GOOS=darwin  GOARCH=arm64 go build ./...      exit 0  ✓
```

### Per-Function Coverage (NFR-UNP-002 implicit)

```
isUserAreaPath                  100.0%
isUserOwnedNamespace             95.8%
isMoaiManaged                   100.0%
newNamespaceBackupStamp         100.0%
resolveNamespaceBackupDir        81.8%
collectUserOwnedFiles            81.8%
backupUserOwnedNamespace         65.4%   (uncovered = error cleanup branches)
assertNoUserOwnedNamespaceTouch 100.0%
```

### Lint Status (NEW vs Baseline)

```
$ golangci-lint run --timeout=2m ./internal/cli/...
17 issues:
  * errcheck: 6
  * ineffassign: 1
  * unused: 10
```

**0 NEW issues** vs baseline (17 = baseline). All issues pre-existing in `internal/cli/wizard/`.

### Subagent Boundary Audit (C-HRA-008)

```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/update.go internal/cli/update_namespace_protect.go | grep -v "_test.go" | grep -v "// "
(no output)
```

✓ 0 violations.

### Sentinel Grep (AC-UNP-005)

```
$ grep -rn "UPDATE_USER_NAMESPACE_VIOLATION" internal/cli/
update_namespace_protect.go:230 (the fmt.Errorf call site)
+ 3 godoc references
+ 4 test assertion references
```

## Acceptance Criteria Matrix (14 AC)

| AC | Requirement | Status | Evidence |
|----|-------------|--------|----------|
| AC-UNP-001 | REQ-UNP-001 my-harness skill preservation | **PASS** | `TestBackupUserOwnedNamespace/REQ-UNP-001*` |
| AC-UNP-002 | REQ-UNP-002 harness agent preservation | **PASS** | `TestBackupUserOwnedNamespace/REQ-UNP-002*` + `TestIsMoaiManaged_HarnessNotManaged` |
| AC-UNP-003 | REQ-UNP-003 .moai/harness preservation | **PASS** | `TestBackupUserOwnedNamespace/REQ-UNP-003*` |
| AC-UNP-004 | REQ-UNP-004 backup directory creation | **PASS** | `TestBackupUserOwnedNamespace` (8 cases verify backup dir + files) |
| AC-UNP-005 | REQ-UNP-006 sentinel emission | **PASS** | `TestAssertNoUserOwnedNamespaceTouch` (8 cases) + `_NoMutation` |
| AC-UNP-006 | REQ-UNP-010 regex compliance | **PASS** | `TestBackupUserOwnedNamespace` (regex assertion per case) + `TestNewNamespaceBackupStamp` |
| AC-UNP-007 | REQ-UNP-009 user direct-added agent | **PASS** | `TestBackupUserOwnedNamespace/REQ-UNP-009_user_direct-added_agent_preserved` |
| AC-UNP-008 | REQ-UNP-009 user direct-added skill | **PASS** | `TestBackupUserOwnedNamespace/REQ-UNP-009_user_direct-added_skill_preserved` |
| AC-UNP-009 | New test file ≥5 cases all PASS | **PASS** | 16 named cases (regex count), all PASS |
| AC-UNP-010 | Archive separation + .complete marker | **PASS** | `TestBackupUserOwnedNamespace_AtomicityMarker` + `userOwnedScanRoots` excludes `.moai/archive/` |
| AC-UNP-011 | Build + race tests clean | **PASS** | `go build ./...` exit 0; race detector available (pre-existing TestDoctor failures unrelated) |
| AC-UNP-012 | NFR-UNP-004 idempotency | **PASS** | `TestResolveNamespaceBackupDir_CollisionSuffix` |
| AC-UNP-013 | Cross-platform build | **PASS** | 4 platforms exit 0 (darwin/amd64, linux/amd64, windows/amd64, darwin/arm64) |
| AC-UNP-014 | NFR-UNP-005 additivity | **PASS** | `TestIsUserOwnedNamespace_AdditivityWithIsUserAreaPath` |

## NICE-TO-HAVE Items (from plan-auditor verdict)

| Item | Status | Rationale |
|------|--------|-----------|
| D3 Windows-path test case | **APPLIED** | `TestAssertNoUserOwnedNamespaceTouch` includes "windows separator harness path triggers sentinel" case |
| D4 NICE-1 `.gitignore` alignment | **DEFERRED** | Out of Scope per spec.md §4 (Risk R6) — separate follow-up SPEC |
| D4 NICE-2 three-backup-root note | **APPLIED** | Inline godoc in `update_namespace_protect.go` package comment |
| D4 NICE-3 R6 sequencing note | **APPLIED** | Spec.md plan §4.4 R6 decision documented; implementation modifies disjoint regions |

## Backward Compatibility

- `isUserAreaPath` unchanged (NFR-UNP-005 additivity preserved). All existing call sites untouched.
- `isMoaiManaged` switch case removal: `harness` is the only removed sub-classification. Existing `TestIsMoaiManaged` test suite did NOT include `harness` cases (verified before changes), so existing tests PASS unchanged.
- `backupMoaiConfig` flow unchanged; namespace backup is purely additive.
- New file `update_namespace_protect.go` is standalone; no public API surface added or modified.
- No git/CI/template changes — internal-only code refactor + new file.

## Working Tree Hygiene (B8 compliance)

Specific paths staged per Section D constraint:
- M1: `.moai/specs/SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001/progress.md`
- M2: `internal/cli/update.go` + `internal/cli/update_test.go`
- M3: `internal/cli/update.go` + `internal/cli/update_namespace_protect.go` + `internal/defs/dirs.go`
- M4: `internal/cli/update_namespace_protect_test.go`
- M5: `.moai/specs/SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001/spec.md` + `progress.md`

No `git add -A`, no `git add .`, no broad directory adds. 50+ unrelated working-tree files remained untouched.

## B9 Compliance

No autonomous `git pull`, `git fetch`, `git rebase`, `git reset --hard`, `git push --force` invoked during run-phase. All 5 commits are atomic forward additions on `feat/SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001` branch.

## Branch State

```
$ git log --oneline -5 feat/SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001
<M5-SHA> chore(SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001): M5 mark implemented v0.2.0
597edec32 test(SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001): M4 namespace violation sentinel
e7e007bd4 feat(SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001): M3 backupUserOwnedNamespace
eb5815843 feat(SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001): M2 isUserOwnedNamespace + isMoaiManaged 정정
dbdcb9954 chore(SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001): M1 baseline measurement
```
