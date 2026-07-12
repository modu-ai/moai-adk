# SPEC-HARNESS-EVOLVE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-12
plan_version: 0.1.2   # iter-1 FAIL 0.75 → D1-D4+S1-S6; iter-2 PASS 0.89 → D-1 + N-1/N-2 amendment
tier: M
artifacts: [spec.md, plan.md, acceptance.md, progress.md]
req_count: 21
ac_count: 27          # 34 verification line items (022a-d, 023a-e)
open_clarifications: 0   # all resolved — 3 pinned user decisions (plan.md §H)
```

Plan-phase notes:
- Design SSOT: `.moai/reports/harness-self-evolving-redesign-final-20260712.html`
  (§4 3-Zone, §5 Loop 0 schema + A2/A4 deltas, §7 M1).
- Plan-audit iter-1: FAIL 0.75 (`.moai/reports/plan-audit/SPEC-HARNESS-EVOLVE-001-review-1.md`).
  v0.1.1 applies D1 (HOI dual-gate reconciliation — REQ-HEV-016 rewrite +
  spec.md §D.3 activation precondition + AC-HEV-021/Scenario 1/5/6 fixtures),
  D2 (CLI registration re-pinned to `newHarnessRouterCmd()`/harness_route.go +
  `v3r5RequiredHarnessVerbs` step + AC-HEV-027), D3 (AC-HEV-011 write-surface
  re-scope), D4 (§H markers struck), S1-S6.
- Resolved clarifications (AskUserQuestion round 2026-07-12, pinned in
  plan.md §H): (1) KEEP HOI-gated transport, activation precondition codified,
  no new gate / no default flip; (2) `request_class` INCLUDED in schema v1
  (coarse keyword enum, non-verbatim); (3) v1 no-rotation, retention deferred
  to EVOLVE-003 with `retention.go` reuse preserved.
- Gate note (spec.md §D.3): Stop-path outcome finalization requires
  `hook.opt_in.enabled: true` (fail-closed, default OFF per
  SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001) THEN `learning.enabled` (fail-open,
  default true). Default-config Stop-path dormancy is EXPECTED shipped
  behavior; this dev repo enables the opt-in locally during M4.
- AC baselines measured against main working tree 2026-07-12 (all
  discriminating tokens at 0 on registration surfaces; `.moai/state/`
  gitignore rule pre-existing; `"ledger"` in harness_retirement_test.go = 0;
  template workflow mirrors routing-ledger = 0 ×3).
- Pinned decisions D1-D6 in plan.md §A.2 (HOI-gated Stop-hook transport
  reuse, hook-side finalization authority, multi-turn pending, dual-gate
  inheritance, per-session pending isolation, best-effort identity
  resolution).

## §E.2 Run-phase Evidence

Milestones M1-M4 complete (cycle_type=tdd). All 27 AC / 34 verification line
items PASS. Evidence gathered against this tree (worktree branch, FF'd to local
main tip; commits stacked there — orchestrator integrates to main per B9
exception, parallel-session race). Verify-log dir: `/tmp/hev/`.

| AC | Status | Verify command | Actual output |
|----|--------|----------------|---------------|
| AC-HEV-001 | PASS | `go test ./internal/harness/routing/...` | `ok  .../internal/harness/routing` |
| AC-HEV-002 | PASS | `go test -cover ./internal/harness/routing/` | `coverage: 91.6% of statements` (≥90) |
| AC-HEV-003 | PASS | `go test -race -run TestConcurrentAppend ./internal/harness/routing/` | `ok` (race-clean) |
| AC-HEV-004 | PASS | `grep -rl "routing-ledger.jsonl" internal/harness/routing/*.go \| wc -l` | `1` (types.go LedgerFileName const) |
| AC-HEV-005 | PASS | `grep -c "matched_subcommand" internal/harness/routing/types.go` | `2` (≥1) |
| AC-HEV-006 | PASS | `grep -c "convergence_class" internal/harness/routing/types.go` | `2` (≥1) |
| AC-HEV-007 | PASS | `grep -c "delegations" internal/harness/routing/types.go` | `2` (≥1) |
| AC-HEV-008 | PASS | `go test -run TestConvergenceNullWhenNoSignal ./internal/harness/routing/` | PASS (null when no signal) |
| AC-HEV-009 | PASS | `go test -run TestRequestDigestNoVerbatim ./internal/harness/routing/` | PASS (digest `sha256:[0-9a-f]{12}`, no verbatim) |
| AC-HEV-010 | PASS | `go test -run TestDeriveOutcome ./internal/harness/routing/` | PASS (abort>fail>success>pending, no override) |
| AC-HEV-011 | PASS | `go run ./cmd/moai harness ledger {record,evidence} --help \| grep -c -- '--outcome'` | `0` AND `0` (write surfaces) |
| AC-HEV-012 | PASS | `grep -rn "usage-log" internal/harness/routing/ \| wc -l` | `0` |
| AC-HEV-013 | PASS | `grep -rn "harness\.Event\b" internal/harness/routing/ \| wc -l` | `0` |
| AC-HEV-014 | PASS | `go test -run TestReroute ./internal/harness/routing/` | PASS (same-session reroute) |
| AC-HEV-015 | PASS | `go test -run TestFinalize_SelfGate_NoPendingNoOp ./internal/harness/routing/` | PASS (no pending → no-op) |
| AC-HEV-016 | PASS | `go test -run TestFinalize_NonTerminalStaysPending ./internal/harness/routing/` | PASS (multi-turn stays pending) |
| AC-HEV-017 | PASS | `go test -run TestStaleSweepAbort ./internal/harness/routing/` | PASS (age-guard a/b/c + liveness guard) |
| AC-HEV-018 | PASS | `go test -run TestFinalize_FailOpen ./internal/harness/routing/` | PASS (inject write fail → sink+nil, pending kept) |
| AC-HEV-019 | PASS | `grep -c "harness/routing" internal/cli/hook.go` | `1` (≥1, finalizer wired) |
| AC-HEV-020 | PASS | `go run ./cmd/moai harness ledger --help; echo exit=$?` | `exit=0`, lists record/evidence/list |
| AC-HEV-021 | PASS | `go test -run TestHarnessObserveStop_RoutingLedgerGated ./internal/cli/` | PASS (HOI off/on × learning matrix) |
| AC-HEV-022a | PASS | `grep -c "routing-ledger" .claude/skills/moai/SKILL.md` | `1` (≥1) |
| AC-HEV-022b | PASS | `grep -c "routing-ledger" .claude/skills/moai/workflows/plan.md` | `1` (≥1) |
| AC-HEV-022c | PASS | `grep -c "routing-ledger" .claude/skills/moai/workflows/run.md` | `1` (≥1) |
| AC-HEV-022d | PASS | `grep -c "routing-ledger" .claude/skills/moai/workflows/sync.md` | `1` (≥1) |
| AC-HEV-023a | PASS | `grep -c "routing-ledger" internal/template/.../moai/SKILL.md` | `1` (≥1) |
| AC-HEV-023b | PASS | `grep -c "SPEC-HARNESS-EVOLVE" internal/template/.../moai/SKILL.md` = `0` + `TestTemplateNeutralityAudit`/`TestTemplateNoInternalContentLeak` green | `0`, guards PASS |
| AC-HEV-023c | PASS | `grep -c "routing-ledger" internal/template/.../workflows/plan.md` | `1` (≥1) |
| AC-HEV-023d | PASS | `grep -c "routing-ledger" internal/template/.../workflows/run.md` | `1` (≥1) |
| AC-HEV-023e | PASS | `grep -c "routing-ledger" internal/template/.../workflows/sync.md` | `1` (≥1) |
| AC-HEV-024 | PASS | `find internal/template/templates \( -name "routing-ledger*" -o -name "routing-pending*" \) \| wc -l` | `0` |
| AC-HEV-025 | PASS | `go test ./...` = 0 FAIL; `go build ./...` + `GOOS=windows go build ./...` exit 0 | all exit 0, no existing test modified |
| AC-HEV-026 | PASS | `git check-ignore -v .moai/state/routing-ledger.jsonl` | `.gitignore:265:.moai/state/` exit 0 (no touch) |
| AC-HEV-027 | PASS | `grep -c '"ledger"' internal/cli/harness_retirement_test.go` = `1` + `TestHarnessV3R5VerbSurface` PASS | `1`, guard PASS |

Deferred AC: none. All 27 AC PASS at run-phase.

@MX tags: `@MX:ANCHOR` on `writer.go` `Writer.Append` (fan-in ≥3: Store
reroute/sweep/finalize + CLI + hook), `@MX:NOTE` on `outcome.go` `DeriveOutcome`
(fixed-precedence, no override).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-07-12
run_commit_sha: pending-backfill-run   # worktree branch commits; orchestrator integrates + backfills
cycle_type: tdd
ac_pass_count: 27          # 34 verification line items (022a-e, 023a-e)
ac_fail_count: 0
ac_deferred_count: 0
routing_coverage_pct: 91.6   # go test -cover ./internal/harness/routing/ (target ≥90)
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues on routing + cli (baseline was 0)
subagent_boundary_grep: 0     # AskUserQuestion/mcp__askuser in routing/ + hook.go non-test = 0
cross_platform_build:
  linux_darwin: pass          # go build ./... exit 0
  windows: pass               # GOOS=windows GOARCH=amd64 go build ./... exit 0
full_suite: pass              # go test ./... exit 0, 0 FAIL
spec_lint: clean              # moai spec lint spec.md → no findings
template_neutrality: green    # TestTemplateNeutralityAudit + TestTemplateNoInternalContentLeak
hoi_dogfood_enable: true       # M4 N-1: system.yaml hook.opt_in.enabled false→true (committed; template default stays false)
total_run_phase_files: 15      # 6 routing src + 5 routing test + harness_ledger.go + _test.go + hook.go + harness_route.go + harness_retirement_test.go + 4 skill + 4 template mirror + catalog.yaml + system.yaml + spec.md + progress.md
m1_to_mN_commit_strategy: per-milestone   # M1 core (+style), M2 CLI+hook, M3 skills, M4 verify — 5 commits on worktree branch
l44_pre_commit_fetch: n-a-worktree-branch  # commits on worktree branch (ancestor of local main); orchestrator owns origin push (B9 exception a)
l44_post_push_fetch: n-a-deferred-to-orchestrator
```

Residual risk: the Stop-path `success`/`fail` finalization is exercised by
`TestHarnessObserveStop_RoutingLedgerGated` (unit) but not by a live Claude Code
Stop hook this run (dormant without an actual session-end event); the HOI
dogfood-enable arms it for future live sessions. The worktree branch is a clean
FF-descendant of local main `3a3a7e56c`; local main is 5 unpushed plan commits
ahead of `origin/main` (other SPECs) — origin push is the orchestrator's call.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-07-12
sync_commit_sha: f12ebc7eae51f17674301fa711a87234d5744a58
changelog_entry_position: Added (top of [Unreleased])
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"
  plan_md: "n/a (no frontmatter status field — Tier M convention, status lives only in spec.md)"
  acceptance_md: "n/a (no frontmatter status field — Tier M convention, status lives only in spec.md)"
readme_sync: skipped   # internal self-evolving-harness observation infra, dormant by default, no user-facing workflow change
ac_count_verified: 27   # cross-checked against acceptance.md SSOT (grep -oE 'AC-HEV-[0-9]+' unique count)
```

Sync-phase notes:
- CHANGELOG dedup pre-check: `grep -c 'HARNESS-EVOLVE-001' CHANGELOG.md` = 0 before emission (no parallel-session duplicate).
- CHANGELOG entry drafted from direct `Read` of the actual implementation files
  (`internal/harness/routing/*.go`, `internal/cli/harness_ledger.go`,
  `internal/cli/hook.go` grep) — not from plan.md prose alone.
- README: reviewed, skipped. The new `moai harness ledger` CLI verbs are
  internal harness self-evolution observation infrastructure, gated dormant by
  default (HOI dual-gate); no documented user-facing workflow changes.

## §F Phase 0.95 Mode Selection

Input parameters: tier=M; scope≈8-10 files (internal/harness/routing/ new pkg +
internal/cli/ ledger cmd + hook.go finalizer + SKILL.md + 3 workflow bodies +
6 template mirrors); domain count=2 (Go source + doc/skill wiring); file
language mix=Go-heavy + markdown; concurrency benefit=LOW (coding-heavy).

Mode evaluation:
- Mode 1 (trivial): not selected — multi-file semantic implementation.
- Mode 2 (background): not selected — write-heavy, foreground required.
- Mode 3 (agent-team): not selected — RETIRED tombstone.
- Mode 4 (parallel): not selected — coding-heavy, not research fan-out
  (Anthropic coding-task parallelism caveat).
- Mode 5 (sub-agent): SELECTED — sequential manager-develop per milestone.
- Mode 6 (workflow): not selected — not ≥30-file uniform mechanical transform;
  this is new-code TDD work.

Decision: sub-agent

Justification: Tier M coding-heavy, single primary domain (the routing Go
package) with dependent milestones (M1 core → M2 CLI+hook depends on M1 types →
M3 wiring). Sequential sub-agent (Mode 5) is the Anthropic-recommended default
for coding work; the milestone dependency chain forbids parallel fan-out.
cycle_type=tdd (greenfield package, ≥90% coverage target).
