# progress.md — SPEC-WORKFLOW-CACHE-OPT-001

> Canonical §E section skeleton. Plan-phase populates §E.1 only; §E.2/§E.3 are
> owned by manager-develop (run-phase), §E.4 by manager-docs (sync-phase).

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-13
- plan_audit: iter-1 FAIL 0.84 → D1-D12 fixed (v0.1.1); iter-2 PASS-WITH-DEBT 0.93 → residual D13/D14 fixed (spec.md v0.1.2, 2026-07-13)
- tier: L
- artifacts: spec.md, plan.md, acceptance.md, design.md, research.md (Tier L 5-artifact set) + progress.md skeleton
- REQ count: 36 (REQ-SNAP-001..011, REQ-GATE-001..006, REQ-DELEG-001..006, REQ-AUDIT-001..004, REQ-BOOK-001..005, REQ-GUARD-001..004)
- AC count: 36 (AC-WCO-001..036, 1:1 REQ traceability)
- depends_on: SPEC-GOAL-ENGINE-001 (status: completed — fulfilled)
- edit-target inventory: 19 files (17 workflow byte-parity + 2 agent sanitized-parity) per plan.md §A
- open decisions: 0 — all 3 clarifications resolved by user decision (plan-audit iter-1 D1): porcelain-v2 + diff-HEAD key digest (strengthened per iter-2 D13) / key-equality+10-min-TTL freshness / exact byte-string stop-goal match. Recorded in plan.md § Settled decisions; zero markers remain.

## §E.2 Run-phase Evidence

### M1 — Shared diagnostic snapshot contract (2026-07-13, manager-develop cycle_type=tdd)

Scope delivered: `internal/verify/` package (schema/key/freshness/store/source, 5 files + 6 test files), `moai verify record|check` CLI (`internal/cli/verify.go` + root-tree registration), `internal/goal/evaluate.go` snapshot-aware Tier-1 path + `internal/cli/hook_stop_goal.go` production wiring, doctrine injection into gate.md / run/task-decomposition.md Phase 2.75 / sync/quality-gates-context.md Phase 0 / loop.md Steps 1·3·1.5, template mirrors (4 files, byte-parity) + `make build`.

TDD evidence: RED observed per surface (`go test ./internal/verify/...` build-failure on undefined package; `go vet ./internal/goal/` unknown-field `Snapshot`; `go vet ./internal/cli/` undefined `newVerifyCmd`) before each GREEN implementation.

#### M1 AC matrix (AC-WCO-001..011)

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-WCO-001 | PASS | `go test -run 'TestSnapshotSchema' ./internal/verify/...` | `ok github.com/modu-ai/moai-adk/internal/verify` — TestSnapshotSchema + TestSnapshotSchemaLoopVerdictDecode PASS (schema fields + loop-verdict conditions decode) |
| AC-WCO-002 | PASS | `go test -run 'TestSnapshotKey' ./internal/verify/...` | PASS — table: clean / dirty-edit / RE-EDIT-of-dirty (D13) / staged / add-untracked / HEAD-advance all yield distinct keys; same tree deterministic |
| AC-WCO-003 | PASS | `go test -run 'TestFreshness' ./internal/verify/...` | PASS — E2E: record → same-tree in-TTL accept → tracked-mutate stale, untracked-add stale, injected clock past 10-min default stale; configured TTL honored |
| AC-WCO-004 | PASS | `go run ./cmd/moai verify --help` | exit 0; COMMANDS lists `check` + `record` (root-tree registered; `TestVerifyRegisteredOnRoot` PASS) |
| AC-WCO-005 | PASS (M1 legs) | `grep -c 'moai verify' gate.md`=2, `grep -c '\.moai/state/verify/' gate.md`=1, `grep -c '\-\-fresh' gate.md`=2, all from verified-0 baselines; AC-WCO-004 PASS | gate.md Phase 1.5 consumption + Snapshot Recording + `--fresh` force-fresh mode |
| AC-WCO-006 | PASS | `grep -c 'moai verify' run/task-decomposition.md` 0 → 2; placement inside Phase 2.75 section (windowed read) | Snapshot consumption + recording paragraphs added to Phase 2.75 |
| AC-WCO-007 | PASS | quality-gates-context.md Phase 0 Step 0.0.1 windowed read | snapshot consumption + citation-as-evidence wording added (`grep -c 'moai verify'` 0 → 1 line) |
| AC-WCO-008 | PASS | Step-3 window `grep -c 'moai verify record'` 0 → 1; Step-1 window `grep -c '\.moai/state/verify/'` 0 → 1 | loop Step 3 mechanical writer + Step 1 shared-store read surface |
| AC-WCO-009 | PASS | Step 1.5 window: shall-NOT-consume sentence 0 → 1; `--fresh` 0 → 1 | independence carve-out + `/moai gate --fresh` invocation landed in Step 1.5 itself |
| AC-WCO-010 | PASS | `go test -run 'TestEvaluateSnapshot' ./internal/goal/...` | PASS — 7 tests: exact-match reuse (fake-runner call-count 0), miss/stale re-execution, near-miss variant executes, attribution in verdict payload, nil-source additive, reused-failure, constant lookups |
| AC-WCO-011 | PASS | `go test -run 'TestFreshness.*Stale' ./internal/verify/...` PASS; doctrine grep: VCI cites gate.md 0→1, task-decomposition 0→1, quality-gates-context 0→1, loop.md 2→3 | stale-never-reusable proven (partial key match + past-TTL both stale); path+key+command citation wording in all 4 files |

