---
id: SPEC-ADVISOR-RUNG-001
title: "Executor-Advisor Escalation Rung for /moai fix and /moai loop + GLM Judgment Carve-Out"
version: "0.1.0"
status: superseded
superseded_by: SPEC-AGENT-ARCH-V2-001
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/skills/moai/workflows"
lifecycle: spec-anchored
era: V3R6
tier: S
depends_on: [SPEC-MODEL-ROUTING-WIRE-001]
tags: "advisor-rung, escalation, moai-fix, moai-loop, glm-carve-out, cg-leader-review, per-spawn-model, workflow-reflex"
---

# SPEC-ADVISOR-RUNG-001 — Executor-Advisor Escalation Rung + GLM Judgment Carve-Out

## Epic Context

**Epic**: Workflow-Reflex (6-SPEC epic derived from the 3-lens workflow audit: model-tier routing / Loop Engineering / Harness Engineering). This SPEC is **4 of 6**.

- **Dependency notes**: **Depends on SPEC-MODEL-ROUTING-WIRE-001 (2 of 6)** — the per-spawn model/effort runtime-arg doctrine and the `moai route` mechanical value surface must land first; the advisor spawn rides that exact channel. Independent of SPEC 1 (HARNESS-RATCHET-REWIRE) and SPEC 3 (LOOP-VERDICT-CONTRACT), though SPEC 3 shares the loop.md edit surface (see Constraints #4).
- **Tier**: S (minimal envelope) — see plan.md §A.4 for evidence.
- **era**: V3R6 (modern 3-phase close: plan→run→sync).

## Traceability (audit findings provenance)

| Finding ID | Severity | Summary |
|------------|----------|---------|
| R3 | MED | No Executor-Advisor shape exists on the /moai surface: every stuck-state escalation in /moai fix and /moai loop goes straight to the USER via AskUserQuestion; no intermediate "consult a stronger model before interrupting the user" rung, despite the per-spawn `Agent(general-purpose, model: opus)` primitive and `[1m]`-safe per-spawn model args already existing unwired |
| R6 | MED | Cost-asymmetry inverse: under `moai glm` (all-GLM mode) judgment gates (plan-auditor/sync-auditor) and planning run on GLM, while CLAUDE.md §15 tells CG mode to avoid exactly that class of work — the all-GLM mode has NO such carve-out |

---

## User Story

**As a** user whose `/moai fix` or `/moai loop` run is stuck repeating the same failed fix,
**I want** the workflow to consult a read-only strong-model advisor after repeated failures on the same diagnostic — re-seeding the executor with the advisor's diagnosis — and to be interrupted only when the advisor rung also fails, plus an honest reduced-assurance flag on judgment gates when everything runs on GLM,
**so that** cheap-executor/expensive-advisor cost asymmetry is exploited instead of inverted, my attention is the LAST escalation rung rather than the first, and all-GLM sessions stop silently running Opus-class judgment work on a worker-class model.

---

## Problem — Measurable Gap Definition (vci §2 attribution)

All gaps measured 2026-07-09 by this agent via Bash/Read. Line numbers indicative; content anchors are authoritative.

### GAP-1 — Every stuck-state escalation goes straight to the user (R3)

- **Measured source**: `.claude/skills/moai/workflows/fix.md` — Level 3 approval ("Level 3 (Review): User approval required", observed near line 146; "Level 3 fixes require AskUserQuestion approval", near line 184) and the CI-loop path (Related Skills § moai-workflow-ci-loop, observed lines 312-314: "패치 실패 시 AskUserQuestion 경유 escalation" / "semantic 분류 또는 patch 실패 시 user escalation"); `.claude/rules/moai/workflow/ci-autofix-protocol.md` § iteration-limit ("iteration 4+ → MANDATORY BLOCKING AskUserQuestion", CONST-V3R5-006 no-auto-resume); `.claude/skills/moai/workflows/loop.md` ("Level 3 (Approval): AskUserQuestion required", observed near line 133).
- **Observed pattern**: Three distinct stuck/escalation paths, all terminating directly at the user. Meanwhile the primitives that could host an intermediate advisor already exist and are verified unwired-for-this-purpose: `archived-agent-rejection.md` §C rows 7-12 document the per-spawn `Agent(subagent_type: "general-purpose", model: "opus", tools: <whitelist>, prompt: <domain instructions>)` pattern, and `model-policy.md` (observed near line 96) confirms per-spawn model args are `[1m]`-safe ("the per-spawn `model` parameter is a runtime arg, distinct from the frontmatter field that triggers the bug"). Nothing triggers them conditionally on repeated failure.

### GAP-2 — All-GLM mode carries no judgment carve-out (R6)

- **Measured source**: `.claude/skills/moai/team/glm.md` § LLM Mode Detection table (observed near line 104): row `| glm | GLM-only | GLM | GLM |` — leader AND teammates on GLM, so plan-auditor/sync-auditor verdicts and planning all run on GLM; `CLAUDE.md` §15 CG-mode guidance: "**Avoid**: planning/architecture (needs Opus reasoning), security reviews, complex debugging"; `grep -n "Avoid\|avoid" .claude/skills/moai/team/glm.md` → 0 matches.
- **Observed pattern**: CG mode (Claude leader + GLM teammates) carries the Avoid list precisely because judgment work needs the stronger model; the all-GLM mode — where the ENTIRE session including judgment gates runs on the worker-class model — carries no equivalent carve-out or reduced-assurance flag. The cost asymmetry is inverted: the cheapest configuration silently owns the highest-judgment work.

### Aggregate defect claim

**The escalation ladder is missing its middle rung, and the cheapest mode silently owns the most judgment-sensitive work.** This SPEC inserts a read-only strong-model advisor rung between repeated executor failure and user escalation (doc-layer, riding the SPEC-MODEL-ROUTING-WIRE-001 per-spawn channel), documents the GLM judgment carve-out mirroring CLAUDE.md §15, and names the existing CG leader-review behavior as the on-demand advisor for teammate-stuck cases.

---

## Requirements (GEARS notation)

> **Subject convention**: generalized subjects ("the fix workflow", "the loop workflow", "the advisor spawn", "the GLM-mode doc"). No legacy `IF/THEN` modality.

### REQ-ADV-001 — Event-driven (When) — advisor rung trigger and re-seed

**When** `/moai loop` (per-iteration cycle) or `/moai fix` (repeat-failure path: Level-3-class fix re-failure or CI-loop patch failure) accumulates N consecutive failed iterations on the SAME diagnostic (default N=2; same-diagnostic identity per plan.md §D D3), the workflow SHALL instruct the orchestrator to spawn a READ-ONLY strong-model advisor with the accumulated failure evidence (diagnostic, attempted patches, verbatim failure output), and SHALL re-seed the executor's next attempt with the advisor's diagnosis before any user escalation.

### REQ-ADV-002 — Ubiquitous — advisor spawn contract

The advisor spawn SHALL be a per-spawn `Agent(general-purpose)` with (a) a read-only tool whitelist (no Write/Edit), (b) foreground execution (`run_in_background: false` semantics not required since read-only, but result must return synchronously to the escalation decision), and (c) per-spawn model/effort runtime arguments resolved per the SPEC-MODEL-ROUTING-WIRE-001 doctrine (`[1m]`-safe runtime-arg channel; strong-model default per plan.md §D D4) — never a frontmatter model pin.

### REQ-ADV-003 — Unwanted behavior — escalation-contract preservation

The advisor rung SHALL NOT fire on a first failure, SHALL NOT replace or delay any mandatory user-escalation contract — the ci-autofix 3-iteration ceiling and its iteration-4+ MANDATORY BLOCKING AskUserQuestion (CONST-V3R5-006), fix.md Level 3/4 approval requirements, and loop safety escalations all remain intact — and SHALL NOT permit the advisor's diagnosis to authorize semantic-failure auto-patching (CONST-V3R5-010 preserved). The advisor consult happens WITHIN the existing iteration budget; user escalation at the ceiling remains unconditional.

### REQ-ADV-004 — Ubiquitous — GLM judgment carve-out

The GLM-mode doc (`.claude/skills/moai/team/glm.md`) SHALL carry a judgment carve-out for all-GLM mode (`team_mode: glm`) stating that judgment gates (plan-auditor / sync-auditor verdicts) and plan-phase architecture reasoning are reduced-assurance under all-GLM and SHOULD either run in a Claude session or be explicitly flagged as reduced-assurance in their outputs — mirroring the CLAUDE.md §15 Avoid list (planning/architecture, security reviews, complex debugging) — and `.claude/skills/moai/team/run.md` § CG Mode SHALL cross-reference the carve-out.

### REQ-ADV-005 — Ubiquitous — CG leader-review named as on-demand advisor

The CG-mode docs SHALL name the existing leader-review behavior as the on-demand advisor rung for GLM teammates' mid-task-stuck case — teammate blocker report → Claude-leader advisory diagnosis → re-delegation with the diagnosis injected — as an EXTENSION of the existing blocker-report boundary (agent-common-protocol.md § Blocker Report Format), not a replacement of it.

### REQ-ADV-006 — Capability gate (Where) — template-first boundary

**Where** an edited surface has a template mirror under `internal/template/templates/` (verified present 2026-07-09 for: fix.md, loop.md, ci-autofix-protocol.md, team/glm.md, team/run.md), the run-phase SHALL apply edits template-first (edit template source, `make build`) or identically in both trees.

---

## Constraints

1. **Frozen escalation clauses untouched (HARD)** — `ci-autofix-protocol.md`'s [ZONE:Frozen] 3-iteration ceiling and mandatory blocking AskUserQuestion are PRESERVED verbatim; the advisor note lands in an Evolvable subsection (plan.md §D D2) and never modifies frozen text.
2. **Subagent boundary (HARD)** — the advisor is a subagent: it returns a diagnosis to the orchestrator and NEVER prompts the user (askuser-protocol.md § Orchestrator–Subagent Boundary). User escalation remains orchestrator-owned AskUserQuestion territory.
3. **Doc-layer SPEC** — skill/rule edits + template mirrors only; no Go changes expected. The advisor trigger is a prompt-layer instruction, not a mechanical hook.
4. **Sibling edit-surface coordination** — SPEC-LOOP-VERDICT-CONTRACT-001 (3 of 6, run-phase pending) edits loop.md Steps 1/4/9; this SPEC's loop.md advisor rung MUST be sequenced after (or explicitly coordinated with) that SPEC's run-phase to avoid conflicting hunks (plan.md §B B10).
5. **Cost-safety** — the advisor fires only when stuck (N≥2 same-diagnostic failures); it is never a per-iteration overhead.
6. **GEARS notation; era V3R6; 12 canonical frontmatter fields.**

---

## Out of Scope

> Per the `OutOfScopeRule` lint, this section uses `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.

### Out of Scope — AskUserQuestion boundary changes

- The orchestrator-subagent asymmetric boundary, the AskUserQuestion channel monopoly, and blocker-report semantics are all unchanged. The advisor rung slots BENEATH the existing user-escalation contracts; it does not modify them.

### Out of Scope — first-failure auto-advisor

- Auto-invoking an advisor on every failure (or every iteration) — the rung is conditional on N≥2 consecutive same-diagnostic failures; anything more aggressive inverts the cost-safety rationale.

### Out of Scope — team.enabled default

- `workflow.team.enabled` stays `false`; no Agent Teams activation changes. The CG leader-review naming (REQ-ADV-005) documents existing behavior.

### Out of Scope — mechanical trigger implementation

- A Go/hook layer that mechanically counts failures and spawns the advisor. This SPEC's wiring is prompt-layer instruction (same deliberate layer choice as SPEC-MODEL-ROUTING-WIRE-001's DD2 precedent); runtime enforcement is follow-up territory.

### Out of Scope — GLM model tables and llm.yaml

- GLM model resolution, `llm.yaml`, `moai glm`/`moai cg` Go launch code, and the GLM context-window doctrine are untouched. Only the glm.md/run.md documentation surface changes.

---

## Cross-References

- **EXTEND base (doc)**: `.claude/skills/moai/workflows/fix.md` (Level 3 approval near line 146/184; CI-loop escalation lines 312-314); `.claude/skills/moai/workflows/loop.md` (Level 3 approval near line 133; iteration cycle); `.claude/rules/moai/workflow/ci-autofix-protocol.md` (Evolvable subsection only); `.claude/skills/moai/team/glm.md` (§ LLM Mode Detection, near line 104); `.claude/skills/moai/team/run.md` (§ CG Mode, near line 87). All five have verified template mirrors.
- **Primitives cited (unmodified)**: `archived-agent-rejection.md` §C rows 7-12 (per-spawn `Agent(general-purpose, model: opus)` pattern); `model-policy.md` per-spawn runtime-arg `[1m]`-safety (near line 96); `agent-common-protocol.md` § Blocker Report Format.
- **Prerequisite**: SPEC-MODEL-ROUTING-WIRE-001 (2 of 6) — `moai route` value surface + pre-spawn per-spawn-arg instruction.
- **Mirror source**: CLAUDE.md §15 CG Avoid list (mirrored into glm.md by REQ-ADV-004; CLAUDE.md itself unmodified).
- **Epic**: Workflow-Reflex 4 of 6. Siblings: SPEC-HARNESS-RATCHET-REWIRE-001 (1), SPEC-MODEL-ROUTING-WIRE-001 (2), SPEC-LOOP-VERDICT-CONTRACT-001 (3), SPEC-CADENCE-BRIDGE-001 (5), SPEC-OBSERVE-HYGIENE-001 (6).

---

## History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-09 | manager-spec | Initial draft — plan-phase artifacts (spec + plan + acceptance + progress). Workflow-Reflex Epic 4 of 6. Advisor rung for fix/loop stuck states + GLM judgment carve-out + CG leader-review naming. Tier S. |
