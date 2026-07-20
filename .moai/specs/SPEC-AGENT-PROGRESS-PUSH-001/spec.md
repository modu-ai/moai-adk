---
id: SPEC-AGENT-PROGRESS-PUSH-001
title: "Dual-channel agent progress reporting + orchestrator narration + background-default realignment"
version: "0.2.0"
status: completed
created: 2026-07-12
updated: 2026-07-13
author: manager-spec
priority: High
phase: "v3.0.0 target"
module: ".claude/agents/moai, .claude/rules/moai, internal/template/templates"
lifecycle: spec-anchored
tags: "agent, progress, sendmessage, tasklist, orchestration, doctrine, template-mirror"
tier: L
era: V3R6
---

# SPEC-AGENT-PROGRESS-PUSH-001 — Agent Interim Progress Reporting

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-07-12 | manager-spec | Initial plan-phase authoring. Tier L. 26 REQ / 32 AC. Single channel (SendMessage). M0 canary as a blocking gate. |
| 0.2.0 | 2026-07-12 | manager-spec | Amendment after official-docs verification. Single channel → **dual channel** (Task\* primary/documented, SendMessage secondary/undocumented). M0 gate **deleted** — resolved empirically (foreground push works) and replaced by a standing regression check. **New scope**: background-default realignment. Both clarifications resolved. 37 REQ / 48 AC. |

---

## §A Context

### A.1 Problem (user-reported)

A long-running delegated agent gives the user **no interim progress signal**. The user waits with only a "Waiting for 1 background agent to finish" line, sometimes for an extended period. The user wants step-by-step progress **during** the run, not only a terminal report.

### A.2 Root cause — an allowlist omission, not a platform limitation

Official documentation (`code.claude.com/docs/en/sub-agents` § Available tools) states that subagents inherit the main conversation's internal tools by default, and enumerates the tools that are **not** available to subagents even when listed in `tools`:

`AskUserQuestion` · `EnterPlanMode` · `ExitPlanMode` (unless `permissionMode: plan`) · `ScheduleWakeup` · `WaitForMcpServers`

`SendMessage` and the `Task*` family are **not** on that list. Both are documented internal tools whose names are the exact strings used in subagent tool lists. Because MoAI agents declare an **explicit** `tools:` allowlist, anything omitted from it is excluded.

**Root cause, precisely**: MoAI agents lack a progress channel not because the platform forbids one, but because our own `tools:` allowlist omits the tools that provide it. Granting them is an officially supported path.

Measured baseline (research.md §A): `SendMessage` is absent from all 18 agent files (9 agents × 2 trees). The `Task*` family is present in 5 of 9 agents and absent from 4 (`manager-design`, `plan-auditor`, `super-advisor`, `sync-auditor`).

### A.3 Channel viability — verified, but asymmetrically documented

| Channel | Documented? | Verified? |
|---|---|---|
| `TaskCreate` / `TaskUpdate` / `TaskList` (shared task list) | **Yes** — tools-reference | Standard MoAI usage |
| `SendMessage({to: "main"})` | **No** — documented recipients are an agent-team teammate, or a subagent resumed by ID or name. The `"main"` recipient exists **only** in the runtime tool schema | **Yes** — empirically, on Claude Code v2.1.206, from **both** background and foreground subagents. Return: `{"success":true,"message":"Message queued for the main conversation's next turn."}`, and the message rendered in the main conversation |

The runtime schema annotates `"main"` as *"background subagents only"*. That annotation is **not enforced** at the tool-call layer — a foreground push succeeds. An undocumented behavior that works carries a different risk profile from a documented one: it may change without notice.

That asymmetry is the reason for the dual-channel design (§B Layer 1).

### A.4 Polling is closed

The `Agent` tool result explicitly warns against reading or tailing the agent's output file — it is the full subagent transcript and reading it overflows the orchestrator's context. `TaskOutput` is deprecated for local_agent tasks for the same reason. Both orchestrator-side pull paths are closed, which is what makes agent-side push and the shared task list the only mechanisms.

### A.5 Delivery timing

