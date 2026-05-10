---
paths: "**/.moai/specs/**,**/internal/worktree/**,**/internal/cli/worktree/**"
---

# Worktree State Guard

Operational rule for the orchestrator's defense layer against
`Agent(isolation: "worktree")` regressions. Wave 5 of SPEC-V3R3-CI-AUTONOMY-001
introduces a Bash-invocable Go primitive (`moai worktree snapshot|verify|restore`)
that captures pre-call state, detects post-call divergence, and surfaces empty
`worktreePath` responses as suspect flags.

This rule defines **when** and **how** the orchestrator must invoke the
primitive. The primitive itself lives in `internal/worktree/` and is exposed as
CLI subcommands by `internal/cli/worktree/guard.go`.

> Cross-references: `worktree-integration.md` (broader worktree patterns),
> `agent-common-protocol.md` § User Interaction Boundary (AskUserQuestion HARD),
> SPEC `.moai/specs/SPEC-V3R3-CI-AUTONOMY-001/spec.md` § 3.6 (REQ-CIAUT-031~036).

## When to Snapshot

The orchestrator SHOULD invoke `moai worktree snapshot` immediately before any
`Agent(isolation: "worktree")` call that:

- Modifies tracked files (any agent with `permissionMode: acceptEdits`)
- Operates in team mode where parallel teammates are spawned with isolation
- Performs cross-file refactors (e.g., expert-refactoring, manager-develop)
- Has historically triggered worktree regressions on this project

Snapshots SHOULD be skipped for:

- Read-only agents (`permissionMode: plan`) — no state can drift
- One-shot file reads (Read tool only)
- Trivial edits to a single config file (low blast radius, snapshot overhead exceeds benefit)

## Divergence Threshold

Per OQ4 in strategy-wave5.md §7: **binary detection**.

A snapshot is considered divergent if **any** of the following changed between
pre-call and post-call:

- HEAD SHA (`git rev-parse HEAD`)
- Current branch (`git rev-parse --abbrev-ref HEAD`)
- Untracked file set under `.moai/specs/` (added or removed)
- Porcelain status lines (`git status --porcelain`, excluding `.moai/state/`,
  `.moai/reports/`, `.moai/cache/`, `.moai/logs/`)

There is no "soft" threshold; a configurable threshold is explicitly out of
Wave 5 scope and may be added as a follow-up SPEC if false-positive rates require.

## Escalation Path

When `moai worktree verify` exits non-zero:

| Exit Code | Meaning | Orchestrator Action |
|-----------|---------|--------------------|
| 0 | Clean | Continue normally |
| 1 | Divergence detected | Read the JSON report (stdout), then `AskUserQuestion(restore / accept / abort)` |
| 2 | Suspect (empty worktreePath in agent response) | Warn the user, set push-block flag, optionally trigger claude-code-guide investigation |
| 3 | Both divergence + suspect | Combine 1 + 2 escalation steps |

[HARD] AskUserQuestion is invoked by the **orchestrator only**. The Go CLI
returns exit codes and JSON; the orchestrator translates these into user-facing
prompts. See `agent-common-protocol.md` § User Interaction Boundary.

## Invocation Pattern (Bash CLI Sequence)

The canonical sequence the orchestrator follows around an isolated agent call:

```bash
# 1. Pre-call snapshot
moai worktree snapshot --agent-name <agent> > /tmp/snap.json
SNAP_PATH=$(jq -r .path /tmp/snap.json)

# 2. Run the isolated agent
# (Claude Code runtime: Agent(isolation: "worktree", ...))
# Capture the agent response JSON if available; many agents do not expose this.

# 3. Post-call verify
moai worktree verify \
  --snapshot "$SNAP_PATH" \
  --agent-response /tmp/agent-response.json \
  --agent-name <agent>

# Exit code:
#   0 = clean, no orchestrator action
#   1 = divergence detected (read the JSON report and AskUser restore/accept/abort)
#   2 = suspect (empty worktreePath; warn and consider blocking subsequent push)
#   3 = both
```

Restore (only after user selects "restore" via AskUserQuestion):

```bash
moai worktree restore --snapshot "$SNAP_PATH"
# Untracked files are listed but NOT recreated (manual restoration required).
```

The JSON report on stdout from `verify` includes:

- `snapshot_id` (matches the pre-call snapshot)
- `divergence` block (HeadChanged, BranchChanged, UntrackedAdded, UntrackedRemoved, PorcelainDelta)
- `suspect_flag` (when present)
- `report_path` and `json_sidecar` (the markdown + JSON divergence logs)
- `exit_code` (mirror of the process exit code for convenience)

## Out of Wave 5 Scope

The following items are deliberately deferred (see strategy-wave5.md § 14):

- Orchestrator-side wiring inside `Skill("moai")` workflows (`/moai run`,
  `/moai sync`) — Wave 5 ships primitives + this rule only.
- Untracked file content snapshot — paths-only restoration; users must
  recreate untracked content from external sources.
- Configurable divergence threshold — binary detection only.
- Direct AskUserQuestion from Go — the CLI surfaces exit codes; the
  orchestrator owns user interaction.
- Concurrency safety for parallel `moai worktree snapshot` invocations —
  sequential orchestrator invocation is assumed.

## Anti-Patterns

- **Skipping snapshot before high-risk agents**: If an agent has historically
  caused divergence, snapshot is mandatory regardless of perceived overhead.
- **Reading exit code without parsing JSON**: The orchestrator should always
  parse the verify JSON to know whether divergence is HEAD/branch/untracked
  and present that detail to the user.
- **Calling AskUserQuestion from inside a subagent**: Forbidden by
  agent-common-protocol; subagents return blocker reports and the orchestrator
  runs the AskUser round.
- **Force-restoring without user confirmation**: `moai worktree restore` is
  destructive against tracked files; always present `AskUserQuestion(restore /
  accept / abort)` first.

---

Version: 1.0.0
Source: SPEC-V3R3-CI-AUTONOMY-001 Wave 5 (W5-T08)
