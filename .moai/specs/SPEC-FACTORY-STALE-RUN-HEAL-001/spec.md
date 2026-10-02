---
id: SPEC-FACTORY-STALE-RUN-HEAL-001
title: "Factory stale-run self-healing — automatic lane re-registration on run switch, executable relaunch command in the stale-run notice, and the moai factory relaunch verb"
version: "0.2.0"
status: completed
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
| 2026-10-02 | 0.1.0 | Initial plan-phase draft (card t1345, Tier M, Class C), tree `802a72235536958ada5b7cd5876a168e4b8c325f`. Scope decided by the leader: Factory Mode only (Kanban Mode is being removed by card t1399). |
| 2026-10-02 | 0.2.0 | Plan-audit iteration 1 revision (FAIL 0.65, `.moai/reports/t1345/plan-audit-iter1.md`, defects D1-D18). Codex launch line corrected to the one line the Codex launcher accepts (REQ-SRH-016); honest account of what the hook rebind heals (current-vocabulary sessions) and what the measured legacy-vocabulary incident gets (§B); per-requirement reconciliation with SPEC-STALE-RUN-LABEL-001 and an explicit supersession clause (§D.9); notice state machine (REQ-SRH-003); the additive read of the retired run's broker dropped (REQ-SRH-008); single-measurement design, budgets, and Stop behaviour (REQ-SRH-008..010, `plan.md` §B); notice-line table and message catalogue (§D.7, §D.8); lane admission decision (DP11); Codex per-card loop excluded with its residual risk (§F); behavioural RED-now probe added (`probe/hook-probe.sh`). Decision points renamed DP1-DP12 to avoid colliding with the audit's defect ids. Counts: 16 requirements, 16 acceptance criteria. |

## B. Problem Statement

A factory lane session carries its run identity in its launch environment, and that identity is frozen for the session's life: a hook subprocess cannot change its parent's environment, and `/clear` does not re-read it. When the run that environment names dies or is replaced, three things go wrong (measured 2026-09-28/30 on run `tlwgk9`, retired by the retirement machinery with `{"classification":"dead","basis":"stamp"}`, successor `tm3yoq`):

1. **A running lane never re-registers.** The leader of the new run dispatches through the new run's broker; the lane's hook keeps opening the old run's broker, so nothing reaches it. The t1330 join repair (SPEC-FACTORY-LANE-JOIN-SOCKET-001) fixed the launcher entry only — a lane launched after the run switch joins correctly — and has no consumer for a lane session that is already running.
2. **The stale-run notice prescribes a sentence, not a command.** SPEC-STALE-RUN-LABEL-001 stopped the every-turn repetition for legacy labels and added the one-time unbind notice, but the text is prose: "end this session, retire the run with 'moai factory runs --retire X', then relaunch" (the last step names no command), and a re-bind line printing the placeholder `moai cc -f lane-<n>` (4 occurrences, one per locale).
3. **There is no verb for the way back.** Returning a stale lane to a live run takes a retire step, a manual join, and knowledge of the lane's own label and provider. No `moai factory relaunch` exists.

**What this SPEC heals, stated precisely.** The measured incident lanes (`worker-69`, `worker-72`) carried the legacy vocabulary. Legacy-vocabulary sessions are never re-bound (SPEC-ROLE-NAMING-CODE-001 REQ-RNC-009: detection only, never a mapping), so for them the cure is the printed command (REQ-SRH-001) and the verb (REQ-SRH-012): a **new** session. The 2h40m of in-session context of such a lane is not preserved by this SPEC; its worktree and branch persist on disk regardless. The hook rebind (REQ-SRH-004) heals **current-vocabulary** sessions (`lane-<n>`) **in place**, which is the class the legacy repair leaves behind: those sessions today print `factory messaging degraded: NO_ACTIVE_FACTORY` on every prompt and receive nothing (§C.2, measured).

This SPEC delivers exactly the card's three behaviours: (1) lane re-registration into the live run, (2) an executable command in the notice, (3) the `moai factory relaunch` verb.

## C. Root Cause (mechanism; read from code and observed through the behavioural probe at `cda6913d127c959cee93b54254e6c7241f8b2032`, Go code identical to `802a72235`)

