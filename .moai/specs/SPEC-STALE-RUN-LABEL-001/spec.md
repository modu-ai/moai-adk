---
id: SPEC-STALE-RUN-LABEL-001
title: "Stale factory run label — run-state-gated prescription, /clear env residue, and the orphan-label unbind path"
version: "0.2.0"
status: completed
created: 2026-09-30
updated: 2026-10-01
author: manager-spec
priority: P1
phase: "v3.2.1 target"
module: "internal/hook, internal/factorymsg, internal/homestate, internal/cli"
lifecycle: spec-anchored
tier: M
tags: "factory, stale-run, env-residue, hook-prescription, unbind, card-t1373"
card: t1373
related_specs: [SPEC-ROLE-NAMING-CODE-001, SPEC-FACTORY-RUN-RETIRE-001, SPEC-FACTORY-SELF-DISPATCH-001, SPEC-FACTORY-BOOTSTRAP-001]
---

# SPEC-STALE-RUN-LABEL-001 — Stale factory run label repair

## A. History

| Date | Version | Change |
|---|---|---|
| 2026-09-30 | 0.1.0 | Initial draft (card t1373, plan phase). Three measured defect cases 2026-09-30: run tlwgk9 retired/dead, lanes worker-69/72 carried the dead label, prescription repeated on every turn, 2h40m session work lost (card t1345 record), 3 lanes locked out of dispatch reception. Worker-70 carried the same residue yet kept receiving dispatches — separation evidence. |
| 2026-09-30 | 0.2.0 | Audit iteration 1 repair (`.moai/reports/t1373/plan-audit-r1.md`, FAIL 0.81, D1-D5 blocking): corrected the §C.1 persisted-carrier premise to the measured fact — `.claude/settings.local.json` carries NO factory keys (auditor-measured 0 `MOAI_*` keys live, no writer in the tree); the residue lives in process env + tmux pane env. DROPPED the persisted-carrier scrub requirement (old REQ-SRL-005; REQ-SRL-006..010 renumbered 005..009). Replaced the env-literal sweep instrument with a diff-scoped added-line extraction, baseline 28/12 pinned as its RED cell. Mapped `TestPrescriptionGateUnavailableFailsOpen` to M1 and annotated AC owning milestones. Added `tier: M` (D5). Applied optional D6 (file-absence phrasing) and D7 (cross-surface dedup carrier). |

## B. Problem Statement

After a factory run is retired (`runs.status='retired'` in the factory DB, `internal/homestate/factory_run_retire.go`), the lane sessions of that run keep the run's factory identity in their process environment (`MOAI_KANBAN_ID`, `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS`, `MOAI_FACTORY_ROLE`, … — stamped at launch by the factory launcher, e.g. `internal/cli/codex_factory.go` env-key list). `/clear` never re-evaluates these keys, so the dead run's label survives into every subsequent turn.

The hook path compounds the residue: `registerFactoryHookPeer` (`internal/hook/factory_messages.go:54-71`) checks the launch label BEFORE any run-state measurement. A legacy-vocabulary label (`worker-<n>`, per `kanban.IsLegacyFactoryRoleValue`) short-circuits into `legacyFactoryHookNotice` (`internal/hook/session_stale_run.go:140`), which renders the unconditional prescription "end this session, retire the run with 'moai factory runs --retire <id>', then relaunch" — on every `UserPromptSubmit` (and the SessionStart surfaces at `session_start_factory.go:52-53` / `session_start.go:466`). `factorymsg.ValidateActiveRun` (`internal/factorymsg/store.go:152`), which reads `runs.status` and distinguishes active from retired/dead, is only reached by current-vocabulary labels.

For an already-retired run the prescription is unsatisfiable: the retire was performed, the run measures retired/dead, and the hook re-instructs anyway — an infinite repetition that consumed 2h40m of live session work and locked 3 lanes out of new dispatch reception.

## C. Root Cause (mechanism, measured)

Three independent surfaces compose the defect:

1. **Env residue root** — factory identity keys are injected into the lane child process environment at launch (launcher env stamp, `internal/cli/codex_factory.go:159-161` and the Claude-launcher equivalent, plus the tmux pane assignments the launcher scrubs-then-restamps per pane launch) and never re-evaluated at the `/clear` boundary. The persisted settings carrier is NOT part of the residue — measured at audit iteration 1 (`.moai/reports/t1373/plan-audit-r1.md` §2): the retire path writes only the runs DB, no code path in this tree writes factory identity into `.claude/settings.local.json`, and the live carrier carries zero `MOAI_*` keys mid-run. A hook subprocess cannot unset the parent session's process environment, so the residue itself is not scrub-able from a live session; what CAN be fixed is that the residue stays *authoritative* — handled by REQ-SRL-004's non-authoritative judgment.
2. **Unconditional prescription** — the legacy-label branch in `registerFactoryHookPeer` (and its SessionStart siblings) returns the retire prescription without consulting `ValidateActiveRun`. The run-state accessor exists and is one call away; the branch predates or ignores it.
3. **No unbind path** — a session carrying a label for a nonexistent run has exactly one prescribed remedy (terminate + relaunch). There is no defined transition to a clean unbound state, and no re-bind entry toward an active run.

