---
id: SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001
title: "Workflow orchestration architecture fix: 17 skip phases + autonomous mode selection"
version: "0.1.1"
status: draft
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P0
phase: "v3.0.0"
module: ".claude/skills/moai/workflows + .claude/rules/moai/* + .claude/agents/*"
lifecycle: spec-anchored
tags: "workflow, orchestration, gears, delegation, hierarchical, mode-selection, sprint-10, cohort-7"
depends_on: [SPEC-V3R6-GEARS-MIGRATION-001, SPEC-V3R6-SKILL-GEARS-ALIGN-001, SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001]
tier: L
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-25 | manager-spec | Initial Tier L spec authored from research.md synthesis (3 parallel research agents) — covers 6 findings + 12 recommendations R1-R12 + user-flagged Finding #6 (5-mode autonomous selection) |
| 0.1.1 | 2026-05-25 | manager-spec | iter-2 focused fix per plan-auditor iter-1 PASS-WITH-DEBT 0.8625. D1 RESOLVED (acceptance.md §B REQ-WOF-013 trace fixed `spec.md §G R9` → `research.md §D.3 R9`). D3 RESOLVED (plan.md §C.1 Tier 5 → Tier 6 rename, spec.md §C.1 Tier 5 unchanged as canonical). D6 RESOLVED (NEW AC-WOF-018 multi-spawn parallel preference verifying REQ-WOF-013 Compound; AC total 17→18; §B Traceability Matrix updated). spec.md body unchanged in this iter; frontmatter version + HISTORY only. Predicted iter-2 plan-auditor: ~0.90 skip-eligible. |

---

## §A — Goals

Restore the MoAI orchestrator's behavior to match the documented Plan-Run-Sync workflow architecture by:
(a) reactivating the 17 phases observed silently skipped across the recent 4-SPEC cohort (sync-phase 3 quality specialists + run-phase manager-strategy chain + 5 HUMAN GATE decision points + plan-phase Explore/research.md/Issue/BODP audit + sub-skill on-demand loading);
(b) implementing the missing autonomous execution-mode selection logic (5 modes: sequential / parallel / background / sub-agent / agent-team) per user-flagged Finding #6;
(c) aligning the orchestrator's delegation pattern with Anthropic-official hierarchical chain semantics + revfactory/harness Hybrid-orchestrator template + Karpathy 4 Coding Principles.

The fix is constitutional in scope (affects every `/moai plan`, `/moai run`, `/moai sync` invocation) and is therefore classified Tier L per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier.

---

## §B — Background

### §B.1 Observed gap

Across the recent 4-SPEC cohort (`SPEC-V3R6-SKILL-GEARS-ALIGN-001`, `SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001`, `SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001`, `SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001`) the MoAI orchestrator executed approximately 70-93% of phases silently — 17 of 25 documented workflow phases were not invoked. The orchestrator collapsed multi-agent hierarchical delegation chains into flat single-spawn dispatches, never invoked the sync-phase quality specialists (`manager-quality`, `expert-security`, `manager-develop` coverage), never spawned the run-phase `manager-strategy` analysis-only role, never engaged `evaluator-active` for Producer-Reviewer cycles under the `thorough` harness level, never invoked the Skill router (Read-loaded SKILL.md body content directly instead of calling `Skill()`), and consistently bypassed the 5 documented HUMAN GATE decision points.

### §B.2 Root cause analysis (cite research.md §A Verdict)

The orchestrator's current behavior violates the spirit of every documented orchestration framework it inherits:

- **Anthropic official 4-phase pattern** (Explore → Plan → Implement → Commit) — the Explore subagent is never spawned; HUMAN GATE Step 2 is autonomously skipped; the chain pattern is collapsed.
- **revfactory/harness mandatory Phase 0 Plan Confirmation** — autonomous-flow contract violates "users are harness architects, not passive observers" verbatim.
- **Karpathy 4 Coding Principles** — 6 of 8 derivative anti-patterns exhibited (Over-Engineering: 17-agent catalog unused; Drive-By Refactoring: subagent prompts bundle M1+M2+M3+M4; Silent Assumption: autonomous flow assumes user wants no gates; Guessing Over Clarifying: paste-ready resume skips clarify rounds; Sycophantic Agreement: never pushes back on `/moai run`; Claiming Without Evidence: SPECs closed without HUMAN GATE verification).
- **Context7-confirmed `Task` tool delegation contract** — 8 orchestration patterns enumerated, only Pattern #1 (Sequential) used.

