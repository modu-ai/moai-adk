---
id: SPEC-GLM-FLASH-DEFAULT-001
title: "GLM-5.3-Flash as the default coding model"
version: "0.2.0"
status: in-progress
created: 2026-08-27
updated: 2026-08-27
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/config, internal/template, internal/cli, internal/statusline, internal/web"
lifecycle: spec-anchored
tags: "glm, flash, default-model, effort-overlay, web-console"
tier: M
---

# SPEC-GLM-FLASH-DEFAULT-001 — GLM-5.3-Flash as the default coding model

## HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-08-27 | manager-spec | Plan-phase authoring (card t289, Class C, Tier M) from operator directive 2026-08-27 + mid-dispatch additions |
| 2026-08-27 | manager-spec | Plan-audit iter-1 revision (v0.2.0): defects D1-D7 applied — AC-013 added for REQ-005; boot-smoke recipe pinned to env-injection map; substring-matching wording corrected to registration-time guidance; REQ-002 twin-scope disambiguated; off-schema `related_specs:` removed; REQ-001 ordering wording softened |

**Origin**: Kanban card t289. GLM-5.3-Flash official release (docs.z.ai/guides/vlm/glm-5.3-flash); operator directive to switch the DEFAULT coding model to it, plus two binding mid-dispatch additions (flash max-only reasoning; web console surface).

## §1 Context and Problem

GLM-5.3-Flash is officially released. The project currently pins every GLM tier slot to `glm-5.3` (Go defaults `internal/config/defaults.go` `DefaultGLMHigh/Medium/Low/Fable` + legacy `DefaultGLMHaiku/Sonnet/Opus`; template twin `internal/template/templates/.moai/config/sections/llm.yaml` `llm.glm.models.*`). The offered closed set is `ValidGLMModels()` = {glm-5.3, glm-5.1, glm-4.7, glm-4.5-air}. The operator directs that the default coding model become `glm-5.3-flash`.

Official specs carried from the card (documentation of record):

- Model id: `glm-5.3-flash`
- Context: 1M tokens
- Architecture: 320B total / 18B active (sparse + linear hybrid attention; 4.44× KV-cache reduction)
- Multimodal: text, image, video, file
- Z.ai Code Bench v1.0: 29.0 (Opus 4.8 = 29.5)
- Recommended params: temperature 1, top_p 0.95, reasoning_effort max, stream + tool_stream true, thinking.type enabled only

**Problem**: flash is not registered anywhere in the tree, the defaults still point at glm-5.3, and the existing effort-collapse overlay maps Claude effort `low` → z.ai `reasoning_effort: low` — a wire state that DOES NOT EXIST on flash (flash accepts `reasoning_effort: max` only). Shipping flash as default without a model-aware overlay branch would emit an invalid effort value on every low-effort request.

## §2 Terminology

- **Tier slots** — `llm.glm.models.{high,medium,low,fable}` (plus legacy opus/sonnet/haiku aliases): the four model stand-ins the launcher maps onto `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL` env (`internal/cli/glm.go:354-357`).
- **Effort overlay** — the runtime SSOT at `internal/template/glm_effort_overlay.go` collapsing Claude effort {low, medium, high, xhigh, max} onto z.ai reasoning states (low | max, post SPEC-GLM-EFFORT-MAX-001).
- **Profile matrix** — `llm.yaml` `profiles` + `harness_agents` per-agent effort columns (t205 structure). A SEPARATE axis from the GLM tier slots; untouched by this SPEC.

## §3 Requirements (GEARS)

### REQ-001 — Registration (closed set)

The GLM model registry shall offer `glm-5.3-flash` as a member of `ValidGLMModels()` (`internal/config/closed_sets.go`), listed first in the set (default-first placement; no capability ordering between flash and glm-5.3 is claimed).

### REQ-002 — Default switch (tier slots)

