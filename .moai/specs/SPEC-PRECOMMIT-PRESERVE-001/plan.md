# SPEC-PRECOMMIT-PRESERVE-001 — Implementation Plan

Milestones are ordered by decision-reversibility: the attribution design (M1) is the decision most
likely to change under review, so it leads. Mechanical work sits at the bottom.

**Scope at v0.4.0: M1 and M2 only.** A third milestone — the `pre-commit.local` extension point —
was planned here and has moved to a successor card (spec.md §D Out of Scope). It was the only
milestone that changed the distributed hook body, and every blocking defect of plan-audit
iteration 2 followed from that one fact. What remains touches installer logic and tests, and
nothing else.

## §A Context

`internal/cli/hook_install_precommit.go` (179 lines `@294b4b6ab`) is the implementation surface. Two
callers, both invoking `installPreCommitHookOptional` and both gated by `--no-hooks`:
`internal/cli/update_template_sync.go:575 @294b4b6ab` and `internal/cli/init.go:898 @294b4b6ab`.

**Each caller changes by exactly one line, and no more.** Version 0.2.0 of this plan said "neither
caller changes", which made REQ-PCP-004 ("on stderr") unsatisfiable: the installer has a single
writer, and the `moai update` caller binds it to **stdout** (`out := cmd.OutOrStdout()` at
`update_template_sync.go:69 @7b2f42be0`, inside `runTemplateSyncWithReporter`; a second,
unrelated `out := cmd.OutOrStdout()` at `:604` belongs to a later function and is not this one), while `moai init` binds it to stderr
(`init.go:898 @7b2f42be0`). Leaving that contradiction for run-phase would have driven the cheapest
repair — weakening AC-PCP-004 to capture `os.Stderr` — which is exactly the criterion erosion this
SPEC exists to prevent.

The permitted change is: add a warning-writer parameter to `installPreCommitHookOptional`, and pass
the stderr writer each caller already holds — `errOut` (`update_template_sync.go:72 @7b2f42be0`) and
`cmd.ErrOrStderr()` (`init.go:898 @7b2f42be0`). Nothing else in either caller moves: not the
`--no-hooks` gate, not the ordering against `installPrePushHookOptional`, not the existing progress
writer. spec.md §C.4 carries the same permission; anything wider is out of scope.

## §B Known Issues carried in

- **Line numbers are per-tree, not absolute.** Card t230 cites `init.go:773`, correct on the primary
  checkout `@a1b1ca696`; this worktree `@294b4b6ab` puts the same call at `:898`. Both are right.
  Version 0.1.0 of this SPEC wrongly called the card stale — see spec.md §A.0 and §A.2. Resolve by
  symbol (`installPreCommitHookOptional`), never by line number alone, and cite the SHA when
  recording one.
- The card's "distributed original 6,454 B" is a figure from the downstream `mo.ai.kr` checkout;
  this tree's template is 3,245 B `@294b4b6ab`. Different trees, not a contradiction. Neither number
  is load-bearing for the fix.

## §C Pre-flight

1. `git rev-parse --short HEAD` — confirm the tree the work starts from.
2. `go test ./internal/cli/ -run TestPreCommit -count=1` — record the passing baseline before any
   edit, so a later failure is attributable.
3. `grep -n 'installPreCommitHookOptional' internal/cli/update_template_sync.go internal/cli/init.go`
   — re-confirm the call sites have not drifted again.
4. `git show v3.1.2:internal/template/templates/.git_hooks/pre-commit | cmp - internal/template/templates/.git_hooks/pre-commit`
   — record the AC-PCP-005 sub-case (c) fixture baseline (identical, rc 0) before any edit, so a
   later divergence is attributable rather than assumed. This is also the fixture the sub-case reads
   from git at test time; confirming it resolves here catches a shallow clone or a missing tag
   before it presents as a test failure.
   (Card t237 edits the constant/template pair this SPEC does not touch, so no **rebase** check
   against it is needed — spec.md §C.2. Its **release-order** constraint is a different question and
   is live: spec.md §A.5 Decision 3, checked at sync-phase against the integration ref per
   acceptance.md §D.3. If this baseline is anything but rc 0, a body change has already reached the
   tree the work starts from — stop and read Decision 3 before implementing.)

