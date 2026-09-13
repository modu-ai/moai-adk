### Pre-Edit Sync Check (Direct-Edit Race Mitigation)

[ZONE:Evolvable] [HARD] Direct main-session edits to shared working-tree paths (Edit/Write/Bash — any direct edit) bypass the spawn gate above, so the orchestrator MUST run the parallel-session detection **before a non-trivial direct edit** to shared paths. (Incident record + enforcement-placement assessment: `agent-common-protocol-reference.md` § Pre-Edit Sync Check — rationale and enforcement record.)

#### The rule, at the moment of the edit

**TRIGGER** — the gate fires when ALL three hold:

| Condition | Test |
|---|---|
| Tool | an `Edit`, `Write`, or file-mutating `Bash` call |
| Target | a shared path another session could also mutate: `.claude/`, `.moai/`, `internal/`, `pkg/`, `cmd/`, or repo-root config files |
| Location | CWD is the primary checkout. Exempt: an already-isolated worktree, `/tmp`, or a session-private scratch dir |

**CHECK** — before the FIRST triggered edit of a task, as one parallel batch:
```bash
# 1. live foreign sessions (own session filtered out; then liveness-probe each PID)
moai session list --json | jq '[.[] | select(.cwd == "<project-root>" and .session_id != "<own>")] | length'
# 2. divergence vs origin/main
git fetch origin main 2>&1; git rev-list --count --left-right origin/main...HEAD
```

**DECIDE and ACT** — no outcome permits "proceed in the shared checkout anyway":

| Probe result | Required action |
|---|---|
| 0 live foreign sessions AND `0 0` / `0 N` | Proceed in the shared checkout |
| ≥1 live foreign session | **ISOLATE before editing**: `moai cc -w <name>` / `EnterWorktree(<path>)` / `Agent(isolation: "worktree")`. If isolation is impossible, surface via `AskUserQuestion` (isolate / wait / abort) |
| `N 0` / `N M` divergence | STOP; `AskUserQuestion` (rebase / inspect / abort) per the Pre-Spawn Sync Check matrix |

> **Stale-registry caveat**: registry entries can hold dead PIDs. Probe each foreign entry's liveness with `kill -0 <pid>`; ignore confirmed-dead, treat indeterminate as live and isolate anyway. ANY live-or-indeterminate foreign entry ⇒ isolate (`worktree-integration.md` § Parallel-Session Branch Conflict Auto-Isolation).

**RE-CHECK** — the probe decays. Re-run it before ANY commit in the shared checkout, and after any long pause in the task.

#### The sweep prohibition
