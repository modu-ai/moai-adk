---
id: SPEC-V3R6-AUDIT-MODEL-PIN-001
title: "Pin cross-model audit backend model+effort via the workflow.audit config block"
version: 1.1.0
status: implemented
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: high
phase: "v3.1.4 target"
module: internal/cli
lifecycle: spec-anchored
tags: "mcp, cross-model-audit, workflow-config, glm, codex"
tier: M
related_specs: [SPEC-MOAI-MCP-SERVER-001, SPEC-AUDIT-MULTI-MODEL-001]
---

# SPEC-V3R6-AUDIT-MODEL-PIN-001 — Audit backend model+effort pinning

> Revision 1.1.0 (plan-audit iter 1 findings MF1-MF6 applied). MF1 design
> relocation: the pin lives in the EXISTING `workflow.audit` block
> (`AuditConfig`), not a new `llm.audit` section — `.moai/config/sections/llm.yaml`
> is gitignored (`.gitignore:192`) and wiped by `moai update`, so an llm.yaml
> pin is uncommitable and non-durable (lead ruling C).

## §A Problem / Context

The cross-model audit backends (`codex_audit`, `glm_audit`, and the `audit_multi`
convergence fan-out) do not control which model and effort they actually run on:

1. **Codex** — `codexSSOTModelEffort` (`internal/cli/mcp_codex.go:191`) resolves the
   SSOT sync-auditor cell via `template.ResolveAgentModelEffort(llm, "sync-auditor")`,
   then filters the result through `codexServableModel` (prefix set at
   `internal/cli/mcp_codex.go:158`: `gpt-`, `o1`, `o3`, `o4`, `codex`). The SSOT cell
   resolves to `opus`/`high` (a Claude id) → prefix mismatch → the pair is DROPPED →
   the request carries no model/effort. This machine only runs `gpt-5.6-sol`/`high`
   by accident, via `~/.codex/config.toml` local defaults. The pin is
   machine-dependent and invisible to the web console.
2. **GLM** — `resolveGLMAuditModel()` (`internal/cli/mcp_glm.go:127`) resolves a model
   id ONLY. Effort is never delivered: the `glmMessagesRequest` body
   (`internal/cli/mcp_glm.go:102`) has no reasoning/effort field at all, and no
   environment variable applies because `glm_audit` performs its own direct HTTP POST
   to the z.ai Anthropic-compatible endpoint.

Consequence: audit quality is governed by unmanaged local machine state, and the GLM
audit's reasoning depth cannot be raised to max even when the project wants it.

## §B Goal

Make the audit backends' model+effort an explicit, project-visible, COMMITTABLE
moai-side instruction, with zero deployment impact for projects that do not opt in:

- Extend the existing `workflow.audit` block (`AuditConfig`,
  `internal/config/audit_models.go:59`; attached at
  `internal/config/types.go:486` `Audit AuditConfig \`yaml:"audit"\``) with two
  `{model, effort}` pairs: `audit.codex` and `audit.glm`, reusing
  `config.ModelEffort`.
- The audit resolvers read this pin with 3-tier precedence: `workflow.audit` pin >
  existing SSOT sync-auditor cell > empty/absent fallback (legacy behavior).
- Pinned values live ONLY in this project's local TRACKED
  `.moai/config/sections/workflow.yaml` (codex: `gpt-5.6-sol`/`high`; glm:
  `glm-5.3`/`max`); the template workflow.yaml gains the same sub-keys with EMPTY
  defaults. `gpt-5.6-sol` never appears in any template-managed surface.
- The GLM effort actually reaches z.ai — proven by live observed evidence with a
  numeric decision rule, not by config-read assertions (resolves the UNVERIFIED
  flag AC-MTP-032b).

## §C Scope

**In scope**: `AuditConfig` extension in `internal/config/audit_models.go`; a
workflow-section load helper beside the existing llm one; the codex and GLM audit
resolvers and the GLM audit request body; the web console typed edit path for the
new fields (the existing Audit panel); the template's empty-default sub-keys; this
project's local pin values in workflow.yaml.

