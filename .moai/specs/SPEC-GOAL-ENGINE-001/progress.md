# progress.md — SPEC-GOAL-ENGINE-001

> Canonical §E section skeleton. Plan-phase populates §E.1 only; §E.2/§E.3 are
> owned by manager-develop (run-phase), §E.4 by manager-docs (sync-phase).

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready (v0.2.1 — 2 D2 fixes from plan-auditor v0.2.0 audit applied; pending re-audit)
- plan_complete_at: 2026-07-12
- tier: L (LEAN: 3 core artifacts + progress.md; design folded into plan.md § Technical Design; research.md shared from SPEC-ANALYZE-FIRST-ROUTING-001)
- artifacts: spec.md, plan.md, acceptance.md, progress.md
- REQ count: 29 (REQ-GLE-001..025 + REQ-GLE-026..029 added v0.2.0 amendment D8; no new REQs in v0.2.1 — REQ-GLE-010/028 reworded in-place for the D2-1 reconciliation)
- AC count: 34 (AC-GLE-001..026 + AC-GLE-027..034 added v0.2.0 amendment D8; no new ACs in v0.2.1 — AC-GLE-021(a) re-anchored and AC-GLE-029 amended in-place for the D2 fixes)
- depends_on: SPEC-ANALYZE-FIRST-ROUTING-001
- v0.2.1 changes (plan-auditor v0.2.0 D2 fixes): D2-1 = enrich §B.5 checkpoint JSON with `failed_conditions: [{cmd, exit, tail}]` + reconcile REQ-GLE-010 ↔ REQ-GLE-028 (failed-condition+tail present in BOTH modes) + amend AC-GLE-029 to assert `failed_conditions`; D2-2 = re-anchor AC-GLE-021(a) from stale `grep -ic "goal evaluator\|goal engine" CLAUDE.md` (baseline 1, non-discriminating) to `awk '/^## 2\./,/^## 3\./' CLAUDE.md | grep -ic "goal evaluator"` (verified baseline 0, discriminating).
- open decisions: 0 remaining — all 4 iteration-2 decisions resolved + 2 D8 amendment decisions resolved (progression-mode axis = kickoff-time choice NOT gate bypass; semi-autonomous confirm via orchestrator-bridge NOT hook prompt). See plan.md Settled Decisions. v0.2.1 pending plan re-audit per the orchestrator.

### Deferred to run-phase (plan-auditor D3, v0.2.0 audit)

2 cosmetic/alignment D3 defects deferred to run-phase per orchestrator directive
(NOT fixed in v0.2.1 — D2 fixes only this iteration):

- **D3-1** — AC-GLE-032/033 use a 2-alternative OR-regex
  (`semi-autonomous|progression.mode`) while AC-GLE-031 uses a single token
  (`semi-autonomous`) (`acceptance.md` AC-GLE-032 ~line 351, AC-GLE-033 ~line
  362 vs AC-GLE-031 ~line 338). Align the 3 doc-surface reachability ACs to a
  consistent token shape in run-phase.
- **D3-2** — AC-GLE-029/030 detail-block headers use `REQ-GLE-028a`/`REQ-GLE-028b`
  sub-clause notation while the §D matrix row uses `REQ-GLE-028`
  (`acceptance.md` AC-GLE-029 ~line 307, AC-GLE-030 ~line 319 vs §D matrix
  ~line 39-40). Cosmetic header notation alignment in run-phase.

## §E.2 Run-phase Evidence

M1–M7 implemented (TDD, cycle_type=tdd). Files created/modified:

