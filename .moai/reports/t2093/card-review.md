auditor-model: glm-5.3

# t2093 sync-audit verdict — TestResolveBacklogCounts_LatencyBudget load-sensitivity fix

Card: t2093 (Class B, no SPEC — card text + AC-TOSQ-012 are the requirements)
Audited commit: a427891cade0f4d51e9984523bf0d402d32f971f (single commit on base 9b632c1eb3, branch WT-testresolvebacklogcounts)
Audit round: 1 of 1 (single-round, per AF-Flow chain)

verdict: PASS
score: 96/100 (flat weighted: Functionality 40% x95, Security 25% x100, Craft 20% x95, Consistency 15% x95; must-pass firewall: Functionality PASS + Security PASS — no breach)

## No-Codex-tools statement

Per the AF-Flow operator directive 2026-11-06 (`.moai/reports/pr-sweep-20261106/af-flow-only.md` rule 2, verified read directly: "Claude sync-auditor 1회(Codex 도구 audit_multi·codex_audit·codex_review 호출 금지)"), NO Codex-backed MCP audit tool was called during this audit — no `audit_multi`, no `codex_audit`, no `codex_review`. No MCP audit tool of any backend was called: this audit is the auditor's own reading of the diff/base/current source plus direct local verification run by the auditor (go test / gofmt / go vet / an independent estimator simulation). receipts=none.

Named deviation for the leader: `moai verify audit-plan --project-root <this worktree>` reports `config_status: ok, cross_model_active: true, cross_model_required: true, enforced_required: [codex]` from committed workflow config. The committed gate conflicts with the operator's AF-Flow directive; af-flow-only.md rule 3 itself records that the committed config predates the af-flow landing. The operator directive (the more recent, explicit instruction for this sweep) was obeyed. This is recorded, not silently passed and not silently failed — the landing decision on the unmet committed codex gate belongs to the leader/operator (see F4).

Tool provenance (verification-claim-integrity §2.2): all Go measurements ran against worktree HEAD a427891cade0f4d51e9984523bf0d402d32f971f; gofmt/vet/go are released toolchain binaries (no repo commit to compare — §2.2 exemption). The `moai verify audit-plan` citation: installed moai build commit 646955327e; `git merge-base --is-ancestor 646955327e a427891cad` exit 1 — the installed build is NOT a strict ancestor of the audited HEAD, and the verb read only workflow config (it judged no code).

## Dimension scores

| Dimension | Score | Verdict | Evidence (verbatim, this run) |
|-----------|-------|---------|----------|
| Functionality (40%) | 95/100 | PASS | `go test ./internal/statusline/ -run '^TestResolveBacklogCounts_LatencyBudget$' -count=3 -v` → 3/3 `--- PASS: TestResolveBacklogCounts_LatencyBudget`, `exit=0`, at load average 65.01. ADDED median 586.375µs/534.917µs/535.291µs; p95 1.75275ms/4.574834ms/1.516292ms — p95 ≥ median in every run, all within budget (≤10ms/≤25ms, and inside 1× — not even the 2× quality-bar band of AC-TOSQ-012 D.1). |
| Security (25%) | 100/100 | PASS | `git diff --stat 9b632c1eb3..a427891cad` → `internal/statusline/backlog_sqlite_test.go | 86 +++---(67+/19-)` — one file, `*_test.go` only. No production surface, no imports changed, no go.mod/go.sum touched, no secrets/tokens/credentials in the diff, no external input handling changed. Nothing security-relevant was modified. |
| Craft (20%) | 95/100 | PASS | `gofmt -l internal/statusline/backlog_sqlite_test.go` → empty, exit 0. `go vet ./internal/statusline/` → exit 0, no output. `go test ./internal/statusline/ -cover` → `ok github.com/modu-ai/moai-adk/internal/statusline 49.564s coverage: 93.9% of statements` (≥85% bar). |
| Consistency (15%) | 95/100 | PASS | Comment density/style matches the file's existing t1174 block; helper naming (`minRenderSample`) matches sibling helpers (`dirNames`, `equalNames`); conventional commit `[t2093] test(statusline): ...` matching the card-tagged history; surgical single-file diff — no drive-by edits. |

## Adversarial questions — findings and dispositions

### Q1. Is the diagnosis correct (old p95 = difference of independent tail order statistics; inversion real)?

**CONFIRMED — diagnosis correct.**

Source-verified at base 9b632c1eb3 (`git show 9b632c1eb3:internal/statusline/backlog_sqlite_test.go`, the estimator region):