Pushes are "queued for the main conversation's next turn" and empirically surface at tool-call boundaries. If the orchestrator ends its turn and idles immediately after a background spawn, delivery is deferred. The doctrine must therefore bind orchestrator engagement, not only agent emission.

### A.6 Named-spawn hazard

A spawn with the `name:` parameter failed with `Internal error: team file for "session-<id>" not found`; the same spawn without `name:` succeeded. Teammates receive Agent-Teams tools by framework injection — but that injection path is the one that failed. The progress channel must therefore be declared explicitly in `tools:` and must not depend on team-runtime initialization.

### A.7 Doctrine-versus-runtime drift (new scope)

Official documentation (`code.claude.com/docs/en/sub-agents` § Run subagents in foreground or background) states that **as of v2.1.198, subagents run in the background by default**; Claude runs one in the foreground when it needs the result before continuing. The default changes *where* a subagent runs, not *what it is allowed to do* — background subagents still surface every permission prompt in the main session. Separately (v2.1.186), a background subagent's permission prompt surfaces in the main session naming the asking subagent, with Esc denying that one call without stopping the subagent.

The running runtime is **v2.1.206** — the flipped default is active. MoAI doctrine still asserts the pre-flip world on four surfaces (measured, research.md §D):

1. `CLAUDE.md` §14 — Background Agent Write Restriction bullet
2. `.claude/rules/moai/core/agent-common-protocol.md` § Background Agent Execution
3. `.claude/rules/moai/workflow/worktree-integration.md` line 114
4. `.claude/rules/moai/core/zone-registry.md` — `CONST-V3R2-020`, `CONST-V3R2-044`

No doctrine surface mentions v2.1.198. This is doctrine-versus-runtime drift, and it is directly entangled with this SPEC: forcing write-capable agents to the foreground is what makes the orchestrator block, and a blocked orchestrator has no tool-call boundary at which a queued push can drain.

---

## §B Requirements (GEARS)

### Layer 1 — Dual-channel agent contract

**REQ-APP-001** (Ubiquitous)
Each of the 9 retained MoAI agent definitions shall declare `SendMessage` in its frontmatter `tools:` CSV, in both the working tree and the template mirror.

**REQ-APP-002** (Ubiquitous)
Each of the 9 retained MoAI agent definitions shall declare `TaskCreate`, `TaskUpdate`, `TaskList`, and `TaskGet` in its frontmatter `tools:` CSV, in both trees.

**REQ-APP-003** (Ubiquitous)
Each of the 9 retained MoAI agent bodies shall carry a `## Progress Reporting Contract` section naming that agent's milestone boundaries, its maximum push count, and the canonical protocol rule by path.

**REQ-APP-004** (Event-driven — primary channel)
**When** a delegated agent begins its run, the agent shall register its declared milestones on the shared task list via `TaskCreate`; **when** it reaches each milestone boundary, it shall record progress via `TaskUpdate`. This is the primary, officially documented progress channel.

**REQ-APP-005** (Event-driven — secondary channel)
**When** a delegated agent completes a milestone boundary, the agent shall additionally push one status message via `SendMessage` with `to: "main"`, carrying a short summary and a message body. This is the secondary channel, providing immediate delivery.

**REQ-APP-006** (Ubiquitous — honest provenance)
The canonical protocol rule shall state plainly that `to: "main"` is an **undocumented runtime behavior**, shall record the Claude Code version it was empirically verified against, and shall state that it may break without notice, at which point the primary channel still carries progress.

**REQ-APP-007** (Unwanted behavior)
No rule, agent body, or doctrine text shall imply that `to: "main"` is an officially sanctioned or documented recipient.

**REQ-APP-008** (Event-detected — degradation)
**When** a `SendMessage` push fails or is rejected, the agent shall continue its actual work unchanged: it shall not enter a retry loop, shall not abort, and shall not surface the failure as an error. A progress-reporting failure is never a work-stopping failure.

**REQ-APP-009** (Unwanted behavior — boundary preservation)
A progress report on either channel shall not contain a question, a request for user input, an option list, or any invocation of the user-question tool. A progress report is a **statement**, never a **question**.

