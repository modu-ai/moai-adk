# plan.md — SPEC-DISPATCH-INTEGRITY-001 (card t1595)

Stateless plan artifact — the SPEC's lifecycle lives in `spec.md`.

## §A Context

Card t1595 bundles 8 P1 review findings from the t1498 round: three bundle
predecessor-semantics defects, three card-selection/after defects, one
delivery-ordering defect, one memory-fold concurrency defect. All four
production files live in `internal/cli` (`factory_bundle.go`,
`factory_card.go`, `factory_card_pr.go`, `memory_fold.go`). Development mode
is tdd. Working tree: card worktree `.moai/worktrees/t1595`, branch
`WT-dispatch-integrity`, baseline SHA 81786284e.

RED assets: the t1498 overlay suite was relayed through `/tmp` (volatile) and
is archived machine-locally at `.moai/reports/t1595/overlay/` — gitignored by
policy (`.gitignore:235`), so assets enter the branch only as COMMITTED TEST
FILES of the owned subset (M0/M1). The suite's canonical drop-in set is ONE
file per package: `cli_all_test.go` is the superset (7 tests;
`cli_more_test.go` is a strict subset and `internal_cli_review_observation_test.go`
an earlier 4-test variant — drop only the superset), plus the worktree file
(1 test) and the hook file (2 tests). `final-results.txt` is the FAIL
baseline record of the original observation (2026-10-07, different tree).

## §B Known Issues — baseline classification (measured, tree 81786284e)

Baseline-attribution: commands run 2026-10-08 on THIS tree at 81786284e,
env-scrubbed compound form, verbatim outputs below. M0 re-affirms before any
fix — absorbing develop invalidates pinned classifications.

| Defect | Card-text site | RED anchor | Baseline on 81786284e | Class |
|---|---|---|---|---|
| (1) multi-hub after | factory_bundle.go:211 (obs-tree) | none — author in M1 | not measured; code read: record stores a single tail hint, selection-level `factoryHubWaitUnmerged` mitigates (hypothesis, not measurement) | UNDETERMINED → M1 |
| (2) first-member hub | factory_bundle.go:121 (obs-tree) | none — author in M1 | not measured; code read: head hub-hint block present (t1533 r2a) — hypothesis | UNDETERMINED → M1 |
| (3) duplicate member | factory_bundle.go:84 (obs-tree) | none — author in M1 | not measured; code read: no dedup in `runFactoryBundleLocked` — expected LIVE | UNDETERMINED → M1 |
| (4) explicit-after overwrite | factory_card.go:1312/:1316 | TestReviewFindingNominatedOverwritesDependency | PASS — `err=factory next: predecessor card not merged: t1 has not reached merged-local (git-flow) or merged-pr (github-flow) t3 state=picked after="t1"` | NOT REPRODUCED → regression guard M2 |
| (5) merged-pr successor | factory_card.go:719 | TestReviewFindingMergedPRPredecessor | PASS — `predecessor=merged-pr; factory next leased="t2"` | NOT REPRODUCED → regression guard M2 |
| (6) b2 whole-dispatch abort | factory_card.go:823 | none — author in M2 | not measured; code read: b2 pre-filter present (t1533 r6/r7) — hypothesis | UNDETERMINED → M2 |
| (7) retry before lease check | factory_card_pr.go:254 | none — author in M3 | not measured; code read: holder+expiry checks precede remote mutations (t1533 r2c/r5) — hypothesis | UNDETERMINED → M3 |
| (8) fold store race | memory_fold.go:647 | TestReviewFindingFoldConcurrentWrite + TestReviewFindingFoldInterleavedArchiveLoss | FAIL both — `concurrent update between recheck and rename lost without refusal`; `completed fold B's line disappeared from both indexes after fold A resumed` | RED-LIVE → fix in M4 |

Four-element baseline cells for the (4)/(5)/(8a)/(8b) rows: acceptance.md
evidence ledger EL-001..EL-004 (single-invocation command, raw stdout,
exit code, tree SHA 544462a8d — Go bytes identical to 81786284e). The §B
fragments above remain the at-a-glance view; the ledger is the adoption
carrier.

