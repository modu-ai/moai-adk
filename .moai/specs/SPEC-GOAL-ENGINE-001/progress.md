# progress.md — SPEC-GOAL-ENGINE-001

> Canonical §E section skeleton. Plan-phase populates §E.1 only; §E.2/§E.3 are
> owned by manager-develop (run-phase), §E.4 by manager-docs (sync-phase).

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready (v0.3.0 in-place amendment — SUPERSEDES the v0.2.1 signal; plan RE-AUDIT REQUIRED — see "### Amendment 0.3.0 plan-phase re-entry" below) [prior: v0.2.1 — 2 D2 fixes from plan-auditor v0.2.0 audit applied]
- plan_complete_at: 2026-07-12
- tier: L (LEAN: 3 core artifacts + progress.md; design folded into plan.md § Technical Design; research.md shared from SPEC-ANALYZE-FIRST-ROUTING-001)
- artifacts: spec.md, plan.md, acceptance.md, progress.md
- REQ count: 34 (REQ-GLE-001..025 + REQ-GLE-026..029 added v0.2.0 amendment D8 + REQ-GLE-030..034 added v0.3.0 amendment; no new REQs in v0.2.1)
- AC count: 39 (AC-GLE-001..026 + AC-GLE-027..034 added v0.2.0 amendment D8 + AC-GLE-035..039 added v0.3.0 amendment; no new ACs in v0.2.1)
- depends_on: SPEC-ANALYZE-FIRST-ROUTING-001
- v0.2.1 changes (plan-auditor v0.2.0 D2 fixes): D2-1 = enrich §B.5 checkpoint JSON with `failed_conditions: [{cmd, exit, tail}]` + reconcile REQ-GLE-010 ↔ REQ-GLE-028 (failed-condition+tail present in BOTH modes) + amend AC-GLE-029 to assert `failed_conditions`; D2-2 = re-anchor AC-GLE-021(a) from stale `grep -ic "goal evaluator\|goal engine" CLAUDE.md` (baseline 1, non-discriminating) to `awk '/^## 2\./,/^## 3\./' CLAUDE.md | grep -ic "goal evaluator"` (verified baseline 0, discriminating).
- open decisions: 0 remaining — all 4 iteration-2 decisions resolved + 2 D8 amendment decisions resolved (progression-mode axis = kickoff-time choice NOT gate bypass; semi-autonomous confirm via orchestrator-bridge NOT hook prompt). See plan.md Settled Decisions. v0.2.1 pending plan re-audit per the orchestrator.

### Amendment 0.3.0 plan-phase re-entry (2026-07-12) — arm CLI + prune wiring reachability

- plan_status: audit-ready (v0.3.0 in-place amendment; status completed → in-progress; `amendment_of` self-ref) — **plan RE-AUDIT REQUIRED** (the amendment invalidates the cached plan-auditor PASS from v0.2.1; the plan-artifact hash changed because spec.md was modified — see spec-workflow.md § Report Persistence "Amendment as cache-invalidating event").
- amendment_re_entry_at: 2026-07-12
- prior_completed: 0.2.1 @ sync_commit_sha 624ae8491 (§E.4 preserved below — the SPEC remains **V3R6** modern-era via §E.2 + §E.4 + sync_commit_sha; during the amendment, frontmatter status is `in-progress`, so the `internal/spec/audit.go` completed-no-drift predicate does NOT fire and normal drift detection resumes — no Go change required).
- amendment REQ delta: +5 (REQ-GLE-030..034) → 34 total REQs.
- amendment AC delta: +5 (AC-GLE-035..039) → 39 total ACs.
- verified defects (domain tooling, this tree — not text inference): (1) `grep goalCmd internal/cli/` → 0 command hits (arm CLI absent — only a `--goal` flag string in `handoff.go`); (2) `grep PruneOrphans internal/ cmd/` non-test → definition only (`internal/goal/prune.go`), ZERO call sites; (3) `ClearGoal` `os.Remove`s the file (delete, not tombstone) → confirms the §D.6 resume-deferral rationale.
- run-phase ownership boundary: the §E.2/§E.3 evidence below is the PRIOR run (v0.2.x). The amendment's NEW REQ-GLE-030..034 + AC-GLE-035..039 are implemented in a FRESH run-phase by manager-develop AFTER plan re-audit + Implementation Kickoff Approval. The §E.2/§E.3/§E.4 markers below are PRESERVED verbatim (era classification + prior-run provenance); they are NOT this amendment's run evidence and were NOT modified by manager-spec.
- residual-risk (D5, plan-auditor iter-2 — session-id convergence is conditional): the arm-CLI session id (`resolveCurrentSessionID`, `internal/cli/session.go:214`) and the `stop-goal` hook session id (Stop-hook stdin `input.SessionID`) converge on the SAME `.moai/state/goal/<id>.json` file ONLY because the SessionStart hook mirrors `input.SessionID` into the side-channel file `.moai/state/current-session-id.txt` (`internal/hook/session_start.go:263`, gated on `input.SessionID != ""`), which `resolveCurrentSessionID` reads (verified this tree). Reachability of REQ-GLE-033 is therefore CONDITIONAL on the runtime exposing `session.id`: when `input.SessionID == ""` the side-channel write is bypassed (`session_start.go:80-83`) and `resolveCurrentSessionID` degrades to the environment-fallback (`CanonicalFallbackSessionID`, source="fallback") — the documented degrade (§D.2 edge case + REQ-GLE-008 `WriterPidKey()` fallback). AC-GLE-037 pins the no-silent-pid-fallback property when a real id IS resolvable; the no-id degrade is out of that AC's scope by design.

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

