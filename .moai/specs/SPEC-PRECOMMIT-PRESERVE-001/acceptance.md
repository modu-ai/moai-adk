# SPEC-PRECOMMIT-PRESERVE-001 — Acceptance Criteria

Every criterion below states the command that decides it, the value it returns against the
**untouched tree at `294b4b6ab`**, and the value that constitutes a pass. Criteria whose pass value
equals today's value are labelled **behaviour-preserving** — they exist to catch regression, and
they are the only criteria permitted to be un-failable on the pre-implementation tree.

Test-name placeholders (`TestPreCommit…`) are the names run-phase is expected to create; a
differently-named test satisfying the same Given-When-Then is acceptable.

## §D AC Matrix

### AC-PCP-001 — provenance record is written (REQ-PCP-001)

**Given** a repository with no `.git/hooks/pre-commit`,
**When** `InstallPreCommitHook(false)` runs,
**Then** `.git/hooks/.moai-pre-commit.sha256` exists and contains the hex SHA-256 of
`preCommitHookContent`.

- Decides: `go test ./internal/cli/ -run TestPreCommitProvenanceRecorded -count=1`
- Today: the test does not exist; `grep -c 'sha256' internal/cli/hook_install_precommit.go` → `0`
- Pass: test present and passing; the grep is non-zero

### AC-PCP-002 — matching provenance overwrites silently (REQ-PCP-002)

**Given** an installed hook whose bytes hash to the value recorded in the provenance sidecar (the
state produced by any prior MoAI install, including one of an older version),
**When** the installer runs with different `preCommitHookContent`,
**Then** the hook is replaced, no `pre-commit.bak.*` file is created, and stderr contains no backup
notice.

- Decides: `go test ./internal/cli/ -run TestPreCommitVersionBumpIsSilent -count=1`
- Today: no such test; the installer overwrites unconditionally so this outcome is reached by
  accident rather than by decision
- Pass: test present and passing, asserting **zero** files matching `pre-commit.bak.*`

### AC-PCP-003 — a modified hook is backed up before replacement (REQ-PCP-003)

**Given** an installed hook carrying the MoAI marker whose content has been altered after
installation (provenance mismatch),
**When** the installer runs,
**Then** a file matching `.git/hooks/pre-commit.bak.*` exists whose bytes equal the pre-run hook
content.

- Decides: `go test ./internal/cli/ -run TestPreCommitModifiedHookIsBackedUp -count=1`
- Today: `grep -c 'pre-commit\.bak' internal/cli/hook_install_precommit.go` → `0`; the pre-run
  content is destroyed with no copy
- Pass: test present and passing; the grep is non-zero

### AC-PCP-004 — the backup is disclosed on stderr (REQ-PCP-004)

**Given** the AC-PCP-003 scenario,
**When** `installPreCommitHookOptional` runs against a captured writer,
**Then** the captured output contains the backup path, states the hook was replaced, and names
`pre-commit.local`.

- Decides: `go test ./internal/cli/ -run TestPreCommitBackupNoticeContent -count=1`
- Today: the only output is `  Pre-commit hook installed (.git/hooks/pre-commit)`
  (`hook_install_precommit.go:172`) — it contains none of the three
- Pass: test present and passing on all three substrings

### AC-PCP-005 — a legacy install with no sidecar degrades safely (REQ-PCP-005)

**Given** a marker-bearing hook, no `.moai-pre-commit.sha256`, and content differing from
`preCommitHookContent`,
**When** the installer runs,
**Then** a backup is taken and a notice emitted; **and given** the same setup with content
*identical* to `preCommitHookContent`, no backup is taken.

- Decides: `go test ./internal/cli/ -run TestPreCommitLegacyNoSidecar -count=1` (both sub-cases)
- Today: no such test; both sub-cases overwrite silently
- Pass: test present and passing on both branches

### AC-PCP-006 — no silent overwrite of a modified hook (REQ-PCP-006)

**Given** the AC-PCP-003 scenario,
**When** the installer runs,
**Then** the run produces **both** a backup file and a notice — neither alone satisfies this.

- Decides: `go test ./internal/cli/ -run TestPreCommitNoSilentOverwrite -count=1`
- Today: the run produces neither
- Pass: test present, asserting both conditions in one case

### AC-PCP-007 — the success line is not the only output after a backup (REQ-PCP-007)

**Given** the AC-PCP-003 scenario,
**When** `installPreCommitHookOptional` runs,
**Then** the captured output is not equal to the bare success line.

