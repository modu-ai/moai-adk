---
id: SPEC-GOAL-STOPFAILURE-CLEAR-001
title: "Disarm an armed goal when a turn dies on an unrecoverable API error, so the loop stops spinning idle turns"
version: "0.1.0"
status: implemented
created: 2026-08-19
updated: 2026-08-20
author: manager-spec
priority: P2
phase: "v3.1.1"
module: "internal/hook, internal/goal"
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "goal, stop-failure, hook, unrecoverable-error, disarm"
related_specs: "SPEC-INFINITE-GOAL-001, SPEC-GOAL-HTML-WIRING-001"
---

## Problem

`/moai goal` is a MoAI-owned reimplementation of native `/goal` semantics, with its own state file (`.moai/state/goal/<session-id>.json`) and its own Stop-hook evaluator. Because it is a reimplementation rather than a wrapper, it inherits nothing from the native command — including the self-clear that Claude Code 2.1.234 added for a turn that dies on an unrecoverable error.

The consequence is stated by the goal doctrine itself: an armed goal with nothing running "spins idle turns until the ceiling, because each turn-end finds the condition unmet and no work advancing it" (`goal-directive.md` § Goal-Presentation Timing). A revoked credential or an exhausted credit balance ends the turn that was doing the work; the goal survives it, still armed.

Premise verified before authoring, in this worktree at `e7aeec088`:

```
$ grep -rn "unrecoverable\|revoked\|overflow\|credit" internal/goal/ ; echo exit=$?
exit=1                      # no match

$ grep -rn "checkin\|CheckIn\|CHECKIN" internal/goal/ ; echo exit=$?
exit=1                      # no match
```

## Detection feasibility — the gating question, answered

The evaluator cannot see the failure. `moai hook stop-goal` runs on the `Stop` event, and a turn that dies on an API error does not end through `Stop`.

A different event does see it. `StopFailure` (Claude Code v2.1.78+) fires exactly when a turn ends on an API error, and carries `error_type` and `error_message` (`internal/hook/types.go` HookInput). MoAI already registers a handler for it (`internal/hook/stop_failure.go`, wired at `internal/cli/deps.go` and in `settings.json`), and that handler already receives `session_id` — the key the goal state file is named by.

So detection is possible, and the disarm belongs on `StopFailure`, not in the goal evaluator.

**Partial, and the gap is named.** The documented `error_type` enum is `rate_limit, overloaded, authentication_failed, oauth_org_not_allowed, billing_error, invalid_request, model_not_found, server_error, max_output_tokens, unknown` (`hooks-system.md`). Two of the three cases the upstream note names map onto it — revoked auth (`authentication_failed`, and the org-policy sibling `oauth_org_not_allowed`) and an exhausted credit balance (`billing_error`). **Context overflow does not appear in the enum at all**, so this SPEC cannot cover it, and does not claim to.

## Requirements

- **REQ-GSF-001** — While a `StopFailure` event carries an unrecoverable `error_type`, the armed goal for that session SHALL be disarmed.
- **REQ-GSF-002** — Unrecoverable means exactly `authentication_failed`, `oauth_org_not_allowed`, `billing_error`. Every other value — including `rate_limit`, `overloaded`, `server_error`, `max_output_tokens`, and `unknown` — SHALL leave the goal armed.
- **REQ-GSF-003** — The disarm SHALL be reported in the handler's `systemMessage`, so a goal that disappears is visible rather than silent.
- **REQ-GSF-004** — The handler SHALL remain non-blocking and fail-open: an unreadable state directory, an absent goal, or a failed remove SHALL degrade to the existing error-class message, never to an error return.
- **REQ-GSF-005** — `oauth_org_not_allowed` SHALL be added to the `StopFailure` matcher in `settings.json` and its template, since a matcher that excludes it means the handler is never invoked for it.

### Why rate_limit must not disarm

`rate_limit`, `overloaded`, and `server_error` are transient: the work resumes on retry, and the goal is exactly the state that should survive to see it. `max_output_tokens` is classified withheld-recoverable by `runtime-recovery-doctrine.md` §1, which obliges the agent to recover rather than to treat the turn as failed. Disarming on any of these would destroy live state on a condition that resolves itself — the opposite failure from the one this SPEC fixes, and a harder one to notice.

## Relation to the existing bounds

The goal loop already has three exits, and this is a fourth that does not overlap them:

| Bound | Fires while | Blind to |
|---|---|---|
| Turn ceiling (default 30) | turns keep completing | a turn that never completes |
| Runtime consecutive-block cap (default 8) | turns keep completing | same |
| Stagnation guard (N no-progress turns) | turns keep completing | same |
| **This SPEC** | a turn dies on an unrecoverable error | everything the three above catch |

All three existing bounds are counters over completed turns. They are the reason an armed goal eventually stops, and the reason it stops *late* — they need turns to burn before they fire, and the turns they burn are the idle ones this SPEC prevents. None of their semantics change here.

## Scope

### Out of Scope — adjacent mechanisms and the existing bounds

- The 30-minute background-task check-in (the other 2.1.234 behavior). Deliberately not bundled: it is an unrelated mechanism on an unrelated trigger, and the card's own instruction is to route it to a separate card rather than grow this one.
- Context-overflow disarm, which the `error_type` enum cannot express (see above).
- Any change to `/moai goal`'s three existing bounds.
