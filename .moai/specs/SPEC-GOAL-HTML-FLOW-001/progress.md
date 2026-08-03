# progress.md — SPEC-GOAL-HTML-FLOW-001

> Lifecycle skeleton emitted at plan-phase. §E.1 (plan-phase audit-ready signal) is populated by manager-spec. §E.2 / §E.3 (run-phase evidence + audit-ready) are populated by manager-develop. §E.4 (sync-phase audit-ready) is populated by manager-docs. The literal `§E.2` / `§E.3` / `§E.4` heading tokens are parsed by `internal/spec/era.go` for V3R6 era classification — preserve them verbatim.

## §A. Plan-phase Artifacts

- spec.md: `.moai/specs/SPEC-GOAL-HTML-FLOW-001/spec.md` (12-field frontmatter, GEARS REQ-GHF-001..010, §B Out of Scope with `### Out of Scope — <topic>` H3 sub-headings).
- plan.md: `.moai/specs/SPEC-GOAL-HTML-FLOW-001/plan.md` (§A-§H, 6 milestones decision-reversibility-first).
- acceptance.md: `.moai/specs/SPEC-GOAL-HTML-FLOW-001/acceptance.md` (§D matrix, 11 ACs GWT, §D.1 severity, §D.3 indirect, §D.4 closure, §D.5 forward-looking).
- progress.md: this file.

## §B. Plan-phase Tier

Tier M (4 artifacts: spec.md + plan.md + acceptance.md + this progress.md). plan-auditor PASS threshold: `0.80`.

## §C. Pre-plan Audit-ready Checklist (manager-spec self-verification)

- [x] SPEC ID regex `PASS` (verbatim): `ID="SPEC-GOAL-HTML-FLOW-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS` → `PASS`.
- [x] Frontmatter 12 canonical fields present (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags). No snake_case aliases.
- [x] `phase: "v3.1 target"` — release target label (NOT a lifecycle stage token; the prohibition on `plan`/`run`/`sync`/`mx` is respected).
- [x] `status: draft` — initial plan-phase emission.
- [x] Requirements in GEARS notation (Where / While / When / The <subject> shall).
- [x] Out of Scope section carries `### Out of Scope — <topic>` H3 sub-headings with `-` bullets (5 sub-headings: Dashboard per-turn auto-refresh / Orchestrator AskUserQuestion simplification / Plan-auditor JSON sidecar / Mechanical re-arm pipeline / Web live dashboard v3.1 target).
- [x] Artifact set matches Tier M (spec.md + plan.md + acceptance.md + progress.md).
- [x] spec.md carries no implementation detail (functions named as consumer-facing API contracts only; no code).

## §D. ID Uniqueness

`SPEC-GOAL-HTML-FLOW-001` verified absent from `.moai/specs/` prior to artifact creation (no duplicates; no sibling SPEC claims the GOAL-HTML-FLOW domain).

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-08-04
plan_tier: M
plan_artifact_count: 4
plan_artifacts:
  - spec.md
  - plan.md
  - acceptance.md
  - progress.md
plan_auditor_verdict: pending
```

## §E.2 Run-phase Evidence

### AC matrix (11/11 PASS, attributed evidence)

| AC | Status | Command | Observed output |
|----|--------|---------|-----------------|
| AC-GHF-001 | PASS | `go test -run TestRenderDashboard_DOMParseShowsGoalFailedCondAndSections ./internal/goal/` | `ok github.com/modu-ai/moai-adk/internal/goal` — DOM parse via `golang.org/x/net/html` asserts goal text + failed-cond Cmd/Exit/Tail + Turn/Ceiling + 5 verdict section headings verbatim |
| AC-GHF-002 | PASS | `go test -run TestRenderDashboard_XSSPayloadInert ./internal/goal/` | DOM parse of `<script>` payload in every untrusted field → `len(scriptNodes) == 0` |
| AC-GHF-003 | PASS | `go test -run TestClearGoalRemovesBothJSONAndHTML ./internal/goal/` | ClearGoal removes both `<session>.json` AND `<session>.html`; idempotent second call nil |
| AC-GHF-004 | PASS | `go test -run TestOrphanPruneMovesHTMLSibling ./internal/goal/` | PruneOrphans moves `.html` alongside `.json` to consumed/; best-effort variant (json-only) green |
| AC-GHF-005 | PASS | `go test -run TestRenderPlanHTML_DOMShowsAllFields ./internal/report/planhtml/` | DOM parse shows goal + all 8 contract field labels + verdict score 0.85 + milestones M1/M2/M3 |
| AC-GHF-006 | PASS | skill-layer inspection of spec-assembly.md Step 2.3.3a | Emission step lands AFTER plan-auditor PASS, BEFORE Phase 12; `[HARD]` clause: gate stays MANDATORY + score-independent; report is additive prose context, NOT a gate substitution (AP-4) |
| AC-GHF-007 | PASS | `go test -run TestRenderDashboardReArm ./internal/goal/` | Three fixture states: embedded_goal indicator + post-/clear re-armed view + D8 unbounded-rejection banner all render via DOM parse |
| AC-GHF-008 | PASS | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/goal.go \| grep -v _test.go \| grep -v '^[^:]*:[0-9]*:[[:space:]]*//'` | exit 1 (zero non-comment matches); Go guard test `TestGoalRenderCmd_NoAskUserQuestion` enforces comment-line exclusion |
| AC-GHF-009 | PASS | `go test -run TestRenderPlanHTML_Determinism ./internal/report/planhtml/` | Two `RenderPlanHTML` runs over the same fixture → byte-identical output |
| AC-GHF-010 | PASS | `go test -run TestRunGoalRender_NoGoalExitsNonZero ./internal/cli/` | No armed goal → non-zero exit + stderr names session id + NO `.html` written |
| AC-GHF-011 | PASS | `go test -run TestRenderDashboard_NilVerdictNoPanic ./internal/goal/` | `RenderDashboard(g, nil)` no panic; goal metadata + "no verdict yet" placeholder; 5 section names absent |

