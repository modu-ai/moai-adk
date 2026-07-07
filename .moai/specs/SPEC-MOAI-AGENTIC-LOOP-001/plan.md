---
id: SPEC-MOAI-AGENTIC-LOOP-001
title: "Implementation plan — /moai agentic completion loop"
version: "0.1.0"
status: in-progress
created: 2026-07-07
updated: 2026-07-07
author: manager-spec
priority: P1
phase: "v3.x"
module: ".claude/skills/moai"
lifecycle: spec-anchored
tags: "plan, orchestration, agentic-loop, mode-catalog"
---

# Implementation Plan — SPEC-MOAI-AGENTIC-LOOP-001

## §A. Context

### §A.1 Current state (verified 2026-07-07)

- `/moai` dispatch: first-word Intent Router (`SKILL.md` L36-94) + 10-step Execution Directive (L229-305). No requirement-analysis step; mode selection described as team/solo binary (L46-52).
- `workflows/moai.md` (default pipeline, 247L): Phase 0 exploration → Phase 1 SPEC → Phase 1.5 annotation → Phase 2 implementation → Phase 3 sync. No plan-auditor/sync-auditor/kickoff-approval declaration in the body; thresholds inlined at L47/L190/L200.
- `workflows/run.md` (181L): stale "5-mode" label at L39; Implementation Kickoff Approval + `/goal ac_converge` wiring at L114-163 (this SPEC generalizes, does not move it).
- `workflows/run/phase-execution.md` L214-235: scale table (Fix/Focused/Standard/Full Pipeline/Team) with drifted Team row (`score ≥ 7 AND --team`).
- `workflows/run/context-loading.md` L35-67: `--mode` dispatch axis {autopilot, loop, team, pipeline} + resolver pseudocode + sentinels.
- Mirror baselines: skills tree 42/42 git-tracked markdown identical — parity checks use `diff -rq --exclude=.moai` because an UNTRACKED runtime-state dir (`internal/template/templates/.claude/skills/moai/workflows/.moai/`, statusline leak) sits inside the template mirror and is deleted at M1 pre-flight; `orchestration-mode-selection.md` MIRRORED-DIVERGED (pre-existing); other rules in scope identical.

### §A.2 PRESERVE list (must not regress)

- Sentinel literals: `MODE_UNKNOWN`, `MODE_TEAM_UNAVAILABLE`, `MODE_PIPELINE_ONLY_UTILITY` (CI-audited) in `run.md` / `plan.md` / `sync.md` / `context-loading.md`.
- `/moai loop` ≡ `/moai run --mode loop` alias contract + its CI-audit cross-reference text in `loop.md` L34-46.
- `run.md` § Run-phase Autonomy (`ac_converge`) section — extended by reference, not rewritten.
- All existing HUMAN GATE definitions (`gate-run-1`, `gate-sync-1`, `gate-sync-2`) and evolvable-block markers.
- `orchestration-mode-selection.md` §G two-axis warning ([ZONE:Frozen] — the crosswalk must be additive and consistent with it).
- Custom Harness Extension `@.moai/harness/*-extension.md` include lines in each workflow file.
- Frontmatter `progressive_disclosure` / `triggers` blocks in every skill file.

### §A.3 EXTEND envelope (files this SPEC may modify)

Exactly the files in the §C change map (live + template mirror pairs) plus `.moai/config/sections/workflow.yaml` (+ its template counterpart, independently — config files are not byte-parity-bound). Nothing else.

---

## §B. Known Issues / Risks

| # | Risk | Mitigation |
|---|------|------------|
| B1 | Line anchors drift between plan and run phases (parallel sessions active on this checkout) | All anchors are content-token anchored in §C; run-phase MUST re-grep tokens before editing, never trust raw line numbers |
| B2 | `orchestration-mode-selection.md` mirror pre-diverged | M1 pre-flight: `diff` both copies, record the delta; land SPEC delta in both; full-resync only if divergence is stale-content-only, else blocker report (REQ-MAL-042) |
| B3 | Neutrality violation in shipped surfaces | All new prose written generically (no SPEC ID / REQ tokens / dates); M6 neutrality grep gate before commit |
| B4 | Sentinel/CI-audit breakage from rewording | PRESERVE list §A.2; M6 grep verifies all sentinels still present |
| B5 | Scope creep into agent bodies or Go code | Out of Scope §E in spec.md; M6 `git diff --stat` gate on `*.go` and `.claude/agents/` |
| B6 | Shared-checkout parallel-session commit race | Path-limited `git add` per milestone commit; pre-spawn sync check before run-phase delegation |
| B7 | Two-axis confusion (the exact trap §G warns about) | The crosswalk table carries an explicit "correspondence, not merge" banner; no `--mode` value additions; AC verifies the `--mode` value set is unchanged |
| B8 | SKILL.md size pressure (309L; router must stay lean) | Requirement-analysis step budgeted ≤ 25 lines (STEP-LEVEL budget: the new Step 2.8 block only); detail lives in moai.md + the rule crosswalk. Distinct from the M2 TOTAL-FILE delta budget (≤ ~35 lines, which includes the mode-flags rewrite on top of the ≤25-line step) |

