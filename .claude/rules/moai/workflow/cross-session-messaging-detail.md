---
description: "Detail companion for cross-session-messaging.md — the mechanism-selection table, addressing/reply frictions, the full configuration surface (inbound controls, isolation, dialog expiry, deny rules, inbox burst refusal), and the shared feature-flag slot that decides whether the channel exists at all"
paths: "**/cross-session-messaging*.md,**/kanban-dispatch*.md"
---

# Cross-Session Messaging — Detail Companion

> Detail companion of `cross-session-messaging.md` (the always-loaded stub). The stub owns what the
> channel is, its availability constraints, every rule, the concurrency-check integration, the
> idle-notice clause, and the anti-pattern list. This file owns the mechanism-selection table, the
> addressing and reply frictions observed in practice, the configuration surface, and the shared
> feature-flag slot behind the stub's fifth availability constraint. Load it when a send does not
> arrive, when choosing between messaging and a handoff, when configuring how a session accepts
> inbound messages, or when a third-party-backend session has lost the channel for no visible
> reason.

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


## Addressing, sending, and replying

A session answers to the name set at launch or by rename; unset, the runtime derives one from the working directory, so parallel sessions in one project collide on a shared prefix and are told apart only by a short reference. Where a launcher starts a session bound to a known unit of work, passing an explicit name makes peers addressable by what they are doing rather than by where they run.

Three frictions are observed in practice and are worth expecting rather than rediscovering:

- **A bare name usually resolves; the short reference is the exception.** The runtime delivers on the name alone when exactly one live session answers to it, and reaches for a short reference only when several sessions share the name or it could not check everywhere your sessions run. So treat a refusal as that exception rather than as the norm: re-send with the reference the error supplies, rather than assuming the peer is unreachable. These appear only in the discovery tool's output, not the user-facing listing. A same-named in-process agent fails differently: with the team namespace on it takes the bare name silently, and a `routing` object on the result is the only sign it went there and was lost. Conditional — read the result rather than always reaching for the reference.
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

A fifth path stops a message and is not a setting at all. Each inbox accepts only so many messages in quick succession; once a rapid burst would exceed what the addressed session takes, further sends to it are **refused up front** rather than reported sent and then dropped. Fan-out is the shape that reaches it — a lead nudging N lanes within one turn (Factory Mode, `moai cc -f <N>`) is precisely a rapid burst. A refusal there is the channel working, not a channel fault, and it costs nothing: delegation rides the queue on disk, never the message (`kanban-dispatch.md` § The delegation channel is the queue). Read the send result rather than assuming it, and where every lane genuinely needs nudging, spread the sends across turns instead of firing them together.

The two ways a message is held do not expire alike. A message the inbound **default** holds waits on `dialogExpiry` and is then dropped, and the sender is told it expired; a message held by an explicit `crossSessionInbound: hold` does not expire at all, and is delivered only when an `accept` later applies. A non-interactive worker cannot show an approval dialog, but a default-held message there still runs the same deadline rather than waiting indefinitely — so a worker meant to take messages unattended needs `accept` in its own settings. One asymmetry is worth knowing: while a background session has no terminal attached, the default-held dialog stays open past its deadline, and the countdown only runs properly once you attach.

Two further facts bear on unattended workers. A `claude -p` session binds an inbox socket like an interactive one, but a session started in **bare mode** binds none — it neither receives messages nor appears in listings. And the `/config` row that selects `crossSessionInbound` (v2.1.232+) does not appear while `--settings` or managed settings set the key — a companion session launched with an injected inbound value cannot change it from its own `/config`, only from the settings source that injected it.

**Availability trap**: a session where the peer-listing command is unrecognized does not have the feature — see § Availability constraints for the OS, provider, version, flag-evaluation, and shared-slot reasons; a session where listing works but a send never arrives is being blocked by something narrower — a deny rule, the receiver's inbound control, or, for a target beyond this machine, the version and listing conditions above.


## The shared flag slot

The channel's gate reads one machine-global boolean, `cachedGrowthBookFeatures.tengu_harbor_kite` in `~/.claude.json`. Two facts turn it into a shared resource rather than a per-session setting, and together they produce a failure that looks like a backend fault and is not one.

**Only a first-party session writes it.** The remote evaluation endpoint is `api.anthropic.com`, hardcoded — it does not follow `ANTHROPIC_BASE_URL`. What `ANTHROPIC_BASE_URL` contributes instead is its host, sent as a targeting attribute (`apiBaseUrlHost`) whenever that host is something other than `api.anthropic.com`. A session pointed at a third-party endpoint therefore receives no payload of its own to write, and the slot's write timestamp does not move while it runs; a first-party session started alongside it fetches and overwrites the same slot.

**The gate is re-read on every call, not cached at startup.** A session that had the channel can lose it mid-run because another session wrote `false`, and can regain it the same way. Neither transition errors.

The consequence for MoAI is direct: `moai glm`, and the GLM panes of `moai cg`, read the slot and never write it, so their messaging tracks whatever a first-party session last left on this machine. Attributing an outage there to the GLM backend is the natural reading and the wrong one — the same session works or does not according to a value it has no part in setting.

**The measurement this rests on.** Holding model and environment fixed and flipping only the slot, a session with `ANTHROPIC_BASE_URL` pointed at a third-party endpoint was started four times: `true` produced an inbox socket twice, `false` produced none twice, with no exception. Separately, a first-party session was observed overwriting the slot from `false` to `true`, while a third-party session left the write timestamp unchanged across six reads in 36 seconds — it never wrote the slot once.

**The manual escape hatch.** The gate checks `CLAUDE_CODE_HARBOR_KITE` before it reads the slot, so exporting `CLAUDE_CODE_HARBOR_KITE=1` for a session turns the channel on whatever the slot holds. MoAI does not inject it. It is an upstream internal flag rather than a documented interface, so reaching for it is an operator decision taken knowing the name can change without notice — worth having when a session is cut off and the slot is out of reach, not something to wire in by default.

---

Classification: Lazy companion — selection guidance, observed frictions, and configuration
reference only. Every rule and prohibition stays in `cross-session-messaging.md`.
