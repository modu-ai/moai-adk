---
id: SPEC-KANBAN-BOOTSTRAP-001
title: "Design — Kanban session topology, bootstrap, and dispatch"
version: "0.3.0"
created: 2026-08-10
updated: 2026-08-11
author: manager-spec
priority: High
phase: "v3.1.0 target"
module: internal/kanban
lifecycle: spec-anchored
tags: "kanban, design, topology, bootstrap, dispatch, backend, sole-writer, role-resolution"
tier: L
dependencies: [SPEC-KANBAN-RENAME-001, SPEC-KANBAN-BOARD-001, SPEC-KANBAN-WORKTREE-001]
related_specs: [SPEC-KANBAN-MULTISESSION-001, SPEC-FACTORY-MODE-001]
---

## §A. Shape of the thing

Five sessions and a board. The board and the trees are elsewhere (`SPEC-KANBAN-BOARD-001`, `SPEC-KANBAN-WORKTREE-001`); what is designed here is the population of sessions, how it comes into being, and the one sentence one session is allowed to say to another.

```
operator
   │  runs the entry switch
   ▼
entry switch ──[no topology config]──▶ single-session chain  (identical to pre-change)
   │
   └──[topology config]──▶ print guidance to stderr
                              │  operator launches by hand
                              ▼
                          poll for quorum ──[bound elapses]──▶ name who is missing, exit non-zero
                              │
                              ▼ quorum
                          lead ──dispatch(id, path, section)──▶ plan → run → review → sync
                              ▲     (addressed by declared role)       │
                              └────── progress.md is the truth ────────┘
                                      (a reply is only a nudge)

  operator admits a card ──▶ plan   …   sync ──▶ done  (lead's terminal write)
  (not a dispatch: backlog has no session to report)
```

Two properties hold the whole design together and both come from the substrate rather than from taste. **Sockets cannot launch sessions**, so the operator is inside the loop at bootstrap and there is no design in which they are not. **Replies are not reliably routed**, so the reply is demoted from a control signal to a nudge, and the board advances on evidence read from a shared file.

## §B. The topology

Five roles, four of which own a column and one of which owns none:

| Role | Column | Sessions deployed |
|---|---|---|
| `lead` | none — watches all six | exactly 1, fixed backend; the board's sole writer per `REQ-KB-017` |
| `plan` | `plan` | 1 |
| `run` | `run` | 1 by default, 2 by configuration |
| `review` | `review` | 1 |
| `sync` | `sync` | 1 |

The lead does no card work. That is not a division of labor for its own sake — it falls out of the fact that WIP admission is a whole-board rule a worker cannot evaluate (a `run` session cannot know how many cards are in `run`), so one actor has to hold the board's view, and an actor holding the board's view while also driving a card would be doing both jobs badly.

The same fact settles a second question this design deliberately does not answer for itself: **who writes the board.** `REQ-KB-017` makes the `lead` the sole writer, and the reasoning converges — the actor that must hold the whole-board view to make an admission decision is the only actor positioned to record the result of one. This design elects that actor; it defines no write path, and the atomicity and board-wide locking that make a single writer safe are `REQ-KB-018` and `REQ-KB-019`, not this document's. What binds here is the negative form: nothing this design prints and nothing it sends may put a board write in a worker's hands.

### B.0 A session says which role it is, and that is a second datum

A table of roles is not a runtime lookup. Something has to answer "which of these running sessions is the `review` session", and the answer has four consumers rather than one: this design's dispatch routing, its quorum accounting, the board sibling's runtime refusal of a non-`lead` write (`AC-KB-017`), and the worktree sibling's two gates that resolve the `lead` occupant from a session that is not it (`REQ-KW-007`, `REQ-KW-011`). Until v0.3.0 each of the four pointed at `REQ-KS-004`, which elects the role set and says nothing about occupancy.

The obvious carrier — the launch label — cannot serve, and the reason is structural rather than incidental. Labels must all differ (`REQ-KS-014`), an unconfigured worker role prints one command per supported backend, and the `run` role may deploy two sessions. One role therefore corresponds to two-or-more possible labels, and the operator picks which exists when they choose a command to run. The mapping does not invert, and making it invert would cost either the distinct-label rule or the operator's backend choice.

