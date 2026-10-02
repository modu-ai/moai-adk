---
id: SPEC-FACTORY-STALE-RUN-HEAL-001
title: "Factory stale-run self-healing — automatic lane re-registration on run switch, executable relaunch command in the stale-run notice, and the moai factory relaunch verb"
version: "0.1.0"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
priority: P1
phase: "v3.2.1 target"
module: "internal/hook, internal/factorymsg, internal/cli, internal/kanban"
lifecycle: spec-anchored
tier: M
tags: "factory, stale-run, self-healing, lane-rebind, relaunch-verb, hook-notice, card-t1345"
card: t1345
depends_on: [SPEC-FACTORY-LANE-JOIN-SOCKET-001, SPEC-STALE-RUN-LABEL-001]
related_specs: [SPEC-FACTORY-RUN-RETIRE-001, SPEC-ROLE-NAMING-CODE-001, SPEC-FACTORY-SELF-DISPATCH-001, SPEC-FACTORY-LANE-WORKTREE-HANDOFF-001, SPEC-FACTORY-BOOTSTRAP-001]
---

# SPEC-FACTORY-STALE-RUN-HEAL-001 — Factory stale-run self-healing

## A. History

| Date | Version | Change |
|---|---|---|
| 2026-10-02 | 0.1.0 | Initial plan-phase draft (card t1345, Tier M, Class C). Authored from a read of the join-repair base (SPEC-FACTORY-LANE-JOIN-SOCKET-001, completed), the stale-prescription repair (SPEC-STALE-RUN-LABEL-001, completed — it carved this card out by name), and the hook/launcher code at tree `802a72235536958ada5b7cd5876a168e4b8c325f`. Scope decided by the leader: Factory Mode only (Kanban Mode is being removed by card t1399). |

## B. Problem Statement

A factory lane session carries its run identity in its launch environment, and that identity is frozen for the session's life: a hook subprocess cannot change its parent's environment, and `/clear` does not re-read it. When the run that environment names dies or is replaced, three things go wrong, and all three were measured on 2026-09-28/30 (run `tlwgk9` retired by the retirement machinery with `{"classification":"dead","basis":"stamp"}`; its successor `tm3yoq` live):

1. **The lane never re-registers.** The leader of the new run dispatches through the new run's broker, but the lane's hook keeps opening the old run's broker. Three lanes were locked out of new dispatch reception. One session lost 2h40m of work bound to the dead run. The t1330 join repair (SPEC-FACTORY-LANE-JOIN-SOCKET-001) fixed the launcher entry only — a lane launched after the run switch joins correctly — and has no consumer for a lane session that is already running.
2. **The stale-run notice prescribes a sentence, not a command.** SPEC-STALE-RUN-LABEL-001 stopped the every-turn repetition and added the one-time unbind notice, but what the notices say is still prose: "end this session, retire the run with 'moai factory runs --retire X', then relaunch" (the final step names no command), and a re-bind line that prints the placeholder `moai cc -f lane-<n>` (4 occurrences, one per locale). An operator cannot paste either.
3. **There is no verb for the way back.** Getting a stale lane session back into a live run takes a retire step, a manual join, and knowledge of the lane's own label and provider. No `moai factory relaunch` exists.

This SPEC delivers exactly those three behaviours: (1) automatic lane re-registration into the live run, (2) an executable command in the notice, (3) the `moai factory relaunch` verb.

## C. Root Cause (mechanism, read from code at `802a72235`)

