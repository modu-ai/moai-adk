# Progress — SPEC-WORKTREE-SQUASH-MERGE-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-02
plan_revision: 4          # iteration-4 revision after plan-audit FAIL (0.79), user-override scoped to N5
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
requirements: 14   # REQ-WSM-001 .. REQ-WSM-014        (Tier M cap 16)
acceptance_criteria: 16   # AC-WSM-001 .. AC-WSM-016   (Tier M cap 16)
open_clarifications: 0
```

Scope decisions resolved during authoring, each from executed git experiments rather than
reasoning: the false-positive boundary (`spec.md` §5 decision 1), synthetic-object handling
(decision 2), the `--merged-only` co-change (decision 3), and the preserved safety
constraints (decision 4). Four falsified alternatives are recorded in `spec.md` §2 so they
are not re-proposed: plain per-commit `git cherry` as the squash detector; history-only
patch-id equivalence with no state check; enumerating S5's paths with rename detection on;
and splitting the enumeration's output on newline rather than NUL.

### Iteration-2 revision (plan-audit D1-D5 + D6-D8)

- **D2** — REQ-WSM-007 narrowed with a measured per-probe verdict-code table; the earlier
  wording made S3/S4 unreachable because `git diff --quiet` uses rc 1 as a verdict.
- **D1** — added the S5 state conjunct (REQ-WSM-014): S3/S4 are no longer sufficient alone.
  Matrix re-run at 9 scenarios × 5 signals; SC-2 (rebase) verified not to regress.
- **D3** — AC-WSM-009's judge repaired (`-r` dropped, exclusion anchored on the literal
  argument pair) and its falsification restated using `git prune`, then executed.
- **D4** — the clean-worktree-with-unpushed-commits removal case recorded in §5 decision 3,
  accepted with the branch-retention mitigation, and listed in §4 as out of scope.
- **D5** — AC-WSM-015 (REQ-001 + REQ-014) and AC-WSM-016 (REQ-011 + REQ-012) added;
  every requirement now has a binding criterion beyond the AC-WSM-014 catch-all.
- **D6** — AC-WSM-012 rewritten to match on the `false` argument and pin both call sites.
- **D7** — the unrelated-histories case folded into AC-WSM-007 as a second row.
- **D8** — §2's no-false-positive claim explicitly scoped to the constructed matrix.

Two residual risks the audit flagged are addressed rather than carried: `plan.md` §F's
`git cherry` cost claim is corrected with measurement (the bound holds for neither S3 nor
S4), and AC-WSM-008 now records the inputs that produce its hash.

### Iteration-3 revision (plan-audit N1-N4)

- **N1** — the rename-fold false positive closed: S5's enumeration takes `--no-renames`.
  Matrix extended to 11 scenarios with SC-10 and SC-11; SC-1..SC-9 unchanged.
- **N2** — non-vacuity guards added to AC-WSM-001, AC-WSM-010, AC-WSM-015.
- **N3** — SC-9's construction pinned to one recipe, with the SHA-collision degeneration
  recorded so the cherry-pick phrasing is not substituted.
- **N4** — the S5 shell snippet rewritten as an array; `plan.md` §D gained the
  `[]string`-not-a-joined-string subsection.

### Iteration-4 revision (plan-audit N5-N6, user override)

Scoped by explicit user override of the three-iteration ceiling to **one finding**. No
resolved finding was re-opened and no structural change was made beyond the two below.

- **N5** — the path-encoding false positive closed: S5's enumeration takes `-z` and its
  output is split on NUL. `git diff --name-only` C-quotes any path containing a backslash,
  tab, double quote, or control character — and, under the documented default
  `core.quotePath=true`, any non-ASCII path — and the quoted rendering matches nothing,
  which by §2 finding 5 reads as merged. Matrix extended to 12 scenarios with SC-12;
  SC-1..SC-11 re-executed and cell-for-cell unchanged. §C.1's mutation grid re-run at
  6 × 12; the new `newline-split` mutation flips exactly SC-12, disjoint from
  `folded-names`, which is the evidence that `--no-renames` and `-z` are separate guards.
  AC-WSM-006 was extended to carry both falsifications rather than a seventeenth criterion
  being added, keeping the SPEC at the Tier M ceiling of 16.
  The load-bearing part of this repair is a **correction**, not an addition: `plan.md` §D
  previously stated that a path list built with `strings.Split(out, "\n")` is *safe*. That
  claim was false in the unsafe direction — it endorsed the exact construction that
  produces the SC-12 false positive — and leaving it in place was the stated reason the
  override was authorized.
- **N6** — run-phase judging commands added to AC-WSM-002 through AC-WSM-008, so §F's
  "the judging command and its observed output recorded" requirement now holds for every
  criterion as written rather than for a subset. All twelve `-run` selectors were executed
  against the pre-implementation tree and their baseline `--- PASS:` counts recorded in §A
  (nine at 0, and 3 / 1 / 2 for the three pre-existing tests).

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
