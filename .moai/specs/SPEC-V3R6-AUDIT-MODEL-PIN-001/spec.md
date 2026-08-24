---
id: SPEC-V3R6-AUDIT-MODEL-PIN-001
title: "Pin cross-model audit backend model+effort via llm.audit config section"
version: 1.0.0
status: draft
created: 2026-08-24
updated: 2026-08-24
author: manager-spec
priority: high
phase: "v3.1.4 target"
module: internal/cli
lifecycle: spec-anchored
tags: "mcp, cross-model-audit, llm-config, glm, codex"
tier: M
related_specs: [SPEC-MOAI-MCP-SERVER-001, SPEC-AUDIT-MULTI-MODEL-001]
---

# SPEC-V3R6-AUDIT-MODEL-PIN-001 — Audit backend model+effort pinning

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

Make the audit backends' model+effort an explicit, project-visible moai-side
instruction, with zero deployment impact for projects that do not opt in:

- New `llm.audit` config section: `audit.codex{model, effort}` and
  `audit.glm{model, effort}`.
- The audit resolvers read this section with PRECEDENCE over the SSOT sync-auditor
  cell; empty/absent section → existing fallback path byte-identical.
- Pinned values live ONLY in this project's local
  `.moai/config/sections/llm.yaml` (codex: `gpt-5.6-sol`/`high`; glm:
  `glm-5.3`/`max`), never in the distributed template.
- The GLM effort actually reaches z.ai — proven by live observed evidence, not by
  config-read assertions (resolves the UNVERIFIED flag AC-MTP-032b).

## §C Scope

**In scope**: `internal/config` LLM section schema; the codex and GLM audit
resolvers and the GLM audit request body; the web console typed edit path for the
new fields; the template's empty-default `audit:` block; this project's local
pin values.

**Naming disambiguation**: this section is `llm.audit` — a sub-key of the `llm:`
block in `llm.yaml`. It is DISTINCT from the pre-existing `workflow.audit` block
(`internal/config/types.go:486`, `AuditConfig` — the audit_model token + per-auditor
gates in `workflow.yaml`). The two never merge.

## §D Requirements (GEARS)

- **REQ-AMP-001** (Ubiquitous) — The llm config section shall carry an `audit`
  sub-structure with `codex` and `glm` keys, each a `{model, effort}` pair reusing
  `config.ModelEffort`, loaded by the existing llm.yaml load path
  (`loadLLMSectionOnly` / `Loader.Load` llm section) with no new section file.

- **REQ-AMP-002** (capability gate) — **Where** `audit.codex.model` is non-empty and
  codex-servable, the codex audit resolver shall return the `audit.codex`
  `{model, effort}` pair in place of the SSOT sync-auditor cell, so that both
  transmission seams (`thread/start` model at `openCodexSession`; `turn/start`
  model+effort at `buildCodexReviewParams`) carry the pinned pair.

- **REQ-AMP-003** (capability gate) — **Where** `audit.glm.model` is non-empty, the
  GLM audit resolver shall return the `audit.glm` `{model, effort}` pair in place of
  the SSOT-derived model id, bypassing the non-GLM-session fallback
  (`glmAuditDefaultModel`) for this explicitly pinned value.

- **REQ-AMP-004** (state-driven) — **While** the `audit` section is absent, empty,
  or its model field fails backend-servability validation, the audit resolvers shall
  resolve exactly as before this SPEC (SSOT cell → filter/fallback), preserving
  backward compatibility for every project that never opts in.

- **REQ-AMP-005** (unwanted) — The distributed template
  (`internal/template/templates/.moai/config/sections/llm.yaml`) and the Go-side
  config defaults shall NOT contain any non-empty audit model or effort value; the
  literal `gpt-5.6-sol` shall never appear under `internal/template/templates/`.

- **REQ-AMP-006** (event-driven) — **When** `audit.glm.effort` is configured, the
  GLM audit HTTP request body shall carry the reasoning directive on the field the
  live z.ai Anthropic-compatible endpoint honors, and shall not carry it on a field
  the live evidence shows is ignored.

