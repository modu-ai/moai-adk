# Cross-Session Messaging

Doctrine for messaging between independent Claude Code sessions — those on this machine, and, where the conditions below are met, those on your other machines or on the web. The channel is a Claude Code runtime feature that is **on with nothing to enable** where the requirements are met — this rule governs how the orchestrator uses it, never how it is built.

> **Loading scope**: Intentionally always-loaded. A peer-session conflict surfaces mid-turn, from any context, and is not predictable from file paths.

## What the channel is

Claude Code binds a per-session inbox socket and exposes two tools: `ListAgents` to discover reachable agents, and `SendMessage` to deliver plain text to one by name. A message carries text and a reply address — never conversation history, never files. A send may additionally carry an opt-in `notify_when_idle` request: the runtime returns one notice when the addressed session next goes idle — one-shot, no polling, and on the same platforms as the channel itself. What that notice does and does not establish is § An idle notice is a scheduling hint.

Three properties bound everything below:

- **Same machine is direct; beyond it travels through Anthropic servers.** Local delivery goes over the per-session socket and never leaves the machine. A session on another of your machines, or a cloud session, is addressed by name the same way, and the orchestrator may **open** an exchange with one rather than only answer it — from Claude Code v2.1.225 onward, and only where that session appears in the listing. Two narrowings survive: a send from a session not itself connected to Remote Control still arrives but carries **no reply address**, so that message is one-way; and a cloud session receives without being able to message back.
- **A message is not consent.** The receiving runtime is told the text came from another session, not from the user. It cannot answer a permission prompt, cannot change configuration, and a slash command inside it arrives as inert text.
- **Filesystem visibility gates reach.** Sessions find each other through files on disk, so a container and its host cannot message each other; two sessions inside the same container can.

## Availability constraints

"On with nothing to enable" holds only where the platform provides the channel. Five constraints bound where it exists at all — and because Kanban Mode delegates through the queue on disk, using this channel only to nudge companions, they bound where its nudges reach:

The five axes, and the one diagnostic that separates "absent" from "blocked": **operating system**, **provider**, **runtime version**, **feature-flag evaluation** (the four opt-out env vars), and **the shared machine-global flag slot** that third-party-backend sessions inherit and can lose mid-session. Per-axis versions, provider splits, and the slot's mechanism and manual escape hatch: `cross-session-messaging-detail.md` § Availability constraints and § The shared flag slot.

Where a constraint bites, the failure is quiet — nothing errors, dispatch just has no channel. Surface the constraint to the operator instead of retrying or re-spawning.

## Where it sits among MoAI's existing mechanisms

Each mechanism answers a different question, and reaching for the wrong one is the common error:
the session registry answers *is another session working here*, messaging answers *tell a live peer
something now*, a paste-ready handoff answers *continue this work after `/clear` or elsewhere*,
subagents answer *coordinate workers this session spawned*, and resuming a session answers *move a
whole conversation*. Messaging complements the registry rather than replacing it, and neither
carries context — a message needing the recipient to hold prior state is the wrong tool. Full
table: `cross-session-messaging-detail.md` § Where it sits.

## Rules

[ZONE:Evolvable] [HARD] **Never route a user decision through a peer.** The user-facing question channel is unchanged: questions go to the user through the orchestrator's question tool. A peer session is not a proxy for the user, and its reply is not approval. Asking a peer to approve, to confirm, or to decide on the user's behalf is prohibited.

[ZONE:Evolvable] [HARD] **Never ask a peer to do what this session may not do.** Work blocked or denied here does not become permissible by delegation. When a needed action is outside this session's permissions, route it back to the user, not sideways to another session.

[ZONE:Evolvable] [HARD] **Send facts, not instructions to mutate shared state.** A message may report what landed, what broke, what a decision was, or ask a question. It must not direct a peer to edit configuration, rewrite doctrine, or take a hard-to-reverse action; those remain gated in the receiving session by its own rules and prompts.

[ZONE:Evolvable] [HARD] **Never address a stopped teammate by name.** A teammate stopped with
`TaskStop` keeps a live name address: one message delivered to that name resumes the agent from its
transcript, reviving it as an ownerless writer mid-card. When composing a message to a teammate by
name, confirm the addressee is not a stopped teammate before sending, and route coordination about
a stopped teammate through the owning orchestrator or lead only — never to the stopped name.
Observing a stopped teammate running again is a process defect: report it to the lead immediately
and record it in the progress record; do not continue quietly or send that teammate more messages.

