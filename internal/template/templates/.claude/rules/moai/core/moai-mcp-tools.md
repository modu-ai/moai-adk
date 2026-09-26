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

The caller is the only party that holds the answer, which is why it is an input
rather than something inferred.

| Situation | What to pass | What happens |
|---|---|---|
| Session in a worktree | `project_root: <git rev-parse --show-toplevel>` | the call acts on that tree |
| Session in the primary checkout | nothing | resolves exactly as it always has |
| Path that is not a MoAI project root | — | the call is REJECTED with an error naming the path |

The rejection is deliberate and is not a rough edge. A silent fallback to the
default would send a caller who mistyped its own worktree path back to acting on
the primary checkout — the exact failure the parameter exists to prevent —
while reporting success.

An accepted path is **canonicalized** before use — symlinks are resolved, so the
call acts on the real directory rather than on whichever spelling reached it, and
a later containment check cannot be walked through by pointing a link at a tree
outside the boundary. A path that cannot be canonicalized is rejected on the same
terms as any other unusable one.

A registered linked worktree of a repository that keeps `.moai` untracked is also
accepted; rules and caveats: `moai-mcp-tools-catalogue.md` § Linked worktrees.

For `audit_multi` the root reaches every backend in the fan-out: Claude and GLM
use it to collect the diff sent to their isolated reviewer, while codex receives
it as the working directory it reviews in. Passing it keeps all independent
opinions about the same tree.

## Tool families (35 of the 39 tools; the session-messaging family follows below)

| Family | Tools | Wired consumers |
|---|---|---|
| SPEC lifecycle | `spec_progress`, `spec_audit`, `spec_drift` | manager-spec, manager-docs, plan-auditor, super-advisor |
| Verification snapshots | `verify_snapshot`, `verify_trend` | manager-develop, sync-auditor, super-advisor |
| Goal + session | `goal_arm`, `goal_status`, `session_list` | orchestrator only / manager-develop, manager-lead |
| Cross-model audit | `audit_multi`, `claude_audit`, `codex_audit`, `glm_audit`, `audit_cache` | plan-auditor, sync-auditor |
| Codex delegation | `codex_task`, `codex_setup`, `codex_job_{status,result,cancel}` | super-advisor |
| Codex read-only roles | `codex_role_audit`, `codex_role_audit_status`, `codex_role_audit_result` | the Codex lane orchestrator — starts `plan-auditor`, `sync-auditor`, `mission-governor`, `super-advisor` as a top-level read-only process instead of through `spawn_agent` |
| GLM delegation | `glm_task`, `glm_job_{status,result,cancel}` | super-advisor |
| Code queries | `graph_file_api`, `graph_find_code`, `graph_trace_calls`, `graph_shortest_path` | any agent (signature-level code navigation from the code-derived edge layer; every answer carries tree+commit provenance) |
| Judgment (gated) | `jev_ask` | gated-unavailable at the shipped default (`workflow.jev.enabled: false`) — no request constructed, no network call; while the chain's fitness gate stands unrun it is not presented as available. Display-only: a labelled model signal a person reads, never a completion predicate, merge approval, queue mutation, or gate input |
| Factory messaging | `factory_msg_send`, `factory_msg_list`, `factory_msg_body`, `factory_msg_receipt`, `factory_msg_status` | attributed factory lead/worker sessions; `status` is the read-only operational roster/count surface |

Per-tool purpose, consumer, and CLI equivalent: `moai-mcp-tools-catalogue.md`. Claude, codex, and
GLM audit transports fail open at the tool boundary: an unavailable backend returns
`inconclusive`, never a hard Go error. An explicitly required gate can still turn that unmet
evidence requirement into `overall_verdict: fail`.

### Session messaging broker (Claude ↔ Codex)

| Tool | Purpose | Consumer | CLI equivalent |
|------|---------|----------|----------------|
| `mcp__moai__session_msg_register` | Register this session (kind: claude or codex, plus name) in the local message broker; idempotent — same kind+name returns the same agentId | Any Claude or Codex agent session | — |
| `mcp__moai__session_msg_list` | List registered broker agents (agentId, name, kind, online, pending count) — the family's only read-only tool | Any Claude or Codex agent session | — |
| `mcp__moai__session_msg_send` | Send a short, self-contained fact message to another registered agent's mailbox | Any Claude or Codex agent session | — |
| `mcp__moai__session_msg_poll` | Claim pending messages from this agent's mailbox (at-least-once) and optionally ack processed ids | Any Claude or Codex agent session | — |

Reach any session kind symmetrically: a Codex session has no native peer-messaging runtime, so the broker is its only path, while claude↔claude keeps the native runtime as the recommended route. Delivery is poll-based — a send is a record, not a delivery guarantee. Newly added tools become visible only after the session restarts its MCP server (a long-lived server does not see tools added after it started), so the restart is procedure step zero whenever a new tool seems missing.

## Unwired-by-design

`goal_arm` is intentionally wired to NO agent. Arming an autonomous loop is an
orchestrator concern (preserves the orchestrator-only arming surface and the
flat hierarchy invariant). Agents that need a goal's state read `goal_status`;
only the orchestrator arms.

---

Classification: Evolvable reference rule — the MCP tool surface map. Update this
file whenever a tool is added/removed/renamed on the `moai mcp-server` (the Go
producer lives in `internal/cli/mcp_server.go`).