**REQ-APP-010** (Ubiquitous)
A `SendMessage` progress push shall be at most 2 lines and shall carry an `[n/N]` milestone counter as its leading token.

**REQ-APP-011** (Ubiquitous — noise cap)
An agent shall emit at most 6 `SendMessage` progress pushes per run. A per-agent contract may set a lower cap; it shall never set a higher one.

**REQ-APP-012** (Unwanted behavior — noise cap)
An agent shall not emit a progress push for a per-tool-call event, an individual file read or write, a search, or any sub-step that is not a named milestone boundary in its contract.

**REQ-APP-013** (Ubiquitous — language)
The progress message body shall be written in English, as an internal agent-to-orchestrator transfer; the orchestrator shall render the user-facing text in `conversation_language` on relay.

**REQ-APP-014** (Capability gate — blocker channel unchanged)
**Where** an agent requires user input that was not supplied in its spawn prompt, the agent shall return a structured blocker report to the orchestrator, and shall not attempt to obtain that input through either progress channel.

### Layer 2 — Orchestrator narration and relay

**REQ-APP-015** (Event-driven — step roadmap)
**When** the orchestrator is about to delegate to an agent whose Progress Reporting Contract declares `N` of 3 or more milestones, the orchestrator shall first emit a step roadmap carrying four markers — NOW / NEXT / LATER / GATE — rendered in the user's `conversation_language`.

**REQ-APP-016** (Event-driven — relay)
**When** an agent progress message arrives in the main conversation, the orchestrator shall relay it to the user in `conversation_language` within the same turn, naming the emitting agent and preserving the `[n/N]` counter.

**REQ-APP-017** (Ubiquitous — durable view)
The orchestrator shall treat the shared task list, read via `TaskList`, as the durable progress view, and shall not depend on the `SendMessage` channel for progress correctness.

**REQ-APP-018** (State-driven — non-idle, read-only)
**While** a background agent delegation is in flight, the orchestrator shall not end its turn and idle; it shall continue independent work so that queued messages are delivered at tool-call boundaries, and all such concurrent work shall be **read-only**.

**REQ-APP-019** (Unwanted behavior — polling prohibition)
The orchestrator shall not read or tail a background agent's transcript output file, and shall not invoke `TaskOutput` to poll a local_agent task.

**REQ-APP-020** (Ubiquitous — spawn safety)
The orchestrator shall spawn MoAI agents without the `name:` parameter by default; a named spawn depends on team-runtime initialization and shall not be relied upon as the progress-channel enabler.

### Layer 3 — Doctrine and registration

**REQ-APP-021** (Ubiquitous)
A canonical rule file `.claude/rules/moai/workflow/progress-reporting-protocol.md` shall exist and shall be the single source of truth for the progress reporting protocol.

**REQ-APP-022** (Ubiquitous)
The `## Background Agent Execution` section of `.claude/rules/moai/core/agent-common-protocol.md` shall carry a pointer to the canonical protocol rule by file path.

**REQ-APP-023** (Ubiquitous)
Section 14 of `CLAUDE.md` shall carry a pointer to the canonical protocol rule by file path.

**REQ-APP-024** (Ubiquitous)
Every new HARD clause introduced by this work shall be registered in `.claude/rules/moai/core/zone-registry.md` under the `CONST-V3R6-NNN` namespace, each entry carrying `id`, `zone`, `zone_class`, `file`, `anchor`, `clause`, and `canary_gate`.

**REQ-APP-025** (Ubiquitous — boundary backing)
The canonical protocol rule shall cite the official documentation finding that the user-question tool is unavailable to subagents even when listed in `tools`, as the platform-level backing for the subagent question prohibition.

### Layer 4 — Background-default realignment

**REQ-APP-026** (Ubiquitous — audit)
Every doctrine surface asserting the superseded background-execution default shall be enumerated, and each shall be brought into alignment with the documented runtime behavior.

**REQ-APP-027** (Ubiquitous — decision)
MoAI shall align with the documented runtime default rather than forcing foreground execution for write-capable agents. The doctrine shall state this decision and its reasoning explicitly.

