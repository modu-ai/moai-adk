# progress.md — SPEC-GOAL-HTML-WIRING-001

> Run-phase evidence is populated by manager-develop (§E.2/§E.3) and manager-docs (§E.4). manager-spec emits only the §E.1 plan-phase signal and the §E.2-§E.4 placeholder headings.

## §A. Branch + Tier + Mode

- **Tier**: M (3 counted artifacts: spec.md + plan.md + acceptance.md)
- **Branch**: `plan/SPEC-GOAL-HTML-WIRING-001` (per repo-local PR-mandatory policy — `enforce_admins: true`)
- **Mode**: Mode 5 (sub-agent sequential) — default for coding-heavy Tier M work per orchestration-mode-selection.md §B
- **Predecessors**: SPEC-GOAL-HTML-FLOW-001 (completed), SPEC-INFINITE-GOAL-001 (completed)

## §B. Working Tree State

- Main checkout, feature branch (opt-in L2 worktree not used for plan-phase per spec-workflow.md Step 1)
- Unmodified at plan-phase start: `.moai/specs/SPEC-GOAL-HTML-FLOW-001/**`, `internal/report/planhtml/`, `internal/goal/dashboard.go`, `internal/hook/handoff_inject.go`

## §C. Plan-phase Decisions

- **Verdict persistence shape**: sidecar file `.moai/state/goal/<session-id>.verdict.json` (single writer: stop-goal evaluator; single reader: `runGoalRender`). NOT extending `<session>.json`; NOT recomputing at render time.
- **CLI namespace**: NEW `moai plan` Cobra parent command (currently does not exist as a CLI verb — `/moai plan` is a slash-command skill routing through the skill system, not the binary). `render-html` is the first subcommand.
- **Surface 1 c2 deferral**: per user-confirmed scope decision 1, the per-turn Stop-hook `.html` auto-refresh is out of scope — separate follow-up SPEC.

## §D. Open Questions

None — intent 100% drained upstream (3 user-confirmed scope decisions: c1-only, new SPEC, CLI verb + closeout rewrite). No `[NEEDS CLARIFICATION]` markers.

## §E.1 Plan-phase Audit-Ready Signal

_<pending plan-auditor — emit verdict + score here after audit>_

## §E.2 Run-phase Evidence

### AC binary matrix (E1)

All 13 ACs PASS (MUST). Verbatim commands + observed output:

| AC | Status | Command | Observed |
|----|--------|---------|----------|
| AC-WIRE-001 (Surface 1 e2e) | PASS | `go test -run TestRunGoalRender_LoadsVerdictAndDOMShowsSections ./internal/cli/` | `--- PASS: TestRunGoalRender_LoadsVerdictAndDOMShowsSections (0.00s)` — DOM shows 5 CeilingVerdict section headings; placeholder absent |
| AC-WIRE-002 (placeholder regression) | PASS | `go test -run TestRunGoalRender_NoVerdictShowsPlaceholder ./internal/cli/` | `--- PASS` — "no verdict yet" present, 5 sections absent |
| AC-WIRE-003 (Cobra registration) | PASS | `go test -run 'TestPlanCmd_RegisteredOnRoot\|TestPlanRenderHTML_HelpExitsZero' ./internal/cli/` | `--- PASS` x2 — `plan` + `plan render-html` resolve on rootCmd; help names SPEC-ID |
| AC-WIRE-004 (Surface 2 e2e) | PASS | `go test -run TestPlanRenderHTML_WritesFileAndDOM ./internal/cli/` | `--- PASS` — `.moai/reports/plan-html/SPEC-FIXTURE-001-plan.html` written; DOM shows goal + 8-field contract + score 0.85 + must-pass + M1/M2 |
| AC-WIRE-005 (Surface 2 fail-open + missing-SPEC) | PASS | `go test -run 'TestPlanRenderHTML_FailOpenOnMissingReview\|TestPlanRenderHTML_MissingSpecDirExitsNonZero' ./internal/cli/` | `--- PASS` x2 — (a) fail-open placeholder + exit 0; (b) non-zero + stderr names SPEC-ID + no html |
| AC-WIRE-006 (template rewrite) | PASS | `go test -run TestSpecAssembly_RewrittenToCLIPath ./internal/cli/` | `--- PASS` — `RenderPlanHTML(specDir=` absent, `moai plan render-html` present, [HARD] + Fail-open preserved |
| AC-WIRE-007 (Surface 3 e2e) | PASS | `go test -run TestRunGoalRender_ReArmIndicatorFromEmbeddedGoal ./internal/cli/` | `--- PASS` — "re-arm on /clear" indicator present, embedded condition text present; source-level: `buildReArmContext` constructs `*goal.ReArmContext` in `internal/cli/goal.go` (non-test) |
| AC-WIRE-008 (Surface 3 DOM, 3 states) | PASS | `go test -run 'TestRunGoalRender_ReArmIndicator\|TestRunGoalRender_ReArmed\|TestRunGoalRender_D8' ./internal/cli/` | `--- PASS` x3 — (a) indicator only; (b) "Re-armed under <id>" mentions new session id; (c) D8 banner present + indicator absent (precedence) |
| AC-WIRE-009 (no-re-arm byte-identity) | PASS | `go test -run TestRunGoalRender_NoReArmStateIsByteIdentical ./internal/cli/` | `--- PASS` — `runGoalRender` output == `RenderDashboard(g, nil)` byte-for-byte |
| AC-WIRE-010 (C-HRA-008 boundary) | PASS | `go test -run 'TestPlan_NoAskUserQuestion\|TestGoal_NoAskUserQuestion' ./internal/cli/` | `--- PASS` x2 — 0 AskUserQuestion/mcp__askuser in non-comment source of plan.go + goal.go |
| AC-WIRE-011 (PRESERVE + Template-First + §25) | PASS | `go test -run TestSpecAssembly_NoNewInternalTokens ./internal/cli/` + `strings bin/moai \| grep -c 'moai plan render-html'` | `--- PASS`; embedded count = 1 (rewrite carried into binary via `make build`); neutrality: 0 SPEC/REQ/AC tokens; renderer signatures byte-identical (no diff to dashboard.go / renderer.go) |
| AC-WIRE-012 (predecessor regression) | PASS | `go test -run 'TestRenderDashboard\|TestRenderDashboardReArm\|TestHTMLPath\|TestClearGoal\|TestOrphanPrune\|TestRenderPlanHTML' ./internal/goal/ ./internal/report/planhtml/` + `go test -run TestAC008 ./internal/hook/` | all `ok` — 11/11 GOAL-HTML-FLOW-001 green; INFINITE-GOAL re-arm tests green |
| AC-WIRE-013 (write-frequency) | PASS | `go test -run TestSaveVerdictWriteFrequency_AtCeilingOnly ./internal/cli/` | `--- PASS` — `saveVerdictFn` calls == 0 across 3 non-exiting turns, == 1 on the ceiling-exit turn |

### Cross-platform build (E2)

