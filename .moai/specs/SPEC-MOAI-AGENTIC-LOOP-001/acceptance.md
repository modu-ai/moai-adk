---
id: SPEC-MOAI-AGENTIC-LOOP-001
title: "Acceptance criteria — /moai agentic completion loop"
version: "0.1.0"
status: completed
created: 2026-07-07
updated: 2026-07-08
author: manager-spec
priority: P1
phase: "v3.x"
module: ".claude/skills/moai"
lifecycle: spec-anchored
tags: "acceptance, orchestration, agentic-loop"
---

# Acceptance Criteria — SPEC-MOAI-AGENTIC-LOOP-001

All file paths relative to repo root. Commands are the verification mechanism; PASS requires the stated expected output, reported verbatim per verification-claim-integrity §3.

## §A. Given-When-Then Scenarios

### Scenario 1 — Full-pipeline natural-language request (happy path)

- **Given** a user issues `/moai "add a retry wrapper to the fetcher"` (default route) and project docs exist
- **When** the router executes
- **Then** a requirement-analysis record is produced (summary + `/goal`-form completion condition + `full-pipeline` contract + shape pre-signal) BEFORE the workflow body loads; the pipeline declares plan-audit gate → Implementation Kickoff Approval (exactly once) → Phase 0.95 (6-mode catalog) → run → auto-chained sync → sync-audit gate; the loop iterates with per-iteration visible reports until the condition evaluates met, then terminates and reports completion.

### Scenario 2 — Explicit single-phase invocation (no silent chain)

- **Given** a user issues `/moai run SPEC-XYZ-001`
- **When** run-phase completes successfully
- **Then** the pipeline contract is `single-phase`; sync is offered as the "(Recommended)" first option of the next-step AskUserQuestion; no auto-chain fires without user selection.

### Scenario 3 — Edge: no-progress escalation

- **Given** the agentic loop is active post-kickoff
- **When** the same failure signature (same failing check + same error class) appears in two consecutive iterations
- **Then** the loop halts, a structured report is surfaced, and the orchestrator runs AskUserQuestion (continue manually / revert + re-plan / abort). No third identical iteration is attempted.

### Scenario 4 — Edge: context-threshold suspension

- **Given** the loop is active and context usage crosses the model-specific handoff threshold
- **When** the current iteration boundary is reached
- **Then** the loop suspends (no new iteration), state persists to progress.md, and a paste-ready resume message is emitted per session-handoff format.

### Scenario 5 — Edge: kickoff gate never bypassed

- **Given** a derived completion condition exists and the plan-auditor returned PASS with a skip-eligible score
- **When** the pipeline reaches the plan→run boundary
- **Then** the Implementation Kickoff Approval AskUserQuestion round is STILL issued (score-independent, exactly once); the condition/goal is set only AFTER user approval; a `/goal` line never substitutes for the gate.

### Scenario 6 — Edge: trivial-route exemption

- **Given** a user issues `/moai feedback "typo in docs"`
- **When** the router executes
- **Then** requirement-analysis emission is skipped (trivial exemption) and the feedback workflow proceeds directly.

## §B. AC Matrix (machine-verifiable)

Path shorthand: `S=` `.claude/skills/moai`; `R=` `.claude/rules/moai/workflow`; `T=` `internal/template/templates`.

