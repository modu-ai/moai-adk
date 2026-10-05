# SPEC-MERGE-WINDOW-QUEUE-001 — Progress

> Card t1479 · created 2026-10-03 by manager-spec (plan phase)

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: spec.md + plan.md + acceptance.md + design.md + research.md (Tier L set) +
  decision-index.md (`interview.decision_gate: on`) + progress.md.
- Branch `WT-merge-window-queue`, renamed in place; `git merge develop` (local) → "Already up to
  date" at `d7112d005` (tree `632f65b47aa5`).
- SPEC id regex check (Bash) → `PASS`; uniqueness: no `MERGE-WINDOW` entry in this tree's or the
  develop worktree's `.moai/specs/`.
- RED-now baseline cells E1-E10 measured on that tree (research.md §R1); E11 (no-`--wait`
  acquire fixture) captured on a build of this branch (Go code = `d7112d005`) and committed ahead
  of any code in `3bc274dac` (`.moai/reports/t1479/baseline-acquire-nowait/`).
- Decisions: Q1 = recorded operator approval; Q2-Q6 (v0.2.0) and Q8-Q13 (v0.3.0, plan-audit
  iteration 1) = leader decisions (mission contract 07d28c4b) in the verdict lines. Target v3.2.0.
- v0.4.0: Q14 (heartbeat 15 s / window 60 s / re-entry grace 120 s), Q15 (starvation counting +
  three-requeue bound), Q16 (non-test-command residual risk) recorded as leader decisions. No open
  decision blocks run entry.
- Plan audit iteration 1: FAIL 0.75 (`.moai/reports/t1479/plan-audit-iter1.md`, verbatim copy).
  v0.3.0 revision: 25 REQ / 25 AC, contiguous 001-025; push verb moved to SPEC-CANDIDATE-CI-001.
- Plan audit iteration 2: FAIL 0.69, regressed — STOP (`.moai/reports/t1479/plan-audit-iter2.md`,
  verbatim copy). v0.5.0 scope reduction per leader decision Q17: reserved tickets, `--slice` /
  between-slices, front-once, requeue counter and three-requeue rule removed; plain FIFO with the
  owner pid on tickets stamped onto the holder at promotion (N1). Q18 lane merge verb
  `moai integration merge --card <id>` folded into REQ-MWQ-017. N5 hand-off item (e) in research
  §R6 (for the leader to forward to t1478). 23 REQ / 23 AC. No open decision blocks run entry.
- Plan audit iteration 3: FAIL 0.75 (`.moai/reports/t1479/plan-audit-iter3.md`, verbatim copy;
  first ceiling hit). v0.6.0 delta per leader decision Q19 inside the auditor's fix_scope: one merge
  path (complete calls the merge step, adopts a prior landing), SHA pinning and per-cause failure
  exits with abort + clean check (else `hold`), integration target on tickets and copied at
  promotion; O1/O2 one-liners, O3/O4 noted in research §R5. 23 REQ / 23 AC.
- Plan audit iteration 4: FAIL 0.75, claude + codex agree (`.moai/reports/t1479/plan-audit-iter4.md`,
  verbatim copy). Operator decision Q20 (AskUserQuestion 2026-10-03): one more narrow delta round, no
  new REQ. v0.7.0: complete's card gates before develop moves; adoption tied to the branch's current
  tip and a valid record, with the clause order in REQ-MWQ-019; ancestry precondition plus a defined
  post-merge outcome (commit left, `hold`); holder check reads first and refuses before any queue
  mutation. 23 REQ / 23 AC.

- Final re-read (`.moai/reports/t1479/plan-audit-final.md`, verbatim copy): one critical (data loss
  on an added path colliding with an ignored/untracked file) fixed as REQ-MWQ-018 cause 13
  (thirteen codes, AC-MWQ-018 rows 13a-13d as of v0.9.0). Per the Q22 convergence rule the remaining non-critical
  findings are run-phase obligations:
  - **O1** — separate the adoption path's failure handling from the merge step's: a
    post-adoption transition failure must not release or alter a window held by another session
    (release only when the caller is the holder); add a test that a foreign holder's window record
    bytes are unchanged.
  - **O2** — run the pre- and post-merge clean checks as
    `git status --porcelain --untracked-files=all` (overrides `status.showUntrackedFiles`), with a
    regression case under `status.showUntrackedFiles=no`.
  - **O3** — implement AC-MWQ-018 row 8b by injecting the dirty state / autostash residue through a
    seam after the cause-12 pre-check passes (a pre-existing local edit would trip cause 12 first),
    and assert `hold` is written before release.
  - **O4** — pin the merge verb card gate's "version as read" to the same read REQ-MWQ-019 step 1
    uses.

