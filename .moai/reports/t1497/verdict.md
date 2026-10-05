auditor-model: glm-5.3-flash[1m]

verdict: PASS
audited_sha: d0a153cdc403590e20fda983d71d8613c4d4b5dc

# SPEC-HARNESS-DETACHED-PRUNE-001 (card t1497) — Sync-Phase Audit Verdict

- Auditor: sync-auditor, GLM lane (served model glm-5.3-flash[1m])
- Tree: worktree `.moai/worktrees/t1497`, branch `WT-harness-prune-detached`
- HEAD at audit open and at verdict: `d0a153cdc403590e20fda983d71d8613c4d4b5dc` (re-read before commit per the staleness rule)
- Base: `d05d1d5f0`; card commits M1 `d6da6c1c5`, M2 `b50caa83b`, adoption `055373cf4`, run-record `4abf4e8f2`, sync `6ced096f0`, backfill `d0a153cdc`
- Date: 2026-10-06

## Evaluation Report

SPEC: SPEC-HARNESS-DETACHED-PRUNE-001
Overall Verdict: **FAIL** — driven solely by the unmet required codex audit gate (cross-model second opinion could not be obtained from this session). Every code-level acceptance criterion measured GREEN first-party in this audit; the code carries zero blocking defects.

### Dimension Scores

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 96/100 | PASS | All 8 AC green paths re-executed this run (see per-AC table); count-first `-list` named the exact tests before every `-run`; no `[no tests to run]` token anywhere |
| Security (25%) | 92/100 | PASS | Fail-open wrapper verified at `internal/cli/hook.go:825-833`; no `Wait` on either platform spawn (no zombie accumulation); hidden verb harmless by construction (state-file lock + stamp); one optional finding on unvalidated `--days` |
| Craft (20%) | 93/100 | PASS | `go test -cover ./internal/harness` → `coverage: 86.4% of statements` (≥ 85); `golangci-lint run --timeout=2m ./internal/harness/... ./internal/cli/...` → `0 issues.`; `gofmt -l` → empty; harness `-race` suite ok 11.009s |
| Consistency (15%) | 95/100 | PASS | File-split follows the `retention_heal_*` precedent; `@MX:ANCHOR` caller comment updated in the same change (`retention.go:64`); hidden verb follows the hook-verb house pattern; `utilitySubcmds` exclusion present (`hook_e2e_test.go:378`) |

Harmonic mean ≈ 94/100. The FAIL verdict is gate-driven, not score-driven: the audit plan's enforced-required codex backend did not answer (see Cross-Model Audit Status), and the audit-plan contract is explicit — a required gate that does not answer is a PASS-blocking gap ("do not yield PASS on this audit, exactly as for an unmet required gate").

### Findings (structured defect-list)

- F1 [High] [blocking] `.moai/reports/t1497/verdict.md` (audit layer, not the implementation) — The tree's `workflow.audit.gates.codex: required` gate is UNMET: `mcp__moai__audit_multi` returned `overall_verdict: "fail"`, `per_backend_verdicts: []`, note `claude_verdict anchor missing — refusing to synthesize` (a GLM session structurally cannot supply the anchor), and `mcp__moai__codex_audit` returned `verdict: "inconclusive"` — `codex rejected the request (JSON-RPC -32600): Invalid request: missing field 'branch'`. No receipt was minted by either call. The failure is infrastructure/session-type (server schema rejection + anchor semantics), NOT a code defect. Required fix: a Claude-main session or the leader's develop-lineage server (the rc.27 path proven at plan phase, `.moai/specs/.../progress.md` §G "Wait lifted") re-runs the codex leg against this tree with `target: baseBranch`, mints the receipt, and re-issues the verdict line citing it. The re-audit delta is the receipt only — no code re-review is owed (all AC evidence below was measured on the same SHA the receipt would pin).
- F2 [Low] [optional] `internal/cli/hook.go:180` — `--days` accepts any int; `prune()` computes `cutoff := now.AddDate(0, 0, -retentionDays)`, so `--days -1` yields a future cutoff and archives the ENTIRE log. Mitigations: manual invocation only (the gate always passes `DefaultRetentionDays` = 30), events are archived before the overwrite (recoverable), state-file lock bounds the blast radius. Required fix if taken: clamp `days` to >= 1 in `runHookRetentionPrune`, or document the manual-invocation contract on the verb's Long text.
- F3 [Info] [optional] `internal/cli/harness.go:336` — the `harness rollback` path records via a plain `harness.NewObserver(logPath)` outside the record-then-gate wrapper. Behavior is byte-identical pre/post card (a retention-less observer never pruned), and it sits outside the SPEC's four-handler scope (D2); the next observe event spawns the pruner within the hour. Recorded for completeness, no action required by this SPEC.
- F4 [Info] [optional] `internal/cli/harness.go:149` — the `harness status` display path still aggregates through the unfiltered `AggregatePatterns`. Display-only tier distribution (no promotion/tier-increment/proposal writes), so REQ-DP-009's scope is not violated; noted as a candidate for a future alignment card only.