### Cross-platform build (AC-E2)

- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (no syscall use; stdlib html/template + CLI only)

### Lint (AC-E5)

- `golangci-lint run --timeout=2m ./internal/goal/... ./internal/report/... ./internal/cli/...` → 0 issues

### Coverage (AC-E3)

- `internal/goal/` → 78.1% (dashboard.go + new lifecycle code well-covered; package baseline includes large pre-existing evaluate.go)
- `internal/report/planhtml/` → 87.2% (exceeds 85% target)

### Race (AC-E3 sub)

- `go test -race ./internal/goal/... ./internal/report/...` → ok (no goroutines/channels in new code)

### Milestone commits (M1–M6)

- M1 `feat(SPEC-GOAL-HTML-FLOW-001): M1 dashboard renderer substrate + escaping` (transitions `status: draft → in-progress`)
- M2 `feat(SPEC-GOAL-HTML-FLOW-001): M2 goal render verb + .html sibling lifecycle`
- M3 `feat(SPEC-GOAL-HTML-FLOW-001): M3 plan-HTML renderer + 8-field derivation`
- M4 `feat(SPEC-GOAL-HTML-FLOW-001): M4 plan-HTML emission step (Kickoff gate enriched, not replaced)`
- M5 `feat(SPEC-GOAL-HTML-FLOW-001): M5 re-arm UI (render-only, consumes landed state)`
- M6 this commit (verification + §E.2/§E.3 population)

### M3 design decision (plan.md constraint 6a)

- **plan-HTML renderer location**: `internal/report/planhtml/renderer.go` (NEW package) — chosen over extending `internal/goal/dashboard.go` because the plan-HTML renderer parses plan-auditor markdown + derives an 8-field contract from SPEC artifacts, which is a DISTINCT concern from the goal `Verdict` renderer. Cohesion favors the separate package.
- **design-report citation** (plan.md §H recommendation): NOT committed in this run (the report at `.moai/reports/moai-autonomy-workflow-redesign-20260803.html` is gitignored; the §3.2 contract is already inlined verbatim into spec.md §A, so traceability holds without committing the binary). Judgment call deferred — no functional impact.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-04
run_commit_sha: "pending-backfill-m6"
run_status: audit-ready
ac_pass_count: 11
ac_fail_count: 0
preserve_list_post_run_count: 6  # the 6 PRESERVE-list files (plan.md §A.4) untouched
l44_pre_commit_fetch: n/a        # Route B branch push; pre-push fetch discipline applied
l44_post_push_fetch: pending     # to be observed after git push
new_warnings_or_lints_introduced: 0
cross_platform_build:
  goos_linux_amd64: pass
  goos_windows_amd64: pass
  goos_darwin_arm64: pass         # dev host