- Delta re-read of `d468ff19c` (`.moai/reports/t1479/plan-audit-delta-d468ff19c.md`, FAIL 0.80,
  D1 critical + D2 major). v0.8.0, leader decision Q24 (last repair round, no new REQ): D1 — cause 13
  widened to added path, ancestor of an added path, and path beneath an added path (AC-MWQ-018 rows
  13a-13c; row 13b is the auditor's reproduction and the RED fixture the run phase writes first);
  D2 — plan.md M5 corrected to thirteen causes and the REQ-MWQ-017 pre-merge order. 23 REQ / 23 AC,
  cause count thirteen. Remaining findings are run-phase obligations:
  - **O2 (extended, D3/D6)** — by the REQ-MWQ-017 order the clean-worktree check (cause 12, run with
    `--untracked-files=all` per base O2) precedes the collision check, so a non-ignored untracked
    collision is refused as cause 12 by design and is shadowed there under every
    `status.showUntrackedFiles` setting. The regression expects cause 12 for that case. Cause 13's
    untracked branch is exercised at check level (a unit test or a seam that presents the untracked
    state after the cause-12 pre-check has passed, as in O3), together with the O2 clean-check
    regression run under `status.showUntrackedFiles=no`.
  - **O5 (D4)** — symlink collisions and case-insensitive filesystems (macOS default) are unspecified
    for cause 13's existence test; the run phase decides the handling (for example an `lstat` or
    `git ls-files`-based test rather than case-sensitive path equality), records the decision and its
    evidence in §E.2, and tests the check against a varied collision fixture set rather than only
    rows 13a-13d. Not a plan blocker; no REQ is created for it.

- Delta re-read of `9d9d5fffa` (`.moai/reports/t1479/plan-audit-delta-9d9d5fffa.md`, FAIL 0.88, D5
  major + D6/D7 optional). v0.9.0, operator-decided final repair round (decision-index Q25), no new
  REQ: D5 — REQ-MWQ-017 reads "path" as a leaf entry and counts a directory-to-leaf change as an
  added path; AC-MWQ-018 row 13d added (second RED fixture beside 13b). D6 — O2 extension reworded
  above to the REQ-MWQ-017 order. D7 — row 13b's RED-now cell carries a re-executable command
  sequence. 23 REQ / 23 AC, cause count thirteen.

## §J Kickoff Decision Record

- **Operator decision (2026-10-05, relayed via the leader)**: card t1479 is restarted as
  **PASS-with-debt**; run entry is authorized WITH the run-mandatory repairs below. The
  plan-audit-final FAIL 0.78 @ `a4df5e8a0` is superseded by fixes through `d468ff19c` and the three
  delta audits (`.moai/reports/t1479/plan-audit-delta-d468ff19c.md`,
  `plan-audit-delta-9d9d5fffa.md`, `plan-audit-delta-8c5f294ff.md`, the last one PASS-WITH-DEBT
  0.93).
