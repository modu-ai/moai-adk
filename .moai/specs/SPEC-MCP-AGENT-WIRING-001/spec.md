---
id: SPEC-MCP-AGENT-WIRING-001
title: "Agent MCP tool wiring — phase-scoped mcp__moai__* grants across the 12 retained agents"
version: "0.2.0"
status: completed
created: 2026-08-12
updated: 2026-08-12
author: manager-spec
priority: P1
phase: "v3.2 target"
module: ".claude/agents/moai, internal/template/templates/.claude/agents/moai"
lifecycle: spec-anchored
tier: M
tags: "mcp, agents, frontmatter, template-first, blast-radius, model-ssot"
depends_on: [SPEC-MCP-DEFAULT-ON-001]
---

# SPEC-MCP-AGENT-WIRING-001 — Agent MCP tool wiring (Epic SPEC-B)

## HISTORY

- 2026-08-12 (plan-phase, iter-1) — Initial Tier M authoring. Second SPEC of the moai-MCP integration Epic (A → B → C). Grants each retained agent the `mcp__moai__*` tools matching its phase scope, mirrors every edit per Template-First, and records the state of the `ResolveAgentModelEffort` SSOT across the model-invoking tool surface. Depends on SPEC-MCP-DEFAULT-ON-001, which makes the server present by default; without it these grants would name a server most projects do not have.
- 2026-08-12 (plan-phase, iter-2) — Revision after plan-audit iter-1 returned FAIL (0.81). The blocking finding D1 corrected a factually wrong §C premise: the codex session model-resolution path is **already** structurally guarded by `TestMCPAudit_NoDirectFrontmatterRead` (`internal/cli/mcp_audit_test.go:148`), which scans `mcp_codex.go` — where `openCodexSession`, `openCodexSessionOn`, and the `threadParams["model"] = me.Model` assignment all live — and `codex_task.go:222` delegates to `openCodexSessionOn` with no independent resolution. The redundant M2 Go-test milestone was dropped; REQ-B-7 (add a regression guard) was removed; the deliverable is now honestly frontmatter + mirror only, with **no Go code change**. D2 (auditor false-positive on tool count) refuted — `grep -c "mcp.NewTool(" mcp_server.go` = 17, matching the plan. D3 (minor): "write-capable" in REQ-B-2 pinned to its semantic definition.

## §A. User Story

**As a** MoAI agent working a SPEC phase,

**I want** the moai state tools relevant to my phase available as typed MCP tools in my own `tools:` list,

**so that** I read SPEC lifecycle state, verification history, and audit results through a declared schema instead of guessing CLI flags — and so that tools outside my phase scope are not available to me at all.

## §B. Scope Summary

This SPEC edits **agent frontmatter only**. It adds `mcp__moai__*` entries to the `tools:` CSV of the retained agents whose phase scope calls for them, mirrors each edited file into the template source, and rebuilds the embedded catalog.

It changes no tool handler, no schema, no Go code, and no `.mcp.json`. The model/effort SSOT (`ResolveAgentModelEffort`) is already complete across all four model-invoking tools at base — see §C and `plan.md` §C for the measurement. This SPEC adds **no Go test and no structural guard**; the existing `TestMCPAudit_NoDirectFrontmatterRead` already covers the codex session path because the entire path lives in `mcp_codex.go`, which that test scans.

The delivery pattern is already established in the tree: `plan-auditor.md:7` and `sync-auditor.md:9` each already carry `mcp__moai__audit_multi` in their `tools:` CSV at base. This SPEC extends that same pattern; it invents nothing.

### Out of Scope — provisioning and defaults

- Whether the `moai` MCP server is present in a project at all → **SPEC-MCP-DEFAULT-ON-001** (SPEC-A), on which this SPEC depends.

### Out of Scope — the web console

- Per-tool enablement UI, codex OAuth, GLM API-key configuration → **SPEC-MCP-CONSOLE-001** (SPEC-C).

### Out of Scope — new or changed MCP tools

- No tool is added, removed, renamed, or re-schema'd. The set stays at the 17 registered by `registerMoaiMCPTools` (`internal/cli/mcp_server.go:105`).

### Out of Scope — user-authored harness agents

- `.claude/agents/harness/**` specialists are user-owned per the harness namespace doctrine and are not edited here.

## §C. Requirements (GEARS)

> Domain prefix `REQ-B-N`. Citations read at base commit `ed70e4354`. This SPEC changes **no Go code**: the model/effort SSOT was measured at base and is already complete across all four model-invoking tools (see `plan.md` §C); REQ-B-6 below is verification-shaped, recording that finding honestly rather than authoring new work.

### M1 — Phase-scoped tool grants

**REQ-B-1** (Ubiquitous) Each of the 12 retained agents shall carry, in its `tools:` CSV frontmatter, exactly the `mcp__moai__*` entries matching its phase scope, and no others. The per-agent grant matrix — including the agents that receive **no** tools, which is a valid and documented outcome — is enumerated in `plan.md` §B with a justification for every add and every omission.

