# Progress — SPEC-AGENT-MODEL-INHERIT-001

## §E.1 Plan-phase Audit-Ready Signal

- Card: t1246. Branch `WT-agent-model-inherit`, base develop `d6992e3a0`. Tier L.
- Artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md; run-entry touch set `.moai/reports/t1246/touch-set.{sh,txt}`.
- Inventory measured 2026-09-26 (research.md). Budget baseline: 77539 / 77600 tokens, headroom 61.
- Run gate: `git merge-base --is-ancestor WT-rules-diet develop` exit 0 (REQ-AMI-001; exit 1 on 2026-09-26); t1257 overlap re-measured at run entry (REQ-AMI-002; intersection 0 at `024b95f77`).
- Operator questions — RESOLVED 2026-09-26 (answered by the operator in the lane window, relayed by the coordinator):
  - Q1 — web console: delete the agent-settings tab entirely (UI + API). → REQ-AMI-011, design D7.
  - Q2 — leftover user config keys: `moai update` removes them and lists them in the update report. → REQ-AMI-014, design D8/D14.
  - Q3 — extra scope included: remove `workflow_agents`, `model_routing`, `model_routing_profiles`, `performance_tier`, and dynamic-workflow `agent()` model/effort values. → REQ-AMI-007, REQ-AMI-013, design D12.
  - Q4 — `init`/`update --profile`: keep the flag, no-op, emit a deprecation warning. → REQ-AMI-016, design D10; extended by measurement to `--model-policy` / `--high` / `--medium-alias` / `--low` (design D13).
  - Q5 — harness v4 manifest specialist `model`/`effort`: optional; `/moai:harness` stops emitting them on generation; existing manifests still parse. → REQ-AMI-003, REQ-AMI-006, design D11.
  - Q6 — docs-site: in scope as M8 of this SPEC. → REQ-AMI-025.
- plan-audit iter-1: FAIL 0.76 (`.moai/reports/plan-audit/SPEC-AGENT-MODEL-INHERIT-001-review-1.md`, commit `5faf93bb8`). D1–D16 addressed in spec v0.3.0 (see spec.md HISTORY).
- Counts: 25 requirements (REQ-AMI-001..025, contiguous), 25 acceptance criteria (AC-AMI-001..025) — both at the Tier L ceiling of 25.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
