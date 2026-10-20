# t1912 narrow re-verification (관측 시각 미상 — 2026-10-20 경과 후 기록, lane-11)

Card t1912, run tn564n. Tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1865 (WT-update-ownership-repair). HEAD at verification: 9764613c59, tree clean.

## ① Narrow re-verification — PASS (this run, this HEAD)

Command: `go test -timeout 30m ./internal/cli -run '^(TestUnregisteredSymlinkOutsideProjectIsNotRead|TestPathBoundaryCounterexamples|TestNonRegularFilesAreKeptWithoutBeingRead|TestRegularFileReplacedByFIFOAfterLstatIsKeptWithoutBlocking)$' -count=1 -v`
Observed: all 4 F1/F2 tests PASS (incl. 4 MutantSymlink/MutantHung/MutantReadFailed/MutantFifo subtests of TestNonRegularFilesAreKeptWithoutBeingRead), `ok github.com/modu-ai/moai-adk/internal/cli 7.114s`, NARROW_EXIT:0.
Prior GREEN (4.738s @ 563745ea66) carried forward — this run confirms the fix survives at 9764613c59.

## ② Package timeout — CONFIRMED from records (no re-run)

`go test -timeout 30m ./internal/cli/...` FAIL at 1801.588s (timeout); subpackages all ok; root cause unproven (deadlock/external load/git-child wait — lane-58's package-timeout-investigation.md). Leader ruling stands: re-runs coordinated one-by-one by the leader after existing load and lane-59 verification end. NOT re-run here.

## ③ Formal audit refusal/hold — IDENTIFIED (verbatim basis)

plan-audit-iter4-delta.md: iteration 3 FAIL 0.84 vs threshold 0.85, ceiling reached (3), hold recorded; iteration 4 delta PASS-WITH-DEBT 0.92 (codex required pass) under leader rulings d-20261018T044408Z-3e9d + standing 9e76, scope limited to commit 073b3e1382. 21 disclosed debts (N4/N14–N21 dispose_in=run; N16 = card t1702 hold scope, not bypassed). The formal-audit refusal/hold = ceiling not extended; no bypass performed.

## Dispatch status: ①②③ complete — reporting to the leader.

## Ownership/tree convention record (leader reinforcement ②, 2026-10-20T10:4xZ)

- Card t1912 = handover-verification card on the EXISTING t1865 tree @ 9764613c59. No source changes made by lane-11; new source changes await leader coordination.
- Original commit history preserved (git log lineage, verified this run): … 4a625be156 → 563745ea66 (final sync-audit FAIL + hold record) → **aeab9f2b0b (F1/F2 repair — guard ownership reads after link and fifo races; PRESERVED, not overwritten)** → 3b1a032fa2 (preserved audit refusal clarification) → 9764613c59 (CLI timeout investigation).
- The aeab9f2b0b repair commit and .moai/reports/t1865/f1-f2-correction.md stand as the F1/F2 evidence of record — this card did NOT re-run their verification as "unrepaired" and did not modify them.
- Baseline HEAD attachment: every measurement in this record is pinned to 9764613c59 (narrow re-verification ①) and to the recorded evidence for ②③ (3b1a032fa2-era records + factory record).