1. **The environment pins the run; the broker is per run.** The two in-session readers of `MOAI_KANBAN_ID` are the peer-registration path (`internal/hook/factory_messages.go:55`) and the inbox claim path (`factory_messages.go:135`). A run's broker is its own database file (`factorymsg.BrokerPath(root, runID)`), so a lane whose environment names the dead run claims from the dead run's broker while the new leader writes to the new run's broker. Nothing connects the two: the lane receives nothing, silently — the inbound lock. (The lane's own card verbs are not pinned — see item 5.)
2. **A current-vocabulary lane never reaches the stale-run gate.** SPEC-STALE-RUN-LABEL-001's run-state gate (`staleRunPrescriptionGate`, `internal/hook/stale_run_gate.go`) is called only from the legacy-label branch (`factory_messages.go:68-70`). A current-vocabulary label (`lane-<n>`) whose run is retired falls through to `ValidateActiveRun` (`factory_messages.go:90-92`) and returns `factory messaging degraded: NO_ACTIVE_FACTORY` on every prompt — the same repetition defect, in the vocabulary the gate does not cover.
3. **The notice text is not executable.** `internal/hook/session_stale_run.go` — `roleValueRetire` / `laneLabelRetire` (retire-then-"relaunch" prose, final step unnamed) and `laneLabelUnbindRebind` (placeholder `lane-<n>`). Measured: `grep -c "lane-<n>" internal/hook/session_stale_run.go` reads 4; `grep -c "factory relaunch" internal/hook/session_stale_run.go` reads 0.
4. **The relaunch clear-policy loop resolves its run once.** `runFactoryLaneRelaunch` (`internal/cli/factory_lane_relaunch.go:56`) reads `MOAI_KANBAN_ID` a single time and passes that id to every `factoryNextLeaseOnce` iteration; card selection lists cards by run id (`factoryNextSelectAndLease` → `ListCards(ctx, runID)`) with no run-status check. A lane launcher that outlives its run keeps leasing from the run it started in. Measured: `grep -c "enterFactoryLaneRun(" internal/cli/factory_lane_relaunch.go` reads 0 (positive control `internal/cli/cc.go` reads 1). This is an observation from reading; the failing behaviour is observed at the run phase's RED step.
5. **The card verbs already follow a run switch.** `moai factory next|stage|complete|assign|status|decide` resolve their run through `resolveFactoryCardRun` (`internal/cli/factory_card.go:1159`): the explicit `--run`, else the single active run. They are not pinned to the environment and are out of the repair surface. The pinned surfaces are the hook (items 1-2) and the relaunch loop (item 4).
6. **No verb wraps the t1330 join for an existing session.** The shared lane-join gate (`enterFactoryLaneRun`, `internal/cli/factory.go:505`) — resolution, verified leader discovery on `NO_ACTIVE_FACTORY`, run resume, re-entry — runs only inside the `moai cc|glm|codex -f` launchers. The existing `moai factory runs --retire` retires; nothing joins.

**Separation evidence (preserved from SPEC-STALE-RUN-LABEL-001):** worker-70 carried the same stale environment yet kept receiving dispatches, because the inbound claim path is keyed on the broker peer record, not on the environment label. This SPEC adds a reader of the live run's broker for a rebound session; it never gates the existing claim path (REQ-SRH-008, REQ-SRH-009).

## D. Requirements (GEARS)

### D.1 The notice names an executable command

- **REQ-SRH-001** (Ubiquitous): The factory stale-run notice and the factory unbind notice shall name the way back into a live run as an executable command line — a complete `moai factory relaunch` invocation whose every argument value is resolved from the session's own launch facts (provider, lane label, run id) and which contains no angle-bracket placeholder — one line per candidate run, in place of the multi-step retire-then-relaunch prose and the `lane-<n>` placeholder line.
- **REQ-SRH-002** (Ubiquitous): The command grammar shall be one shared definition used both to print the line in the notices and to parse it in the verb, so that a printed line is accepted verbatim by the verb; the command line (command name, flags, run id, lane label, provider token) shall be emitted verbatim in all four locales (en, ko, ja, zh) with only the surrounding prose localized; the provider token shall reflect the session's own launch provider (`cc`, `glm`, or `codex`); and the command shall pin the lane label for a current-vocabulary session and omit it for a legacy-vocabulary session, so that a legacy label is never mapped to a number (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-009).
- **REQ-SRH-003** (Ubiquitous): The notice cadence fixed by SPEC-STALE-RUN-LABEL-001 shall be unchanged: the active-run prescription at most once per session identity, the unbind notice exactly once per session identity and final, and the command line in the unbind notice only while an active run exists in the project root — this SPEC changes what the notices say, never when.

### D.2 Automatic lane re-registration in the hook