- NEW `internal/goal/schema.go`, `state.go`, `prune.go`, `evaluate.go` + tests (`state_test.go`, `prune_test.go`, `evaluate_test.go`).
- NEW `internal/cli/hook_stop_goal.go` (`moai hook stop-goal` verb; registered under `hookCmd`).
- NEW `.claude/hooks/moai/handle-stop-goal.sh` wrapper + template mirror `handle-stop-goal.sh.tmpl`.
- NEW `.claude/skills/moai/workflows/goal.md` (4 verbs + progression-mode + checkpoint/orchestrator-bridge) + template mirror.
- EXTEND `.claude/skills/moai/SKILL.md` (P1 `**goal**` + Quick Reference `### goal`) + template mirror.
- EXTEND `internal/template/templates/.claude/settings.json.tmpl` Stop-hook COMPOSE (handle-stop-goal.sh entry, timeout 120; handle-stop.sh preserved).
- EXTEND doctrine: `goal-directive.md` (`/moai goal` row + Axis B), `native-invocation-model.md` (Axis B illustration), `session-handoff.md` (Block 5 `/moai goal` variant), `moai.md` (phase-granular vs task-granular boundary), `run.md` + `orchestration-mode-selection.md` (progression-mode axis), `CLAUDE.md` §2 stage ⑤ (goal evaluator) + §2 stage ④ (progression-mode axis) — all mirrored to template.
- EXTEND `internal/cli/hook_test.go`, `hook_pre_push_test.go` (subcommand count 36→37), `hook_e2e_test.go` (utilitySubcmds +stop-goal).

### AC PASS/FAIL matrix (AC-GLE-001..034)

All 34 ACs PASS. Verification commands run in the worktree (this run, against this tree):

| AC | Status | Verification |
|----|--------|--------------|
| 001 | PASS | `test -f .claude/skills/moai/workflows/goal.md` (file present, 4 verbs documented) |
| 002 | PASS | `awk '/^### Priority 1/,/^### Priority 2/' SKILL.md \| grep -c '\*\*goal\*\*'` → 1 |
| 003 | PASS | `awk '/^## Workflow Quick Reference/,0' SKILL.md \| grep -c '^### goal'` → 1 |
| 004 | PASS | `go test ./internal/goal/ -run TestStatePathPerSession` → ok |
| 005 | PASS | `go test ./internal/goal/ -run TestSchemaFields` → ok |
| 006 | PASS | `go test ./internal/goal/ -run TestAtomicWrite` → ok |
| 007 | PASS | `go test ./internal/goal/ -run TestOrphanPrune` → ok |
| 008 | PASS | `go test ./internal/goal/ -run TestWriterPidFallback` → ok |
| 009 | PASS | `echo '{}' \| go run ./cmd/moai hook stop-goal` → exit 0 |
| 010 | PASS | `go test ./internal/goal/ -run TestTier1Block` → ok |
| 011 | PASS | `go test ./internal/goal/ -run TestTier2Gate` → ok |
| 012 | PASS | `go test ./internal/goal/ -run TestAllPassNoBlock` → ok |
| 013 | PASS | `go test ./internal/goal/ -run TestCeilingVerdict` → ok (5 section names present) |
| 014 | PASS | boundary grep `AskUserQuestion\|mcp__askuser` on internal/goal/ + hook_stop_goal.go + handle-stop-goal.sh (excl _test.go + comments) → 0 |
| 015 | PASS | `go test ./internal/goal/ -run TestNoKickoffBypass` → ok; goal.md states no-bypass |
| 016 | PASS | `go test ./internal/goal/ -run TestNativeGoalYield` → ok |
| 017 | PASS | `go test ./internal/goal/ -run TestStagnationStop` → ok (E1/E3 note in verdict) |
| 018 | PASS | `grep -ic "/moai goal" session-handoff.md` → 1 |
| 019 | PASS | `grep -ic "/moai goal" goal-directive.md` → 2; `grep -ic "Axis B"` → 2 |
| 020 | PASS | `grep -ic "/moai goal" native-invocation-model.md` → 1 |
| 021 | PASS | (a) `awk '/^## 2\./,/^## 3\./' CLAUDE.md \| grep -ic "goal evaluator"` → 1 (both local + template); (b) `grep -ic "task-granular\|phase-granular\|goal engine" moai.md` → 1 |
| 022 | PASS | `go test ./internal/config/ -run TestAgentic` → ok (distinctness guard green) |
| 023 | PASS | `ls schema.go state.go prune.go evaluate.go hook_stop_goal.go` → 5 files |
| 024 | PASS | `go test -cover ./internal/goal/` → coverage: 86.5% of statements (≥85) |
| 025 | PASS | `test -f templates/.claude/skills/moai/workflows/goal.md` (MIRROR_OK); neutrality grep → 0; `make build` → exit 0 |
| 026 | PASS | `grep -c 'handle-stop\.sh' settings.json.tmpl` → 2 (preserved); `grep -c 'handle-stop-goal\.sh'` → 2 (added) |
| 027 | PASS | `grep -ic "semi-autonomous" goal.md` → 8 |
| 028 | PASS | `go test ./internal/goal/ -run TestAutonomousModeNoCheckpoint` → ok |
| 029 | PASS | `go test ./internal/goal/ -run TestSemiAutonomousCheckpointSignal` → ok (mode/failed_conditions present) |
| 030 | PASS | `grep -ic "checkpoint" goal.md` → 6; `grep -ic "orchestrator" goal.md` → 10 |
| 031 | PASS | `grep -ic "semi-autonomous" CLAUDE.md` → 1 |
| 032 | PASS | `grep -ic "semi-autonomous\|progression.mode" run.md` → 1 |
| 033 | PASS | `grep -ic "semi-autonomous\|progression.mode" orchestration-mode-selection.md` → 1 |
| 034 | PASS | (a) `grep -ic "both.mode\|in both modes" goal.md` → 2; (b) `go test ./internal/goal/ -run TestKickoffMandatoryBothModes` → ok |

