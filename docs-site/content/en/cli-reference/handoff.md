---
title: moai handoff Handoff Records
weight: 68
draft: false
---

`moai handoff` manages the auto-resume handoff pending record. It saves or clears the paste-ready resume body used to continue work across a session boundary (`/clear`). When `handoff.mode: auto` is set, the saved record is auto-injected at the next session start.

## Subcommands

| Command | Description |
|--------|------|
| `moai handoff save` | Save the paste-ready resume body as a pending record |
| `moai handoff clear` | Remove the pending record |

Both accept the common flag `--project-dir <path>` (project root, default: current directory).

## moai handoff save

```bash
moai handoff save --stdin --spec SPEC-AUTH-001 --phase run < resume.txt
```

| Flag | Description |
|--------|------|
| `--body <text>` | Resume body (verbatim 6-block paste-ready) |
| `--stdin` | Read the body from stdin instead of `--body` |
| `--spec <id>` | The SPEC id this handoff resumes |
| `--phase <plan\|run\|sync>` | Phase |
| `--session <uuid>` | saved_by_session uuid (attribution) |
| `--lang <lang>` | conversation_language snapshot |
| `--ultrathink` | Record the ultrathink directive (for restore guidance) |
| `--ultracode` | Record the ultracode directive (for restore guidance) |
| `--goal <condition>` | Record the `/goal` condition (for restore guidance) |

## moai handoff clear

```bash
moai handoff clear
```

Removes the pending handoff record.

## Fail-open guarantee

Even if the `moai` CLI is not on PATH or `moai handoff save` exits non-zero, the orchestrator's paste-ready output is preserved unchanged. A save failure never blocks handoff emission, and the manual paste path works fully without the save — the save is merely an additive persistence step, not a gate.

## Related documents

- [Autonomous Continuation Loops](/en/advanced/autonomous-loops)
- [moai goal](/en/cli-reference/goal)
- [CLI Overview](/en/getting-started/cli)
