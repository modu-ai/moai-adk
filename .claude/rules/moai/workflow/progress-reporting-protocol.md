# Progress Reporting Protocol

Canonical single source of truth for how a delegated MoAI agent reports interim progress, and for how the orchestrator narrates and relays it. The protocol binds both sides — every agent run and every delegation — so it is always-loaded (no `paths:` restriction), the same reasoning that makes the question-channel protocol and the session-handoff rule always-loaded.

## 1. Why this exists

A long-running delegated agent used to give the user no interim signal — only a terminal "waiting for 1 background agent" line. The cause was an allowlist omission, not a platform limit: MoAI agents declare an explicit `tools:` allowlist, and it simply omitted the tools that carry progress. Granting them closes the gap. The orchestrator-side pull paths are closed by design (see §8), so an agent-side push plus the shared task list are the only mechanisms available.

## 2. The two channels

Progress rides two channels at once, and the difference in their guarantees is the whole design.

The **primary** channel is the shared task list. The agent registers its milestones with `TaskCreate` and marks each boundary with `TaskUpdate`, and the orchestrator reads the list with `TaskList`. It is the documented channel, and it is the one the orchestrator relies on for correctness. If anything else fails, progress is still visible here.

The **secondary** channel is a `SendMessage` push to the `to: "main"` recipient. It buys immediacy — the push surfaces at the orchestrator's next tool-call boundary without a pull. It is undocumented (see §3). If it fails, progress degrades to primary-only; it is never work-stopping.

The primary channel is load-bearing; the secondary is an immediacy enhancement layered on top of it.

## 3. Provenance and honesty

The `to: "main"` recipient is an **undocumented** runtime behavior. It is not in the documented `SendMessage` recipient list — it exists only in the runtime tool schema. It was empirically verified to deliver on Claude Code **v2.1.206**, from both foreground and background subagents. Because it is undocumented, it may break without notice. If it does, the primary channel still carries progress: the feature degrades from immediate to pull-visible, it does not disappear. No text in this protocol, in any agent body, or in any doctrine surface may describe `to: "main"` as sanctioned or official — it is honestly labeled undocumented everywhere it appears.

## 4. Agent-side contract

At the start of a run the agent registers its declared milestones on the shared task list with `TaskCreate`, and marks each boundary with `TaskUpdate`. At each boundary it additionally pushes one short status line via `SendMessage({ to: "main", ... })`. Each push:

- leads with an `[n/N]` milestone counter;
- is at most 2 lines;
- reports a **milestone-only** boundary — never a per-tool-call event, an individual file read or write, a search, or any sub-step;
- counts toward a hard cap of **at most 6** pushes per run (a per-agent contract may set a lower N; it never sets a higher one).

## 5. The boundary — status, never a question

A progress report is a **statement**, never a **question**. An agent MUST NOT ask the user anything through either channel — no question, no option list, no request for input. This is backed at the platform level: the user-question tool is unavailable to subagents even when listed in `tools`. When an agent needs user input, it returns a structured blocker report to the orchestrator, which is the only path to the user. This protocol adds a status channel, never a question channel.

## 6. Language

The push body is written in **English**, as an internal agent-to-orchestrator transfer. The orchestrator renders the user-facing text in `conversation_language` when it relays. Translation belongs at the relay, not in the agent body — duplicating it into every agent body would fork the language policy.

## 7. Orchestrator-side

**Roadmap.** Before delegating to an agent whose contract declares N of **3 or more** milestones, the orchestrator first emits a four-marker step roadmap in `conversation_language`:

| Marker | en | ko |
|---|---|---|
| NOW | `[NOW]` | `[지금]` |
| NEXT | `[NEXT]` | `[다음]` |
| LATER | `[LATER]` | `[이후]` |
| GATE | `[GATE]` | `[게이트]` |

`GATE` names the next point at which the orchestrator will stop and ask, so the user knows in advance where their turn comes. The `N >= 3` trigger reuses the per-agent N already fixed in each agent's contract — it introduces no new threshold.

**Relay.** When a push arrives, the orchestrator relays it in the same turn, in `conversation_language`, naming the emitting agent and preserving the `[n/N]` counter. It translates rather than passing the English body through verbatim.

**Durable view.** The orchestrator treats the shared task list, read via `TaskList`, as the durable progress view, and never depends on the `SendMessage` channel for progress correctness.

**Non-idle, read-only.** While a background delegation is in flight, the orchestrator shall not end its turn and idle — it continues independent work so that queued pushes drain at tool-call boundaries. All such concurrent work is **read-only**: MoAI does not run two write-capable agents concurrently, and orchestrator work concurrent with a write-capable agent is read-only. This is the concurrency safeguard against file-write races, replacing the superseded background-write restriction.

## 8. Prohibitions

- The orchestrator shall not read or tail a background agent's **transcript** output file — it is the full subagent transcript, and reading it overflows context.
- The orchestrator shall not invoke `TaskOutput` to poll a local_agent task; it is deprecated for the same reason. Note the asymmetry: `TaskList` is a permitted pull because it reads the shared task list, not the transcript.
- The progress channel shall not depend on spawning with the `name:` parameter. A named spawn depends on team-runtime initialization (the team file), which has failed in practice; both tool families are declared explicitly in `tools:` instead of relying on teammate injection.

## 9. Graceful degradation

A progress-reporting failure is never a work-stopping failure. When a `SendMessage` push fails or is rejected, the agent continues its actual work unchanged — it does not retry-loop, does not abort, and does not surface the failure as an error. This best-effort discipline is what makes the undocumented secondary channel safe to lean on for immediacy without staking correctness on it.

## 10. Cross-references

- `.claude/rules/moai/core/agent-common-protocol.md` § Background Agent Execution — the background-default policy and the compact invariant contract.
- `CLAUDE.md` § 14 Parallel Execution Safeguards — the orchestrator-side parallel/background safeguards.
- `.claude/rules/moai/core/askuser-protocol.md` — the user-question channel monopoly the boundary in §5 preserves.
- `.claude/rules/moai/core/zone-registry.md` — the registered HARD clauses this protocol introduces.

---

Version: 1.0.0
Classification: Canonical Reference — single source of truth for the progress reporting protocol. Do not duplicate content; cross-reference this file instead.
