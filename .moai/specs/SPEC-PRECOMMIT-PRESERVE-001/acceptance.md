# SPEC-PRECOMMIT-PRESERVE-001 — Acceptance Criteria

## How to read a criterion

Each criterion carries five parts. The last two are what make it a check rather than a wish.

- **Given / When / Then** — the scenario, binary-testable.
- **Decides** — the command whose result is the verdict.
- **Baseline** — what that command returns against the untouched tree at `294b4b6ab`.
- **Mutant** — a concrete implementation that would **pass this criterion while violating the
  requirement it protects**, plus what defeats it. A criterion with no constructible mutant is too
  weak and must be rewritten.
- **Failing input** — the input against which the check must be **observed failing** before it is
  considered complete. A check never seen red has not been shown to observe anything.

Criteria whose pass value equals the baseline are labelled **behaviour-preserving**; they exist to
catch regression and are the only ones permitted to be un-failable on the pre-implementation tree.
Their "failing input" is therefore a mutation of the *implementation*, not of the tree.

**The mutant this whole card is about**, stated once because several criteria must jointly defeat
it: an implementation that overwrites a user-modified hook *correctly and recoverably* — right
content, backup written — but **prints nothing**. Any criterion that asserts only post-state file
contents lets this mutant through. AC-PCP-004, AC-PCP-006 and AC-PCP-007 assert the notice itself,
which is the only thing that stops it.

Test names are proposals; a differently-named test satisfying the same Given-When-Then is fine.

## §D AC Matrix

### AC-PCP-001 — provenance record is written (REQ-PCP-001)

**Given** a repository with no `.git/hooks/pre-commit`,
**When** `InstallPreCommitHook(false)` runs, **and then** runs a second time with a
`preCommitHookContent` that differs,
**Then** after each run `.git/hooks/.moai-pre-commit.sha256` contains the hex SHA-256 of the content
written by *that* run.

- **Decides**: `go test ./internal/cli/ -run TestPreCommitProvenanceRecorded -count=1`
- **Baseline**: test absent; `grep -c 'sha256' internal/cli/hook_install_precommit.go @294b4b6ab` → `0`
- **Mutant**: writes the sidecar once at first install and never refreshes it. Passes a single-run
  assertion, then permanently misclassifies every post-upgrade hook as user-modified. Defeated by
  the second run in this criterion — a single-install assertion is insufficient and must not be
  substituted.
- **Failing input**: a stub that creates `.moai-pre-commit.sha256` empty. Must be observed red.

### AC-PCP-002 — a version bump stays silent (REQ-PCP-002)

**Given** an installed hook whose bytes match the recorded digest,
**When** the installer runs with a **different** `preCommitHookContent`,
**Then** the hook is replaced, no file matching `pre-commit.bak.*` exists, and the captured output
contains no backup notice.

- **Decides**: `go test ./internal/cli/ -run TestPreCommitVersionBumpIsSilent -count=1`
- **Baseline**: test absent. Today's installer reaches this outcome by accident (it never notifies
  at all), not by decision.
- **Mutant**: an implementation that never notifies under any circumstance. Passes trivially and
  violates REQ-PCP-006. Defeated by AC-PCP-003/004/006, which demand a notice in the sibling case —
  this criterion is only sound *paired with* them.
- **Failing input**: an implementation that backs up whenever `installed != incoming` (the naive
  two-way design). Must be observed red — this is the noise regression the design exists to prevent.

### AC-PCP-003 — a modified hook is backed up before replacement (REQ-PCP-003)

**Given** a marker-bearing hook altered after installation (digest mismatch),
**When** the installer runs,
**Then** a file matching `.git/hooks/pre-commit.bak.*` exists whose bytes equal the **pre-run** hook
content.