---

## §C. File-by-File Change Map

Every skill-file row implies a byte-identical template-mirror write (`internal/template/templates/` + same relative path). Anchors verified 2026-07-07; re-verify by token at run time.

| # | File | Anchor (token) | Change | REQ |
|---|------|----------------|--------|-----|
| 1 | `.claude/rules/moai/workflow/orchestration-mode-selection.md` | after §G / before §H | NEW §: "Dispatch-Axis Crosswalk" — table mapping `--mode` axis values + scale-table labels → catalog modes; "correspondence, not merge" banner; §B.1 gains one sentence naming `workflow.yaml` `auto_selection` as machine SSOT | 011, 013, 016 |
| 2 | `.claude/skills/moai/SKILL.md` | "Execution Mode Flags (mutually exclusive)" (L46-52); Execution Directive Step 2.5→3 boundary (L275-283) | (a) Replace binary flags prose with 6-mode catalog pointer (`--team`→Mode 3 / `--solo`→Mode 5 forced overrides; auto-select per rule §B); threshold numbers → pointer. (b) NEW Step 2.8 "Requirement Analysis & Completion Condition" (summary + condition + pipeline contract + shape pre-signal; trivial exemptions; goal-directive form) | 001-005, 010, 013 |
| 3 | `.claude/skills/moai/workflows/moai.md` | L46-48 default-behavior; L114-122 Phase 0-completion/Phase 1; L129-138 annotation; L140-163 Phase 2; L165-172 Phase 3; L174-190 Team Mode; L192-241 Execution Summary | (a) Declare named gates in pipeline body: plan-audit gate (plan-auditor) after Phase 1, Implementation Kickoff Approval (exactly once, score-independent) before Phase 2, Phase 0.95 6-mode selection, sync-audit gate (sync-auditor) after Phase 3. (b) NEW "Agentic Completion Loop" section: lifecycle (entry post-kickoff / iterate run→sync→verify ONLY — plan re-entry solely via escalation + renewed kickoff approval / terminate on condition-met, ceiling, escalation), ceiling key `workflow.agentic_loop.max_iterations` (default 10), no-progress rule, dark-flow per-iteration report, semantic-failure escalation, context-threshold suspension, blocker-report boundary. (c) run→sync chaining policy (pipeline contract). (d) Thresholds ×3 sites → pointers; Team Mode section references catalog | 010, 013, 015, 020-028, 030-032 |
| 4 | `.claude/skills/moai/workflows/run.md` | L37-41 Phase Owners ("5-mode"); L43-52 routing table; L66-85 Quick Reference; after § Run-phase Autonomy | (a) Fix stale label: Phase 0.95 = 6-mode catalog decision + rule pointer. (b) Mode 4 operational entry (research fan-out 3-5 concurrent read-only Agent(), single turn). (c) Mode 6 operational entry (§C.3 gate + progress.md §F logging + launch procedure). (d) single-phase chaining note (recommended-option, no silent chain) | 012, 014, 015, 031 |
| 5 | `.claude/skills/moai/workflows/run/phase-execution.md` | "Phase 0.95: Scale-Based Execution Mode Selection" (L214-235) | Re-express scale table in catalog terms (Fix/Focused/Standard→Mode 5 envelope variants; Full Pipeline→Mode 5 full; Team→Mode 3 via capability gate — drift fix: gate-based, not `AND --team`); crosswalk pointer to rule | 011, 014, 016 |
| 6 | `.claude/skills/moai/workflows/run/context-loading.md` | "Multi-Mode Router" (L35-67) | Additive note: `--mode` axis is the dispatch axis, distinct from the Phase 0.95 catalog; pointer to crosswalk section. Resolver pseudocode + sentinels untouched | 011 |
| 7 | `.claude/skills/moai/workflows/run/mode-orchestration.md` | "Team Mode Routing" (L35-52); "Completion Criteria" (L67-80) | Team/solo prose → catalog reference (Mode 3 gate); completion criteria gain chaining step (contract-conditional) | 010, 030, 031 |
| 8 | `.claude/skills/moai/workflows/loop.md` | after "Invocation Routes" (L36-46) | Relationship paragraph: Ralph engine = diagnostic fix-loop; pipeline-level completion loop may invoke it at verify step; granularity distinction; alias contract untouched | 029 |
| 9 | `.claude/skills/moai/workflows/sync.md` | Quick Reference / Invocation Flow (L61-88) | Chain-entry note: sync may be entered via auto-chain (full-pipeline contract) or explicitly; FAIL gates halt the chain | 030, 032 |
| 10 | `.moai/config/sections/workflow.yaml` (+ template counterpart, independent edit) | `workflow:` block | Add documented `agentic_loop: { max_iterations: 10 }` key (prose-read) | 023 |

Estimated live-file count: 10; with mirrors ≈ 19-20 writes. Markdown/YAML only.

---

## §D. Technical Approach

