# moai-mcp Tool Catalogue

> Single source of truth for the 21 tools exposed by the self-hosted `moai` MCP
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

## Tool families (21 tools)

| Family | Tools | Wired consumers |
|---|---|---|
| SPEC lifecycle | `spec_progress`, `spec_audit`, `spec_drift` | manager-spec, manager-docs, plan-auditor, super-advisor |
| Verification snapshots | `verify_snapshot`, `verify_trend` | manager-develop, sync-auditor, super-advisor |
| Goal + session | `goal_arm`, `goal_status`, `session_list` | orchestrator only / manager-develop, manager-lead |
| Cross-model audit | `audit_multi`, `codex_audit`, `glm_audit`, `audit_cache` | plan-auditor, sync-auditor |
| Codex delegation | `codex_task`, `codex_setup`, `codex_job_{status,result,cancel}` | super-advisor |
| GLM delegation | `glm_task`, `glm_job_{status,result,cancel}` | super-advisor |

Per-tool purpose, consumer, and CLI equivalent: `moai-mcp-tools-catalogue.md`. Both codex and GLM
are OPTIONAL and fail open — an unavailable backend returns `inconclusive`, never a hard error.

## Unwired-by-design

`goal_arm` is intentionally wired to NO agent. Arming an autonomous loop is an
orchestrator concern (preserves the orchestrator-only arming surface and the
flat hierarchy invariant). Agents that need a goal's state read `goal_status`;
only the orchestrator arms.

---

Classification: Evolvable reference rule — the MCP tool surface map. Update this
file whenever a tool is added/removed/renamed on the `moai mcp-server` (the Go
producer lives in `internal/cli/mcp_server.go`).