**Mechanism layer (landed).** The prohibition above is mechanically enforced at the
PreToolUse layer: a SendMessage whose recipient — bare name, agent id, or the
`name [ref]` form (suffix parsed and stripped) — matches a live entry in the same
session's stop registry (`.moai/state/agent-stops/<session-id>.json`, populated by the
matcher-less PostToolUse dispatch on every TaskStop completion) is refused with a
sentinel-prefixed deny (`STOPPED_TEAMMATE_VIOLATION:`). Deliberate revival stays
possible through an explicit fresh spawn carrying the same name — the entry clears
before the spawn proceeds; session end removes the session's entries. Every stop,
send, deny, and clear appends one JSONL row to `.moai/logs/agent-stop-audit.jsonl`.
The deny layer is opt-in (`workflow.agent_stop_guard.enabled` — template default
false; this repository's local config enables it for dogfood). The flag is a kill
switch for the DENY only: recording and audit continue regardless of its state. A
`STOPPED_TEAMMATE_VIOLATION` deny is never a bug to route around — it means the
doctrine held: route coordination through the owning orchestrator, or respawn the
name deliberately.

[ZONE:Evolvable] **Role-boundary dispatch is permitted; offloading is not.** Where sessions are standing roles in a declared topology — one coordinating session and workers that each own a stage of the pipeline — a coordinating session may dispatch a work item to the session whose role owns that stage, and may ask for its completion status. Three conditions make this dispatch rather than offloading: the target's role is declared in advance rather than chosen because it happened to be idle, the work item is a **pointer into shared source of truth** (an identifier, a path, a contract section) rather than the work itself, and each worker writes to an isolated tree so concurrent workers cannot collide. All three must hold together; absent any one of them, it is offloading — see the anti-pattern below.

[ZONE:Evolvable] **Do not let a dispatch depend on the reply arriving.** Because reply routing is not guaranteed, completion must also be observable in the shared source of truth — a progress record the coordinator can read — with the message serving as prompt notification rather than as the record. A coordinator that advances only on received replies stalls silently when one is lost.

[ZONE:Evolvable] **Prefer a message over a stall when a peer holds the answer.** When the working tree shows a concurrent session and the orchestrator would otherwise stop and ask the user to mediate, asking the peer directly is usually faster and spares the user a mediation round-trip. It is not free: a delivered message counts toward the recipient's usage exactly as a typed prompt does — what is saved is the user's attention, not tokens. Ask the user when the decision is theirs; ask the peer when the fact is theirs.

[ZONE:Evolvable] **Keep messages short and self-contained.** The recipient has none of this session's context. One or two sentences naming the artifact, the change, and the consequence beats a summary that assumes shared history.

## Integration with the concurrency checks

The Pre-Spawn and Pre-Edit Sync Checks (`agent-common-protocol.md`) detect a foreign session and then stop for user mediation. Where the detected peer is reachable, messaging adds a step between detection and escalation:

1. Detect the concurrent session (registry query + divergence check) — unchanged.
2. **Ask the peer what it is holding** (`SendMessage`), when the blocking question is a fact the peer knows: which paths it is editing, whether its work is committed, when it expects to land.
3. Escalate to the user only when the answer does not resolve the conflict, or when the resolution is a decision rather than a fact.

Worktree isolation remains the structural fix for a write conflict. Messaging shortens the diagnosis; it does not make two sessions safe to write the same path.

Conversely, after landing a change that invalidates what a peer is building on — a schema change, a renamed symbol, a merged branch — notifying the affected peer is appropriate without being asked.

## A send result has three shapes, and none of them says "read"

[ZONE:Evolvable] [HARD] **A successful send means the message reached the session, not that its Claude read it.** The result answers where the text went; it never answers whether a model consumed it. Three shapes, and the result text is what tells them apart:

| Result | What happened | What to do |
|---|---|---|
| Queued to the addressed session | The text is in that session's inbox and drains at its next tool round | Nothing — but completion still comes from the evidence, not from this |
| A `routing` object | An in-process mailbox took it; the peer never sees it | Re-send to `name [ref]` |
| A `[Cross-session delivery notice]` follows | The receiving session's permission policy is **holding** the message for its user's approval, or refused it outright | Treat it as undelivered: surface it to the operator rather than re-sending, because the same policy holds the next copy too |

The third shape is the one that used to leave no trace. A session running in a different permission mode than the sender's holds inbound peer messages until its user approves them, and may let them expire; for a session on this machine the notice reports that, and the notice is the only signal — nothing in the original send result predicts it.

**A notice never arrives for a Remote Control, cloud, or Claude Desktop peer.** Silence there is not agreement and not delivery; it is the absence of a channel to report either. Never read it as a reply.

**The queue is what survives all three shapes.** Because a dispatch is delegated through the queue on disk and completion is read from evidence (`kanban-dispatch.md` § The delegation channel is the queue, § Completion is read, never trusted), a held or lost message costs the board nothing. That is exactly why reading the send result matters: it tells the sender whether a *nudge* landed, and nothing more. Advancing a card because a send reported success is an unobserved completion claim (`verification-claim-integrity.md` §1.1 surface 1).

## An idle notice is a scheduling hint

A send may ask the addressed session to report back once, when it next goes idle (`notify_when_idle`). It is opt-in per send and one-shot — the request is spent on the first notice, so a second notice needs a second request — and it replaces a polling loop on the asking side.

[ZONE:Evolvable] [HARD] **An idle notice is not completion evidence.** A session goes idle when it finishes, when it stops at a permission prompt, and when it dies, and the notice cannot tell those three apart. What it establishes is *when to go look*; what it says about the work is nothing. Treating it as a completion signal converts the [HARD] read-don't-trust rule (`kanban-dispatch.md` § Completion is read, never trusted) into an unobserved completion claim (`verification-claim-integrity.md` §1.1 surface 1) — the notice arrives, the card advances, and no one read the evidence.

Used for what it is, it removes waste: instead of re-reading a progress file on a guessed interval, ask for the notice and read the evidence once, when there is something to read.

## Addressing and configuration

A session answers to the name set at launch or by rename; the bare name delivers when exactly one
live session answers to it, and the short `[ref]` is the exception the error text supplies. An
arriving message carries both the sender's name and a reply address — reply to the name as given,
and fall back to the address only when that name does not resolve. Reply routing is not guaranteed,
so a message carries enough identification for a human or a peer to route the answer by hand.

Inbound acceptance, cross-machine isolation, dialog expiry, deny rules, and the inbox's rapid-burst
refusal (the shape a lead's fan-out nudge reaches) are configuration, not doctrine:
`cross-session-messaging-detail.md` § Addressing, sending, and replying and § Configuration
surface. The availability trap is diagnostic — a session where the peer-listing command is
unrecognized does not have the feature at all (§ Availability constraints above); one where listing
works but a send never arrives is being blocked by something narrower.

## Anti-patterns

- **Peer-as-user.** Treating a peer's reply as approval for a gated action.
- **Peer-as-handoff.** Sending a work summary to a peer that has no context, where a resume or a paste-ready handoff was the correct mechanism.
- **Peer-as-worker.** Offloading work this session should have done — or should have given to a subagent it supervises — onto an independent session, because that session is idle. Distinct from role-boundary dispatch (below), which is permitted.
- **Reviving a stopped teammate.** Addressing a stopped teammate by name — one delivered message
  resumes it from the transcript as an ownerless writer. Coordination about it goes through the
  owning orchestrator or lead, never the stopped name (now also mechanically denied — see the
  mechanism layer note above).
- **Silent write race.** Messaging a peer about a shared path and then writing it anyway, without isolation, because the peer answered.
- **Broadcast noise.** Messaging every listed session rather than the one whose work is affected.

## Codex broker path (session messaging tools)

A Codex peer is unreachable by the channel above — a Codex session has no Claude Code runtime — and rides the moai MCP broker instead (`session_msg_register` / `session_msg_list` / `session_msg_send` / `session_msg_poll`, poll-based, symmetric for both session kinds). **Every rule above extends to it unchanged**, and a Codex reader never loads this rules tree: the tool descriptions carry the discipline for that side. Tool surface, the per-clause broker reading, and the MCP-server restart caveat: `cross-session-messaging-detail.md` § The Codex broker path.

> Origin: SPEC-CODEX-SESSION-MSG-001 (design.md §8 mapping).

## Cross-references

- `.claude/rules/moai/core/agent-common-protocol.md` — Pre-Spawn / Pre-Edit Sync Check, the detection layer this composes with
- `.claude/rules/moai/core/askuser-protocol.md` — the user-question channel monopoly, unchanged by this rule
- `.claude/rules/moai/workflow/session-handoff.md` — crossing a context boundary, the mechanism messaging does not replace
- `.claude/rules/moai/workflow/worktree-integration.md` — isolation, the structural fix for a write conflict
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` — why concurrency is assumed rather than proven absent

---

Version: 1.3.0
Classification: Evolvable operational rule — peer-session communication; changes no gate semantics.