1. **The environment pins the run; the broker is per run.** The two in-session readers of `MOAI_KANBAN_ID` are the peer-registration path (`internal/hook/factory_messages.go:55`) and the inbox claim path (`factory_messages.go:135`). A run's broker is its own database file (`factorymsg.BrokerPath`: `<factory dir>/messages/<run>/broker.db`), so a lane whose environment names the dead run claims from the dead run's broker while the new leader writes to the new run's broker. The lane receives nothing, silently — the inbound lock. The lane's own card verbs are not pinned (item 5).
2. **A current-vocabulary lane never reaches the stale-run gate — observed.** The run-state gate of SPEC-STALE-RUN-LABEL-001 (`staleRunPrescriptionGate`, `internal/hook/stale_run_gate.go`) is called only from the legacy-label branch (`factory_messages.go:68-70`). A current-vocabulary label whose run is retired falls through to `ValidateActiveRun` (`factory_messages.go:90-92`). Observed through `probe/hook-probe.sh rebind` (E8): prompt 1 and prompt 2 each print `factory messaging degraded: NO_ACTIVE_FACTORY`, and `probe/hook-probe.sh session-start-silent` (E9) shows the same string at SessionStart. Today such a session never receives an unbind notice (`unbindFactoryHookNotice` returns "" for a non-legacy label, `stale_run_gate.go:208-210`); it receives the per-prompt degraded string.
3. **The notice text is not executable.** `internal/hook/session_stale_run.go` — `laneLabelRetire` (and `roleValueRetire`) carry retire-then-"relaunch" prose, `laneLabelUnbindRebind` carries the placeholder. Observed (E10): the legacy-label hook output for an active run is the one-line prose; for a retired run with another active, the re-bind line reads `moai cc -f lane-<n>` in the JSON.
4. **The relaunch clear-policy loop resolves its run once.** `runFactoryLaneRelaunch` (`internal/cli/factory_lane_relaunch.go:40`, run id read at `:56`) passes that id to every `factoryNextLeaseOnce` iteration; card selection lists cards by run id (`factoryNextSelectAndLease` → `ListCards(ctx, runID)`, `factory_card.go:385`) with no run-status check. Its signature carries neither the explicit run selector nor the leader target of the launch. This is an observation from reading; the failing behaviour is observed at the run phase's RED step (the loop launches an interactive `claude` child, which no read-only single command can drive).
5. **The card verbs already follow a run switch.** `moai factory next|stage|complete|assign|status|decide` resolve their run through `resolveFactoryCardRun` (`factory_card.go:1159`): the explicit `--run`, else the single active run. They are not pinned to the environment and are out of the repair surface.
6. **No verb wraps a lane join for an existing session, and the Codex entry is different.** For `moai cc|glm -f lane`, the shared lane-join gate (`enterFactoryLaneRun`, `internal/cli/factory.go:505`) does resolution, verified leader discovery on `NO_ACTIVE_FACTORY`, run resume, and re-entry. The Codex entry does not use that gate: `moai codex -f lane` resolves its run through `enterSelectedFactoryRun(root, "", true, ...)` (`codex_launcher.go`, `runCodexFactoryLane`) — the single active run, no discovery, no `--factory-run`, no `--leader` — and `codexFactoryEntryClassify` refuses every shape except the bare token `lane`: `moai codex -f lane-3` and `moai codex -f lane --factory-run runA` both exit 1 with `FACTORY_MODE_UNSUPPORTED_BACKEND` (E7).
7. **A lane's slot lives in the `workers` registry, pid-keyed.** The launcher claims the lane slot through `ClaimFactoryLane` (cc/glm, no capacity bound) or `ClaimFactoryLaneWithin` (one Codex door, `codex_factory.go:149`); the `workers` row records the label, the session's pid and a `run_id`. The leader's free-slot view (`FactoryFreeSlots`) reads pid liveness only. A rebind therefore finds the slot already held by the session's own live pid.

**Separation evidence (preserved from SPEC-STALE-RUN-LABEL-001):** worker-70 carried the same stale environment yet kept receiving dispatches, because the inbound claim path is keyed on the broker peer record, not on the environment label. A session that is not rebound keeps exactly that path (REQ-SRH-008).

