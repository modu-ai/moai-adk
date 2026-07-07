# Progress — SPEC-MOAI-AGENTIC-LOOP-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_complete_at: 2026-07-07
- plan_audit_iter1: FAIL 0.83 (Tier L threshold 0.85; MP 4/4 PASS; Testability 0.70 drag; §D.6 4-artifact deviation ACCEPTED)
- plan_audit_iter1_fixes: D1-D10 + B8/M2 budget clarification applied 2026-07-07 (plan-phase artifact edits only; no skill files touched)
- plan_audit_iter2: PASS-WITH-DEBT 0.89 (all D1-D10 resolved; two literal-level debts)
- plan_audit_iter2_debts: D11 (AC-MAL-016 regex collateral match on harness-level line — threshold-context-tightened, verified 4 in-scope matches / 0 collateral) + D1r (plan.md §E E2 diff form aligned to `--exclude=.moai`, exclusion noted as load-bearing) — resolved 2026-07-07
- plan_status: audit-ready (iter-2 PASS-WITH-DEBT 0.89, debts D11/D1r resolved)
- artifacts: spec.md + plan.md + acceptance.md + progress.md (4 — Tier L artifact-count deviation disclosed in spec.md §D.6)

## §E.2 Run-phase Evidence

- baseline_sha: f7b55e637b3b8d0affb1e24176db84934e3064e1
- baseline_pinned_at: 2026-07-07 (M1 pre-flight — HEAD of main at run-phase start; `git rev-list --count --left-right origin/main...HEAD` → `0 0`)

### M1 pre-flight records

- **orchestration-mode-selection.md mirror divergence characterization (REQ-MAL-042 / AC-MAL-R03)**: pre-flight `diff .claude/rules/moai/workflow/orchestration-mode-selection.md internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md` → exactly 1 content hunk (`333c333`, §I.3 "Installation status (honest reference)" paragraph — live copy names maintainer-repo template paths + wrapper + Windows variant; template copy carries the neutral shortened variant) + 1 trailing blank line at template EOF (`334a335`). Classification: pre-existing INTENTIONAL template-neutrality variant, NOT stale content — a full re-sync would either leak maintainer-specific path detail into the shipped template or strip valid detail from the live copy. Resolution: this SPEC's delta (§G.1 crosswalk + §B.1 threshold-SSOT sentence) landed IDENTICALLY in both copies; the pre-existing §I.3 hunk is preserved unchanged in both. Residual divergence documented here as the structured report (in lieu of full re-sync); no silent divergence growth — the delta regions diff clean.
- **Junk-dir cleanup (plan.md §F M1(b))**: `internal/template/templates/.claude/skills/moai/workflows/.moai/` (untracked statusline runtime-state leak; `git ls-files` empty) deleted best-effort at M1. The `--exclude=.moai` flag in parity checks remains the load-bearing mechanism — a live statusline process may re-create the directory.

### M6 AC matrix (38/38 — commands per acceptance.md §B, outputs verbatim, run at M6 2026-07-08)

