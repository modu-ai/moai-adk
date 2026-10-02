# SPEC-FACTORY-STALE-RUN-HEAL-001 — Progress

## Plan Phase (2026-10-02, card t1345)

- Status: draft — spec.md, plan.md, acceptance.md authored by manager-spec (Tier M, 3 artifacts plus this record).
- Research: read-only; mechanism read at tree `802a72235536958ada5b7cd5876a168e4b8c325f` (`internal/hook/factory_messages.go`, `stale_run_gate.go`, `session_stale_run.go`, `session_start_factory.go`; `internal/factorymsg/run_state.go`, `store.go`; `internal/cli/factory.go`, `factory_lane_relaunch.go`, `factory_card.go`, `cc.go`; `internal/homestate/factory_run_retire.go`, `factory_run_resume.go`; `internal/discovery/`).
- Scope boundary (leader): Factory Mode only — Kanban Mode is being removed by card t1399; no kanban-only surface is touched or depended on.
- Root cause summary: the environment pins the run (the hook's only in-session reader), a run's broker is its own database, so a lane stays on the dead run's broker while the new leader writes to the new run's; current-vocabulary labels never reach the SPEC-STALE-RUN-LABEL-001 gate and degrade on every prompt; the notice text is prose with a `lane-<n>` placeholder; the `--clear-policy relaunch` loop resolves its run once; no verb wraps the t1330 join gate for an existing session. The card verbs (`next/stage/complete/...`) already resolve the run from the database and are not part of the repair.
- SPEC ID self-check: `SPEC-FACTORY-STALE-RUN-HEAL-001` — regex check executed as Bash, output `PASS`; uniqueness confirmed against `.moai/specs/` (0 existing matches).
- Counts: 15 requirements (Tier M ceiling 16), 15 acceptance criteria (ceiling 16).
- Top decision points (spec.md §G): D1 where re-registration lives for a running session; D2 what the verb relaunches; D3 recognition basis (status only, no in-hook owner probe).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-02

- Artifacts: spec.md, plan.md, acceptance.md, progress.md under `.moai/specs/SPEC-FACTORY-STALE-RUN-HEAL-001/`.
- Frontmatter: 12 canonical fields plus `tier: M`, `card`, `depends_on`, `related_specs`; `status: draft`; version 0.1.0.
- Out of Scope: seven H3 topics including the explicit Kanban Mode exclusion.
- Open items for the plan-audit: Go-test RED cells are scheduled at M1/M2/M3 RED (not executable at plan phase); AC-SRH-001/002/005 RED cells are measured and pinned in `acceptance.md` D.1 (E1-E4); decisions D1-D10 await operator or auditor confirmation at the Kickoff.

## §E.2 Run-phase Evidence

_pending run-phase_

## §E.3 Run-phase Audit-Ready Signal

_pending run-phase_

## §E.4 Sync-phase Audit-Ready Signal

_pending sync-phase_
