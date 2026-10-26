verdict: PASS
audited_sha: 139c7bb2f89af3bd8f268b6330dcf5c6d2fd8e91

# t1973 — lane completion verdict (Class B, reproduce-first)

- card: t1973 · run tnbn7e · lane-14
- tree: /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1973 · branch WT-factory-run-rebind · **head 139c7bb2f8** (full 139c7bb2f89af3bd8f268b6330dcf5c6d2fd8e91, parent ab107885ed, absorb parent line via 57c22079c9 → origin/main 840053635f) · clean
- scope of change: internal/homestate/factory_run_rebind.go (+ repair), internal/homestate/factory_run_rebind_t1973_test.go (new, reproduction tests), fixture-touching test files per diagnostics; no `.github/workflows` changes
- defect provenance: t1966 sync-audit-verdict-iter5 §Out-of-scope (codex leg rcpt-ee107d9287, never auditor-reproduced) — this card reproduced both findings against the real `FactoryDB.RebindRun` (5/5 subtests RED at 57c22079c9)

## Claim

P1 (stale-done target row eats the source run's live work) and P2 (a third run's living worker / valid lease is ignored when the leader is dead with a matching stamp) are both reproduced (RED), then repaired: the consolidation now refuses a terminal or evidence-divergent target row and refuses a third run that holds anything live.

## Evidence (command + verbatim output)

- RED (delegate, at 57c22079c9, tests-only commit ab107885ed): `unset GOFLAGS && go test ./internal/homestate/ -run 'TestRed_T1973' -count=1` → exit 1 — P1: `--- FAIL: TestRed_T1973_P1_StaleDoneTargetEatsLiveWork` (3 subtests: terminal_done_target_row_with_its_own_evidence_and_approval / terminal_target_row_even_with_matching_evidence / non-terminal_target_row_with_divergent_evidence — each "operation succeeded, want a refusal naming [c1 run-new]"); P2: `--- FAIL: TestRed_T1973_P2_LivingWorkerAndLeaseInThirdRun` (2 subtests: valid_lease_on_the_dead-leader_third_run's_row / living_worker_registered_on_the_dead-leader_third_run — each "operation succeeded, want a refusal naming [c1 run-third]"). Full verbatim in .moai/reports/t1973/progress.md.
- GREEN (delegate, at 139c7bb2f8): targeted rebind-adjacent suite → `ok github.com/modu-ai/moai-adk/internal/homestate 15.056s` exit 0; full package `unset GOFLAGS && go test ./internal/homestate/... -count=1` → `ok github.com/modu-ai/moai-adk/internal/homestate 174.977s` exit 0; `gofmt -l` clean on touched files; `go vet ./internal/homestate/` PASS.
- card-review (codex_review scope=card, base 840053635f): verdict pass, findings none → `.moai/reports/t1973/card-review.md`.

## Repair summary

- P1: the `held > 0` unconditional abandon is replaced by a `SELECT state, evidence_sha` of the target row — a terminal target row OR an evidence-divergent one returns a rebindRefusal naming card, run, state and both evidence SHAs (whole transaction rolls back; the source's live row and run are untouched). Only a non-terminal, evidence-matching target row keeps the existing abandon-with-pointer path (t1963 behavior preserved — rbBase c4 and lceSeedChain c2 keep it).
- P2: refuseThirdRunHolder's dead-owner skip now proves the third run holds nothing live before applying — the holder row's lease via the package's holdsValidLease (newly selected columns), the third run's workers via workerPIDs + the caller's workerProbe (the same probe refuseLiveWorker uses). A valid lease, a worker not positively dead (indeterminate counts as live), or an unreadable worker listing refuses the consolidation. Signature gained `now, probe` params; single caller updated.

## Baseline-attribution

Every figure above was measured in this run against this tree: delegate runs at head 139c7bb2f8; lane observation of the RED commit and repair commit at their landing SHAs; HEAD re-read immediately before this verdict (`git rev-parse` → 139c7bb2f89af3bd8f268b6330dcf5c6d2fd8e91).

## t1975 overlap (record-only)

factory_run_rebind.go:215 (step-5 card_dispatch repoint) is untouched by this card; the P1 refusal prevents the terminal-target shape from reaching it, reducing — not replacing — t1975's concern there. leader_approval.go:295 (upsertDispatchBindingTx) is a different writer of the same binding table; no interaction beyond sharing the table.

## Gaps

- Reproductions use the suite's injected classifier/probe seams — no live-process reproduction was attempted.
- Coverage percentage not measured this run; all new refusal and skip-preserving branches are exercised by the five new subtests plus the pre-existing dead-third-run test.

## Residual-risk

- The new refusal strings could interact with string-matching tests elsewhere; the full homestate package regression (174.977s, exit 0) covers the package, and origin/main CI owns the repository-wide verdict after landing.