**Design shape (single audit block)**: the pin EXTENDS `AuditConfig` — the
struct behind the `workflow.audit` block. No `audit:` block exists in either
the local or the template workflow.yaml today (grep-verified: 0 matches in
both); M1 CREATES the template block with empty pin sub-keys and M5 CREATES the
local block with the pinned values. There is no second audit structure
anywhere; the llm-side section proposed in revision 1.0.0 is dead (gitignore +
update-wipe, plan-audit MF1) and is not partially retained.

## §D Requirements (GEARS)

- **REQ-AMP-001** (Ubiquitous) — The workflow config's existing `audit` block shall
  carry `codex` and `glm` sub-keys, each a `{model, effort}` pair reusing
  `config.ModelEffort`, loaded by the existing workflow.yaml load path, with the
  Go defaults and the distributed template carrying empty values only.

- **REQ-AMP-002** (capability gate) — **Where** `audit.codex.model` is non-empty and
  codex-servable, the codex AUDIT resolution shall return the `audit.codex`
  `{model, effort}` pair in place of the SSOT sync-auditor cell, so that both
  transmission seams (`thread/start` model at `openCodexSessionOn`; `turn/start`
  model+effort at `buildCodexReviewParams`) carry the pinned pair on the
  `codex_audit` path.

- **REQ-AMP-003** (capability gate) — **Where** `audit.glm.model` is non-empty, the
  GLM audit resolver shall return the `audit.glm` `{model, effort}` pair in place
  of the SSOT-derived model id, bypassing the non-GLM-session fallback
  (`glmAuditDefaultModel`) for this explicitly pinned value.

- **REQ-AMP-004** (state-driven) — **While** the pin sub-keys are absent, empty, or
  the pinned model fails backend-servability validation, the audit resolvers shall
  resolve exactly as before this SPEC (SSOT cell → filter/fallback), preserving
  backward compatibility for every project that never opts in.

- **REQ-AMP-005** (unwanted) — The distributed template
  (`internal/template/templates/.moai/config/sections/workflow.yaml`) and the
  Go-side config defaults shall NOT contain any non-empty audit model or effort
  value; the literal `gpt-5.6-sol` shall never appear under
  `internal/template/templates/`.

- **REQ-AMP-006** (Ubiquitous) — The `audit.glm.effort` input set shall be exactly
  the z.ai reasoning-state names `{low, high, max}`, stored and transmitted
  verbatim with a single interpretation (the GLM state — even where a name such as
  `high` or `max` also exists in the Claude vocabulary, no Claude-vocabulary
  reading applies); any other non-empty value is invalid, and an invalid value
  shall cause the reasoning directive to be omitted while the model pin still
  applies. No effort collapse runs on this path — `CollapseClaudeEffortToGLM`
  (`internal/template/glm_effort_overlay.go:129`) is the vocabulary REFERENCE, not
  a runtime dependency here.

- **REQ-AMP-007** (event-driven) — **When** the reasoning directive is delivered
  with effort `max`, a live `glm_audit` call shall produce observed backend-side
  evidence distinguishing delivery from non-delivery by a numeric decision rule
  (per AC-AMP-006).

- **REQ-AMP-008** (state-driven) — **While** the shared resolution bodies serve
  NON-audit consumers — `glm_task` through `resolveGLMModelForAgent`
  (`internal/cli/glm_task.go:125`) and `codex_task` through
  `resolveCodexModelEffort` reachable from `openCodexSessionOn`
  (`internal/cli/codex_task.go:226` → `internal/cli/mcp_codex.go:565`) — the pins
  shall affect ONLY the audit entry points (`codex_audit`, `glm_audit`, and the
  `audit_multi` convergence fan-out), leaving both task paths' resolutions
  unchanged.

- **REQ-AMP-009** (Ubiquitous) — The web console typed edit path shall expose
  `audit.codex.{model,effort}` and `audit.glm.{model,effort}` as editable fields
  in the existing Audit panel (the surface already rendering `workflow.audit.*`
  fields), with i18n labels in all four console locales, so the pin is visible
  and correctable without hand-editing yaml.

## §E Constraints

- **Template-First**: the template sub-keys land in
  `internal/template/templates/.moai/config/sections/workflow.yaml` first;
  `make build` regenerates the embed + catalog hashes before the local mirror.
