---
id: SPEC-AGENT-TIER-001
title: "Three-tier subagent model-effort configuration (max/medium/low) with operator-fixed audit pins"
version: "0.1.0"
status: draft
created: 2026-10-01
updated: 2026-10-01
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/config, internal/cli, internal/web, internal/template/templates/.moai/config/sections"
lifecycle: spec-anchored
tags: "agent-tier, model-effort, audit-pins, terminal-bench, llm-constants, web-console"
tier: M
related_specs: [SPEC-AGENT-MODEL-INHERIT-001, SPEC-MODEL-MATRIX-UPDATE-001, SPEC-V3R6-AUDIT-MODEL-PIN-001]
---

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-10-01 | manager-spec | Initial draft for card t1391 (operator directive 2026-10-01, grounded in the Terminal-Bench 4.0 cost/accuracy design report). Audit-pin location re-measured in this tree: the Go default audit block lives at `internal/config/defaults.go:1218-1238` (the delegation brief's `:1202` was stale); the claude pin literals sit at `:1221-1222`. One deviation from the lane's suggested constant naming is recorded at REQ-TIER-002. |

## §A Context

The operator directed (2026-10-01) a 3-tier subagent model·effort configuration grounded in the
Terminal-Bench 4.0 accuracy-vs-cost chart (Sonnet 5.5 / Opus 5.5 / Sonnet 5 / GPT-5.6 Sol across
the effort axis). The grounding report (`.moai/reports/agent-tier-design-20261001.md`, untracked,
primary checkout) supplies the chart figures, the 3-tier design table, the rejected-candidate
rationale, and the risk table; the load-bearing figures are carried in §E below so this SPEC is
self-sufficient.

Two operator decisions bind this SPEC and override the grounding report where they differ:

1. **Audit-surface pins are operator-fixed defaults, NOT tier-derived.**
   claude audit = {claude-opus-5-5, **high**}; codex audit = {gpt-6.1-sol, **high**};
   glm audit = {glm-5.3, **max**}. These replace the t1368 pins
   (SPEC-MODEL-MATRIX-UPDATE-001: claude {claude-opus-5-5, medium}, codex {gpt-6.1-sol, high},
   glm empty). The moai web UI already displays these values; web and CLI defaults must stay
   consistent.
2. **The 3-tier design applies to the GENERAL subagent profile matrix; audit surfaces are
   explicitly OUT of scope for the tiers.** max = {sonnet-5-5, max} (70.6% on the chart) —
   super-advisor, manager-spec design judgments; medium = {sonnet-5-5, high} (45%) —
   manager-develop, manager-docs, e2e-tester, general implementation lanes; low =
   {sonnet-5-5, medium} (29%) — Explore, light fixes, observation/summary tasks. The max-tier
   fallback {claude-opus-5-5, xhigh} (65% @ ~$5, availability/dispersion alternative) is recorded,
   not automated. (The grounding report's own tier table placed plan-auditor and sync-auditor in
   the max tier; the operator decision supersedes that placement — the auditors ride the audit
   pins instead.)

### A.1 Relationship to SPEC-AGENT-MODEL-INHERIT-001 (completed, spec-anchored)

SPEC-AGENT-MODEL-INHERIT-001 retired the 39-cell per-agent profile matrix: no per-agent
`model:`/`effort:` frontmatter (REQ-AMI-003/004), no doctrine instructing spawn-time
model/effort passing (REQ-AMI-007), no per-agent resolver or `moai model profile` command
(REQ-AMI-008), and `ModelEffort` survives only for the audit pins. This SPEC **partially
supersedes** that SPEC's universality: it reinstates a class-level tier axis resolved by
configuration. The supersession is deliberately bounded (REQ-TIER-010): agent files still carry
no `model:`/`effort:` frontmatter, doctrine still forbids hand-passing model/effort at spawn,
the tier table is binary-resolved configuration rather than a per-agent matrix, and the
declaration-only spawn observer (`internal/hook/agent_model_guard.go`) stays non-enforcing. No
other REQ-AMI requirement is touched; the back-pointer edit on the old SPEC's frontmatter is
out of scope (§C).

### A.2 Terminology decision (Q2)

Tier names (`max`/`medium`/`low`) are **configuration-key names**, not effort values. The
effort field keeps the existing chart-axis vocabulary (low | medium | high | xhigh | max)
everywhere it already exists; the GLM audit pin keeps the z.ai reasoning-state vocabulary
{low, high, max} verbatim (REQ-AMP-006). Rationale: reusing tier tokens as effort values would
collide tier-max with effort-max — the max tier's effort IS max while the medium tier's effort
is high — so any sentence saying "medium effort" would be ambiguous between tier-medium and
effort-medium. Note the settings-layer precedent is consistent with keeping the vocabularies
distinct: `effortLevel` accepts low/medium/high/xhigh only, and a resolved `max` rides the
launcher launch-argument path instead.

## §B Requirements (GEARS)

### B.1 Terminology

- **REQ-TIER-001** (Ubiquitous): The tier tokens `max`, `medium`, and `low` shall be configuration-key names that identify {model, effort} pairs — max → {sonnet-5-5, max}, medium → {sonnet-5-5, high}, low → {sonnet-5-5, medium} — and shall not be effort values: the effort field shall keep the existing chart-axis vocabulary (low | medium | high | xhigh | max) on every surface where it exists today, and the GLM audit pin shall keep the z.ai vocabulary {low, high, max} verbatim; no configuration surface shall accept a tier token in a slot whose contract is an effort value.

### B.2 Tier constants

- **REQ-TIER-002** (Ubiquitous): `internal/config` shall declare, in the same single-source block as the `DefaultGLM*` slot-constant family (`defaults.go`), one named declaration per agent tier keyed by the tier token — `DefaultClaudeTierMax`, `DefaultClaudeTierMedium`, `DefaultClaudeTierLow` — carrying exactly {sonnet-5-5, max}, {sonnet-5-5, high}, and {sonnet-5-5, medium} respectively, and no other file shall restate these model ids or effort values as inline literals.
  - *Naming deviation (recorded per the delegation brief's Q2 clause):* the brief's suggested shape `DefaultClaudeTierMax/High/Medium` keys the suffix by EFFORT slot (the `DefaultGLM{High,Medium,Low}` convention, where the suffix names the effort slot). Keying by effort slot would alias tier-medium to a constant named `...TierHigh`, recreating exactly the tier/effort ambiguity Q2 exists to prevent. The constants are therefore keyed by TIER token, which keeps constant names 1:1 with configuration keys; alignment with the `DefaultGLM*` convention is preserved at the family level (same block, same `Default<Family><Slot>` shape, closed-set derivation). The VALUES are exactly the operator's pairs — no value deviation.
- **REQ-TIER-003** (Ubiquitous): The max-tier fallback pair {claude-opus-5-5, xhigh} shall be recorded as a named single-source declaration beside the tier constants (suggested `DefaultClaudeTierMaxFallback`) whose doc comment carries the chart grounding (65% @ ~$5 — availability/dispersion alternative per the grounding report's rejected-candidate record); this SPEC requires the record only, not automatic failover.

### B.3 Audit pins (operator-fixed)

- **REQ-TIER-004** (Ubiquitous): The `workflow.audit` pin defaults shall be operator-fixed at claude = {claude-opus-5-5, high}, codex = {gpt-6.1-sol, high}, glm = {glm-5.3, max}, carried consistently by the Go default (`NewDefaultWorkflowConfig`), the distributed template (`workflow.yaml`), and the audit resolver terminal fallbacks (`resolveClaudeAuditModelEffort`, `resolveCodexAuditModelEffort`, the GLM audit pin path); the pins shall not be derived from the tier table, and a user-set pin shall keep precedence over any default (REQ-AMI-019 lineage).
- **REQ-TIER-005** (Event-driven): **When** an audit pin default changes, the web console closed sets and schema rendering (the `ValidCodexAuditModels` family, the GLM model closed set, and the audit-pin fields the console renders) shall be updated in the same change set, so no surface offers or renders a superseded default.
- **REQ-TIER-006** (Ubiquitous): The GLM audit pin shall ship non-empty ({glm-5.3, max}) replacing the empty default; the effort value shall be sent verbatim under the REQ-AMP-006 z.ai vocabulary semantics (`max` is a valid z.ai reasoning state); and the GLM task-delegation default (`glm_task` path) shall remain unchanged — the audit pin is audit-only and never task delegation (REQ-AMP-008).

### B.4 Profile matrix (tier axis)

- **REQ-TIER-007** (Ubiquitous): The subagent profile model shall carry a tier axis with exactly the closed set {max, medium, low}, each value resolving to its REQ-TIER-002 pair; audit surfaces — plan-auditor, sync-auditor, and the claude/codex/glm audit backends — shall be excluded from the tier matrix and shall resolve exclusively through the `workflow.audit` pins.
- **REQ-TIER-008** (Ubiquitous): The default agent-class→tier assignment shall be: super-advisor and manager-spec → max; manager-develop, manager-docs, e2e-tester, and general implementation lanes → medium; Explore and light-fix/observation/summary tasks → low; and the assignment shall be project-overridable through configuration.
- **REQ-TIER-009** (Event-driven): **When** the configuration loader reads a tier assignment whose token is outside the closed set, it shall reject the load with an error naming the offending token and its file (the config `Validate()` convention).
- **REQ-TIER-010** (Ubiquitous): Tier→{model, effort} resolution shall be machine-driven configuration resolution: agent definition files (C1/C2/C3) shall continue to carry no `model:`/`effort:` frontmatter, doctrine shall continue not to instruct agents to hand-pass model or effort at spawn time, and the declaration-only spawn observer (`agent_model_guard`) shall remain non-enforcing; the partial supersession of SPEC-AGENT-MODEL-INHERIT-001 is bounded to this configuration-driven tier axis and touches none of its other requirements.

### B.5 moai web widget

- **REQ-TIER-011** (Ubiquitous): The moai web console shall present the three-tier model with each tier's chart grounding — max: 70.6% · ~$11; medium: 45% · ~$2.3; low: 29% · ~$0.8 — and shall offer tier selection limited to the closed set.

### B.6 Verification discipline

- **REQ-TIER-012** (Ubiquitous): Every check this SPEC adopts shall follow two-cell adoption discipline (verification-completeness.md §2) — a RED-now observation on the pre-implementation tree (command, verbatim output, exit code, tree SHA) and a green-path cell naming the milestone that flips it; the tier-constants unit test shall be RED at run start because the constants do not exist (compile failure), and GREEN after M1.
- **REQ-TIER-013** (Ubiquitous): The audit pin values and the tier-table values shall appear as inline literals in exactly their single-source declaration locations (the Go constant/default declarations in `internal/config`, plus the template `workflow.yaml` mirror), a grep sweep shall verify no third location restates them, and the sweep shall have been demonstrated on a known-violation input — with the swept count reported — before its green is read (observed-failure completion).
- **REQ-TIER-014** (Ubiquitous): Template-surface changes shall follow Template-First (template source first, `make build`, tracked mirror byte-identical) and template neutrality (no new SPEC IDs, card IDs, or internal dates introduced under `internal/template/templates/**`; 16-programming-language neutrality preserved).

## §C Out of Scope

### Out of Scope — Automatic tier failover machinery
- No runtime availability detection, health probing, or automatic failover to {claude-opus-5-5, xhigh}; the fallback is a recorded, operator-invokable default (REQ-TIER-003). Activation semantics are revisited only when availability signals exist (§D.5 of acceptance.md).

### Out of Scope — Per-agent frontmatter tier keys and spawn-argument doctrine
- No `tier:` (or any model/effort) key is added to agent definition files by this SPEC, and no doctrine file is rewritten to instruct hand-passing values at spawn; the spawn-time injection channel is decided at run phase within REQ-TIER-010's constraints (plan.md §F M2 names the candidates and the default).

### Out of Scope — GLM task path and main-session effort policy
- The `glm_task` default model, the GLM effort overlay (`CollapseClaudeEffortToGLM`), and the main-session model policy (`model_policy` / `effort_level`, REQ-AMI-017) are untouched.

### Out of Scope — Benchmark re-adjudication
- Terminal-Bench 4.0 is the sole evidence base; FrontierCode/CursorBench tab reversal triggers a re-adjudication card, not an amendment here. Rejected candidates (Sonnet 5, GPT-5.6 Sol, Opus 5.5 low/medium) stay rejected per the grounding report's record.

### Out of Scope — SPEC-AGENT-MODEL-INHERIT-001 frontmatter housekeeping
- Adding `partially_superseded_by: [SPEC-AGENT-TIER-001]` to the old SPEC's frontmatter is a separate housekeeping delegation; this SPEC declares the supersession on its own side (§A.1, REQ-TIER-010).

### Out of Scope — Audit gate topology
- The per-auditor gates (claude required, codex required, glm advisory) and the convergence engine are unchanged; only the {model, effort} pin values move.

### Out of Scope — Pre-existing template comment residues
- The template `workflow.yaml` audit comment already carries an internal date ("operator directive 2026-09-30"). This SPEC's edits must not ADD neutrality violations; cleaning that pre-existing residue is not tasked here.

## §D Success Criteria

Acceptance criteria live in `acceptance.md` (§D AC Matrix): 15 criteria, of which 13 are
release-blocking. The phase gate is plan-auditor PASS at the Tier M threshold (0.80). The
binding operator criteria: the tier-constants test RED at run start / GREEN after M1; audit pin
values byte-exact against the operator table in the Go default and the template mirror; no
third location hardcoding the pin values (grep sweep with positive control); the web widget
rendering the three tiers with the chart figures.

## §E Evidence Base (figures carried from the grounding report)

The grounding report is untracked (`.moai/reports/` is gitignored); the SPEC-critical figures
are carried here so the SPEC survives independently.

| Model | effort | score (%) | cost/attempt ($) |
|---|---|---|---|
| Sonnet 5.5 | medium | ~29 | ~0.8 |
| Sonnet 5.5 | high | ~45 | ~2.3 |
| Sonnet 5.5 | xhigh | ~62 | ~5.5 |
| Sonnet 5.5 | max | ~70.6 | ~11 |
| Opus 5.5 | xhigh | ~65 | ~5.0 |

Tier design: max = Sonnet 5.5 @ max (highest score 70.6% — accuracy-first); medium = Sonnet 5.5
@ high (callout ② — 45% @ ~$2.3, cost-efficiency sweet spot); low = Sonnet 5.5 @ medium (29% @
~$0.8 — beats Sonnet 5's best and dominates Opus low). Fallback: Opus 5.5 @ xhigh (65% @ ~$5,
−5.6p vs max tier at ~55% of the cost). Rejected: Sonnet 5 (≤11% everywhere), GPT-5.6 Sol
(38% @ ~$8 — dominated), Opus 5.5 low/medium (dominated at equal or higher cost).

## §F Cross-References

- Grounding report: `.moai/reports/agent-tier-design-20261001.md` (primary checkout, untracked)
- Lineage: SPEC-AGENT-MODEL-INHERIT-001 (partial supersession, §A.1) · SPEC-MODEL-MATRIX-UPDATE-001 (t1368 pins being replaced) · SPEC-V3R6-AUDIT-MODEL-PIN-001 (pin precedence semantics, REQ-AMP-001..008)
- Code surfaces measured 2026-10-01 at `f130aa041bb90c81b235029afffa806a65e88480`: `internal/config/defaults.go` (DefaultGLM* family :248-276; Audit block :1218-1238), `internal/config/audit_models.go` (ModelEffort, AuditConfig), `internal/config/closed_sets.go` (DefaultCodexAuditModel :96, ValidCodexAuditModels), `internal/cli/mcp_claude.go:181` / `mcp_codex.go:213` (resolvers), `internal/cli/mcp_glm.go:54` (glmAuditDefaultModel), `internal/cli/launch_effort_settings.go` (resolveLaunchEffort), `internal/hook/agent_model_guard.go` (declaration-only observer), `internal/web/` (templ console), template mirror `internal/template/templates/.moai/config/sections/workflow.yaml:98-121`
