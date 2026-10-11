# moai-mcp Tool Catalogue

> Moved to the detail companion: `moai-mcp-tools-catalogue.md` ("moai-mcp Tool Catalogue").

## MCP-over-CLI rule

Prefer the MCP tool when it is in the calling agent's `tools:` list. The MCP path
and the Bash CLI back the SAME implementation; the MCP path returns structured
output, avoids shell-quoting hazards, and is lower-latency inside a subagent
where Bash may be restricted. Use the Bash CLI only when the MCP tool is absent
from the agent's `tools:` list, or when orchestrating from the main session and
the CLI form reads more naturally inline.

## The `project_root` input — name your own tree

> Moved to the detail companion: `moai-mcp-tools-catalogue.md` ("The `project_root` input — name your own tree").

[HARD] **An agent working inside a worktree MUST pass it**, and the value is its
own `git rev-parse --show-toplevel`. This is not a convenience. The server cannot
work the answer out for itself: it is a long-lived subprocess, so its working
directory cannot follow a worktree switch, and the environment variable it falls
back on names the PROJECT — the primary checkout — even for a session working in
a worktree. Omit the parameter from a worktree and the call acts on the primary
checkout instead, which means a SPEC that exists only on the card's branch is not
in the catalogue the auditor reads. It is not reported missing; it is simply
absent.

## Cross-reference

> Moved to the detail companion: `moai-mcp-tools-catalogue.md` ("Cross-reference").

---

Classification: Evolvable reference rule — the MCP tool surface map. Update this
file whenever a tool is added/removed/renamed on the `moai mcp-server` (the Go
producer lives in `internal/cli/mcp_server.go`).