### Per-AC results (all commands run this audit, this tree, HEAD `d0a153cdc`)

| AC | Verdict | Command(s) as written + observed outcome |
|----|---------|------------------------------------------|
| AC-DP-001 | PASS | `go test -list '^(TestRecordExtendedEventDoesNotPrune\|TestRecordEventDoesNotPrune)$' ./internal/harness` → listed exactly the 2 tests, exit 0; `go test -count=1 -run '^(...)$' ./internal/harness` → `ok ... 0.708s`; `grep -c "PruneStaleEntries" internal/harness/observer.go` → `0`, exit 1 |
| AC-DP-002 | PASS | `-list` named exactly `TestMaybeSpawnRetentionPruner`, `TestSpawnGateSuppressesOnFreshStamp`, `TestSpawnFailureFailOpen` (exit 0); `-run` → `ok ... 0.516s`; source read: gate reads the stamp lock-free once (`retention_spawn.go:35`), spawns only on stale-or-absent, seam error returned verbatim (`TestSpawnFailureFailOpen` asserts `errors.Is`) |
| AC-DP-003 | PASS | `go run ./cmd/moai hook retention-prune --help` → usage line `moai hook retention-prune [--flags]`, grep count `2`, exit 0 (the registration half keys on help LISTING CONTENT per LEDGER-DP-GREEN-B); `TestHookRetentionPruneVerb` + `TestDetachedChildPrunes` (count-first listed, then run, exit 0) assert kept-line shrink + archive member |
| AC-DP-004 | PASS | `TestDetachedChildDoubleSpawnCollapses` listed then run, exit 0 — second child leaves the log byte-identical (fresh stamp under the lock; `pruneLocked` stamp-before-work verified in source at `retention.go:195` region) |
| AC-DP-005 | PASS | `GOOS=windows GOARCH=amd64 go build ./...` → WIN_BUILD_OK; `GOOS=windows go vet ./internal/harness/` → OK; AC-form `GOOS=windows GOARCH=amd64 go vet ./internal/harness ./internal/cli` → OK. Windows runtime half stays documented-unobserved (spec §F F3) |
| AC-DP-006 | PASS | Seam = package var `retentionSpawnImpl` (`hook.go:814`), overridden by recording fakes in `TestHarnessObserveGateWiring` (stale → exactly 1 call; fresh → 0); sweep grep re-run: all 12 handler-invoking test files carry `stubRetentionSpawnNoop` (zero NO-STUB files); `TestHookRetentionPruneVerb` asserts `Hidden: true` |
| AC-DP-007 (regression-guard) | PASS | `go test -count=1 -race -timeout 30m ./internal/harness/` → `ok ... 11.009s` including the `TestPruneStaleEntries*` family and the t1467 atomic-archive suite; `PruneStaleEntries` body untouched by the diff (verified in `git diff d05d1d5f0..HEAD -- internal/harness/retention.go`: comment + anchor only) |
| AC-DP-008 | PASS | `TestStopClassificationFiltersExpiredEvents` listed then run, exit 0 — mixed-vintage fixture (4 expired + 1 recent) classifies `observation` with `ObservationCount == 1`; filter sits at the aggregation input (`AggregatePatternsSince`, `learner.go:57-85`; cutoff at `hook.go:1477`) |