So the role is **declared**, alongside the label rather than inside it. Three properties, and the third is the one an implementation would omit: it is the routing key, it is the quorum key, and it resolves **from a session that is not the `lead`**. A lead-private declaration satisfies everything visible from inside this design and breaks both of the worktree sibling's gates — a failure that surfaces in another document's criteria, which is precisely the kind this design is now shaped to avoid, having lost the sole-writer rule to the same shape once already.

What is *not* designed here is where the declaration lives. The launch command, the session registry, and the peer-discovery output are all plausible carriers and nothing measured favours one, so the contract is fixed and the carrier is left to run-phase. Fixing it here would be fixing a decision on preference and calling it a design.

### B.1 Why the session count is not the WIP limit

The `run` column admits two cards; the deployment ships one coder session. Those are different numbers on purpose, and the temptation to collapse them into one is strong enough that the board sibling has a requirement (`REQ-KB-010`) whose only job is to keep them apart.

Collapsing them would mean admission is gated on a session being free. The result is that the effective WIP silently equals the session count, and raising WIP has no effect until someone also raises the session count — which is the kind of coupling that is discovered by an operator wondering why a configuration change did nothing.

Kept apart, the second card enters `run` and waits **unheld**. That state already exists in the board (`REQ-KB-011`); the lead dispatches the card the moment a coder session frees up. It converges, too, with the state a released holder leaves behind (`REQ-KW-011`), so the board needs one representation rather than two.

## §C. Bootstrap

### C.1 Manual, by necessity

There is no mechanism by which the lead spawns a peer. This is not a limitation being routed around; it is the shape of the transport. The bootstrap therefore prints, the operator launches, and the lead polls.

The guidance must be **executable** — copyable commands, each carrying the label the lead will later address. Anything less than executable pushes the operator into transcribing flags by hand at exactly the moment they are trying to start four processes.

### C.2 The bootstrap cannot ask

CLI code may not call `AskUserQuestion` or `mcp__askuser__*`. This is inconvenient in exactly one place, and it happens to be this one: "launch these four sessions, press enter when ready" is the natural shape of a bootstrap and is prohibited.

So the bootstrap emits to stderr and then **polls**. It never blocks on an answer it is not allowed to ask for. The polling is what replaces the "press enter" — instead of the operator signalling readiness, the lead observes it.

### C.3 The quorum bound aborts

Bounded by configuration, defaulting to 300 seconds. On expiry: name the roles that answered, name the roles that did not, exit non-zero.

Proceeding with a partial team is the option that feels generous and is not. A missing role means its column has no owner. The first card to reach that column sits there while the lead, correctly, refuses to dispatch it to a session that does not declare that role. Nothing errors. The board just stops, and stopping is a much more expensive symptom to diagnose than failing, because there is no message to search for.

Aborting converts a silent late failure into a loud immediate one, and the recovery costs one relaunch.

### C.4 The backends

The lead is fixed. The workers are chosen. Both facts have to be expressed by a program that may not ask a question, and the resolution is to move the choice into the output:

- backend named in configuration → **one** command printed for that role
- backend not named → **both** commands printed; the operator runs one

Printing both is a choice offered without a question asked. It costs roughly a doubled line count in the unconfigured case, which is a fair price for staying inside the prompt prohibition without taking the choice away.

The lead's own line is in neither branch either — it is fixed, and fixity has to be *observable* to be a design property rather than a hope. So each channel a backend could arrive through carries a defined response: a configured `lead` backend fails the load with a named error, a backend argument on the launcher fails to parse because no such parameter exists, and an environment variable carrying one leaves the emitted line untouched. Only the first needs to be loud; a configuration key silently dropped is indistinguishable from one silently honored, and indistinguishable is the state this design is trying to leave.

The honest boundary sits one step past that. An operator who ignores the printed guidance and launches the lead under the other launcher is not stopped by anything here, because the entry switch prints and does not supervise. The property being designed is *no offered path to a non-Claude lead* — not *no such lead can exist* — and stating the smaller true thing costs nothing while the larger false one would cost a criterion nobody can evaluate.

