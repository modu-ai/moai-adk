# Cross-Session Messaging

Doctrine for messaging between independent Claude Code sessions — those on this machine, and, where the conditions below are met, those on your other machines or on the web. The channel is a Claude Code runtime feature that is **on with nothing to enable** where the requirements are met — this rule governs how the orchestrator uses it, never how it is built.

> **Loading scope**: Intentionally always-loaded. A peer-session conflict surfaces mid-turn, from any context, and is not predictable from file paths.

## What the channel is

Claude Code binds a per-session inbox socket and exposes two tools: `ListAgents` to discover reachable agents, and `SendMessage` to deliver plain text to one by name. A message carries text and a reply address — never conversation history, never files.

Three properties bound everything below:

- **Same machine is direct; beyond it travels through Anthropic servers.** Local delivery goes over the per-session socket and never leaves the machine. A session on another of your machines, or a cloud session, is addressed by name the same way, and the orchestrator may **open** an exchange with one rather than only answer it — from Claude Code v2.1.225 onward, and only where that session appears in the listing. Two narrowings survive: a send from a session not itself connected to Remote Control still arrives but carries **no reply address**, so that message is one-way; and a cloud session receives without being able to message back.
- **A message is not consent.** The receiving runtime is told the text came from another session, not from the user. It cannot answer a permission prompt, cannot change configuration, and a slash command inside it arrives as inert text.
- **Filesystem visibility gates reach.** Sessions find each other through files on disk, so a container and its host cannot message each other; two sessions inside the same container can.

## Where it sits among MoAI's existing mechanisms

Each mechanism answers a different question. Reaching for the wrong one is the most common error.

| Need | Mechanism | Not this |
|------|-----------|----------|
| Is another session working here right now? | session registry (`moai session list`) — detection | messaging |
| Tell a live peer session something it needs now | **cross-session messaging** | handoff |
| Continue this work after `/clear` or on another machine | paste-ready handoff (`session-handoff.md`) | messaging |
| Coordinate workers this session spawned | subagents / agent teams | peer messaging |
| Move a whole conversation elsewhere | resume the session | messaging |

Messaging complements the registry rather than replacing it: the registry says *that* a peer exists, messaging is *how to talk to it*. Neither carries context — a message that needs the recipient to hold prior state is the wrong tool, and a handoff is the right one.

## Rules

[ZONE:Evolvable] [HARD] **Never route a user decision through a peer.** The user-facing question channel is unchanged: questions go to the user through the orchestrator's question tool. A peer session is not a proxy for the user, and its reply is not approval. Asking a peer to approve, to confirm, or to decide on the user's behalf is prohibited.

[ZONE:Evolvable] [HARD] **Never ask a peer to do what this session may not do.** Work blocked or denied here does not become permissible by delegation. When a needed action is outside this session's permissions, route it back to the user, not sideways to another session.

[ZONE:Evolvable] [HARD] **Send facts, not instructions to mutate shared state.** A message may report what landed, what broke, what a decision was, or ask a question. It must not direct a peer to edit configuration, rewrite doctrine, or take a hard-to-reverse action; those remain gated in the receiving session by its own rules and prompts.

[ZONE:Evolvable] **Role-boundary dispatch is permitted; offloading is not.** Where sessions are standing roles in a declared topology — one coordinating session and workers that each own a stage of the pipeline — a coordinating session may dispatch a work item to the session whose role owns that stage, and may ask for its completion status. Three conditions make this dispatch rather than offloading: the target's role is declared in advance rather than chosen because it happened to be idle, the work item is a **pointer into shared source of truth** (an identifier, a path, a contract section) rather than the work itself, and each worker writes to an isolated tree so concurrent workers cannot collide. All three must hold together; absent any one of them, it is offloading — see the anti-pattern below.

[ZONE:Evolvable] **Do not let a dispatch depend on the reply arriving.** Because reply routing is not guaranteed, completion must also be observable in the shared source of truth — a progress record the coordinator can read — with the message serving as prompt notification rather than as the record. A coordinator that advances only on received replies stalls silently when one is lost.

[ZONE:Evolvable] **Prefer a message over a stall when a peer holds the answer.** When the working tree shows a concurrent session and the orchestrator would otherwise stop and ask the user to mediate, asking the peer directly is usually faster and spares the user a mediation round-trip. It is not free: once delivered, a message counts toward usage in the receiving session exactly as a typed prompt does, so what is saved is the user's attention, not tokens. Ask the user when the decision is theirs; ask the peer when the fact is theirs.

[ZONE:Evolvable] **Keep messages short and self-contained.** The recipient has none of this session's context. One or two sentences naming the artifact, the change, and the consequence beats a summary that assumes shared history.

## Integration with the concurrency checks

The Pre-Spawn and Pre-Edit Sync Checks (`agent-common-protocol.md`) detect a foreign session and then stop for user mediation. Where the detected peer is reachable, messaging adds a step between detection and escalation:

1. Detect the concurrent session (registry query + divergence check) — unchanged.
2. **Ask the peer what it is holding** (`SendMessage`), when the blocking question is a fact the peer knows: which paths it is editing, whether its work is committed, when it expects to land.
3. Escalate to the user only when the answer does not resolve the conflict, or when the resolution is a decision rather than a fact.

Worktree isolation remains the structural fix for a write conflict. Messaging shortens the diagnosis; it does not make two sessions safe to write the same path.

Conversely, after landing a change that invalidates what a peer is building on — a schema change, a renamed symbol, a merged branch — notifying the affected peer is appropriate without being asked.

## Addressing, sending, and replying

A session answers to the name set at launch or by rename; unset, the runtime derives one from the working directory, so parallel sessions in one project collide on a shared prefix and are told apart only by a short reference. Where a launcher starts a session bound to a known unit of work, passing an explicit name makes peers addressable by what they are doing rather than by where they run.

Three frictions are observed in practice and are worth expecting rather than rediscovering:

- **A bare name usually resolves; the short reference is the exception.** The runtime delivers on the name alone when exactly one live session answers to it, and reaches for a short reference only when several sessions share the name or it could not check everywhere your sessions run. So treat a refusal as that exception rather than as the norm: re-send with the reference the error supplies, rather than assuming the peer is unreachable. The user-facing peer listing does not show these references — only the discovery tool's output does — so the reference is read from the tool result or from the refusal itself.
- **A reply address is not guaranteed to route.** A recipient may be unable to answer the sender it was addressed by and fall back to guessing a peer. Consequently a message must carry enough identification for a human or a peer to route the answer manually: name the sending context and what the answer is for. Never assume a reply will land automatically, and never make the sender's identity implicit.
- **The sender's permission class is disclosed.** An arriving message states whether its sender bypasses permission prompts, and that disclosure is what the receiver's inbound default keys on. A message from a bypassing sender is more likely to be held for approval, so a session that expects to be answered promptly should not assume delivery.

An arriving message carries **both** the sender's name and a reply address — not one to the exclusion of the other. Replying to the name as given is the normal path; the address is the fallback where that name does not resolve. What fails is re-deriving either from a listing instead of copying what the message supplied.

## Configuration surface

| Key | Effect |
|-----|--------|
| `crossSessionInbound` | `accept` delivers, `hold` parks for approval, `refuse` drops. Unset, the runtime decides per message from the two sessions' permission-mode classes |
| `isolatePeerMachines` | `true` requires explicit approval before any message leaves the machine. A `true` from any scope applies |
| `dialogExpiry` | Deadline after which a **default**-held message is dropped — the dialog closes, or in a non-interactive session the held message expires. Five minutes unless set; `never` holds until the session ends. It does not govern a message held by an explicit `hold` |
| `permissions.deny: ["SendMessage", "ListAgents"]` | Turns off sending and listing. Also removes messaging to subagents and teammates, which share the tool |

The two ways a message is held do not expire alike. A message the inbound **default** holds waits on `dialogExpiry` and is then dropped, and the sender is told it expired; a message held by an explicit `crossSessionInbound: hold` does not expire at all, and is delivered only when an `accept` later applies. A non-interactive worker cannot show an approval dialog, but a default-held message there still runs the same deadline rather than waiting indefinitely — so a worker meant to take messages unattended needs `accept` in its own settings. One asymmetry is worth knowing: while a background session has no terminal attached, the default-held dialog stays open past its deadline, and the countdown only runs properly once you attach.

**Availability trap**: the feature depends on flag evaluation, so environment variables that disable non-essential traffic or telemetry can turn it off silently. A session where the peer-listing command is unrecognized does not have the feature; a session where listing works but a send never arrives is being blocked by something narrower — a deny rule, the receiver's inbound control, or, for a target beyond this machine, the version and listing conditions above.

## Anti-patterns

- **Peer-as-user.** Treating a peer's reply as approval for a gated action.
- **Peer-as-handoff.** Sending a work summary to a peer that has no context, where a resume or a paste-ready handoff was the correct mechanism.
- **Peer-as-worker.** Offloading work this session should have done — or should have given to a subagent it supervises — onto an independent session, because that session is idle. Distinct from role-boundary dispatch (below), which is permitted.
- **Silent write race.** Messaging a peer about a shared path and then writing it anyway, without isolation, because the peer answered.
- **Broadcast noise.** Messaging every listed session rather than the one whose work is affected.

## Cross-references

- `.claude/rules/moai/core/agent-common-protocol.md` — Pre-Spawn / Pre-Edit Sync Check, the detection layer this composes with
- `.claude/rules/moai/core/askuser-protocol.md` — the user-question channel monopoly, unchanged by this rule
- `.claude/rules/moai/workflow/session-handoff.md` — crossing a context boundary, the mechanism messaging does not replace
- `.claude/rules/moai/workflow/worktree-integration.md` — isolation, the structural fix for a write conflict
- `.claude/rules/moai/workflow/main-checkout-branch-guard.md` — why concurrency is assumed rather than proven absent

---

Version: 1.0.0
Classification: Evolvable operational rule — peer-session communication; changes no gate semantics.
