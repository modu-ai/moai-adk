# moai-mcp Tool Catalogue

> Single source of truth for the 17 tools exposed by the self-hosted `moai` MCP
> server (`.mcp.json` → `{command: "moai", args: ["mcp-server"]}`). Each tool is
> prefixed `mcp__moai__` at the call site. This rule tells agents and the
> orchestrator WHEN to prefer an MCP tool over its CLI/slash equivalent.
>
> Wiring parity (local ↔ template) and per-agent `tools:` lists are owned by the
> agent definitions; this file owns the capability map + the MCP-over-CLI rule.

## MCP-over-CLI rule

Prefer the MCP tool when it is in the calling agent's `tools:` list. The MCP path
and the Bash CLI back the SAME implementation; the MCP path returns structured
output, avoids shell-quoting hazards, and is lower-latency inside a subagent
where Bash may be restricted. Use the Bash CLI only when the MCP tool is absent
from the agent's `tools:` list, or when orchestrating from the main session and
the CLI form reads more naturally inline.

## Tool catalogue (17 tools)

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
| `mcp__moai__goal_status` | Read armed-goal state | manager-develop, manager-kanban | `moai goal status` |
| `mcp__moai__session_list` | List active moai sessions | manager-kanban | `moai session list` |

`goal_arm` is orchestrator-only and arms an autonomous loop — never inside an
agent (preserves the flat-hierarchy arming surface). `goal_status` lets
manager-develop / manager-kanban read the armed condition's progress; `session_list`
lets manager-kanban detect concurrent sessions on the same checkout for race
mitigation before fan-out.

### Cross-model audit (second opinion)

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__audit_multi` | Multi-auditor convergence (claude + codex + glm) | plan-auditor, sync-auditor | — (MCP-only convergence entry) |
| `mcp__moai__codex_audit` | codex backend single audit (native/adversarial) | plan-auditor, sync-auditor | — |
| `mcp__moai__glm_audit` | GLM (z.ai) backend single audit | plan-auditor, sync-auditor | — |
| `mcp__moai__audit_cache` | plan-audit PASS cache (compute_hash/lookup/store, process-shared) | sync-auditor | `moai audit cache` (none — MCP-only) |

Single-backend audit mode is selected per the project's `audit_model`:
`codex+glm` (default, converge via `audit_multi`) | `glm` | `codex` | `none`
(Claude-only, no backend call). All backends are fail-open: an unavailable
backend returns `inconclusive`, never a Go error.

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

## Unwired-by-design

`goal_arm` is intentionally wired to NO agent. Arming an autonomous loop is an
orchestrator concern (preserves the orchestrator-only arming surface and the
flat hierarchy invariant). Agents that need a goal's state read `goal_status`;
only the orchestrator arms.

---

Classification: Evolvable reference rule — the MCP tool surface map. Update this
file whenever a tool is added/removed/renamed on the `moai mcp-server` (the Go
producer lives in `internal/cli/mcp_server.go`).