Auxiliary overlay observations (same run): `TestReviewFindingLaneProbe` FAIL
(3/3 subtests) — deferred to t1596 (decision-index Q1);
`TestReviewFindingExplicitCandidateFiles` FAIL, `TestReviewFindingDryRunFactoryMigration`
FAIL — t1596 scope; `TestReviewFindingWhitespaceLanding` PASS — t1561 (#1795)
landed, mapping-resolved; `TestReviewFindingZoneExistingDotDot` PASS —
methodology control green; `TestReviewFindingBackgroundReceiptRecycling` FAIL
— t1562 (#1803 in flight), excluded.

Defect (8) mechanism note (from the interleaved observation): the fold writes
the archive BEFORE memory ("archive-preceding"); fold B completing inside
fold A's seam window gets its archive update overwritten by A's pre-B
snapshot rename while A then refuses the memory write — B's line lands in
neither index. The lock must therefore span BOTH index writes of one fold,
not per-file.

## §C Pre-flight

- Env-scrub verification runs as ONE compound invocation per package:
  `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test …`.
- Heavy internal/cli family runs take the slot lease
  (`moai slot acquire --resource <shared-target> --max-duration <bound>`
  first; release before the completion report) per gitflow-lane-protocol §8.
- Re-read `git rev-parse --short HEAD` + `git branch --show-current` before
  every commit; stage by explicit pathspec; never `git add -A`.
- Worktree guard: git operates on THIS tree only; no `git -C` to other trees.
- No full `go test ./...` in-lane (C2).

## §D Constraints

Spec C1–C6 apply. Additional:

- Committed test set boundary (decision-index Q2): only t1595-owned tests
  (MergedPRPredecessor, NominatedOverwritesDependency,
  FoldConcurrentWrite, FoldInterleavedArchiveLoss) plus the green control
  (ZoneExistingDotDot) enter the branch; foreign-card tests stay in the
  overlay archive for their owners.
- `.moai/reports/` is gitignored: run evidence lives in `progress.md` §E.2
  and the leader-read verdict file; scratch measurement output stays in
  `.moai/state/verify/` and is never cited.

## §E Self-Verification (plan-phase)

Verified this phase: overlay baseline measured on this tree (§B, three
package runs, verbatim; four-element ledger cells EL-001..004 captured
2026-10-09 in the round-1 repair session); SPEC-ID regex check PASS
(verbatim `PASS`); ID collision check (no SPEC-DISPATCH-INTEGRITY in
catalog); frontmatter validated against the 12-field schema SSOT; overlay
assets archived into the card tree. Spec lint: run at initial authoring
(exit 0, observed in the authoring session), re-run by the round-1 audit
(exit 0), and re-run after the round-1 repairs (result stated in the
repair session's report); the durable record is M0's ledger duty — the
first §E.2 entry carries the lint command and its output. Not verified
(gaps): live status of defects (1)(2)(3)(6)(7) — classification requires
the M1–M3 characterization tests; hook-package helper surface assumed from
the passing baseline run only.

## §F Milestones

Review-priority note (decision-reversibility): M4 carries the leading design
decision — the cross-process lock shape (new mechanism, highest
change-likelihood). M1-(3) is the only other defect expected live. M2/M3 are
classification-heavy (expected regression guards) and mechanically lowest.
Execution order below is the dependency order; reviewers should read M4 and
M1 first.

- **M0 — Overlay intake & baseline re-verification** (P1). Drop the
  canonical set (internal/cli superset → `internal/cli/review_observation_test.go`;
  worktree file; hook file), run the env-scrubbed `-run 'TestReviewFinding'`
  pass per package, compare per-test against §B, re-classify any divergence
  (the tree may have absorbed develop since 81786284e), record the
  classification into `progress.md` §E.2, then trim the drop-ins to the
  owned subset before any commit — INCLUDING the re-authoring of
  `TestReviewFindingFoldInterleavedArchiveLoss` to the serialized shape
  AC-DI-010 pins (fold B a separate process completing after A releases
  the lock; strengthened-assertion duty per AC-DI-010) and the measurement
  of that re-authored body's RED-now into the ledger. Exit:
  classification table re-affirmed on the current tree, and every baseline
  cell — including the re-authored AC-DI-010 body's RED-now and the
  strengthened AC-DI-005/009 bodies — captured into the acceptance.md
  evidence ledger in the §2.1 four-element form (single-invocation
  command, raw stdout, exit code, tree SHA), with anchored selectors
  (`-run '^TestName$'`, D11) and the measurement input named — the
  canonical committed drop-in per `owned-tests/README.md`, never the tree
  SHA alone (D10).
- **M1 — Bundle predecessor semantics, defects (1)(2)(3)** (P1). (a) Author
  characterization tests RED-first into the committed internal/cli test file
  from the existing helper surface (`fcFixture`, `fcQueue`, `fcClassify`,
  `fcPlace`, `fbSeedFiles`, `runFactory "bundle"`, `fbLeasedCard`):
  (3) duplicate-member refusal; (2) first-member hub constraint; (1) member
  not lease-eligible past an unmerged hub sharer. Observe each on the
  current tree, classify two-cell. (b) Fix the LIVE ones — (3) expected
  live: refuse duplicate ids before any record (a refusal, never silent
  dedup). (c) Re-measure the `factory_bundle_test.go` family (13 tests); if
  selection code was touched, the `factory_card_test.go` family (10 tests)
  too.
- **M2 — Card selection & after-overwrite, defects (4)(5)(6)** (P1). Commit
  the (4)(5) overlay tests (trimmed) as regression guards with the
  not-reproduced evidence from §B re-affirmed on the current tree. Author
  the (6) characterization RED-first (b2 unmerged-sharer skip while a ready
  card progresses), classify — expected regression guard (pre-filter
  present). Re-measure the `factory_card_test.go` (10 tests) AND
  `factory_nominate_test.go` families — the nomination path's own family,
  which the codex gate's injection flipped to FAIL while the listed
  families passed (D14).
- **M3 — Merging-retry lease validation, defect (7)** (P1). Author the
  characterization RED-first from the `factory_card_pr_test.go` delivery
  fixtures (local-remote pattern): a foreign-lane retry and an expired-lease
  retry must refuse BEFORE any remote mutation (the fixture remote observes
  zero mutation). Classify — expected regression guard. Re-measure the
  `factory_card_pr_test.go` (15 tests) + `factory_card_pr_guard_test.go`
  families.
- **M4 — Memory-fold cross-process serialization, defect (8)** (P1, LIVE).
  Fix: a cross-process lock spanning the fold's whole write transaction
  (both index writes inside one lock hold — §B mechanism note), with the
  write geometry brought to the consistent four-step shape — byte
  comparison (cmp1) → seam probe → NEW final byte comparison → rename —
  the post-probe comparison being the last check before each rename (D9;
  it is the only defense against a non-cooperating writer — the
  FoldConcurrentWrite author writes without any lock). The serialized
  AC-DI-010 test body was re-authored and RED-measured in M0 (gate-P2
  sequencing); M4 delivers the lock that flips it green and verifies
  GREEN. Guarantee scope: the lock closes
  the write window for cooperating writers; for non-cooperating writers the
  detection at the last observable byte comparison is the defense, and the
  irreducible TOCTOU tail between that comparison and the rename is stated
  as residual risk in AC-DI-009 — never absolutized. Both fold tests flip
  green. Re-measure: `go test -count=5 -race` over the memory-fold family
  (`memory_fold_test.go`, 14 tests, plus `memory_fold_wiring_test.go`);
  record the exact selector in §E.2 before running — a selector matching
  zero tests is a failed measurement, not a pass. MX (autonomous): the lock
  helper is an `@MX:ANCHOR` candidate at fan-in ≥ 3 and an `@MX:WARN`
  candidate (cross-process locking) per `mx-tag-protocol.md`.

## §G Anti-Patterns

- Do NOT satisfy `TestReviewFindingFoldConcurrentWrite` by relocating the
  `orderProbe("bytes-done")` call site — in EITHER direction: not UPSTREAM
  of the byte comparison that precedes it (that existing recheck would
  then refuse the test's write while the post-comparison window — where
  the original defect manifested — stays unguarded), and not DOWNSTREAM
  into a window a real concurrent writer could not occupy. The post-fix
  write geometry is fixed: byte comparison (cmp1) → seam probe → NEW final
  byte comparison → rename. A byte comparison must follow the probe in
  every acceptable geometry, and that follower IS the last check before
  the rename — "final byte comparison" is a ROLE the post-probe comparison
  takes, not a fixed call site (plan-audit iter-2 D9; codex gate 재발화 2,
  gate option (a)). The refusal must be real detection at that boundary;
  the guarantee is scoped to the comparison that follows the probe, and
  the irreducible TOCTOU tail between it and the rename is AC-DI-009's
  stated residual risk — do not absolutize the guarantee and do not weaken
  the lock span (AC-DI-010) to make an old test body pass.
- Do NOT re-implement (4) or (5) — they measure PASS on this tree;
  re-implementing double-writes behavior already pinned.
- Do NOT commit foreign cards' RED tests (t1596 ×2, t1562 ×1; t1561's green
  test likewise stays with its owner).
- Do NOT "fix" (3) by silently dropping the duplicate id — REQ-DISPATCH-003
  demands a refusal.
- Do NOT run the full suite in-lane; do NOT cite `/tmp` or
  `.moai/state/verify/` scratch as evidence.
- No time estimates in any artifact or report.

## §H Cross-references

`spec.md`, `acceptance.md`, `decision-index.md`, `progress.md`, and the
tracked deferral archive `deferred/` (same directory);
`.moai/reports/t1595/{handoff-to-card-session,lane-wait-20261008}.md`
and `.moai/reports/t1595/overlay/` (local-only); cards t1498, t1533, t1561,
t1562, t1596; `verification-claim-integrity.md`; `verification-completeness.md`;
`gitflow-lane-protocol.md` §8.
