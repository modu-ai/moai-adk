---
description: "Detail companion for moai-mcp-tools.md — the full 31-tool moai MCP catalogue with per-family tables, consumers, and CLI equivalents"
paths: "**/moai-mcp-tools.md,**/internal/cli/mcp_server.go,**/.claude/agents/moai/*.md"
---

# moai-mcp Tool Catalogue — Detail Companion

> Detail companion of `moai-mcp-tools.md` (the always-loaded stub). The stub owns the
> MCP-over-CLI preference rule, the family index, and the unwired-by-design note. This file owns
> the per-tool catalogue: purpose, wired consumer, and CLI equivalent for each of the 31 tools.
> Load it when wiring a tool into an agent's `tools:` list, or when choosing between an MCP tool
> and its Bash equivalent for a specific capability.

## Tool catalogue (31 tools)

### SPEC lifecycle

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__spec_progress` | List SPEC docs + frontmatter | manager-spec, manager-docs | `moai spec list` |
| `mcp__moai__spec_audit` | SPEC lifecycle audit (era + drift) | manager-spec, manager-docs, plan-auditor, super-advisor | `moai spec audit` |
| `mcp__moai__spec_drift` | Modern-era V3R6 drift findings | manager-spec, plan-auditor | `moai spec audit` (drift view) |

Reach for these in **plan-phase** (manager-spec authoring a new SPEC, checking era
classification + drift) and **sync-phase** (manager-docs verifying lifecycle
closure). `spec_progress` enumerates existing SPECs + frontmatter; `spec_audit`
classifies era and detects drift across the catalog; `spec_drift` is the focused
modern-era V3R6 drift slice. plan-auditor uses `spec_audit`/`spec_drift` for
plan-phase skeptical review.

### Verification snapshots

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__verify_snapshot` | Read/record per-key verification snapshot | manager-develop | `moai verify check` |
| `mcp__moai__verify_trend` | Per-key verification check history | manager-develop, sync-auditor, super-advisor | `moai verify check` |

Used by **manager-develop** during run-phase self-verification (§E), and by
sync-auditor / super-advisor for trend review. `verify_snapshot` reads or records
the per-key snapshot keyed by HEAD digest; `verify_trend` surfaces the check
history to judge convergence over time. The orchestrator's attributable diff-check
consults the current snapshot key before re-executing tests.

### Goal + session (autonomous loop)

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__goal_arm` | Arm a condition-declared goal | **orchestrator main session ONLY** — wired to NO agent (arming an autonomous loop is an orchestrator concern) | `moai goal arm` / `/moai goal` |
| `mcp__moai__goal_status` | Read armed-goal state | manager-develop, manager-lead | `moai goal status` |
| `mcp__moai__session_list` | List active moai sessions | manager-lead | `moai session list` |

`goal_arm` is orchestrator-only and arms an autonomous loop — never inside an
agent (preserves the flat-hierarchy arming surface). `goal_status` lets
manager-develop / manager-lead read the armed condition's progress; `session_list`
lets manager-lead detect concurrent sessions on the same checkout for race
mitigation before fan-out.

### Cross-model audit (second opinion)

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__audit_multi` | Multi-auditor convergence (claude + codex + glm) | plan-auditor, sync-auditor | — (MCP-only convergence entry) |
| `mcp__moai__claude_audit` | Independent Claude subscription audit (`claude -p` with read-only isolation and structured output) | plan-auditor, sync-auditor; automatically by `audit_multi` in GPT/GLM sessions | — |
| `mcp__moai__codex_audit` | codex backend single audit (native/adversarial) | plan-auditor, sync-auditor | — |
| `mcp__moai__glm_audit` | GLM (z.ai) backend single audit | plan-auditor, sync-auditor | — |
| `mcp__moai__audit_cache` | plan-audit PASS cache (compute_hash/lookup/store, process-shared) | sync-auditor | `moai audit cache` (none — MCP-only) |

`claude_audit` accepts `target`, `focus`, optional `model`/`effort`, and
`project_root`. It permits subscription login only (`authMethod=claude.ai`,
`apiProvider=firstParty`), defaults to `sonnet/high`, strips gateway/provider
environment variables, disables tools and session persistence, and reports
resolved-model plus usage provenance without exposing credentials.

`audit_multi` chooses the Claude participant from the launch provider. A Claude
main session reuses its in-session `claude_verdict` as an anchor. A GPT, GLM, or
unknown-origin session ignores any caller-supplied Claude verdict and invokes
the independent `claude_audit` backend. Codex and GLM continue to run as their
configured gates require. Every backend tool fails open to `inconclusive`; an
explicitly required gate left inconclusive makes the convergence verdict fail.

### Codex delegation (background jobs)

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__codex_task` | Delegate a coding/investigation task to codex (sync or background) | super-advisor | `moai codex task` |
| `mcp__moai__codex_setup` | Probe local codex install (LookPath + version + auth) | super-advisor | `moai codex setup` |
| `mcp__moai__codex_job_status` | Read a background codex job's status/record | super-advisor | `moai codex job status` |
| `mcp__moai__codex_job_result` | Read a background codex job's output | super-advisor | `moai codex job result` |
| `mcp__moai__codex_job_cancel` | Stop a running background codex job | super-advisor | `moai codex job cancel` |

The codex delegation family is wired into `super-advisor` because the on-demand
high-reasoning consultation agent is the natural consumer of background
cross-model delegation: it arms a codex task via `codex_task`, polls completion
via `codex_job_status`/`codex_job_result`, and cancels via `codex_job_cancel`.
`codex_setup` probes whether codex is available before delegating. codex is
OPTIONAL: a missing or unavailable codex yields a fail-open `inconclusive`, never
a hard error.

### GLM delegation (background jobs)

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__glm_task` | Delegate a task (arbitrary prompt) to GLM (z.ai) (sync or background) | super-advisor | — (no `moai glm task` CLI exists) |
| `mcp__moai__glm_job_status` | Read a background GLM job's status/record | super-advisor | — |
| `mcp__moai__glm_job_result` | Read a background GLM job's output | super-advisor | — |
| `mcp__moai__glm_job_cancel` | Stop a running background GLM job | super-advisor | — |

The GLM delegation family mirrors the codex delegation family against the z.ai
HTTP backend and is wired into `super-advisor` the same way: it arms a GLM task
via `glm_task` (sync returns the completed text, background returns a job id),
polls completion via `glm_job_status`/`glm_job_result`, and cancels via
`glm_job_cancel`. There is no `codex_setup` counterpart — availability is
learned from `glm_task` itself, which reports a structured failed result when
the key is missing or z.ai is unreachable. GLM is OPTIONAL: a missing or
unavailable GLM yields a fail-open result, never a hard error.

### Judgment (gated, display-only)

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__jev_ask` | Ask the gated judgment capability typed questions over one supplied state; typed answers with probability | gated-unavailable at the shipped default (`workflow.jev.enabled: false`); display-only — a labelled model signal a person reads, never a completion predicate, merge approval, queue mutation, or gate input | — (MCP-only) |

The tool is registered unconditionally so its gate-off contract is invocable
and countable, but with the gate off it constructs no request and makes no
network call, and while the chain's fitness measurement gate stands unrun no
surface presents it as available. The question-design rules for authoring
well-formed questions live in the reference skill; the call path lives in
`internal/jev` and the tool wraps it without a second implementation.


---

Classification: Lazy companion — catalogue tables and per-family guidance only. The preference rule
stays in `moai-mcp-tools.md`. Update this file whenever a tool is added, removed, or renamed on the
`moai mcp-server` (the Go producer lives in `internal/cli/mcp_server.go`).