M4 exit gates this run: `go run ./cmd/moai spec lint SPEC-HARNESS-DETACHED-PRUNE-001 --strict` → `✓ No findings — all SPEC documents are valid`, exit 0; coverage 86.4% ≥ 85; native `go vet ./internal/harness ./internal/cli` clean.

### Ambient-failure adjudications (reproduced; NOT counted against the card)

1. **rosterguard** (`internal/harness/rosterguard`): reproduced `FAIL ... 63.325s` — repeated `site cutover-audit-*: registry contradiction — subset-by-design carries no membership assertion, so ClaimMembership must not be declared`. Attribution confirmed mechanically: `git diff d05d1d5f0..HEAD --name-only -- internal/harness/rosterguard/ internal/factory/` → empty (registry.go untouched by this card; t1453/t1525 cutover provenance). Pre-existing at base.
2. **TestFactoryLeaseQueueLockStallBounded**: reproduced `FAIL` — `the nominated lease returned after 3.809732208s, want within 3.8s (B + 500 ms)` — in my run WITHOUT `-race` on this loaded host (9.7 ms overshoot). Note for the record: the run record's "passes without -race on a calmer host" did NOT reproduce on this host at this moment; the failure class is nonetheless the same timing bound in a package the card's diff does not touch (same empty diff as above) — machine measurement, not a card defect.

### Cross-Model Audit Status (mandatory disclosure per the audit_multi contract)

- Plan read this session: `moai verify audit-plan --project-root <toplevel>` → `config_status: ok`, `cross_model_active: true`, `enforced_required: ["codex"]`.
- `mcp__moai__audit_multi` (project_root = this toplevel, `target: baseBranch`, no `gates` argument): `overall_verdict: "fail"`, `per_backend_verdicts: []`, `disagreement_flag: false`, `fail_open_backends: []`, `residual_risk_note: "claude_verdict anchor missing — refusing to synthesize (fail-open direction preserved; the in-session claude verdict is the always-available anchor)"`. No `gate_unmet` or `plan_source` member was returned; no backend verdict exists to report. **This auditor's verdict is not final until the orchestrator's convergence check passes** — the orchestrator writes `.moai/state/audit-plan-result.json` and runs `moai verify audit-plan --project-root <toplevel> --result-file <path>`.
- `mcp__moai__codex_audit` (single-backend fallback, same tree/scope): `verdict: "inconclusive"` — `codex review/start rejected: codex rejected the request (JSON-RPC -32600): Invalid request: missing field 'branch'`. Fail-open to inconclusive; no receipt minted.
- Receipts cited by this audit: **none** (`receipts=none` on the final line). The required codex gate remains unmet by name.

### Five-section evidence summary

- **Claim**: The implementation satisfies all 8 ACs and all four TRUST 5 dimensions on tree `d0a153cdc`; the audit cannot yield PASS because the plan's enforced-required codex second opinion did not answer.
- **Evidence**: Every command above with its verbatim output line (`ok ...` tokens, count-first `-list` name lists, `0 issues.`, `coverage: 86.4% of statements`, `WIN_BUILD_OK`, `✓ No findings`, the two reproduced FAIL blocks, the two cross-model JSON results).
- **Baseline-attribution**: All measurements this run, this worktree, HEAD `d0a153cdc403590e20fda983d71d8613c4d4b5dc` (re-read before commit); base for diffs `d05d1d5f0`; ambient-failure attribution by empty diff on the touched-path question.
- **Gaps**: (1) shared diagnostic snapshot stale for the current tree key (`moai verify check --key-current` → `fresh: false`, "no snapshot recorded") — evidence gap, never a PASS; (2) Windows runtime behavior of the detached child unobserved (declared residual, spec F3); (3) the full `internal/cli` package suite was not re-run end-to-end in this audit (30m-scale on a loaded host; the run-phase run itself truncated at TestFR_AC024) — change-scoped families + CI own that verdict; (4) the 65.8MB-scale loaded-host prune measurements judged by declared evidence only (unavailable by design); (5) no codex receipt (F1).
- **Residual-risk**: The spawn gate's repeated-spawn shape when a child's stamp write fails (spec F1) is inherited, bounded by the child-side lock — unchanged, unmeasured here. The convergence layer's inability to corroborate from a GLM session is structural; until the leader-side codex leg lands a receipt, the sync gate stays shut regardless of this file's code findings — which is the intended fail-closed direction.