**What the two launchers do, now measured.** All of the above assumed the launchers are per-session backend selectors. Measured (`research.md` §J), each is that *and* a project-global mutation. The selector half holds because the backend goes into the process environment and the launcher execs into `claude`, which inherits it — so the printed-command mechanism genuinely delivers per-worker backends, and had that failed the whole of §C.4 would have been designing something that does not work. The half nobody had checked is that `moai glm` persists `team_mode` and, in tmux, the shared session environment, while `moai cc` clears both, strips GLM settings, and removes `worker-`-prefixed worktrees. Since the bootstrap prescribes running these commands up to five times per project, the interleaving is the ordinary case.

The design does not change; two readings of it do. Persisted `team_mode` is a record of launch **order**, not of team composition, so nothing here may read it as the latter — composition is the role declarations of §B.0. And `worker-` is a name the worktree sibling cannot use for a per-card tree, because launching a Claude-backed worker would delete it.

The mixed backend appears in none of these branches. It is rejected with its existing sentinel, and the rejection is preserved rather than reconsidered: a mixed leader/teammate backend contradicts the one-session/one-backend premise the state record encodes, and since sockets cannot launch sessions, the messaging layer is not a replacement for the process model that backend assumes.

### C.5 The no-config path, and the baseline that guards it

With no topology configuration, the entry switch behaves exactly as it did before. That is the property most likely to be assumed rather than checked, so it is compared mechanically against a recorded baseline: parse result, environment mutation, exec argument vector.

The recording has a **window**, and the window is the part the predecessor missed. The rename retires every identifier the baseline would encode. Record before the rename and the baseline holds the old flag token and the old environment names; every later comparison then fails, and it fails for the rename rather than for a regression. A guard that fires on a sanctioned change trains its readers to re-record instead of investigate, and after the second or third re-record it is no longer a guard at all.

So: record after the rename lands, before the entry switch is touched. Narrow, sequencing-sensitive, and the only window in which the recording means what the comparison will read.

**And the same shape is owed to every other equivalence claim this SPEC makes.** The entry-switch baseline was the one place the discipline appeared; three other claims — the human gates unchanged, the coder chain's carried-over behaviors unchanged, each mirror pair's measured relationship preserved — asked for a comparison against a "before" that nothing obliged anyone to write down. That is a weaker failure than recording late: a claim with no comparand at all cannot fail, because at verification time the original is gone and the only remaining evidence is the verifier's memory of it. Each of the three now names a durable artifact and the point before which it is taken, and each is judged on provenance as well as content, exactly as the entry-switch baseline already was.

## §D. Dispatch

### D.1 The message is a pointer

Three fields: SPEC identifier, file path, contract section reference. The SPEC file on disk is the truth; the message is an address for it.

This is the doctrine's pointer-only condition, and it is also the only design that survives unreliable replies. A pointer is re-sendable and idempotent; a body is not — re-sending a body risks the receiver acting on a stale copy of a contract that has since changed on disk.

### D.2 Structural, not textual

Checking "no requirement text" by grepping the payload for requirement tokens is a proxy that fails in the interesting case: a payload that paraphrases the requirement in fresh words passes the grep and violates the rule.

The property is therefore made **unrepresentable**. The payload is a typed record whose fields are the three above, with no free-text field anywhere in it. There is nowhere to put a body, so nobody puts one, and no reviewer has to notice that somebody did.

### D.3 Addressing, and the refusal that carries its own fix

Selection and addressing are separate steps and use different keys: the lead *selects* by declared role (§B.0) and then *addresses* the selected session by its label plus reference. Collapsing them — selecting by label — would be selecting by whichever printed command the operator happened to run.

Sending by name alone is refused; the send needs the peer's short reference. That refusal is not a failure state — it **carries the reference**. So the lead's handling is: send, and on refusal re-send with what the refusal supplied. Reporting the peer unreachable at that point would be reporting a fixable condition as a fatal one.

### D.4 Nudge and truth

A reply is a hint that something happened. `progress.md` is what actually happened. The lead advances a card only on evidence read from the shared source of truth, and reaches the same board state whether the reply arrives promptly, arrives late, or never arrives at all.

This is also where the sole-writer rule becomes a design property rather than a policy. The worker writes its own `progress.md`; the lead reads it and writes the board. Letting the worker record its own progression in board state would look like a shortcut — one fewer hop, and the worker knows first — and it would install a second writer, which is exactly what `REQ-KB-017` forbids and what the atomicity and locking of `REQ-KB-018` / `REQ-KB-019` are sized for one of. The hop is the design.

