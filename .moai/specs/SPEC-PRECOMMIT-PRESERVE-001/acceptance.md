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

### AC-PCP-002 — a version bump stays silent (REQ-PCP-002) — implementation-mutant-falsifiable only

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
- **Failing input** (implementation mutation, not a tree mutation): an implementation that backs up
  whenever `installed != incoming` (the naive two-way design). Must be observed red — this is the
  noise regression the design exists to prevent.
- **Falsification class**: this criterion's Given ("bytes match the recorded digest") is
  unconstructible on the untouched tree, because no sidecar mechanism exists there — it goes red at
  setup rather than by observing today's behaviour. It is therefore falsifiable **only** against a
  named implementation mutant, the same class as the three criteria labelled behaviour-preserving,
  and must not be counted as evidence about the pre-implementation tree.

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

### AC-PCP-004 — the backup is disclosed, and it reaches stderr (REQ-PCP-004)

**Given** the AC-PCP-003 scenario,
**When** `installPreCommitHookOptional` runs against **two** captured writers — the progress writer
and the warning writer,
**Then** (i) the **warning** writer's captured output contains **both** of: the backup file's path,
and a statement that the hook was replaced; **and** (ii) both call sites bind the warning-writer
parameter to the command's stderr.

The notice asserts these two elements and no third. It must **not** name `pre-commit.local`: that
extension point left this SPEC at v0.4.0 (spec.md §D Out of Scope), so a notice naming it would
point the user at a file the installed hook never reads. A test asserting the string
`pre-commit.local` here would therefore lock in a false instruction.

- **Decides**: `go test ./internal/cli/ -run TestPreCommitBackupNoticeContent -count=1` for clause
  (i); for clause (ii), **a Go test is the deciding check** — one that invokes each call site (or
  asserts its wiring) and confirms the warning-writer argument resolves to the command's
  `ErrOrStderr`. A grep is *not* sufficient on its own: under the correct implementation the update
  call site reads `installPreCommitHookOptional(projectRoot, getBoolFlag(cmd, "no-hooks"), out,
  errOut)` — the line contains `errOut`, and the `cmd.ErrOrStderr()` derivation lives two lines away
  at `update_template_sync.go:72 @7b2f42be0`. A checker grepping the call line for `ErrOrStderr`
  finds nothing even when the wiring is right, and a checker grepping for `errOut` cannot tell the
  right variable from an identically-named wrong one. Where a grep is used as a cheap smoke check,
  it must cover the derivation too (`grep -n 'errOut :=' internal/cli/update_template_sync.go`
  alongside the call-site line) — but the Go test is what decides.
- **Baseline**: the only output is `  Pre-commit hook installed (.git/hooks/pre-commit)`
  (`internal/cli/hook_install_precommit.go:172 @294b4b6ab`) — it contains none of the three, and no
  warning writer exists. Clause (ii) baseline: `update_template_sync.go:575 @7b2f42be0` passes
  `out` (bound to `cmd.OutOrStdout()` at `:69`), i.e. **stdout**.
- **Mutant**: prints the notice but never writes the backup — all disclosure, no artifact. Passes,
  and violates REQ-PCP-003. Defeated by AC-PCP-003 and, in a single case, by AC-PCP-006.
  A second mutant, which clause (ii) exists for: adds the warning-writer parameter but wires the
  `moai update` call site to the existing `out`. Clause (i) passes — the notice lands on *a*
  captured writer — while the notice is silently on stdout and a scripted
  `moai update >/dev/null` swallows it. Only clause (ii) catches this.
- **Failing input**: the **card's headline mutant** — an implementation that backs up and overwrites
  correctly but prints nothing. Must be observed red here against clause (i). Clause (ii) must be
  observed red against `update_template_sync.go` at `7b2f42be0` unchanged.

### AC-PCP-005 — a legacy install with no record degrades safely (REQ-PCP-005)

**Given** a marker-bearing hook and no `.moai-pre-commit.sha256`:
**(a)** content differing from the incoming content → **Then** a backup is taken and a notice
emitted; **(b)** content identical to the incoming content → **Then** no backup is taken;
**(c)** content that is the **`v3.1.2` released hook body read from git** — not a synthetic fixture,
not the incoming constant — → **Then** no backup is taken, no notice is emitted, and the provenance
record is written.

