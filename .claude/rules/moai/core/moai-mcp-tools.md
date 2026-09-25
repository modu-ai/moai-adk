# moai-mcp Tool Catalogue

> Single source of truth for the 39 tools exposed by the self-hosted `moai` MCP
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

## The `project_root` input — name your own tree

Thirteen tools accept an optional `project_root` string: `spec_progress`,
`spec_audit`, `spec_drift`, `verify_snapshot`, `verify_trend`, `codex_audit`,
`glm_audit`, `claude_audit`, `audit_multi`, `graph_file_api`, `graph_find_code`,
`graph_shortest_path`, and `graph_trace_calls`. It names the tree the call
should act on.

[HARD] **An agent working inside a worktree MUST pass it**, and the value is its
own `git rev-parse --show-toplevel`. This is not a convenience. The server cannot
work the answer out for itself: it is a long-lived subprocess, so its working
directory cannot follow a worktree switch, and the environment variable it falls
back on names the PROJECT — the primary checkout — even for a session working in
a worktree. Omit the parameter from a worktree and the call acts on the primary
checkout instead, which means a SPEC that exists only on the card's branch is not
in the catalogue the auditor reads. It is not reported missing; it is simply
absent.

The caller is the only party that holds the answer, which is why it is an input.

| Situation | What to pass | What happens |
|---|---|---|
| Session in a worktree | `project_root: <git rev-parse --show-toplevel>` | the call acts on that tree |
| Session in the primary checkout | nothing | resolves exactly as it always has |
| Path that is not a MoAI project root | — | the call is REJECTED with an error naming the path |

The rejection is deliberate, not a rough edge: a silent fallback would send a
caller who mistyped its own worktree path back to the primary checkout — the
exact failure the parameter exists to prevent — while reporting success.

An accepted path is **canonicalized** before use, so the call acts on the real
directory rather than on whichever spelling reached it and a containment check
cannot be walked through by pointing a link outside the boundary. A path that
cannot be canonicalized is rejected on the same terms.

For `audit_multi` the root reaches every backend in the fan-out, keeping all
independent opinions about the same tree.

## Cross-reference

`moai-mcp-tools-catalogue.md` — the lazy companion. Load it for § Tool catalogue
(39 tools) · § Tool families (the family-to-consumer map) · § Session messaging
broker (Claude ↔ Codex) · § Unwired-by-design (why `goal_arm` reaches no agent).

---

Classification: Evolvable reference rule — the MCP tool surface map. Update this
file whenever a tool is added/removed/renamed on the `moai mcp-server` (the Go
producer lives in `internal/cli/mcp_server.go`).
