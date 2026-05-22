---
id: SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001
title: "moai update User-Owned Namespace Protection + Backup Standardization — Acceptance"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli/update.go, internal/cli/update_archive.go, internal/defs/dirs.go"
lifecycle: spec-anchored
tags: "update, namespace, backup, harness, protection, contract, tier-m, acceptance"
tier: M
---

# Acceptance Criteria — SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001

All acceptance criteria are **binary verifiable** (PASS/FAIL via concrete commands). Each AC carries an explicit verification command. Subjective judgments are prohibited per `agent-common-protocol.md` §Skeptical Evaluation Stance.

## §1. AC Matrix (Requirement ↔ AC Traceability)

| Requirement | Acceptance Criterion | Test Strategy |
|-------------|----------------------|---------------|
| REQ-UNP-001 | AC-UNP-001 | Integration test in `update_namespace_protect_test.go` |
| REQ-UNP-002 | AC-UNP-002 | Integration test |
| REQ-UNP-003 | AC-UNP-003 | Integration test |
| REQ-UNP-004 | AC-UNP-004 | Filesystem assertion post-update |
| REQ-UNP-005 | AC-UNP-005 | Violation simulation test |
| REQ-UNP-006 | AC-UNP-005 (sentinel) + AC-UNP-009 (exit code) | Stderr grep + exit code assertion |
| REQ-UNP-007 | AC-UNP-010 (backup atomicity marker) | Filesystem check for `.complete` marker |
| REQ-UNP-008 | N/A | Future-proof, not implemented in this SPEC |
| REQ-UNP-009 | AC-UNP-007, AC-UNP-008 | Integration test |
| REQ-UNP-010 | AC-UNP-006 (regex naming) | Regex assertion |
| NFR-UNP-001 | (not gated by AC; M5 smoke) | — |
| NFR-UNP-002 | AC-UNP-011 (build + race tests) | Implicit via test execution |
| NFR-UNP-003 | AC-UNP-013 (cross-platform) | Cross-build matrix |
| NFR-UNP-004 | AC-UNP-012 (idempotency) | Sequential invocation test |
| NFR-UNP-005 | AC-UNP-014 (additivity) | `isUserAreaPath` still callable |

## §2. Given-When-Then Scenarios

### AC-UNP-001 — `.claude/skills/my-harness-*` preservation

**Given** a project containing `.claude/skills/my-harness-test/SKILL.md` with arbitrary user content,
**When** `moai update` runs to completion,
**Then** the file `.claude/skills/my-harness-test/SKILL.md` exists with byte-identical content.

**Verification**:
```bash
# Setup (in test fixture / synthetic tmp project)
mkdir -p .claude/skills/my-harness-test
echo "user content $(date +%s)" > .claude/skills/my-harness-test/SKILL.md
EXPECTED_HASH=$(sha256sum .claude/skills/my-harness-test/SKILL.md | awk '{print $1}')

# Run update
moai update --force

# Verify
test -f .claude/skills/my-harness-test/SKILL.md
ACTUAL_HASH=$(sha256sum .claude/skills/my-harness-test/SKILL.md | awk '{print $1}')
test "$EXPECTED_HASH" = "$ACTUAL_HASH"  # exit 0 == PASS
```

### AC-UNP-002 — `.claude/agents/harness/` preservation

**Given** a project containing `.claude/agents/harness/test-specialist.md` (a user-authored harness teammate),
**When** `moai update` runs to completion,
**Then** the file `.claude/agents/harness/test-specialist.md` exists with byte-identical content.

**Verification**:
```bash
mkdir -p .claude/agents/harness
echo "user harness agent" > .claude/agents/harness/test-specialist.md
EXPECTED=$(sha256sum .claude/agents/harness/test-specialist.md | awk '{print $1}')

moai update --force

ACTUAL=$(sha256sum .claude/agents/harness/test-specialist.md | awk '{print $1}')
test "$EXPECTED" = "$ACTUAL"  # exit 0 == PASS
```

### AC-UNP-003 — `.moai/harness/` preservation

**Given** a project containing `.moai/harness/test.md` (a user-authored harness extension),
**When** `moai update` runs to completion,
**Then** the file `.moai/harness/test.md` exists with byte-identical content.

**Verification**:
```bash
mkdir -p .moai/harness
echo "harness extension" > .moai/harness/test.md
EXPECTED=$(sha256sum .moai/harness/test.md | awk '{print $1}')

moai update --force

ACTUAL=$(sha256sum .moai/harness/test.md | awk '{print $1}')
test "$EXPECTED" = "$ACTUAL"  # exit 0 == PASS
```

### AC-UNP-004 — Backup directory created and populated