## D. Requirements (GEARS)

### D.1 The notice names an executable command

- **REQ-SRH-001** (Ubiquitous): The factory lane stale-run notice and unbind notice shall name the way back into a live run as executable command lines — complete `moai factory relaunch` invocations with every argument resolved from the session's own launch facts and the measured run state, on their own lines, containing no angle-bracket placeholder, exactly as tabulated in §D.7 — in place of the retire-then-relaunch prose and the `lane-<n>` placeholder line; the notice shall frame the line as the operator's to run from a terminal after ending the session, never as an instruction to the agent.
- **REQ-SRH-002** (Ubiquitous): The command grammar shall be one shared definition used both to print the line in the notices and to parse it in the verb, so that a printed line is accepted verbatim by the verb; the command lines shall be emitted verbatim (command, flags, ids, tokens) in every locale in which a notice is rendered — en, ko, ja, zh on the bootstrap surface, en on the peer-registration surface — with only the surrounding prose localized.
- **REQ-SRH-003** (Ubiquitous): The notice cadence shall follow a state machine. For a legacy-vocabulary session the cadence of SPEC-STALE-RUN-LABEL-001 is unchanged: the active-run prescription at most once per session identity, the unbind notice exactly once and final, and a command line in the unbind notice only while an active run exists. For a current-vocabulary session each distinct state — rebound to run R, unbound (no active run), ambiguous among a set of runs, refused by run R — emits its one notice when entered and is silent while that state persists, and the unbound state is not final: a session that was told no active run exists rebinds when exactly one run later becomes active.

### D.2 Automatic lane re-registration in the hook

- **REQ-SRH-004** (Event-driven): **When** a factory lane session carrying a current-vocabulary lane label submits a prompt while the run its launch environment names measures not active and exactly one other run in the same project root measures active, the hook shall register that session as the lane's peer in the active run under the same lane slot, shall emit the rebound notice (§D.8) when the session enters that state, and shall hand the rebound run to the inbox claim of the same invocation.
- **REQ-SRH-005** (State-driven): **While** a current-vocabulary lane session's named run measures not active, the SessionStart hook shall establish no peer binding and shall emit nothing for it — in particular not the `factory messaging degraded` string — because the first prompt measures the state and carries the rebound, unbound, ambiguity, or refusal notice.
- **REQ-SRH-006** (Event-driven): **When** a current-vocabulary lane session's named run measures not active and the project root holds zero active runs or more than one, the hook shall register no peer in any run; for zero it shall emit the unbound notice and no command line, for several the ambiguity notice with one command line per candidate run (at most three, then a count), and it shall not select among candidates (SPEC-FACTORY-RUN-RETIRE-001 REQ-014) — both replacing the per-prompt degraded string.
- **REQ-SRH-007** (Unwanted): The hook shall not register a legacy-vocabulary session into any run and shall not displace a live owner of the target slot: when the slot in the target run is held by a different live session the registration shall be refused with no write and the refusal notice returned; and the rebind shall not touch the `workers` registry — the slot is already held by the session's own live pid, so the leader's free-slot view is unchanged (DP11).
- **REQ-SRH-008** (Ubiquitous): For a session rebound by REQ-SRH-004 the inbox claim of that invocation shall read the broker of the rebound run and not the environment run's broker (the environment run is not active and the MCP tools refuse a non-active run); for every other session, and for every Stop event, the claim path shall be exactly as before — keyed on the broker peer record of the environment run — and a Stop event shall not rebind.
- **REQ-SRH-009** (Ubiquitous): For a current-vocabulary lane session whose named run measures active, the hook's behaviour — bind, registration, claim, notice output — and the number of run-state measurements per invocation (one path resolution, one query) shall be unchanged from before this SPEC.
- **REQ-SRH-010** (Unwanted): The hook shall fail open and shall not write run records: a measurement or registration failure shall degrade to the existing degraded answer or to silence, never to a hook error or a prescription; the added steps shall run under the existing budgets (the registration steps under the bind budget, the claim under the inspection deadline) and shall be exercisable under injected budgets; and the hook's only writes shall be the peer registration and the notice marker — it shall not retire, reactivate, or create any run.