No source explicitly endorses the current "flat single-manager dispatch" pattern. See `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/research.md` §A Executive Summary for full corroboration matrix.

### §B.3 User-flagged Finding #6 (autonomous mode selection)

The user separately flagged that the orchestrator should autonomously judge between 5 execution modes (sequential / parallel / background / sub-agent / agent-team) based on observable task complexity signals — but currently defaults to sequential single-spawn always. The Anthropic multi-agent research blog quantifies the cost of this default: parallel 3-5 subagents cut research time up to 90% for multi-domain investigations. See research.md §C Decision Tree for the synthesized selection logic.

### §B.4 Impact on downstream artifacts

The skip pattern manifests as observable downstream defects:
- Sprint 10 cohort 4/8 plan-phase iter-3 final score 0.870 PASS-WITH-DEBT (Consistency 0.72-0.74 stagnated through count-drift L63 pattern) — direct consequence of single-spawn manager-spec carrying full Tier M lifecycle without manager-strategy chain.
- Sync-phase 4-artifact `sync_commit_sha:` partial-backfill defects (L60 reinforced 3 times across sibling SPECs) — direct consequence of `manager-docs` operating without `manager-quality` audit feedback loop.
- Multi-session race incidents (L52 case 5 confirmed 9+ times) — direct consequence of orchestrator skipping pre-spawn `gh pr checks --json` + `git fetch --dry-run` parallel verification batch.

---

## §C — Scope

### §C.1 In-Scope file inventory

**Tier 1 — Workflow router skills (3 files)**:
- `.claude/skills/moai/workflows/plan.md` — restore Explore phase + research.md emission + HUMAN GATE Decision Point 1 + GitHub Issue auto-creation reference + BODP audit reference
- `.claude/skills/moai/workflows/run.md` — restore Phase 0.5 Plan Audit Gate (already documented, verify), Phase 0.95 Mode Selection (NEW), Phase 1 manager-strategy chain restoration, Phase 2.0 Sprint Contract for thorough harness, Phase 2.5 GitHub Issue creation, HUMAN GATE Plan Approval
- `.claude/skills/moai/workflows/sync.md` — restore Phase 0.5.4 manager-quality invocation, Phase 0.55 expert-security manifest audit, Phase 0.7 manager-develop coverage, HUMAN GATE 1 (working tree + tests) and HUMAN GATE 2 (doc scope), 4-phase close with Mx delegation

**Tier 2 — Sub-skill modules (7 files)**:
- `.claude/skills/moai/workflows/plan/context-discovery.md` — codify Explore parallel-spawn pattern
- `.claude/skills/moai/workflows/plan/clarity-interview.md` — verify Socratic interview AskUserQuestion contract
- `.claude/skills/moai/workflows/plan/spec-assembly.md` — verify Tier judgment AskUserQuestion + Decision Point 1
- `.claude/skills/moai/workflows/run/context-loading.md` — add Mode Selection Phase 0.95 invocation
- `.claude/skills/moai/workflows/run/phase-execution.md` — Phase 0.5 → 0.95 → 1 sequence + HUMAN GATE Plan Approval
- `.claude/skills/moai/workflows/run/task-decomposition.md` — restore manager-strategy → manager-develop hierarchical chain
- `.claude/skills/moai/workflows/run/mode-orchestration.md` — codify 5-mode decision tree

**Tier 3 — Rule files (5 files)**:
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — update §1 to reflect mode-selection awareness + Phase 0.95 sequence
- `.claude/rules/moai/workflow/session-handoff.md` — clarify Block 5 `실행:` line MUST trigger Skill router (NOT manual SKILL.md body Read)
- `.claude/rules/moai/workflow/spec-workflow.md` — document Phase 0.95 Mode Selection between Phase 0.5 Plan Audit Gate and Phase 1; cross-reference 5-mode decision tree
- `.claude/rules/moai/workflow/orchestration-mode-selection.md` (NEW) — canonical 5-mode autonomous selection rule per Finding #6
- `.claude/rules/moai/workflow/human-gates.md` (NEW) — canonical inventory of 5 HUMAN GATE decision points + AskUserQuestion contract

