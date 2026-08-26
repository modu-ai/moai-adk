# Proposal — Rule Amendment (Mechanism Pointer)

> SPEC-TEAMMATE-REVIVAL-GUARD-001 · M3 deliverable · run-phase.
> Status: PROPOSAL TEXT ONLY — nothing in this file has been applied to `.claude/rules/**`;
> the orchestrator routes the amendment at sync through the rules owner.
> The proposed amendment body (§A.1) is written to be neutral (no SPEC IDs, SHAs, or
> internal dates) so the rules owner can adopt it verbatim.
> Sibling deliverable: `audit-commit-correlation-recipe.md` (same directory) — the
> direction-3 visibility recipe.

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

## §B Default-flip verdict (recorded; NOT executed)

Template default stays **false**. Measured today: two live probes (a recording-only
probe and a full deny/respawn recipe), zero false-positive denies, zero wedged
legitimate sends. That is probe-scale, not the sustained dogfood the flip trigger
demands. Flip is a named upgrade trigger, not an intent: revisit only after
(1) a sustained, multi-hour, multi-teammate session with the gate enabled,
(2) zero false-positive denies in its audit trail, and
(3) zero reports of wedged legitimate sends.
Record the flip against that evidence when it exists.