## §D Constraints

- Non-interactive throughout (§C.3 of spec.md).
- M1/M2 must not touch `preCommitHookContent` or its template twin. If a change there appears
  necessary during M1 or M2, the sidecar decision (spec.md §A.5, Decision 1) has been abandoned —
  stop and re-open that decision rather than absorbing the edit quietly.
- `make build` is not expected at any point: no template or embedded content changes.
- Caller edits are limited to the single-line warning-writer argument at each of the two call sites
  (§A). A caller change beyond that is a scope violation, not an implementation detail.
- The hook body stays byte-identical throughout. AC-PCP-005 sub-case (c) and AC-PCP-013 observe it;
  if either goes red, an edit has reached the constant or the template and the run has left scope.
- **Release composition is a constraint on this SPEC's release, not on this SPEC's diff** (spec.md
  §A.5 Decision 3). Nothing in M1 or M2 can violate it and nothing here can check it: the run-phase
  gates below read the working tree, where the constraint is green by construction. It is checked at
  sync-phase against the integration ref (acceptance.md §D.3) and again, outside this SPEC, at the
  release-candidate cut (acceptance.md §D.4). Recorded here so a run-phase reader does not mistake
  the M2 scope gate for the composition check.

## §E Self-Verification

After each milestone: `go vet ./internal/cli/...` and `go test ./internal/cli/... -count=1`. Cite
the verbatim output in `progress.md §E.2`, not a summary.

Additionally, per criterion delivered: **observe it failing against its stated failing input before
accepting it green.** acceptance.md names the input for every criterion. A green-only check has not
been shown to observe anything; record the red run's output alongside the green one.

## §F Milestones

### M1 — attribution: the three-way classifier (highest change likelihood)

The decision under review. Delivers REQ-PCP-014, REQ-PCP-001, REQ-PCP-002, REQ-PCP-005.

- Write `.git/hooks/.moai-pre-commit.sha256` after every successful hook write.
- On install with a marker-bearing hook present, classify the file as `unmodified` /
  `user-modified` / `undecidable-legacy` per the spec.md §A.5 table.
- No backup and no notice yet — this milestone only decides, so the classifier can be reviewed and
  argued with before any behaviour hangs off it.
- Tests: AC-PCP-014, AC-PCP-001, AC-PCP-002, AC-PCP-005.

Reviewer's question for this milestone: does a routine version bump land in the silent row? If it
does not, everything downstream is noise. AC-PCP-014 case one is the check that answers it, and the
two-way implementation must be observed failing it before the milestone is accepted.

### M2 — disclosure: backup and notice

Delivers REQ-PCP-003, REQ-PCP-004, REQ-PCP-006, REQ-PCP-007, REQ-PCP-009, REQ-PCP-010, and confirms
REQ-PCP-008 and REQ-PCP-013. (REQ-PCP-013 was M3's at v0.3.0; with M3 gone it is confirmed here,
where its guard is what proves the hook body went untouched.)

- On `user-modified`, copy to `.git/hooks/pre-commit.bak.<UTC RFC3339-ish timestamp>` before
  writing; pick a distinct path if occupied.
- Emit the notice from `installPreCommitHookOptional` on its new warning writer (stderr at both call
  sites, §A), naming the backup path and the replacement — **and nothing else**. It must not name
  `pre-commit.local`; that facility left this SPEC at v0.4.0 (REQ-PCP-004).
