---
description: "Detail companion of agent-common-protocol.md — the relocated non-binding bodies (rationale, companion pointer paragraphs, incident pointers) plus the migrated cross-reference material. Loads on paths contact with the agent-common-protocol rule family."
paths: "**/agent-common-protocol*.md"
---


## Migrated from the core body


### Agent Common Protocol


Shared protocol for all MoAI agent definitions. This rule is automatically loaded for all agents, eliminating the need to duplicate these sections in each agent body.


> **Detail companion**: `agent-common-protocol-reference.md` — verbatim verification batch, output contracts, CLI idioms, Ledger Closure clause bodies, sync-check rationale + incident records, and the bodies relocated from here: § Orchestrator Obligations · § Re-delegation Procedure · § Skeptical Evaluation Stance · § CLAUDE.md Reference · § File Operations Pattern · § Search Pattern · § Tool Selection by Task · § Bash Timeout · § Error Recovery Pattern · § Super-Advisor Escalation (E1-E4) · § Read-only verification batching · § Attributable diff-check doctrinal switch. Load it when composing a verification batch, selecting a tool, recovering from a failed call, escalating to super-advisor, or handling an aborted delegation.


### Subagent Prohibitions


Rationale: subagents run in isolated, stateless contexts — prompting there is a dead channel, and the orchestrator stays the user's single point of contact (askuser-protocol.md).


### Hook Invocation Surface


Three hook scripts enforce orchestrator-discipline obligations — `status-transition-ownership.sh` (PostToolUse on SPEC-artifact writes), `sync-phase-quality-gate.sh` (Stop on sync-phase commit, blocking only under `MOAI_SYNC_GATE_BLOCKING=1`), `team-ac-verify.sh` (TaskCompleted in team mode; registered in no settings surface, so no flag activates it). All three exit 0 always and signal through stdout JSON, honored only on exit 0 — on exit 2 it is discarded and only stderr surfaces. Per-row triggers, JSON shapes, owning policy, and the subagent-boundary criterion: `agent-common-protocol-reference.md` § Hook Invocation Surface detail.


### Ledger Closure


The **ledger-closure invariant**: an aborted `Agent()` delegation leaves no **dangling tool_use** —
an open promise with no matching result — in the orchestrator's context. It is the in-session
analogue of the model-API rule that every `tool_use` receives a `tool_result`.


**Scope-boundary note.** Ledger Closure is a sibling of (not nested in) Hook Invocation Surface
under the User Interaction Boundary H2.


### Pre-Spawn Sync Check (Multi-Session Race Mitigation)


> **Spawn-gate boundary**: this fires only at the write-agent spawn boundary; direct main-session edits bypass it — see § Pre-Edit Sync Check below. Defense-in-depth: `.moai/docs/generic-patterns-guide.md` § Multi-Session Race Mitigation Procedure; worktree-as-race-elimination: `session-handoff.md` § Worktree-Anchored Resume Pattern.


### The sweep prohibition


**Ambient signal.** The SessionStart hook already lists foreign active sessions in a `<system-reminder>` — the always-on detection layer; this check is the decision layer that turns detection into isolation.

