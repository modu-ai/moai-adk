---
id: SPEC-CODEX-DUAL-AGENTS-001
title: "Codex Dual Harness M5 — Dual Publication of the 11 Agent Definitions from a Neutral Source"
version: "0.1.0"
status: in-progress
created: 2026-08-22
updated: 2026-08-22
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/template"
lifecycle: spec-anchored
tags: "codex, agents, dual-harness, template, publication"
tier: M
---

# SPEC-CODEX-DUAL-AGENTS-001 — Codex Dual Harness M5: Dual Agent Publication

## §A User Story and Background

**User story.** As a MoAI template maintainer, I want the 11 retained agent definitions to have a
single neutral source from which both the Claude Code agent files (`.claude/agents/moai/*.md`) and
the Codex agent files (`.codex/agents/moai/*.toml`) are published deterministically, so that
Codex-harness users receive the same agent capabilities without a hand-maintained fork of the
agent definitions — while existing Claude Code users see **zero diff** from `moai update`.

**Background.** This SPEC is card t89 of the Codex dual-harness series (M5). The M0 measurement
card (t91, report at `.moai/reports/t91/README.md`, primary checkout) verified on codex-cli
0.147.0 that:

- A probe file `.codex/agents/t91probe.toml` with fields `name`, `description`,
  `developer_instructions`, `model_reasoning_effort`, `sandbox_mode` was loaded and a subagent
  delegation returned successfully (`T91PROBE-OK`) — the agents-TOML format works (t91 §2).
- t91 §8 M5 row: "전제 성립. `sandbox_mode`·`model_reasoning_effort` 필드 동작 확인" (premise
  holds; both fields verified working).
- `codex mcp add moai -- moai mcp-server` registers `[mcp_servers.moai]` in `config.toml`; all 21
  moai MCP tools were recognized in-session (t91 §5).
- **CAVEAT (binding)**: unknown config values and event names are **silently ignored**
  (`T91BogusEvent` produced zero warnings — t91 §1 [HARD]). An emitter cannot rely on Codex
  validating field values; the generator side must validate its own output.

**Scope statement.** This SPEC is M5 ONLY: neutral definition layer + deterministic dual
publication of the 11 agents. Sibling cards M1 (skills canonicalization to `.agents/skills`),
M2 (AGENTS.md), M3 (hook adapter), M4 (`moai init --agent` wiring generator), and M6 (plugin
surface) are explicitly out of scope (§ Out of Scope). Where M4 will consume M5's output, the
seam is noted in plan.md §H.

## §B Glossary

| Term | Meaning |
|---|---|
| Neutral definition layer | The single source pair from which both publications derive: the agent `.md` definitions (name / description / model / effort / skills / tools frontmatter + body) plus one machine-readable Codex mapping manifest |
| Emitter | The deterministic generator in the template pipeline that reads the neutral layer and produces the published artifacts |
| Byte-identity (regression ban) | Published `.md` content is byte-equal to the committed template `.md` files, so `moai update` produces zero diff for existing users |
| Documented drop | A Claude-side capability with no Codex equivalent, recorded with rationale in the mapping manifest rather than silently discarded |
| Probe | A run-phase measurement of an unmeasured Codex semantic, executed through the t91 harness pattern (isolated `CODEX_HOME`, `codex exec --json`) |

## §C Requirements (GEARS)

### R-001 — Single neutral source (Ubiquitous) `[AC-003, AC-013]`

The agent publisher shall derive both published forms of every retained agent — the Claude Code
`.md` and the Codex `.toml` — from a single neutral definition layer consisting of the agent
`.md` definitions plus exactly one machine-readable Codex mapping manifest, with no
hand-maintained second representation of any agent.

### R-002 — `.md` byte-identity (Event-driven) `[AC-001]`

**When** the emitter publishes the agent set, the published `.claude/agents/moai/*.md` content
shall be byte-identical (sha256 equality) to the current committed template files for all 11
agents.

### R-003 — No `.md` rewrite (Unwanted) `[AC-002]`

The publication pipeline shall not re-render, reformat, re-order, or re-flow any committed
template `.md` file; the regression ban for existing `moai update` users shall hold by
construction, not merely by review.

### R-004 — TOML emission (Event-driven) `[AC-003]`

**When** an agent definition is published, the emitter shall emit a Codex agent TOML carrying at
minimum `name`, `description`, and `developer_instructions`, where `developer_instructions`
equals the `.md` body text verbatim and `name` equals the `.md` frontmatter `name`.

### R-005 — Verbatim body round trip (Ubiquitous) `[AC-001, AC-003]`

The agent body prose shall survive the neutral→publication round trip byte-unmodified in both
artifacts (identity in the `.md`; byte-equal body inside `developer_instructions` in the TOML).

### R-006 — Determinism (Event-driven) `[AC-004]`

**When** the emitter runs twice against identical inputs, it shall produce byte-identical
outputs — stable key order, no timestamps, no absolute paths, and no environment-derived values
embedded in any artifact.