- **REQ-SRH-004** (Event-driven): **When** a factory lane session carrying a current-vocabulary lane label submits a prompt while the run its launch environment names measures not active and exactly one other run in the same project root measures active, the hook shall register that session as the lane's peer in the active run under the same lane slot, and shall emit one rebound notice per session identity naming the previous run, the new run, the slot, and the registration generation.
- **REQ-SRH-005** (State-driven): **While** a session is eligible for rebind under REQ-SRH-004, the SessionStart hook shall establish no peer binding and shall emit no unbind notice for it: an eligible session is rebind-pending, not unbound, and the first prompt performs the registration.
- **REQ-SRH-006** (Event-driven): **When** a current-vocabulary lane session's named run measures not active and the project root holds zero active runs or more than one, the hook shall register no peer in any run; for zero active runs it shall emit the once-only unbind notice (replacing the per-prompt degraded string of §C.2); for several it shall emit, once only, an ambiguity notice naming every candidate run with its own command line, and shall not select among them (SPEC-FACTORY-RUN-RETIRE-001 REQ-014).
- **REQ-SRH-007** (Unwanted): The hook shall not register a legacy-vocabulary session into any run, and shall not displace a live owner of the target slot: when the slot in the target run is held by a live session the registration shall be refused with no write, and the session shall receive one notice naming the refusal and the executable command.
- **REQ-SRH-008** (Ubiquitous): For a rebound session the inbox claim shall continue to read the broker of the run its environment names exactly as before, and shall additionally read the broker of the run it is rebound into within the same inspection budget, surfacing messages from both; the environment-label judgment shall not gate, filter, or delay delivery from either broker (SPEC-STALE-RUN-LABEL-001 REQ-SRL-007).
- **REQ-SRH-009** (Ubiquitous): For a current-vocabulary lane session whose named run measures active, the hook's behaviour — bind, registration, claim, notice output, and the number of run-state measurements — shall be unchanged from before this SPEC (SPEC-STALE-RUN-LABEL-001 REQ-SRL-008).
- **REQ-SRH-010** (Unwanted): The hook shall fail open and shall not write run records: a measurement or registration failure shall degrade to the existing degraded answer or to silence, never to a hook error or a prescription; the added measurements shall complete inside the existing hook inspection deadline; and the hook's only writes shall be the peer registration and the once-per-identity notice marker — it shall not retire, reactivate, or create any run.

### D.3 The relaunch clear-policy loop follows the run

- **REQ-SRH-011** (Event-driven): **When** the `--clear-policy relaunch` lane loop begins an iteration, it shall re-resolve the run through the same selection the launch used — an explicit `--factory-run` stays explicit; otherwise the shared lane-join gate, including verified leader discovery — before leasing any card, shall hand the resolved run to the child session it starts, and, when resolution refuses, shall stop the loop with the gate's refusal text without leasing a card.

### D.4 The `moai factory relaunch` verb

- **REQ-SRH-012** (Ubiquitous): The `moai factory relaunch` verb shall accept `--lane lane-<n>`, `--provider cc|glm|codex`, `--run <id>`, `--from-run <id>`, and `--dry-run`, and shall re-enter a lane session into a live run by launching the provider's lane-join entry — `-f lane-<n>` when `--lane` is given, `-f lane` otherwise, and `--factory-run <id>` when `--run` is given — through the shared lane-join gate, so that run resolution, verified leader discovery, and run resume are exactly those of the ordinary join; it shall refuse a legacy lane spelling naming the canonical form, and under `--dry-run` it shall print the launch line it would run, write nothing, and exit 0.
- **REQ-SRH-013** (Event-driven): **When** `--from-run <id>` names a run that measures active and whose owner classifies dead under the existing run-retirement predicate, the verb shall retire that run through that predicate before it launches; where the run is not active, or its owner classifies live or indeterminate, the verb shall leave the run untouched, launch anyway, and name the outcome.
- **REQ-SRH-014** (Unwanted): The verb shall not create, edit, or reactivate run records beyond the retirement of REQ-SRH-013 and what the shared join gate itself performs, shall not read or change the lane's clear policy, and shall not migrate card leases or worktrees; its help text shall state that it is distinct from the `--clear-policy relaunch` loop.

### D.5 Scope

- **REQ-SRH-015** (Unwanted): The change shall be Factory Mode only: it shall not alter, extend, or depend on any kanban-only surface, including the kanban relaunch notice prose and the kanban session-start notice.

### D.6 Naming resolution — two things called "relaunch"

| | `--clear-policy relaunch` (existing) | `moai factory relaunch` (new) |
|---|---|---|
| What it is | A lane clear-policy value; the launcher stays the parent process and loops | A one-shot operator verb |
| What it relaunches | One interactive session per leased card, inside one run | One lane session, into a live run |
| Crosses runs? | No — except where REQ-SRH-011 makes it follow the run | Yes — that is its purpose |
| Touches run records? | No | Only the proof-gated retirement of `--from-run` (REQ-SRH-013) and what the join gate performs |
| Invoked by | `moai cc\|glm -f lane --clear-policy relaunch` | The operator, from the line the notice prints |