| AC | Verifies (REQ) | Command | Expected |
|----|----------------|---------|----------|
| AC-MAL-001 | 001 | `grep -ci "requirement analysis" S/SKILL.md` | ≥ 1 |
| AC-MAL-002 | 001, 002 | `grep -c "Requirement Analysis & Completion Condition" S/SKILL.md` | ≥ 1 (the new step-heading literal — the generic phrase "completion condition" pre-exists ×2 in SKILL.md and is NOT a valid signal) |
| AC-MAL-003 | 002, 003 | `grep -c "goal-directive" S/SKILL.md S/workflows/moai.md` | each file ≥ 1 |
| AC-MAL-004 | 001c, 030, 031 | `grep -ci "pipeline contract" S/SKILL.md S/workflows/moai.md` | each file ≥ 1 |
| AC-MAL-005 | 011 | `grep -ci "crosswalk" R/orchestration-mode-selection.md` | ≥ 1 (new section present) |
| AC-MAL-005b | 042 | `grep -ci "crosswalk" T/.claude/rules/moai/workflow/orchestration-mode-selection.md` | ≥ 1 (delta landed in mirror) |
| AC-MAL-006 | 012 | `grep -c "5-mode" S/workflows/run.md` | 0 |
| AC-MAL-007 | 010, 012 | `grep -c "orchestration-mode-selection" S/SKILL.md S/workflows/moai.md S/workflows/run.md` | each file ≥ 1 |
| AC-MAL-008 | 014 | `grep -ci "3-5 concurrent" S/workflows/run.md S/workflows/run/phase-execution.md` | sum ≥ 1 |
| AC-MAL-009 | 015 | `grep -icE "Mode 6\|workflow fan-out" S/workflows/moai.md` | ≥ 1 (moai.md operational entry) |
| AC-MAL-009b | 015 | `grep -c "v2.1.154" S/workflows/run.md S/workflows/moai.md` | sum ≥ 1 (capability gate stated) |
| AC-MAL-009c | 015 | `grep -ci "launch procedure" S/workflows/run.md` | ≥ 1 (operational-entry discriminator — the bare token "Mode 6" pre-exists ×4 in run.md as mentions; the launch-procedure literal distinguishes an operational entry from a mention) |
| AC-MAL-010 | 021 | `grep -c "plan-auditor" S/workflows/moai.md` | ≥ 1 |
| AC-MAL-010b | 021 | `grep -ci "Implementation Kickoff Approval" S/workflows/moai.md` | ≥ 1 |
| AC-MAL-010c | 021 | `grep -c "sync-auditor" S/workflows/moai.md` | ≥ 2 (Phase 0 exploration + pipeline gate) |
| AC-MAL-011 | 020 | `grep -ci "exactly once" S/workflows/moai.md` | ≥ 1 (kickoff-once invariant literal) |
| AC-MAL-011b | 020 | `grep -icE "never authorize\|does NOT authorize" S/workflows/moai.md` | ≥ 1 (condition ≠ run-entry authorization) |
| AC-MAL-012 | 023 | `grep -c "agentic_loop" S/workflows/moai.md .moai/config/sections/workflow.yaml` | each file ≥ 1 (key documented + present) |
| AC-MAL-012b | 023 | `grep -A2 "agentic_loop:" .moai/config/sections/workflow.yaml \| grep -c "max_iterations: 10"` | ≥ 1 (block-scoped — a bare `max_iterations` grep is vacuous: `loop_prevention.max_iterations: 100` pre-exists in the same file) |
| AC-MAL-012c | 023, 040 | `grep -A2 "agentic_loop:" T/.moai/config/sections/workflow.yaml \| grep -c "max_iterations"` | ≥ 1 (template counterpart gains the block — independent edit, not byte-parity-bound) |
| AC-MAL-013 | 024 | `grep -icE "no.progress\|same failure" S/workflows/moai.md` | ≥ 1 |
| AC-MAL-013b | 025 | `grep -icE "per-iteration\|every iteration" S/workflows/moai.md` | ≥ 1 (dark-flow guard: visible iteration report) |
| AC-MAL-014 | 026 | `grep -ci "semantic failure" S/workflows/moai.md` | ≥ 1 |
| AC-MAL-015 | 027 | `grep -icE "session-handoff\|paste-ready" S/workflows/moai.md` | ≥ 1 |
| AC-MAL-016 | 013 | `grep -rnE "(domains? ?(>=\|≥) ?3\|(>=\|≥) ?3 domains\|files ?(>=\|≥) ?10\|\((>=\|≥)10\)\|score ?(>=\|≥) ?7\|\((>=\|≥)7\))" S/SKILL.md S/workflows/moai.md \| wc -l` | 0 (threshold-context-tightened: pre-edit this regex matches EXACTLY the 4 in-scope threshold sites — SKILL.md "Execution Mode Flags" line + moai.md ×3, all phrasing variants — and does NOT collaterally match moai.md's harness-level auto-detection line `file_count >= 10 AND multi_domain (domain_count >= 2)`, which belongs to a different SSOT (harness level, not orchestration-mode-selection §B.1) and stays untouched) |
| AC-MAL-016b | 013 | `grep -c "auto_selection" R/orchestration-mode-selection.md` | ≥ 1 (prose SSOT cites machine SSOT) |
| AC-MAL-017 | 016 | `grep -c "AND --team flag" S/workflows/run/phase-execution.md` | 0 (Team-row drift reconciled to capability gate) |
| AC-MAL-018 | 029 | `grep -icE "agentic\|pipeline-level" S/workflows/loop.md` | ≥ 1 (relationship paragraph) |
| AC-MAL-018b | 029 | `grep -c "run --mode loop" S/workflows/loop.md` | ≥ 1 (alias contract preserved) |
| AC-MAL-019 | 030, 032 | `grep -icE "auto-chain\|chain" S/workflows/sync.md` | ≥ 1 (chain-entry note) |
| AC-MAL-020 | 028 | `grep -ci "blocker report" S/workflows/moai.md` | ≥ 1 |
| AC-MAL-021 | 040 | `diff -rq --exclude=.moai .claude/skills/moai internal/template/templates/.claude/skills/moai` | empty output, exit 0 (byte parity over git-tracked content; `--exclude=.moai` masks the UNTRACKED runtime-state leak dir `T/.claude/skills/moai/workflows/.moai/`, which M1 pre-flight also deletes — see plan.md §F M1) |
| AC-MAL-022 | 041 | `grep -rn "SPEC-MOAI-AGENTIC-LOOP\|REQ-MAL-" .claude/skills/ .claude/rules/ internal/template/templates/` | 0 matches (neutrality; `.moai/specs/` is outside the searched roots) |
| AC-MAL-023 | PRESERVE | `grep -c "MODE_PIPELINE_ONLY_UTILITY" S/workflows/run/context-loading.md S/workflows/plan.md S/workflows/sync.md` | each file ≥ 1 (sentinels intact) |
| AC-MAL-023b | PRESERVE | `grep -c "MODE_UNKNOWN" S/workflows/run.md` | ≥ 1 |
| AC-MAL-024 | 011 (§G frozen) | `grep -rn -- "--mode workflow" S/` | 0 (no new --mode axis value) |
| AC-MAL-025 | scope | `git diff --stat $(grep -oE "baseline_sha: [0-9a-f]+" .moai/specs/SPEC-MOAI-AGENTIC-LOOP-001/progress.md \| awk '{print $2}')..HEAD -- '*.go'` | empty (zero Go changes). Baseline pinning mechanism: run-phase M1 pre-flight records `baseline_sha: <pre-SPEC HEAD SHA>` in progress.md §E.2 BEFORE the first edit; this AC dereferences that pinned SHA (no floating `<pre-SPEC-baseline>` placeholder) |
| AC-MAL-026 | artifacts | `moai spec lint` (scoped to this SPEC) | 0 errors (frontmatter schema + OutOfScopeRule pass) |

> **AC-MAL-022 transitive-coverage note (neutrality scope)**: the grep targets this SPEC's ID and REQ tokens directly. The remaining forbidden content classes for shipped surfaces (internal dates C3, commit SHAs C7) are transitively covered by (a) AC-MAL-021 mirror parity — live and template copies are byte-identical, so template-side scanning covers both — and (b) the `internal_content_leak_test.go` CI guard, which owns C3/C7 scanning on template content per the template internal-content isolation doctrine.

## §C. Non-Grep (review-verified) Criteria

| AC | Verifies | Mechanism |
|----|----------|-----------|
| AC-MAL-R01 | REQ-MAL-011 crosswalk correctness | Reviewer confirms the crosswalk maps ALL 4 `--mode` values + ALL 5 scale labels, each to exactly one catalog mode, with the "correspondence, not merge" banner; consistent with §G |
| AC-MAL-R02 | REQ-MAL-004/005 | Reviewer confirms trivial-exemption list + Socratic-before-condition ordering are stated in the SKILL.md step |
| AC-MAL-R03 | REQ-MAL-042 | Pre-flight divergence diff of orchestration-mode-selection.md recorded in run-phase evidence (progress.md §E.2); resolution or blocker report present |
| AC-MAL-R04 | REQ-MAL-022 loop lifecycle | Reviewer confirms moai.md loop section defines entry (post-kickoff only), iteration (run→sync→verify ONLY — plan re-entry solely via user-approved escalation that re-crosses the Implementation Kickoff Approval gate; no autonomous plan→run re-cross), and all 4 termination causes (condition met / ceiling / escalation / context suspension) |

## §D. Quality Gate & Definition of Done

- [ ] All §B AC rows PASS with verbatim command output (E1)
- [ ] All §C review criteria confirmed with cited prose
- [ ] Mirror parity clean (AC-MAL-021) and neutrality clean (AC-MAL-022)
- [ ] All PRESERVE sentinels + gate IDs + evolvable markers intact (AC-MAL-023*)
- [ ] Zero Go diffs (AC-MAL-025); zero `.claude/agents/` diffs
- [ ] `moai spec lint` clean for this SPEC's artifacts (AC-MAL-026)
- [ ] Per-milestone Conventional Commits, path-limited adds, no `--no-verify`
- [ ] progress.md §E.2/§E.3 populated by run-phase owner (manager-develop)

Severity: AC-MAL-011/011b/020/021/022 (kickoff-once, boundary, parity, neutrality) are BLOCKING — any FAIL halts close. Remaining ACs are standard-blocking per DoD; no PASS-WITH-DEBT is pre-authorized for the [HARD]-tagged group.