#### M1 quality gates

- `go test ./...` → exit 0, 100 packages ok, 0 FAIL (evidence: `.moai/state/verify/wco-m1/wco-full-test.log`)
- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (B1)
- `go vet ./...` → exit 0
- `golangci-lint run` → `0 issues.` (pre-flight baseline `0 issues.` → NEW = 0; logs: `wco-preflight-lint.log` / `wco-lint-post.log`)
- Coverage: `go test -cover ./internal/verify/...` → **86.6%** (≥85% threshold); `./internal/goal/...` → **87.4%** vs measured base-commit baseline **86.5%** (detached-worktree measurement at 13c74c9c3) — non-regressing
- Subagent-boundary grep (B3): 0 matches in `internal/verify` + `internal/goal` (+ per-package `TestSubagentBoundary` CI guards added)
- Advisory-Check Discipline (Custom-1): `verify.Source` memoizes key computation (1 per turn-end regardless of condition count — `TestSourceConstantCost` regression guard) + `DefaultLookupTimeout` 2s time-box with re-execution fallback (`TestSourceDeadlineFallback`)
- Template parity: gate.md / loop.md / run/task-decomposition.md / sync/quality-gates-context.md `cmp -s` local↔template all EQ; `make build` exit 0; `go test ./internal/template/...` exit 0; no internal SPEC IDs introduced into templates
- Settled-decision conformance: package name `internal/verify` (no collision — pre-flight verified `moai verify` top-level free; existing `verify` verb is nested under `harness`); key = HEAD SHA + sha256(porcelain-v2 ∥ NUL ∥ diff-HEAD)[:16]; TTL default 10 min (`verify.DefaultTTL`), configurable via `--ttl` flag / `Source.TTL`; snapshot layout `.moai/state/verify/snapshots/<head12>-<digest16>.json` atomic temp+rename

#### M1 notes / discovered behavior

- Test repos require `.moai/` gitignored (mirrors real project): the snapshot store writes under `.moai/state/`, which must be ignored so recording does not invalidate the key it was recorded under (porcelain-v2 lists untracked non-ignored paths). Documented in `initTestRepo` helpers.

### M2-M5 — Gate merging / audit improvement / delegation relaxation / bookkeeping batching (2026-07-13, manager-develop cycle_type=tdd doc-milestone variant)

Executed in an isolated worktree fast-forwarded to 2d55ee6f3 (M1 base); per-milestone commits, no push (orchestrator integrates). Every AC baseline grep was re-measured BEFORE editing (all matched acceptance.md recorded baselines — no mismatch blocker).

Scope delivered per milestone:

- **M2 (`c7f8febad`)** — moai.md Steps 11.3+11.5 single-call merge + full-pipeline completion close; codebase-analysis.md Stage B single multi-question call (L152 [HARD] SEPARATE-scope-boundary text preserved); harness-build-entry.md Phase 1.6 proposal folded into the Phase 3 approval round; feedback.md single collection round; sync/delivery.md Phase 4 full-pipeline suppression + "(Recommended)" chain. 5 local + 5 mirror files.
- **M3 (`e47ff0210`)** — run.md + run/phase-execution.md Tier S single-audit-pass default ("audit always runs (once)"; SKIP-불가 wording preserved; new § Tier S Single-Audit-Pass Default + § FAIL Defect-List Delta Re-Check); plan-auditor.md + sync-auditor.md structured defect-list output contract (finding id / location / severity / required fix); moai.md Pipeline Gates #4 sync-audit delta re-check; doc-generation.md Phase 3.1 retry cap 3→1. Agent-file mirror via sanitized-parity (sync-auditor HRN-003 line); catalog.yaml hash regen cascade (gen-catalog-hashes --all) for the 2 edited agent bodies.
- **M4 (`490a9a0b6`)** — fix.md Level-1 orchestrator-direct formatter + mandate re-scoped Level 2+ (static dispatch shape kept; `TestAgentlessUtilityNoLLMControlFlow` PASS); loop.md Step 6 same exception; codemaps.md 3→1 spawn (Phase 2/3 orchestrator-direct); clean.md 4→2 combined spawns + Phase 5.5 orchestrator-direct-only; mx.md <5-item orchestrator-direct Pass 3.
- **M5 (`f8a00d3ac`)** — loop.md Steps 5/6/7 one batched bookkeeping turn per iteration (TaskCreate + `in_progress` + `completed` aggregated) + Step 7.5/§ MX Tag Report aggregate-at-exit MX_TAG_REPORT; review.md Phase 2 Mode-4 parallel 4-lens fan-out (sync-auditor stays binding synthesis/verdict owner) + incremental secrets scan (`.moai/state/secrets-scan-checkpoint.txt`, `<last-sha>..HEAD`, full `--all` first-run/flag fallback).