**Tier 4 — Agent definitions (3 files)**:
- `.claude/agents/core/manager-strategy.md` — verify code-prohibited HARD assertion + analysis-only scope discipline
- `.claude/agents/core/manager-quality.md` — verify sync-phase Phase 0.5.4 invocation surface
- `.claude/agents/expert/expert-security.md` — verify sync-phase Phase 0.55 dependency manifest audit invocation surface

**Tier 5 — SPEC artifacts (4 files — this SPEC)**:
- `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/spec.md` (this file)
- `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/plan.md`
- `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/acceptance.md`
- `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/design.md`

**Estimated scope total**: 22 in-scope files (excluding the 4 SPEC artifacts which manager-spec authors in plan-phase). Mirror parity into `internal/template/templates/.claude/skills/moai/workflows/` + `.claude/rules/moai/workflow/` is REQUIRED (additional ~10 files in run-phase). Total Tier L scope: ~32 files / >1000 LOC equivalent in markdown plus state-machine documentation.

### §C.2 Out of Scope

Explicit non-goals (see §F Non-Goals for full enumeration):
- Code-level implementation of Go-side hook enforcement (R10 hook-based PostToolUse/Stop enforcement is deferred to a follow-up SPEC).
- A/B testing methodology execution (R11 deferred — skill-ab-testing.md is the SSOT, not this SPEC).
- Status Transition Ownership PostToolUse hook implementation (R12 follow-up SPEC).
- Modification of any predecessor SPEC bodies in `.moai/specs/SPEC-V3R6-*` directories (strict L48 SSOT discipline).
- Modification of the 16-language rule files under `.claude/rules/moai/languages/` (out of orchestration scope).
- Modification of the SPEC frontmatter schema SSOT (already canonicalized in `.claude/rules/moai/development/spec-frontmatter-schema.md`).

---

## §D — Requirements (GEARS Notation Self-Dogfood)

This SPEC enforces GEARS notation self-dogfood ≥80% across the 15 REQ-WOF-XXX entries below. Pattern distribution: 4 Ubiquitous + 4 Event-driven `When` + 3 State-driven `While` + 2 Capability-gate `Where` + 1 Event-detected (Unwanted) + 1 Compound. IF/THEN modality is forbidden per SPEC-V3R6-GEARS-MIGRATION-001.

### REQ-WOF-001 (Ubiquitous)

The MoAI orchestrator shall restore all 5 HUMAN GATE decision points (Decision Point 1 plan + Plan Approval run + GATE 1 sync working-tree-and-tests + GATE 2 sync doc-scope + Decision Point 2 plan-close) and invoke each via `AskUserQuestion` with the canonical Socratic interview structure (≤4 questions per round, first option marked `(권장)`, conversation_language match).

### REQ-WOF-002 (Ubiquitous)

The MoAI orchestrator shall invoke the sync-phase 3 quality specialists in their documented sequence: `manager-quality` at Phase 0.5.4 (TRUST 5 validation), `expert-security` at Phase 0.55 (dependency manifest audit, always-runs HARD), and `manager-develop` at Phase 0.7 (coverage verification).

### REQ-WOF-003 (Ubiquitous)

The MoAI orchestrator shall maintain the run-phase hierarchical delegation chain `manager-strategy → manager-develop → manager-quality + expert-security + manager-docs` as 3+ separate `Agent()` spawns with orchestrator-mediated context passing; single-spawn delegation covering multiple manager roles is forbidden.

### REQ-WOF-004 (Ubiquitous)

The MoAI orchestrator shall implement autonomous execution-mode selection logic covering all 5 documented modes (sequential / parallel / background / sub-agent / agent-team) and shall record the selected mode and selection rationale in `.moai/specs/<SPEC-ID>/progress.md` § Mode Selection.

### REQ-WOF-005 (Event-driven `When`)

When the orchestrator processes a `/moai <subcommand>` instruction from a paste-ready resume Block 5, the orchestrator shall invoke `Skill("moai", arguments: "<subcommand> $ARGUMENTS")` rather than reading the corresponding `SKILL.md` body file directly.

### REQ-WOF-006 (Event-driven `When`)

When the run-phase enters Phase 1 (post Phase 0.95 Mode Selection), the orchestrator shall spawn `manager-strategy` (analysis-only, code-prohibited HARD) before any `manager-develop` spawn, and the manager-strategy spawn shall produce a `tasks.md` artifact under `.moai/specs/<SPEC-ID>/` enumerating M1-M6+ task decomposition.

