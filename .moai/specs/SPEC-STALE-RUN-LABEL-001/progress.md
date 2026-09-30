# SPEC-STALE-RUN-LABEL-001 — Progress

## Plan Phase (2026-09-30, card t1373)

- Status: draft — spec.md / plan.md / acceptance.md authored by manager-spec, then repaired at audit iteration 1.
- Research: read-only; mechanism verified in this worktree (`internal/hook/factory_messages.go`, `internal/hook/session_stale_run.go`, `internal/hook/session_start_factory.go`, `internal/factorymsg/store.go` `ValidateActiveRun`, `internal/homestate/factory_run_retire.go` `retireRun`, `internal/cli/codex_factory.go` env stamp, `internal/kanban/bootstrap.go` legacy vocabulary).
- Root cause summary: legacy-label branch short-circuits to the retire prescription before any run-state measurement; env residue survives /clear because the identity lives in the session process env (plus tmux pane env); no unbind path exists. Worker-70 separation: inbound Claim path is peer-record-keyed and never reads the env label.
- SPEC ID self-check: `SPEC-STALE-RUN-LABEL-001` regex PASS (executed), uniqueness confirmed against `.moai/specs/`.
- Next: plan-audit iteration 2 (delta scope D1-D5), then M1 (run-state-gated prescription + unbind state).

## Audit record

- **Iteration 1 — FAIL, score 0.81** (`.moai/reports/t1373/plan-audit-r1.md`, audited_sha d194083fb, 2026-09-30). 5 blocking defects (D1 carrier-premise false, D2 sweep instrument report-not-verdict, D3 unmapped release-blocking test, D4 key-set unenumerated, D5 tier field missing) + 3 optional (D6, D7, D8 — D8 no fix required).
- **Iteration 2 fix pass (spec/plan/acceptance 0.2.0, this update)**:
  - D1: §C.1 carrier sentence rewritten to the measured fact (settings.local.json carries NO factory keys — auditor measured 0 `MOAI_*` keys live, no writer in the tree; residue is process env + tmux pane env). Took the auditor's preferred **DROP branch** for the persisted-carrier scrub requirement (old REQ-SRL-005): no writer exists to guard against (Enforce Simplicity). Cascaded: acceptance AC-SRL-005 shrinks to unbind-notice + re-bind line; plan M2 scrub item removed; spec Out of Scope gains the dropped-scrub bullet.
  - D2: AC-SRL-009 / E5 instrument replaced with diff-scoped added-line extraction (`TestNoNewEnvNameLiteralsInDiff`, added-line count logged); measured baseline 28 total (hook+factorymsg 12, cli 16) pinned as the RED cell with its false-positive sources.
  - D3: `TestPrescriptionGateUnavailableFailsOpen` added to plan M1 (M1.1 + M1 test list); AC matrix Test column now carries owning milestones for all 9 ACs.
  - D4: dissolved by the D1 drop (no key-set to enumerate).
  - D5: `tier: M` added to spec.md frontmatter.
  - D6 (optional): REQ-SRL-003 clarifying parenthetical — the accessor's own DB-file probe is part of the shared measurement; hook-side re-derivation from broker-file absence is the prohibited act.
  - D7 (optional): plan M1.2 clause — the once-per-session-identity dedup carrier is shared across ALL prescription surfaces (SessionStart bootstrap + UserPromptSubmit peer path), so turn 1 cannot emit twice.
  - Renumbering: old REQ-SRL-006..010 → REQ-SRL-005..009 (sequential, no gaps); AC IDs unchanged AC-SRL-001..009 with mappings updated (005→REQ-005/006, 006→REQ-007, 007→REQ-008, 008→REQ-003, 009→REQ-009).

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: spec.md, plan.md, acceptance.md, progress.md (Tier M set) under `.moai/specs/SPEC-STALE-RUN-LABEL-001/`.
- Frontmatter: 12 canonical fields present + `tier: M`; `status: draft`; version 0.2.0.
- Out of Scope: four H3 topics including the explicit card t1345 exclusion and the audited dropped-scrub note.
- Open items for audit iteration 2: RED-now cells for the Go-test release-blocking ACs are scheduled at M1 RED (plan-phase author cannot execute unexported-function tests without writing test files); AC-SRL-009's RED cell is already measured and pinned (auditor baseline 28/12 @ d194083fb); unbind-state carrier (session-record vs factorymsg peer marker) remains the M1 review decision, now with the cross-surface sharing clause (D7) attached.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending run-phase>_
