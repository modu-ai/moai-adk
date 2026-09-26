# progress.md — SPEC-AUTONOMY-GATE-REWIRE-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-09-26
- spec_version: 0.3.1 (revision base 781ddc355 = v0.3.0; operator re-decision 2026-09-26 — no Jev-solo decider, `llm+jev` cross-check only; decision rules owned by A3; plan-audit iter-1 dispositions in plan.md §I)
- tier: L (spec.md, plan.md, acceptance.md, design.md, research.md)
- requirements: 25 (REQ-GR-001..025, contiguous; old 022+023 merged into 022, old 024/025 -> 023/024, new 025 = amendment linkage) / acceptance criteria: 25 (AC-GR-001..025)
- a1_reference_baseline: 65e0a9167 (SPEC-AUTONOMY-CONTRACT-001 0.5.1, schema owner); history 8f77d9a33 / 98cb7879d / 4208a3a3b / 652243c72 / 6d98ca466 / 67a2f55cb
- a2_reference: lead-stated final format (escalation/<class>-<fingerprint>.md + YAML header + revoke kind); current A2 commit d8926ff9a carries the withdrawn JSON format
- a2b_reference: t1245 owns push serialization, agent-origin sign deny, decide refusal in MOAI_FACTORY_ROLE=agent sessions (SPEC not read; marker kept)
- store: $MOAI_HOME/db/<project-key>/contract/{receipts.jsonl,events.jsonl} (lead decision R10)
- deciders: guided=human; contract autonomous = llm (default) | llm+jev; jev refused (R5); outcome in {approve, reject, human} derived by A3 rules R1-R5
- open_clarifications: none (defaults D-2, D-4 recorded; operator may override)
- lead_confirmations_pending: A1 SPEC text amendment ownership for the interim-rule lift and REQ-CONTRACT-019 (design.md §2 row 27, research.md §10.4); moai-store provenance value (research.md §10.4)
- run_preconditions: t1234 (A1 >= 0.5.1), t1235 (A2, with final escalation format), t1245 (A2b) merged into develop; t1175 merged and absorbed; BASE recorded in §E.2; CLAUDE.local.md §29 edit only after operator confirms in the lane session
- autonomous_kickoff_activation: see design.md §7.1 (single source); amendment linkage REQ-GR-025
- Implementation Kickoff Approval: not requested at plan phase

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
