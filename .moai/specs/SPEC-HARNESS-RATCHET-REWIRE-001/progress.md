# SPEC-HARNESS-RATCHET-REWIRE-001 — Progress

> Plan-phase artifact. Run-phase evidence (§E.2, §E.3) and sync-phase evidence (§E.4) are populated by manager-develop and manager-docs respectively. This file is parser-load-bearing: `internal/spec/era.go` `hasAnyProgressMarker` greps for the literal `§E.2` / `§E.3` / `§E.4` substrings — the headings below MUST be preserved verbatim.

---

## §E.1 Plan-phase Audit-Ready Signal

```
plan_status: audit-ready
plan_complete_at: 2026-07-09
tier: M
epic: Workflow-Reflex (1 of 6)
artifacts: 4 (spec.md + plan.md + acceptance.md + progress.md)
gears_requirements: 9
acceptance_criteria: 13
out_of_scope_topics: 6
audit_findings_traced: H1, H2, L2
spec_id_self_check: PASS (SPEC-HARNESS-RATCHET-REWIRE-001 → ^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$)
```

**Plan-phase self-verification**:
- [x] All GEARS requirements use valid patterns (Ubiquitous / When / Where / shall-not; no IF/THEN).
- [x] All gap claims cite measured source + observed pattern (vci §2 attribution; measured 2026-07-09).
- [x] §Out of Scope has ≥1 `### Out of Scope — <topic>` H3 sub-heading with `-` bullets (6 topics).
- [x] 12 canonical frontmatter fields + era + tier; no snake_case aliases.
- [x] Human apply gate (`auto_apply: false`) preservation stated as HARD constraint.
- [x] Fail-open + 5s hook-budget constraint stated for the Stop-path propose chain.

---

## §E.2 Run-phase Evidence

### Pre-flight baseline (measured 2026-07-09 by manager-develop)

```
git branch --show-current → worktree-agent-a949030202916a3bf (fast-forwarded from main)
git rev-parse HEAD        → ab4f38def (== origin/main, divergence 0 0 after ff)
go build ./...            → exit 0
GOOS=windows GOARCH=amd64 go build ./... → exit 0
golangci-lint run         → 0 issues (clean baseline)
go test ./internal/harness/... ./internal/hook/... ./internal/cli/harness/... → all ok (exit 0)
grep "classifyHarnessPatterns" internal/cli/hook.go → line 772 (Stop chain anchor)
grep "Retired\|superseded" internal/harness internal/cli/harness → no output (B2 clean)
```

### Open decision resolutions (plan.md §D)

- **D1 (failure signature)**: RESOLVED → low-cardinality error-class token. Reuse the existing `ErrorCategory` (TimeoutError/ExitError/etc. ~7 tokens) as the `<signature>` ContextHash. Avoids cardinality explosion.
- **D2 (propose invocation seam)**: RESOLVED → extract a callable `generateProposals(root)` in cli/hook.go reusing `proposalgen.ReadPromotions`/`MapPromotions`/`WriteProposals`. No subprocess (within 5s budget).
- **D3 (inbox stub schema)**: RESOLVED → append-only JSONL (ts / event_key / summary / source). Drain marking deferred to the orchestrator's Lessons Protocol (named by M3 doctrine cross-ref).
- **D4 (eligibility predicate placement)**: RESOLVED → filter at classification time (in `classifyHarnessPatterns` loop, before `WritePromotion`).

### D2-audit note (AC-HRR-012 grep refinement)

acceptance.md §D.4 AC-HRR-012 grep matches 6 string-literal lines in the clean baseline (propose.go:48/50/59, execute.go:320, install.go:117, scaffolder.go:111). The E4 self-verification anchors on the call pattern `AskUserQuestion(` to avoid the false-positive FAIL.

_<M1/M2/M3 evidence appended below as milestones complete>_

### M1 — failure-event recording (PASS)

- AC-HRR-001: `tool_failure:Bash:ExitError` event recorded via harness.Observer (signature = ErrorCategory token, D1).
- AC-HRR-002: `test_fail:internal/hook:` event recorded on test-fail detection (package from extractTestPackage).
- EC-1: empty tool name → `unknown-tool` placeholder (non-degenerate key).
- Tests: `go test -run "TestPostToolFailure_Records|TestPostToolFailure_EmptyToolName|TestEvidenceWriter_RecordsTestFail|TestExtractTestPackage" ./internal/hook/` → ok.
- Files: types.go (EventTypeToolFailure/EventTypeTestFail + PatternBearing), post_tool_failure.go, evidence_writer.go, failure_observer.go (new), failure_event_test.go (new).
- Commit: 8d49e286d. Frontmatter draft → in-progress (manager-develop M1 ownership).