The GLM tier-slot defaults shall resolve to `glm-5.3-flash` for all four slots (high, medium, low, fable) in the template twin (`internal/template/templates/.moai/config/sections/llm.yaml` `llm.glm.models.*` — the four slot keys are the template's only surface) and for the three Go legacy-alias constants (`DefaultGLMHaiku` / `DefaultGLMSonnet` / `DefaultGLMOpus` in `internal/config/defaults.go`) in the Go twin — the legacy aliases live only in the Go constants; the template carries no alias keys.

### REQ-003 — glm-5.3 stays selectable

Where a user explicitly selects `glm-5.3` in any tier slot (existing llm.yaml or the settings widget), the system shall preserve that selection and its existing behavior — glm-5.3 remains a member of the offered closed set, backed by a named Go constant (the set is derived from constants; a bare retarget of `DefaultGLMHigh` would silently drop glm-5.3 from the set).

### REQ-004 — Flash reasoning max-only (operator-binding addition 1)

**When** the resolved GLM model is `glm-5.3-flash`, the effort overlay (`internal/template/glm_effort_overlay.go`, runtime SSOT) shall resolve every Claude effort level — including `low` — to `reasoning_effort: max` with thinking enabled; the flash resolution path shall not emit the `low` reasoning state, because the 3-level GLM reasoning control (low/high/max) does not exist on flash. The non-flash collapse behavior (glm-5.3-family: low→low, above-low→max) shall remain unchanged.

### REQ-005 — Documentation twin of the overlay rule

The template `llm.yaml` effort-mapping comment block (documentation-only twin of the overlay) shall state the flash max-only rule alongside the existing collapse table.

### REQ-006 — Context window table

The statusline context-window table (`internal/statusline/memory.go` `glmContextWindows`) shall carry an explicit `"glm-5.3-flash": 1_000_000` entry, so the 1M window (and the 50% handoff threshold it drives via `MOAI_STATUSLINE_CONTEXT_SIZE` / `CLAUDE_CODE_AUTO_COMPACT_WINDOW`) holds by direct entry rather than by longest-substring fallback through `"glm-5.3"`. Registration-time guidance: an unregistered `glm-5.3-*` variant inherits 1M via longest-substring matching; any future divergent `glm-5.3-*` id MUST add its own explicit table entry at registration time.

### REQ-007 — Web console surface (operator-binding addition 2)

Where the web console settings surface lists or selects GLM models — the tier-slot selects and the audit GLM-model select, whose options are derived from `ValidGLMModels()` in `internal/settings/schema_sections.go` — the console shall offer `glm-5.3-flash`, with per-locale option labels in `internal/web/assets/i18n.js` for all four locales (en, ko, ja, zh) under both option-key families (`f.llm.glm.models.opt.*` and `f.workflow.audit.glm.model.opt.*`).

### REQ-008 — Template mirror discipline

The template twin edits (`llm.yaml` model literals + overlay comment) shall land in `internal/template/templates/` first, followed by `make build` (embedded-template regeneration), with catalog.yaml unchanged and template content free of SPEC-IDs, internal dates, and commit SHAs (template-neutrality, CLAUDE.local.md §2 / §25).

### REQ-009 — Docs surfaces

Where README (4-locale set) and docs-site pages name the GLM default model or enumerate GLM models, the documentation shall name `glm-5.3-flash` as the default — scoped to the files that actually carry GLM model lists (grep-verified; full list in plan.md §D).

### REQ-010 — Boot smoke

**When** `moai glm` launches with default configuration, the launcher shall inject `glm-5.3-flash` into the `ANTHROPIC_DEFAULT_{OPUS,SONNET,HAIKU,FABLE}_MODEL` env set (process env + tmux session env), observable in the injection path (`internal/cli/glm.go` buildTmuxInjectVars / env Setenv block).

## §4 Constraints

- **Language**: code, comments, and template content in English (language.yaml).
- **Preserved invariants**: (a) explicit glm-5.3 selection keeps its current behavior; (b) the judgment-weighted profile matrix (profiles / harness_agents) is NOT modified; (c) the overlay remains TOTAL over unrecognized effort input (default → max).
- **Verified finding — no sampling-params surface**: repo-wide grep for `top_p` in `internal/**.go` returns zero hits, and no `temperature` construction exists in the GLM launcher path. The launcher injects env vars only; the Anthropic-compatible request body is constructed by Claude Code itself. Therefore the card's recommended sampling params (temperature 1, top_p 0.95, stream/tool_stream) have NO landing surface in this codebase. Per the dispatch instruction, this SPEC does not invent one; the absence is recorded here as a finding. `reasoning_effort` is the only recommended parameter with an existing surface (the overlay/thinking path, t175-measured).
- **Local checkout note**: `.moai/config/sections/llm.yaml` does not exist in this tree (verified) — the template is the only llm.yaml twin to edit; no local-override merge is needed.

## §5 Findings (verified, tree 410da655f = origin/main tip)

| ID | Finding | Evidence |
|----|---------|----------|
| F-1 | No request-params construction surface exists | `grep -rn "top_p" internal --include='*.go'` → 0 hits; no temperature in glm.go/launcher.go |
| F-2 | flash already resolves 1M context via longest-substring match on `"glm-5.3"` | `internal/statusline/memory.go` `matchContextWindow` priority-2 path |
| F-3 | `ValidGLMModels()` is derived from DefaultGLM* constants, so a tier-slot retarget alone would drop glm-5.3 from the offered set | `internal/config/closed_sets.go:76-78` |
| F-4 | Web options are schema-derived; only labels are hand-authored | `internal/settings/schema_sections.go:229,399` + `internal/web/glm_tier_test.go` asserts the exact set |
| F-5 | Resolution path: tier slots → `ANTHROPIC_DEFAULT_*_MODEL` env (process + tmux), context window derived from the High slot | `internal/cli/glm.go:354-366,497-514` |

## §6 Out of Scope

### Out of Scope — glm-5.2 restoration

- No change to the withdrawn status of glm-5.2 in the offered set (`DefaultGLM52` stays a loadable-but-unoffered constant).

### Out of Scope — GLM pricing and cost surfaces

- No pricing tables, cost estimates, or billing surfaces for flash.

### Out of Scope — fable-tier semantics

- The fable tier's semantic behavior (window class, handoff thresholds) is unchanged beyond its default model id switching with the other slots.

### Out of Scope — reasoning-state widget redesign

- The per-agent reasoning-state display domain (`GLMReasoningStateNames` low/high/max, agentfm surfaces) is not redesigned per-model. Stored per-agent states record intent under the existing documented semantics; only the resolution output for flash is pinned to max (REQ-004).

### Out of Scope — new request-params surface

- No new mechanism for injecting temperature / top_p / tool_stream is built (see §4 finding). If z.ai later exposes these through the Anthropic-compat shim in a way this repo can carry, that is a separate SPEC.

### Out of Scope — profile matrix

- The t205 judgment-weighted profile matrix (llm.yaml `profiles` + `harness_agents`) is untouched; effort columns collapse onto the new model through the overlay unchanged.