**Separation evidence (must preserve):** worker-70 carried the same stale env yet kept receiving dispatches normally. Inbound message delivery (`factoryHookBatch`, `internal/hook/factory_messages.go:130`) is keyed on the broker peer record (session UUID) and never reads the env label; the env-label judgment lives only in the bind/notice path. The two code paths are separate by construction and the repair MUST NOT couple them.

## D. Requirements (GEARS)

### D.1 Prescription gating (axis ② — no re-prescribing a retired run)

- **REQ-SRL-001**: **When** the factory run named by the session's launch environment measures not active (`runs.status` is not `active` in the factory DB under the hook's project root), the hook prescription path shall emit no retire prescription — the response shall be silence or the unbind notice of REQ-SRL-004, never the "retire the run" remedy.
- **REQ-SRL-002**: **While** a session's launch label carries a legacy factory role value and the named run measures active, the hook shall emit the stale-run prescription at most once per session identity — a session that has already received the prescription in its current session identity shall receive silence on subsequent turns. (Today's every-turn repetition is the noise mechanism; the active-run case keeps its one informative emission.)
- **REQ-SRL-003**: The prescription-silence judgment of REQ-SRL-001 shall be measured from the same run-state store the CLI retire path writes (`homestate.FactoryDBPath`, `runs.status`), through a shared accessor — never re-derived from file absence, text patterns, or a second status encoding (the shared accessor's own DB-file probe is part of the shared measurement; what is prohibited is hook-side re-derivation from broker-file absence).

### D.2 Env residue (axis ① — the /clear root)

- **REQ-SRL-004**: **When** a session enters the `/clear` boundary (SessionStart, source `clear`) carrying factory identity env for a run that is not active, the session-start handler shall treat the identity as non-authoritative: no factory peer binding shall be established, and the fresh session shall not inherit an effective dead-run binding from the stale label.

### D.3 Orphan-label path (axis ③ — re-bind or clean unbound state)

- **REQ-SRL-005**: **When** a session's env label references a run that no longer exists as active, the hook path shall transition the session to a defined unbound state: exactly one final unbind notice naming the orphan label and the measured run state, then silence on every subsequent turn of that session.
- **REQ-SRL-006**: **Where** an active factory run exists in the same project root at unbind time, the unbind notice shall name the documented re-bind entry (the `moai cc -f lane-<n>` join) so the operator can return the session to the active run's slot set. Where no active run exists, the notice shall omit the re-bind line.

> A persisted-carrier scrub requirement was considered and **DROPPED** at audit iteration 1 — measured (plan-audit-r1.md §2): the retire path writes only the runs DB, no code path in this tree puts factory identity keys into `.claude/settings.local.json`, and the live carrier carries zero `MOAI_*` keys mid-run; a scrub there guards a path nothing produces (Enforce Simplicity). See History §A row 0.2.0.

### D.4 Separation preservation (worker-70 invariant)

- **REQ-SRL-007**: The inbound message Claim path (`factoryHookBatch`) shall remain keyed on the broker peer record alone: the stale-env judgment of REQ-SRL-001/004/005 shall not gate, filter, delay, or alter message claim delivery for any session. The system shall not couple the env-label judgment into the message channel.
- **REQ-SRL-008**: The stale-env repair shall not change the behavior of a session whose env label is current-vocabulary (`lane-<n>`) and whose run measures active — the existing bind path (`ValidateActiveRun` → `BindLaunchPending`/`RegisterPeer`) is preserved byte-for-byte in behavior.

### D.5 Configuration carriers

- **REQ-SRL-009**: New environment-name references introduced by this SPEC shall use the `internal/config/envkeys.go` constants — no new hardcoded env-name literals in `internal/hook/`, `internal/factorymsg/`, or `internal/cli/`.

## E. Acceptance Overview

Full Given-When-Then matrix, severity, and traceability: `acceptance.md`. Each AC is mechanically verifiable without a live factory run (unit/integration on a `t.TempDir()` factory DB).

## F. Out of Scope

### Out of Scope — Self-healing lane re-registration (card t1345)

- The broader self-healing feature — automatic lane re-registration on run switch and the `moai factory relaunch` verb — is card t1345, a SEPARATE card. This SPEC does not implement auto re-registration; REQ-SRL-007 only names the existing manual join command in a notice.

### Out of Scope — Legacy vocabulary retirement itself

- The leader/lane rename and the legacy-vocabulary refusal semantics are SPEC-ROLE-NAMING-CODE-001 (REQ-RNC-009/-022). This SPEC changes WHEN the refusal notice fires, not the vocabulary or the refusal.

### Out of Scope — Live process env scrubbing

- A hook subprocess cannot unset the parent session's process environment. No requirement of this SPEC shall be implemented by attempting to mutate a live session's environment; the live-session case is covered by the non-authoritative judgment (REQ-SRL-004).
- A persisted-carrier scrub (`.claude/settings.local.json`) was audited and dropped — no writer in this tree produces factory keys in that carrier (plan-audit-r1.md §2); guarding a path nothing produces is scope without a defect.

### Out of Scope — Run lifecycle semantics

- Run retirement semantics, owner-liveness classification, and the runs DB schema are SPEC-FACTORY-RUN-RETIRE-001 / SPEC-CODEX-FACTORY-RETIRE-001. This SPEC is a read-side consumer of `runs.status`.