Sub-case (c) is the real installed base, and it exists because nothing else measures it. Sub-case
(b) builds its fixture from the incoming content, so it passes by construction whatever the shipped
bytes happen to be; it proves the code path, not the population. (c) fixes the fixture to bytes
that actually shipped — every `v3.1.0`-`v3.1.2` install carries them — and so it is also what keeps
this SPEC honest about leaving the hook body untouched (spec.md §C.1): the day someone changes
`preCommitHookContent`, (c) goes red and says so, in the place where the consequence lands.

- **Decides**: `go test ./internal/cli/ -run TestPreCommitLegacyNoRecord -count=1` (all three
  sub-cases). Sub-case (c)'s fixture is obtained from git rather than hand-copied —
  `git show v3.1.2:internal/template/templates/.git_hooks/pre-commit` — so it cannot silently drift
  into a copy of the current constant. A test that cannot reach git (shallow clone, no tags) must
  **skip loudly**, never pass quietly; a skipped (c) is a gap, not a pass.
- **Baseline**: test absent; today all three sub-cases overwrite silently. The `v3.1.2` blob is
  byte-identical to the current template at `7b2f42be0`
  (`git show v3.1.2:internal/template/templates/.git_hooks/pre-commit | cmp - <same path>` → rc 0),
  which is exactly why (c) passes on the correct implementation and is not a burden.
- **Mutant**: treats a missing record as "unmodified" and stays silent. Passes sub-cases (b) and
  (c) and violates REQ-PCP-006 for the entire pre-upgrade population — the migration case, which is
  every existing installation. Defeated only because sub-case (a) is mandatory; (b) and (c) alone
  are not acceptable evidence. A second mutant, which (c) exists for: an implementation correct on
  synthetic fixtures, shipped in a tree whose hook body has quietly changed — (b) still passes,
  (c) goes red and names the real regression.
- **Failing input**: sub-case (a) run against a "silent when unknown" implementation. Must be
  observed red. For sub-case (c), a tree with any edit to `preCommitHookContent` / the template
  twin — it must be observed red there before (c) is accepted, since a criterion that has only ever
  been green proves nothing.

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
**When** `installPreCommitHookOptional` runs against both captured writers,
**Then** (i) the warning writer's captured output is non-empty and (ii) contains the backup path —
so the run's total output is never the bare success line alone.

- **Decides**: `go test ./internal/cli/ -run TestPreCommitBackupOutputNotBareSuccess -count=1`
- **Baseline**: the run's entire output **is** exactly
  `  Pre-commit hook installed (.git/hooks/pre-commit)\n`, on the single writer that exists at
  `294b4b6ab`; there is no warning writer, so clause (i) is red at baseline.
- **Mutant**: appends a blank line or trailing whitespace to the success line. Under the original
  single-writer wording a pure inequality check passed while the user learned nothing; under the
  two-writer wording the padded run leaves the warning writer empty, so clause (i) now catches it
  directly. Clause (ii) additionally defeats a mutant that writes a blank or generic line to the
  warning writer — non-empty output that names nothing is not disclosure.
- **Failing input**: the whitespace-padded success line, with the warning writer empty. Must be
  observed red against clause (i).

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

### AC-PCP-010 — a failed supporting write is non-fatal, and never costs the user bytes (REQ-PCP-010)

Two sub-cases, both mandatory. They differ in post-state because the two supporting writes sit on
opposite sides of the hook write.

**(a) Backup write fails.** **Given** a user-modified hook and a hooks directory in which the backup
write cannot succeed, **When** `installPreCommitHookOptional` runs, **Then** it returns normally,
emits a warning naming the failure, does not panic or abort the caller, **and the installed hook's
bytes are byte-identical to their pre-run value** — no replacement occurred.

**(b) Provenance write fails after a successful hook write.** **Given** a run in which the hook write
succeeds but `.git/hooks/.moai-pre-commit.sha256` cannot be written, **When**
`installPreCommitHookOptional` runs, **Then** it returns normally, emits a warning, does not abort
the caller, **and the hook IS replaced** — and a second run with the same content takes no backup
and emits no notice (the record self-heals via REQ-PCP-005).

