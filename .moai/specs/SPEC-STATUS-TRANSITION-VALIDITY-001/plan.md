# Implementation Plan — SPEC-STATUS-TRANSITION-VALIDITY-001

Card: **t376**. Tier M. Tree of record: `3f03d9c36` (`WT-status-transition-gap`).

## §A — Context

`internal/spec` already reads status transitions from git (`lookupOwnershipTransitionFromGit`,
`lint_ownership.go:194-246`) and already carries the `(prev, curr, subject, sha, trailer)` tuple in
`ownershipTransitionRecord` (`:180-186`). What is missing is a consumer that judges the `(prev, curr)`
pair on its own terms. This card adds that consumer; it does not rebuild the extraction.

## §B — Known issues this plan must not trip over

- `OwnershipTransitionRule` returns nil in two places before it can judge anything —
  `expected == ownerNone` (`:404-407`) and `rec.AuthoredByAgent == ""` (`:412-415`). The new rule
  must not inherit either guard; that inheritance is the defect.
- `applyEraDemotion` marks **every** warning of a demoted document advisory (`lint.go:296-311`), so a
  warning-severity finding on a `completed` SPEC will not gate under `--strict`. This is accepted
  scope (spec.md §C), not something to route around inside the new rule.
- Module convention (`internal/spec/CLAUDE.md`): a new rule must not false-flag closed sibling SPECs,
  and rules are observation-only — no writes, no `os/exec` beyond the existing cached git query path.
- Git queries inside `Lint()` go through the per-run cache (`gitquery_cache.go`). Adding a second
  independent `git log` walk per document would double the run cost; reusing the existing lookup is
  the intended shape.

## §C — Pre-flight

