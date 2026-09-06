# SPEC-ERRCHECK-CATALOG-HASH-001 — Acceptance Criteria

## §A Verification Command (single judge, all ACs)

```bash
golangci-lint run ./internal/template/... --timeout=2m
```

Every AC below is judged on the **errcheck match count** in this command's output, never on the
exit code (the exit code conflates errcheck with any other linter signal; the measured defect set
is exactly one errcheck finding).

## §D AC Matrix

### AC-GREEN — Fixed tree reports errcheck count 0

**Given** the worktree at branch `WT-errcheck-catalog-hash` with the fix applied at
`internal/template/catalog_tree_hash.go:60`
**When** the verification command runs
**Then** the output contains zero errcheck matches at BOTH scopes: (a) file-scoped — the count of
lines matching `catalog_tree_hash.go.*errcheck`-class findings is 0, AND (b) package-scoped — the
summary line reports `* errcheck: 0` (binding the package-wide REQ-001, not merely the flagged
file).

### AC-MUTANT — Reverted (pre-fix) tree reports errcheck count 1 — anti-vacuous guard

**Given** the fix is reverted by restoring the pre-fix content of
`internal/template/catalog_tree_hash.go` — either
`git checkout 615d18c1f -- internal/template/catalog_tree_hash.go` (tree-ish source) or
`git show 735b7d7b4856088dc7702a56fd450c0800327666 > internal/template/catalog_tree_hash.go`
(recorded pre-fix blob, whose line 60 is the unchecked call; verified in this run)
**When** the same verification command runs
**Then** the output reports exactly 1 errcheck finding at `catalog_tree_hash.go:60:14` — proving
the check actually inspects this file under the committed config. **After** observing count 1, the
fix is re-applied and AC-GREEN re-confirmed before closure.

### AC-ORDERING — Baseline commit precedes fix commit

**Given** the RED baseline artifact `.moai/reports/t489/red-baseline.md` (measured by lane-15 in
this run)
**When** `git log` on the card branch is examined
**Then** the commit introducing `.moai/reports/t489/red-baseline.md` is an ancestor of (precedes)
the commit modifying `internal/template/catalog_tree_hash.go`, so the commit graph — not the
commit message — witnesses the baseline-first ordering (verification-claim-integrity §2.3).

### AC-SCOPE — Verification scope is the linter only, and the diff shape is the sanctioned form

**Given** the run-phase closure report
**When** its verification section and the branch diff are examined
**Then** (a) the only verification command cited is the §A linter command (no full test-suite run
is claimed or required), (b) the diff contains no new test file and no `.golangci.yml` change, AND
(c) the diff shape discriminates the sanctioned repair from a prohibited alternative:

- the branch diff at `internal/template/catalog_tree_hash.go` shows the bare call
  `fmt.Fprintf(h, "%s:%s\n", e.rel, e.sum)` prefixed with `_ =` plus an English comment citing the
  `hash.Hash` never-error contract (satisfying REQ-002), and
- the closure check `git diff 615d18c1f..HEAD -- '*.go'` contains NO `//nolint` directive —
  scoped to Go sources deliberately, because the SPEC prose itself (spec.md / plan.md /
  acceptance.md, committed by AC-ORDERING/AC-EVIDENCE) carries the literal string `//nolint` and a
  whole-branch grep would false-positive on a correct implementation. A `//nolint:errcheck`
  comment at line 60 of the Go file would otherwise pass
  AC-GREEN/AC-MUTANT/AC-ORDERING/AC-EVIDENCE while violating REQ-002, REQ-003, and the SPEC's own
  Out of Scope — this clause closes that hole.

Rationale: the fix is a `_ =` discard of a structurally-always-nil error and cannot alter program
semantics, so the existing package tests remain the behavior evidence and no new test is required.

### AC-EVIDENCE — Evidence exported and traceable

**Given** the completed run phase
**When** the repository is examined
**Then** (a) evidence artifacts exist under `.moai/reports/t489/` and are committed, and (b) every
commit on the card branch carries the card id `t489` in its message.

## §D.1 Severity

| AC | Severity | Gate |
|----|----------|------|
| AC-GREEN | Must-pass | blocks closure |
| AC-MUTANT | Must-pass | blocks closure (anti-vacuous guard) |
| AC-ORDERING | Must-pass | blocks closure |
| AC-SCOPE | Must-pass | blocks closure |
| AC-EVIDENCE | Must-pass | blocks closure |

## §D.5 Closure Gates

- All five ACs PASS with verbatim command output cited (no summaries).
- AC-MUTANT's revert cycle is fully closed: fix re-applied and AC-GREEN re-observed after the
  mutant measurement.
- Working tree clean of stray modifications beyond the single flagged line and the evidence files.

## Definition of Done

REQ-001 through REQ-006 satisfied, evidenced by AC-GREEN, AC-MUTANT, AC-ORDERING, AC-SCOPE, and
AC-EVIDENCE, with the five-section evidence-bearing report (Claim / Evidence / Baseline-attribution
/ Gaps / Residual-risk) in the run-phase completion report.