**Given** a project containing user-owned artifacts in any of the three protected categories,
**When** `moai update` runs to completion,
**Then** a directory matching `.moai/backups/update-*/` exists and contains a backup of every user-owned artifact present before update.

**Verification**:
```bash
mkdir -p .claude/skills/my-harness-x .claude/agents/harness .moai/harness
echo "a" > .claude/skills/my-harness-x/file
echo "b" > .claude/agents/harness/file
echo "c" > .moai/harness/file

moai update --force

# Find the most recent backup directory
BACKUP=$(ls -1d .moai/backups/update-* 2>/dev/null | tail -1)
test -n "$BACKUP" && test -d "$BACKUP"  # exit 0 == backup exists
test -f "$BACKUP/.claude/skills/my-harness-x/file"
test -f "$BACKUP/.claude/agents/harness/file"
test -f "$BACKUP/.moai/harness/file"
```

### AC-UNP-005 — Violation sentinel emitted on destructive attempt

**Given** a synthetic `cmdUpdate` invocation where the deploy-plan is artificially modified (via test hook) to include a path matching `isUserOwnedNamespace`,
**When** `assertNoUserOwnedNamespaceTouch` runs,
**Then** stderr contains the literal string `UPDATE_USER_NAMESPACE_VIOLATION` and the exit code is non-zero, and no file modification has occurred.

**Verification** (Go test):
```go
func TestUpdate_NamespaceViolation_AbortsBeforeWrite(t *testing.T) {
    tmpDir := t.TempDir()
    setupSyntheticProject(t, tmpDir)

    plan := []deployOp{{rel: ".claude/agents/harness/contraband.md", action: "overwrite"}}
    err := assertNoUserOwnedNamespaceTouch(plan)

    if err == nil {
        t.Fatal("expected sentinel violation, got nil")
    }
    if !strings.Contains(err.Error(), "UPDATE_USER_NAMESPACE_VIOLATION") {
        t.Fatalf("expected sentinel, got: %s", err.Error())
    }
    // Assert no filesystem mutation occurred
    if _, statErr := os.Stat(filepath.Join(tmpDir, ".claude/agents/harness/contraband.md")); !os.IsNotExist(statErr) {
        t.Fatal("file was created before sentinel fired")
    }
}
```

### AC-UNP-006 — Backup directory regex compliance

**Given** a successful `moai update` invocation that created a user-owned namespace backup,
**When** the backup directory name is inspected,
**Then** the name matches the regex `^\.moai/backups/update-\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}Z$`.

**Verification**:
```bash
BACKUP=$(ls -1d .moai/backups/update-* | tail -1)
echo "$BACKUP" | grep -qE '^\.moai/backups/update-[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}-[0-9]{2}-[0-9]{2}Z$'  # exit 0 == PASS
```

### AC-UNP-007 — User direct-added agent file preserved

**Given** a project containing `.claude/agents/custom-agent.md` (filename does not begin with any system-agent prefix and is not under `{core,expert,meta,harness}/`),
**When** `moai update` runs to completion,
**Then** the file exists with byte-identical content.

**Verification**:
```bash
mkdir -p .claude/agents
echo "user-authored top-level agent" > .claude/agents/custom-agent.md
EXPECTED=$(sha256sum .claude/agents/custom-agent.md | awk '{print $1}')

moai update --force

ACTUAL=$(sha256sum .claude/agents/custom-agent.md | awk '{print $1}')
test "$EXPECTED" = "$ACTUAL"
```

### AC-UNP-008 — User direct-added skill directory preserved

**Given** a project containing `.claude/skills/custom-skill/SKILL.md` (directory name does not begin with `moai-` and is not equal to `moai`),
**When** `moai update` runs to completion,
**Then** the directory and all files within remain byte-identical.

**Verification**:
```bash
mkdir -p .claude/skills/custom-skill
echo "user skill" > .claude/skills/custom-skill/SKILL.md
EXPECTED=$(sha256sum .claude/skills/custom-skill/SKILL.md | awk '{print $1}')

moai update --force

ACTUAL=$(sha256sum .claude/skills/custom-skill/SKILL.md | awk '{print $1}')
test "$EXPECTED" = "$ACTUAL"
```

### AC-UNP-009 — New test file exists and passes

**Given** the run-phase implementation is complete,
**When** `go test -race ./internal/cli/... -run TestUpdate_Namespace -v` runs,
**Then** the test file `internal/cli/update_namespace_protect_test.go` exists, contains at least 5 named table-driven cases (verified by case count), and every case exits PASS.

