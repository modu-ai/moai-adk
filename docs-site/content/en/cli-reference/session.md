---
title: moai session Session Registry
weight: 65
draft: false
---

`moai session` manages the multi-session coordination registry in `.moai/state/active-sessions.json`. It is a tool for mitigating the races that arise when multiple Claude Code sessions work on the same project simultaneously.

## Subcommands

| Command | Description |
|--------|------|
| `moai session register <session_id> <spec_id> <phase>` | Register a new active session |
| `moai session heartbeat <session_id>` | Update an existing session's last_heartbeat (idempotent) |
| `moai session deregister <session_id>` | Remove a session (idempotent) |
| `moai session list` | List active sessions (filterable with `--filter-spec`) |
| `moai session purge` | Remove stale entries (default: more than 30 minutes since the last heartbeat) |
| `moai session current` | Print this orchestrator's session UUID |
| `moai session doctor` | Diagnose why the registry is empty |

Most subcommands support machine-readable output via the `--json` flag.

## moai session list

```bash
moai session list
moai session list --filter-spec SPEC-AUTH-001
```

| Flag | Description |
|--------|------|
| `--json` | Machine-readable JSON output (orchestrator pre-spawn check format) |
| `--filter-spec <id>` | Return only entries matching the given spec_id |

## moai session purge

```bash
moai session purge
```

| Flag | Description |
|--------|------|
| `--json` | JSON output |
| `--threshold-minutes <n>` | Stale-heartbeat cutoff in minutes (default 30) |

## moai session current

```bash
moai session current
```

Prints the orchestrator's own session UUID. If the runtime does not expose a session ID, it returns the canonical fallback string.

| Flag | Description |
|--------|------|
| `--json` | JSON output |
| `--show-fallback` | Print only the canonical fallback string (for paste-ready resume generation) |

## moai session doctor

```bash
moai session doctor
```

Diagnoses why the multi-session coordination registry is empty (write-path diagnostics).

| Flag | Description |
|--------|------|
| `--json` | JSON output |

## Usage context

This registry is used by the orchestrator to detect concurrent-session races before spawning implementation agents. If `moai session list --json --filter-spec <SPEC-ID>` returns entries from another session, the orchestrator halts and confirms with the user.

## Related documents

- [moai inventory](/en/cli-reference/inventory) — unified view of sessions, worktrees, and harnesses
- [CLI Overview](/en/getting-started/cli)