#### M2-M6 AC matrix (AC-WCO-012..036; M1 ACs 001..011 re-confirmed post-edit)

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-WCO-012 | PASS | `grep -c 'single AskUserQuestion call carrying both' moai.md`; `grep -c 'score-independent'`; `grep -c 'exactly once per pipeline entry'` | 0→1; preserved 1; preserved 2 |
| AC-WCO-013 | PASS | windowed `sed -n '/### Stage B/,/## Stage B Round 4/p'`: separate-per-axis phrase 1→0; `single AskUserQuestion call` 0→1; PRESERVE `as a SEPARATE AskUserQuestion` 1→1 | all legs green |
| AC-WCO-014 | PASS | `grep -c 'instead of two sequential rounds' harness-build-entry.md` 0→2; standalone-offer remnant "last question the interview asks" 1→0 | merged round; two-round flow removed |
| AC-WCO-015 | PASS | `grep -c 'NO manufactured next-step question'` moai.md 0→1 + delivery.md Phase-4 window 0→1; "(Recommended)" moai.md 5, delivery.md Phase-4 window 4 | both surfaces; single-phase chain kept |
| AC-WCO-016 | PASS | feedback.md Phase-1 window: `ONE AskUserQuestion round` 0→1; per-item "Solicit … via AskUserQuestion" 2→0 | 3 sequential rounds → 1 |
| AC-WCO-017 | PASS | inventory greps: (a) moai.md merged round names Kickoff (run-phase entry/review/abort) + shape (worktree/sub-agent); (b) Stage B window verification=2 external_systems=2 ui_surface=1 team_sharing=1; (c) proposal+approval both present; (d) type=4 title=3 description=3 | 4 surfaces, all tokens present |
| AC-WCO-018 | PASS | fix.md `orchestrator-direct formatter command` 0→2 + `Level 2 and above` 0→1; `go test -run TestAgentlessUtilityNoLLMControlFlow ./internal/template/` | PASS (4/4 subtests) |
| AC-WCO-019 | PASS | loop.md Step 6 Level-1 exempt clause 0→1 | orchestrator-direct exception landed |
| AC-WCO-020 | PASS | codemaps.md `Delegate.*to the manager-docs subagent` 2→0; Agent Chain lists Phase 1 Explore as the single Agent() spawn | 3 spawns → 1 |
| AC-WCO-021 | PASS | clean.md Agent Chain lists exactly 2 spawn phases (1+2 combined, 4+5 combined); Phase 5.5 window "or a per-spawn" 1→0 | 4 spawns → 2; 5.5 orchestrator-direct only |
| AC-WCO-022 | PASS | mx.md `fewer than 5 items` 0→2 (Pass 3 rule + Agent Chain) | <5-item direct clause landed |
| AC-WCO-023 | PASS | preserve greps: fix.md Level-3 AskUserQuestion approval 1; clean.md removal-plan AskUserQuestion 1; @MX:ANCHOR protection 1 | present→present |
| AC-WCO-024 | PASS | run.md `single audit pass` 0→1 + `audit always runs (once)` 0→1 + `SKIP 불가` preserved 1; phase-execution § Tier S Single-Audit-Pass Default 0→1 + `Never skipped` preserved 1 | no skip-audit wording introduced |
| AC-WCO-025 | PASS | `Required fix` plan-auditor 0→2, sync-auditor 0→1; `defect-list` 0→2 each; delta re-check flow phase-execution 0→1 + moai.md 0→1 | output contracts + owning-body flows |
| AC-WCO-026 | PASS | doc-generation.md `max 1 retry` heading; `iteration = 3` 1→0; `max 3 iterations` 1→0; escalation at iteration=2 | retry cap 3 → 1 |
| AC-WCO-027 | PASS | `never substitutes an orchestrator self-assessment` pa=1 sa=1 phase-execution=1 moai.md=1 | verdict authority preserved, 4 surfaces |
| AC-WCO-028 | PASS | loop.md per-fix [HARD] TaskUpdate lines 2→0; `ONE batched turn per iteration` 0→1 | 3-calls-per-issue → 1 batch/iteration |
| AC-WCO-029 | PASS | `After each iteration, include MX_TAG_REPORT` 1→0; Step 7.5 `Generate MX_TAG_REPORT` 1→0; `aggregated across all iterations` 0→2 | aggregate-at-exit |
| AC-WCO-030 | PASS | review.md `parallel read-only fan-out` 0→1; `sequentially` 1→0; `binding synthesis and verdict owner` 0→1 | Mode-4 lenses; sync-auditor verdict kept |
| AC-WCO-031 | PASS | `.moai/state/secrets-scan-checkpoint` 0→1; `<last-sha>..HEAD` 0→1; full `git log -p --all -G` retained 1 | incremental + fallback |
| AC-WCO-032 | PASS | inventories: (a) TaskCreate/in_progress/completed 3/3 in Step-5 batch; (b) Tags Added/Removed/Updated 3/3; (c) 4 perspective headings 4/4; (d) PRIVATE KEY/AKIA/ghp_ 2 each | no record loss, 4 surfaces |
| AC-WCO-033 | PASS | `grep -c 'score-independent'` vs 2d55ee6f3 baseline: run.md 1→1, moai.md 1→1 (non-decreasing); no edited file conditions Kickoff on score | invariant intact |
| AC-WCO-034 | PASS | per-surface AskUserQuestion greps: moai.md 11, codebase-analysis 7, harness-build-entry 15, feedback 3, delivery 4 | channel monopoly intact; no prose-question instruction introduced |
| AC-WCO-035 | PASS | VCI cites: gate.md 1 (0→1 at M1), task-decomposition 1, quality-gates-context 1, loop.md 3 (baseline 2→3) | per-file delta matched |
| AC-WCO-036 | PASS | 19-surface sweep: 17 workflow + plan-auditor `cmp -s` ALL EQ; sync-auditor sanitized-diff empty (HRN-003 transform); `grep -rn 'SPEC-V3R2-HRN-003' internal/template/templates/` = 0; `make build` exit 0; `go test ./internal/template/...` exit 0 | split parity + neutrality + build green |
| AC-WCO-001..011 | PASS (re-confirmed) | `go test -run 'TestSnapshotSchema\|TestSnapshotKey\|TestFreshness' ./internal/verify/...` ok; `TestEvaluateSnapshot ./internal/goal/...` ok; `moai verify --help` exit 0 lists record+check; doctrine legs re-grepped after M4/M5 loop.md edits (Step-3 `moai verify record` 1, Step-1 `.moai/state/verify/` 1, Step-1.5 `--fresh` 1 + shall-NOT 1) | M1 intact after M2-M5 edits |