total_run_phase_files: 8          # dashboard.go(+test), html_sibling_test.go, state.go, prune.go, goal.go, goal_render_test.go, renderer.go(+test), plan.html.mustache
m1_to_mN_commit_strategy: per-milestone conventional commits with 🗿 MoAI trailer
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-04
sync_commit_sha: "535a88710"
sync_status: audit-ready
sync_phase_close: "3-phase close (plan→run→sync) — completed transition rides this sync commit"
changelog_entry_added: true   # CHANGELOG.md [Unreleased] → ### Added
readme_4_locale_sync: true     # README.md / README.ko.md / README.ja.md / README.zh.md — `render` verb added to `moai goal` row
spec_md_frontmatter_transition: "in-progress → implemented → completed (single sync commit)"
spec_md_body_touched: false    # frontmatter status + updated only
plan_acceptance_body_touched: false   # forbidden ownership crossing respected
spec_lint_result: "0 errors, 1 warning (StatusGitConsistency — resolved by this sync commit's in-progress→completed transition)"
b12_changelog_self_test:
  pre_emission_grep_count: 0       # grep -c 'SPEC-GOAL-HTML-FLOW-001' CHANGELOG.md == 0 before emission (no duplicate)
  ac_count_acceptance_md: 11       # distinct AC-GHF-* identifiers in acceptance.md
  ac_count_changelog_entry: 11     # CHANGELOG entry references the same 11 ACs (AC-GHF-001..011)
  file_paths_verified: true        # ls confirmed: internal/goal/dashboard.go, internal/cli/goal.go, internal/report/planhtml/renderer.go
frontmatter_status_transitions:
  spec_md: "in-progress → completed (3-phase close chain in-progress→implemented→completed rides this single sync commit)"
  plan_md: "n/a (no YAML frontmatter)"
  acceptance_md: "n/a (no YAML frontmatter)"
canary_compliance_check:
  spec_lint_clean: true             # 0 errors on spec.md
  status_git_consistency: resolved  # the pre-sync warning (in-progress vs git-implied implemented) is resolved by this commit's completed transition
```

## §F Phase 4 Mode Selection

**Input parameters:**
- tier: M (spec.md frontmatter `tier: M`)
- scope (file count): ~8–10 files — NEW `internal/goal/dashboard.go`(+test), `internal/cli/goal.go`, `internal/goal/state.go`, `internal/goal/prune.go`, plan-HTML renderer(~2 files), `plan.html.mustache`, `spec-assembly.md`, `goal.md` workflow — within Tier M (5–15 files)
- domain count: 2 (internal/goal + internal/cli, plus skill/workflow markdown edits) — below Mode 4 ≥3 threshold
- file language mix: Go source (renderer, CLI, lifecycle) + markdown (skill/workflow) + mustache (template slot)
- concurrency benefit: LOW — coding-heavy TDD (sequential RED-GREEN-REFACTOR), per Anthropic's coding-task parallelism caveat

**Mode evaluation:**
- Mode 1 (trivial): not selected — multi-file feature, not a typo.
- Mode 2 (background): not selected — write-capable implementation, not read-only.
- Mode 3 (agent-team): RETIRED — never selected.
- Mode 4 (parallel): not selected — coding-heavy (concurrency benefit LOW); Mode 4 reserved for research-heavy multi-domain.
- Mode 5 (sub-agent): SELECTED — coding-heavy TDD, safe default per Anthropic's coding-task parallelism caveat; sequential per-milestone delegation.
- Mode 6 (workflow): not selected — not a high-volume mechanical transform; new-code work stays Mode 5.

**Decision:** sub-agent

**Justification:** Coding-heavy Go implementation (dashboard renderer, CLI verb, sibling lifecycle, plan-HTML report renderer) driven by TDD. Per Anthropic's finding that most coding tasks involve fewer truly parallelizable tasks than research, the sequential sub-agent path is the correct default. The 6 milestones are sequentially dependent (M5 re-arm UI consumes M1's renderer; M4 emission consumes M3 report) and cannot fan out. Implementation Kickoff Approval obtained in prior session (per resume); Mode 5 is strictly downstream of that gate.

## §H Recursive Self-Diagnosis Log

_<pending run-phase>_

## §I Token Accounting

```yaml
# Sync-close per-SPEC token spend measurement (best-effort, per progress.md §I convention).
# Attribution confidence: LOW — this sync-phase was executed inside a worktree branch off
# the already-merged run-PR squash (d1fcf09b0); the token figures are self-reported by the
# sync agent (manager-docs) and NOT measured by a mechanical token-accounting mechanism.
tokens_spent_estimate: "~25K (sync-phase scope only: CHANGELOG + 4×README + spec.md frontmatter + progress.md §E.4/§I)"
attribution_confidence: "low (self-reported; no mechanical meter — the token-accounting mechanism is documentation-only at this layer)"
phase: sync
spec_id: SPEC-GOAL-HTML-FLOW-001
```