This is the direct consequence of the measured routing unreliability, and it has a pleasant side effect: the board becomes testable by replay, because its state is a function of files rather than of message timing.

### D.4a The cycle, both of its ends, and the role it used to omit

The dispatch cycle is `plan → run → review → sync`, one dispatch per arrow, each arrow taken when the lead reads completion evidence in the card's `progress.md`. `review` belongs in it: it owns a column and runs the verify exit gate, and earlier revisions of this family wrote the sequence without it — a small omission in prose that, copied into an implementation, is a role that quietly never receives work.

Neither end of the cycle is a dispatch, and both are worth stating rather than leaving to inference:

- **`sync → done`** is the last arrow with the dispatch removed. Nothing lives in `done` to address, so the lead reads the sync session's evidence and writes the terminal transition.
- **`backlog → plan`** cannot be an arrow. `backlog` has no owning session, so no completion report exists, and a lead that admitted cards on its own would be inventing work rather than scheduling it. Admission is an operator act — the same operator who launches the sessions decides what enters the working columns — and the lead's loop starts at the first dispatchable column.

The *mechanism* of that operator admission is designed nowhere in this family. That is a gap, it is recorded as one (`plan.md` §B.4), and it is deliberately not filled here: this design has no measurement that would decide between a CLI verb, a hand edit, and an import, and inventing one would be the third time a contract in this family got written by whoever happened to be nearby.

### D.5 The three prohibitions

Each one closes a path that a naive lead would take, and all three come from "a message is not consent":

- **No operator decision through a peer.** The peer's runtime cannot ask the user either; routing a decision there produces a dead channel with extra steps.
- **No peer reply as approval.** The receiving runtime is told the text came from another session. Treating that as user approval would launder an unapproved action through a hop.
- **No asking a peer to do what the lead may not do.** Otherwise the permission boundary is defeated by delegation, which is the whole point of having one.

### D.6 Idempotency

A re-sent dispatch for a card already in progress advances nothing and corrupts nothing. The substrate makes this load-bearing rather than tidy: messages are held for approval more often than expected, so re-sends are a normal event and not an exceptional one.

## §E. The coder session's internal chain

Deliberately the smallest section here. Each `run` session drives its card through the chain the single-session mode already defines: the verify exit gate at the exit of run-phase, the revision dedup predicate per card, the goal preset over one card.

The design decision is **not to design**. No second chaining mechanism, no relocated gate semantics, no new preset. If this part of the implementation produces a large diff, something has been rebuilt that should have been pointed at.

The human gates are untouched in number, ordering, and meaning. A board is a scheduling change; it is not a licence to ship work past a review nobody removed on purpose.

## §F. What this design does not touch

Board state has a single origin under the primary checkout, resolved through the common git directory, and a single writer — the `lead` (`REQ-KB-017`). This design's whole obligation toward both is negative: nothing it prints and nothing it sends tells a session to read or write its own tree's copy, and nothing it prints or sends gives a session other than the `lead` a way to write the board at all. A worker resolving board state relative to its own working tree would find a different file, or none, and the board would fork silently per worktree — a failure with no error message, which is the expensive kind. A worker *writing* the shared file would not fork it; it would corrupt the one thing every session reads, which is worse.

Neither rule is restated here, and the second is deferred by name rather than by silence, because silence is how it went missing: the board sibling records that its predecessor bundled the ownership rule with a rejected storage mechanism, the split deleted both, and each side's exclusions pointed at the other. "This design decides who is told" was true and was not a claim about who writes.

## §G. Out of Scope

### Out of Scope — carried from spec.md §C

- The board model, the trees, the mixed backend, any spawn mechanism, any change to the messaging channel, and any interactive prompt.

### Out of Scope — this document specifically

- The carrier of a session's role declaration — launch command, session registry, or peer-discovery output. The contract is designed (§B.0); the carrier is a run-phase choice with no measurement here to decide it.
- The mechanism by which an operator admits a card into `plan` (§D.4a). Owned by no SPEC in this family; recorded as a gap rather than invented here.
- Concrete type names, function signatures, file layouts, and package boundaries. Those are run-phase decisions; this document fixes the shape and the reasons, not the identifiers.
- Any wire format or serialization choice for the dispatch record. The constraint is the field set and the absence of a free-text field; how it is encoded is not decided here.
