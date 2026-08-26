# Proposal — Rule Amendment (Mechanism Pointer) + Audit Correlation Recipe

> SPEC-TEAMMATE-REVIVAL-GUARD-001 · M3 deliverable · run-phase.
> Status: PROPOSAL TEXT ONLY — nothing in this file has been applied to `.claude/rules/**`
> or `.moai/docs/`; the orchestrator routes both at sync through the proper owners.
> The proposed amendment body (§A.1) is written to be neutral (no SPEC IDs, SHAs, or
> internal dates) so the rules owner can adopt it verbatim.

## §A Doctrine-pointer amendment (for `.claude/rules/moai/workflow/cross-session-messaging.md`)

Target: the "Never address a stopped teammate by name" [ZONE:Evolvable] [HARD] clause —
append the paragraph below at the end of the clause body. Optional companion: extend the
"Reviving a stopped teammate" anti-pattern bullet with "(now also mechanically denied —
see the mechanism layer note above)".

### A.1 Proposed addition (neutral text, adopt verbatim)

> **Mechanism layer (landed).** The prohibition above is mechanically enforced at the
> PreToolUse layer: a SendMessage whose recipient — bare name, agent id, or the
> `name [ref]` form (suffix parsed and stripped) — matches a live entry in the same
> session's stop registry (`.moai/state/agent-stops/<session-id>.json`, populated by the
> matcher-less PostToolUse dispatch on every TaskStop completion) is refused with a
> sentinel-prefixed deny (`STOPPED_TEAMMATE_VIOLATION:`). Deliberate revival stays
> possible through an explicit fresh spawn carrying the same name — the entry clears
> before the spawn proceeds; session end removes the session's entries. Every stop,
> send, deny, and clear appends one JSONL row to `.moai/logs/agent-stop-audit.jsonl`.
> The deny layer is opt-in (`workflow.agent_stop_guard.enabled` — template default
> false; this repository's local config enables it for dogfood). The flag is a kill
> switch for the DENY only: recording and audit continue regardless of its state. A
> `STOPPED_TEAMMATE_VIOLATION` deny is never a bug to route around — it means the
> doctrine held: route coordination through the owning orchestrator, or respawn the
> name deliberately.

### A.2 Sibling surface (informational)

`.claude/rules/moai/core/agent-common-protocol.md` and the doctrine SPEC's own clause
already carry the discipline text; only the cross-session-messaging clause needs the
mechanism pointer. No other rule file requires an edit for the pointer.

## §B Audit → commit-window correlation recipe (`.moai/docs/` candidate)

Purpose: when a stopped teammate is found to have run again (the revival incident
shape), bound the revival window from the stop-guard audit trail and attribute the
commits inside it post hoc. This is the visibility deliverable (direction 3) of the
mechanism: rows exist so the window is provable after the fact.

1. **Bound the window.**
   `grep '"name":"<teammate>"' .moai/logs/agent-stop-audit.jsonl`
   Window = `stop_recorded` (row N) → the earliest of `respawn_cleared`,
   `session_cleared`, or the next `stop_recorded` for that name. Each row carries a
   UTC RFC3339 timestamp, session_id, name, agent_id, and decision — enough to key the
   window to one session.
2. **Attribute commits inside the window.**
   `git log --since='<window start UTC>' --until='<window end UTC>' --format='%h %cI %an%n%b'`
   Match author/committer identity to the teammate where possible. Where the revived
   agent committed under a shared identity, the audit row's session_id plus the window
   boundaries are the attribution evidence.
3. **Classify.** Commits inside a stop→window-end span with no intervening
   `respawn_cleared` are revival-window output — re-verify them independently. The
   originating incident precedent: the revived output itself passed re-verification;
   the defect was ownership (an unowned writer), not artifact quality. The risk the
   next revival carries is that it passes WITHOUT re-verification.
4. **When the deny layer is enabled**, step 2 usually returns an empty result for the
   window — the send was refused. An empty result is the guard working; the
   `send_denied` row is the record that the attempt was made and when.

**Worked example (measured in this card's worktree, live rows).** Session
`703df7e1-…` stopped `t267-ep1-probe2` at `2026-08-26T17:12:48Z`; two sends were
denied (`17:12:53Z`, `17:13:10Z` — the second teammate-issued, the incident vector);
a same-name spawn cleared the entry at `17:13:21Z`;
`git log --since='2026-08-26T17:12:48Z' --until='2026-08-26T17:13:21Z'` → 0 commits:
no revival contamination; the deny held for the whole window.

**Operational note.** A registry file whose session ended before the SessionEnd
cleanup existed (or on a binary predating it) is inert: per-session scoping means no
other session's sends consult it. Such files are manual-cleanup garbage, not a
correctness hazard.

## §C Default-flip verdict (recorded; NOT executed)

Template default stays **false**. Measured today: two live probes (a recording-only
probe and a full deny/respawn recipe), zero false-positive denies, zero wedged
legitimate sends. That is probe-scale, not the sustained dogfood the flip trigger
demands. Flip is a named upgrade trigger, not an intent: revisit only after
(1) a sustained, multi-hour, multi-teammate session with the gate enabled,
(2) zero false-positive denies in its audit trail, and
(3) zero reports of wedged legitimate sends.
Record the flip against that evidence when it exists.