### D.3 The relaunch clear-policy loop follows the run

- **REQ-SRH-011** (Event-driven): **When** the `moai cc|glm -f lane --clear-policy relaunch` loop begins an iteration, it shall re-resolve the run through the same selection the launch used — an explicit `--factory-run` stays explicit; otherwise the shared lane-join gate, including verified leader discovery with the launch's leader target — before leasing any card, shall hand the resolved run to the child session it starts, and, when resolution refuses, shall stop the loop with the gate's refusal text without leasing a card.

### D.4 The `moai factory relaunch` verb

- **REQ-SRH-012** (Ubiquitous): The `moai factory relaunch` verb shall accept `--provider cc|glm|codex`, `--lane lane-<n>`, `--run <id>`, `--from-run <id>`, and `--dry-run`, and shall re-enter a lane session into a live run by launching the provider's own lane-join entry — for `cc` and `glm`: `-f lane-<n>` when `--lane` is given, `-f lane` otherwise, and `--factory-run <id>` when `--run` is given, through the shared lane-join gate; for `codex`: `-f lane` (REQ-SRH-016) — and under `--dry-run` it shall print the launch line it would run, write nothing, and exit 0; every line it prints shall be accepted by the target launcher's own argument classifier.
- **REQ-SRH-013** (Event-driven): **When** `--from-run <id>` names a run that measures active and whose owner classifies dead under the existing run-retirement predicate, the verb shall retire that run through that predicate before it launches; where the run is not active, or its owner classifies live or indeterminate, or the id is unknown, the verb shall leave the run untouched, launch anyway, and name the outcome.
- **REQ-SRH-014** (Unwanted): The verb shall refuse a legacy lane spelling naming the canonical form, shall not create, edit, or reactivate run records beyond the retirement of REQ-SRH-013 and what the shared join gate itself performs, shall not read or change the lane's clear policy, and shall not migrate card leases or worktrees; its help text shall state that it is distinct from the `--clear-policy relaunch` loop.

### D.5 Scope and provider limits

- **REQ-SRH-015** (Unwanted): The change shall be Factory Mode only: it shall not alter, extend, or depend on any kanban-only surface, including the kanban relaunch notice prose and the kanban session-start notice.
- **REQ-SRH-016** (Ubiquitous): The provider token shall reflect the session's own launch provider — environment backend `glm` maps to `glm`, `gpt` to `codex`, and any other or empty value to `cc` — and for provider `codex` the verb and the notices shall name only `moai codex -f lane` as the launch (no lane pin, no run flag, no discovery: the Codex entry takes the next free lane in the single active run), the verb refusing `--lane` and `--run` with `--provider codex` and naming the Codex limitation.

### D.6 Naming resolution — two things called "relaunch"

| | `--clear-policy relaunch` (existing) | `moai factory relaunch` (new) |
|---|---|---|
| What it is | A lane clear-policy value; the launcher stays the parent process and loops | A one-shot operator verb |
| What it relaunches | One interactive session per leased card, inside one run | One lane session, into a live run |
| Crosses runs? | No — except where REQ-SRH-011 makes the cc/glm loop follow the run | Yes — that is its purpose |
| Touches run records? | No | Only the proof-gated retirement of `--from-run` (REQ-SRH-013) and what the join gate performs |
| Invoked by | `moai cc\|glm -f lane --clear-policy relaunch` | The operator, from the line the notice prints |

The verb relaunches the **lane session into a run**. It does not relaunch, create, or restore the run record.

### D.7 Notice-line table

Provider token `P` per REQ-SRH-016. `X` is the run the environment names, `S` the session's lane label, `A` the set of runs measuring active in the project root. Command lines are printed one per line, no indentation, flag order `--provider`, `--lane`, `--run`, `--from-run`. "Prose" is the existing or catalogued notice text of §D.8; only the command lines are protocol tokens. For `P = codex` the `--lane` and `--run` flags are never printed (REQ-SRH-016).