**Verification**:
```bash
test -f internal/cli/update_namespace_protect_test.go
# Count table case names — must be >= 5
CASES=$(grep -cE '^\s+name:\s+"' internal/cli/update_namespace_protect_test.go)
test "$CASES" -ge 5
# Execute
go test -race ./internal/cli/... -run TestUpdate_Namespace -v 2>&1 | tee /tmp/namespace_test.log
! grep -q -- '--- FAIL' /tmp/namespace_test.log  # exit 0 == no failures
```

### AC-UNP-010 — Archive backup separation maintained + atomicity marker present (REQ-UNP-007 + REQ-UNP-010)

**Given** the implementation is complete,
**When** filesystem state is inspected after `moai update --force` on a project that triggers both archive-drift AND user-owned namespace backup,
**Then** (a) archive-drift backup resides under `.moai/archive/skills/v2.16-drift-*/`, (b) user-owned namespace backup resides under `.moai/backups/update-*/`, (c) the two directory trees do not overlap, **and (d) every completed namespace backup directory contains a `.complete` marker file (REQ-UNP-007 atomicity)**.

**Verification**:
```bash
moai update --force
# Both directory trees must coexist disjoint
DRIFT_COUNT=$(find .moai/archive/skills -maxdepth 1 -type d -name 'v2.16-drift-*' | wc -l)
BACKUP_COUNT=$(find .moai/backups -maxdepth 1 -type d -name 'update-*' | wc -l)
# At least one of each tree exists (when applicable scenario fixture is set up)
test "$DRIFT_COUNT" -ge 0  # 0 is OK if no drift scenario
test "$BACKUP_COUNT" -ge 1  # at least 1 user-owned backup expected
# Disjointness check (no path under .moai/backups starts with archive/skills)
! find .moai/backups -path '*/archive/skills/*' 2>/dev/null | grep -q .
# REQ-UNP-007 atomicity marker: .complete file MUST exist after backup completion
LATEST_BACKUP=$(find .moai/backups -maxdepth 1 -type d -name 'update-*' | sort | tail -1)
test -f "$LATEST_BACKUP/.complete"
```

### AC-UNP-011 — Build + race tests clean

**Given** the run-phase implementation is complete on the working tree,
**When** `make build` and `go test -race ./internal/cli/...` execute,
**Then** both exit with code 0.

**Verification**:
```bash
cd /Users/goos/MoAI/moai-adk-go
make build
go test -race ./internal/cli/...
echo "EXIT=$?"  # must be 0
```

### AC-UNP-012 — Idempotency: duplicate-second invocation handling

**Given** two successive `moai update` invocations launched within the same UTC second,
**When** both complete,
**Then** either (a) exactly one backup directory exists for that second and the second invocation skipped the backup because user-owned content was byte-identical, OR (b) two directories exist with the second using suffix `-1` (NFR-UNP-004).

**Verification**:
```bash
# Launch two updates rapidly
moai update --force &
moai update --force &
wait
# Inspect — count must be 1 OR 2 (with -1 suffix on the second)
COUNT=$(ls -1d .moai/backups/update-* 2>/dev/null | wc -l)
test "$COUNT" -ge 1 && test "$COUNT" -le 2
# If COUNT == 2, the second directory MUST have a numeric suffix
if [ "$COUNT" -eq 2 ]; then
    ls -1d .moai/backups/update-* | tail -1 | grep -qE '\-[0-9]+$'
fi
```

### AC-UNP-013 — Cross-platform build PASS

**Given** the run-phase implementation is complete,
**When** cross-platform builds run from the host (darwin),
**Then** all three target builds exit with code 0.

**Verification**:
```bash
cd /Users/goos/MoAI/moai-adk-go
GOOS=darwin  GOARCH=amd64 go build -o /tmp/moai-darwin   ./cmd/moai
GOOS=linux   GOARCH=amd64 go build -o /tmp/moai-linux    ./cmd/moai
GOOS=windows GOARCH=amd64 go build -o /tmp/moai-windows.exe ./cmd/moai
ls -la /tmp/moai-darwin /tmp/moai-linux /tmp/moai-windows.exe  # all must exist
```

### AC-UNP-014 — `isUserAreaPath` additivity preserved

**Given** the run-phase implementation is complete,
**When** `internal/cli/update.go` is inspected,
**Then** the function `isUserAreaPath` (existing) is still present and unchanged in its return semantics for the two patterns it previously covered (NFR-UNP-005).

**Verification**:
```bash
grep -n 'func isUserAreaPath' internal/cli/update.go  # function exists
# Behavior preservation — function still returns true for previously-covered patterns
go test -race ./internal/cli/... -run 'TestIsUserAreaPath'  # PASS
```

## §3. Edge Cases