- **Backup failure and sidecar failure differ in precedence** (REQ-PCP-010): a failed backup warns
  and **leaves the hook untouched**; a failed sidecar write, which happens after the hook write,
  warns and leaves the replacement in place. Neither fails the caller. Getting this backwards — warn
  and overwrite anyway on a failed backup — destroys the patch this SPEC exists to protect;
  AC-PCP-010(a)'s post-state assertion is the check.
- Tests: AC-PCP-003, AC-PCP-004, AC-PCP-006, AC-PCP-007, AC-PCP-008, AC-PCP-009, AC-PCP-010,
  AC-PCP-013.
- Scope gate at the end of M2: re-run
  `git show v3.1.2:internal/template/templates/.git_hooks/pre-commit | cmp - internal/template/templates/.git_hooks/pre-commit`
  and confirm rc 0. M2 is the last milestone, so this is the last chance to catch a hook-body edit
  that crept in — the property AC-PCP-005 sub-case (c) and AC-PCP-013 assert.

## §G Anti-Patterns

- **Two-way comparison as attribution.** Comparing the installed hook against the content about to
  be written and calling any difference a user patch. Forbidden by REQ-PCP-014; it produces a
  warning on every version bump. AC-PCP-014 exists to fail exactly this implementation.
- **Asserting file state instead of output.** Tests that check only the post-run hook and backup
  bytes admit the card's headline mutant — a correct, recoverable, entirely silent overwrite.
  AC-PCP-004/006/007 assert the notice text itself; do not weaken them to file-state checks.
- **Preserve-instead-of-overwrite.** Rejected in spec.md §A.5 Decision 2. Silently freezing a
  project on an old hook is the same failure in the other direction.
- **Interactive confirmation.** Hangs `moai update` in CI.
- **One-sided template edit.** Editing the constant without the template (or vice versa) fails
  `TestPreCommitTemplateMatchesConstant` — but only if the test runs; do not skip `internal/cli`.
- **Warn-and-overwrite-anyway on a failed backup.** Satisfies "warn, do not fail the caller" while
  destroying the user's patch with no recoverable backup. REQ-PCP-010 sub-case (a) forbids it and
  AC-PCP-010(a)'s post-state assertion is the only criterion that reaches it.
- **Routing the backup notice to the progress writer.** Under `moai update` that writer is stdout,
  so a scripted `moai update >/dev/null` swallows the data-loss notice entirely — the card's own
  failure mode, reintroduced through the stream choice. AC-PCP-004 clause (ii) checks the call
  sites, not just the installer.
- **Absorbing the extension point back into this SPEC.** Appending the `pre-commit.local`
  delegation "while we are in here" restores every defect the v0.4.0 trim removed, and makes
  `installed != incoming` true for 100% of users on the release that introduces the classifier. It
  belongs to the successor card (spec.md §D Out of Scope). AC-PCP-005 sub-case (c) and AC-PCP-013
  are what notice.
- **Naming `pre-commit.local` in the backup notice.** The facility does not exist yet; a notice that
  names it hands the user a recovery path that silently does nothing (REQ-PCP-004).
- **Widening to pre-push.** Out of scope; keeps the diff attributable.

## §H Cross-References

- `internal/cli/hook_install_precommit.go` — the whole implementation surface
- `TestPreCommitTemplateMatchesConstant` — `internal/cli/hook_install_precommit_test.go:38
  @294b4b6ab`; note its `t.Skipf` branch at `:45 @294b4b6ab` (AC-PCP-013's mutant)
- `internal/template/templates/.git_hooks/pre-commit` — the template twin
- Card t237 / issue #1641 — edits the constant/template pair; open, with a verified patch on
  `t312-precommit-vet @ b6f478b1a`. No merge collision with this SPEC (spec.md §C.2), but bound by
  the release-order constraint of spec.md §A.5 Decision 3
- Successor card — the `pre-commit.local` extension point, split out at v0.4.0 (spec.md §D Out of
  Scope); depends on this SPEC landing first so provenance records exist when the body changes
- Card t235 / issue #1639 — the gate-serialization axis, out of scope here