#### M6 quality gates (verification sweep)

- `go test ./...` → exit 0, 100 packages ok, 0 FAIL (`.moai/state/verify/wco-m2m6/m6-full-test.log`)
- `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (B1)
- `golangci-lint run` → `0 issues.` (pre-flight baseline `0 issues.` → NEW = 0; logs under `.moai/state/verify/wco-m2m6/`)
- `make build` exit 0 (catalog hash regen idempotent — no diff after M3's committed regen)
- `moai spec lint spec.md --strict` → 0 errors, 1 warning: the known `StatusGitConsistency` (in-progress vs git-implied implemented) — expected mid-run state
- Subagent-boundary grep (B3/E4): 0 matches in `internal/verify` + `internal/goal`
- Coverage re-confirmation: `internal/verify` 86.6%, `internal/goal` 87.4% (unchanged from M1 — no Go edits in M2-M5)
- Template neutrality: 0 internal-SPEC-ID matches under `internal/template/templates/`

#### M2-M6 notes / discovered behavior

- catalog.yaml hash-guard cascade: editing agent bodies (plan-auditor/sync-auditor) trips `TestManifestHashFormat` (CATALOG_HASH_UNSTABLE) — resolved by `go run ./internal/template/scripts/gen-catalog-hashes.go --all` in the same milestone commit (M3).
- Parallel-session caution: during this run a parallel session renamed phase numbering ("Phase 0.5"→"Phase 1", "Phase 0.95"→"Phase 4") in MAIN-checkout rule files (`spec-workflow.md`, `orchestration-mode-selection.md`, `spec-frontmatter-schema.md`). This worktree's edits are against the 2d55ee6f3 baseline vocabulary ("Phase 0.5"/"Phase 0.95" in run.md/moai.md/phase-execution.md). Merge-time reconciliation may be needed if that rename also sweeps the workflow skill bodies.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-13
run_commit_sha: pending-backfill-m6   # M6 commit cannot self-reference; backfill = the commit carrying this §E.3
run_status: complete
ac_pass_count: 36
ac_fail_count: 0
preserve_list_post_run_count: 8   # plan.md §D PRESERVE list — all 8 classes verified untouched (Kickoff wording, askuser-protocol cite-only, VCI cite-only, sync-phase hook, cadence-bridge, loop-verdict schema names, Tier M/L ceilings, run.md Run-phase Autonomy body)
l44_pre_commit_fetch: n/a-worktree   # isolated worktree; push + origin fetch owned by the orchestrator at integration
l44_post_push_fetch: n/a-worktree    # no push performed by this agent (per spawn contract)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  native: exit 0
  windows_amd64: exit 0
total_run_phase_files: 33   # M2-M5 diff vs 2d55ee6f3: 16 local surfaces + 16 template mirrors + catalog.yaml (+ progress.md in the M6 commit)
m1_to_mN_commit_strategy: per-milestone commits — M1 2d55ee6f3 (merged), M2 c7f8febad, M3 e47ff0210, M4 490a9a0b6, M5 f8a00d3ac, M6 (this commit); no --amend, no --no-verify, no push
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

- plan_complete_at: 2026-07-13T05:38:51Z
- plan_status: audit-ready

## §F Phase 0.95 Mode Selection

- Inputs: tier=L, scope=Go new pkg + 19 doc mirrors, domains=3 (Go/workflow-doc/rules), lang mix=Go+markdown, concurrency benefit=LOW (coding-heavy)
- Evaluation: trivial=no(semantic) / background=no(write) / agent-team=RETIRED / parallel=no(coding-heavy per Anthropic caveat) / workflow=no(not uniform-mechanical, inter-file deps) / sub-agent=SELECTED
- Decision: sub-agent
- Justification: M1 is new Go code (snapshot engine + CLI + consumers) — coding-heavy work stays sequential per Anthropic's coding-task parallelism caveat. M2-M5 doc edits have cross-file consistency requirements (template mirrors) that a single sequential agent handles more safely than fan-out.
- Kickoff: Implementation Kickoff Approval obtained (user selected run-phase entry); preferences drained.