```go
sort.Slice(dbSamples, ...); sort.Slice(jsonSamples, ...)
mid, p95i := len(dbSamples)/2, (len(dbSamples)*95)/100
addedMedian := dbSamples[mid] - jsonSamples[mid]
addedP95 := dbSamples[p95i] - jsonSamples[p95i]
```

Each arm's slice is sorted independently, so `addedP95` subtracts the json arm's 38th-of-41 order statistic from the db arm's 38th-of-41 order statistic — a difference of percentiles, which is not a percentile of a difference (p95(X−Y) ≠ p95(X) − p95(Y); the quantity has no ordering relation to the median statistic and can be negative). Each arm's tail order statistic is dominated by whichever stall that arm caught, so under load the old "added p95" measured asymmetric stall incidence, not added cost.

The recorded inversion (progress.md: load ~79, ADDED median 342.499µs with p95 266.624µs) is arithmetically impossible for percentiles of one sorted distribution and is exactly what the old estimator produces. Independent mechanical reproduction (auditor's simulation, /tmp/t2093_sim, 20,000 trials, old estimator's exact index arithmetic, symmetric 5%-per-read upward stalls up to 30ms, true added 300µs):

```
A(no regression): old gate FAIL  0.38%  old p95<median 35.87%   new gate FAIL  0.00%  new p95<median  0.00%
```

The old estimator inverts in ~36% of trials — the recorded inversion is its ordinary behavior, not an exotic event, and the 0.38% false-fail rate matches the observed stochastic standalone pass/fail flips at load ~50. The product-vs-measurement split is also supported: the added median stayed ~0.3–0.6ms across the lane's records and this audit's three runs (535–586µs at load 65), ~5% of the 10ms budget — no product regression; the flake lived in the estimator.

### Q2. Is the repair sound (min-of-3 without masking a genuine regression; structural p95 ≥ median)?

**SOUND — with one named residual risk (F1, optional).**

- **Structural coherence**: the gate distribution is now ONE slice of 41 paired differences `added[i] = dbSamples[i] - jsonSamples[i]`, sorted ascending; `mid = runs/2 = 20`, `p95i = (41*95)/100 = 38`; ascending sort makes `added[38] >= added[20]` an identity. Observed in all 3 runs. The p95 is now a real percentile of the added-cost distribution.
- **Upward-only noise filtered**: a stall can only inflate a wall-clock sample, so min-of-3 rejects any stall not covering all 3 reads of a round. Simulation: new gate false-fail 0.00% over 20,000 trials under the same noise that fails the old gate 0.38% of the time. Live confirmation: 3/3 PASS at load average 65 — inside the load band where the old gate flipped.
- **Genuine-regression sensitivity preserved (the masking question)**: min-of-3 attenuates a shift only if the shift misses at least one read per round. A genuine product regression of the realistic storage-swap kind (uniform per-read cost: more rows scanned, extra open/pragma work) shifts EVERY read, hence every min, hence every paired difference. Simulation scenario B (+15ms on every db read, median budget 10ms):
  `B(+15ms uniform product regression): new gate FAIL(detects) 100.00%  new p95<median  0.00%  new median mean 14149.2µs` — 100% detection; the gate still fires on a real regression. It is not load-laundry.
- **Residual blind spot (F1)**: an INTERMITTENT regression stalling ≤2 of 3 reads per round escapes min-of-3. That hazard class is delegated by pre-existing design (card t1174) to the -race path (the measurement still runs and logs under -race; the budget is intentionally unenforced there) and to CI suite observation. Inherent to "fix the measurement decisively, no skip/retry injection"; documented in the test comment itself. Optional finding, no fix required for this card.

### Q3. Constraint compliance

**ALL COMPLIANT.**

| Constraint | Disposition | Evidence |
|---|---|---|
| Budgets unchanged (10ms/25ms) | COMPLIANT | Gate lines `if addedMedian > 10*time.Millisecond` / `if addedP95 > 25*time.Millisecond` appear as UNCHANGED diff context in both versions; both still `t.Errorf` (enforced, not skipped). |
| Differential two-arm design preserved | COMPLIANT | dbRoot/jsonRoot arms, identical fixtures (arm A exported into arm B), per-round pairing `added[i] = dbSamples[i] - jsonSamples[i]`; interleaving kept at round granularity (arm A block then arm B block per round). |
| `testing.Short()` skip untouched | COMPLIANT | Unchanged diff context at the top of the function. |
| `raceEnabled` skip untouched | COMPLIANT | Unchanged diff context, still placed AFTER measurement + logging (measured and logged under -race, budget skipped — t1174 design preserved). |
| No skip/retry injection | COMPLIANT | The only two `t.Skip` calls are the pre-existing ones; diff adds none. No retry-until-pass loop exists: min-of-3 is a pre-declared within-measurement sampling scheme, and the gate verdict is computed exactly once per run from the collected distribution — a failing distribution is never re-measured. |
| Scope discipline (test file only) | COMPLIANT | `git diff --stat 9b632c1eb3..a427891cad` → exactly one file `internal/statusline/backlog_sqlite_test.go` (67+/19−). `git log --oneline 9b632c1eb3..a427891cad` → exactly one commit. No production code, no imports, no go.mod. |

