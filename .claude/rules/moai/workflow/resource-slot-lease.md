---
description: "Resource slot lease: the moai slot verbs, the declared bound, and the opt-in PreToolUse guard"
paths: "**/.moai/config/sections/workflow.yaml,**/resource-slot-lease.md"
---

# Resource Slot Lease

A lease on a NAMED resource, so two sessions sharing one machine do not start the same heavy work at once. Checking that a resource looks free and then starting is not mutual exclusion: both sessions can observe "free" between the check and the start. An acquire decides and writes inside one cross-process critical section, so exactly one of any set of concurrent acquires is admitted.

## The three verbs

```bash
moai slot acquire --resource <name> [--name <lane>] [--command <text>] [--max-duration 20m] [--force] [--json]
moai slot status  [--resource <name>] [--json]
moai slot release --resource <name> [--force] [--json]
```

- The record lives in the PRIMARY checkout's `.moai/state/slot-leases/<resource>.json`, so every linked worktree sees the same lease.
- A resource name is 1-64 characters of `a-z`, `0-9` and `-`. It becomes part of a path, so anything else is refused and nothing is written.
- The recorded pid is the OWNING SESSION's, never the pid of the process that ran the verb — that one is gone the moment the command returns. An unresolvable owner is recorded as 0 and reads as live.
- Exit codes: `0` success, `3` held by another session, `4` busy (a peer was mid-write; retry), `1` anything else. Held and busy are different facts and never stand in for each other.

## The declared bound is a promise, not a timeout

`--max-duration` (default `workflow.slot_lease.default_max_duration`) is the holder's own declaration. Past it the lease reads EXPIRED and another session may take it over without `--force`, even while the owner is alive; the holder re-acquiring restarts the bound. A lease whose owning session is gone reads STALE and is taken over the same way. `--force` takes over a live, unexpired holder, and every takeover names the displaced holder in the new record and in the audit log — it is never silent.

The asymmetry is deliberate: a lease held open by a forgotten release costs everyone waiting, while a takeover costs one overlap. A resource that must never overlap needs a bound long enough for its slowest honest run.

## The guard is opt-in

```yaml
workflow:
    slot_lease:
        enabled: false          # the guard only; the verbs work either way
        default_max_duration: 30m
        resources:
            <resource-name>:
                commands:
                    - '<command-regex>'
```

While `enabled` is true, a PreToolUse guard refuses a shell command that matches a resource's patterns when that resource's holder is a different, live, unexpired session. It reads nothing at all while `enabled` is false or unreadable.

Patterns are RE2, matched against the command with quoted spans removed, and they come only from this file — no command list ships, because any list would name one language's tools.

The guard FAILS OPEN. An unreadable record, an unknown caller, a project root it cannot resolve, a pattern that does not compile, or a resource entry that is not a list of strings all ALLOW the command, write a `[moai:slot-lease] advisory:` line, and record why in `.moai/logs/slot-lease-audit.jsonl`. It also allows when the holder is the calling session, when the holder is gone, when the bound has elapsed, and when nobody holds the resource — each with its own audit line. A deny needs positive evidence, and its reason starts with `SLOT_LEASE_VIOLATION:`.

## What it does not do

- It does not serialize work inside ONE session: a session's own subagents read as that session and all pass.
- It does not require a lease. Two sessions that both skip `acquire` still collide; the guard only refuses a command while someone else's lease is live.
- It is independent of the release-integration window (`moai integration`): separate record, separate lock, separate config key.