**REQ-APP-028** (Unwanted behavior — retained safeguard)
MoAI shall not run two write-capable agents concurrently, and orchestrator work performed concurrently with a write-capable agent shall be read-only. The safeguard against write races is **concurrency control**, not foreground forcing.

**REQ-APP-029** (Ubiquitous — registry amendment)
`CONST-V3R2-020` and `CONST-V3R2-044` shall be amended in place to reflect the realigned behavior, and shall not be left asserting the superseded default.

**REQ-APP-030** (Ubiquitous — marker reconciliation)
The inline zone markers on the background-execution clauses shall be reconciled with their zone-registry classification, and the unregistered HARD clause in `worktree-integration.md` shall be brought into the registry or removed. The zone registry is the declared single source of truth for zone classification.

**REQ-APP-031** (Unwanted behavior)
The `background:` frontmatter field shall not be set on any MoAI agent definition; the runtime shall be left to choose the execution mode.

### Layer 5 — Template mirroring and guards

**REQ-APP-032** (Ubiquitous)
Every distributed file changed under `.claude/` shall have a corresponding mirror under `internal/template/templates/.claude/`, and every changed `CLAUDE.md` shall have its mirror under `internal/template/templates/CLAUDE.md`.

**REQ-APP-033** (Ubiquitous — neutrality)
The template mirror of every file changed by this work shall be free of internal-development content: SPEC identifiers, requirement or acceptance-criterion tokens, internal work dates, and commit hashes.

**REQ-APP-034** (Ubiquitous — mirror CI enrollment)
The canonical protocol rule shall be enrolled in the byte-parity mirror allowlist consumed by the template mirror CI guard, so that a future single-tree edit is caught at CI rather than shipping stale.

**REQ-APP-035** (Ubiquitous — parity-class preservation)
The live-versus-template parity class of each changed file shall be preserved: files byte-identical before the change shall remain byte-identical, and files carrying a pre-existing sanitization divergence shall retain exactly that divergence and no more.

**REQ-APP-036** (Event-driven — build and guard)
**When** files under `internal/template/templates/` change, the embedded template filesystem shall be regenerated and the template guard suite shall pass with no new failure.

### Layer 6 — Standing regression check

**REQ-APP-037** (Ubiquitous)
The undocumented `to: "main"` channel shall carry a standing regression check that re-verifies delivery on the current runtime, and the canonical rule shall record the Claude Code version against which the channel was last verified.

---

## §C Out of Scope

### Out of Scope — runtime and CLI changes

- Any change to Go production logic. This work changes agent definitions, rule documents, and one CI-guard allowlist entry only.
- Any new `moai` CLI subcommand, hook script, or configuration loader for progress reporting.
- Any persisted progress ledger, progress state file, or progress telemetry beyond the runtime's own shared task list.

### Out of Scope — user-interaction channel

- Any relaxation of the user-question channel monopoly. The user-question tool is unavailable to subagents at the platform level regardless of the `tools:` field; this work adds a status channel, never a question channel.
- Any bidirectional channel. The progress channels are one-directional: agent to orchestrator.

### Out of Scope — deferred tuning

- A configurable noise cap in project configuration. The cap is a fixed doctrine constant in this work; making it configurable is a follow-up.
- Progress reporting for agents that have no MoAI-owned definition file to edit.
- Re-enabling or re-designing the retired Agent Teams static-orchestration layer.

### Out of Scope — worktree isolation policy

- The worktree opt-in policy itself. `worktree-integration.md` is touched only where it asserts the superseded background-execution default; its isolation guidance is not otherwise revised.

---

## §D Success Criteria

1. All 9 agents declare both channels' tools and carry a Progress Reporting Contract, in both trees.
2. A delegated agent's milestones appear on the shared task list, and its pushes reach the user relayed in `conversation_language`.
3. The doctrine states honestly which channel is documented and which is not, and records the runtime version the undocumented one was verified against.
4. The subagent question boundary is provably unchanged and now cites platform-level backing.
5. No doctrine surface asserts the superseded background-execution default; the two affected registry clauses are amended in place, not left stale.
6. Template mirrors are neutral, parity-class preserved, mirror-CI enrolled, build and guard suite green.
