# SPEC-CODEX-FACTORY-RETIRE-001 — design

## 1. Shape after retirement

```
moai codex [head] [-- tail]
  head scan (pre-verb, pre-`--`):
    -k / --kanban*            -> refuse KANBAN_MODE_UNSUPPORTED_BACKEND, rc 1
    -f / --factory* / --factory-run -> refuse FACTORY_MODE_UNSUPPORTED_BACKEND, rc 1
  -w strip -> verb lookup -> readout | launch
  launch:
    child env = parent env - {11 lane keys, CLAUDE_CODE_SESSION_ID, MOAI_SESSION_PID} + CODEX_HOME
    direct POSIX: chdir -> syscall.Exec            (no factory.db I/O)
    direct Windows: Start -> Wait -> exit code     (no factory.db I/O)
    spawn: tmux command with KEY= for 11 lane keys; pane identity probe only when -w anchors
```

The refusal is a scan, not a parser: it does not need to classify role or worker-label
values, only to recognize that an entry token is present. That removes the reason
`isFactoryRoleToken`, `isFactoryLaneShape`, and the kanban name parsing existed on the
codex door.

## 2. Why the environment is scrubbed rather than the registration guarded

Two options were weighed.

| Option | Effect | Rejected because |
|---|---|---|
| Keep the codex launch-pending call, guard it with "only when this process entered -f" | codex never self-registers | The Codex child still inherits the lane identity, and its own hooks (`--harness codex`) and its `moai mcp-server` (via `env_vars`) would attribute to the Claude lane's run |
| Scrub the eleven lane keys from every codex child (chosen) | No codex process can present a lane identity | — |

Scrubbing covers the three downstream readers at once: the launcher's own
registration, the Codex hooks, and the MCP server subprocess. It also makes D1
(keeping `env_vars`) safe: the forwarded keys are empty.

## 3. Join refusal placement

`enterSelectedFactoryRun` is the single point both `cc` and `glm` pass for worker joins
and for `--factory-run` leads. Reading `runs.lead_backend` there refuses both shapes
with one check. The refusal precedes `resolveFactoryWorkerName` (registry claim) and
`recordFactoryRunStart` (run-row upsert), so a stale codex run is neither joined nor
silently re-owned by a Claude lead.

## 4. What stays and why

| Kept | Reason |
|---|---|
| `factory_launch_pending.go`, `factory_run_owner.go` | cc/glm launch paths (`launch_exec_posix.go`, `launch_exec_windows.go`) |
| `factorymsg` handoff API + all tables | Claude hook per-prompt path calls `refuseWhileHandoffPending` / bind; schema stability |
| `hook/factory_handoff_bind.go` | inert without a producer; removal is a Claude-path change (out of scope) |
| `moai factory handoff abandon-lane`, `recover-resume`, `moai factory runs` | operator recovery for rows that already exist |
| `mcpServerEnvVarsValue` | existing-project drift (D1); F2 needs it |

## 5. Rebuild seam left for F2

After this SPEC, `-f` on `moai codex` is a refusal constant. F2 replaces the refusal for
the `-f agent` shape with the headless worker; nothing else in the codex launcher needs
to be restored.