| Row | Vocabulary | `X` state | `A` | Notice | Command lines printed |
|---|---|---|---|---|---|
| R1 | current | active | — | none (unchanged, REQ-SRH-009) | none |
| R2 | current | not active | exactly one, `Y` | rebound (N1) | none |
| R3 | current | not active | exactly one, `Y`, slot held by a live other session | refused (N4) | `moai factory relaunch --provider P --run Y` |
| R4 | current | not active | zero | unbound (N2) | none |
| R5 | current | not active | two or more | ambiguity (N3) | per candidate `Yi` (sorted, at most three): `moai factory relaunch --provider P --lane S --run Yi`; codex: one line `moai factory relaunch --provider codex` |
| R6 | legacy | active | — | stale-run prescription (N5) | `moai factory relaunch --provider P --from-run X` |
| R7 | legacy | not active | zero | legacy unbind (N6), final | none |
| R8 | legacy | not active | exactly one | legacy unbind (N6) | `moai factory relaunch --provider P` |
| R9 | legacy | not active | two or more | legacy unbind (N6) | per candidate `Yi` (sorted, at most three): `moai factory relaunch --provider P --run Yi`; codex: one line `moai factory relaunch --provider codex` |

Why R6 carries `--from-run X` and not `--run X`: `--run X --from-run X` would retire `X` and then demand `X` (`NO_ACTIVE_FACTORY`), while `--from-run X` alone retires it only when its owner is dead and otherwise lets the join take the live `X` as `-f lane` — the outcome the old "retire, then relaunch" prose aimed at. R3 omits `--lane` because a pinned slot the live session holds would be refused by the launcher's claim; the auto-numbered join takes the next free slot.

### D.8 Message catalogue