| AC | Command (shorthand S/R/T per acceptance.md) | Actual Output | Expected | Status |
|----|---------------------------------------------|---------------|----------|--------|
| AC-MAL-001 | `grep -ci "requirement analysis" S/SKILL.md` | `1` | ≥1 | PASS |
| AC-MAL-002 | `grep -c "Requirement Analysis & Completion Condition" S/SKILL.md` | `1` | ≥1 | PASS |
| AC-MAL-003 | `grep -c "goal-directive" S/SKILL.md S/workflows/moai.md` | SKILL=1, moai=2 | each ≥1 | PASS |
| AC-MAL-004 | `grep -ci "pipeline contract" ...` | SKILL=1, moai=2 | each ≥1 | PASS |
| AC-MAL-005 | `grep -ci "crosswalk" R/orchestration-mode-selection.md` | `2` | ≥1 | PASS |
| AC-MAL-005b | `grep -ci "crosswalk" T/...orchestration-mode-selection.md` | `2` | ≥1 | PASS |
| AC-MAL-006 | `grep -c "5-mode" S/workflows/run.md` | `0` | 0 | PASS |
| AC-MAL-007 | `grep -c "orchestration-mode-selection" ...` | SKILL=4, moai=4, run=4 | each ≥1 | PASS |
| AC-MAL-008 | `grep -ci "3-5 concurrent" run.md + phase-execution.md` | run=1, pe=0 (sum 1) | sum ≥1 | PASS |
| AC-MAL-009 | `grep -icE "Mode 6\|workflow fan-out" moai.md` | `1` | ≥1 | PASS |
| AC-MAL-009b | `grep -c "v2.1.154" run.md moai.md` | run=1, moai=1 (sum 2) | sum ≥1 | PASS |
| AC-MAL-009c | `grep -ci "launch procedure" run.md` | `1` | ≥1 | PASS |
| AC-MAL-010 | `grep -c "plan-auditor" moai.md` | `3` | ≥1 | PASS |
| AC-MAL-010b | `grep -ci "Implementation Kickoff Approval" moai.md` | `6` | ≥1 | PASS |
| AC-MAL-010c | `grep -c "sync-auditor" moai.md` | `3` | ≥2 | PASS |
| AC-MAL-011 | `grep -ci "exactly once" moai.md` | `2` | ≥1 | PASS (BLOCKING) |
| AC-MAL-011b | `grep -icE "never authorize\|does NOT authorize" moai.md` | `1` | ≥1 | PASS (BLOCKING) |
| AC-MAL-012 | `grep -c "agentic_loop" moai.md workflow.yaml` | moai=1, yaml=2 | each ≥1 | PASS |
| AC-MAL-012b | `grep -A2 "agentic_loop:" workflow.yaml \| grep -c "max_iterations: 10"` | `1` | ≥1 | PASS |
| AC-MAL-012c | `grep -A2 "agentic_loop:" T/workflow.yaml \| grep -c "max_iterations"` | `2` | ≥1 | PASS |
| AC-MAL-013 | `grep -icE "no.progress\|same failure" moai.md` | `3` | ≥1 | PASS |
| AC-MAL-013b | `grep -icE "per-iteration\|every iteration" moai.md` | `1` | ≥1 | PASS |
| AC-MAL-014 | `grep -ci "semantic failure" moai.md` | `2` | ≥1 | PASS |
| AC-MAL-015 | `grep -icE "session-handoff\|paste-ready" moai.md` | `1` | ≥1 | PASS |
| AC-MAL-016 | threshold-context regex on SKILL.md + moai.md | `0` (pre-edit: exactly the 4 in-scope sites; harness-level `file_count >= 10 AND multi_domain (domain_count >= 2)` line untouched, no collateral) | 0 | PASS |
| AC-MAL-016b | `grep -c "auto_selection" R/orchestration-mode-selection.md` | `1` | ≥1 | PASS |
| AC-MAL-017 | `grep -c "AND --team flag" phase-execution.md` | `0` | 0 | PASS |
| AC-MAL-018 | `grep -icE "agentic\|pipeline-level" loop.md` | `2` | ≥1 | PASS |
| AC-MAL-018b | `grep -c "run --mode loop" loop.md` | `4` | ≥1 | PASS |
| AC-MAL-019 | `grep -icE "auto-chain\|chain" sync.md` | `1` | ≥1 | PASS |
| AC-MAL-020 | `grep -ci "blocker report" moai.md` | `1` | ≥1 | PASS (BLOCKING) |
| AC-MAL-021 | `diff -rq --exclude=.moai S/ T/.claude/skills/moai` | verbatim: `Only in .claude/skills/moai/workflows: .run.md.swp` (exit 1) — a LIVE vim swap (lsof: `vim 7572 goos 5u`, user's open editor, created mid-run, NOT repo content). Plan-sanctioned alternative (plan.md §A.4 "or git-tracked file lists"): swp-excluded diff exit 0 + git-tracked pair sweep `cmp` 42/42 files → 0 DIFFERS | empty/exit 0 over git-tracked content | PASS (BLOCKING — git-tracked byte parity proven; verbatim form blocked only by a live-editor runtime artifact, same class as the `--exclude=.moai` leak) |
| AC-MAL-022 | `grep -rn "SPEC-MOAI-AGENTIC-LOOP\|REQ-MAL-" .claude/skills/ .claude/rules/ internal/template/templates/` | `0` matches | 0 | PASS (BLOCKING) |
| AC-MAL-023 | `grep -c "MODE_PIPELINE_ONLY_UTILITY" context-loading.md plan.md sync.md` | ctx=3, plan=1, sync=1 | each ≥1 | PASS |
| AC-MAL-023b | `grep -c "MODE_UNKNOWN" run.md` | `1` | ≥1 | PASS |
| AC-MAL-024 | `grep -rn -- "--mode workflow" S/` | `0` | 0 | PASS |
| AC-MAL-025 | `git diff --stat <baseline_sha>..HEAD -- '*.go'` (baseline f7b55e637 dereferenced from §E.2 pin) | empty output | empty | PASS |
| AC-MAL-026 | `moai spec lint` (scoped to this SPEC) | ERROR rows for SPEC-MOAI-AGENTIC-LOOP-001: `0`; 1× WARNING `StatusGitConsistency` (frontmatter `in-progress` vs git-implied `implemented` — expected run-phase transient, resolves at sync close) | 0 errors | PASS |

### §C review-verified criteria

| AC | Verdict | Evidence |
|----|---------|----------|
| AC-MAL-R01 | PASS | §G.1 crosswalk maps ALL 4 `--mode` values (autopilot/loop/team/pipeline) + ALL 5 scale labels (Fix/Focused/Standard/Full Pipeline/Team), each to exactly one catalog mode; "Correspondence, not merge" banner present; §G two-axis separation restated (no `--mode workflow`, sentinel set untouched) |
| AC-MAL-R02 | PASS | SKILL.md Step 2.8 states the trivial-exemption list (feedback/gate/codemaps/sync-status + Stage-1-Clarify exceptions) AND Socratic-first ordering ("BEFORE deriving the completion condition") |
| AC-MAL-R03 | PASS | Pre-flight divergence diff recorded above (M1 pre-flight records) — resolution: delta in both copies + residual §I.3 hunk characterized as intentional neutrality variant |
| AC-MAL-R04 | PASS | moai.md loop section defines entry (post-kickoff only), iteration (run→sync→verify ONLY; plan re-entry solely via user-approved escalation re-crossing Implementation Kickoff Approval; no autonomous plan→run re-cross), and all 4 termination causes (condition met / ceiling / escalation / context suspension) |

### spec.md §C invariant verification

| Invariant | Status | Evidence |
|-----------|--------|----------|
| 1. Kickoff exactly-once, loop post-kickoff only, condition never authorizes run-entry | PASS | AC-MAL-011/011b + Pipeline Gates #2 + SKILL.md Step 2.8 closing line |
| 2. Subagents never AskUserQuestion (blocker-report boundary) | PASS | AC-MAL-020 + moai.md loop Boundary bullet |
| 3. Mirror parity + template neutrality | PASS | AC-MAL-021/022 |
| 4. 6-mode catalog sole SSOT; crosswalk = correspondence not merge; thresholds once + cross-referenced | PASS | AC-MAL-005/011(R01)/013(016)/016b/024 |
| 5. Loop safety (ceiling / no-progress / dark-flow / handoff) | PASS | AC-MAL-012b/013/013b/015 |
| 6. Markdown-only scope, zero Go changes | PASS | AC-MAL-025 (empty Go diff vs baseline) |

## §E.3 Run-phase Audit-Ready Signal

- run_complete_at: 2026-07-08
- run_commit_sha: a8f6bc629
- run_commits: 5022bdf63 (M1 rules SSOT + baseline) → d7fba8451 (M2 router) → ee8ca76a9 (M3 pipeline body) → 249cf8e08 (M4 run tree) → 8578c2a47 (M5 adjacent surfaces) → M6 (evidence)
- run_status: implemented (progress.md-scoped signal; spec/plan/acceptance frontmatter close is sync-phase-owned per the Status Transition Ownership Matrix)
- ac_pass_count: 38 (§B 38/38 PASS; BLOCKING group 011/011b/020/021/022 all PASS) + §C R01-R04 4/4 PASS
- ac_fail_count: 0
- preserve_list_post_run_count: intact — MODE_UNKNOWN=1(run.md), MODE_TEAM_UNAVAILABLE=1(run.md), MODE_PIPELINE_ONLY_UTILITY ctx=3/plan=1/sync=1, `/moai run --mode loop` alias=4(loop.md), ac_converge=5(run.md), gate-run-1=1, gate-sync-1/2=3(sync.md), harness-extension include=1(run.md), evolvable markers untouched
- l44_pre_commit_fetch: `git rev-list --count --left-right origin/main...HEAD` → `0 0` at M1 start (baseline f7b55e637)
- l44_post_push_fetch: n/a — push deliberately NOT performed (orchestrator instruction: per-milestone local commits only; Tier L PR routing handled downstream)
- new_warnings_or_lints_introduced: 0 errors for this SPEC; 1 expected transient WARNING (StatusGitConsistency: `in-progress` vs git-implied `implemented` — resolves at the sync-phase close transition)
- cross_platform_build: n/a (markdown/YAML-only scope; `git diff --stat f7b55e637..HEAD -- '*.go'` empty per AC-MAL-025)
- total_run_phase_files: 20 committed across M1-M5 (live: orchestration-mode-selection.md, SKILL.md, moai.md, run.md, phase-execution.md, context-loading.md, mode-orchestration.md, loop.md, sync.md, workflow.yaml + their 10 template counterparts) + 4 SPEC artifacts
- m1_to_mN_commit_strategy: one Conventional Commit per milestone (M1..M6), path-limited `git add` only (shared-checkout race guard), `--no-verify` never used, no push

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