### Amendment 0.3.0 (M8) run-phase evidence — arm CLI + prune wiring (REQ-GLE-030..034)

M8 implemented (TDD, cycle_type=tdd) on a worktree fast-forwarded to the current
main HEAD (which carries the M1–M7 engine). Files created/modified by this
amendment run:

- NEW `internal/cli/goal.go` — `moai goal` cobra command under `rootCmd`
  (arm/status/clear; bare `goal "<cond>"` aliases arm; NO resume). Reuses the
  `internal/goal` engine (`NewGoal`/`SaveGoal`/`LoadGoal`/`ClearGoal`); no engine
  rewrite. Registered via `init()` → `rootCmd.AddCommand(newGoalCmd())`.
- NEW `internal/cli/goal_test.go` — AC-035/036/037/039 + status/clear round-trip
  + model-condition parse + `--all`/`--json` + no-session-id degrade tests.
- EXTEND `internal/hook/session_start.go` — `pruneGoalOrphans` + `activeGoalSessionIDs`
  helpers; `goal.PruneOrphans(` call site wired on the session-start path
  (fail-open); imports `internal/goal`.
- NEW `internal/hook/session_start_goal_prune_test.go` — AC-038 (orphan → consumed/)
  + fail-open test.
- EXTEND `.claude/skills/moai/workflows/goal.md` + template mirror — annotate
  `resume` verb as deferred / out-of-scope (§D.6); §25-neutral (no SPEC IDs).
- REGEN `internal/template/catalog.yaml` (goal.md mirror hash via `make build`).

#### AC PASS/FAIL matrix (AC-GLE-035..039)

| AC | Status | Verification |
|----|--------|--------------|
| 035 | PASS | `go run ./cmd/moai goal --help` exit 0; independent `grep -qw` → arm/status/clear all PRESENT (`ALL3_PRESENT`); `TestGoalCmdListsDeliveredVerbs` ok |
| 036 | PASS | `go test ./internal/cli/ -run TestGoalArmEvalLinkage` → ok. Drives `goal arm "false exits 0" --session X` THROUGH the registered rootCmd; asserts `.moai/state/goal/X.json` written (exact path, not pid-*); `LoadGoal(root,"X")` returns armed goal (1 mechanical cond, cmd=`false`, ceiling 30); then `runStopGoalHook` given `{"session_id":"X"}` emits `"decision":"block"` — arm↔eval share the SAME file |
| 037 | PASS | `go test ./internal/cli/ -run TestGoalArmResolvesSessionId` → ok. Side-channel `current-session-id.txt`=`real-sess-77`; arm without `--session` writes `real-sess-77.json`, NO pid-*.json (no silent pid fallback) |
| 038 | PASS | (a) `grep -nE 'goal\.PruneOrphans\(' internal/hook/session_start.go \| grep -vE ':[0-9]+:[[:space:]]*//' \| wc -l` → 1 (was 0); (b) `go test ./internal/hook/ -run 'TestSessionStartPrunesGoalOrphans\|TestSessionStartGoalPruneFailOpen'` → ok (orphan → consumed/; prune error does not block session start) |
| 039 | PASS | (a) `go run ./cmd/moai goal --help \| grep -qw resume` → RESUME_ABSENT; (b) `grep -Eic 'resume[^.]*(defer\|out of scope\|follow-up)\|(defer\|out of scope\|follow-up)[^.]*resume' goal.md` → 1 (local + template) |