### EC-UNP-001 — Empty user-owned namespace (no files in any of the three protected categories)

**Expectation**: No `.moai/backups/update-*/` directory is created. `moai update` completes normally.

**Verification**: After `moai update --force` on a project with no user-owned content, `find .moai/backups -maxdepth 1 -type d -name 'update-*' | wc -l` returns `0`.

### EC-UNP-002 — Pre-existing `.moai/backups/update-*/` from previous run

**Expectation**: New invocation creates a new directory with a newer timestamp. No deletion of old backups by this SPEC (out of scope — see `spec.md` §4).

### EC-UNP-003 — User-owned path coincides with template overlay target

**Expectation**: Template overlay write loop calls `isUserOwnedNamespace` and skips. No silent overwrite. If the loop is somehow bypassed, `assertNoUserOwnedNamespaceTouch` catches it via REQ-UNP-006 / AC-UNP-005.

### EC-UNP-004 — Symlinks in user-owned namespace

**Expectation**: Symlinks are copied as symlinks (not dereferenced) into the backup directory. Behavior matches existing `copyFileFn` semantics in `update.go`.

### EC-UNP-005 — Read-only files in user-owned namespace

**Expectation**: Backup copy preserves source mode bits up to the limit of `defs.FilePerm` (`0o644`). Subsequent `moai update` runs do not touch the original read-only file.

### EC-UNP-006 — Concurrent invocation collision (NFR-UNP-004)

**Expectation**: Handled by AC-UNP-012 — suffix `-1`, `-2`, or skip-if-byte-identical.

### EC-UNP-007 — Backup creation fails mid-copy

**Expectation**: Function returns error, the `cmdUpdate` flow aborts, no destructive operation runs. The partially-written backup directory is removed (defensive cleanup) similar to the existing pattern at `update.go:1415` and `update_archive.go:91`.

## §4. Quality Gate Criteria (Definition of Done)

The SPEC is **DONE** when all of the following PASS:

- [ ] `make build` exits 0
- [ ] `go test -race ./internal/cli/...` exits 0
- [ ] `golangci-lint run ./internal/cli/...` exits 0 NEW issues over baseline
- [ ] Cross-platform builds (darwin/linux/windows) exit 0 each
- [ ] `internal/cli/update_namespace_protect_test.go` exists with >= 5 table cases, all PASS
- [ ] `isUserAreaPath` unchanged, `isUserOwnedNamespace` added, `isMoaiManaged` amended (harness removed from switch)
- [ ] `internal/defs/dirs.go` contains `NamespaceBackupsSubdir = "backups"` constant
- [ ] AC-UNP-001 through AC-UNP-014 each individually verified PASS
- [ ] Sentinel `UPDATE_USER_NAMESPACE_VIOLATION` grep-able in stderr on violation
- [ ] No `git pull/fetch/rebase` invoked autonomously by manager-develop during run-phase (B9 prohibition compliance)
- [ ] No drive-by refactor outside the named files (Agent Core Behavior 5 — Scope Discipline)
- [ ] `spec.md` `status` updated to `implemented`, `version` bumped to `0.2.0`, `updated:` date incremented
- [ ] `progress.md` written with M1-M5 evidence and AC matrix

## §5. Out of Scope (cross-reference)

### §5 Out of Scope

- Backup retention/cleanup verification for `.moai/backups/update-*/` (no AC at this layer; deferred to follow-up SPEC)
- `--no-backup` flag verification (REQ-UNP-008 is future-proof; no AC in this SPEC)
- Sync mechanism regression beyond the named files (no AC; out of scope for this SPEC)
- Sequential Thinking deprecation acceptance (handled by `SPEC-V3R6-SEQ-THINKING-RETIRE-001`)
- Migration of existing `.moai-backups/` content (no AC; coexistence accepted)
- Archive-drift backup behavior (covered by `SPEC-V3R6-UPDATE-ARCHIVE-CONTRACT-001`)

Canonical list: `spec.md` §4 Out of Scope.

## §6. Follow-up SPEC Candidates

Identified during plan-phase, deferred to separate SPECs:

1. **Backup retention/cleanup for `.moai/backups/update-*/`** — mirror or generalize `cleanup_old_backups`.
2. **Consolidation of three backup roots** — unify `.moai-backups/`, `.moai/archive/skills/v2.16-drift-*/`, `.moai/backups/update-*/`.
3. **`--no-backup` flag** — implement REQ-UNP-008.
4. **`.gitignore` template patch** — add `.moai/backups/` entry to `internal/template/templates/.gitignore`.
5. **Audit ledger** — log every namespace violation attempt to `.moai/logs/namespace-protect.log`.