### REQ-WOF-007 (Event-driven `When`)

When the sync-phase enters Phase 0.55, the orchestrator shall invoke `expert-security` for dependency manifest audit covering all `go.sum` / `package-lock.json` / `Pipfile.lock` / `Cargo.lock` artifacts present in the SPEC-modified scope, with the invocation HARD-marked as always-runs regardless of the harness level.

### REQ-WOF-008 (Event-driven `When`)

When the run-phase manager-develop subagent returns a blocker report indicating SPEC body content modification is required mid-run (D-NEW-1 inline-fix pattern), the orchestrator shall halt the run-phase and re-delegate to `manager-spec` with the user-confirmed scope before re-delegating back to `manager-develop` to continue.

### REQ-WOF-009 (State-driven `While`)

While the task scope encompasses ≥10 files OR ≥3 distinct domains (as judged by the manager-spec scope inventory in §C of the SPEC's plan.md), the orchestrator shall route to Full Pipeline mode (parallel multi-spawn of researcher + analyst + architect for plan-phase; implementer + tester + reviewer for run-phase per workflow.yaml role profiles).

### REQ-WOF-010 (State-driven `While`)

While the SPEC is classified `harness: thorough` (per `.moai/config/sections/harness.yaml`) AND the SPEC tier is M or L, the orchestrator shall invoke `evaluator-active` for Sprint Contract negotiation at Phase 2.0 (before Phase 2 implementation) with `max_iterations: 3` per harness Producer-Reviewer pattern.

### REQ-WOF-011 (State-driven `While`)

While the orchestrator is processing any user-facing question, the orchestrator shall route the question exclusively through `AskUserQuestion` (with mandatory `ToolSearch(query: "select:AskUserQuestion")` preload immediately before each call); free-form interrogative prose in response body text is forbidden.

### REQ-WOF-012 (Capability-gate `Where`)

Where the project has `workflow.team.enabled: true` in `.moai/config/sections/workflow.yaml` AND environment variable `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` is set AND the SPEC harness level is `thorough`, the orchestrator shall consider Agent Teams mode (parallel teammate spawn via `Agent(subagent_type: "general-purpose")` with workflow.yaml role profile overrides) as a candidate mode in the Phase 0.95 Mode Selection decision tree.

### REQ-WOF-013 (Capability-gate `Where`)

Where the SPEC tier is L (per `tier:` frontmatter field) OR the run-phase delegation prompt includes explicit `--pr` flag, the orchestrator shall route post-implementation through the `manager-git` PR creation path (`feat/SPEC-XXX` branch + `gh pr create`) rather than the Hybrid Trunk main-direct push path documented in CLAUDE.local.md §23.7.

### REQ-WOF-014 (Event-detected — replaces Unwanted IF/THEN)

When the orchestrator detects an autonomous-flow attempt to skip any of the 5 HUMAN GATE decision points (e.g., a paste-ready resume containing `머지 후:` instruction that bypasses Plan Approval), the orchestrator shall halt the autonomous flow, emit a structured warning to the user via `AskUserQuestion` listing the bypassed gate and the canonical recovery path, and shall not proceed until the user explicitly confirms the gate decision.

### REQ-WOF-015 (Compound — combines `Where` + `While` + `When`)

`[Where the harness level is standard or thorough] [While the task scope is multi-domain (≥3 domains)] [When the orchestrator selects an execution mode in Phase 0.95]`, the orchestrator shall prefer parallel multi-spawn (3-5 concurrent `Agent()` calls in a single message) over sequential single-spawn per the Anthropic multi-agent research recommendation ("3-5 parallel cut research time 90%").

---

## §E — Acceptance Criteria Reference

See `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/acceptance.md` for the full mandatory acceptance criteria matrix (≥15 AC-WOF-XXX entries with Given/When/Then format, severity, evidence command, pass criterion, and REQ-WOF-XXX traceability mapping).

---

## §F — Non-Goals

The following items are explicitly out of scope for this SPEC:

1. **Hook-based mechanical enforcement** (R10 from research.md §D.4) — PostToolUse/Stop hooks for unbypassable contracts are deferred to a follow-up SPEC. Current SPEC delivers declarative orchestrator-discipline at the markdown rule layer.
2. **A/B test methodology execution** (R11) — skill-ab-testing.md is the SSOT; running comparative tests is a separate workflow.
3. **Status Transition Ownership PostToolUse hook** (R12 / REQ-ARR-009 follow-up) — declarative ownership matrix in spec-frontmatter-schema.md is sufficient for this SPEC; hook enforcement is deferred.
4. **Go-side CLI implementation changes** — this SPEC is markdown-only (workflow rules + skills + agent definitions). Any Go code change (e.g., `internal/spec/lint.go` rule additions) is out of scope.
5. **Migration of the 88 pre-v3 EARS SPECs to GEARS notation** — backward-compatibility window remains in effect through 2026-11-22; this SPEC neither blocks nor accelerates that migration.
6. **Modification of any SPEC body in `.moai/specs/SPEC-V3R6-*` predecessor directories** — strict L48 SSOT discipline; only the current SPEC's 4 artifacts may be authored.
7. **Modification of the 16 language rule files under `.claude/rules/moai/languages/`** — out of orchestration scope.
8. **`/moai feedback`, `/moai review`, `/moai e2e`, `/moai design` subcommand classification** — deferred per spec-workflow.md "Out of scope of this matrix" callout.

---

## §G — Risks

| # | Risk | Severity | Likelihood | Mitigation |
|---|------|----------|------------|------------|
| R1 | Restoring HUMAN GATE 5종 may regress autonomous-flow user experience by introducing 5+ AskUserQuestion rounds per SPEC | HIGH | HIGH | Per REQ-WOF-014 and acceptance.md AC-WOF-013, the Phase 0.5 Plan Audit Gate skip-eligibility policy (PASS ≥ 0.90, no plan-PR commit landed) carries forward: gates SHOULD be present but MAY be auto-confirmed when audit verdict is sufficiently strong. Document this skip-eligibility in plan.md §B Run-phase Strategy explicitly. |
| R2 | Hierarchical delegation chain restoration (`manager-strategy → manager-develop`) may increase Tier L SPEC wall-time by 20-40% vs flat single-spawn | MEDIUM | HIGH | Per REQ-WOF-015 compound clause, parallel multi-spawn pattern compensates for the chain depth cost. Anthropic verbatim "3-5 parallel cut 90%" benchmark cited in research.md §B.1 as quantitative target. Measure wall-time delta in evaluation phase (M6). |
| R3 | Skill router invocation discipline (REQ-WOF-005) may break existing paste-ready resume messages currently formatted with explicit `/moai <subcommand>` syntax | LOW | MEDIUM | The `/moai <subcommand>` syntax already triggers Skill() per Claude Code official behavior; the issue is orchestrator's manual SKILL.md Read bypass. Updating session-handoff.md §Block 5 to clarify "Skill router auto-triggers from /moai prefix; do NOT Read SKILL.md body manually" resolves without breaking existing resume formats. |
| R4 | Autonomous mode selection (REQ-WOF-004) introduces classification ambiguity at the boundary between modes (e.g., 9-file scope vs 10-file scope threshold) | MEDIUM | MEDIUM | The decision tree in design.md §B.3 includes tie-breaker rules: when scope falls at threshold ±1, default to the simpler mode (sub-agent over agent-team) and log the boundary case to `progress.md § Mode Selection` for retrospective analysis. The 5-mode taxonomy is deliberately discrete; future fine-tuning is OK. |
| R5 | The 22-file scope inventory may diverge from the actual run-phase modification set due to discovered downstream coupling (e.g., session-handoff.md cross-references that fan out into additional rule files) | MEDIUM | MEDIUM | Per L46 path-specific staging discipline + manager-develop-prompt-template.md §1.B B10 PRESERVE list invariant, run-phase will enforce strict scope adherence. If genuine scope expansion is required, return blocker report and re-delegate to manager-spec for explicit §C.1 expansion. Plan-auditor MP-3 Traceability check will detect divergence. |
| R6 | New rule files (`orchestration-mode-selection.md`, `human-gates.md`) introduce additional rule-loading cost (~2,500 tokens combined) at session start, marginally degrading initial token budget | LOW | LOW | Both new rules use `paths:` frontmatter where applicable. orchestration-mode-selection.md applies only when `paths: "**/.moai/specs/**"` matches; human-gates.md applies only at orchestrator turns processing `/moai plan|run|sync`. Estimated cost ≈ 800 tokens conditional load, ≈ 200 tokens always-load metadata. Net acceptable. |

---

## §H — Cross-References

### §H.1 Research synthesis

- `.moai/specs/SPEC-V3R6-WORKFLOW-ORCHESTRATION-FIX-001/research.md` — full 3-agent parallel research synthesis (Anthropic official + Context7 MCP + GitHub repos) with 6 findings + 12 recommendations R1-R12

### §H.2 Predecessor SPECs (read-only inputs per L48 SSOT discipline)

- `SPEC-V3R6-GEARS-MIGRATION-001` — GEARS notation v0.2.0 canonical migration (depends_on)
- `SPEC-V3R6-SKILL-GEARS-ALIGN-001` — Skill authoring guide GEARS alignment (depends_on)
- `SPEC-V3R6-WORKFLOW-PLAN-GEARS-ALIGN-001` — Workflow plan skill GEARS alignment (depends_on, in-progress at SHA `27afbca1e`)
- `SPEC-V3R6-FOUNDATION-CORE-GEARS-ALIGN-001` — Foundation core GEARS alignment (closed at `0156c7003`)
- `SPEC-V3R6-PLAN-AUDITOR-GEARS-ALIGN-001` — Plan-auditor GEARS alignment (closed at `ebe492670`)

### §H.3 Canonical rule SSOTs

- `.claude/rules/moai/workflow/spec-workflow.md` — Plan/Run/Sync phase canonical SSOT + SPEC Complexity Tier (S/M/L) + Phase 0.5 Plan Audit Gate
- `.claude/rules/moai/core/agent-common-protocol.md` — User Interaction Boundary (AskUserQuestion HARD) + Parallel Execution batching
- `.claude/rules/moai/core/askuser-protocol.md` — Channel Monopoly + Socratic Interview Structure + Free-form Circumvention Prohibition
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Section A-E 5-section delegation template + B1-B12 Known Issues
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — Canonical 12 frontmatter fields + Status Transition Ownership Matrix
- `.claude/rules/moai/workflow/session-handoff.md` — Paste-ready resume 6-block + Block 0 worktree anchor + Block 5 Skill router invocation
- `.claude/rules/moai/workflow/context-window-management.md` — Model-specific threshold (1M = 50% / 200K = 90%)
- `.claude/rules/moai/workflow/verification-batch-pattern.md` — 7-item Trust-but-verify parallel batch
- `.claude/rules/moai/development/branch-origin-protocol.md` — BODP audit trail (CONST-V3R5-030..036)
- `.claude/rules/moai/workflow/ci-autofix-protocol.md` — CI auto-fix loop max-3 iter + AskUserQuestion escalation
- `.claude/rules/moai/development/agent-patterns.md` — 6 harness orchestration patterns (Pipeline / Fan-out-Fan-in / Expert Pool / Producer-Reviewer / Supervisor / Hierarchical Delegation)
- `.claude/rules/moai/development/orchestrator-templates.md` — 3 templates (Team / Sub / Hybrid) decision matrix
- `.claude/skills/moai/references/anti-patterns.md` — Karpathy 8 anti-patterns catalog

### §H.4 External authoritative sources

- claude.com/docs/en/agents — Agents documentation
- claude.com/docs/en/skills — Skills + Skill() invocation
- claude.com/docs/en/best-practices — 4-phase canonical pattern (Explore → Plan → Implement → Commit)
- claude.com/docs/en/sub-agents — "Subagents cannot spawn other subagents" constraint
- anthropic.com/engineering/built-multi-agent-research-system — "3-5 parallel cut research time 90%"
- platform.claude.com/docs/en/about-claude/models/whats-new-claude-4-7 — Opus 4.7 Adaptive Thinking + Principle 4/5

### §H.5 Out-of-scope follow-up SPECs

- (future) `SPEC-V3R6-WORKFLOW-HOOK-ENFORCEMENT-001` — implements R10 hook-based PostToolUse/Stop enforcement
- (future) `SPEC-V3R6-STATUS-OWNERSHIP-HOOK-001` — implements R12 / REQ-ARR-009 follow-up Status Transition Ownership PostToolUse hook
- (future) `SPEC-V3R6-WORKFLOW-AB-TEST-001` — executes R11 A/B test methodology against the new workflow skills

---

Version: 0.1.1
Status: draft (plan-phase initial authoring + iter-2 focused fix)
Tier: L (constitutional scope, 5-artifact set including this spec.md + plan.md + acceptance.md + design.md + research.md)
