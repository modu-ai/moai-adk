---
title: Factory Mode
weight: 6
draft: false
new: true
added_in: "v3.2"
---

{{< new-badge v3.2 >}}

# Factory Mode

{{< callout type="info" >}}
{{< icon flash primary >}} <strong>Value affiliation</strong>: multi-session orchestration · tokenomics
{{< /callout >}}

Factory Mode is the second form of Kanban Mode. Where kanban is a board on which three role companions carry one card between columns, the factory is an assembly line on which **N numbered lanes** carry several cards at once. A card does not hop between columns — it goes **whole** into one free lane, and that lane owns it end to end through `plan → run → sync`, in order, in-session.

The entry is the dedicated token `-f`: as the kanban chain uses `-k`, the factory uses `-f`. The queue, evidence reading, integration, and disposal rules are identical to kanban — the only thing that changes is the shape of how cards flow across the board.

## Where factory diverges from kanban

| | Kanban Mode (`-k`) | Factory Mode (`-f`) |
|---|---|---|
| Session makeup | 1 lead + 3 role companions (`plan` · `run` · `sync`) | 1 lead + lanes `lane-1` … `lane-N` |
| Card movement | column to column, session to session | whole into one lane, phases in order inside the lane |
| Phase execution | each column's companion session owns it | the lane spawns each phase as an `Agent()` sub-agent and runs it |
| Card classes | A/B/C define which **columns** the card **passes through** | A/B/C name only the **ceremonies** the card skips — no card ever changes sessions |
| `/clear` boundary | between phases | between cards |

In kanban, the card class defined the shape of the shortcut — Class A skipped the `plan` column, and so did Class B. The factory has no session seated per column, so this distinction dissolves. The class still names the ceremonies the card skips (Class B proceeds without `plan`, so no SPEC exists; Class A goes straight to closing), but it no longer changes which session works the card. Every lane performs whatever phases remain for its card, serially.

## Entry — opening the lead and the lanes

```bash
# Lead — a 4-lane factory run (prints the launch commands for lane-1..lane-4)
$ moai cc -f 4

# Lanes — each in its own terminal; the number goes into the command
$ moai cc -f lane-1
$ moai cc -f lane-2
$ moai cc -f lane-3
$ moai cc -f lane-4

# GLM-backend lanes take the same form
$ moai glm -f lane-3
```

Attach no count to `-f` and the run starts with one lane (`lane-1`) by default. As the queue piles up, add one lane at a time with `moai cc -f lane-<n>` (or `moai glm -f lane-<n>`). Like kanban's companions, lanes are launched **by hand, each in its own terminal** — there is no path by which a session launches another session.

One launch takes one entry token — passing `-k` and `-f` together is an error. The v1.2.0 unified entry forms — `-k <N>` (lead) and `-k <N> --name lane-<i>` (lane) — remain valid compatibility forms (a bare `-k --name lane-<i>` with no N defaults to 8 lanes). The mixed-backend launcher `moai cg` refuses the factory for the same reason as kanban (`FACTORY_MODE_UNSUPPORTED_BACKEND`). As the kanban lead's socket opens at `/tmp/moai-socket-kanban/<run-id>`, the factory lead's socket opens at `/tmp/moai-socket-factory/<run-id>`, and the bootstrap notice carries the actual path.

## The lead's routing — whole cards to free lanes

What the factory lead does differs from a kanban lead. Where a kanban lead coordinates the phases of one card, the factory lead **routes already-picked cards to free lanes**. A "free lane" here means the previous card has reached `done` and the lead has read its evidence — a lane holding a card is not called, and when every lane is busy, the card is not routed and waits in the queue.

The actor that picks a card is always the operator (`moai todo next <n>`); the factory lead does not scan the queue and line cards up. The kanban foreman loop (a bare `/loop`) is no exception — it dispatches the next card already marked `picked` and never picks one itself. The dispatch block takes the same form as kanban, with the `cmd` field pointing at the entry phase the class dictates — `/moai plan` for Class C, `/moai run` for Class B, direct close for Class A. The lane proceeds through the remaining phases on its own, with no further dispatches.

## The 3 stages inside a lane — serial

