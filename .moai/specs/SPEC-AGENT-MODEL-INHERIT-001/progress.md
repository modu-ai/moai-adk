# Progress — SPEC-AGENT-MODEL-INHERIT-001

## §E.1 Plan-phase Audit-Ready Signal

- Card: t1246. Branch `WT-agent-model-inherit`, base develop `d6992e3a0`. Tier L.
- Artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md.
- Inventory measured 2026-09-26 (research.md). Budget baseline: 77539 / 77600 tokens, headroom 61.
- Run gate: card t1175 merged to develop (REQ-AMI-001); re-measure t1257 overlap at run entry (REQ-AMI-002).
- Operator questions — RESOLVED 2026-09-26 (answered by the operator in the lane window, relayed by the coordinator):
  - Q1 — web console: delete the agent-settings tab entirely (UI + API). → REQ-AMI-012, design D7.
  - Q2 — leftover user `llm.yaml` keys: `moai update` removes them and lists them in the update report. → REQ-AMI-015, design D8.
  - Q3 — extra scope included: remove `workflow_agents`, `model_routing`, `model_routing_profiles`, `performance_tier`, and dynamic-workflow `agent()` model/effort values. → REQ-AMI-008, REQ-AMI-014, design D12.
  - Q4 — `init`/`update --profile`: keep the flag, no-op, emit a deprecation warning. → REQ-AMI-017, design D10.
  - Q5 — harness v4 manifest specialist `model`/`effort`: optional; `/moai:harness` stops emitting them on generation; existing manifests still parse. → REQ-AMI-004, REQ-AMI-007, design D11.
  - Q6 — docs-site 48 files: in scope as M8 of this SPEC. → REQ-AMI-026.
- Counts after folding: 25 requirements (REQ-AMI-003 folded into 001/002; REQ-AMI-026 added), 25 acceptance criteria — both at the Tier L ceiling of 25.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
