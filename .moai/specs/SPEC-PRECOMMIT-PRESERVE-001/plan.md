# SPEC-PRECOMMIT-PRESERVE-001 — Implementation Plan

Milestones are ordered by decision-reversibility: the attribution design (M1) is the decision most
likely to change under review, so it leads. Mechanical work sits at the bottom.

## §A Context

`internal/cli/hook_install_precommit.go` (179 lines `@294b4b6ab`) is the whole surface. Two callers,
both invoking `installPreCommitHookOptional` and both gated by `--no-hooks`:
`internal/cli/update_template_sync.go:575 @294b4b6ab` and `internal/cli/init.go:898 @294b4b6ab`.
Neither caller changes in this SPEC.

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
4. Check whether card t237 has landed on `origin/main`; if it has, rebase before starting M3.

## §D Constraints

- Non-interactive throughout (§C.3 of spec.md).
- M1/M2 must not touch `preCommitHookContent` or its template twin. If a change there appears
  necessary during M1 or M2, the sidecar decision (spec.md §A.5, Decision 1) has been abandoned —
  stop and re-open that decision rather than absorbing the edit quietly.
- `make build` after M3 only.

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
REQ-PCP-008.

- On `user-modified`, copy to `.git/hooks/pre-commit.bak.<UTC RFC3339-ish timestamp>` before
  writing; pick a distinct path if occupied.
- Emit the stderr notice from `installPreCommitHookOptional`, naming the backup path, the
  replacement, and `pre-commit.local`.
- Backup and sidecar write failures degrade to a warning; the caller is never failed.
- Tests: AC-PCP-003, AC-PCP-004, AC-PCP-006, AC-PCP-007, AC-PCP-008, AC-PCP-009, AC-PCP-010.

### M3 — the local extension point (paired edit; collision-prone — land last)

Delivers REQ-PCP-011, REQ-PCP-012, REQ-PCP-013.

- Append the `pre-commit.local` delegation as the hook's final step, before `exit 0`, propagating
  its exit status.
- Edit **both** `preCommitHookContent` and
  `internal/template/templates/.git_hooks/pre-commit`, then `make build`.
- Land as its own commit, last (§C.2 — card t237 edits the same pair).
- Tests: AC-PCP-011, AC-PCP-012, AC-PCP-013.

Note the ordering consequence: M2's notice points users at `pre-commit.local`, which does not exist
until M3. Either land M3 before the SPEC closes, or the notice names a facility the user cannot yet
use — an acceptable intermediate state within one SPEC, not an acceptable shipping state.

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
- **Widening to pre-push.** Out of scope; keeps the diff attributable.

## §H Cross-References

- `internal/cli/hook_install_precommit.go` — the whole implementation surface
- `TestPreCommitTemplateMatchesConstant` — `internal/cli/hook_install_precommit_test.go:38
  @294b4b6ab`; note its `t.Skipf` branch at `:45 @294b4b6ab` (AC-PCP-013's mutant)
- `internal/template/templates/.git_hooks/pre-commit` — the template twin
- Card t237 / issue #1641 — collides on M3's file pair
- Card t235 / issue #1639 — the gate-serialization axis, out of scope here
