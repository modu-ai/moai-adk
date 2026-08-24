# SPEC-PRECOMMIT-PRESERVE-001 — Implementation Plan

Milestones are ordered by decision-reversibility: the attribution design (M1) is the decision most
likely to change under review, so it leads. Mechanical work sits at the bottom.

## §A Context

`internal/cli/hook_install_precommit.go` (179 lines at `294b4b6ab`) is the whole surface. Two
callers: `internal/cli/update_template_sync.go:575` and `internal/cli/init.go:898`, both gated by
`--no-hooks`. Neither caller changes in this SPEC.

## §B Known Issues carried in

- The card's line numbers (`:557`, `:773`) are stale; the correct call sites are recorded in
  spec.md §A.2. Do not re-derive from the card.
- The card's "distributed original 6,454 B" is a downstream-checkout figure; this tree's template is
  3,245 B. Neither number is load-bearing for the fix.

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

## §F Milestones

### M1 — attribution: the provenance sidecar (highest change likelihood)

The decision under review. Delivers REQ-PCP-001, REQ-PCP-002, REQ-PCP-005.

- Write `.git/hooks/.moai-pre-commit.sha256` after every successful hook write.
- On install with a marker-bearing hook present, classify the file as `unmodified` /
  `user-modified` / `undecidable-legacy` per the spec.md §A.5 table.
- No backup and no notice yet — this milestone only decides, so the classifier can be reviewed and
  argued with before any behaviour hangs off it.
- Tests: AC-PCP-001, AC-PCP-002, AC-PCP-005.

Reviewer's question for this milestone: does a routine version bump land in the silent row? If it
does not, everything downstream is noise.

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

- **Byte comparison as attribution.** Comparing the installed hook against the content about to be
  written and calling any difference a user patch. This is the defect spec.md §A.3 exists to
  prevent; it produces a warning on every version bump.
- **Preserve-instead-of-overwrite.** Rejected in spec.md §A.5 Decision 2. Silently freezing a
  project on an old hook is the same failure in the other direction.
- **Interactive confirmation.** Hangs `moai update` in CI.
- **One-sided template edit.** Editing the constant without the template (or vice versa) fails
  `TestPreCommitTemplateMatchesConstant` — but only if the test runs; do not skip `internal/cli`.
- **Widening to pre-push.** Out of scope; keeps the diff attributable.

## §H Cross-References

- `internal/cli/hook_install_precommit.go` — the whole implementation surface
- `internal/cli/hook_install_precommit_test.go:38` — `TestPreCommitTemplateMatchesConstant`
- `internal/template/templates/.git_hooks/pre-commit` — the template twin
- Card t237 / issue #1641 — collides on M3's file pair
- Card t235 / issue #1639 — the gate-serialization axis, out of scope here