**REQ-B-2** (Unwanted) No agent shall receive a **write-capable** MCP tool unless its phase scope requires the write. The four write-capable tools are `goal_arm`, `verify_snapshot`, `codex_task`, and `codex_job_cancel`; the remaining 13 carry read-only hint annotations (verified against the `mcp.WithReadOnlyHintAnnotation(true)` calls in `registerMoaiMCPTools`, `internal/cli/mcp_server.go:105-200`). Each write-capable grant shall be individually justified in `plan.md` and shall be visible as such in the acceptance matrix. "Write-capable" here is a **semantic** classification — the tool mutates persistent state (e.g. `verify_snapshot` wraps `verify.RecordCheck`, which writes the verification baseline keyed by HEAD) — and is NOT derived from the per-tool `WithReadOnlyHintAnnotation` value; whether each tool's own annotation is accurate is an out-of-scope implementation concern of the MCP server, not of this SPEC.

**REQ-B-3** (Ubiquitous) Every edited `.claude/agents/moai/*.md` file shall be mirrored to `internal/template/templates/.claude/agents/moai/`, and `make build` shall regenerate the embedded catalog in the same change. The mirrored copy shall carry no SPEC ID, REQ token, commit SHA, internal date, `/Users/` path, or `CLAUDE.local.md` reference.

**REQ-B-4** (Unwanted) No agent's existing `tools:` entries shall be removed or reordered destructively; `mcp__moai__*` entries are appended to the existing CSV. The two agents that already carry `mcp__moai__audit_multi` (`plan-auditor.md:7`, `sync-auditor.md:9`) shall retain it.

**REQ-B-5** (Unwanted) The `tools:` field shall remain a **CSV string**, never a YAML array. (`skills:` is the only frontmatter field in this project that takes a YAML array.)

### M2 — Model/effort SSOT (verification-shaped; no code change)

**REQ-B-6** (State-driven) **While** any MCP tool resolves a model or an effort level, it shall do so through `template.ResolveAgentModelEffort` (`internal/template/profile_matrix.go:385`) and shall not read agent frontmatter or `llm.agent_overrides` directly. This requirement is **verification-shaped rather than change-shaped**: at base, all four model-invoking tools (`codex_audit`, `codex_task`, `glm_audit`, `audit_multi`) already resolve through the SSOT, and the existing structural guard `TestMCPAudit_NoDirectFrontmatterRead` (`internal/cli/mcp_audit_test.go:148`) already scans `mcp_codex.go` — where the entire codex session path (`openCodexSession`, `openCodexSessionOn`, and the `threadParams["model"] = me.Model` assignment at `mcp_codex.go:566`) lives. `internal/cli/codex_task.go` delegates to `openCodexSessionOn` at line 222 with no independent model resolution. The SSOT scope is therefore already complete and already structurally guarded at base; this SPEC records that finding honestly and adds no Go test, no structural guard, and no migration.

## §D. Constraints

- **C-B-1** — Frontmatter-only. No agent **body** prose is rewritten; only the `tools:` line changes.
- **C-B-2** — Template-First is mandatory: a missing mirror fails CI.
- **C-B-3** — `manager-lead` is the sole `Agent`-carrying retained agent and its depth-2 seal (`internal/template/manager_lead_depth_test.go`) is untouched. Adding MCP tools to a leaf worker does not add `Agent` to it.
- **C-B-4** — An agent receiving zero MCP tools is a valid outcome and shall be recorded as a decision, not left as an unexplained gap.

## §E. Risks

- **R-B-1 — Blast-radius creep via write-capable tools.** A coordinator or auditor granted `goal_arm` could change a session's termination condition; an agent granted `verify_snapshot` could write the verification baseline that later attributes its own claims. Mitigation: REQ-B-2 plus the per-grant justification in `plan.md` §B, and the deliberate omission of `goal_arm` from `manager-lead` (see `plan.md` §B.2).
- **R-B-2 — Grants that reference an absent server.** If SPEC-A has not landed, most projects have no `moai` MCP server and the granted tool names resolve to nothing. Mitigated by the `depends_on` declaration; a tool name that does not resolve is inert rather than an error.
- **R-B-3 — Mirror drift.** 12 source files and 12 mirrors is a shape that invites a missed copy. Mitigated by REQ-B-3's `make build` obligation and the existing CI parity guard.

## §F. Exclusions

### Out of Scope — agent body content

- No agent's role prose, workflow steps, or skill-loading section is edited. Only the `tools:` frontmatter line.

### Out of Scope — the `Explore` built-in

- `Explore` is an Anthropic built-in with no MoAI file; its tool set is not ours to edit. It is counted in the catalog of 12 but receives no grant by construction.

### Out of Scope — MCP tool implementation

- No handler, schema, annotation, or registration in `internal/cli/mcp_server.go` is modified.

### Out of Scope — model-policy changes

- The profile matrix, its cells, and the per-agent `{model, effort}` resolution are unchanged. REQ-B-6 records verification (the SSOT is already complete at base), not new resolution behavior.
