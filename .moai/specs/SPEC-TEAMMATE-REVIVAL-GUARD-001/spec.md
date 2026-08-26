---
id: SPEC-TEAMMATE-REVIVAL-GUARD-001
title: "Mechanism: stop-registry + SendMessage deny guard against revival of stopped teammates"
version: "0.1.1"
status: in-progress
created: 2026-08-26
updated: 2026-08-26
author: manager-spec (card t267)
priority: P2
phase: "v3.1.4 target"
module: "internal/hook, .claude/settings.json, internal/template/templates/.claude/settings.json.tmpl, internal/config"
lifecycle: spec-anchored
tags: "agent-teams, hooks, pretool-use, stop-registry, guard, audit, template-mirror"
tier: M
related_specs:
  - SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001
  - SPEC-ZONE-REGISTRY-RESYNC-001
---

# SPEC: Stop-registry + SendMessage deny guard — the mechanism layer beneath the stopped-teammate doctrine

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-26 | manager-spec | Initial plan-phase emission (card t267, Class C). Incident evidence from SPEC-ZONE-REGISTRY-RESYNC-001 (card t232, 2026-08-25); doctrine boundary per SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001 §1.3 (landed 2026-08-26, card t269). Per-layer feasibility findings measured on tree `c9eed8ac6` — plan.md §C carries command + observed output for each claim. |
| 0.1.1 | 2026-08-26 | manager-spec | Plan-audit iteration-2 fix pass (card t267): D1 enforcement default RESOLVED false (operator decision 2026-08-26, AskUserQuestion round; clarification marker stripped); D2 unmatcher'd PostToolUse dispatch corrected to `handle-harness-observe.sh` (evidence row B3); D3 full E-P1 recipe attributed to M2; D4 template twin path corrected to `settings.json.tmpl`; D5 `name [ref]` recipient matching added to REQ-TRG-003 + AC-TRG-011. |

## 1. Problem — measured shape

### 1.1 The revival mechanism (proven incident)

`TaskStop` halts a named teammate's execution but does not reclaim its name address. Any later `SendMessage` delivered to that name **resumes the agent from its transcript** — this is a deliberate, runtime-documented Claude Code feature, not an error:

> `.moai/research/cc-changelog-snapshot-2.1.233.md:3236-3237` (v2.1.77): "The Agent tool no longer accepts a `resume` parameter — use `SendMessage({to: agentId})` to continue a previously spawned agent" / "`SendMessage` now auto-resumes stopped agents in the background instead of returning an error."

The same release removed the Agent tool's `resume` parameter, making `SendMessage` the **only** programmatic resume path — which is precisely why denying the send closes the known vector.

Incident (card t267 evidence, from t232): teammate `zrr-spec-amend` was stopped via TaskStop, then revived by ONE coordination `SendMessage` from another agent; in the revived window it executed M2 drafts (`49630cba2`/`adde4cfc9`), M3 (`a74362427`), and a sync 3-phase close (`a35ff0c60..0d8e3ce32`). Commit `ef93a9d1e` self-describes as "stop-resurrected zrr-spec-amend teammate" — that commit's existence is itself proof the resume path is real. Harm shape: ownership violation (unowned writer), not artifact pollution — t232 output passed full independent re-verification. The risk is that the NEXT revival passes without re-verification.

Rejected hypothesis (do not reinvestigate): stop-goal-series evaluators or SubagentStop block decisions caused the revival. Lane-4's first-party account refuted it; the cause is address survival alone. This rejection bars hooks as a **cause** explanation only — hooks as an **enforcement** layer remain this SPEC's subject.

### 1.2 Why a mechanism, not more discipline (the representative mutant)

The lead already broadcast "don't message that name" on 2026-08-25. Broadcasting harder is human discipline, and discipline breaks when the next participant doesn't know. The doctrine layer (SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001, status: completed) prohibits the send in rule text; **nothing today denies it mechanically**. This SPEC supplies that mechanism.

### 1.3 Boundary with t269 (doctrine) and t232 (origin)

- t269 (SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001) owns the **doctrine layer**: rule text in `.claude/rules/moai/workflow/cross-session-messaging.md` and `.claude/rules/moai/core/agent-common-protocol.md`. Landed; this SPEC does not touch `.claude/rules/**` at plan-phase. A rule amendment that references the mechanism (and its kill switch) is a PROPOSED run-phase deliverable (plan.md M3), authored only after the mechanism lands.
- t232 (SPEC-ZONE-REGISTRY-RESYNC-001) is the originating incident; its outputs passed re-verification and need no repair here.

## 2. Requirements (GEARS)

### 2.1 User story

As a kanban lead / orchestrator working an agent team, I want a `SendMessage` addressed to a teammate I have stopped to be mechanically refused — with the attempt recorded — so that the next participant's ignorance cannot resurrect an ownerless writer inside my card's worktree; and when a revival does happen by any path, I want the audit trail to let me attribute the commit window afterward.