```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

### Coverage (E3) — touched packages, whole-package baseline

```
$ go test -cover ./internal/goal/ ./internal/cli/
ok  github.com/modu-ai/moai-adk/internal/goal  0.434s  coverage: 78.1% of statements
ok  github.com/modu-ai/moai-adk/internal/cli   172.220s coverage: 76.3% of statements
```

Whole-package coverage includes pre-existing surface; the SPEC-touched files
(`internal/goal/verdict.go`, `internal/goal/prune.go`, `internal/cli/goal.go`
runGoalRender path, `internal/cli/plan.go`, `internal/cli/hook_stop_goal.go`)
are exercised by the new wiring tests with at-ceiling-only / fail-open /
DOM-parse / Cobra-registration / byte-identity coverage.

### Subagent-boundary grep (E4)

```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/goal.go internal/cli/plan.go internal/cli/hook_stop_goal.go | grep -v _test | grep -v '// '
(0 matches — exit 1)
```

### Lint (E5)

```
$ golangci-lint run --timeout=2m ./internal/goal/ ./internal/cli/ ./internal/report/planhtml/ ./internal/hook/
0 issues.   (matches pre-flight baseline)
```

### Commits + push (E6)

Run-phase commits stacked on `feat/SPEC-GOAL-HTML-WIRING-001` (origin/main..HEAD = `0 6`, all pushed):

- `5b21a362b` — feat: M1 verdict sidecar + Surface 1 wiring
- `094170e6b` — feat: M2 re-arm UI construction (Surface 3)
- `fd8630e80` — feat: M3 moai plan render-html CLI verb (Surface 2)
- `d4bbbd958` — feat: M4 spec-assembly.md rewrite + boundary tests
- `3ed9d364c` — chore: M5 lint+gofmt cleanup

`git push origin feat/SPEC-GOAL-HTML-WIRING-001` → each commit pushed (98cac1da3..3ed9d364c).

### Blocker report (E7)

None. All 13 ACs PASS with no user-decision dependency. The status transition
`draft → in-progress` is the only transition this agent owns (on spec.md
frontmatter; manager-docs owns the `in-progress → implemented → completed`
close via the single sync commit).

### RED failing-test output (E8) — TDD falsifiability evidence

The first RED test of each milestone was confirmed failing BEFORE its GREEN
implementation was written:

- **M1 RED** (`internal/goal/verdict_test.go` before `verdict.go` existed):
  ```
  internal/goal/verdict_test.go:12:9: undefined: VerdictPath
  internal/goal/verdict_test.go:47:12: undefined: SaveVerdict
  internal/goal/verdict_test.go:56:14: undefined: LoadVerdict
  FAIL  github.com/modu-ai/moai-adk/internal/goal [build failed]
  ```
- **M1 CLI RED** (`goal_wiring_test.go` before `saveVerdictFn` + `LoadVerdict` in runGoalRender):
  ```
  internal/cli/goal_wiring_test.go:202:12: undefined: saveVerdictFn
  internal/cli/goal_wiring_test.go:209:21: undefined: saveVerdictFn
  FAIL  github.com/modu-ai/moai-adk/internal/cli [build failed]
  ```
- **M2 RED** (`goal_rearm_test.go` before `buildReArmContext` in runGoalRender): tests ran
  successfully (compile-clean — `RenderDashboardReArm` already existed) but the DOM
  assertions FAILED — the rendered body lacked "re-arm on /clear" / "Re-armed under
  session" / "D8 rejection" because `runGoalRender` still called `RenderDashboard(g, v)`
  with no reArm. `FAIL  github.com/modu-ai/moai-adk/internal/cli  1.001s`.
- **M3 RED** (`plan_test.go` before `plan.go` existed):
  ```
  internal/cli/plan_test.go:125:9: undefined: newPlanCmd
  FAIL  github.com/modu-ai/moai-adk/internal/cli [build failed]
  ```
- **M4 boundary RED** (`plan_subagent_boundary_test.go` before comment-filter):
  ```
  --- FAIL: TestPlan_NoAskUserQuestion (0.00s)
      plan_subagent_boundary_test.go:19: internal/cli/plan.go must NOT reference AskUserQuestion
  --- FAIL: TestGoal_NoAskUserQuestion (0.00s)
  ```
  (Fixed by mirroring internal/goal/subagent_boundary_test.go's comment filter
  per AC-WIRE-010's "comments out of scope" note — the naive `strings.Contains`
  was flagging docstring mentions of the boundary itself.)

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-07
run_commit_sha: 3ed9d364c   # M5 final commit on feat/SPEC-GOAL-HTML-WIRING-001
run_status: complete
ac_pass_count: 13
ac_fail_count: 0
preserve_list_post_run_count: 3   # RenderDashboard/RenderDashboardReArm/RenderPlanHTML signatures + GOAL-HTML-FLOW-001 artifacts + INFINITE-GOAL re-arm mechanism
l44_pre_commit_fetch: clean (0 6 — local ahead by 6, pushed)
l44_post_push_fetch: clean (origin tracked each commit)
new_warnings_or_lints_introduced: 0
cross_platform_build.linux: exit 0
cross_platform_build.windows: exit 0
total_run_phase_files: 12   # verdict.go + verdict_test.go + state.go + prune.go (modified) + hook_stop_goal.go + goal.go + goal_wiring_test.go + goal_rearm_test.go (new) + plan.go + plan_test.go + plan_subagent_boundary_test.go (new) + spec-assembly.md (template + local mirror) + root.go
m1_to_mN_commit_strategy: per-milestone Conventional Commit (5 commits + plan-phase base)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-07
sync_commit_sha: 7901fd704   # main merge SHA of #1388 (squash); backfilled from 6622de0f6 (in-branch sync commit)
sync_status: audit-ready
ac_pass_count: 13
ac_fail_count: 0
changelog_entry_position: 1   # first entry under [Unreleased] ### Added
frontmatter_status_transitions:
  spec_md: "in-progress -> completed"   # single sync-commit terminal transition (3-phase close)
  plan_md: n/a                          # no frontmatter on plan.md
  acceptance_md: n/a                    # no frontmatter on acceptance.md
  progress_md: n/a                      # progress.md carries this signal, not a status field
canary_compliance_check:
  grep_changelog_count: 1               # exactly one CHANGELOG entry for this SPEC-ID
  grep_acceptance_ac_count: 13          # AC-WIRE-001..013 distinct identifiers in acceptance.md
  readme_edit_required: false           # README (en/ko/ja/zh) has no pre-existing mentions of these commands; out of sync-phase scope
```

## §F. Phase 4 Mode Selection

_<pending run-phase — populated by orchestrator before first Agent() spawn>_
