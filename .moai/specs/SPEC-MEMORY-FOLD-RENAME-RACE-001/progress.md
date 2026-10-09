# progress.md — SPEC-MEMORY-FOLD-RENAME-RACE-001

card: t1568 · phase: plan · tier: M · baseline tree: `2aab5f797` (base absorbed from `81786284e`, lane decision D-1)

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-09
- revision: 0.1.3 (audit iter-2 fix round — D6 Windows judge corrected to the release multi-OS gate's Windows leg + DoD timing recorded, D7 AC-MRR-009 extended to the 4-tuple with the between-writes seam, D8 P1 premise reworded, A4 selector 28→29 with re-measured baseline; HISTORY v0.1.3; the v0.1.2 iter-1 fix round and v0.1.1 base-absorb re-pin records are retained below)
- tier: M
- artifacts: spec.md + plan.md + acceptance.md (Tier M set) + design.md (lane-directed mechanism-rationale record) + decision-index.md (conditional — `interview.decision_gate: on`) + progress.md
- req_count: 7 (REQ-MRR-001…007, gap-free, GEARS notation)
- ac_count: 9 (AC-MRR-001…009, gap-free; traceability table at acceptance.md §2 — every REQ covered, no orphan criterion)
- baseline_tree: 2aab5f797
- certain_file_count: 6 (Tier M band 5-15; the added file is `fold_store_lock_windows_test.go`, AC-MRR-008's Windows judge; two named contingencies in plan.md §A.2)
- development_mode: tdd (RED-first is the card's mandate and plan.md M1)
- card_id: t1568
- decision_index: 5 rows, all `FOUNDER`/implementation-level, all `Default:`-carrying, all stamped `DEFAULT-APPLIED 2026-10-09T02:43:00Z glm via manager-spec` — no operator-blocking row
- re-pin record: +2 below the import block, +3 after the `foldClosedCardMemory` hunk (`os.Rename` `:647`→`:649`, `applyFold` `:412`→`:414`, `foldOnDoneStep` caller `:1172`→`:1175`); `memory_fold_test.go` and `internal/sessionmsg/lock_*.go` verified unchanged (`git diff 81786284e 2aab5f797 --stat` empty)

### Plan-phase lint (tool provenance: judging build vs measured tree)

- judging build (current): `/tmp/moai-t1568-lint2`, built from THIS tree at HEAD `2aab5f797` (`go build -o /tmp/moai-t1568-lint2 ./cmd/moai`), invoked by path. Re-run after the revision 0.1.3 fix-round edits: `spec lint SPEC-MEMORY-FOLD-RENAME-RACE-001` → `✓ No findings — all SPEC documents are valid`, exit 0 — the run of record for 0.1.3.
- Historical (old base `81786284e`, binary built from that tree): run 1 `0 error(s), 11 warning(s)` (unanchored `-run` patterns), run 2 after anchoring `✓ No findings` — both superseded by the run above; kept for the record only.
- Anchoring note: the family selector is `-run '^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$'` — end-anchored via `.*$`, selecting exactly the fold regression family: 27 `TestMemoryFold*` tests (including the 13 `TestMemoryFoldOnDone_*` wiring tests and the write-ordering regression `TestMemoryFold_ArchiveRecheckedBeforeMemoryRename`), `TestReviewArchiveUpdateDuringEffectiveScan`, and `TestReviewSequentialAbandonedTempOwnership` (added at iter-2 A4 — the codex-review round-2 per-worker temp-ownership regression the 28-test form missed). Selector history: iter-1 (D1) established the first selector `^Test(Fold|Review).*$` swept 53 Review-prefix tests and ZERO fold tests; the iter-2 28-test form superseded it; the final form's set is verified by the plan §C.1 `-list` gate before every use: `go test -list '^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$' ./internal/cli` → 29 tests, exit 0 (verbatim list persisted at `.moai/reports/t1568/t1568-family-list3.txt`).

### Pre-fix baseline (green-before for AC-MRR-007)

- Attempt 1 (base `81786284e`, `-timeout 240s`): **timeout artifact, not a red** — verbatim: `panic: test timed out after 4m0s` (log line 15), `FAIL github.com/modu-ai/moai-adk/internal/cli 244.760s`, and ZERO `--- FAIL` test-level lines in the full log (grep-verified). The goroutine dump's parked test (`TestReviewUnrecordedPickedHubWait` in the `homestate` busy-retry path) is the dump's illustration, not a failed assertion. Claim policy: this run proves nothing about the family's health in either direction.
- Attempt 2 (base `2aab5f797`, wrong selector `^Test(Fold|Review).*$`): **mis-measurement, kept on record as such** — verbatim `ok  	github.com/modu-ai/moai-adk/internal/cli	414.385s`, exit 0, but the swept set was 53 Review-prefix tests and ZERO of the 27 `TestMemoryFold*` fold tests (audit iter-1 D1). Never cite it as the green-before.
- Attempt 3 (base `2aab5f797`, 28-test selector `^(TestMemoryFold|TestReviewArchiveUpdate).*$`, `-timeout 900s`, env-scrubbed compound): **GREEN, then superseded as the record** — verbatim `ok  	github.com/modu-ai/moai-adk/internal/cli	38.008s`, exit 0; log persisted at `.moai/reports/t1568/t1568-baseline3.txt`. Kept honest: iter-2 A4 established this selector missed `TestReviewSequentialAbandonedTempOwnership`, so its command is no longer AC-MRR-007's exact command.
- Attempt 4 (base `2aab5f797`, FINAL 29-test selector `^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$`, `-timeout 900s`, env-scrubbed compound): **GREEN — the green-before of record** — verbatim `ok  	github.com/modu-ai/moai-adk/internal/cli	42.974s`, exit 0 (own task output `EXIT=0`), this run / this tree. Logs persisted in-tree: `.moai/reports/t1568/t1568-baseline4.txt` (plus attempts 1-3 and all three `-list` verifications). AC-MRR-007's post-fix green must equal this run.

### RED-now discipline status

- AC-MRR-001's RED cell is PENDING EXECUTION **by choice, not by impossibility** (audit iter-1 D4 corrected the premise): the existing `orderProbe("bytes-done")` seam (`memory_fold.go:643-645`) fires inside the window, so a plan-phase RED is executable today by setting it per plan.md B-3's set/restore discipline. The deferral to run-phase M1 is the clean-dedicated-seam choice: `orderProbe`'s contract is stage recording (owned by the ordering regression), not mutation; the dedicated `mutateBeforeRename` seam keeps the mutation contract separate and fires after the abandonment check — the true last observable moment. The observation itself is M1's first deliverable and its verbatim output + exit code + tree SHA land in progress.md §E.2; M2 may not start before it is on record. Every executable plan-phase check (ID regex PASS, uniqueness, chokepoint grep, lint, corrected-selector baseline) is on record in `.moai/reports/t1568/plan-evidence.md`.

## §F Phase 4 Mode Selection

- Inputs: tier M · scope 6 files · domains 1 (Go, internal/cli) · language mix Go-only · concurrency benefit LOW (coding-heavy, strict RED→GREEN ordering)
- Evaluation: direct — not selected (multi-file, test-first loop with milestones) · fanout — not selected (coding-heavy per the Anthropic coding-task parallelism caveat) · sweep — not selected (not mechanical-uniform) · agent-team — not requested
- Decision: **serial**
- Justification: single-package defect fix whose milestones are strictly ordered (M1 RED must be observed before M2 GREEN; plan.md §F). One writer per tree; sequential per-milestone delegation is the safe default for coding work.
- Baseline-attribution: decided by lane-22 at run entry, plan-audit iter-3 PASS (0.95, rcpt-d7cf366e7f8d3eb44edd6dae) at HEAD 2f436db17.

## §E.2 Run-phase Evidence

### M1 — RED: the loss reproduced (AC-MRR-001's RED cell, observed before any fix code)

- **Observed at**: 2026-10-09, on the M1 work tree = HEAD `2f436db17` + the M1 test-only seam fields and test only (production `atomicWriteFoldFile`/`applyFold` checks unchanged; the seam call sites are additive). This exact tree is snapshotted by the M1 commit.
- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -race -count=10 ./internal/cli -run '^TestFoldRenameWindowConcurrentWriter$'`
- **Exit code**: 1
- **Verbatim** (every iteration red — 10/10 top-level FAIL, 20/20 subtests FAIL; full raw log `.moai/reports/t1568/t1568-red-m1.txt`):

```
--- FAIL: TestFoldRenameWindowConcurrentWriter (0.09s)
    --- FAIL: TestFoldRenameWindowConcurrentWriter/archive-append (0.01s)
        memory_fold_test.go:1466: REQ-MRR-004 violated: the fold returned success while the writer's line published inside the archive-append rename window was destroyed from project_card_archive_2026_10.md
    --- FAIL: TestFoldRenameWindowConcurrentWriter/memory-rewrite (0.00s)
        memory_fold_test.go:1466: REQ-MRR-004 violated: the fold returned success while the writer's line published inside the memory-rewrite rename window was destroyed from MEMORY.md
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.220s
```

  (the `--- FAIL`/`REQ-MRR-004 violated` pair repeats identically for all 10 iterations; the block above is iteration 1 verbatim)
- **Why red for the right reason**: the fold returned success (`runFoldOK` passed) while the writer's line published at the `mutateBeforeRename` point — after every pre-rename check — was destroyed by `os.Rename`, on BOTH write surfaces (archive append; MEMORY.md rewrite). The first attempt's red (`.moai/reports/t1568/t1568-red-m1-first-fixture.txt`) was a fixture defect — `no archive index to fold into` (the store lacked the archive file) — corrected by seeding `minimalArchive()`; it is kept on record as a wrong-reason red, never cited as the AC-MRR-001 RED.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owned by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs; sync_commit_sha populated by the single sync commit>_