## §E.3 Run-phase Audit-Ready Signal

- run_complete_at: 2026-07-12
- run_commit_sha: pending-backfill (single run-phase commit; SHA backfilled post-land)
- run_status: audit-ready (34/34 AC PASS)
- ac_pass_count: 34
- ac_fail_count: 0
- preserve_list_post_run_count: 4 (run.md ac_converge section, agentic_loop_distinctness_test.go, internal/ralph+internal/loop+internal/cli/loop.go, existing handle-stop.sh Stop-hook entry)
- l44_pre_commit_fetch: n/a (worktree-isolated L1; no remote fetch required)
- l44_post_push_fetch: n/a (no push — commits left local per orchestrator directive)
- new_warnings_or_lints_introduced: 0 (`golangci-lint run` → 0 issues; `go vet ./...` clean)
- cross_platform_build:
  - `go build ./...` → exit 0
  - `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- total_run_phase_files: 27 (4 new Go source + 3 new Go test + 1 new CLI verb + 1 new wrapper + 2 new wrapper/workflow template + 1 new workflow + 6 modified doctrine × 2 (local+template) + 1 CLAUDE.md × 2 + 1 settings.json.tmpl + 3 modified CLI tests + 1 catalog.yaml)
- m1_to_mN_commit_strategy: single run-phase commit carrying draft→in-progress frontmatter transition (M1 ownership) + the full M1–M7 implementation + progress.md §E.2/§E.3 evidence
- subagent_boundary_grep: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/goal/ internal/cli/hook_stop_goal.go .claude/hooks/moai/handle-stop-goal.sh | grep -v _test.go | grep -v '^[^:]*:[0-9]*:[ \t]*//'` → 0 (REQ-GLE-014 preserved)
- spec_lint: `moai spec lint spec.md` → 0 errors (1 StatusGitConsistency warning resolves on this commit landing the draft→in-progress transition)
- template_neutrality: `grep -rn 'SPEC-GOAL-ENGINE\|SPEC-ANALYZE-FIRST\|AGENTIC-CORE\|REQ-GLE' internal/template/templates/.claude/` → 0

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — owned by manager-docs>_
