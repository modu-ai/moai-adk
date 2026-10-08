# decision-index.md — SPEC-DISPATCH-INTEGRITY-001

Stateless companion (no status axis; the SPEC's lifecycle lives in
`spec.md`). Authored because `interview.decision_gate: on`
(`.moai/config/sections/interview.yaml`). Every row follows
Detect → Explain → Ask; no row carries an embedded recommendation beyond a
`Default:` selected by the published rule.

### Q1: Which card owns the overlay's lane-probe finding (`TestReviewFindingLaneProbe` / `productionLaneFilesProbe`)?

Label: FOUNDER
Class: implementation-level
Why unresolved: the t1595 card text enumerates exactly 8 defects;
productionLaneFilesProbe (github-flow empty list + whitespace path
splitting) is none of them. The predecessor handoff names t1596/t1597 as
candidates without deciding, and no committed-tree authority — product.md,
a prior SPEC's HISTORY/Amendments row, a config-section setting, or the
constitution — assigns ownership of this finding, so the row cannot route
DECIDED or POLICY-COVERED. Card evidence under `.moai/reports/` is not an
authority register.
Default: Defer to t1596 (rule: preserves current behavior — this SPEC's
scope stays the card's 8 defects; undo is a single revert of this SPEC's
own commits)
Alternate: Scope the probe fix into SPEC-DISPATCH-INTEGRITY-001 (expands
the card beyond its 8-defect contract and adds a second production
surface — productionLaneFilesProbe — the card text never names)
Evidence note: RED on this tree at 81786284e, all 3 subtests (2026-10-08
run, plan.md §B) — the deferral records ownership, not the defect's
existence. Inheritance made durable on the branch (plan-audit round-1
repair, D5): the deferred test source and mapping are committed under the
TRACKED path `.moai/specs/SPEC-DISPATCH-INTEGRITY-001/deferred/`
(`lane-probe_test.go.txt`, `README.md` naming owner card t1596 and the
intake path). The owner-side record in t1596's queue body could not be
authored by this card — queue mutation is prohibited for factory lanes and
their spawned agents — so the tracked deferral archive is the two-sided
side this card lands; t1596's operator-side dispatch is where the owner
body gains its pointer. The overlay archive
(`.moai/reports/t1595/overlay/`, origin `/tmp/t1498-review-overlay/`)
stays the local evidence supplement.
Operator verdict: DEFAULT-APPLIED 2026-10-08T15:14Z manager-spec (lane t1595)

### Q2: Do the overlay tests owned by OTHER cards (t1596 ×2, t1561 ×1, t1562 ×1) get committed on the t1595 card branch?

Label: FOUNDER
Class: implementation-level
Why unresolved: the overlay archive carries tests owned by four other
cards — three of them RED on this tree (t1596's
ExplicitCandidateFiles/DryRunFactoryMigration, t1562's
BackgroundReceiptRecycling) and one green (t1561's WhitespaceLanding, fix
landed in #1795). Committing foreign tests would turn this branch's package
suites red on findings this SPEC does not own, and no constitution clause
or operator setting states the test-set boundary for a multi-card overlay
relay as written — so the row cannot route POLICY-COVERED.
Default: Commit only t1595-owned tests (MergedPRPredecessor,
NominatedOverwritesDependency, FoldConcurrentWrite,
FoldInterleavedArchiveLoss) plus the green control (ZoneExistingDotDot);
foreign-card tests stay in the overlay archive for their owners (rule:
preserves current behavior — branch suites stay green, CI judges the full
suite)
Alternate: Commit the full overlay suite as-is (three foreign REDs break
internal/cli and internal/hook on this branch; t1561's green test
duplicates another card's deliverable)
Operator verdict: DEFAULT-APPLIED 2026-10-08T15:14Z manager-spec (lane t1595)
