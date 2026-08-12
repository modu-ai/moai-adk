---
title: Factory Mode
weight: 30
draft: false
---

## What is Factory Mode?

Factory Mode lets one **lead** session drive a `plan -> run -> verify -> sync`
chain while four **companion** sessions join the same run to parallelize the
work. Every session in the run — lead or companion — gets the raised
Stop-hook block cap so a mid-session goal keeps running past the default
consecutive-block ceiling.

The lead seeds the chain; companions do not. Each companion carries a
factory-membership flag (`-f`) plus a role label (`--name <role>-<run-id>`)
so the dispatcher classifies it correctly and the SessionStart hook
announces its membership.

## Entry switches

### Lead entry

```bash
moai cc -f                     # lead on Claude backend
moai cc -f SPEC-AUTH-001       # lead tied to a SPEC
moai glm -f                    # lead on GLM backend
```

The lead session:

{{< icon check-circle ok >}} Sets `MOAI_FACTORY` + `MOAI_FACTORY_ID` (chain seed).
{{< icon check-circle ok >}} Prints the run id and four companion launch lines at SessionStart.
{{< icon x-circle danger >}} Does NOT set `MOAI_FACTORY_LABEL` (that is the companion signal).

### Companion entry

```bash
moai cc -f --name plan-abc123    # plan companion
moai cc -f --name run-abc123     # run companion
moai cc -f --name review-abc123  # review companion
moai cc -f --name sync-abc123    # sync companion
moai glm -f --name run-abc123    # same, on the GLM backend
```

The `<run-id>` is the identifier the lead announced at startup; the four
roles are `plan`, `run`, `review`, `sync`.

The companion session:

{{< icon check-circle ok >}} Sets `MOAI_FACTORY_LABEL` (membership + role label).
{{< icon check-circle ok >}} Gets the same raised Stop-hook block cap as the lead.
{{< icon x-circle danger >}} Does NOT set `MOAI_FACTORY` — it does not seed a chain.

### No-op (unchanged session)

```bash
moai cc --name mysession         # no -f, no factory membership
moai cc --name run-abc123        # companion-shape --name but no -f → no-op
```

Without `-f`, the launcher is a no-op regardless of `--name` shape. The
`--name` flag passes through to Claude untouched.

## Multi-session bootstrap flow

```
Terminal 1 (lead)          Terminal 2-5 (companions)
─────────────────          ────────────────────────
moai cc -f                 moai cc -f --name plan-<run-id>
                           moai cc -f --name run-<run-id>
                           moai cc -f --name review-<run-id>
                           moai cc -f --name sync-<run-id>
```

Bootstrap is manual: a session cannot launch another session. The lead
SessionStart notice prints the exact four commands to copy, one per new
terminal. Substitute `moai glm` for `moai cc` on any companion to run it on
the GLM backend.

## Cross-session messaging

Inter-session communication uses Claude Code's cross-session messaging
(`ListAgents` / `SendMessage`). The `crossSessionInbound` settings field
controls whether an inbound message is accepted, held, or refused.

Factory Mode auto-accepts inbound messages: the launcher writes a transient
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

The lead notice announces the run id, the four companion launch lines, the
leader socket path, and the inbound-automation status. A companion notice
is a single role-less line acknowledging the join. Neither notice prompts;
both are informational stdout only.