### M2 — classifier eligibility + propose auto-run (PASS)

- AC-HRR-003: regression test proves `subagent_stop:unknown:` (41 obs) NO LONGER promoted; `session_stop::`, `user_prompt::` excluded.
- AC-HRR-004: `IsEligibleForPromotion(p)` predicate — ineligible when empty context hash AND empty/unknown subject. Filter at classification time (D4).
- AC-HRR-005: Stop path chains `generateProposals(root)` after classify when promoCount > 0 → proposal dirs land in `.moai/harness/proposals/`.
- AC-HRR-006: propose chain fail-open — faulted path → stderr log, exit 0 (never blocks session end).
- AC-HRR-007: no apply invocation on Stop chain (`grep "harness apply\|ApplyProposal\|\.Apply(" internal/cli/hook.go` → 0 matches); `auto_apply: false` unchanged (harness.yaml:116).
- Decision recorded: propose-chain output dir = `.moai/harness/proposals/` (learning-loop home, per SPEC REQ-HRR-004 / AC-HRR-005 / AC-HRR-010; distinct from legacy manual-propose `.moai/proposals` default). Generation core (ReadPromotions/MapPromotions/WriteProposals) reused; only output-dir param differs.
- Cascade fix: `seedUsageLog` helper updated to seed eligible events (non-empty context_hash) — the prior degenerate `user_prompt::` seed is now correctly excluded by the eligibility filter.
- Tests: `go test -run "TestIsEligibleForPromotion|TestClassifyHarnessPatterns_ExcludesDegenerate|TestRunHarnessObserveStop_ProposeChain" ./internal/harness/ ./internal/cli/` → ok.
- Pre-existing environmental note: 6 status/doctor golden tests (`TestStatus_*`, `TestDoctor_*`) fail in the worktree due to golden-snapshot/env mismatch — confirmed pre-existing on clean M1 state (git stash verified), NOT caused by SPEC changes. They test `moai status`/`moai doctor` rendering (no overlap with classify/propose/eligibility code).

### M3 — lessons inbox + doctor + doctrine (PASS)

- AC-HRR-008: failure handlers append structured stubs (ts/event_key/summary/source) to `.moai/lessons-inbox.jsonl` (D3: append-only JSONL, 0o600 perms). Both tool_failure and test_fail paths append; append-only semantics verified (3 stubs from 3 failures).
- AC-HRR-009: moai-constitution.md § Lessons Protocol cross-references the inbox drain (live + template mirror, template-first per REQ-HRR-009/011). `grep -c lessons-inbox` → 1 in each.
- AC-HRR-010: `moai harness doctor` emits dormancy warning when promotions ≥1 AND proposals dir absent; NOT emitted when proposals present or promotions absent.
- Files: failure_observer.go (lessonsInboxStub + appendLessonsInboxStub + truncateSummary, wired into both recorders), doctor.go (checkPipelineDormancy + countPromotionLines), lessons_inbox_test.go (new), doctor_dormancy_test.go (new), moai-constitution.md (live + template mirror).
- Tests: `go test -run "TestRecordToolFailureEvent_AppendsLessonsInbox|TestRecordTestFailEvent_AppendsLessonsInbox|TestLessonsInbox_AppendOnly|TestDoctor_Dormancy|TestDoctor_NoDormancy" ./internal/hook/ ./internal/cli/harness/` → ok.

## §E.3 Run-phase Audit-Ready Signal

```
run_complete_at: 2026-07-09
run_commit_sha: 87471fb18 (coverage hardening, final run-phase commit)
run_status: audit-ready
ac_pass_count: 13
ac_fail_count: 0
preserve_list_post_run_count: 3 (applier.go, rubric.go, regression_gate.go — semantics untouched)
l44_pre_commit_fetch: origin/main fetched pre-push; divergence 0 0 after push
l44_post_push_fetch: origin/main == HEAD (0 0); 4 commits on origin/main
new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues vs clean baseline)
cross_platform_build:
  darwin_amd64: exit 0
  windows_amd64: exit 0
total_run_phase_files: 16 (6 Go source edited/new, 6 test files new, 2 doctrine files live+mirror, 4 SPEC frontmatter/evidence)
m1_to_mN_commit_strategy: one feat commit per milestone (M1/M2/M3) + one test commit (coverage hardening); rebased onto a parallel-session divergence (origin gained 4 complementary commits mid-run) — rebase clean, no conflicts; pushed HEAD:main (Route A Hybrid Trunk)
```

Run-phase status: `in-progress` (draft→in-progress set on M1 commit by manager-develop). The `in-progress → implemented → completed` close is owned by manager-docs (sync-phase).

## §E.4 Sync-phase Audit-Ready Signal