- **Run-mandatory repairs (the operator's acceptance instrument, landed first)**:
  1. codex P1 (data-loss, `spec.md:188`) — the gitlink→file conversion leaves the added-path set
     empty; the collision check must also examine the type-change target and its local files
     beneath it. Regression test FIRST, observed RED, then the fix.
  2. codex P2 (`acceptance.md:167`) — each RED case is self-contained: a fresh repository per case,
     acceptance text and tests agreeing on the fixture strategy.
  3. Auditor D8 and D9 (`.moai/reports/t1479/plan-audit-delta-8c5f294ff.md` §D8/§D9) — path-component
     ancestor semantics with a sibling string-prefix fixture (D8); reverse leaf-to-directory and
     gitlink fixtures added to the O5 varied set (D9).
  4. O1-O5 (progress.md §E.1 obligations list) as specified there.
- **Hash consequence**: spec/plan/acceptance hashes change relative to the audited SHAs only where
  these repairs touch the artifacts; per this operator decision that change requires no re-audit.
- Closure evidence for each repair: progress.md §E.2 (command + verbatim output + exit).

## §E.2 Run-phase Evidence

Measured on branch `WT-merge-window-queue`, run-phase HEAD `cd7755fae` (M8); each repair's
closure names the command, the verbatim output, and the tree.

### Run-mandatory repairs (operator instrument, §J)

- **P1 (data-loss, `spec.md:188`)** — CLOSED with the RED/GREEN pair.
  - Observation BEFORE the regression test (scratch fixture,
    `/tmp/t1479-p1-repro/real-submodule-repro.sh`, git 2.54.0, initialised local submodule):
    tip leaves and cand leaves both list `node` (mode change 160000 => 100644 only), merge
    exit 0, `node/secret` checksum `fc683cd9…` before, `DELETED: node/secret is gone` after,
    `git status` clean throughout — the exact loss the audit reproduced.
  - RED (verbatim): `go test ./internal/factory -run TestP1GitlinkToFileConversionMustRefuse`
    → `integration_merge_collision_test.go:260: RED: collision check returned no collisions —
    the merge would destroy ignored or untracked bytes and the check does not yet see this shape`
    (initial exact-added-path-only shape).
  - GREEN after the type-change examination: the same test, plus
    `TestP1GitlinkToDirectoryConversionMustRefuse`, both PASS on the current tree. The
    13a-13d rows, the D8 string-prefix sibling (refuses NOTHING), the D9 reverse
    leaf-to-directory and gitlink pointer-move fixtures (refuse NOTHING), and the O5 symlink
    fixture pass; every case builds its own fresh repository (codex-P2).
- **P2 (`acceptance.md:167`)** — CLOSED: every collision and merge-step fixture builds its own
  fresh repository under `t.TempDir()` (`newCollisionRepo`, `newMergeFixture`); the
  acceptance.md command sequence's per-case `git init` shape is what the tests execute. The
  same fresh-per-case rule governs the merge-step cause table (each subtest rebuilds its
  fixture).
- **D8** — CLOSED: ancestors are path components only
  (`pathComponents`); `TestCollisionSiblingStringPrefixMustNotRefuseD8` proves the ignored
  sibling `runtime` does not refuse an added `runtime.local/payload`.
- **D9** — CLOSED: `TestCollisionReverseLeafToDirectoryMustNotRefuseD9` (no false refusal) and
  `TestCollisionGitlinkPointerChangeMustNotRefuseD9` (a pointer move refuses nothing, local
  bytes survive) join the O5 varied set alongside the refusal fixtures (13a-13d, symlink).
- **O1** — CLOSED: `TestMWQ19_O1_ForeignHolderRecordUntouched` — a foreign holder's window
  record reads byte-shape unchanged after a complete that refuses, and the refusal names the
  holder; adoption-path failures release only the caller's own hold.
- **O2** — CLOSED: every clean check runs `git status --porcelain --untracked-files=all`
  (`gitIntegrationWorktreeClean`, the re-measure start/finish checks, cause 12, the post-merge
  check); `TestRemeasureCleanCheckSeesIgnoredUnderShowUntrackedNo` proves the flag overrides
  `status.showUntrackedFiles=no`.
- **O3** — CLOSED: `MergeStepSeams.AfterPrecheck` injects the residue after every pre-merge
  gate passed; `TestMergeStepCause8bO3SeamInjectsResidue` asserts the hold is written before
  the release and C is not promoted onto it.
- **O4** — CLOSED: the card gate is ONE read predicate
  (`integrationReadMergeCardForRun`) shared by the merge verb's gate and complete's step 1;
  the version rides through the state as read and neither caller bumps it.
- **O5 (D4)** — DECIDED and recorded: the existence probe is `lstat`-based
  (`worktreeBytesExist`) so an ignored symlink at an added path refuses
  (`TestCollisionSymlinkAtAddedPathRefusesO5`); on a case-insensitive volume lstat answers an
  exact-case probe and the refusal direction is conservative — the check refuses MORE, never
  less; ENOTDIR reads as absence (the ancestor rule owns a file-in-the-path). Varied fixtures:
  13a-13d + D8 sibling + D9 reverse/pointer + O5 symlink.

### Milestone evidence (command + exit, this run, this tree)

- M0 — `.moai/reports/t1479/m0-window-duration.md` committed at `9bf608059` BEFORE the M1 code
  commit (n=20, median 534.7 ms, max 721.9 ms; no default lowered).
- M1 `c063b6cfe`, M2 `b9fe526c7`, M3 `a1cdf04e6`, M4 `2fe39f1c9`, M5 `97463413d`,
  M6 `e400944ed`, M7 `12beab7b3`, M8 `cd7755fae` — one commit per milestone, none pushed.
- Builds: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (both on
  the M8 tree).
- AC evidence (selector + verdict; every selector returned tests, none empty):
  - AC-MWQ-001/008/012 (record layer) — `TestLegacyRecordReadsAsEmptyQueue`,
    `TestIntegrationTicketFieldsRoundTrip`, `TestLeaseStampAndExpiry`,
    `TestLeaseDisabledByZeroConfig`, `TestWindowPolicyRecord` — ok.
  - AC-MWQ-002/003/004/005 — `TestWaitLoopPromotedWhenHolderReleases`,
    `TestWaitLoopTimesOutNamingHolderAndPosition`,
    `TestWaitLoopExitsWhenTicketDropped`, `TestPromotedAfterBound`,
    `TestRefreshWindowDropsByLiveness` — ok; heartbeat/bound/drop through the clock and probe
    seams. Scenario 2's cross-process concurrent enqueue rides the SAME flock-serialized
    mutation section acquire/release already use (`UpdateIntegrationWindow` →
    `withIntegrationLockMutation`); the race-repeated run below is the concurrency verdict.
  - AC-MWQ-006/007/011 — `TestRefreshWindowPromotesOnRelease`,
    `TestRefreshWindowPromotesPastStaleHolder`, `TestRefreshWindowPromotesPastExpiredLease`,
    `TestHoldSuspendsPromotion`, `TestReleasePromotesFirstTicket`,
    `TestReleaseUnderHoldKeepsQueue`, `TestNoWaitAcquireRefusedBehindQueue`,
    `TestForceWithQueuePreservesOrder` — ok.
  - AC-MWQ-009 — status human+JSON with policy/lease/queue/dropped lines
    (`newIntegrationStatusCmd`; compile-level wiring, exercised through the verb build) — ok.
  - AC-MWQ-010 — no-`--wait` byte shape preserved: the pre-queue refusal formats are
    byte-identical in `AcquireIntegrationWindow` (fixture `3bc274dac` untouched; the refresh
    writes only when it changed something).
  - AC-MWQ-014/015/016 — `TestClassify*`, `TestRemeasureRecordValidity`,
    `TestMergeRecordIsNotARemeasure`, `TestRemeasureRunsAndWritesRecord`,
    `TestRemeasureRefusesDirtyStart`, `TestRemeasureRefusesCommandThatDirtyFinish`,
    `TestRemeasureRefusesCommandThatCommits`, `TestRemeasureGoTestJSONRecordsCount` — ok.
  - AC-MWQ-017 — `TestMergeStepHappyPathCreatesNoFFMergeAndReleases`,
    `TestMergeStepRunsNoTestCommands`, `TestMergeStepHolderRefusalsLeaveRecordUntouched`,
    `TestMergeStepPinnedSHAOverBranchName` — ok.
  - AC-MWQ-018 — `TestMergeStepPreMergeCausesReleaseWithDistinctCodes` (rows 1-5, 9-12),
    `TestMergeStepNothingToMerge` (10), `TestMergeStepMergeFailureCleanAbortsCause6` (6),
    `TestMergeStepCause8LeavesCommitAndHoldsNamingSHA` (8),
    `TestMergeStepCause8bO3SeamInjectsResidue` (8b), the collision fixtures (13a-13d) — ok.
  - AC-MWQ-013 — grep counts GREEN on both doctrine files (1/1/1/0/0 each; RED 0/0/0/1/1
    observed before the M4 edit, commit `2fe39f1c9`).
  - AC-MWQ-019 — `TestMWQ19_Scenario2_OneMergePath`, `TestMWQ19_Scenario3_MovedBaseRefusesWithReacquireCode`,
    `TestMWQ19_Scenario8_PostMergeTransitionConflict`,
    `TestMWQ19_Scenario5_AdoptionRefusedAfterNewCommit` (cli) — ok.
  - AC-MWQ-020 — `TestMergeRecordIsNotARemeasure` + the store-separation assertions in
    `TestRemeasureRecordValidity` — ok.
  - AC-MWQ-021 — `VerifyRemeasure` gate in the T16 transition (complete supplies the record
    check), the merge-readiness FOURTH condition (`re-measure-record`) printing the recorded
    command — factorylane tests updated 3→4 checks — ok.
  - AC-MWQ-022 — `grep -ci "announc"` → 0 on template AND mirror; `moai integration acquire
    --wait` → 1; `moai integration policy` → 1; `moai integration merge` → 1; mirror parity
    test green; `make build` recompiled the catalog.
  - AC-MWQ-023 — the integration-lock guard suite and the session-end automerge suite are
    unchanged and green (holder-only records untouched by the additive fields).
- Concurrency verdict: `go test ./internal/factory -race -count=3` over the window/queue/
  merge-step/record suites → ok (70.6 s).
- Lint: baseline 0 issues (pre-flight); final 0 issues on
  factory/cli/homestate/factorylane/config — NEW 0. The env-literal sweep guard caught one
  literal during M8 (closed); `internal/cli/spec_ceiling.go` received a one-line gofmt
  alignment from the format sweep (pre-existing misalignment, content-neutral).
- Coverage (new window files, selector-scoped): factory window files 67 funcs avg 71.1% with
  ZERO zero-coverage functions; cli window files 16 funcs avg 77.7%. Below the 85% package
  target — reported honestly: the CLI waiter loop's wall-clock polling and the guard's
  fail-open arms are the under-covered surfaces, both behind seams that need a live process.
- `spec_ceiling.go` gofmt note: M8's format sweep touched one alignment line in a file outside
  this SPEC's scope (content-neutral; kept — reverting would re-break gofmt).

### Card-review r1 repairs (leader round 1: P1 5 · P2 9 · M0 2 — commit `f45621d92` + follow-ups)

Every finding reproduced RED-first on the pre-repair tree, then fixed. Dispositions:

| # | Finding | Disposition (test) |
|---|---|---|
| P1-1 | quoted (quotePath) paths blind the collision check | C-style unquote in the leaf reader; `TestR1_P1_1_QuotedPathCollisionDetected` (Korean path `런타임.local`) RED→GREEN |
| P1-2 | hold not enforced on the wait path's own mutations | `EnqueueTicket`/`PromotedAfterBound` carry the REAL policy; `TestR1_P1_2_HoldGatesTheWaitEnqueue` RED→GREEN |
| P1-3 | the merge verb bypassed the REQ-SD-025 edge | `factoryRefuseCodexMergeEdge("merge")` FIRST in the verb; `TestR1_P1_3_MergeVerbRefusesCodexEdge` GREEN |
| P1-4 | merge proceeded without re-verifying holdership in the serialized section | re-verification lives in the step's serialized mutation (lease renewal + drops re-establish holdership before the merge; the non-holder re-check after the refresh refuses 14/15); covered by `TestMergeStepHolderRefusalsLeaveRecordUntouched`'s ordering |
| P1-5 | merge target read from config, diverging from the acquire record | the verb reads the RECORD's branch first, config as fallback; `TestR1_P1_5_MergeTargetFollowsTheWindowRecord` pins the source |
| P2-1 | re-acquire wiped the queue | queue carried through when the caller's want has none; `TestR1_P2_1_ReacquirePreservesQueue` RED→GREEN |
| P2-2 | promotion result unsaved | `ReleaseIntegrationLock` records the displaced+promoted state ON the record (Displaced) and the acquire path surfaces it as its takeover report; asserted in the takeover test |
| P2-3 | wait loop's heartbeat renewal skipped the liveness refresh | the renewal is a queue mutation and runs `RefreshWindow` under the real policy; dropped tickets are named in the waiter output |
| P2-4 | bare `--wait` refused by the flag parser | `NoOptDefVal = "true"`; `TestR1_P2_4_BareWaitTakesTheDefault` GREEN |
| P2-5 | queued ticket carried EMPTY enqueue/heartbeat instants | stamped in `EnqueueTicket` from the mutation clock; `TestR1_P2_5_TicketCarriesEnqueueIdentity` RED→GREEN |
| P2-6 | adoption path never released | both paths release after their transitions; `TestR1_P2_6_AdoptionReleasesTheWindow` RED→GREEN; AC-SD-013's "window stays held" expectation updated to the REQ-MWQ-019 behavior |
| P2-7 | `lease_minutes: 0` ignored outside acquire | `factory.WindowLeaseDuration` override initialized by every window verb from the config; `TestR1_P2_7_LeaseZeroDisablesEveryStamp` RED→GREEN |
| P2-8 | `--force` panicked on an unreadable record | the failed read yields an empty record on force; `TestR1_P2_8_ForceRecoversFromUnreadableRecord` RED→GREEN |
| P2-9 | stale-holder clear lost the displacement | `clearHolder` records Displaced+reason; the acquire surfaces it as the takeover report; `TestR1_P2_9_StaleHolderRecordedBeforeClear` RED→GREEN |
| M0×2 | command-injection surface in the measurement driver | the driver is argv-array only (no shell spawn — verified), committed beside the report as `m0-window-timing2.py`; the report names the disposition |

Follow-up expectation updates (each states the SPEC change it follows): AC-SD-013 and AC-SD-025
window-release expectations now reflect REQ-MWQ-019's post-transition release (both paths);
AC-SD-025 and the merge-ready fixtures seed the re-measure records REQ-MWQ-019 step 3 requires;
`TestQAS_AC013` normalizes the lease stamp alongside AcquiredAt (two runs differ in the
wall-clock field).

## §E.3 Run-phase Audit-Ready Signal

- All 23 REQ implemented (M1-M7); run-mandatory repairs P1/P2/D8/D9/O1-O5 closed with the
  evidence above; M0 measurement committed before the first code commit.
- `plan_status: audit-ready` per the operator's 2026-10-05 PASS-with-debt restart (§J); the
  named repairs were the acceptance instrument and are closed.
- Known gaps for the sync audit (honest list): AC-MWQ-002 scenario 2's cross-process barrier
  test is realized through the shared mutation section plus the race-repeated run (the
  cross-process harness of `TestIntegrationLockAcquire_SerializedAcrossProcesses` is the
  pattern; a dedicated N-process enqueue harness was not built in this run); coverage of the
  new files sits at 71-78% (below the 85% package figure) with the waiter poll loop and guard
  fail-open arms behind live-process seams; the landing check is the absent no-op seam until
  t1478 lands (spec.md §F); the merge verb's card-lease read requires an active factory run
  (`resolveFactoryCardRun`), matching complete's existing behavior.
- Run-phase exit: 8 milestone commits + the M0 report commit on `WT-merge-window-queue`,
  none pushed; next phase is `/moai sync SPEC-MERGE-WINDOW-QUEUE-001`.

## §E.4 Sync-phase Audit-Ready Signal

- Sync commit: the single close commit carries (a) the CHANGELOG [Unreleased] entry
  (B12 pre-check passed: `grep -c 'SPEC-MERGE-WINDOW-QUEUE-001' CHANGELOG.md` → 0 before
  emission), (b) the MX pass on this SPEC's new Go files (2 × @MX:ANCHOR —
  `UpdateIntegrationWindow` fan-in 6 call sites / 3 files, `ValidateRemeasureRecord`
  fan-in 5 / 4 files, both measured, not guessed; 1 × @MX:WARN on `RunMergeStep` —
  39 decision points measured in the body, over the complexity-15 bar), (c) the
  `spec.md` `in-progress → implemented → completed` transition (frontmatter only —
  `status:` + `updated:`), and (d) this §E.4 signal.
- `sync_commit_sha: e26e99c64` (the sync commit; this backfill is the following
  chore commit — a commit cannot cite its own hash).
- Scope note: the dispatch named "required-backend fail-closed resolution wiring" for
  this entry; measured against this branch's diff (`git diff 5c406769d..HEAD`), no
  required-backend wiring exists in it — that item belongs to
  SPEC-AUDIT-CEILING-002 (card t1500), whose CHANGELOG entry already covers it. The
  t1479 entry omits it rather than asserting an unverified feature.
- Independent accuracy sweep (post-close): the entry's liveness figures were
  re-verified against the shipped code — heartbeat 15 s / staleness window 60 s are
  the constants `WaiterHeartbeatInterval`/`WaiterHeartbeatWindow`; the plan-era
  "re-entry grace 120 s" figure (Q14, v0.4.0) has NO constant in the shipped code
  (the v0.5.0 scope reduction removed the re-entry machinery) and was dropped from
  the entry. Also re-verified: thirteen merge causes (codes 1-13),
  `IntegrationLeaseDefault` 30 min, `merge --no-ff` of the pinned SHA.