1. Re-read `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum and
   § Status Transition Ownership Matrix. The transition set is derived from there — except for
   `draft → implemented`, which spec.md §A.5 D1 adopts from the Go arm against the SSOT's silence.
   Do not write a second definition of the DAG in Go prose.
2. Re-run the baseline before touching code, in this tree, and keep the output:
   `moai spec lint --json > <evidence>/lint-baseline-preflight.json`. The §A.1 table in spec.md is
   the comparison target for AC-STV-013.
3. Read `.moai/reports/t376/probe-transition-gap.log` and
   `.moai/reports/t376/transition-census.log` — the probe is the ground truth the regression test set
   replaces; the census is the population every count in spec.md §A.4-§A.6 rests on.

## §D — Constraints

- No implementation outside `internal/spec` (plus the one message-assertion test wherever it lives).
- Emit at warning severity with no emission-site `Advisory` flag (REQ-STV-009, spec.md §A.5 D2). Do
  not add either new code to `eraDemotableCodes` — that map is consulted only for `SeverityError`
  findings, so an entry there would be inert and read as intent.
- Do not touch `terminalStatusEnum`, the `StatusGitConsistencyRule` early return, or the demotion
  decision at `lint.go:239`.
- Follow the table-driven test convention (`[]struct{name, ...}` + `t.Run`), so each case is
  independently runnable.

## §E — Milestones

Ordered by decision-reversibility: the decisions most likely to change come first.

### M1 — The transition set, the check order, and the two finding shapes

All five decisions are settled (spec.md §A.5); no clarification remains open. What M1 still owns is
the shape they are written in.

- Represent the canonical set of spec.md §A.7 as one explicit structure with the SSOT cited beside
  it. Terminal targets (`superseded` / `archived` / `rejected`) are `* → X`, so they are a
  target-side test rather than a pair enumeration.
- Two codes, not one: `StatusTransitionInvalid` (the pair is outside the set) and
  `StatusTokenUnrecognized` (a side names a token outside the 8-value enum). Each message names what
  it actually observed.
- **Check order is load-bearing and must be decided explicitly**, because it changes which code a
  document gets: `(none)` skip (D5) → token recognition (D3) → pair validity (D1/D2/D4). spec.md §A.6
  notes one census row — the quote-wrapped `(none) → "in-progress"` — whose reported code depends
  entirely on this order. Write the order down and assert it in a test rather than letting it fall
  out of statement sequence.
- Findings name previous status, current status, and the performing commit SHA (AC-STV-001/002).
- Severity warning, **no emission-site `Advisory`** (D2, AC-STV-018).

### M2 — Wire the rules and measure the corpus

- Register both rules and run them over the live corpus.
- Report the per-code count against the §A.1 table (AC-STV-013), and compare the two new counts with
  the §A.6 hand projections (~98 / ~7). The projections are a comparison target, not a pass
  condition — a wide miss is a finding to explain, since it means the census population and the
  rule's own reading of it diverge.
- Run the non-overlap comparison against `StatusValueInvalid` (AC-STV-016) here, on real output. A
  non-empty intersection fails that AC and is not dischargeable by reporting it.
- Split the post-change findings by `advisory` and report the non-advisory count (AC-STV-019). A
  non-zero count needs a recorded decision in `progress.md` §E.2 before the card closes; the strict
  gate's current headroom is zero.
- Any movement in a code this card did not touch is explained, not absorbed.

### M3 — Regression test set (replaces both scratch probes)

- Convert the probe's 8 cases into a proper table-driven test covering AC-STV-001..AC-STV-009 and
  AC-STV-014/015/017/018, in one execution, with the AC-STV-010 live control asserted in that same
  run. Use the census's real tokens (`synced`, `approved`, `cancelled`, `Completed`) as the
  unrecognized-token fixtures rather than invented ones.
- Include the no-trailer case explicitly (AC-STV-003) — it is the property that separates this rule
  from the existing one.
- Include a closed-sibling case per the module convention, so the rule is shown not to false-flag a
  SPEC closed before it shipped.
- Delete `internal/spec/zz_t376_probe_test.go` and `internal/spec/zz_t376_census_test.go` in this
  milestone, not before: both are the reference the test set is written against. Their logs under
  `.moai/reports/t376/` stay.

### M4 — Demotion message cause (mechanical)

- `applyEraDemotion` currently receives one boolean and appends one fixed string. Give it the cause
  and let the annotation name it (AC-STV-011).
- Update every existing test asserting the old literal. Search before editing:
  `grep -rn "grandfathered era — downgraded" internal/`.

### M5 — Close-out (mechanical)

- `go test ./internal/spec/...` on the affected package; cite the output. Full-suite verdict is CI's.
- Record the AC-STV-013 per-code comparison in `progress.md` §E.2, and the AC PASS/FAIL matrix in
  §E.3.

## §F — Risks

| Risk | Consequence | Mitigation |
|---|---|---|
| A census edge decided wrong (D1/D4/D5) | A corpus-wide false-positive class — 217, 136, or 50 findings | AC-STV-007a and AC-STV-017 assert the three largest silently; M2 measures against the §A.6 projection |
| New rule duplicates the git walk | Lint run cost doubles | Reuse `lookupOwnershipTransitionFromGit` / the per-run cache |
| Non-advisory severity reddens the integration branch | `spec-lint --strict` runs on every `main` / `develop` push and currently has **zero** non-advisory findings, so the first one this rule emits turns it red | **AC-STV-019** binds it: M2 reports the advisory split, and a non-zero non-advisory count requires a recorded decision before close. This mitigation is an AC, not plan prose — the audit's D2 finding was that a risk mitigated only in prose has no gate |
| Check order left implicit | A document gets the wrong code, and the message misstates the cause | M1 writes the order down and asserts it |
| Message fix breaks an assertion elsewhere | Red test in an unrelated package | M4 greps before editing |

## §G — Anti-patterns

- Re-deriving the lifecycle DAG in Go from memory instead of from the SSOT.
- Judging the transition only when a trailer is present — that is the defect, restated.
- Reporting the post-change baseline as a single aggregate delta.
- Reading an all-silent test run as a pass without a live control.

## §H — Cross-references

- `spec.md` §A.3 (the four measured layers), §A.4 (the derived edge set), §C (non-goals)
- `acceptance.md` (the binding AC set)
- `internal/spec/CLAUDE.md` (module conventions)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` (DAG SSOT)