1. **SSOT-first ordering** — the rule crosswalk (M1) lands before any skill edit so every skill pointer has a valid target from the first commit.
2. **Pointer discipline** — skill surfaces never restate catalog content; they name the mode, the rule section, and the local operational consequence (what to spawn, what to log). Keeps SKILL.md lean (risk B8).
3. **Additive gates** — moai.md gate declarations reference the existing sub-skill implementations (plan.md Phase 2.3, run.md Kickoff section, sync.md auditor flow); no gate logic is duplicated, only made visible in the pipeline body.
4. **Loop as orchestrator prose** — the agentic loop is specified as orchestrator behavior (conditions, ceilings, escalations), not as a new engine. `/goal` carries the condition when available; otherwise identical text evaluated per-turn by the orchestrator (REQ-MAL-003).
5. **Neutral authorship** — all shipped prose written template-neutral from the first draft (never "strip later").

## §E. Self-Verification (run-phase deliverables)

Per verification-claim-integrity 5-section format. The run-phase agent MUST report:

- E1: AC PASS/FAIL matrix (per acceptance.md §B) with verbatim command outputs
- E2: Mirror parity — `diff -rq --exclude=.moai .claude/skills/moai internal/template/templates/.claude/skills/moai` verbatim (empty expected; aligned with AC-MAL-021 — the `--exclude=.moai` is the LOAD-BEARING mechanism here because the untracked junk dir is actively re-created by a live statusline process, so M1's deletion is best-effort hygiene only)
- E3: Neutrality grep — `grep -rn "SPEC-MOAI-AGENTIC-LOOP\|REQ-MAL-" .claude/skills/ .claude/rules/ internal/template/templates/` verbatim (0 matches expected; `.moai/specs/` excluded)
- E4: Sentinel preservation grep (all PRESERVE §A.2 tokens)
- E5: Go-diff gate — `git diff --stat <baseline_sha>..HEAD -- '*.go'` empty, where `<baseline_sha>` is the pre-SPEC HEAD pinned in progress.md §E.2 at M1 pre-flight (AC-MAL-025 dereferences the same pin)
- E6: `moai spec lint` (or `go run ./cmd/moai spec lint`) clean for this SPEC
- E7: Commit SHAs + push state; blocker reports if any

## §F. Milestones (priority-ordered, no time estimates)

| M | Scope (file cluster) | Files | Exit criterion |
|---|----------------------|-------|----------------|
| M1 | Rules SSOT layer: crosswalk section + threshold SSOT sentence; pre-flight: (a) mirror-divergence characterization, (b) delete untracked junk dir `internal/template/templates/.claude/skills/moai/workflows/.moai/` (statusline leak, not git-tracked), (c) pin pre-SPEC baseline SHA as `baseline_sha: <sha>` in progress.md §E.2 | map #1 (+ mirror) | Crosswalk present in both copies; divergence characterized (resolved or blocker-reported); junk dir removed (best-effort — a live statusline process may re-create it; `--exclude=.moai` in parity checks is the load-bearing mechanism); baseline_sha pinned; AC-MAL-005/016b |
| M2 | Router layer: SKILL.md requirement-analysis step + catalog wiring + threshold pointer | map #2 (+ mirror) | AC-MAL-001..004; SKILL.md TOTAL file delta ≤ ~35 lines (the new step block itself ≤ 25 lines per risk B8) |
| M3 | Pipeline body: moai.md gates + agentic loop lifecycle + chaining + threshold pointers | map #3 (+ mirror) | AC-MAL-010..015 (moai.md group) |
| M4 | Run tree: run.md label fix + Mode 4/6 entries; phase-execution crosswalk + drift fix; context-loading note; mode-orchestration catalog ref | map #4-7 (+ mirrors) | AC-MAL-006..009, 017; sentinels intact |
| M5 | Adjacent surfaces: loop.md relationship + sync.md chain-entry + workflow.yaml key | map #8-10 (+ mirrors/counterpart) | AC-MAL-012, 018, 019 |
| M6 | Verification batch: full mirror re-sync check, neutrality grep, sentinel grep, Go-diff gate, spec lint, AC matrix | (verification only) | E1-E6 all PASS; acceptance.md DoD |

Commit convention: one Conventional Commit per milestone (`feat(SPEC-MOAI-AGENTIC-LOOP-001): M{N} <subject>`), path-limited `git add` (risk B6), `--no-verify` prohibited. First run-phase commit carries the `draft → in-progress` frontmatter transition (manager-develop owned).

## §G. Anti-Patterns (do not do)

- Adding `workflow` as a `--mode` axis value (violates §G two-axis frozen separation)
- Restating threshold numbers in any skill surface (REQ-MAL-013)
- Embedding this SPEC's ID in any shipped skill/rule/template file (REQ-MAL-041)
- Rewriting the `run.md` `ac_converge` section instead of referencing it
- Auto-chaining on single-phase contracts (REQ-MAL-031)
- Removing/reformatting evolvable-block markers or gate IDs
- `git add -A` on this shared checkout

## §H. Cross-References

- spec.md §B (requirements), §D (decisions), §E (out of scope / follow-ups)
- acceptance.md §B (AC matrix — verification commands for every milestone exit criterion)
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Tier L delegation template applies (Section A-E full form)
