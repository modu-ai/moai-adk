# SPEC-HARNESS-DETACHED-PRUNE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

audit_ready: true
plan_status: audit-ready
plan_complete_at: 2026-10-05
attribution: plan-audit iter2 PASS 0.94 (verdict .moai/reports/t1497/plan-audit-iter2.md, commit fd08d4da8; repair 2d29b0509; codex adversarial enforced-required gate pass, zero findings; audit_receipt absent — measured stale-server gap documented in the verdict appendix, disposition routed to the leader).
attribution: delta-2 PASS (verdict .moai/reports/t1497/plan-audit-delta-2.md, commit 7d37b0c85; receipt rcpt-e10d81162cc1ac86b41e811c codex_audit pass; final plan-artifact hash as pinned in the delta-2 file).

## §E.2 Run-phase Evidence

**Delivery path note**: manager-develop landed M1 (`d6da6c1c5`) and M2 (`b50caa83b`, 7 files +280/−2) then was terminated mid-probe after two stalls on the delegation wave (GLM 429 static-stop at 00:42 + silent no-activity); the lane adopted its uncommitted M3/M4 work after a full diff read (t1495 stall-adoption protocol) and completed the remaining verification itself. Adoption commit: `055373cf4` (M3 wrapper wiring + REQ-DP-009 filter + six gate-review fixes + fixture refreshes; 20 files).

Verification matrix (all measured by lane-6 on this tree at `055373cf4`, env-scrubbed compound form):
- `go build ./...` → BUILD_OK; `GOOS=windows GOARCH=amd64 go build ./...` → WIN_BUILD_OK
- `gofmt -l internal/cli/ internal/harness/` → empty; `go vet ./internal/harness/... ./internal/cli/...` → clean
- `golangci-lint run --timeout=2m ./internal/harness/... ./internal/cli/...` → `0 issues.`
- `go test -count=1 -race -timeout 30m ./internal/harness/...` → `ok internal/harness 9.300s` (all sub-packages ok)
- Targeted cli `-race` (the new tests + every previously-failing family: PrePushSubcommandCount, DetachedChild×2, StopClassificationFiltersExpiredEvents, GateWiring, RetentionPruneVerb, RecordsBaseline, RecordsWhenEnabled, StopChainGateCutOff) → `ok internal/cli 10.923s`
- Gate-repro command (classify + high-water fixtures post-refresh) → `ok internal/cli 1.559s`
- AC-DP-001..008: flipped by the M1..M4 commits per acceptance.md's green paths (LEDGER-DP-GREEN-A/B ran as written: `-list` named the exact tests, `-run` executed them)

Ambient failures — attributed to the loaded host / pre-existing base, NOT this diff (change-scoped rule, AGENTS.md §4):
- `rosterguard` TestRegistryIsWellFormed/TestRegisteredSitesMatchTheirDeclaredAxis: cutover-* registry sites declare ClaimMembership on subset-by-design — registry.go untouched by this card (pre-existing at base; t1453/t1525 cutover provenance)
- `TestFactoryLeaseQueueLockStallBounded` timing bound (4.31s vs 3.8s) under `-race` on the loaded host; PASSES without `-race` on a calmer host (`ok 12.265s`) — machine measurement
- 30m package timeout truncated the full cli `-race` run (TestFR_AC024 in flight) — full-suite green is CI's verdict per the lane-local rule

Gaps: Windows RUNTIME unobserved (build+vet parity only, documented residual); the full-package cli suite green on an idle host is deferred to CI.

Residual risk: the spawn gate on a loaded host still costs one lock-free stamp read per observe event (sub-millisecond; REQ-HL-001 budget intact); an orphaned prune child completes outside the hook budget by design.

## §E.3 Run-phase Audit-Ready Signal

run_status: implemented
run_complete_at: 2026-10-06
attribution: M1 `d6da6c1c5` + M2 `b50caa83b` (manager-develop, Authored-By-Agent trailer on the draft→in-progress transition commit); M3/M4 adopted and verified by lane-6 (`055373cf4`) after the delegate's two-stall termination — full diff read before adoption per the stall-adoption protocol.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
spec_id: SPEC-HARNESS-DETACHED-PRUNE-001
card: t1497
phase: sync
sync_status: complete
sync_complete_at: 2026-10-06
sync_commit_sha: 6ced096f0ed72f75b54371eb430c85d1314686b2  # D3 backfilled — written as pending-backfill inside the sync commit itself (a commit cannot cite its own SHA), resolved here
sync_branch: WT-harness-prune-detached
frontmatter_status_transitions:
  implemented_to_completed: carried-by-sync-commit  # status + updated only, spec.md frontmatter