- **Decides**: `go test ./internal/cli/ -run TestPreCommitModifiedHookIsBackedUp -count=1`
- **Baseline**: `grep -c 'pre-commit\.bak' internal/cli/hook_install_precommit.go @294b4b6ab` → `0`
- **Mutant**: backs up on *every* install, unmodified hooks included. Passes, and violates
  REQ-PCP-002 by making every version bump noisy. Defeated by AC-PCP-002.
- **Failing input**: today's installer at `294b4b6ab` — no backup is written at all.

### AC-PCP-004 — the backup is disclosed on stderr (REQ-PCP-004)

**Given** the AC-PCP-003 scenario,
**When** `installPreCommitHookOptional` runs against a captured writer,
**Then** the captured output contains **all three** of: the backup file's path, a statement that the
hook was replaced, and the string `pre-commit.local`.

- **Decides**: `go test ./internal/cli/ -run TestPreCommitBackupNoticeContent -count=1`
- **Baseline**: the only output is `  Pre-commit hook installed (.git/hooks/pre-commit)`
  (`internal/cli/hook_install_precommit.go:172 @294b4b6ab`) — it contains none of the three.
- **Mutant**: prints the notice but never writes the backup — all disclosure, no artifact. Passes,
  and violates REQ-PCP-003. Defeated by AC-PCP-003 and, in a single case, by AC-PCP-006.
- **Failing input**: the **card's headline mutant** — an implementation that backs up and overwrites
  correctly but prints nothing. Must be observed red here; this is the criterion that catches it.

### AC-PCP-005 — a legacy install with no record degrades safely (REQ-PCP-005)

**Given** a marker-bearing hook and no `.moai-pre-commit.sha256`:
**(a)** content differing from the incoming content → **Then** a backup is taken and a notice
emitted; **(b)** content identical to the incoming content → **Then** no backup is taken.

- **Decides**: `go test ./internal/cli/ -run TestPreCommitLegacyNoRecord -count=1` (both sub-cases)
- **Baseline**: test absent; today both sub-cases overwrite silently.
- **Mutant**: treats a missing record as "unmodified" and stays silent. Passes sub-case (b) and
  violates REQ-PCP-006 for the entire pre-upgrade population — the migration case, which is every
  existing installation. Defeated only because sub-case (a) is mandatory; (b) alone is not
  acceptable evidence.
- **Failing input**: sub-case (a) run against a "silent when unknown" implementation. Must be
  observed red.

### AC-PCP-006 — no silent replacement, whatever the policy (REQ-PCP-006)

**Given** the AC-PCP-003 scenario,
**When** the installer runs,
**Then** the single run produces **both** a backup file **and** a notice naming it. Neither alone
satisfies this criterion.

- **Decides**: `go test ./internal/cli/ -run TestPreCommitNoSilentReplacement -count=1`
- **Baseline**: today's run produces neither.
- **Mutant**: writes the backup, then swallows a failed notice write (or vice versa), so one artifact
  is present and the other silently absent. Passes any criterion that checks them in *separate*
  tests. Defeated by asserting both within one case — which is why this criterion is not redundant
  with AC-PCP-003 plus AC-PCP-004.
- **Failing input**: an implementation with the notice line deleted, backup retained. Must be
  observed red. **This criterion must also be re-run and still pass under any future change to the
  overwrite policy** — it is the invariant, not a property of option (a).

### AC-PCP-007 — the success line is never the whole story (REQ-PCP-007)

**Given** the AC-PCP-003 scenario,
**When** `installPreCommitHookOptional` runs,
**Then** the captured output both (i) differs from the bare success line and (ii) contains the
backup path.

- **Decides**: `go test ./internal/cli/ -run TestPreCommitBackupOutputNotBareSuccess -count=1`
- **Baseline**: the captured output **is** exactly
  `  Pre-commit hook installed (.git/hooks/pre-commit)\n`.
- **Mutant**: appends a blank line or trailing whitespace to the success line. Output now differs
  from the bare line, so a pure inequality check passes while the user learns nothing. Defeated by
  clause (ii) — inequality alone is not an acceptable implementation of this criterion.