- **REQ-AMP-007** (event-driven) — **When** the reasoning directive is delivered
  with effort `max`, a live `glm_audit` call shall produce observed backend-side
  evidence distinguishing delivery from non-delivery (per AC-AMP-006).

- **REQ-AMP-008** (state-driven) — **While** the shared GLM model-resolution body
  (`resolveGLMModelForAgent`) serves the `glm_task` path, the `audit.glm` pin shall
  affect only the audit entry points (`glm_audit` handler and the `audit_multi`
  convergence fan-out), leaving `glm_task` resolution unchanged.

- **REQ-AMP-009** (Ubiquitous) — The web console typed edit path (3rd Party LLM
  panel) shall expose `audit.codex.{model,effort}` and `audit.glm.{model,effort}`
  as editable fields with i18n labels in all four console locales, so the pin is
  visible and correctable without hand-editing yaml.

## §E Constraints

- **Template-First**: the template `audit:` block lands in
  `internal/template/templates/` first; `make build` regenerates the embed +
  catalog hashes before the local mirror.
- **Config symmetry CI guard**: `CONFIG_STRUCT_YAML_MISMATCH`
  (`audit_struct_yaml_symmetry_test.go`) requires the new struct fields and the
  template yaml keys to stay symmetric.
- **Effort vocabularies are backend-native**: codex effort uses the codex/Claude
  vocabulary (`low|medium|high|xhigh|max` — pin value `high`); glm effort uses the
  z.ai 3-level reasoning-state vocabulary (`low|high|max` — pin value `max`,
  canonical names from `template.GLMReasoningStateNames()`; the card's
  "reasoning-max" normalizes to stored value `max`).
- **Fail-open unchanged**: an unusable pinned value degrades through the existing
  fail-open paths (codex: fall back to SSOT resolution; GLM: z.ai HTTP error →
  `VerdictInconclusive`), never a hard error.
- **§14 hardcoding prevention**: new identifiers/constants land as named constants
  beside their consumers, not inline literals.

## §F Out of Scope

### Out of Scope — companion statusline patch
- The already-applied working-tree changes to `.claude/skills/moai-kanban-foreman/SKILL.md`,
  `internal/statusline/renderer.go`, `internal/statusline/renderer_test.go`, and
  `internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md` ride the
  same PR but are NOT part of this SPEC; this SPEC neither touches nor reverts them.

### Out of Scope — workflow.audit block
- The `workflow.audit` block (audit_model token, per-auditor gates,
  `AuditConfig`) is not modified, merged with `llm.audit`, or re-keyed.

### Out of Scope — non-audit consumers
- Agent-spawn model injection (`ResolveAgentModelEffort` callers outside the audit
  backends), `glm_task` model resolution, harness-generation effort, and the
  session-global `ANTHROPIC_REASONING_EFFORT` overlay are untouched.

### Out of Scope — codex local config management
- `~/.codex/config.toml` is machine state; this SPEC does not read, write, or
  validate it. The moai-side pin makes it irrelevant, not managed.

## §G HISTORY

- 2026-08-24 — v1.0.0 draft authored (plan phase, kanban card t225). Baseline
  behavior measured in this tree: codex pair dropped at
  `mcp_codex.go:200-202` (opus → prefix mismatch), GLM effort absent from
  `glmMessagesRequest` (`mcp_glm.go:102-107`).

## §H Cross-References

- `SPEC-MOAI-MCP-SERVER-001` — codex backend + `codexSSOTModelEffort` provenance
  (REQ-CX2-002, C4/C7 no-regression guards this SPEC must not break).
- `SPEC-AUDIT-MULTI-MODEL-001` — `audit_multi` convergence; its GLM leg flows
  through the same `callGLMAudit` seam changed here.
- `internal/template/glm_effort_overlay.go` — z.ai reasoning-state vocabulary
  (`CollapseClaudeEffortToGLM`, `GLMReasoningStateNames`) reused for validation.
- Template llm.yaml GLM overlay note — prior measurement that the z.ai
  Anthropic-compat shim honors the Anthropic `thinking` parameter and ignores a
  top-level `reasoning_effort` field; the live AC re-verifies this for the audit
  path specifically.