changelog_entry: added  # [Unreleased] §Fixed — user-visible harness behavior change (prune off the synchronous hook path); entry-worthiness judged against the repo's existing internal-behavior entries
b12_self_test_a: changelog_grep_count_0  # pre-emission `grep -c SPEC-HARNESS-DETACHED-PRUNE-001 CHANGELOG.md` = 0 — no duplicate emission
b12_self_test_b: ac_count_match_8_vs_8  # acceptance.md §C distinct live AC identifiers = 8 (awk counter: live=8 excluded=0 ambiguous=0); CHANGELOG entry references the same 8
b12_self_test_c: all_claimed_paths_exist  # internal/harness/{retention_spawn,retention_spawn_unix,retention_spawn_windows,learner,observer}.go + internal/cli/{hook,hook_retention_prune_test}.go verified via ls/wc before entry authoring
canary_compliance_check:
  single_sync_commit: true  # transition + §E.4 + CHANGELOG in ONE commit, no separate Mx chore
  no_push: true  # dispatch boundary — commits stay local on WT-harness-prune-detached; integration is the lane/leader's
  docs_scope: none  # README / docs-site / .moai/docs/ untouched — internal behavior change, no user-facing doc surface
  codemaps: skipped-not-required  # .moai/project/codemaps covers the touched packages but is generated from the develop integration tree (provenance d0378d37c, 2026-10-05T06:17Z); the sync skill mandates no per-card regeneration — one-line note carried in the sync commit body