### 2.2 Requirements

- **REQ-TRG-001** (Ubiquitous — audit trail): **The stop-guard audit trail** shall append exactly one structured JSONL event to `.moai/logs/agent-stop-audit.jsonl` for every `TaskStop` completion and `SendMessage` issuance it observes — event fields `{timestamp, session_id, kind, name, agent_id, decision}` with `kind ∈ {stop_recorded, send_observed, send_denied, respawn_cleared, session_cleared}` — carrying enough precision (UTC timestamp + session + name/id + decision) to correlate a revival window with commit timestamps after the fact. The audit trail records regardless of the enforcement gate's state. *(fix direction 3 — visibility)*
  - Files: `internal/hook/agent_stop_guard.go` (new), `.moai/logs/agent-stop-audit.jsonl` (runtime artifact)
  - Precedent: `internal/hook/agent_model_guard.go:54,167-195` (JSONL audit, silent-failure append)

- **REQ-TRG-002** (Event-driven — stop recording): **When** a `TaskStop` tool call against a named agent completes successfully, **the stop-guard** shall persist a stop record `{session_id, name, agent_id, stopped_at}` into the per-session stop registry at `.moai/state/agent-stops/<session-id>.json` and emit the `stop_recorded` audit event.
  - Recording rides the already-wired unmatcher'd `PostToolUse` dispatch (`.claude/settings.json`'s matcher-less entry delivers every tool completion to `handle-harness-observe.sh` → `moai hook harness-observe` today — measured, plan.md §B-B3); run-phase wiring extends that dispatch or adds a matcher-bound entry to recognize `TaskStop` completions.
  - Repeated stops of the same name upsert idempotently (same name ⇒ refresh `stopped_at`, no duplicate entries).

- **REQ-TRG-003** (Event-driven — deny the send): **When** the enforcement gate is enabled and a `SendMessage` tool call's recipient — bare name, agent id, or the sanctioned `name [ref]` form (the optional `[ref]` suffix is parsed and stripped before comparison) — matches a live entry in the same session's stop registry, **the stop-guard** shall deny the call with a sentinel-prefixed reason (`STOPPED_TEAMMATE_VIOLATION: …`) that names the stopped teammate and routes coordination through the owning orchestrator, and shall emit the `send_denied` audit event. *(fix direction 2 — reject instead of revive)*
  - Wiring: new `PreToolUse` matcher entry `SendMessage|TaskStop` in `.claude/settings.json` + template twin, routed to the existing `handle-pre-tool.sh` wrapper (matcher-syntax precedent: the existing `Agent|Task` entry — plan.md §C.1-E1).

- **REQ-TRG-004** (Unwanted — fail-open): **The stop-guard** shall not deny a `SendMessage` call on uncertain evidence — unparseable `tool_input`, absent recipient field, unreadable or missing registry file, or ambiguous name/id match — and shall route every such case to observe-only with an audit event. An enforcement bug must never wedge a session (house norm: `branch_guard.go`, `agent_model_guard.go:22-26`).

- **REQ-TRG-005** (Event-driven — deliberate revival is a spawn, not a message): **When** an `Agent`/`Task` spawn payload carries a `name` argument matching a live stop-registry entry, **the stop-guard** shall clear that entry before the spawn proceeds — deliberate revival stays possible through an explicit, visible fresh spawn while remaining impossible through an accidental message.
  - Implementation note: `extractAgentSpawn` (`agent_model_guard.go:87-102`) already parses the spawn payload; it gains a `Name` field read.

- **REQ-TRG-006** (Event-driven — session-end cleanup): **When** a session ends, **the session-end path** shall remove that session's stop-registry entries, so stale records never outlive the session that produced them or leak into a later session's name reuse.

- **REQ-TRG-007** (Capability gate — enforcement default): **Where** `workflow.agent_stop_guard.enabled` is false (shipped template default; house-norm-consistent with `Workflow.BranchGuard.Enabled` and `Workflow.AgentModelGuard.Enabled`), **the stop-guard** shall observe and advise without denying; where true, REQ-TRG-003's deny path is active. The shipped default is **false** — resolved by operator decision 2026-08-26 (orchestrator AskUserQuestion round, recommended option taken; recorded in plan.md §C.3) — alongside the default-flip upgrade trigger: flip the template default only after M3 dogfood shows zero false-positive denies.

- **REQ-TRG-008** (Ubiquitous — template neutrality): **The settings wiring, hook code, and config defaults this SPEC adds** shall ship in the template twins (`internal/template/templates/…`) free of SPEC IDs, REQ tokens, commit SHAs, and internal dates, per the template-internal-isolation doctrine (C1–C8 catalogue; CI guard `template-neutrality-check.yaml`).