```
sync_complete_at: 2026-07-09
sync_commit_sha: <pending — backfill via follow-up chore commit per established convention>
sync_status: audit-ready (3-phase close on single sync commit; completed transition rides this commit)
changelog_entry_position: [Unreleased] > Added (SPEC-HARNESS-RATCHET-REWIRE-001)
frontmatter_status_transitions:
  spec.md: in-progress -> completed (single sync commit, 3-phase close)
  plan.md: in-progress -> completed (same sync commit, atomic)
  acceptance.md: in-progress -> completed (same sync commit, atomic)
  progress.md: n/a (no YAML frontmatter — "# " header start; era.go parses §E.* headings + sync_commit_sha field, not frontmatter)
b12_self_test_a_pre_emission_grep: PASS (grep -c 'SPEC-HARNESS-RATCHET-REWIRE-001' CHANGELOG.md -> 0 pre-emit; no duplicate)
b12_self_test_b_ac_count_match: PASS (acceptance.md SSOT = 13 AC rows via '^\| AC-HRR-[0-9]+ \|'; CHANGELOG entry references 13/13)
b12_self_test_c_file_path_verify: PASS (internal/harness, internal/hook, internal/cli/harness, .claude/rules/moai/core/moai-constitution.md + template mirror all exist via ls)
```

### Sync-phase verification (measured 2026-07-09 by manager-docs)

**SCOPING NOTE**: working tree is dirty with parallel-session modified files (`internal/statusline/{renderer.go,cache_hit_test.go}`, `.claude/agents/moai/{manager-docs,manager-git}.md`, `.moai/config/sections/{llm,system,workflow}.yaml`, template mirror copies, `pkg/version/version.go`) + untracked (`.claude/output-styles/moai/moai-easy.md`, `.moai/reports/*.html`, `.moai/specs/SPEC-SUBCOMMAND-RETIRE-001/`, `.moai/config/astgrep-rules/security/credentials.yml`, `.moai/README.md`, `.claude/rules/.moai/`). A full `go test ./...` would conflate RATCHET scope with parallel-session changes in unrelated packages. Verification is therefore scoped to RATCHET-touched packages only: `./internal/harness/... ./internal/hook/... ./internal/cli/harness/...` (all 3 paths verified to exist via `ls -d`).

```
$ go test ./internal/harness/... ./internal/hook/... ./internal/cli/harness/...
exit=0
ok  	github.com/modu-ai/moai-adk/internal/harness	0.897s
ok  	github.com/modu-ai/moai-adk/internal/harness/capture	(cached)
ok  	github.com/modu-ai/moai-adk/internal/harness/cluster	0.376s
ok  	github.com/modu-ai/moai-adk/internal/harness/proposalgen	1.094s
ok  	github.com/modu-ai/moai-adk/internal/harness/router	(cached)
ok  	github.com/modu-ai/moai-adk/internal/harness/safety	1.449s
ok  	github.com/modu-ai/moai-adk/internal/harness/seeds	(cached)
ok  	github.com/modu-ai/moai-adk/internal/harness/throttle	(cached)
ok  	github.com/modu-ai/moai-adk/internal/harness/tier	(cached)
ok  	github.com/modu-ai/moai-adk/internal/harness/v4manifest	(cached)
ok  	github.com/modu-ai/moai-adk/internal/hook	4.379s
ok  	github.com/modu-ai/moai-adk/internal/hook/dbsync	(cached)
ok  	github.com/modu-ai/moai-adk/internal/hook/handoff	(cached)
ok  	github.com/modu-ai/moai-adk/internal/hook/memo	(cached)
ok  	github.com/modu-ai/moai-adk/internal/hook/memo/taxonomy	(cached)
ok  	github.com/modu-ai/moai-adk/internal/hook/mx	(cached)
ok  	github.com/modu-ai/moai-adk/internal/hook/mx/complexity	(cached)
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	(cached)
ok  	github.com/modu-ai/moai-adk/internal/hook/security	(cached)
ok  	github.com/modu-ai/moai-adk/internal/hook/testutil	(cached)
ok  	github.com/modu-ai/moai-adk/internal/hook/trace	(cached)
ok  	github.com/modu-ai/moai-adk/internal/cli/harness	3.157s
(22 packages, all ok, 0 FAIL)

$ golangci-lint run ./internal/harness/... ./internal/hook/... ./internal/cli/harness/...
exit=0
0 issues. (NEW vs baseline = 0; RATCHET scope clean)
```

### Constraint re-verification (sync-phase, observed this session)