### R-007 — Mapping-table authority (Ubiquitous) `[AC-007, AC-008, AC-013]`

The tools→Codex mapping shall follow the semantic-class mapping table codified in the mapping
manifest (plan.md §A.3), and every tool token appearing in any agent `tools:` CSV shall belong
to a mapped class.

### R-008 — Fail-closed self-validation (Event-driven) `[AC-005, AC-006]`

**When** an unknown tool token, an unmapped effort value, or an invalid sandbox value is
detected during emission, the emitter shall fail with a non-zero exit and a diagnostic naming
the offending file and token, and shall leave no partially-updated artifact set. This
requirement exists because Codex silently ignores unknown config values (M0 t91 §1).

### R-009 — MCP server mapping (Capability gate) `[AC-007]`

**Where** an agent's `tools:` list contains any `mcp__moai__*` entry, the emitted TOML shall
declare the moai server as a `[mcp_servers.moai]` table (`command = "moai"`,
`args = ["mcp-server"]`); the per-tool whitelist distinction shall be recorded as a
documented drop (coarse server-level grant) in the mapping manifest. The table shape is a
run-phase measured correction: codex-cli 0.147.0 rejects the array form
`mcp_servers = ["moai"]` ("invalid type: sequence, expected a map"), so the array literal
in the original requirement text was an unmeasured assumption (progress.md §E.2 MS3b,
commit e6c2239e5).

### R-010 — Effort mapping (Ubiquitous) `[AC-008, AC-P02]`

The emitter shall map each agent's `effort:` value to a Codex `model_reasoning_effort` value
through the mapping manifest's measured enumeration, and an unmapped source value shall block
emission (fail-closed). The v1 value set shall be settled by run-phase probe, not assumed.

### R-011 — Model handling (Ubiquitous) `[AC-009, AC-P03]`

The emitter shall omit the `model` field from every emitted TOML unless the mapping manifest
carries an explicit per-agent override; in v1 no override exists, and the single Claude-side
model pin (`manager-git: sonnet`) shall be recorded as a documented drop (a Claude alias is not
a Codex model id).

### R-012 — Template-first placement and distribution (Ubiquitous) `[AC-010]`

The neutral layer, the emitter, and the generated artifacts shall live in the template pipeline
under `internal/template/`, with the 11 TOML files committed under
`internal/template/templates/.codex/agents/moai/`, embedded by `make build`, and distributed by
`moai update` unchanged.

### R-013 — Documented drops (Capability gate) `[AC-013]`

**Where** an agent capability has no Codex equivalent (Task* family, `Agent` tool, DesignSync,
per-agent hooks, per-agent web-tool grants, per-tool MCP filtering, memory scopes, UI metadata),
the mapping manifest shall record a documented drop with rationale rather than a silent
discard.

### R-014 — Probe verification of unmeasured semantics (Event-detected) `[AC-P01..AC-P06]`

**When** the run-phase probe harness measures a Codex semantic marked unmeasured in plan.md
§A.4 (sandbox value set, `model_reasoning_effort` enumeration, model-omission inheritance,
agents-dir subdirectory scanning, skills.config value set, per-agent MCP tool filtering), the
measured result shall be recorded in progress.md §E.2 and the mapping manifest shall be updated
to the measured enumeration — or the affected field omitted — before SPEC close.

## §D Constraints

1. **[HARD] Regression ban**: the emitted `.md` publication must be byte-identical to the
   current template files; existing users must see zero diff from `moai update` (R-002/R-003).
2. **[HARD] Verbatim body**: the neutral format must carry the agent body prose verbatim; the
   body is the agent's prompt (R-005).
3. **[HARD] Silent-ignore countermeasure**: because Codex does not validate unknown field
   values (M0), the emitter must fail closed on any value outside a measured enumeration
   (R-008); a field whose value set is not probe-confirmed ships OMITTED (inherit Codex
   default), never guessed (R-014).
4. **[HARD] Template-First**: neutral definitions + emitter live in the template pipeline under
   `internal/template/`; `make build` regenerates embedded assets; the `.toml` files are
   committed alongside the `.md` files in the template tree (R-012).
5. **[HARD] Scope fence**: M5 does not implement the `--agent` wiring flag (M4), skills
   canonicalization (M1), AGENTS.md (M2), or the hook adapter (M3).
6. **Template neutrality**: new files under `internal/template/templates/` must pass the
   template-neutrality CI guard (no SPEC IDs, no internal dates/SHAs in distributed content).
   The agent bodies already satisfy this today; the emitter introduces no new tokens.
7. **Measurement baseline**: all Codex-behavior facts cited by this SPEC come from t91
   (codex-cli 0.147.0). Facts not measured there are marked unmeasured and are probe items,
   not assumptions.

## §E Measured Facts vs Unmeasured (Evidence Grounding)