mx_tag_delta: 0  # run-phase landed the @MX:ANCHOR/@MX:WARN tags on the new gate/child symbols; sync adds/removes no tags (documentation-only diff)
ac_summary: 8 PASS / 0 FAIL (acceptance.md §C; §E.2 run-phase matrix + sync-tree re-verification)
sync_verification: go build ./... + GOOS=windows build + AC green-path test families + observer.go PruneStaleEntries count 0 — all measured on this tree at the sync commit
```

## §F Phase 4 Mode Selection

Input parameters: tier=M; scope ~10 files (observer.go, internal/cli/hook.go, 2-4 new platform/exec files, 3-5 test files); domain count 1 (Go harness/cli source); language mix 100% Go; concurrency benefit LOW (coding-heavy); Agent Teams prereqs not requested.

| Mode | Selected | Rationale |
|------|----------|-----------|
| direct | not selected | Semantic multi-file change — not a typo/single-line fix |
| serial | **selected** | Coding-heavy single-domain implementation; Anthropic coding-task parallelism caveat |
| fanout | not selected | Coding-heavy work — fan-out buys no wall-clock and races one tree |
| sweep | not selected | Not ≥~30-file mechanical-uniform transform; coding-heavy stays serial |

Decision: serial

Justification: the card is one cohesive Go change (observer prune removal + spawn gate + platform detached-exec + hook wiring) with strict intra-file ordering and a single-writer tree discipline. A single manager-develop delegate carries the milestones sequentially; per the t1318 lesson its spawn auto-isolates to its own L1 tree and the lane reconciles by fast-forward merge (t1467 §6→§7 precedent). No agent-team request is in play; workflows add overhead with no parallelizable surface.

## §G Plan→Run Kickoff Decision Record

Gate form: **autonomous** (default transition per `.claude/rules/moai/workflow/auto-semantics.md` §9.1; keep-set categories not applicable — no environment-impossible, operator-held, or irreversible external-shared operation in scope).

Evidence criteria:
1. plan-audit iter2 **PASS 0.94** (Tier M threshold 0.80), zero blocking findings — verdict `.moai/reports/t1497/plan-audit-iter2.md`, commit `fd08d4da8`; repair `2d29b0509`.
2. codex adversarial backend (enforced-required gate) re-invoked on iter2: **pass, zero findings**.
3. 8/8 RED ledger cells measured on base tree `d05d1d5f0` this session (LEDGER-DP-A~H); `spec lint --strict` 0 findings (re-verified by the auditor).
4. plan-artifact hash unchanged since the audited state — post-verdict commits (`2d29b0509` is pre-verdict; `82a7088f1` and this record) touch progress.md only, which is outside the ComputeHash subject set.
5. §E.1 plan_status: audit-ready (commit `82a7088f1`).

Phase-1 skip contract (per spec-workflow § Plan Audit Gate skip policy): verdict PASS + score ≥ per-tier threshold + artifact-hash unchanged — the run delegation proceeds WITHOUT re-running the plan audit.

Known gap carried to the leader: codex results carry no `audit_receipt` — the running MCP server build (`c8f245c2c`) predates the receipt minter (`ffe36d8e3`, t686); measured stale-server gap, disposition documented in the verdict appendix (server reconnect by host, or leader exception record).

Decision: **PROCEED to run-phase** (milestones per plan.md, TDD RED→GREEN), delegated to manager-develop (serial).

Recorded by: lane-6 (factory lane, card t1497), 2026-10-05.

**Explicit wait (2026-10-05, lane-6)**: the manager-develop spawn was denied by `AUDIT_RECEIPT_VIOLATION` — the iter2 PASS stands without a corroborated receipt. Measured: the receipt minter (`ffe36d8e3`, t686) IS an ancestor of this tree (develop lineage, ancestor-exit=0) but the INSTALLED rc.24 build is compiled from the main lineage (`c8f245c2c`) which lacks it — a live `codex_audit` call returned a verdict with no `audit_receipt` field (guard present, minter absent: version skew, t1526 family). The guard opts in via `workflow.audit.gates.codex: required` (primary config) and has no bypass short of a minted receipt (`internal/hook/audit_receipt_guard.go:84`). Disposition routed to the leader: (a) rebuild+reinstall from the develop lineage, reconnect MCP, then the plan-auditor re-runs the codex leg and cites the receipt; or (b) relax the gate value (operator decision). Recheck point: the leader's reply. Plan artifacts are fully committed (HEAD `1d8ca5ac0`); run phase NOT started; tree preserved.

**Wait lifted (2026-10-05 22:13, lane-6)**: the leader adjudicated variant (a) — its own MCP server (rc.27, develop lineage, carries the minter) ran `codex_audit` directly against this tree (project_root argument) and minted receipt `rcpt-662d3de90b404764e8918857` (tool codex_audit, codex_verdict pass, tree_root this worktree, created 2026-10-05T13:11:25Z; store `.moai/state/audit-receipts/receipts/`). The gate-widening option (b) was rejected (D8 policy conflict); the guard-vs-minter version skew is being issued as its own card by the leader. Next: the plan-auditor re-issues its verdict line citing the receipt, then the run phase re-spawns.

**Receipt-gap resolution cycle complete (2026-10-06 00:5x, lane-6)** — the plan phase closed through six gate-review rounds and one leader-approved delta protocol:

- Plan-artifact trajectory: `080578aec` (author, 8 REQ/7 AC) → `2d29b0509` (iter1 repair: vacuous-green regex cell) → `eccd62531`+`8a430d101` (cross-package seam: `MaybeSpawnRetentionPruner` parameter injection + `internal/cli` `retentionSpawnImpl` wrapper var) → `949fbfc32` (REQ-DP-009 retention-aware classification + AC-DP-008 + fake sweep obligation) → `73264bddf` (aggregation-input filter mandated + child tests moved to cli) → `c6140ad9a` (utilitySubcmds exclusion) — final contract 9 REQ / 8 AC / 4 milestones, all RED/GREEN ledger cells measured on-tree.
- Audit chain: iter1 FAIL 0.88 → iter2 PASS 0.94 → delta-1 PASS (base ruling by the leader; scope = post-iter2 amendments) → delta-2 PASS → **final verdict registered as `plan-audit-iter3.md`** (parser-compatible score lines; `-iterN` name per the run-gate resolver; commit `42a968564`), machine verdict line: `AUDIT-VERDICT: PASS spec=SPEC-HARNESS-DETACHED-PRUNE-001 receipts=rcpt-1545d502869fd29c3a4accda` (codex_audit pass, this tree, 2026-10-05T15:45:05Z — minted by the auditor's own native driver after the leader's rc.27 server became unavailable; four independent codex passes recorded on the final tree).
- Audit-ceiling governance: the 2/2 Tier M ceiling was exceeded by leader ruling only (delta rounds 1-2 approved explicitly; the ceiling policy's intent — preventing infinite loops, not blocking valid amendments — was the leader's cited basis).
- Version-skew findings recorded for the leader's follow-up card: the guard-without-minter skew (main-lineage install vs develop-lineage minter), the run-gate resolver treating non-`-iterN` verdicts as iteration 1, and the `auditverdict.Parse` duplicated-key rejection of annotated score lines.
- §E.1 carries both notations (`audit_ready: true` factory key + `plan_status: audit-ready` schema prose); the run delegation proceeds with the skip contract satisfied by iter3 (PASS 0.94-equivalent final, hash pinned in iter3, no artifact change since).