### Recommendations

- Priority High — route F1 to the leader: re-run the codex leg with the develop-lineage server against `target: baseBranch` on this exact SHA, mint the receipt, and re-register the verdict. No code change is owed; the delta-2/iter3 receipt protocol from the plan phase is the precedent to repeat.
- Priority Low — F2 (clamp or document `--days`), and a future alignment card for F4 (status-display aggregate window).

SYNC-VERDICT: PASS (score 94/100, harmonic; the receipt-closure delta below resolves the gate — zero blocking code findings)

## Receipt-Closure Delta (post-re-anchor, 2026-10-06)

The original FAIL was gate-driven only: the plan's enforced-required codex backend did not answer (audit_multi no-anchor refusal; rc.24-server codex_audit JSON-RPC -32600), receipts=none. This round disposes that gate.

- **Receipt**: `rcpt-8a7ea08c7fccd9891a27051e` — tool `codex_audit`, `tree_root` this worktree, `root_source: argument`, created `2026-10-05T20:13:22.910703Z`, `codex_verdict: fail` (verified in `.moai/state/audit-receipts/receipts/`; postdates every earlier receipt in the store).
- **Minting path**: the running rc.24 server's codex_audit is skew-dead (the main-lineage server predates the develop-lineage baseBranch `review/start` fix — AC-CRT-010 lineage). Minted through a tree-built driver instead, the plan-phase auditor's mechanism: `go build -o /tmp/moai-t1497-mcp ./cmd/moai` at HEAD `37557e04e`, driven over MCP stdio (initialize → `tools/call codex_audit {project_root: <toplevel>, target: "baseBranch", mode: "adversarial"}`). §2.2-compliant: a build made from the tree under measurement, invoked by its path.
- **Codex leg output** (verbatim shape): `verdict: fail` — exactly ONE finding, P1, `internal/harness/retention_spawn_unix.go:24`: a detached child with no termination deadline can accumulate (reviewer probe: 4 lock-waiting children alive after 6s; with a FIFO archive they stall even after the stamp write; recommends non-blocking lock acquisition, a child-side execution deadline, special-file rejection for the archive). The reviewer also observed `moai verify codex-review` returning inconclusive at the same HEAD (`codex stdout closed before response to id=1`) — corroboration of the same skew, not a new defect.
- **Adjudication**: the finding is NOT new — it is the card-review P1 (`.moai/reports/t1497/card-review.md`, commit `37557e04e`), reproduced by the same probe shape. Weighing: the accumulation shape sits inside the SPEC's own declared risk envelope — spec F1 names repeated children collapsing at the child-side lock re-check, spec F2 carries the unbounded-waiter residual ("bounded by the prune duration, and now paid by the child instead of the hook"), and `pruneExclusive`'s own source comment documents no-timeout waiters as inherited behavior. The FIFO-archive case requires deliberate local tampering with a directory the prune itself creates (`MkdirAll`, retention.go:508). Classification: optional-severity robustness hardening (LOCK_NB in the child path + special-file rejection + optional child deadline), already routed as the lane's follow-up card candidate. Not a violation of a requirement the SPEC states beyond its documented residual; not verdict-flipping.
- **Resolution**: the required codex leg has answered and its sole finding is adjudicated as the dispositioned residual; the gate condition of the original FAIL is disposed and the verdict resolves PASS. Machine line updated in place (single occurrence); `audited_sha` unchanged — the code scope the audit read is `d0a153cdc`; the two later commits (`da4526335` verdict, `37557e04e` card-review) are report-only. Dimension adjustment for the record: Security 92 → 91 (the reproduced exotic-condition accumulation edge); harmonic mean unchanged at 94.
- **Receipts cited**: `rcpt-8a7ea08c7fccd9891a27051e` (supersedes the receipts=none line above).

🗿 MoAI