- **Failing input**: the whitespace-padded success line. Must be observed red against clause (ii).

### AC-PCP-008 — marker-absent behaviour is unchanged (REQ-PCP-008) — behaviour-preserving

**Given** an existing hook with no MoAI marker,
**When** the installer runs,
**Then** it returns `ErrUserHookExists`, the hook bytes are unchanged, and no `pre-commit.bak.*`
file is created.

- **Decides**: the existing marker-absent case in `internal/cli/hook_install_precommit_test.go`,
  extended with the no-backup assertion
- **Baseline**: returns `ErrUserHookExists`, hook unchanged, no backup
  (`internal/cli/hook_install_precommit.go:145 @294b4b6ab`) — **behaviour-preserving**
- **Mutant**: an implementation that returns `ErrUserHookExists` for everything and never installs
  anything. Passes, and violates REQ-PCP-001/002. Defeated by AC-PCP-001 and AC-PCP-002.
- **Failing input** (implementation mutation, since the tree cannot fail it): an implementation that
  routes marker-less hooks into the new backup-and-overwrite path. Must be observed red.

### AC-PCP-009 — backups do not clobber each other (REQ-PCP-009)

**Given** a file already present at the backup path the installer would choose,
**When** a second backup is taken,
**Then** the pre-existing file's bytes are unchanged **and** a second, distinctly-named backup
exists.

- **Decides**: `go test ./internal/cli/ -run TestPreCommitBackupNoClobber -count=1`
- **Baseline**: no backup path exists, so the hazard is unreachable and untested.
- **Mutant**: skips the backup entirely when the path is occupied. Nothing is clobbered — so a
  criterion asserting only "the old file is unchanged" passes — while the new hook is lost
  unrecorded, violating REQ-PCP-003. Defeated by the second clause requiring a new backup to exist.
- **Failing input**: a fixed-filename backup implementation (`pre-commit.bak`, no timestamp). Must
  be observed red.

### AC-PCP-010 — backup failure does not fail update/init (REQ-PCP-010)

**Given** a hooks directory in which the backup write cannot succeed,
**When** `installPreCommitHookOptional` runs,
**Then** it returns normally, emits a warning naming the failure, and does not panic or abort the
caller.

- **Decides**: `go test ./internal/cli/ -run TestPreCommitBackupFailureNonFatal -count=1`
- **Baseline**: no backup is attempted, so no failure mode exists to be non-fatal.
- **Mutant**: never attempts a backup at all — cannot fail, so passes. Violates REQ-PCP-003.
  Defeated by AC-PCP-003. A second mutant: swallows the failure with no warning, which the "emits a
  warning" clause defeats.
- **Failing input**: an implementation that returns the backup error to the caller and aborts
  `moai update`. Must be observed red.

### AC-PCP-011 — the local hook runs and its status propagates (REQ-PCP-011)

**Given** an executable `.git/hooks/pre-commit.local` that exits `3`,
**When** the installed pre-commit hook runs with nothing staged,
**Then** the hook exits `3`.

- **Decides**: a shell-driven test invoking the installed hook, plus
  `grep -c 'pre-commit\.local' internal/template/templates/.git_hooks/pre-commit`
- **Baseline**: grep → `0` `@294b4b6ab`; the hook ends at `exit 0` with no delegation
- **Mutant**: executes `pre-commit.local` but discards its status and exits `0` regardless. Passes
  any criterion asserting only that the script *ran*. Defeated by asserting the specific non-zero
  status — which is why the criterion names `3` rather than "non-zero success/failure".
- **Failing input**: today's hook at `294b4b6ab`, which exits `0`. Must be observed red.

### AC-PCP-012 — absent local hook preserves today's behaviour (REQ-PCP-012) — behaviour-preserving

**Given** no `.git/hooks/pre-commit.local`, and separately one present but not executable,
**When** the hook runs with nothing staged and `moai` absent from PATH,
**Then** it exits `0` in both sub-cases.

