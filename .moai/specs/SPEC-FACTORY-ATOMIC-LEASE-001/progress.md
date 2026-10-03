# SPEC-FACTORY-ATOMIC-LEASE-001 — Progress

Card t1458, Tier M. This file is the phase record. Plan-phase writes only §E.1; §E.2 and §E.3 belong to the
run phase (manager-develop) and §E.4 to the sync phase (manager-docs).

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-03
- plan_start_head: 2de0a2cb613b04765a1554f86685a3b48e0be806
- artifacts: spec.md (14 requirements), plan.md (6 work milestones), acceptance.md (15 criteria), evidence/
  (ten probes and their overlays, re-executable from the repository root)
- plan_audit: iteration 1 FAIL 0.79; iteration 2 FAIL 0.81; iteration 3 FAIL 0.83; four-hunk confirmation
  not run — to be run by plan-auditor; no verdict is claimed here. (Iteration 1: Tier M threshold 0.80,
  audited tree `db692601307c28b6d1dd905ab1ab6f6d9bd1e974`, no must-pass criterion failed, the score was
  driven by Testability 0.55 and Clarity 0.70. Iteration 2: audited tree
  `c8b716fed24564a685188dc94b3f446ac9fc79c8`, no must-pass criterion failed, three must-fix defects PA2-M1
  to PA2-M3 and seven PA2-N1 to PA2-N7; the leader granted one extra repair round and one delta audit past
  the Tier M ceiling of 2, and accepts no PASS-with-debt on it. Iteration 3, that delta audit: audited tree
  `9f73f4cdf7af24af493edfb9e629f70aac915133`, no must-pass criterion failed, one must-fix defect I3-M1 and
  three should-fix defects I3-S1 to I3-S3, five notes I3-N1 to I3-N5. The leader's "Decision 2" approved an
  exception to "a second ceiling hit parks the card": this repair of four hunks, then a re-read of those
  four hunks only; if that confirmation is also blocked the card is parked.)

### 2026-10-03 — iteration 2 repair (spec.md 0.2.0)

Each audit defect id, where it was fixed, and how a re-auditor checks it. Measurements made in this
iteration are ledger rows L9–L18 in `acceptance.md`, on HEAD `db692601307c28b6d1dd905ab1ab6f6d9bd1e974`
(Go files equal to the plan-start tree), judged by `moai` build `0732cc699` (L18).