| Codex semantic | Status | Source / resolution path |
|---|---|---|
| Agents-TOML load + delegation | **Measured, works** | t91 §2 (`T91PROBE-OK`) |
| Fields `name`, `description`, `developer_instructions`, `model_reasoning_effort`, `sandbox_mode` | **Measured, accepted** | t91 §2 probe file |
| Unknown config values silently ignored | **Measured (hazard)** | t91 §1 `T91BogusEvent` → R-008 |
| `codex mcp add moai` + all 21 tools in-session | **Measured, works** | t91 §5 |
| MCP calls in non-interactive `codex exec` need approval-policy handling | **Measured** | t91 §5 note ("user cancelled MCP tool call") |
| `sandbox_mode` allowed value set | **Unmeasured** | Probe AC-P01 |
| `model_reasoning_effort` allowed enumeration | **Unmeasured** | Probe AC-P02 |
| `model` field omission semantics; arbitrary-string acceptance | **Unmeasured** | Probe AC-P03 |
| `.codex/agents/` subdirectory scanning (`moai/` subdir vs flat) | **Unmeasured** | Probe AC-P04 |
| `skills.config` allowed value set | **Unmeasured** | Deferred to M1; optional probe AC-P05 |
| Per-agent tool filtering inside one MCP server | **Unmeasured** | Documented drop; optional probe AC-P06 |

## §F Traceability Matrix (structure)

| Requirement | Acceptance criteria |
|---|---|
| R-001 | AC-003, AC-013 |
| R-002 | AC-001 |
| R-003 | AC-002 |
| R-004 | AC-003 |
| R-005 | AC-001, AC-003 |
| R-006 | AC-004 |
| R-007 | AC-007, AC-008, AC-013 |
| R-008 | AC-005, AC-006 |
| R-009 | AC-007 |
| R-010 | AC-008, AC-P02 |
| R-011 | AC-009, AC-P03 |
| R-012 | AC-010 |
| R-013 | AC-013 |
| R-014 | AC-P01..AC-P06 |
| Documentation grounding | AC-011, AC-012 (§D.6/§D.7 constraints — deliberate non-REQ anchor) |

## Out of Scope

The following are out of scope for this SPEC:

### Out of Scope — Sibling cards M1–M4 and M6 of the Codex dual-harness series

- M1 (skills canonicalization to `.agents/skills`, symlink dual-coverage, `skills.config`
  emission): the `Skill` tool class and `skills:` preload are deferred to M1; M5 emits no
  skills field. Seam: when M1 lands, it extends the mapping manifest — no M5 artifact changes.
- M2 (AGENTS.md budget/merge semantics — t82 lane).
- M3 (hook adapter): per-agent Claude `hooks:` frontmatter has no Codex per-agent equivalent;
  Codex hooks live in project-level `.codex/hooks.json` (t91 §6). M5 records the drop; M3 owns
  the adapter (including the `PostToolUse` + `collaboration*` matcher redesign per t91 §8).
- M4 (wiring generator, `moai init --agent`): M4 consumes M5's committed `.toml` artifacts and
  the emitter's validation surface; M5 implements no CLI flag.
- M6 (plugin surface): `plugin_hooks` is removed on the measured build (t91 §8) — no plugin
  carriage path exists in this SPEC.

### Out of Scope — Agent body prose portability

- Rewriting, annotating, or conditionally adapting agent body prose for Codex semantics. The
  bodies reference Claude-specific surfaces (AskUserQuestion prohibitions, ToolSearch preload,
  Claude tool names); they are carried verbatim into `developer_instructions` and the
  Claude-specific references degrade silently on Codex. Body portability is future work
  outside M5.

### Out of Scope — Codex-side permission enforcement

- Enforcing per-tool whitelists, per-agent permission modes (`permissionMode`), or memory
  scopes on the Codex side. These are documented drops with rationale (plan.md §A.3); building
  Codex-side enforcement machinery is out of scope.

### Out of Scope — Changes to the 11 agent definitions

- Any content, tool-grant, effort, or model change to the agent `.md` files themselves. The
  `.md` files are frozen inputs (R-002/R-003). Catalog changes go through the agent-catalog
  policy, not this SPEC.

## §G History

| Date | Author | Change |
|---|---|---|
| 2026-08-22 | manager-spec | Initial plan-phase authoring (card t89, Tier M). M0 evidence: t91 report (codex-cli 0.147.0). |

## §H Cross-References

- M0 measurement report: `.moai/reports/t91/README.md` + `hook-payloads/` (primary checkout)
- Agent inventory ground truth: `internal/template/templates/.claude/agents/moai/*.md` (11 files)
- Agent namespace SSOT: `.claude/rules/moai/development/agent-authoring.md` § Agent Directory Convention
- Model policy (inherit-by-default, manager-git sonnet pin): `.claude/rules/moai/development/model-policy.md` § Inherit-by-Default Convention
- moai MCP tool catalogue (21 tools): `.claude/rules/moai/core/moai-mcp-tools.md`
- Template-First rule and neutrality guard: `CLAUDE.local.md` §2 / `.moai/docs/template-internal-isolation-doctrine.md`
- plan.md §A.3 — the tools→Codex mapping table (first-class plan-phase deliverable)
- plan.md §H — M4/M1/M3 consumption seams
