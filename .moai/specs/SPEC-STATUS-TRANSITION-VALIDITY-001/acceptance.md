# Acceptance Criteria — SPEC-STATUS-TRANSITION-VALIDITY-001

Card: **t376**. Every AC below is binary-testable by executing a command and reading its output.

The set is deliberately **bidirectional**: the invalid-transition ACs and the legitimate-transition
ACs must be satisfied by the **same execution** of the regression test, so a run that catches
nothing cannot be read as clean when it is actually broken (AC-STV-010 is the live control).

## §D — AC Matrix

### AC-STV-001 — `draft → completed` is caught

*Covers REQ-STV-001, REQ-STV-003.*

**Given** a repository holding a SPEC whose git history records a `draft → completed` status
transition,
**When** the lint engine runs over that SPEC,
**Then** the findings include exactly one `StatusTransitionInvalid` entry for that SPEC whose message
names `draft`, `completed`, and the SHA of the commit that performed the transition.

### AC-STV-002 — `completed → draft` is caught

*Covers REQ-STV-001, REQ-STV-004.*

**Given** a repository holding a SPEC whose git history records a `completed → draft` status
transition,
**When** the lint engine runs over that SPEC,
**Then** the findings include exactly one `StatusTransitionInvalid` entry for that SPEC whose message
names `completed`, `draft`, and the performing commit SHA.

### AC-STV-003 — trailer-independence

*Covers REQ-STV-002.*

**Given** two repositories identical except that the transition commit carries an
`Authored-By-Agent:` trailer in one and no trailer at all in the other, each recording the same
invalid transition,
**When** the lint engine runs over both,
**Then** both emit `StatusTransitionInvalid` for that SPEC, and the two findings differ in no field
other than the commit SHA and the file path.

This is the AC that distinguishes the new rule from `OwnershipTransitionRule`, whose measured
behavior is a silent skip on the no-trailer case (`.moai/reports/t376/probe-transition-gap.log`,
row `implemented_to_completed_no_trailer`).

### AC-STV-004 — `implemented → completed` with the correct owner still passes

*Covers REQ-STV-005.*

**Given** a repository whose SPEC history records `implemented → completed` authored by
`manager-docs`,
**When** the lint engine runs,
**Then** no `StatusTransitionInvalid` finding is emitted for that SPEC.

### AC-STV-005 — `draft → in-progress` still passes

*Covers REQ-STV-005.*

**Given** a repository whose SPEC history records `draft → in-progress`,
**When** the lint engine runs,
**Then** no `StatusTransitionInvalid` finding is emitted for that SPEC.

### AC-STV-006 — `in-progress → implemented` still passes

*Covers REQ-STV-005.*

**Given** a repository whose SPEC history records `in-progress → implemented`,
**When** the lint engine runs,
**Then** no `StatusTransitionInvalid` finding is emitted for that SPEC.

### AC-STV-007 — the declared `completed → in-progress` amendment still passes

*Covers REQ-STV-005.*

**Given** a repository whose SPEC history records `completed → in-progress`,
**When** the lint engine runs,
**Then** no `StatusTransitionInvalid` finding is emitted for that SPEC — the edge is the declared
amendment transition of the SSOT matrix, not a reversal.

### AC-STV-007a — the single-sync-commit close and the adopted `draft → implemented` still pass

*Covers REQ-STV-005.*

**Given** two repositories, one whose SPEC history records `in-progress → completed` (the
single-sync-commit close of §A.5 D4, 217 census occurrences) and one recording `draft → implemented`
(§A.5 D1, 50 occurrences),
**When** the lint engine runs over each,
**Then** neither emits a `StatusTransitionInvalid` finding.

These two edges carry 267 of the 713 census records between them. A rule that flags either is not
shipping a detector, it is shipping a corpus-wide false positive.

### AC-STV-008 — `planned` edges are tolerated

*Covers REQ-STV-006.*

**Given** a repository whose SPEC history records an edge naming the legacy-optional `planned`
status on either side,
**When** the lint engine runs,
**Then** no `StatusTransitionInvalid` finding is emitted for that SPEC.

### AC-STV-009 — no transition, or no git, is silent

*Covers REQ-STV-007.*