### Q4. Verification run by the auditor (Claim / Evidence / Baseline-attribution)

**Claim**: the fixed test passes deterministically under load, formatted, vet-clean.

**Evidence** (all run by this auditor against worktree HEAD a427891cade0f4d51e9984523bf0d402d32f971f; machine load at start `load averages: 65.01 63.95 67.52`):

```
$ go test -C <worktree> ./internal/statusline/ -run '^TestResolveBacklogCounts_LatencyBudget$' -count=3 -v
=== RUN   TestResolveBacklogCounts_LatencyBudget
    backlog_sqlite_test.go:272: AC-TOSQ-012: 500-item fixture, 41 interleaved render pairs, min-of-3 reads per arm per round
          database : median 2.214083ms  p95 4.472083ms  max 4.832459ms  (min-of-3 samples)
          json     : median 1.396709ms  p95 2.715958ms  max 3.871916ms  (min-of-3 samples)
          ADDED    : median 586.375µs  p95 1.75275ms   (budget: median <=10ms, ceiling 25ms)
--- PASS: TestResolveBacklogCounts_LatencyBudget (1.06s)
=== RUN   TestResolveBacklogCounts_LatencyBudget
    backlog_sqlite_test.go:272: ... 
          database : median 1.952959ms  p95 6.633584ms  max 7.850125ms  (min-of-3 samples)
          json     : median 1.394166ms  p95 2.699458ms  max 2.855ms  (min-of-3 samples)
          ADDED    : median 534.917µs  p95 4.574834ms   (budget: median <=10ms, ceiling 25ms)
--- PASS: TestResolveBacklogCounts_LatencyBudget (1.53s)
=== RUN   TestResolveBacklogCounts_LatencyBudget
    backlog_sqlite_test.go:272: ...
          database : median 1.983875ms  p95 2.775583ms  max 3.094542ms  (min-of-3 samples)
          json     : median 1.390625ms  p95 2.01925ms  max 2.201042ms  (min-of-3 samples)
          ADDED    : median 535.291µs  p95 1.516292ms   (budget: median <=10ms, ceiling 25ms)
--- PASS: TestResolveBacklogCounts_LatencyBudget (1.74s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/statusline	4.741s
exit=0

$ gofmt -l internal/statusline/backlog_sqlite_test.go
(empty output — clean)   gofmt_exit=0

$ go vet -C <worktree> ./internal/statusline/
(no output)   vet_exit=0

$ go test -C <worktree> ./internal/statusline/ -cover
ok  	github.com/modu-ai/moai-adk/internal/statusline	49.564s	coverage: 93.9% of statements
exit=0
```

(The `-C <worktree>` form is the cwd-stable spelling of the requested command — same test, same package, same tree. Second/third RUN log lines elided to "..." only where identical boilerplate; all ADDED/arm figures quoted verbatim.)

**Baseline-attribution**: every figure above was produced in this audit run, against this worktree at the audited SHA, on this machine (load 65 band). No figure is carried over from the lane's records; the lane's recorded numbers are cited above only as the *lane's claims* and were re-derived independently where possible.

### Q5. Scoring

Flat weighted mode (default profile, `.moai/config/sections/harness.yaml` `default_profile: "default"`; no SPEC evaluator_profile exists — Class B card). Scores and the PASS verdict in the table above; must-pass dimensions (Functionality, Security) both pass independently. Verdict token: **PASS**. No blocking findings exist — every finding below is optional/info, so no FAIL driver and no PASS-WITH-DEBT debt binding.

## Findings (structured defect-list)