How one lane passes one card fits in three sentences. **Run does not start until plan finishes, and sync does not start until run finishes** — a lane never runs two phases of the same card at once. Each phase's execution is spawned by the lane as an `Agent()` sub-agent; the lane itself only orchestrates. Sub-agent output stays in that session's window, and the lane reads the evidence they leave to assemble the result.

```mermaid
flowchart TD
    Queue["Backlog queue<br/>(the operator picks the card)"] --> Lead["Factory lead<br/>(routes to a free lane)"]
    Lead -->|"one whole card"| Lane["Lane lane-N<br/>(session orchestration)"]
    Lane -->|"Agent() spawn"| Plan["plan<br/>SPEC authoring"]
    Plan -->|"starts only after it ends"| Run["run<br/>implementation"]
    Run -->|"starts only after it ends"| Sync["sync<br/>review lenses + docs · closure"]
    Sync -->|"the lead reads the evidence"| Done["done"]
    Done -->|"/clear, then the next card"| Lead
```

Human gates still fire while the phases proceed. Human approval points such as Implementation Kickoff Approval are never passed automatically inside a lane.

## Parallelism caps and isolation

Each lane can run **up to 10 agents concurrently** — the launcher injects a `CLAUDE_CODE_MAX_CONCURRENT_SUBAGENTS` cap into each lane session, so N lanes fanning out simultaneously divide the machine's capacity by construction rather than by operator restraint.

Parallelism has two axes, and they do not mix:

- **Between cards (fan-out)** — you can attach a worker per card and push several at once. Each worker writes only inside its own card directory (`.moai/specs/<SPEC-ID>/`), so parallel writes do not collide.
- **Within a card (stages)** — the phases of one card are serial. Two write agents are never attached to the same card in parallel — one card, one writer at a time.

When spawning write-capable sub-agents in parallel, always attach worktree isolation (`isolation: "worktree"`) — each write agent works in its own worktree copy, so file writes do not collide even outside the card directory, and the lane integrates through evidence and merges. Read-only investigation and audit fan-outs are not isolated — a place where the worktree setup cost buys nothing.

### Staggered activation and no model override

Never activate every lane at once. Activate the first lane, wait for evidence that it has started producing output (first job or visible progress), then activate the remaining lanes — concurrent requests cannot read a cache entry still being written, so simultaneous activation breaks cache efficiency. For the same reason, do not put a model override on dispatch messages — the GLM tier mapping rides the `ANTHROPIC_DEFAULT_*_MODEL` slot environment variables, and a per-spawn override splits the caches and can bypass the slot→GLM mapping.

## Lane-number ownership — workers.json

Which lane holds which number is recorded in `.moai/state/factory/workers.json`. When a new lane opens, its number skips **only those held by live sessions** and attaches to the next free number — a dead lane's number is released and reused, and its leftover claims are cleared from this file too. The `-f lane-<n>` form already names the lane, so passing `--name`/`-n` alongside it is an error.

## What does not change

The factory changes only the shape of how cards flow. The remaining rules stand word for word as in kanban:

- **The delegation channel is the queue on disk.** A dispatch is a pointer, not a copy; a message is a nudge, never the delegation itself.
- **Completion is judged only on evidence read.** A card advances on whether the progress record was read, not on whether the lane replied. The final PASS/FAIL verdict is always the lead's — the lane that produced the work judging its own output is not an allowed shape.
- **The `/clear` boundary sits between cards.** Losing the between-phase handoffs is the point of this mode, but the between-card handoff does not go away — when a card reaches `done`, the lane is `/clear`-ed before taking the next one.
- **A card worktree is not disposed of until the remote merge lands.** If the branch is not yet merged, the worktree is the only copy of that work.

## Related docs

- [Kanban Mode](/en/advanced/kanban-mode) — the first form, three roles carrying one card between columns. Card classes and the board·queue rules shared with the factory
- [`/moai todo`](/en/utility-commands/moai-todo) — the backlog queue that admits cards onto the board. The operator is the one who picks
- [manager-lead Lead Coordinator](/en/advanced/manager-lead) — the coordination agent that drives dispatch inside a kanban or factory lead session
- [`/moai loop`](/en/utility-commands/moai-loop) — the unattended foreman driven by a bare `/loop`. The same "never picks, only routes" boundary as the factory lead
- [Kanban Board Terms](/en/core-concepts/kanban-board-terms) — the formal glossary with definitions and examples of card, column, lane, and lead