```
$ grep -rn 'AskUserQuestion(' internal/harness/ internal/hook/ internal/cli/harness/ | grep -v "_test.go"
internal/harness/safety/oversight.go:22:  // The orchestrator receives the return value and displays it to users via AskUserQuestion(options=payload.Options).
(1 match — comment-line docstring only, NOT a call site; AC-HRR-012 PASS, call-site pattern = 0)

$ grep -n "auto_apply" .moai/config/sections/harness.yaml
116:    auto_apply: false
(REQ-HRR-005 human apply gate preserved, unchanged)

$ grep -c "lessons-inbox" .claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md
1  (live)
1  (template mirror)
(AC-HRR-009 doctrine cross-ref present in both trees)
```

### Run-phase evidence summary (from §E.2/§E.3, populated by manager-develop)

- **run_commit_sha**: 87471fb18 (coverage hardening, final run-phase commit; §E.3 records this as the final run-phase commit, with 3d81db272 as the §E.3 audit-ready signal commit carrying the run-phase evidence block above).
- **Run-phase commits on origin/main**: e70c77576 (M1 failure-event recording) → 838e38102 (M2 classifier eligibility + propose auto-run) → 69041f969 (M3 lessons inbox + doctor + doctrine) → 87471fb18 (coverage hardening) → 3d81db272 (§E.3 audit-ready signal).
- **AC pass/fail**: 13/13 PASS, 0 FAIL (acceptance.md §D SSOT; all MUST-PASS).
- **Cross-platform build** (run-phase measured): darwin_amd64 exit 0, windows_amd64 exit 0.
- **Coverage hardening** (M4): internal/hook + internal/cli/harness touched files brought to ≥85% per AC-HRR-013.
- **Pre-existing environmental note**: 6 status/doctor golden tests (`TestStatus_*`, `TestDoctor_*`) were observed failing in the run-phase worktree due to golden-snapshot/env mismatch — confirmed pre-existing on clean M1 state, NOT caused by SPEC changes, and OUT of RATCHET scope (they test `moai status`/`moai doctor` rendering, no overlap with classify/propose/eligibility code). These are owned by SPEC-INTERNAL-TEST-002 (DEBT-TEST-001 stale-golden, now resolved at HEAD per TEST-002 sync).

### Sync-phase close

3-phase close per SPEC-V3R6-LIFECYCLE-REDESIGN-001 / Status Transition Ownership Matrix: the single sync commit carries the `in-progress → completed` terminal transition (merged 3-phase close — NO separate Mx chore commit; MX Tag validation is a sync sub-step, not a fourth phase). `sync_commit_sha` backfilled via follow-up chore commit per established convention (SPEC-BRAND-DIR-REMOVE-001 `2c637ece2` / SPEC-TOKEN-BUDGET-STOP-001 `ab4f38def` / SPEC-INTERNAL-TEST-002 `6116a0972`). README unchanged — RATCHET touches internal learning-loop plumbing (failure-event recording, classifier eligibility, propose auto-run seam, lessons inbox, doctor dormancy additive check); README's Harness section documents the v4 Builder + lifecycle, not learning-loop internals or doctor sub-checks, so no user-facing README surface change.

---

## §F Phase 0.95 Mode Selection

### Input parameters
- tier: M
- scope: ~8-12 files (hook.go, post_tool_failure.go, evidence_writer.go, classifier in internal/harness, propose.go reuse seam, doctor.go, new inbox writer, constitution rule + template mirror, tests)
- domain count: 3 (Go source internal/hook + internal/harness + internal/cli; rule doc + template mirror)
- file language mix: Go source + markdown doctrine
- concurrency benefit: LOW (coding-heavy, sequential milestone dependency: M1 → M2 → M3)
- Agent Teams prereqs: NOT met (workflow.team.enabled defaults false; env var unset)

### Mode evaluation
- Mode 1 trivial: not selected (Tier M, 3 milestones, 400-800 LOC)
- Mode 2 background: not selected (write work — implementation modifies files)
- Mode 3 agent-team: not selected (prereqs not met; coding-heavy)
- Mode 4 parallel: not selected (coding-heavy per Anthropic caveat; M2 depends on M1)
- Mode 5 sub-agent: **selected** (sequential milestone delegation; coding-heavy default)
- Mode 6 workflow: not selected (<30 files, not mechanical-uniform transform)

### Decision
sub-agent — sequential manager-develop delegation per milestone (M1 → M2 → M3).

### Justification
Tier M coding-heavy work with sequential milestone dependency (M1 failure-events → M2 classifier consumes them → M3 additive). Anthropic's coding-task parallelism caveat favors sequential sub-agent over parallel/team fan-out. Three small Go features + tests + one doc edit fit a single Section A-E delegation with milestone-ordered execution. plan-auditor iter-1 verdict PASS 0.90 (skip-eligible boundary) recorded 2026-07-09; Implementation Kickoff Approval obtained same session.
