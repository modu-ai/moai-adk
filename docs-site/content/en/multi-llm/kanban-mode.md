---
title: Kanban Mode
weight: 30
draft: false
---

{{< callout type="info" >}}
For the full overview of Kanban Mode and the Origin-Trail Chain design direction, see [Kanban Mode](/en/advanced/kanban-mode). This page covers the multi-session (lead + companion) operating procedure.
{{< /callout >}}

## What is Kanban Mode?

Kanban Mode lets one **lead** session drive a `plan -> run -> sync`
chain while three **companion** sessions join the same run to parallelize the
work. The review verdict is not a separate stage — the sync gate absorbs it.
Every session in the run — lead or companion — gets the raised
Stop-hook block cap so a mid-session goal keeps running past the default
consecutive-block ceiling.

The lead seeds the chain; companions do not. Each companion carries a
kanban-membership flag (`-k`) plus a role label (`--name <role>`)
so the dispatcher classifies it correctly and the SessionStart hook
announces its membership.

## Entry switches

### Lead entry

```bash
moai cc -k                     # lead on Claude backend
moai cc -k SPEC-AUTH-001       # lead tied to a SPEC
moai glm -k                    # lead on GLM backend
```

The lead session:

{{< icon check-circle ok >}} Sets `MOAI_KANBAN` + `MOAI_KANBAN_ID` (chain seed).
{{< icon check-circle ok >}} Prints the run id and three companion launch lines at SessionStart.
{{< icon x-circle danger >}} Does NOT set `MOAI_KANBAN_LABEL` (that is the companion signal).

### Companion entry

```bash
moai cc -k --name plan    # plan companion
moai cc -k --name run     # run companion
moai cc -k --name sync    # sync companion
moai glm -k --name run    # same, on the GLM backend
```

Companions are named by their bare role; the three roles are `plan`, `run`,
and `sync`. The `<run-id>` names the lead session alone and never appears in
a companion name — a second live session claiming the same role takes the
next free number.

The companion session:

{{< icon check-circle ok >}} Sets `MOAI_KANBAN_LABEL` (membership + role label).
{{< icon check-circle ok >}} Gets the same raised Stop-hook block cap as the lead.
{{< icon x-circle danger >}} Does NOT set `MOAI_KANBAN` — it does not seed a chain.

### No-op (unchanged session)

```bash
moai cc --name mysession         # no -k, no kanban membership
moai cc --name run               # companion role name but no -k → no-op
```

Without `-k`, the launcher is a no-op regardless of `--name` shape. The
`--name` flag passes through to Claude untouched.

## Multi-session bootstrap flow

```
Terminal 1 (lead)          Terminal 2-4 (companions)
─────────────────          ────────────────────────
moai cc -k                 moai cc -k --name plan
                           moai cc -k --name run
                           moai cc -k --name sync
```

Bootstrap is manual: a session cannot launch another session. The lead
SessionStart notice prints the exact three commands to copy, one per new
terminal. Substitute `moai glm` for `moai cc` on any companion to run it on
the GLM backend.

## Cross-session messaging

Inter-session communication uses Claude Code's cross-session messaging
(`ListAgents` / `SendMessage`). The `crossSessionInbound` settings field
controls whether an inbound message is accepted, held, or refused.

### Availability constraints

Cross-session messaging is not available in every environment. Kanban Mode
connects the lead and companions through this channel alone, so where the
channel is absent the mode itself cannot form. Check these constraints
before you start.

{{< icon warning warn >}} **Operating system**: available on macOS and Linux (including Linux
inside WSL 2) only. Claude Code does not provide cross-session messaging
on native Windows.
{{< icon warning warn >}} **Providers**: unavailable on Amazon Bedrock, Claude Platform on AWS,
Agent Platform on Google Cloud, and Microsoft Foundry.
{{< icon warning warn >}} **Versions**: requires Claude Code v2.1.224 or later. Initiating a
cross-machine conversation requires v2.1.225+, and @mentions plus the
/config rows require v2.1.232+.
{{< icon warning warn >}} **Flags**: any one of `CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC`,
`DISABLE_TELEMETRY`, `DO_NOT_TRACK`, or `DISABLE_GROWTHBOOK` disables
feature-flag evaluation, which silently turns messaging off.

Quick diagnosis: if the `/list-agents` command is recognized, the feature
is present; if it is not recognized, it is absent.

Kanban Mode auto-accepts inbound messages: the launcher writes a transient
settings file carrying `{"crossSessionInbound": "accept"}` and passes it to
the backend via `--settings`. The file is session-private (cleaned up on
exit) and does not mutate your persistent settings.

### Operator-supplied `--settings`

If you pass `--settings <file>` on the command line, the launcher does NOT
inject its own settings file. Verify your file carries:

```json
{
  "crossSessionInbound": "accept"
}
```

The lead SessionStart notice prints an advisory reminding you to check when
the launcher did not inject.

## SessionStart notice

The lead notice announces the run id, the three companion launch lines, the
leader socket path, and the inbound-automation status. A companion notice
is a single role-less line acknowledging the join. Neither notice prompts;
both are informational stdout only.
