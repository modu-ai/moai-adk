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
- Milestones M2 (gate merging), M3 (audit), M4 (delegation), M5 (bookkeeping), M6 (verification sweep) pending.

## §E.3 Run-phase Audit-Ready Signal

_<pending — run-phase in progress: M1 complete (AC-WCO-001..011 PASS), M2-M6 outstanding>_

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