All 39 ACs hold (AC-GLE-001..034 from the prior run remain PASS — the amendment
ADDED code without touching the engine, hook verb, or their tests; full suite 99
pkgs ok / 0 FAIL confirms no regression).

## §E.3 Run-phase Audit-Ready Signal

- run_complete_at: 2026-07-12
- run_commit_sha: 23a9a7249 (single run-phase commit; SHA backfilled post-land)
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

### Amendment 0.3.0 (M8) run-phase audit-ready signal

- run_complete_at: 2026-07-12
- run_commit_sha: 3f68742d4 (amendment run-phase commit; cherry-picked from worktree b136ae0d8 onto current main by orchestrator per the stale-worktree cherry-pick pattern; SHA backfilled post-land)
- run_status: audit-ready (39/39 AC PASS — AC-GLE-035..039 NEW + AC-GLE-001..034 preserved)
- ac_pass_count: 5 (amendment-new; 34 prior preserved)
- ac_fail_count: 0
- preserve_list_post_run_count: 4 (internal/goal engine schema.go/state.go/prune.go/evaluate.go; internal/cli/hook_stop_goal.go stop-goal verb; goal.md skill body minus resume-deferral; REQ-GLE-001..029 + AC-GLE-001..034)
- l44_pre_commit_fetch: n/a (L1 worktree fast-forwarded to current main c13b0fd26; no remote fetch — local unpushed per repo posture)
- l44_post_push_fetch: n/a (no push — amendment commit left local unpushed per orchestrator directive)
- new_warnings_or_lints_introduced: 0 (`golangci-lint run --timeout=3m` → 0 issues, baseline 0; `go vet` implicit in build clean)
- cross_platform_build:
  - `go build ./...` → exit 0
  - `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- coverage: `go test -cover ./internal/goal/...` → 86.5% (AC-024 ≥85); `internal/cli/goal.go` per-func: parseCondition/newGoalCmd/goalProjectRoot/resolveArmSessionID 100%, runGoalArm 83.3%, runGoalStatus 81.8%, runGoalStatusAll 78.6%, runGoalClear 80.0%, printGoalHuman 84.6%; new `session_start.go` helpers: pruneGoalOrphans 100%, activeGoalSessionIDs 88.9%
- total_run_phase_files: 6 (NEW internal/cli/goal.go + internal/cli/goal_test.go + internal/hook/session_start_goal_prune_test.go; EXTEND internal/hook/session_start.go; EXTEND .claude/skills/moai/workflows/goal.md × 2 (local+template); REGEN internal/template/catalog.yaml)
- m1_to_mN_commit_strategy: single amendment run-phase commit (M8). Frontmatter already `status: in-progress` (set by manager-spec at the 0.3.0 completed→in-progress amendment re-entry) — NO status transition performed by this run.
- subagent_boundary_grep: `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/goal.go internal/hook/session_start.go | grep -v _test.go | grep -v '//'` → 0 (REQ-GLE-014 / C-HRA-008 preserved)
- template_neutrality: `grep -rn 'SPEC-GOAL-ENGINE\|SPEC-ANALYZE-FIRST\|AGENTIC-CORE\|REQ-GLE' internal/template/templates/.claude/` → 0
- template_mirror: goal.md resume-deferral mirrored to `internal/template/templates/.claude/skills/moai/workflows/goal.md`; `make build` → exit 0; mirror-parity tests ok
- full_suite_test: `go test ./...` → 99 packages ok / 0 FAIL

## §E.4 Sync-phase Audit-Ready Signal

- sync_complete_at: 2026-07-12
- sync_commit_sha: 624ae8491
- sync_status: audit-ready
- changelog_entry_added: true (SPEC-GOAL-ENGINE-001 entry added to CHANGELOG.md [Unreleased] ### Added section)
- readme_updated: true (README.md + README.ko.md updated with `/moai goal` subcommand)
- mx_tag_validation: PASS (0 AskUserQuestion/mcp__askuser matches in internal/goal/ + hook_stop_goal.go + handle-stop-goal.sh non-test code)
- spec_lint: `moai spec lint spec.md` → 0 errors
- template_neutrality: `grep -rn 'SPEC-GOAL-ENGINE\|SPEC-ANALYZE-FIRST\|AGENTIC-CORE\|REQ-GLE' internal/template/templates/.claude/` → 0
- cross_platform_build:
  - `go build ./...` → exit 0
  - `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- full_suite_test: `go test ./...` → 96 packages ok / 0 FAIL