- Decides: `go test ./internal/cli/ -run TestPreCommitBackupOutputNotBareSuccess -count=1`
- Today: the captured output **is** exactly `  Pre-commit hook installed (.git/hooks/pre-commit)\n`
- Pass: test present and passing

### AC-PCP-008 — marker-absent behaviour is unchanged (REQ-PCP-008) — behaviour-preserving

**Given** an existing hook with no MoAI marker,
**When** the installer runs,
**Then** it returns `ErrUserHookExists`, the hook bytes are unchanged, and no `pre-commit.bak.*`
file is created.

- Decides: the existing marker-absent test in `internal/cli/hook_install_precommit_test.go`, plus a
  new no-backup assertion
- Today: returns `ErrUserHookExists`, hook unchanged, no backup (`hook_install_precommit.go:145`)
- Pass: identical — **behaviour-preserving**. The added assertion is that the new backup path does
  not fire here

### AC-PCP-009 — backups do not clobber each other (REQ-PCP-009)

**Given** a `.git/hooks/pre-commit.bak.<ts>` already present at the path the installer would choose,
**When** a second backup is taken,
**Then** the pre-existing backup's bytes are unchanged and a second, distinctly-named backup exists.

- Decides: `go test ./internal/cli/ -run TestPreCommitBackupNoClobber -count=1`
- Today: no backup path exists, so the hazard is unreachable and untested
- Pass: test present and passing, asserting two distinct backup files

### AC-PCP-010 — backup failure does not fail update/init (REQ-PCP-010)

**Given** a hooks directory in which the backup write cannot succeed (e.g. read-only),
**When** `installPreCommitHookOptional` runs,
**Then** it returns normally and emits a warning; it does not panic and does not abort the caller.

- Decides: `go test ./internal/cli/ -run TestPreCommitBackupFailureNonFatal -count=1`
- Today: no backup is attempted, so no failure mode exists to be non-fatal
- Pass: test present and passing; the warning branch is exercised

### AC-PCP-011 — the local extension point runs and its status propagates (REQ-PCP-011)

**Given** an executable `.git/hooks/pre-commit.local` exiting non-zero,
**When** the pre-commit hook runs with nothing staged,
**Then** the hook exits with that same non-zero status.

- Decides: a shell-driven test invoking the installed hook, plus
  `grep -c 'pre-commit\.local' internal/template/templates/.git_hooks/pre-commit`
- Today: the grep → `0`; the hook ends at `exit 0` with no delegation
- Pass: grep non-zero **and** the exit-status test passing

### AC-PCP-012 — absent local hook preserves today's behaviour (REQ-PCP-012) — behaviour-preserving

**Given** no `.git/hooks/pre-commit.local`, or one that is present but not executable,
**When** the pre-commit hook runs with nothing staged and `moai` absent from PATH,
**Then** it exits `0`.

- Decides: a shell-driven test invoking the installed hook in both sub-cases
- Today: exits `0`
- Pass: identical — **behaviour-preserving**

### AC-PCP-013 — constant/template parity holds (REQ-PCP-013) — behaviour-preserving

**Given** the repository after this SPEC's implementation,
**When** `go test ./internal/cli/ -run TestPreCommitTemplateMatchesConstant -count=1` runs,
**Then** it passes without skipping.

- Decides: that command
- Today: passes (test verified present at `hook_install_precommit_test.go:38`)
- Pass: still passes — **behaviour-preserving**. This is the guard against REQ-PCP-011 being landed
  as a one-sided edit

## §D.1 Edge cases

| Case | Expected |
|---|---|
| `.git/hooks/` absent | created, as today (`hook_install_precommit.go:131`) |
| `skip=true` | no-op — no write, no backup, no provenance record |
| Sidecar present but unreadable or malformed | treated as absent → legacy path (AC-PCP-005) |
| Installed hook unreadable | existing `read existing hook` error path, unchanged |
| Two installs in the same second, both needing a backup | AC-PCP-009 governs; distinct paths required |

## §D.2 Quality gates

- `go vet ./internal/cli/...` clean
- `go test ./internal/cli/... -count=1` passing
- `go test ./internal/template/... -count=1` passing (template parity surface)
- `make build` run after any change to `preCommitHookContent` or the template twin

## §D.3 Definition of Done

- All 13 AC pass; the 4 behaviour-preserving ones show no change from baseline.
- No prompt is introduced on any path (§C.3).
- REQ-PCP-011's paired edit lands as its own final commit (§C.2).
- `~/go/bin/moai spec lint .moai/specs/SPEC-PRECOMMIT-PRESERVE-001/spec.md` clean.
