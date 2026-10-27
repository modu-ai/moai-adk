verdict: PASS
audited_sha: 27e248839fecc41b91d7211d8e5bd8c6b35700d6

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

---

## Round 2 — 2026-10-27 (leader codex_audit FAIL on 139c7bb2f8, repaired)

- audit receipt: rcpt-5120c225117dd37429a99ea0 @26b6b502b9 (static review, review_base main / merge base cb30e05b3e) — verdict FAIL, both findings inside the RebindRun path repaired in round 1; in-card repairs.
- commits this round: `3c2c65caf4` (RED tests, failing at tree 26b6b502b9) → `a980943254` (repair). No push, no absorb.

### Finding P1 — source card's own valid lease unchecked (factory_run_rebind.go:165)

- RED: `TestRed_T1973b_P1_SourceCardHoldingValidLeaseIsRekeyed` (2 subtests: valid lease with no source worker row; holder registered as a worker of a different run) — command `unset GOFLAGS && go test ./internal/homestate/ -run 'TestRed_T1973b' -count=1` at tree 26b6b502b9 → exit 1, each subtest "operation succeeded, want a refusal naming [c1 run-old]".
- Repair: `consolidationCard` carries `lease_holder`/`lease_expires_at` (sourceCards SELECT extended), and the pre-write guard loop refuses a source card holding a valid (unexpired) lease via the package's existing `holdsValidLease` — the leader's death does not kill a card lease, and the worker guard is run-scoped so a holder registered elsewhere stays invisible to `refuseLiveWorker`.

### Finding P2 — both-empty evidence read as a match (factory_run_rebind.go:193)

- RED: `TestRed_T1973b_P2_BothEmptyEvidenceTreatedAsMatch` — same command/run → exit 1, "operation succeeded, want a refusal naming [c1 run-new]".
- Repair: the target-row conflict check now refuses the both-empty evidence shape as unproven-same-work. Choice made (and stated): the SIMPLE SAFE DEFAULT (refuse), not progress-info comparison — the refuse-side cost is one manual re-approval, while the abandon-side cost is the source row's stranded stage/worktree/spec_id. Named in the refusal via `bothEmptyNote`.

### Round-2 evidence

- GREEN (all round-1 + round-2 t1973 tests, at a980943254 working tree): `unset GOFLAGS && go test ./internal/homestate/ -run 'TestRed_T1973' -count=1` → exit 0, `ok github.com/modu-ai/moai-adk/internal/homestate 2.296s`; `gofmt -l` clean; `go vet ./internal/homestate/` PASS.
- Full package regression: `unset GOFLAGS && go test ./internal/homestate/... -count=1` → exit 0, `ok github.com/modu-ai/moai-adk/internal/homestate 246.491s`.
- Baseline: all measurements in this run against this tree; HEAD re-read immediately before this update (`git rev-parse` → a980943254276978c84bdaedf3cd073c317014e2).

## Rebinding note (2026-10-27) — audited_sha rebound to the post-absorb HEAD

The audit the verdict records (Round-3 post-absorb check by the sync-auditor, PASS, gate leg rcpt-d8421e7ddc1dc81049d70294) is bound to HEAD `d5dc6515f3c3e82dfd37a653eaf6e1ece8a4daf4` — the absorb merge of origin/main 259dbeec6d (t1976 landing). The absorb's own diff touches exactly AGENTS.md + internal/template/templates/AGENTS.md.tmpl (verified by the sync-auditor and the lane; zero card files), so the audited findings stand unchanged at the new HEAD; this machine line rebinds to it per the merge-evidence rule (audited commit must be HEAD or differ only under .moai/reports/t1973/). Prior binding: a980943254276978c84bdaedf3cd073c317014e2 (round-2 repair commit).

## Round 3 — 2026-10-27 (leader merge-hold: round-2 both-empty refusal broke two main paths; narrowed)

- trigger: leader CI reproduction — round-2 P2 (both-empty evidence refusal) broke `TestLeaderStartConsolidatesDeadOwnerRuns` and `TestFactoryRunsRetireRebindTo/a_operator_moves_the_cards_and_retires_the_source` (legitimate t1963 paths living in main). Reproduced by the lane at tree `64fcc418eb`: `unset GOFLAGS && go test ./internal/cli/ -run 'TestFactoryRunsRetireRebindTo|TestLeaderStartConsolidatesDeadOwnerRuns' -count=1` → exit 1 (`active runs = "run-old,run-new", want exactly run-new`; `err = factory run consolidation refused: card cardB: run run-tgt already holds a assigned row (evidence "", source "") ...`). Round-2 P1 (source-card lease) and round-1 P1 (terminal target) were NOT implicated.
- commits this round: `324af52281` absorb merge of origin/main `422b51a80d` (t1977 landing; leader-ordered, --no-ff; diff touches docs-site/hugo.toml + internal/template/testdata/binding_ledger.json only — zero card files) → `c7d2523a2e` narrowing fix. No push.