### Amendment 0.3.0 sync-phase close signal (this section, distinct from the 0.2.1 close above)

- sync_complete_at: 2026-07-12
- sync_commit_sha: pending-backfill-sync (amendment sync commit — cannot know its own SHA; backfilled in a follow-up commit per the SHA placeholder backfill exemption, spec-frontmatter-schema.md § Forbidden ownership crossings)
- sync_status: audit-ready
- amendment_version: "0.3.0" (distinct from the 0.2.1 close recorded above; frontmatter transition this commit: `in-progress → completed`)
- changelog_entry_added: true (new distinct CHANGELOG.md `[Unreleased]` ### Added entry for the 0.3.0 amendment — arm CLI + prune wiring reachability fix; does NOT duplicate the existing 0.2.1 entry. `grep -c 'SPEC-GOAL-ENGINE-001' CHANGELOG.md` 2 → 3)
- readme_updated: false (no user-facing README change required — internal CLI reachability fix only; a separate untracked README redesign in the working tree was left untouched per scope discipline)
- mx_tag_validation: PASS (0 AskUserQuestion/mcp__askuser matches in internal/cli/goal.go + internal/hook/session_start.go non-test code, re-verified this sync)
- spec_lint: `moai spec lint spec.md` → expect StatusGitConsistency warning to clear now that status is `completed`
- template_neutrality: `grep -rn 'SPEC-GOAL-ENGINE\|SPEC-ANALYZE-FIRST\|AGENTIC-CORE\|REQ-GLE' internal/template/templates/.claude/` → 0 (unchanged)
- ac_count: 39/39 AC PASS (AC-GLE-001..039, acceptance.md SSOT — amendment adds AC-GLE-035..039 reachability pins on top of the preserved AC-GLE-001..034)
- files_touched_this_sync: spec.md (frontmatter status/updated only), progress.md (this §E.4 addendum), CHANGELOG.md (one new [Unreleased] entry) — no body edits to spec.md/plan.md/acceptance.md, no README edit, no run-phase code touched

## §F Phase 0.95 Mode Selection

- Decision: sub-agent (Mode 5 — coding-heavy, single-domain, bounded)

Recorded for the Amendment 0.3.0 (M8) run-phase (arm CLI + prune wiring). The
prior M1–M7 run predated this logging section.

- Input parameters: tier = L (retained for PASS threshold + Section A-E), but the
  amendment scope is small — 6 files, single domain (Go CLI + one hook wiring +
  one doc/mirror), all sequential/coding-heavy with inter-file dependency
  (arm CLI depends on the engine; the E2E test drives arm→hook in one flow).
  concurrency benefit = LOW (coding-heavy, not research fan-out).
- Mode evaluation:
  - Mode 1 (trivial): not selected — multi-file behavioral change, not a typo.
  - Mode 2 (background): not selected — write-capable implementation, not read-only.
  - Mode 3 (agent-team): RETIRED — never selected.
  - Mode 4 (parallel): not selected — coding-heavy single-domain work (Anthropic
    coding-task parallelism caveat), no ≥3-domain research fan-out.
  - Mode 5 (sub-agent): SELECTED — sequential coding, one implementation agent.
  - Mode 6 (workflow): not selected — < 30 files, not a uniform mechanical sweep.
- Decision: sub-agent
- Justification: The amendment is coding-heavy, single-domain, and bounded (6
  files, arm↔engine↔hook inter-dependency), so Mode 5 (sequential sub-agent) is
  the correct default per Anthropic's coding-task parallelism caveat — parallel /
  workflow fan-out earns its overhead only on genuinely-parallel high-volume
  mechanical work, which this is not.