**Given** a working directory that is not a git repository, and separately a repository whose SPEC
history records no status change at all,
**When** the lint engine runs in each,
**Then** neither run emits a `StatusTransitionInvalid` finding, and neither emits any finding at
error severity attributable to this rule.

### AC-STV-010 — live control in the same execution

*Covers REQ-STV-011.*

**Given** the regression test set of AC-STV-001..AC-STV-009 executed as one run,
**When** the run completes,
**Then** at least one case in that same run has asserted a finding that MUST fire (AC-STV-001 and
AC-STV-002 both qualify), and the test fails when that case produces no finding.

A run in which every assertion is "no finding" cannot distinguish a correct implementation from a
rule that was never wired in. Precedent: the probe's `implemented_to_completed_wrong_owner` row,
which is the one row proving the probe harness itself was live.

### AC-STV-011 — the demotion annotation names its actual cause

*Covers REQ-STV-008.*

**Given** two demoted documents — one demoted because its directory classifies as grandfathered era,
one demoted solely because its frontmatter status is terminal —
**When** the lint engine runs over both,
**Then** the demoted findings' appended annotations differ, each naming the cause that actually
fired, and no annotation on the terminal-status-only document claims grandfathered era.

Any existing test asserting the previous single string is updated in the same change.

### AC-STV-012 — observation-only

*Covers REQ-STV-012.*

**Given** the repository tree before a lint run,
**When** the lint engine runs over the full corpus,
**Then** `git status --porcelain` reports no change attributable to the run, and the diff of the
landed change touches none of `terminalStatusEnum`, `eraDemotableCodes`, the
`StatusGitConsistencyRule` early return, or the demotion decision at `lint.go:239` other than the
message text required by AC-STV-011.

### AC-STV-013 — per-code baseline re-measurement

*Covers REQ-STV-010.*

**Given** the §A.1 baseline of 1096 findings broken down per rule code (CoverageIncomplete 846,
MovingRefUnpinned 114, LegacyEARSKeyword 43, ModalityMalformed 25, MissingExclusions 24,
StatusGitConsistency 18, FrontmatterInvalid 14, InvalidREQID 6, SyncSHASlotFormat 5,
OwnershipTransitionInvalid 1),
**When** `moai spec lint --json` is re-run over the same corpus after the change and its findings
are grouped by code,
**Then** the run-phase evidence reports the new count for **every** code side by side with its
baseline count, and attributes each non-zero movement to that code individually.

Movement is expected: this card adds two codes, so `StatusTransitionInvalid` and
`StatusTokenUnrecognized` rows appear. spec.md §A.6 projects roughly 98 and 7 respectively **by hand
from the census table, not by observation** — those projections are the comparison target, not the
pass condition. A projection that matches is weak evidence; a projection that misses by a wide
margin is a finding to explain before the card closes. An aggregate "delta N" figure does not satisfy
this AC, and movement in a code this card did not touch is explained, not absorbed.

### AC-STV-014 — the `completed → implemented` reversal is caught

*Covers REQ-STV-001.*

**Given** a repository whose SPEC history records `completed → implemented` — the reversal class the
census counted 48 times —
**When** the lint engine runs,
**Then** a `StatusTransitionInvalid` finding is emitted for that SPEC.

### AC-STV-015 — an unrecognized token fires its own code

*Covers REQ-STV-013.*

**Given** a repository whose SPEC history records a transition naming a token outside the canonical
8-value enum (the census supplies real examples: `synced`, `approved`, `cancelled`, `Completed`,
`Superseded`),
**When** the lint engine runs,
**Then** a `StatusTokenUnrecognized` finding is emitted naming that token, and **no**
`StatusTransitionInvalid` finding is emitted for the same pair.

The second half of the assertion is the load-bearing one: it is what stops the new code from
becoming a second spelling of the first.

### AC-STV-016 — non-overlap with `StatusValueEnumRule`, measured

*Covers REQ-STV-015.*

**Given** the full corpus,
**When** `moai spec lint --json` runs after the change and the documents reported under
`StatusTokenUnrecognized` are intersected with those reported under `StatusValueInvalid`,
**Then** the intersection is **empty** — zero documents appear in both sets.