## 3. Traceability summary

| REQ | AC (acceptance.md) | Milestone (plan.md §F) |
|-----|--------------------|------------------------|
| REQ-TRG-001 | AC-TRG-001, AC-TRG-008 | M1 |
| REQ-TRG-002 | AC-TRG-001 | M1 |
| REQ-TRG-003 | AC-TRG-002, AC-TRG-007, AC-TRG-011 | M2 |
| REQ-TRG-004 | AC-TRG-003, AC-TRG-004 | M2 |
| REQ-TRG-005 | AC-TRG-005 | M2 |
| REQ-TRG-006 | AC-TRG-006 | M2 |
| REQ-TRG-007 | AC-TRG-007 | M2 |
| REQ-TRG-008 | AC-TRG-009 | M1 (mirror), M2 (config) |

## 4. Design anchors (binding constraints)

1. **Fail-open** — the guard denies only on positive evidence (registry hit + enabled gate); every uncertain path allows. Deny reason carries the sentinel prefix for source matching.
2. **No runtime patching** — fix direction 1 (address reclamation inside the Claude Code runtime) is rejected as infeasible for this repository; see plan.md §C.2 for the evidence.
3. **Budget discipline** — the guard performs local-file reads/writes only (registry JSON + JSONL append); it must complete well inside the existing PreToolUse 10s budget. No network, no subprocess fan-out.
4. **Template-First cycle** — every `.claude/` or config artifact lands in `internal/template/templates/` first, then `make build`, then local sync; the `.sh`/`.sh.tmpl` twin discipline applies to any wrapper touched.
5. **Subagent hook contract** — hook handlers never prompt the user; diagnostics via stderr, decisions via structured stdout (exit 2 / `decision: block`).
6. **Per-session scoping** — the registry is keyed by session id (one team per session); cross-session sends are a different addressing layer with their own doctrine (Out of Scope).

## 5. Out of Scope

### Out of Scope — Runtime address reclamation (fix direction 1)

- Reclaiming the name→agent mapping inside the Claude Code runtime (`TaskStop` unregistering the address). The runtime is not patchable from this repository, and auto-resume-on-SendMessage is a documented deliberate feature (`.moai/research/cc-changelog-snapshot-2.1.233.md:3237`). Any true reclamation belongs upstream.
- Consequence accepted: the name stays resolvable; this SPEC makes messaging it refusable-and-recorded instead.

### Out of Scope — Write-path quarantine of revived agents (deferred hardening)

- Denying state-mutating tool calls (e.g. `git commit`) issued by an agent whose identity matches a live stop record. Official docs confirm subagent tool calls fire hooks with `agent_id`/`agent_type`, so the layer is reachable — but `agent_id` stability across a resume is unverified, and a wrong quarantine wedges legitimate work. Deferred until M3 dogfood evidence exists; recorded as a candidate follow-up SPEC, not a hidden scope creep.

### Out of Scope — Cross-session stopped-name tracking

- The registry is per-session. Sends crossing sessions (local socket / Remote Control) follow a different delivery layer with its own availability constraints and doctrine (`.claude/rules/moai/workflow/cross-session-messaging.md` § Availability constraints).

### Out of Scope — sessionmsg broker enforcement

- `internal/sessionmsg` is the claude↔codex A2A broker (`agent.go` AgentRecord: kind/name/cwd/pid — cross-session sessions, not in-process teammates). The native teammate inbox does not transit it; enforcing there would not cover the incident vector. `session_msg_send` symmetry may be added later if a codex-side incident class appears.

### Out of Scope — Doctrine edits during plan-phase

- This SPEC does not edit `.claude/rules/**` (landed t269 scope). The M3 deliverable proposes a rule amendment (mechanism pointer + kill switch) as text for the orchestrator to dispatch; it is not applied by this SPEC.

## 6. Cross-references

- SPEC-TEAMMATE-REVIVAL-SOLE-WRITER-001 — doctrine layer (REQ-TRSW-001..005); this SPEC is the mechanism beneath REQ-TRSW-001/002/004.
- SPEC-ZONE-REGISTRY-RESYNC-001 — originating incident (card t232).
- SPEC-WORKTREE-BRANCH-GUARD-001 / SPEC-WORKTREE-BRANCH-GUARD-OPTIN-001 — the fail-open, opt-in guard precedent this design follows.
- `internal/hook/agent_model_guard.go` — architectural precedent (PreToolUse on agent-management tool, observe/advise/block layering, JSONL audit).
- `.moai/docs/hook-development.md` — hook contract; `internal/hook/CLAUDE.md` — package conventions.
- Card t267 queue text (authoritative incident/design-direction source); t224 (related: lanes unable to spawn agents → ownership matrix blur — separate card).
