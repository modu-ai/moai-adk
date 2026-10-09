# progress.md — SPEC-MEMORY-FOLD-RENAME-RACE-001

card: t1568 · phase: plan · tier: M · baseline tree: `2aab5f797` (base absorbed from `81786284e`, lane decision D-1)

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-09
- revision: 0.1.2 (audit iter-1 fix round — D1 selector corrected + re-measured baseline, D2 span criterion AC-MRR-009, D3 abandonment-release cell, D4 honest RED premise, D5 windows CI judge; HISTORY v0.1.2; the v0.1.1 base-absorb re-pin record is retained below)
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

- judging build (current): `/tmp/moai-t1568-lint2`, built from THIS tree at HEAD `2aab5f797` (`go build -o /tmp/moai-t1568-lint2 ./cmd/moai`), invoked by path. Re-run after the revision 0.1.2 fix-round edits: `spec lint SPEC-MEMORY-FOLD-RENAME-RACE-001` → `✓ No findings — all SPEC documents are valid`, exit 0 — the run of record for 0.1.2.
- Historical (old base `81786284e`, binary built from that tree): run 1 `0 error(s), 11 warning(s)` (unanchored `-run` patterns), run 2 after anchoring `✓ No findings` — both superseded by the run above; kept for the record only.
- Anchoring note: the family selector is `-run '^(TestMemoryFold|TestReviewArchiveUpdate).*$'` — end-anchored via `.*$`, selecting exactly the fold regression family: 27 `TestMemoryFold*` tests (including the 13 `TestMemoryFoldOnDone_*` wiring tests and the write-ordering regression `TestMemoryFold_ArchiveRecheckedBeforeMemoryRename`) plus `TestReviewArchiveUpdateDuringEffectiveScan`. Audit iter-1 (D1) established the prior selector `^Test(Fold|Review).*$` swept 53 Review-prefix tests and ZERO fold tests; the corrected selector's set is verified by the plan §C.1 `-list` gate before every use: `go test -list '^(TestMemoryFold|TestReviewArchiveUpdate).*$' ./internal/cli` → 28 tests, exit 0 (verbatim list persisted at `.moai/reports/t1568/t1568-family-list2.txt`).

### Pre-fix baseline (green-before for AC-MRR-007)

- Attempt 1 (base `81786284e`, `-timeout 240s`): **timeout artifact, not a red** — verbatim: `panic: test timed out after 4m0s` (log line 15), `FAIL github.com/modu-ai/moai-adk/internal/cli 244.760s`, and ZERO `--- FAIL` test-level lines in the full log (grep-verified). The goroutine dump's parked test (`TestReviewUnrecordedPickedHubWait` in the `homestate` busy-retry path) is the dump's illustration, not a failed assertion. Claim policy: this run proves nothing about the family's health in either direction.
- Attempt 2 (base `2aab5f797`, wrong selector `^Test(Fold|Review).*$`): **mis-measurement, kept on record as such** — verbatim `ok  	github.com/modu-ai/moai-adk/internal/cli	414.385s`, exit 0, but the swept set was 53 Review-prefix tests and ZERO of the 27 `TestMemoryFold*` fold tests (audit iter-1 D1). Never cite it as the green-before.
- Attempt 3 (base `2aab5f797`, CORRECTED selector `^(TestMemoryFold|TestReviewArchiveUpdate).*$`, `-timeout 900s`, env-scrubbed compound): **GREEN — the green-before of record** — verbatim `ok  	github.com/modu-ai/moai-adk/internal/cli	38.008s`, exit 0 (own task output `EXIT=0`), this run / this tree. Logs persisted in-tree: `.moai/reports/t1568/t1568-baseline3.txt` (plus attempts 1-2 and both `-list` verifications). AC-MRR-007's post-fix green must equal this run.

### RED-now discipline status

- AC-MRR-001's RED cell is PENDING EXECUTION **by choice, not by impossibility** (audit iter-1 D4 corrected the premise): the existing `orderProbe("bytes-done")` seam (`memory_fold.go:643-645`) fires inside the window, so a plan-phase RED is executable today by setting it per plan.md B-3's set/restore discipline. The deferral to run-phase M1 is the clean-dedicated-seam choice: `orderProbe`'s contract is stage recording (owned by the ordering regression), not mutation; the dedicated `mutateBeforeRename` seam keeps the mutation contract separate and fires after the abandonment check — the true last observable moment. The observation itself is M1's first deliverable and its verbatim output + exit code + tree SHA land in progress.md §E.2; M2 may not start before it is on record. Every executable plan-phase check (ID regex PASS, uniqueness, chokepoint grep, lint, corrected-selector baseline) is on record in `.moai/reports/t1568/plan-evidence.md`.

## §E.2 Run-phase Evidence

_<pending run-phase — owned by manager-develop>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — owned by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs; sync_commit_sha populated by the single sync commit>_