| Defect | Where it was fixed | Check |
|---|---|---|
| D1 | spec REQ-FAL-006 restated as an outcome, mechanism removed from it; plan D2 states both mechanisms and why (L14); spec §A.2 O6 corrected (the iteration-1 probe used a runtime PRAGMA) and O5 cites `factory.go` 473–486 | read REQ-FAL-006 last sentences; plan D2; O6; L14 |
| D2 | acceptance AC-FAL-009 clause (ii): per-path table plus the invariant; checked against `factoryNextClaim`, `factoryNextRecordAndClaim`, `RecordPicked`, `withCardTx`. **One cell differs from the audit's table:** bare arm (b2) makes 0 promotions, not 1, because its queue item is already `picked` | AC-FAL-009 (ii) |
| D3 | spec REQ-FAL-007 names the foreign-worktree directory check as the only outside read; AC-FAL-009 clause (iii) pins the allowed set (the read half stated doctrine-only); plan D4 | REQ-FAL-007; AC-FAL-009 (iii) |
| D4 | acceptance S4 (pass condition) and `-v` on every Command; per-name PASS count N; AC-FAL-010 given a literal selector (L9, 68 names swept, run exit 0); L10 shows the new selectors sweep 0 at the pin | S4; L9; L10 |
| D5 | AC-FAL-002/003 RED cells now cite L11 (clause (i), goroutine form, positive control; `completed inside section = true` at the pin); L1/L3 M2 probes relabeled context; evidence `probe-clause-i_test.go.txt`, `overlay-clause-i.json` | L11; AC-FAL-002/003 |
| D6 | plan §7 MU7 (and MU8–MU11); plan WM1 "cross-process lane helper (what it needs)"; AC-FAL-001 clause (c) and `TestFactoryLeaseSerialCrossProcessExactlyOne`; L12 (10 of 13 breaches from two processes); spec §H DL-3; evidence `probe-xproc_test.go.txt` | AC-FAL-001; plan WM1, §7; L12 |
| D7 | spec REQ-FAL-003 adds "within the queue lock's wait budget" and cites §F R6 | REQ-FAL-003 |
| D8 | spec REQ-FAL-006 bare-form outcome (end the pass at once, error, never an empty-queue report); §H DL-1 for the leader (worst cases: rejected design 5 × 3.3 s = 16.5 s; adopted one budget plus ≤ 50 ms); AC-FAL-007 (c) single outcome with a bound; plan D2; spec §F R13 | §H DL-1; AC-FAL-007 (c) |
| D9 | plan WM1 test list adds `TestHomestateDoesNotImportKanban` (non-test files only; L16: 0 vs 1 with `-test`) and `TestFactoryLeaseSectionRejectsNestedMutate`; AC-FAL-011 reclassified release-blocking until the tests exist (L10); spec REQ-FAL-010 scoped | AC-FAL-011; plan WM1; L16 |
| D10 | plan WM1 now three commits: seam-and-stub (compile-only stubs for every symbol a test names), baseline, RED; the AC-FAL-010 baseline is taken before the RED commit; Definition of Done item 1 | plan WM1; DoD 1 |
| D11 | spec §F R3 corrected (no stale clear on acquisition; the clear is wired to three other locks only, L17), R14 (step-lock crash), R15 (leased without a worktree and its recovery, with the audit's pointer to `factory_lane_relaunch.go` ~110 corrected: that line follows a new lease and is not a recovery); plan D3 failure outcomes; spec §D new Out-of-Scope bullets | §F R3, R14, R15 |
| D12 | AC-FAL-006: deterministic overlap criterion (ordered event log) with the forced-overlap probe (L13: 60 of 60 iterations overlapped); statistics restated as ranges (unforced 5% to 50% over eight runs, forced 55% to 75% over six; "one in ten million" withdrawn); spec §A.1 M5 row | AC-FAL-006; L13; L15 |
| D13 | spec §F R4 second shape (T2 landed, T3 timed out: row `assigned`, item `picked`, verb says `raced`) and third shape (arm (c) has no compensation); REQ-FAL-006 detail wording; AC-FAL-007 (b) mid-claim stall | §F R4; AC-FAL-007 (b) |
| D14 | spec REQ-FAL-004 reworded into the positive form | REQ-FAL-004 |
| D15 | AC-FAL-007 margin stated once (500 ms, a labeled heuristic); `factoryLeaseClaimWaitCap` defined in plan D2 as deadline plus busy timeout and used by AC-FAL-007/-008 | plan D2; AC-FAL-007/-008 |
| D16 | spec O13 before O14, §F R1–R15 in order; O10 replaced by re-measured, attributed counts (L17) | spec §A.2, §F |
| D17 | every `Covers:` line begins with its AC id | acceptance.md |
| D18 | plan D1: the locked handle offers `LoadPure`, no adopting read; spec REQ-FAL-009 names the non-adopting reads; `TestLockedBacklogLoadIsPure` in AC-FAL-011 | plan D1; AC-FAL-011 |

Decisions for the leader are in spec §H (DL-1 bare-form outcome, DL-2 nominated-form outcome, DL-3
cross-process helper). Not observed in this iteration, stated so nobody reads silence as a pass: the
mutants (none can run before implementation); any behavior after the fix; Windows; a multi-lane stall;
whether anything issues the request that applies a lease expiry to a card leased without a worktree.

### 2026-10-03 — iteration 3 (override round) repair (spec.md 0.3.0)

Defect ids below are the plan-audit's iteration-2 ids, written `PA2-…` so they do not collide with the
card's symptom ids M1–M5. Observations are ledger rows L19–L23 in `acceptance.md`, made on HEAD
`c8b716fed24564a685188dc94b3f446ac9fc79c8` (Go files equal to the plan-start tree) plus this round's
uncommitted evidence files; the installed `moai` build is `0732cc699`, diverged from HEAD (neither is an
ancestor of the other) with an identical `internal/spec` (L22). No Go source, other SPEC, doctrine file or
`.moai/reports/` file was edited.

| Finding | Where it was fixed | How a re-auditor checks it |
|---|---|---|
| PA2-M1 (drift-log lock wait is unbounded) | spec §A.2 O15 (source reading + L19), REQ-FAL-006 (time bound includes it), REQ-FAL-007 (the named, bounded exception), new REQ-FAL-014, REQ-FAL-009 and -012 (carve-out and scope), §E (REQ-FR-025 narrowed), §F R16, §H DL-5/DL-7; plan D2 (third device, bounded flow, lock order), WM1 stub, WM4 (the milestone that changes `homestate`), §6 R-K/R-L, §7 MU15–MU17; acceptance AC-FAL-015 (new, release-blocking), AC-FAL-009 (iii), L19, L21, L23 | read REQ-FAL-014 and AC-FAL-015; re-run L19's command (red: elapsed 3.0 s against a 1.3 s limit, skip predicate red, retry green) |
| PA2-M2 (AC-FAL-009 table) | acceptance AC-FAL-009 (ii): table rebuilt with queue-item and row state per row, `queued`/`picked` split for the assigned-row nominee, two failed-claim rows (before and after the assign edge), fixture-builder column | compare each row with `factoryNextClaim`, `factoryNextRecordAndClaim`, `RecordPicked`, `factoryNominateCompensate` (factory_card.go 604–650, 972–1003; card_picked.go 114–177) |
| PA2-M3 (REQ-FAL-003 vs REQ-FAL-009, arm (a)) | spec REQ-FAL-003 second clause narrowed, REQ-FAL-009 (arm (a) preserved), §A.1 M2 row, §B.2, §F R17, §D (arm (a) out of scope), §H DL-6; acceptance AC-FAL-003 clause (iii) (guard) and L20; plan MU18 | read REQ-FAL-003; L20 (`queue=hold record=leased`, 3 of 3) |
| PA2-N1 (DL-1 reasoning) | spec §H DL-1 restated (visible error vs silent stop; both loops end on an error), §F R5, §D (lane-loop retry out of scope) | read DL-1 and R5 against `factory_lane_relaunch.go` 76–79, 103–109 and `codex_launcher.go` 988–994 |
| PA2-N2 (`raced` supersession) | spec §E new row; acceptance AC-FAL-014 | read the §E row; AC-FAL-014 Then |
| PA2-N3 (legacy adoption) | spec §F R13 and REQ-FAL-009 (store construction, before any lock; the race named); plan D1 (the store is constructed before `WithLock`) | read R13 against `todo.go` and `state_dir.go` |
| PA2-N4 (AC-FAL-010 vs plan §5) | acceptance AC-FAL-010 Then (the one recorded removal is excepted); plan §5 intro | read both |
| PA2-N5 (plan §7) | plan §7: MU8 → AC-FAL-011; MU12, MU13, MU14 added (the audit's two AC-only mutants and N14(iii)); MU15–MU18 added for this round | read the §7 table against the acceptance Mutation lines |
| PA2-N6 (absolute sentences) | spec §A.1 M5 row; acceptance L13 text | read both sentences against the re-execution counts they cite |
| PA2-N7 (panic stubs) | plan WM1 (stubs return a zero value or a sentinel error; `release` never nil; homestate marker is the identity); acceptance AC-FAL-008 cell | read plan WM1 step 1 |
| PA2-N8 (D5 not in §H) | spec §H DL-4 (open, non-blocking; now three SPECs); plan D5 | read §H |
| PA2-N9 (shortened wait vs constant) | plan D3 (`factoryWorktreeStepWaitDefault` const + `factoryWorktreeStepWait` var), WM1; acceptance AC-FAL-006 | read plan D3 |
| PA2-N10, PA2-N11 (AC-FAL-014) | acceptance AC-FAL-014: Mutation line added; Command gains `prior_completed_sha` and `status:` counts | read AC-FAL-014 |
| PA2-N12 (probe ranges) | spec §A.2 O6, O14; acceptance L7, L14 (ranges stated as seen over named runs) | read the sentences |
| PA2-N13 (REQ-FAL-007 "read") | spec REQ-FAL-007 ("file or process I/O", environment and clock excepted); acceptance AC-FAL-009 (iii) | read both |
| PA2-N14 (mutant holes) | only (iii) taken: plan §7 MU14; holes (i) and (ii) have no criterion and are named as not claimed (plan §7 note) — skipped because each needs a new criterion clause and fixture, not a one-line edit | read plan §7 note |
| PA2-N15 ("only" unmeasured) | acceptance AC-FAL-013: allowed set enumerated, a complement probe added, sync-phase paths stated as outside it | read AC-FAL-013 |
| PA2-N16 (volume) | nothing asked; the round added L19–L23, REQ-FAL-014 and AC-FAL-015 | — |

Observations made in this round (commands and verbatim output in the ledger): L19 — drift-log lock held
3 s, claim under an 800 ms deadline: 3.016 s, 3.009 s, 3.002 s (exit 1, red for the stated reason); the
uncontended reconciliation costs 0.002–0.028 s (1 entry), 0.018–0.100 s (200), 0.151–0.260 s (2000). L20 —
arm (a) leases a held card with an assigned row, 3 of 3. L21 — the four new test names sweep 0 at the pin.
L22 — tool provenance. L23 — the six existing drift-log tests pass (6 of 6). Also: `grep -c
SPEC-FACTORY-ATOMIC-LEASE-001 .moai/specs/SPEC-FACTORY-RECORD-001/spec.md` printed `0`, exit 1.

Decisions taken in this round that the leader should see: spec §H DL-7 (non-blocking try rather than a
deadline; opt-in per call for the claim's three writes rather than global; both left open by the leader's
design and chosen to change the fewest other callers' behavior); the new third completed SPEC
(SPEC-FACTORY-RECORD-001, REQ-FR-025's "next successful write reconciles", narrowed for the claim's
writes) that the Amendments mechanism now reopens — DL-4's cost grew from two SPECs to three.

Not observed in this round, stated so nobody reads silence as a pass: any behavior after the fix; the
lease-level tests of AC-FAL-015 (they arrive in WM1; L19 is the record-write-level probe); the mutants
(none can run before implementation); Windows (the non-waiting form is compile-verified only once built);
a drift log larger than 2000 entries; whether another process's log append is delayed in practice by the
held lock; the `moai` build from this tree (the installed `0732cc699` judged the lint); the plan-audit's
own L13 and L14 re-executions that §A.1 and L13/L14 cite (they are reported to this author, not in the
ledger).

### 2026-10-03 — exception repair (iteration 4) (spec.md 0.3.1)

Scope: exactly the four hunks of the leader's "Decision 2" (the plan-audit's iteration-3 ids I3-M1 and
I3-S1 to I3-S3; the notes I3-N1 to I3-N5 were left alone, they are optional). No Go source, other SPEC,
doctrine file or `.moai/reports/` file was edited; the HEAD this repair read and measured is
`9f73f4cdf7af24af493edfb9e629f70aac915133` (Go files equal the plan-start tree). Tool provenance for
`moai spec lint` (`verification-claim-integrity.md` §2.2): the installed build is `0732cc699`
(`v3.2.0-rc.27`, built 2026-10-03T03:34:50Z, from `moai version`); `git merge-base --is-ancestor 0732cc699
HEAD` and `git merge-base --is-ancestor HEAD 0732cc699` each exited 1, so the build is not a strict ancestor
of HEAD, and `git diff --name-only HEAD 0732cc699 -- internal/spec` printed nothing, so the lint source it
carries equals the tree's; the tree's own build was not made. With that build, `moai spec lint
SPEC-FACTORY-ATOMIC-LEASE-001` and the same with `--strict` each printed `✓ No findings — all SPEC
documents are valid` and exited 0 after the edits above.

| Finding | Where it was fixed | How a re-reader checks it |
|---|---|---|
| I3-M1 (AC-FAL-015 (i)–(iii) unsatisfiable at the verb) | acceptance AC-FAL-015: a new "Observation points" paragraph, five fixtures (a)–(e) each naming its level, clauses (i)–(iii) restated at the return of `factoryNextNominate` / `factoryNextLeaseOnceGated` with "exit 0 / the verb's output" expressly not observed there, new clause (vii) (the verb's own card-worktree record write waits for the held lock and then reconciles, with probe B's measured expectation), Command, RED-now, green-path, mutation and Not-claimed cells updated; ledger L24; plan §7 MU20, WM1 test list | read AC-FAL-015 "Observation points" and clauses (i)–(iii), (vii) against `factory_card.go` 1083–1112 and 407–430 (the verb calls `factoryEnsureCardWorktree` only after a lease function returned, and that step ends in `db.RecordCardWorktree`); re-run L24's first command shape (the probe is a scratch file, see L24) |
| I3-S1 ("no entry reconciled twice") | spec REQ-FAL-014 (the over-claim replaced by the re-read guarantee); spec §F R16 (the cross-writer window named, audit-measured 2 duplicate events in 3 of 3 runs); acceptance AC-FAL-015 fixture (e) and clause (vi); plan §7 MU19, D2 step 3, WM1 seam and test list | read REQ-FAL-014's third clause and R16's "second window"; AC-FAL-015 (vi) and MU19 |
| I3-S2 (REQ-FAL-009 carve-outs) | spec REQ-FAL-009: the lead-in no longer names the skip; the four carve-outs are a numbered list | read REQ-FAL-009; each of the four items equals one of the four things the audit listed |
| I3-S3 (DL-6 provenance) | spec §H: DL-6 and DL-7 under "Decided (leader, 2026-10-03)" with the provenance "decision.md Decision 2"; the "Taken by the repair author" heading removed (it would be empty); DL-7 carries the leader's added instruction; spec §F R18 added for the limit | read §H DL-6, DL-7 and R18; DL-4 is still open |

Mechanical consequences of those four, listed so none is a surprise: spec HISTORY 0.3.1 and the frontmatter
`version`; spec §B.5 and the §F introduction mention R18; acceptance AC-FAL-012 counts R1–R18; acceptance
S3 and the ledger header name L24; plan.md's header note and §6 R-K. The counts stay at 14 requirements and
15 criteria (Tier M ceilings 16 and 16); the mutants are now MU1–MU20.

Observations made in this repair (commands and verbatim output are ledger row L24): the card-worktree
record write, with one unreconciled entry and the drift log's lock held 2 s, returned after 2.0007 s,
2.0015 s and 2.0006 s with `unreconciled-after=0` (3 of 3; the plan-audit's own probe B printed 2.004 s,
2.001 s and 2.003 s — audit-measured, not a ledger row); the two test names added in this repair
(`TestRecordWriteReconcileBoundedRereadsUnderLock`, `TestFactoryLeaseDriftLogVerbWorktreeWriteWaits`)
sweep 0 tests at the pin. The probe is a scratch file outside the repository (`evidence/` was outside the
file list of this repair); the exit status of both `go test` runs was not captured (the worktree guard
refuses a trailing `echo`), and the rows say so.

Decisions taken in this repair beyond the leader's record, for the leader to see: (1) the test seam of
AC-FAL-015 fixture (e) — an inert package variable between the log's unlocked read and the claim's try for
its lock — is a new line item of the WM1 seam-and-stub commit, because no deterministic test of the
re-read step exists without one; (2) the cross-writer window is named in R16 and the leader's DL-7 limit
is a separate R18, not folded into R16, so each reads as one residual; (3) new clauses (vi) and (vii)
were appended after (v) so no existing clause number, and no cross-reference to one, changed; (4) two
mutants were added, MU19 (no re-read under the lock) and MU20 (the skip reaching the verb's worktree write),
the second being the "connection-scoped cousin" of MU16 the audit named; (5) the verb-level fixture holds
the lock 2 s, the value of the measured probe, and bounds the verb's return by that hold plus the same
500 ms margin as AC-FAL-007, a labeled heuristic.

Not observed in this repair, stated so nobody reads silence as a pass: any test of AC-FAL-015 clauses (i)–(iii),
(vi) or (vii) (they arrive in WM1; clause (vi) has no RED-now observation and none is claimed — its test needs
the seam); the overlap of two live writers in the narrow gap between a commit and its mark (only the audit's
state-forcing probe A exists); the mutants; Windows; any behavior after the fix; the verb itself under a held
drift-log lock (L24 is record level); the four-hunk confirmation by a plan-auditor.

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_

## §F Phase 4 Mode Selection

Input parameters: tier M; scope about 12 Go files in `internal/cli`, `internal/kanban`, `internal/homestate` plus their tests;
domain count 3 (CLI lease verbs, queue store, factory record); file mix 100% Go; concurrency benefit LOW (coding-heavy, ordered
milestones with a shared file `internal/cli/factory_card.go`); Agent Teams prerequisites not applicable (not requested).

| Mode | Selected | Rationale |
|---|---|---|
| direct | not selected | semantic change across three packages |
| serial | **selected** | coding-heavy, milestones depend on each other (WM2 primitive before WM3, WM3 with WM4) |
| fanout | not selected | write-capable work in one tree; one writer per working tree |
| sweep | not selected | not a uniform mechanical transform |

Decision: serial

Justification: the milestones share one tree and one hot file; a single write-capable agent runs at a time. The orchestrator
verifies each round (own `git status`, HEAD, scoped test runs) before the next spawn. Rounds: R1 = WM1 (seam-and-stub commit,
baseline commit, RED commit); R2 = WM2; R3 = WM3 + WM4 (they land together); R4 = WM5; R5 = WM6 closure. The run starts after
the autonomous plan->run Kickoff recorded in `.moai/reports/t1458/decision.md` (git-ignored) at HEAD b27652922, and after
absorbing develop 2b9e4a4d0 (merge 09faf2965; the plan-artifact hashes were re-checked unchanged).