### Narrowing decision + rationale (lane judgment, recorded per leader directive)

The both-empty evidence refusal now applies ONLY when the SOURCE row is in a lease-holding (working) state — `run`, `sync`, `merge-ready`, ... — whose stage/worktree/spec_id an abandon would strand. An evidence-less NON-working source row (`picked` / `assigned` / `needs-decision`) keeps the t1963 abandon-with-pointer dedup the operator-move and leader-start-consolidation paths live on. Distinction axis: the source ROW's state. The run classification cannot distinguish (a consolidation source is always classified dead — step 0 refuses otherwise), and the target row's state cannot either (`picked` appears on both sides of the line). Round-2's RED test is unchanged — its source row is `CardRun`, still refused; new pin test `TestT1973_R3_BothEmptyNonWorkingSourceKeepsAbandon` (picked/assigned/needs-decision) pins the restored abandon from the other side.

### Round-3 evidence

- Reproduced the regression at `64fcc418eb` (exit 1, verbatim above); after the narrowing: the two CLI tests `ok github.com/modu-ai/moai-adk/internal/cli 15.977s` exit 0, and again post-absorb `ok ... 46.662s` exit 0; all t1973 suites `ok github.com/modu-ai/moai-adk/internal/homestate 2.540s` exit 0, post-absorb `ok ... 7.176s` exit 0.
- Full-package: `go test ./internal/homestate/ -count=1` → `ok github.com/modu-ai/moai-adk/internal/homestate 160.860s` exit 0. The full `internal/cli` suite under the same tree exceeded every local ceiling purely on machine load (1-min load 32-81, memory rule gate ≤32): 10m run and 45m run each ended in `panic: test timed out` with ZERO `--- FAIL` assertion lines — every cli test reached before the alarm passed. A load-gated full cli re-run (waits for load < 32, -timeout 120m) was still in flight at verdict-write time; the leader's origin/main CI owns the repository-wide verdict after landing regardless.
- Baseline: HEAD re-read immediately before this update (`git rev-parse` → c7d2523a2e8a4d98aaf4c3f43bcccad74c581fb0).

### Gaps (round 3)

- The load-gated full `internal/cli` run had not completed at verdict-write time (machine load above the entry gate); its verdict lands as a follow-up observation, and origin/main CI is the authoritative full-suite verdict after the leader merges the PR.

## Rebinding note 2 (2026-10-27) — audited_sha rebound to eef35b5abe

audited_sha rebound again to `eef35b5abefac01efa03292507e5862d3258d301` per the leader's direct read (the previous binding c7d2523a2e predated the round-2b repair). The independent sync-audit Round-5 (sync-audit-verdict.md, PASS, gate leg rcpt-d0fbe175c2272a995bba07d6) audits exactly this commit; the delta between c7d2523a2e and eef35b5abe is this file's own rebinding notes only (.moai/reports/t1973/ scope).

## Rebinding note 3 (2026-10-27) — audited_sha rebound to 0b03ab2141

audited_sha rebound to `0b03ab21413d913ab3c4d421bb59fbe96d0a35d2` (the structural whole-row gate commit). The independent sync-audit Round-7 (sync-audit-verdict.md, PASS, gate leg rcpt-5f54398825fb644b32c99e05, delta_base 148f049343) audits exactly this commit; the delta between eef35b5abe and 0b03ab2141 is this file's own rebinding notes plus the round-2b repair commits the rounds cover (.moai/reports/t1973/ scope for the rebinding notes; the repair commits are the audited content itself).

## Rebinding note 4 (2026-10-27) — audited_sha synced to 27e248839f

audited_sha synced to `27e248839fecc41b91d7211d8e5bd8c6b35700d6` (the post-absorb HEAD) per the leader's rebind of the card's evidence_sha to the same commit. The independent sync-audit Round-8 (sync-audit-verdict.md, PASS, gate leg rcpt-5df311471c95f09a8b7bd505) corroborates the tree identity: the absorb's own diff touches zero card files, the audited internal/homestate code at this HEAD is byte-identical to 0b03ab2141, and the merge-tree result tree equals HEAD's tree. The delta between 0b03ab2141 and 27e248839f is the origin/main-side landings plus this file's rebinding notes only.