A non-empty intersection **fails this AC**. It may not be discharged by listing the overlap: an
overlapping document is the duplication REQ-STV-015 prohibits until someone shows otherwise, and the
card does not close on it. Two remedies are available, both of which must land before the AC passes:
narrow the new rule so the overlap disappears, or — where a specific overlap is genuinely
correct — enumerate the exact documents allowed to appear in both, record them in `progress.md` §E.2
with both messages and the reason each is not duplication, and amend this AC to name that bounded
set, so that any document outside it still fails.

The distinction being protected: `StatusValueInvalid` judges the frontmatter's current value,
`StatusTokenUnrecognized` judges a token seen in git history. They are **expected** to be disjoint on
this corpus; this AC is what turns that expectation into a measured outcome, and an AC that could be
satisfied by reporting the overlap would measure nothing at all.

### AC-STV-017 — `(none) → X` is skipped

*Covers REQ-STV-014.*

**Given** a repository whose SPEC history yields the `(none) → X` shape for several targets,
including `completed`,
**When** the lint engine runs,
**Then** no `StatusTransitionInvalid` and no `StatusTokenUnrecognized` finding is emitted for those
documents.

Census scale makes this a gate rather than a detail: 136 of 713 records carry the `(none)` shape,
104 of them targeting `completed`.

### AC-STV-018 — the finding actually gates

*Covers REQ-STV-009.*

**Given** a modern-era SPEC whose frontmatter status is not terminal and whose history records an
invalid transition,
**When** `moai spec lint --strict` runs over it,
**Then** the emitted `StatusTransitionInvalid` finding carries `advisory: false` in the JSON output
and `HasErrors()` reports true.

This is the AC that distinguishes a rule that reports from a rule that blocks. Its counterpart on a
`completed` SPEC — advisory, non-gating — is the accepted limitation recorded in spec.md §C, and is
not a failure of this AC.

### AC-STV-019 — the gating population is measured, and a non-zero one is decided on

*Covers REQ-STV-009.*

**Given** the post-change corpus lint output,
**When** the findings are split by their `advisory` field
(`jq '[.[] | select(.advisory != true)] | length'` and the same filtered to the two new codes),
**Then** the non-advisory count is reported — overall and for each of the two new codes — and, **when
that count is non-zero**, an explicit decision is recorded in `progress.md` §E.2 before the card may
close: either accept that `spec-lint --strict` now reddens on the integration branch, or hold the
finding advisory and record why.

The card does not close on an unreported gating population, and it does not close on a non-zero one
with no decision attached.

Why this AC exists as a gate rather than as plan prose: the measured headroom is **zero**. Of the
1096 baseline findings, the count carrying `advisory != true` is **0**
(`jq '[.[] | select(.advisory != true)] | length' .moai/reports/t376/lint-baseline.json`, re-run in
this tree), and `.github/workflows/spec-lint.yml:40` runs `go run ./cmd/moai spec lint --strict` on
every push to `main` and `develop`. The strict gate is green today because nothing gates, not because the corpus is clean.
The first non-advisory finding this rule emits reddens the integration branch — which is the intended
effect of §A.5 D2, but it is a consequence someone must accept knowingly, at the point where the
number exists.

Note the ordering: this AC is deliberately not satisfiable in advance. The decision is recorded
**after** M2 produces the count, never predicted before it.

## §D.1 — Definition of Done

- [ ] All 20 ACs (AC-STV-001..013 incl. 007a, plus AC-STV-014..019) satisfied, with the command run
      and its output cited.
- [ ] The bidirectional set (AC-STV-001/002/014 firing, AC-STV-004..008 and AC-STV-017 silent) is
      satisfied in one execution, with the AC-STV-010 control asserted in that same execution.
- [ ] The throwaway probes `internal/spec/zz_t376_probe_test.go` and
      `internal/spec/zz_t376_census_test.go` are removed, their coverage having been converted into
      the regression test set above (the census log stays as evidence; the scratch test does not).
- [ ] `go test ./internal/spec/...` passes; the affected-package result is cited, and the full-suite
      verdict is left to CI.
- [ ] The per-code baseline comparison of AC-STV-013, the advisory split and gating decision of
      AC-STV-019, and the intersection result of AC-STV-016 are all recorded in `progress.md` §E.2.
