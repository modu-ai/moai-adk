---
description: Board-and-lens mechanics companion to the kanban dispatch protocol
paths: "**/kanban-dispatch*.md,**/.claude/agents/moai/manager-lead.md,**/.moai/state/integration/**"
---

# Kanban Dispatch — Board and Lens Mechanics

> Owns: the board mechanics, dispatch-cycle walkthrough, review-lens pointer, heavy-run lease, Factory Mode mechanics, the remaining boundaries and cross-references. `kanban-dispatch.md` is the SSOT and keeps every binding clause and its scope declaration; sibling companion `kanban-dispatch-detail.md` keeps the long tables and narratives.

## The board

Five columns, fixed and ordered: `backlog → plan → run → sync → done`. `backlog` and `done` have no owning session; the three working columns each map to exactly one companion role (`plan` / `run` / `sync`), which is what makes dispatch a lookup rather than a decision. There is no `review` column — the sync gate absorbs the review verdict and runs the lenses itself. Column-by-role table: `kanban-dispatch-detail.md` § The board.

## Review lens selection

`review` is not one thing. The factory leader picks the lenses from what the card actually changed and states the choice in the `sync` dispatch, so the gate runs that review rather than re-deriving it. Lens table: `kanban-dispatch-detail.md` § Review lens selection.

`--deep --patch` is opt-in twice over: `--patch` drafts a fix and is absent unless the operator asked for it. Do not add it on the leader's own initiative.

### Serializing a heavy run across lanes

Where two lanes would otherwise start the same heavy verification at once, a lane takes a lease on a named resource first: `moai slot acquire --resource <name> --max-duration <bound>`, and `moai slot release --resource <name>` when the run ends. `moai slot status` says who holds it. This is the same record-and-liveness shape as the integration window, kept deliberately separate from it — its own record, its own lock, its own config key — because a merge window and a heavy-run window are different resources and one window for both makes each wait on the other.

The bound is the holder's own declaration: past it another lane may take the resource over without `--force`, so a forgotten release costs the bound and not the batch. A PreToolUse guard that refuses a matching command while another live session holds the resource exists, and is **opt-in and off by default** (`workflow.slot_lease.enabled`); the verbs work whatever that flag says. Full surface: `.claude/rules/moai/workflow/resource-slot-lease.md`.

## The dispatch cycle walkthrough

`[operator picks a card] → plan → run → sync → [leader marks done]` — each arrow is one dispatch from the leader to one companion session, addressed by name (`SendMessage`; `ListAgents` lists live sessions and says when it could not check them all). Each instruction carries, at minimum: the card, the SPEC ID once one exists, the phase command, and the completion signal to write. Keep it a pointer, not a copy — the companion reads the SPEC artifacts itself.

**`sync → done` is the same act with the dispatch removed.** No session occupies `done`, so the leader reads the sync session's completion evidence and records the terminal transition itself. Session-naming rules and the address-block walkthrough: `kanban-dispatch-detail.md` § The dispatch cycle.

## Factory Mode mechanics

`moai cc -f <N>` launches one leader plus lane sessions labelled `lane-1..lane-N` (`lane-<n>` is the session label, and `-f lane` joins as the next free one). No per-column companions: the leader routes each card WHOLE to a free lane, or a lane receives the card by its own self-promotion of an already-queued card; either way the lane carries it `plan → run → sync` in-session — serial stages, each stage's execution spawned as sub-agents — and owns it end to end. A/B/C collapse into the lane (the class still names which ceremonies are skipped — `plan` for A and B — but no card changes sessions). Queue, evidence-reading, integration, and disposal rules are unchanged. Mechanics: `kanban-dispatch-detail.md` § Factory in-lane 3-stage.

## Boundaries

- **No board state store.** The queue is a plain file; column position is held by the leader within a card's run and re-derived from SPEC status after a clear. Persistent board state, per-card worktree lifecycle, WIP limits, and card/frontmatter reconciliation are separate work, not assumed here.
- **No session spawning.** The leader addresses sessions the operator launched — it never creates one (sub-agents are not sessions).
- **A role with no live session is a fault, not a wait.** The leader reports the empty role and the waiting card; silently holding it presents as a hang, the most expensive failure shape to diagnose. Empty means a complete listing showed none; an incomplete one is a gap, not an empty role (detail companion).

## Cross-references

- `.claude/rules/moai/core/agent-common-protocol.md` § Blocker Report Format — what a companion returns when it cannot proceed
- `.claude/rules/moai/workflow/worktree-integration.md` — the L1/L2 worktree tiers, their lifetimes, and the disposal contract
- `.claude/skills/moai/workflows/gtd.md` — the backlog queue surface
- `.claude/agents/moai/manager-lead.md` — the coordination agent the leader session works through