Surfaces: the **peer-registration surface** (UserPromptSubmit, agent-facing, English — `registerFactoryHookPeer` passes `langEnglish`) and the **bootstrap surface** (SessionStart, English for the agent copy and the operator's `conversation_language` for the operator copy). The four-locale obligation of REQ-SRH-002 applies to the strings that already exist in four locales (N5, N6); N1-N4 are new agent-facing strings on the peer-registration surface and are English-only. Protocol tokens, never translated: the `factory lane ...:` / `stale run:` prefixes, run ids, slot labels, generation numbers, command lines.

| Id | Purpose | Surface and locales | Protocol tokens | Cadence carrier |
|---|---|---|---|---|
| N1 rebound | Names the rebind: previous run, new run, slot, generation, and that messages arrive at turn boundaries | peer-registration, en | prefix `factory lane rebound:`; `X`, `Y`, `S`, generation | the notice marker's state field = `rebound:Y` |
| N2 unbound | One-time notice for a current-vocabulary session with no active run; not final | peer-registration, en | prefix `factory lane unbound:`; `S`, `X`, measured state | state field = `unbound:X` |
| N3 ambiguity | Names the not-active run and every candidate run; points at the command lines | peer-registration, en | prefix `factory lane unbound:`; candidate ids; command lines | state field = `ambiguous:` + candidate ids |
| N4 refusal | The slot in the target run is held by a live session; no write occurred | peer-registration, en | prefix `factory lane rebind refused:`; `S`, `Y`; command line | state field = `refused:Y` |
| N5 stale-run prescription (existing text, command replaced) | Legacy lane, run active: end the session and relaunch under the current vocabulary | bootstrap en/ko/ja/zh; peer-registration en | prefix `stale run:`; label, run id; command line | existing marker kind `prescription` (once per identity) |
| N6 legacy unbind (existing text, rebind line replaced) | Legacy lane, run not active | bootstrap en/ko/ja/zh; peer-registration en | prefix `stale run:`; label, run id, measured state; command lines | existing marker kind `unbind` (once, final) |

N1-N4 share one new marker field, `state`, holding the last emitted state key; a notice is emitted only when the computed state key differs from the stored one. N5 and N6 keep SPEC-STALE-RUN-LABEL-001's two fields untouched.

### D.9 Relation to SPEC-STALE-RUN-LABEL-001 — supersession clause

SPEC-STALE-RUN-LABEL-001 (completed) is **not amended**; where this SPEC and it differ as tabulated, **this SPEC governs for factory sessions**, and the run phase updates the tests that assert the superseded text. Per requirement:

| SRL requirement | Disposition | Where |
|---|---|---|
| REQ-SRL-001 (no retire prescription when the run is not active) | Preserved | REQ-SRH-003 (legacy), rows R7-R9 |
| REQ-SRL-002 (prescription at most once per identity when active) | Preserved in cadence; **text superseded** — the retire-then-relaunch prose becomes the command line | REQ-SRH-001, REQ-SRH-003, row R6 |
| REQ-SRL-003 (measure through the shared accessor, never re-derive) | Preserved and extended — the current-vocabulary path now measures through the same tri-state accessor in place of `ValidateActiveRun` | REQ-SRH-009 |
| REQ-SRL-004 (no binding at the `/clear` boundary for a not-active run) | Preserved at SessionStart for every vocabulary; the rebind is a UserPromptSubmit act into an **active** run, not a binding to the dead run | REQ-SRH-005, REQ-SRH-004 |
| REQ-SRL-005 (exactly one **final** unbind notice, then silence) | Preserved for legacy sessions; **superseded for current-vocabulary sessions** by the state machine (unbound is not final) | REQ-SRH-003 |
| REQ-SRL-006 (the unbind notice names the `moai cc -f lane-<n>` join while an active run exists) | **Superseded**: the line becomes the executable relaunch command; the condition (only while an active run exists) is preserved | REQ-SRH-001, rows R8-R9 |
| REQ-SRL-007, first sentence (the claim path is keyed on the broker peer record alone; the environment-label judgment shall not gate, filter, or delay delivery) | Preserved for every non-rebound session and every Stop event | REQ-SRH-008 |
| REQ-SRL-007, second sentence (the environment-label judgment shall not be coupled into the message channel) | **Amended for a rebound session**: the registration step's measurement of the environment *run's status* selects which run's broker the claim opens; the environment *label* still never gates delivery, and delivery within that broker is still keyed on the peer record | REQ-SRH-008 |
| REQ-SRL-008 (current-vocabulary session on an active run unchanged) | Preserved | REQ-SRH-009 |
| REQ-SRL-009 (environment names through the config constants) | Preserved | `acceptance.md` D.5 |
| AC-SRL-005(b) (asserts the text `moai cc -f lane-`) | **Superseded**: the assertion becomes the exact relaunch line | AC-SRH-005 |

## E. Acceptance Overview

Full matrix, evidence ledger, mutant probes, quality gates: `acceptance.md` (16 criteria, AC-SRH-001..016). Release-blocking criteria carry a plan-phase RED-now cell measured through a binary built from the card tree (the verb absent, or the behavioural probe `probe/hook-probe.sh`); criteria whose RED cannot be observed by one read-only command at plan time are classified High (regression-guard) with the reason stated — none is release-blocking with a deferred RED. Each AC names the packages it touches.

## F. Out of Scope

### Out of Scope — Kanban Mode

- Kanban Mode (`-k`) and every kanban-only surface: the kanban relaunch notice prose (`roleValueRelaunch`), the kanban session-start notice, kanban companion roles. Card t1399 is removing Kanban Mode; this SPEC neither extends nor depends on it. The shared command builder is a new, factory-named file in `internal/kanban` (the package also holds the factory lane-label helpers), so the change is additive and does not collide with t1399's removals.

### Out of Scope — Run lifecycle semantics

- Owner-liveness classification, the retirement predicate, the runs schema, and `ResolveActiveRun` / `ReconcileActiveRuns` (SPEC-FACTORY-RUN-RETIRE-001, SPEC-CODEX-FACTORY-RETIRE-001). The hook does not classify owners or retire runs (REQ-SRH-010); the verb reuses the existing predicate unchanged.
- Healing a run that died without being retired while its record still reads active: the hook measures `runs.status` only (DP3).

### Out of Scope — Codex per-card loop

- The Codex per-card loop `runCodexFactoryLane` (`codex_launcher.go`) resolves its run once through `enterSelectedFactoryRun` and then leases from it in `for {}` — the same read-the-run-once defect as §C.4, on the only Codex factory entry. REQ-SRH-011 covers the cc/glm loop only. Residual risk: a Codex lane keeps leasing from a retired run until the operator relaunches it; `moai factory relaunch --provider codex` is the remedy this SPEC provides.

### Out of Scope — Leader behaviour

- Leader relaunch, leader re-registration, and the creation of a new run. The verb is lane-only; `moai cc|glm|codex -f` (leader entry) is unchanged. A legacy-vocabulary *leader* keeps its existing notice.

### Out of Scope — In-flight work

- Migrating card leases, assigned cards, or worktrees from the dead run to the live run (DP8). A lane's worktree and branch persist on disk regardless; the card record belongs to the run that leased it.

### Out of Scope — Legacy vocabulary

- Mapping legacy labels (`worker-<n>`, `agent-<n>`) to canonical ones, and the refusal semantics themselves (SPEC-ROLE-NAMING-CODE-001). This SPEC changes what the stale-run notice says, not the vocabulary rule. Legacy sessions are not re-bound; they receive the command for a new session.

### Out of Scope — Launcher refusal texts

- The launcher-side refusal messages that also name `moai factory runs --retire` (`internal/cli/factory.go:453`, `:573`; `internal/kanban/factory_slots.go:137`) are CLI errors that already name an executable command; they are unchanged.

### Out of Scope — Live environment mutation

- Rewriting a running session's process environment. Impossible from a hook (SPEC-STALE-RUN-LABEL-001 §F); the rebind is derived from run state on each prompt, never written into the environment.

## G. Assumptions and Decision Points

Items not settleable from code or the card text. Each carries the default this SPEC applies and the alternative. Ordered by consequence. (Prefix DP, not D, so they cannot be confused with the plan-audit's defect ids.)

| # | Decision point | Default (applied) | Alternative |
|---|---|---|---|
| DP1 | **Where re-registration lives for a running session.** | In the hook: the UserPromptSubmit registration path derives the rebound run from measured run state on each prompt (REQ-SRH-004) and hands it to that invocation's claim; the cc/glm relaunch loop follows the run (REQ-SRH-011). No persisted run carrier — only the notice marker's `state` field. | Launcher/verb only: a running lane stays locked until the operator relaunches it — the measured loss. |
| DP2 | **What the verb relaunches.** | The lane session into a live run, plus the existing proof-gated retirement of the stale run named by `--from-run` (REQ-SRH-013). It never creates, edits, or reactivates a run record itself (REQ-SRH-014). | Session only; or session plus run-record resume. |
| DP3 | **Recognition basis for "died or replaced".** | `runs.status` only — the SPEC-STALE-RUN-LABEL-001 REQ-SRL-003 basis. A retired or absent run is replaced; a run still recorded active is healthy. | Classify the owner inside the hook and treat an active-but-dead run as replaced. Heals died-without-retirement at the cost of a process probe on a 200 ms budget and a judgement the retirement machinery owns. |
| DP4 | **More than one active run.** | Fail closed: no registration, ambiguity notice with one command per candidate (REQ-SRH-006). | Pick the newest — prohibited by SPEC-FACTORY-RUN-RETIRE-001 REQ-014. |
| DP5 | **Claim for a rebound session.** | Read only the rebound run's broker (REQ-SRH-008): the environment run is not active and the MCP tools refuse a non-active run, so messages from it cannot be acted on. Stop does not rebind and reads the environment run as before. | Additive read of both brokers — merged-line format, truncation, and the SRL-007 wording problem for no actionable gain; dropped. |
| DP6 | **SessionStart for a current-vocabulary session on a not-active run.** | Silence for every such session, eligible or not (REQ-SRH-005); the first prompt carries the notice or the rebind. | Register at SessionStart — needs a launch-pending row in the new run that does not exist. |
| DP7 | **SPEC-STALE-RUN-LABEL-001 reconciliation.** | Per-requirement table plus supersession clause (§D.9); that SPEC is not amended (it is completed; its tests are updated at the run phase). | Amend it in place through the amendment procedure — heavier ceremony for no behavioural difference. |
| DP8 | **In-flight card leases across the run switch.** | Out of scope (§F). | A lease-migration step in the verb; needs a record-level transfer contract the factory record does not define. |
| DP9 | **Codex provider.** | Option (a): the verb prints and launches only `moai codex -f lane` for `--provider codex`; `--lane`/`--run` are refused with the Codex limitation named (REQ-SRH-016). | Exclude codex from the verb and print the existing launcher line in the notice for Codex sessions — the same line, without a uniform verb. |
| DP10 | **Ambiguity notice width.** | One command line per candidate, at most three, then a count and the `moai factory runs` pointer. | Unbounded; risks the hook context budget. |
| DP11 | **Lane admission and capacity on rebind.** | **Bypass declared, as a tested exclusion.** Reusing `ClaimFactoryLaneWithin` is infeasible: the lane's slot is already held in `workers` by the session's own live pid, so a bounded explicit-slot claim returns `factory lane "lane-3" is already occupied in run <Y>` (`factory_slots.go`, the `taken[n]` branch of `claimFactoryLane` when `maxSlots > 0`), and releasing the own row first is a second write racing the leader. cc/glm joins call `ClaimFactoryLane` with no capacity bound either, so the rebind keeps exact parity with them. Consequence: declared lane capacity is not enforced on rebind (a stale `lane-9` can rebind into a run declared with 3), and the leader's free-slot view is unchanged — `FactoryFreeSlots` is pid-keyed and the pid is the same live session. AC-SRH-008 asserts the `workers` rows and `FactoryFreeSlots` are identical before and after. | Add a capacity refusal reading `runs.lane_capacity` on the rebind path; a new mechanism the cc/glm join does not have. |
| DP12 | **Cadence carrier for N1-N4.** | One new `state` field on the existing notice marker, holding the last emitted state key (§D.8). | Four new marker kinds — more fields, and none re-fires when a state recurs. |

Assumptions (each re-verified at the run phase's RED step; a falsified one returns a blocker report):

- **A1.** A hook subprocess cannot change the session's environment (SPEC-STALE-RUN-LABEL-001 §F); the rebind is therefore derived per prompt.
- **A2.** The MCP `factory_msg_*` tools take an explicit `run_id` (`internal/cli/mcp_factory_msg.go:38`) and refuse a non-active run, so the agent learns the new run id from the rebound notice and the inbox context line.
- **A3.** A broker per run means registering the same slot in a second run's broker is legitimate (`peers.slot` is unique per broker file).
- **A4.** `UserPromptSubmit` fires on operator prompts. Whether an autonomous self-dispatch lane fires it between operator prompts is **not measured**; the consequence is stated in §H.
- **A5.** For `cc` and `glm`, re-executing `moai <provider> -f ...` from the verb is equivalent to the operator typing the line. For `codex` the equivalent line is `moai codex -f lane` only (E7 measures every other shape refused).

## H. Residual Risk

- **Legacy-vocabulary lanes are not healed in place (§B).** Their cure is a new session; the in-session context of a legacy lane is lost.
- **Died-without-retirement is not healed in-session (DP3).** A run whose leader died but whose record still reads active — while a successor is active — keeps its lanes locked until a launcher-side reconcile or `moai factory relaunch --from-run <id>` retires it.
- **Autonomous lanes may never rebind (A4).** A self-dispatch lane driven by the SessionStart rule and the Stop hook with no operator prompt for long stretches fires no UserPromptSubmit, so it registers nothing and stays unnoticed until the operator prompts it or the leader's roster shows the missing peer. The Stop event deliberately does not rebind (REQ-SRH-008).
- **A rebound lane receives the new inbox at prompt boundaries only.** Stop-time delivery for a rebound session reads the environment run's broker, as before; the new run's messages arrive on the next prompt.
- **Declared lane capacity is not enforced on rebind (DP11).**
- **Time-of-check-to-time-of-use.** The sole active run may retire between measurement and registration; the next prompt re-measures and the session rebinds or is told it is unbound. Self-correcting; named, not prevented.
- **Slot contention.** A relaunched lane and a rebinding lane may both want one slot; REQ-SRH-007 refuses the displacement and the loser receives the command line.
- **Nested invocation.** The notice frames the command as the operator's, to run from a terminal after ending the session; no mechanical guard stops an agent from running it.
- **Codex.** `moai codex -f lane` takes the next free slot, so a Codex lane's number can change on relaunch; the Codex per-card loop keeps its read-once defect (§F).