- **Config symmetry CI guard**: the schema extension must be covered by a guard
  that FAILS on yaml↔struct drift — the existing
  `audit_struct_yaml_symmetry_test.go` symmetryCases do NOT cover `AuditConfig`
  (verified: 7 cases — the 4 MIG-003 sections plus StatuslineConfig lineage), so
  the extension adds an `AuditConfig` symmetry case (or an equivalent new
  round-trip test) per plan.md M1 (plan-audit MF3).
- **Effort vocabularies are backend-native and single-reading**: codex effort
  uses the codex/Claude vocabulary (`low|medium|high|xhigh|max` — pin value
  `high`); glm effort uses the z.ai state vocabulary under REQ-AMP-006's
  single-reading rule (pin value `max`; the card's "reasoning-max" normalizes to
  the stored state name `max`).
- **Fail-open unchanged**: an unusable pinned value degrades through the existing
  fail-open paths (codex: fall back to SSOT resolution; GLM: z.ai HTTP error →
  `VerdictInconclusive`), never a hard error.
- **Optional-backend absence**: when a live gate's backend credential/binary is
  absent (`GLM_API_KEY` missing; codex binary/auth missing), the live AC reports
  SKIP with an explicit marker in the evidence file — never FAIL, never a silent
  pass — and a SKIP blocks the AC from being counted PASS (plan-audit MF6).
- **§14 hardcoding prevention**: new identifiers/constants land as named constants
  beside their consumers, not inline literals.

## §F Out of Scope

### Out of Scope — companion statusline patch
- The already-committed working-tree changes to `.claude/skills/moai-kanban-foreman/SKILL.md`,
  `internal/statusline/renderer.go`, `internal/statusline/renderer_test.go`, and
  `internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md` ride the
  same PR but are NOT part of this SPEC; this SPEC neither touches nor reverts them.

### Out of Scope — non-audit consumers
- `codex_task` and `glm_task` model resolution (REQ-AMP-008 names them and their
  seams), agent-spawn model injection (`ResolveAgentModelEffort` callers outside
  the audit backends), harness-generation effort, and the session-global
  `ANTHROPIC_REASONING_EFFORT` overlay are untouched.

### Out of Scope — llm.yaml
- `.moai/config/sections/llm.yaml` (gitignored, update-wiped) receives NO new keys
  from this SPEC; the llm section schema and `loadLLMSectionOnly` are untouched.

### Out of Scope — codex local config management
- `~/.codex/config.toml` is machine state; this SPEC does not read, write, or
  validate it. The moai-side pin makes it irrelevant, not managed.

## §G HISTORY

- 2026-08-24 — v1.0.0 draft authored (plan phase, kanban card t225): llm.audit
  section design.
- 2026-08-24 — v1.1.0 plan-audit iter 1 revision (score 0.875 FAIL): MF1 schema
  relocated llm.audit → workflow.audit `AuditConfig` (gitignore evidence
  `.gitignore:192`; lead ruling C); MF2 codex_task isolation requirement added
  (shared-resolver leak path `codex_task.go:226` → `mcp_codex.go:565`); MF3
  symmetry-guard coverage made real (AuditConfig case; prior citation was
  vacuous — 7 existing cases exclude it); MF4 single-reading effort vocabulary
  rule (REQ-AMP-006 rewritten); MF5 numeric decision rule for the live AC;
  MF6 SKIP semantics for absent credentials.

## §H Cross-References

- `SPEC-MOAI-MCP-SERVER-001` — codex backend + `codexSSOTModelEffort` provenance
  (REQ-CX2-002, C4/C7 no-regression guards this SPEC must not break).
- `SPEC-AUDIT-MULTI-MODEL-001` — `audit_multi` convergence; its GLM leg flows
  through the same `callGLMAudit` seam changed here.
- `internal/config/audit_models.go` — the extended `AuditConfig` and its
  workflow.yaml load contract.
- `internal/template/glm_effort_overlay.go:129` — z.ai reasoning-state vocabulary
  reference (REQ-AMP-006).
- Template llm.yaml GLM overlay note — CORRECTED by the live differential
  (acceptance.md AC-AMP-006 amendment record, 2026-08-24): the overlay note's
  prior claim (Anthropic `thinking` honored, top-level `reasoning_effort`
  ignored) is REVERSED on the audit path — the thinking-budget object measured
  a true null (1.02) and the top-level `reasoning_effort` field is the
  effective delivery field. The template llm.yaml overlay-doc correction itself
  remains a separate follow-up card (reference-only).