- **Decides**: a shell-driven test invoking the installed hook in both sub-cases
- **Baseline**: exits `0` — **behaviour-preserving**
- **Mutant**: a hook that never delegates at all. Passes trivially and violates REQ-PCP-011.
  Defeated by AC-PCP-011.
- **Failing input** (implementation mutation): an unguarded invocation under `set -eu`, which aborts
  non-zero when the file is missing. Must be observed red — this is the realistic regression.

### AC-PCP-013 — constant/template parity holds (REQ-PCP-013) — behaviour-preserving

**Given** the tree after implementation,
**When** `go test ./internal/cli/ -run TestPreCommitTemplateMatchesConstant -count=1 -v` runs,
**Then** it reports PASS and **not** SKIP.

- **Decides**: that command, with the run's skip status inspected
- **Baseline**: passes; test verified present at
  `internal/cli/hook_install_precommit_test.go:38 @294b4b6ab` — **behaviour-preserving**
- **Mutant**: deleting or moving the template file. The test's own `t.Skipf` branch
  (`hook_install_precommit_test.go:45 @294b4b6ab`) then makes the suite green while parity is
  entirely unverified. This is a live hazard, not a hypothetical. Defeated by requiring PASS and
  rejecting SKIP.
- **Failing input**: editing `preCommitHookContent` without the template twin. Must be observed red.

### AC-PCP-014 — attribution is three-way, not two-way (REQ-PCP-014)

**Given** an installed hook whose bytes equal the **previous** `preCommitHookContent`, a provenance
record matching it, and an incoming content that differs from both,
**When** the installer runs,
**Then** the hook is classified unmodified and replaced with no backup and no notice; **and given**
the same installed bytes with a provenance record that does **not** match them, **Then** it is
classified user-modified.

- **Decides**: `go test ./internal/cli/ -run TestPreCommitThreeWayAttribution -count=1` (both cases)
- **Baseline**: test absent; `grep -c 'sha256' internal/cli/hook_install_precommit.go @294b4b6ab` →
  `0`, so no third operand exists to compare against.
- **Mutant**: **the two-way implementation** — compares the installed file against the incoming
  content only. It passes the second case (a difference is present, so it says user-modified, by
  luck) and **fails the first**, which is precisely why the first case is mandatory. Any
  single-case version of this criterion is satisfiable by a two-way design and must be rejected.
- **Failing input**: the two-way implementation, run against case one. Must be observed red.

## §D.1 Edge cases

| Case | Expected |
|---|---|
| `.git/hooks/` absent | created, as today (`internal/cli/hook_install_precommit.go:131 @294b4b6ab`) |
| `skip=true` | no-op — no write, no backup, no provenance record |
| Record present but malformed or unreadable | treated as absent → REQ-PCP-005 legacy path |
| Installed hook unreadable | existing `read existing hook` error path, unchanged |
| Two backups needed within the same timestamp granularity | AC-PCP-009 governs; distinct paths required |
| Hook restored by hand from a `.bak` file | record mismatches → classified user-modified → backed up again. Noisy but safe; the correct direction of error |

## §D.2 Quality gates

- `go vet ./internal/cli/...` clean
- `go test ./internal/cli/... -count=1` passing
- `go test ./internal/template/... -count=1` passing
- `make build` after any change to `preCommitHookContent` or the template twin

## §D.3 Definition of Done

- All 14 AC pass; the 3 behaviour-preserving ones show no change from baseline.
- **Every criterion has been observed failing against its stated failing input** before being
  accepted. A criterion that has only ever been green is not evidence.
- No prompt is introduced on any path (§C.3 of spec.md).
- REQ-PCP-011's paired edit lands as its own final commit (§C.2).
- `~/go/bin/moai spec lint .moai/specs/SPEC-PRECOMMIT-PRESERVE-001/spec.md` clean.