- **Decides**: `go test ./internal/cli/ -run TestPreCommitSupportWriteFailureNonFatal -count=1`
  (both sub-cases)
- **Baseline**: neither a backup nor a provenance write is attempted at `294b4b6ab`, so no failure
  mode exists to be non-fatal, and no post-state distinction exists to assert.
- **Mutant**: **backup fails → warn → overwrite anyway.** This is the destroying implementation: it
  returns normally, emits a warning, does not abort the caller — satisfying every clause of the
  criterion as originally written — while destroying the user's patch with no recoverable backup,
  violating REQ-PCP-006. It is defeated **only** by sub-case (a)'s post-state assertion; no other
  criterion in this set reaches it, because AC-PCP-006's Given is the AC-PCP-003 scenario, in which
  the backup succeeds. A second mutant: never attempts a backup at all — cannot fail, so passes;
  violates REQ-PCP-003 and is defeated by AC-PCP-003. A third: swallows the failure with no warning,
  defeated by the "emits a warning" clause.
- **Failing input**: for sub-case (a), the warn-then-overwrite-anyway implementation — must be
  observed red against the post-state clause specifically, not merely against "returns normally".
  Also red-check an implementation that returns the backup error to the caller and aborts
  `moai update`. For sub-case (b), an implementation that skips the hook write when the provenance
  write is going to fail (over-applying (a)'s precedence to the wrong side of the write).

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
- **Note (v0.4.0)**: this SPEC changes neither the constant nor the template (spec.md §C.1), so this
  criterion has nothing of its own to prove — it guards the invariant while the classifier lands.
  Paired with AC-PCP-005 sub-case (c), which fails when the body changes at all, the two together
  are what make "the hook body is untouched" a checked property rather than an intention.

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
| Backup write fails on a user-modified hook | REQ-PCP-010 sub-case (a): warn, **do not replace**, do not fail the caller. AC-PCP-010(a) asserts the unchanged post-state |
| Backup succeeds, then the hook write fails | no replacement occurred, so **no replacement notice** — the REQ-PCP-004 notice states the hook "was replaced", which would be false. Report the write failure as a warning (REQ-PCP-010's non-fatal path) and name the orphan `pre-commit.bak.<ts>` left beside the unchanged hook, so the artifact is explained rather than mysterious |
| Provenance write fails after a successful hook write | REQ-PCP-010 sub-case (b): warn, keep the replacement. Self-heals on the next run via the REQ-PCP-005 legacy path |

## §D.2 Quality gates

- `go vet ./internal/cli/...` clean
- `go test ./internal/cli/... -count=1` passing
- `go test ./internal/template/... -count=1` passing
- `make build` is **not** expected: this SPEC changes neither `preCommitHookContent` nor the
  template twin (spec.md §C.1). A run-phase in which `make build` becomes necessary has left scope —
  stop and re-open the sidecar decision (spec.md §A.5 Decision 1) rather than absorbing the edit

## §D.3 Definition of Done

- All 12 AC pass; the 2 that show no change from baseline (AC-PCP-008, -013) are
  behaviour-preserving and are falsifiable only against a named implementation mutant — as is
  AC-PCP-002, whose Given is unconstructible on the untouched tree. The other 9 fail against the
  untouched tree. (At v0.3.0 this read 15 AC with 5 mutant-only; AC-PCP-011, -012 and -015 left with
  the extension point — spec.md §D Out of Scope.)
- **Every criterion has been observed failing against its stated failing input** before being
  accepted. A criterion that has only ever been green is not evidence.
- No prompt is introduced on any path (§C.3 of spec.md).
- `preCommitHookContent` and `internal/template/templates/.git_hooks/pre-commit` are byte-identical
  to their pre-implementation state — this SPEC's diff touches installer logic and tests only
  (spec.md §C.1). AC-PCP-005 sub-case (c) and AC-PCP-013 are the two checks that observe it.
- `~/go/bin/moai spec lint .moai/specs/SPEC-PRECOMMIT-PRESERVE-001/spec.md` clean.