The verb relaunches the **lane session into a run**. It does not relaunch, create, or restore the run record: a run is the leader's, and its record changes only through the existing retirement and resume machinery.

## E. Acceptance Overview

Full Given-When-Then matrix, evidence ledger, mutant probes, and quality gates: `acceptance.md` (15 criteria, AC-SRH-001..015). Every AC is mechanically verifiable without a live factory run: the binary-level ones run against a build of this tree invoked by path; the rest are unit or integration tests on a `t.TempDir()` factory database. Each AC names the packages it touches so the run phase scopes re-measurement to them.

## F. Out of Scope

### Out of Scope — Kanban Mode

- Kanban Mode (`-k`) and every kanban-only surface: the kanban relaunch notice prose (`roleValueRelaunch`), the kanban session-start notice, kanban companion roles. Card t1399 is removing Kanban Mode; this SPEC neither extends nor depends on it.

### Out of Scope — Run lifecycle semantics

- Owner-liveness classification, the retirement predicate, the runs schema, and `ResolveActiveRun` / `ReconcileActiveRuns` (SPEC-FACTORY-RUN-RETIRE-001, SPEC-CODEX-FACTORY-RETIRE-001). The hook does not classify owners or retire runs (REQ-SRH-010); the verb reuses the existing predicate unchanged.
- Healing a run that died without being retired while its record still reads active: the hook measures `runs.status` only (decision D3).

### Out of Scope — Leader behaviour

- Leader relaunch, leader re-registration, and the creation of a new run. The verb is lane-only; `moai cc|glm|codex -f` (leader entry) is unchanged.

### Out of Scope — In-flight work

- Migrating card leases, assigned cards, or worktrees from the dead run to the live run (decision D8). A lane's worktree and branch persist on disk regardless; the card record belongs to the run that leased it.

### Out of Scope — Legacy vocabulary

- Mapping legacy labels (`worker-<n>`, `agent-<n>`) to canonical ones, and the refusal semantics themselves (SPEC-ROLE-NAMING-CODE-001). This SPEC changes what the stale-run notice says, not the vocabulary rule.

### Out of Scope — Launcher refusal texts

- The launcher-side refusal messages that also name `moai factory runs --retire` (`internal/cli/factory.go:453`, `:573`; `internal/kanban/factory_slots.go:137`) are CLI errors that already name an executable command; they are unchanged.

### Out of Scope — Live environment mutation

- Rewriting a running session's process environment. Impossible from a hook (SPEC-STALE-RUN-LABEL-001 §F); the rebind is derived from run state each turn, not written into the environment.

## G. Assumptions and Decision Points

Items I could not settle from code or the card text. Each carries the default this SPEC applies and the alternative. Ordered by consequence.