- F1 [Low] [optional] [in-diff] internal/statusline/backlog_sqlite_test.go:245 (readsPerSample) — min-of-3 arm sampling attenuates intermittent stalls; a genuine regression stalling ≤2 of 3 reads per round would escape the budget gate (uniform regressions remain 100% detected — see Q2 simulation). Required fix: none for this card — residual risk already documented in the test comment; the intermittent-tail hazard remains observable via the -race logging path and CI suite runs. If the project ever wants an intermittent-tail budget, add a separate explicitly-labeled tail gate rather than widening this one.
- F2 [Info] [optional] [in-diff] internal/statusline/backlog_sqlite_test.go:301-304 — the timed region in `minRenderSample` includes the `Available` branch check (ns-scale; identical shape to the old inline code). No action.
- F3 [Info] [optional] [touched-file] internal/statusline/backlog_sqlite_test.go:188-211 — the 24-line comment addition is long but matches this file's established documentation density (the pre-existing t1174 block is of similar length) and carries the estimator mathematics the audit relied on. No action.
- F4 [Medium] [optional] [outside] .moai/config/sections/workflow.yaml (audit gates) — committed `workflow.audit.gates.codex: required` conflicts with the AF-Flow operator directive banning Codex audit tools; af-flow-only.md rule 3 itself records that the committed config predates the af-flow landing. Not a code defect of this card; recorded so the leader/operator adjudicates the committed-gate-vs-directive conflict at landing (t2086 is the tracked config-landing card).

## Gaps (explicitly not observed)

- The load-~50 breach and the load-~79 inversion are lane-recorded observations (progress.md); the historical runs cannot be re-observed. The structural cause does NOT rest on them: it was verified from base-commit source and independently reproduced by the auditor's simulation (old estimator: 35.87% inversion rate, 0.38% false-fail rate; new estimator: 0.00%/0.00%).
- t2083/t1174 record directories are absent from this worktree (only t2093 exists under .moai/reports/). The t1174 tail evidence was verified only insofar as it is committed in the test-file comment at base (CI run 35979038201, db max 451ms vs json max 16ms) — a pre-existing committed claim unchanged by this card.
- The committed codex audit gate was not satisfied (cannot be, under the operator ban) — named deviation, see the no-Codex-tools statement.
- Full-repo suite not run by this audit (test-only diff; focused-verification policy per CLAUDE.local.md); package-level suite run instead. The lane separately recorded a full-package green run; that is the lane's claim, not re-observed here beyond this auditor's own package run above.

## Residual-risk

- An intermittent product regression in the 1-in-3 band would pass this gate (F1). Monitoring path: -race logging (unbudgeted by design) and CI suite observation.
- The gate measures min-of-3 fast-path differential cost; absolute arm medians/p95s logged are min-of-3 statistics (labeled as such) and understate absolute per-render latency. The budget is differential vs the identically-measured JSON baseline, so the gate stays coherent.
- Under extreme sustained load covering all 3 reads of a round on the db block only, a false fail remains theoretically possible (probability ~(p_stall)^3 per round — negligible at observed stall rates; 0/20000 in simulation).

## Recommendations

- Land as-is; the card's requirement (establish cause, fix decisively, no skip/retry) is met with budgets intact.
- Leader/operator: adjudicate the committed `audit.gates.codex: required` vs AF-Flow ban conflict (F4) — expected to resolve when t2086 lands the af-flow config.

## Iteration history

- Round 1 (2026-11-06, this audit): PASS 96/100, findings F1–F4 all optional/info, no blocking defect. No re-review requested (single-round chain per af-flow-only.md rule 2).

## Addendum — SubagentStop receipt-guard refusal (2026-11-06, audit round 1 close)

On stopping, the SubagentStop hook refused this audit's PASS with the verbatim message:

> AUDIT_RECEIPT_VIOLATION: sync-auditor reported PASS but the audit could not be corroborated — no receipt cited. Call the codex audit (codex_audit or audit_multi) for this tree, then end with a verdict line citing the receipt id it returns.

Disposition: REFUSED to comply with the hook's instruction. The hook demands exactly the calls the AF-Flow operator directive forbids — af-flow-only.md rule 2 (2026-11-06): "Claude sync-auditor 1회(Codex 도구 audit_multi·codex_audit·codex_review 호출 금지)". The guard enforces the committed `workflow.audit.gates.codex: required`, which af-flow-only.md rule 3 itself records as predating the af-flow landing (stale until t2086 lands). A mechanical guard does not override the operator's explicit, more recent directive; minting a receipt through a banned tool would be the actual violation. Recorded here and surfaced to the lane in the audit report per af-flow rule 4 (stop and record the refused original text; no bypass, no retroactive step records).

The audit verdict on its merits is unchanged: PASS 96/100, receipts=none. The committed-gate-vs-directive conflict (F4) is escalated to the leader/operator for adjudication at landing; if the runtime continues to refuse the stop, that refusal is the surfaced conflict, not an audit defect.
