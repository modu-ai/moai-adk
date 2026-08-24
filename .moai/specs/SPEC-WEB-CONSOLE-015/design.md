# SPEC-WEB-CONSOLE-015 — Design (reference material)

> **Not part of the required artifact set.** This SPEC was reclassified from Tier L to Tier M in
> version 0.3.0, and Tier M carries three artifacts (spec / plan / acceptance). This document is
> retained rather than discarded because it is current and its content is not duplicated elsewhere
> — it is reference material for a reader who wants the decisions behind the requirements, and it
> is not audited as part of the Tier M artifact set.

Deliberately thin. This document carries only what spec.md and plan.md do not: the decisions this
SPEC still owns after the three-way carve-out, and their consequences.

## §1 Where version 0.1.0's decisions went

Version 0.1.0 §G resolved six cross-cutting decisions. Four of them now belong to sibling SPECs
and are **not** restated here; a reader tracing one should follow the owner rather than this file.

| Decision (0.1.0 label) | Subject | Owner now |
|---|---|---|
| G-1 | card identifier derived from the worktree basename, with an override | `SPEC-KANBAN-RECORD-SESSION-KEY-001` |
| G-2 | whether the context-usage move stays in this SPEC | moot — it moved out entirely |
| G-3 | hard cut for the context-usage path, no dual-write window | `SPEC-SESSION-TELEMETRY-001` |
| G-4 | `/todo` as its own route rather than a panel | `SPEC-WEB-TODO-QUEUE-001` |
| G-5 | the todo section lists all three states with a badge | `SPEC-WEB-TODO-QUEUE-001` |
| G-6 | the lane number's Go type and its zero value | `SPEC-KANBAN-RECORD-SESSION-KEY-001` |

Three decisions remain, all of them properties of the console.

## §2 C-1 — lanes sit beside the chain, not inside it

**Decision.** The lane collection is a second collection in the kanban view model, iterated
separately. `ChainRoles` (`internal/web/viewmodel_ops.go:46`) is not widened.

**Why.** The four chain roles are a fixed dispatch vocabulary — lead, plan, run, sync — and the
view's arrow-joined chain rendering encodes that fixed length (`screens.templ` draws an arrow
between consecutive roles). Widening the list to a variable-length set would make every chain
consumer, present and future, defend against a role list whose length it cannot predict, to
express a relationship that does not exist: a lane is not a stage of the chain, it is a parallel
worker carrying a whole card.

**Consequence.** The lane section is additive markup. Nothing about the existing chain rendering
changes, which is what keeps M1 shippable while the dependency SPECs are still in flight.

## §3 C-2 — a duplicate process identifier resolves to nobody

**Decision.** When two or more registered lanes carry the same process identifier, the resolved
session's record is attributed to **none** of them, and each affected lane renders the unresolved
marker (REQ-WC15-047).

**Why this shape, and why it is a requirement rather than a note.** The hazard is reachable at
render time, not hypothetical: `LoadFactoryRegistry` (`internal/kanban/factory_slots.go:55`) is
fail-open, and `PruneFactoryDeadClaims` (`:84`) is a separate call the console is under no
obligation to make — so a stale entry whose PID has been recycled onto a live lane is exactly the
state a read-only console can find. Version 0.1.0 carried this rule as an imperative sentence in a
non-normative edge-case list, with no requirement and no criterion; the same document had already
promoted a different edge case to a criterion for precisely this reason, so the treatment was
inconsistent with its own precedent.

**Rejected alternatives.**

- *Attribute to the lower-numbered lane.* Picks a winner with no evidence for the choice, and
  renders a confident wrong card identifier on an audit board — the failure mode the "not
  recorded" marker exists to avoid.
- *Prune dead claims first, then join.* Pruning is a producer act: it decides an entry is dead and
  rewrites the registry. A read-only console must not perform it (REQ-WC15-002), and reading the
  pruned result without writing it back would make the console's view diverge from the file every
  other consumer reads.
- *Accept mis-attribution and record it as a known limitation.* Available, and explicitly the
  alternative the audit offered. Rejected because a silently wrong card label is worse than an
  absent one, and the correct behaviour costs a counting pass.

**Consequence for the implementation.** The lookup cannot be a `map[pid]session` built by
assignment — that is last-write-wins and silently produces a winner. Count occurrences first, then
resolve only the identifiers whose count is one.

## §4 C-3 — the join gets its process identifier by widening the session view model

**Decision.** `SessionVM` gains the process identifier; `loadSessions`
(`internal/web/viewmodel_ops.go:409-435`) stops dropping it. The lane builder consumes the
already-loaded session collection rather than re-reading the registry file.

**Why.** `loadSessions` already reads `active-sessions.json` once per render and already unmarshals
into `session.Entry`, which carries `PID`; it simply does not copy the field into the view model.
A second reader in the lane builder would read the same file twice per render and could observe a
different state than the chain rows rendered from — two sections of one page disagreeing about
which sessions are live.

**Consequence.** `session.Entry` is untouched (spec.md §C.1 — it is a frozen schema, read only).
The widening is confined to a console-internal view-model type.

## §5 C-4 — what the note banner says after the change

**Decision.** The banner keeps its estimation disclosure and drops both false clauses.

Today (`internal/web/screens.templ:192`) it reads, hard-coded and untranslated:

> Stage is estimated from heartbeat. Model, effort and context usage are not recorded yet, so they
> are left blank — they fill in once kanban.Record is extended.

The first sentence stays true and is load-bearing: it is the prose form of the honesty flag
REQ-WC15-045 requires. The second is falsified twice over — the values become recorded once
`SPEC-SESSION-TELEMETRY-001` lands, and the kanban record was never their producer (spec.md §A.2).

**Consequence.** The rewritten banner is a user-visible string in a view whose every other string
is translated, so it takes a translation key and four locale entries (REQ-WC15-050, REQ-WC15-052).
`noteBanner` (`internal/web/widgets.templ:40-52`) already branches on a non-empty key and emits
`data-i18n` only then, so passing a key is the whole mechanism — no widget change is needed.