| # | Decision point | Default (applied in this SPEC) | Alternative |
|---|---|---|---|
| D1 | **Where "automatic re-registration" lives for a session that is already running.** | In the hook: the UserPromptSubmit registration path derives the effective run each turn from measured run state (REQ-SRH-004), plus the relaunch loop follows the run (REQ-SRH-011). Derived, not persisted — no new carrier beyond the existing once-per-identity notice marker. | Launcher/verb only: no in-session rebind; a running lane stays locked until the operator relaunches it. Simpler, but it leaves a live session locked out — the measured loss. |
| D2 | **What the verb relaunches.** | The lane session into a live run; plus the existing proof-gated retirement of the stale run named by `--from-run` (REQ-SRH-013). It never creates, edits, or reactivates a run record itself (REQ-SRH-014). | (a) Session only — the operator retires separately with `moai factory runs --retire`; (b) session plus the run record, including a discovery-time resume — widens the verb into the join gate's territory. |
| D3 | **Recognition basis for "died or replaced".** | The hook measures `runs.status` only — the SPEC-STALE-RUN-LABEL-001 REQ-SRL-003 basis. A run that is retired or absent is replaced; a run still recorded active is healthy. | Also classify the named run's owner inside the hook and treat an active-but-dead run as replaced when another run is active. Heals the died-without-retirement case, at the cost of a process probe on a 200 ms hook budget and a hook-side judgement the retirement machinery owns. |
| D4 | **More than one active run.** | Fail closed: no registration; the ambiguity notice lists each candidate with its own command (REQ-SRH-006). | Pick the newest — prohibited by SPEC-FACTORY-RUN-RETIRE-001 REQ-014. |
| D5 | **Claim delivery for a rebound session.** | Additive: keep reading the environment run's broker, also read the rebound run's broker (REQ-SRH-008). Preserves SPEC-STALE-RUN-LABEL-001 REQ-SRL-007's "shall not alter claim delivery" literally. | Substitute: read only the rebound run's broker. One store open instead of two, but it alters delivery for a session that still holds a live peer in the old run. |
| D6 | **SessionStart behaviour for an eligible session.** | Bind nothing, emit no unbind notice; the first prompt registers (REQ-SRH-005). Keeps SPEC-STALE-RUN-LABEL-001 REQ-SRL-004 ("no binding at the /clear boundary") literally true. | Register at SessionStart too — needs a launch-pending row in the new run that does not exist, and breaks REQ-SRL-004's wording. |
| D7 | **SPEC-STALE-RUN-LABEL-001 reconciliation.** | Preserved: cadence (REQ-SRH-003), claim independence (REQ-SRH-008), active-run behaviour (REQ-SRH-009). Changed by design: the literal text of the active-run prescription and of the re-bind line (its tests assert the old text and are updated at the run phase, the `plan.md` §B list); a rebind-eligible current-vocabulary session leaves the unbound state instead of receiving the final unbind notice (REQ-SRH-005). No amendment to SPEC-STALE-RUN-LABEL-001's body is needed — none of its requirements pins the literal. | Amend SPEC-STALE-RUN-LABEL-001 in place. Heavier ceremony for no behavioural gain. |
| D8 | **In-flight card leases across the run switch.** | Out of scope (§F). | A lease-migration step in the verb. Needs a record-level transfer contract the factory record does not define. |
| D9 | **Codex lanes.** | The provider token set is `cc`, `glm`, `codex`; the verb re-executes the lane-join entry of the named provider, so no provider-specific code exists. The hook maps backend `gpt` to `codex`. | `cc`/`glm` only; Codex lanes receive the rebind/unbind semantics but a prose notice (non-executable again). |
| D10 | **Ambiguity notice width.** | One command line per candidate, bounded at three lines plus a count and `moai factory runs` pointer. | Unbounded; risks the hook context budget on a root with many active runs. |

Assumptions (each re-verified at the run phase's RED step; a falsified assumption returns a blocker report):

- **A1.** A hook subprocess cannot change the session's environment (SPEC-STALE-RUN-LABEL-001 §F); the rebind is therefore derived per turn.
- **A2.** The MCP `factory_msg_*` tools take an explicit `run_id` (`internal/cli/mcp_factory_msg.go:38`), so the agent learns the new run id from the rebound notice and the inbox context line; no MCP change is needed.
- **A3.** A broker per run means registering the same slot in a second run's broker is legitimate (`peers.slot` is unique per broker, not globally).
- **A4.** `UserPromptSubmit` fires for every operator prompt and the SessionStart self-dispatch rule leads to an operator or agent turn; an eligible session therefore registers on its first prompt.
- **A5.** Re-executing `moai <provider> -f ...` from the verb is equivalent to the operator typing the line; no launcher-private logic is needed (REQ-SRH-012).

## H. Residual Risk

- **Died-without-retirement is not healed in-session (D3).** A run whose leader died but whose record still reads active — while a successor is active — keeps its lanes locked until a launcher-side reconcile or `moai factory relaunch --from-run <id>` retires it. The verb's retire pre-step is the remedy; the hook does not probe.
- **Time-of-check-to-time-of-use.** The sole active run may retire between the hook's measurement and the registration; the registration then lands in a retired run's broker, the next prompt re-measures, and the session is rebound again or unbound. Self-correcting; named, not prevented.
- **Slot contention.** A relaunched lane and a rebinding lane may both claim the same slot in the target run; REQ-SRH-007 refuses the displacement, the loser receives the command line.
- **Nested invocation.** Running the verb from inside a Claude session starts a nested session. The notice says to end the stale session first; no mechanical guard is specified.
- **Codex relaunch exec** is delegated to the existing codex entry (D9, A5); if that entry cannot be re-executed from the verb without provider-specific code, the run phase returns a blocker report rather than adding it.
